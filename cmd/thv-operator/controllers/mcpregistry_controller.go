// Package controllers contains the reconciliation logic for the MCPRegistry custom resource.
// It handles the creation, update, and deletion of MCP registries in Kubernetes.
package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	apitypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	mcpv1alpha1 "github.com/stacklok/toolhive/cmd/thv-operator/api/v1alpha1"
	mcpregistry "github.com/stacklok/toolhive/cmd/thv-operator/controllers/mcp_registry"
	"github.com/stacklok/toolhive/cmd/thv-operator/controllers/types"
	controllertypes "github.com/stacklok/toolhive/cmd/thv-operator/controllers/types"
	"github.com/stacklok/toolhive/pkg/logger"
)

// MCPRegistryReconciler reconciles a MCPRegistry object
type MCPRegistryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *MCPRegistryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the MCPRegistry instance
	mcpRegistry := &mcpv1alpha1.MCPRegistry{}
	logger.Info("Reconciling MCPRegistry", "name", req.NamespacedName)
	err := r.Get(ctx, req.NamespacedName, mcpRegistry)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			logger.Info("MCPRegistry resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get MCPRegistry")
		return ctrl.Result{}, err
	}

	// Check if the MCPRegistry is being deleted
	if mcpRegistry.DeletionTimestamp != nil {
		logger.Info("MCPRegistry is being deleted")
		return r.finalizeMCPRegistry(ctx, mcpRegistry)
	}

	// Add finalizer if it doesn't exist
	if !containsString(mcpRegistry.Finalizers, controllertypes.MCPRegistryFinalizer) {
		controllerutil.AddFinalizer(mcpRegistry, controllertypes.MCPRegistryFinalizer)
		if err := r.Update(ctx, mcpRegistry); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Skip status updates during periodic reconciliation to prevent infinite loops
	// Status will only be updated when there's actual work to do

	// Create MongoDB resources
	if err := mcpregistry.CreateMongoDBResources(ctx, r.Client, r.Scheme, mcpRegistry, mcpRegistry.Name, mcpRegistry.Namespace); err != nil {
		logger.Error(err, "Failed to create MongoDB resources")
		return ctrl.Result{}, err
	}

	// Ensure MongoDB status is always set after MongoDB resources are created
	if err := r.ensureMongoDBStatus(ctx, mcpRegistry); err != nil {
		logger.Error(err, "Failed to update MongoDB status")
		return ctrl.Result{}, err
	}

	// Create registry deployment and service
	mongoSvcName := mcpregistry.GetMongoServiceName(mcpRegistry.Name)
	if err := mcpregistry.CreateRegistryResources(ctx, r.Client, r.Scheme, mcpRegistry, mcpRegistry.Name, mcpRegistry.Namespace, mongoSvcName); err != nil {
		logger.Error(err, "Failed to create registry resources")
		return ctrl.Result{}, err
	}

	// Sync associated servers to MongoDB
	serverCount, err := r.syncAssociatedServers(ctx, mcpRegistry)
	if err != nil {
		logger.Error(err, "Failed to sync associated servers")
		return ctrl.Result{}, err
	}

	// Update status to Ready only if we're not already in Ready state
	if mcpRegistry.Status.Phase != mcpv1alpha1.MCPRegistryPhaseReady {
		now := metav1.Now()
		mcpRegistry.Status.Phase = mcpv1alpha1.MCPRegistryPhaseReady
		mcpRegistry.Status.Message = "Registry synced successfully"
		mcpRegistry.Status.LastSyncTime = &now
		mcpRegistry.Status.ServerCount = int32(serverCount)

		// Update conditions
		readyCondition := metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionTrue,
			Reason:             "Synced",
			Message:            "Registry synced successfully",
			LastTransitionTime: now,
		}
		meta.SetStatusCondition(&mcpRegistry.Status.Conditions, readyCondition)

		if err := r.Status().Update(ctx, mcpRegistry); err != nil {
			logger.Error(err, "Failed to update MCPRegistry status")
			return ctrl.Result{}, err
		}
		logger.Info("Updated status to Ready")
	}

	logger.Info("MCPRegistry reconciled successfully")

	// Requeue after the refresh interval
	refreshInterval, err := time.ParseDuration(mcpRegistry.Spec.RefreshInterval)
	if err != nil {
		// Default to 1 hour if parsing fails
		refreshInterval = time.Hour
		logger.Info("Using default refresh interval of 1 hour")
	}

	logger.Info("Scheduling next reconciliation", "interval", refreshInterval)
	return ctrl.Result{RequeueAfter: refreshInterval}, nil
}

// ensureMongoDBStatus ensures that MongoDB connection information is always set in the status
func (r *MCPRegistryReconciler) ensureMongoDBStatus(ctx context.Context, mcpRegistry *mcpv1alpha1.MCPRegistry) error {
	logger := log.FromContext(ctx)

	// Set MongoDB connection information
	mongoSvcName := mcpregistry.GetMongoServiceName(mcpRegistry.Name)
	mongoServiceURL := fmt.Sprintf("mongodb://%s:%s", mongoSvcName, controllertypes.MongoDBPort)

	// Check if MongoDB status needs to be updated
	statusNeedsUpdate := false

	if mcpRegistry.Status.MongoDBServiceURL != mongoServiceURL {
		mcpRegistry.Status.MongoDBServiceURL = mongoServiceURL
		statusNeedsUpdate = true
		logger.Info("Updated MongoDB service URL", "url", mongoServiceURL)
	}

	if mcpRegistry.Status.MongoDBCollectionName != controllertypes.MongoDBCollectionName {
		mcpRegistry.Status.MongoDBCollectionName = controllertypes.MongoDBCollectionName
		statusNeedsUpdate = true
		logger.Info("Updated MongoDB collection name", "collection", controllertypes.MongoDBCollectionName)
	}

	// Only update if there are changes
	if statusNeedsUpdate {
		if err := r.Status().Update(ctx, mcpRegistry); err != nil {
			return fmt.Errorf("failed to update MongoDB status: %w", err)
		}
		logger.Info("MongoDB status updated successfully")
	}

	return nil
}

// syncAssociatedServers syncs MCPServers associated with this registry to MongoDB
func (r *MCPRegistryReconciler) syncAssociatedServers(ctx context.Context, registry *mcpv1alpha1.MCPRegistry) (int, error) {
	logger := log.FromContext(ctx)
	logger.Info("Starting sync of associated servers", "registry", registry.Name)

	// Find all MCPServers that reference this registry
	var servers mcpv1alpha1.MCPServerList
	if err := r.List(ctx, &servers); err != nil {
		logger.Error(err, "Failed to list MCPServers")
		return 0, err
	}

	// Filter servers that reference this registry
	var associatedServers []mcpv1alpha1.MCPServer
	for _, server := range servers.Items {
		if server.Labels[controllertypes.RegistryNameLabel] == registry.Name {
			associatedServers = append(associatedServers, server)
			logger.Info("Found associated server", "server", server.Name, "namespace", server.Namespace)
		}
	}

	logger.Info("Found servers to sync", "count", len(associatedServers), "registry", registry.Name)

	// Connect to MongoDB
	mongoClient, err := r.connectToMongoDB(ctx, registry)
	if err != nil {
		logger.Error(err, "Failed to connect to MongoDB")
		return 0, err
	}
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			logger.Error(err, "Failed to disconnect from MongoDB")
		}
	}()

	// Get the collection
	collection := mongoClient.Database(controllertypes.MongoDBDatabaseName).Collection(controllertypes.MongoDBCollectionName)
	logger.Info("Using MongoDB collection", "database", controllertypes.MongoDBDatabaseName, "collection", controllertypes.MongoDBCollectionName)

	// Sync each server to MongoDB
	for _, server := range associatedServers {
		logger.Info("Syncing server to MongoDB", "server", server.Name, "namespace", server.Namespace)

		// Check if this server has self-registration annotation
		if _, hasServerDetail := server.Annotations[types.ServerDetailAnnotation]; hasServerDetail {
			logger.Info("Processing self-registration path", "server", server.Name)

			serverDetail, err := r.extractServerDetailsFromMCPServer(&server)
			if err != nil {
				logger.Error(err, "Failed to extract server details", "server", server.Name, "namespace", server.Namespace)
				continue
			}

			logger.Info("Successfully extracted server details", "server", server.Name, "serverID", serverDetail.Server.ID)
			err = r.insertServerDetail(ctx, collection, serverDetail)
			if err != nil {
				logger.Error(err, "Failed to insert server details", "server", server.Name, "namespace", server.Namespace)
				continue
			}
		} else if registeredServerID, hasRegisteredID := server.Labels[types.RegisteredServerIDLabel]; hasRegisteredID {
			logger.Info("Processing pre-registration path", "server", server.Name, "registeredID", registeredServerID)

			// Fetch the pre-registered server details from MongoDB
			existingServerDetail, err := r.fetchServerDetailFromMongoDB(ctx, collection, registeredServerID)
			if err != nil {
				logger.Error(err, "Failed to fetch pre-registered server details", "server", server.Name, "registeredID", registeredServerID)
				continue
			}

			// Add remote configuration for this MCPServer
			err = r.addRemoteConfigurationToServerDetail(ctx, collection, existingServerDetail, &server)
			if err != nil {
				logger.Error(err, "Failed to add remote configuration", "server", server.Name, "registeredID", registeredServerID)
				continue
			}

			logger.Info("Successfully added remote configuration", "server", server.Name, "registeredID", registeredServerID)
		} else {
			logger.Info("Server has neither self-registration annotation nor pre-registration label, skipping", "server", server.Name)
		}
	}

	logger.Info("Completed sync of associated servers", "registry", registry.Name, "count", len(associatedServers))
	return len(associatedServers), nil
}

func (r *MCPRegistryReconciler) insertServerDetail(ctx context.Context, collection *mongo.Collection, serverDetail *mcpv1alpha1.ServerDetail) error {
	// Find all elements in the collection and print them
	filter := bson.M{
		"name": serverDetail.Name,
	}

	var existingEntry mcpv1alpha1.ServerDetail
	err := collection.FindOne(ctx, filter).Decode(&existingEntry)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return fmt.Errorf("error checking existing entry: %w", err)
	}
	fmt.Printf("existingEntry: %+v\n", existingEntry)

	if existingEntry.Server.ID != "" {
		fmt.Printf("updating existing entry %s\n", existingEntry.ID)
		// check that the current version is greater than the existing one
		// if serverDetail.VersionDetail.Version <= existingEntry.VersionDetail.Version {
		// 	return fmt.Errorf("version must be greater than existing version")
		// }
		result, err := collection.UpdateOne(
			ctx,
			bson.M{"id": existingEntry.ID},
			bson.M{"$set": bson.M{"versiondetail.islatest": false}})
		if err != nil {
			return fmt.Errorf("error updating existing entry: %w", err)
		}
		if result.UpsertedCount > 0 {
			logger.Info("Inserted new server document", "server", serverDetail.Name, "UpsertedID", result.UpsertedID)
		} else if result.ModifiedCount > 0 {
			logger.Info("Updated existing server document", "server", serverDetail.Name, "UpsertedID", result.UpsertedID)
		} else {
			logger.Info("No changes to server document", "server", serverDetail.Name, "ID", existingEntry.ID)
		}
	} else {
		// Insert the entry into the database
		fmt.Printf("inserting new entry %s: %s\n", serverDetail.ID, serverDetail.Name)
		result, err := collection.InsertOne(ctx, serverDetail)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				// return ErrAlreadyExists
				fmt.Printf("entry already exists, skipping\n")
				return nil
			}
			return fmt.Errorf("error inserting entry: %w", err)
		}
		logger.Info("Inserted new server document", "server", serverDetail.Name, "InsertedID", result.InsertedID)
	}

	return nil
}

// finalizeMCPRegistry handles the cleanup when an MCPRegistry is being deleted
func (r *MCPRegistryReconciler) finalizeMCPRegistry(ctx context.Context, mcpRegistry *mcpv1alpha1.MCPRegistry) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Perform any cleanup operations here
	logger.Info("Finalizing MCPRegistry", "name", mcpRegistry.Name)

	// Remove finalizer
	controllerutil.RemoveFinalizer(mcpRegistry, controllertypes.MCPRegistryFinalizer)
	if err := r.Update(ctx, mcpRegistry); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MCPRegistryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcpv1alpha1.MCPRegistry{}).
		Watches(
			&mcpv1alpha1.MCPServer{},
			handler.EnqueueRequestsFromMapFunc(r.findRegistriesForServer),
		).
		Complete(r)
}

// findRegistriesForServer maps MCPServer events to registry reconciliation
func (r *MCPRegistryReconciler) findRegistriesForServer(ctx context.Context, obj client.Object) []reconcile.Request {
	logger := log.FromContext(ctx)
	logger.Info("Finding registries for server", "server", obj.GetName())
	server := obj.(*mcpv1alpha1.MCPServer)
	logger.Info("Server", "server", server.GetName())
	var requests []reconcile.Request

	// Find registries that should track this server
	registryName := server.Labels[controllertypes.RegistryNameLabel]
	registryNamespace := server.Labels[controllertypes.RegistryNamespaceLabel]

	if registryName != "" {
		// If no namespace is specified, use the same namespace as the server
		if registryNamespace == "" {
			registryNamespace = server.Namespace
		}

		requests = append(requests, reconcile.Request{
			NamespacedName: apitypes.NamespacedName{
				Name:      registryName,
				Namespace: registryNamespace,
			},
		})
	}

	logger.Info("Found registries to reconcile", "count", len(requests), "registry", registryName, "namespace", registryNamespace)

	return requests
}

// Helper function to check if a string slice contains a specific string
func containsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// connectToMongoDB connects to the MomongoClientngoDB instance
func (r *MCPRegistryReconciler) connectToMongoDB(ctx context.Context, registry *mcpv1alpha1.MCPRegistry) (*mongo.Client, error) {
	logger := log.FromContext(ctx)

	// Build MongoDB connection string
	mongoSvcName := mcpregistry.GetMongoServiceName(registry.Name)
	mongoURL := fmt.Sprintf("mongodb://%s:%s", mongoSvcName, controllertypes.MongoDBPort)

	logger.Info("Connecting to MongoDB", "url", mongoURL)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	logger.Info("Successfully connected to MongoDB")
	return client, nil
}

// extractServerDetailsFromMCPServer extracts ServerDetail from MCPServer spec
func (r *MCPRegistryReconciler) extractServerDetailsFromMCPServer(server *mcpv1alpha1.MCPServer) (*mcpv1alpha1.ServerDetail, error) {
	logger := log.FromContext(context.Background())
	logger.Info("Extracting server details", "server", server.Name, "image", server.Spec.Image)

	// Get the server-detail annotation
	serverDetailAnnotation, exists := server.Annotations[types.ServerDetailAnnotation]
	if !exists {
		return nil, fmt.Errorf("server detail annotation is required")
	}

	// Parse the YAML annotation into ServerDetail
	var serverDetail mcpv1alpha1.ServerDetail
	if err := json.Unmarshal([]byte(serverDetailAnnotation), &serverDetail); err != nil {
		return nil, fmt.Errorf("failed to parse server detail annotation: %w", err)
	}

	// Validate that required fields are present
	if serverDetail.ID == "" {
		return nil, fmt.Errorf("server detail must have an ID")
	}
	if serverDetail.Name == "" {
		return nil, fmt.Errorf("server detail must have a name")
	}
	if serverDetail.Description == "" {
		return nil, fmt.Errorf("server detail must have a description")
	}

	serverImage := server.Spec.Image
	var serverBaseImage string
	var imageTag string
	if strings.Contains(serverImage, ":") {
		parts := strings.Split(serverImage, ":")
		serverBaseImage = parts[0]
		imageTag = parts[1]
	} else {
		serverBaseImage = serverImage
		imageTag = "latest"
	}
	serviceName := fmt.Sprintf("mcp-%s-proxy", server.Name)
	serviceURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", serviceName, server.Namespace, server.Spec.Port)

	operatorServerDetail := &mcpv1alpha1.ServerDetail{
		Server: mcpv1alpha1.Server{
			ID:            serverDetail.ID,
			Name:          serverDetail.Name,
			Description:   serverDetail.Description,
			Repository:    serverDetail.Repository,
			VersionDetail: serverDetail.VersionDetail,
		},
		Packages: []mcpv1alpha1.Package{
			{
				RegistryName: "docker",
				Name:         serverBaseImage,
				Version:      imageTag,
			},
		},
		Remotes: []mcpv1alpha1.Remote{
			{
				TransportType: "sse",
				URL:           serviceURL,
				Headers:       []mcpv1alpha1.Input{}, // Empty headers for now
			},
		},
	}

	return operatorServerDetail, nil
}

// fetchServerDetailFromMongoDB fetches a server detail from MongoDB by ID
func (r *MCPRegistryReconciler) fetchServerDetailFromMongoDB(ctx context.Context, collection *mongo.Collection, serverID string) (*mcpv1alpha1.ServerDetail, error) {
	logger := log.FromContext(ctx)
	logger.Info("Fetching server detail from MongoDB", "serverID", serverID)

	filter := bson.M{"id": serverID}
	var serverDetail mcpv1alpha1.ServerDetail

	err := collection.FindOne(ctx, filter).Decode(&serverDetail)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("server with ID %s not found in MongoDB", serverID)
		}
		return nil, fmt.Errorf("failed to fetch server detail from MongoDB: %w", err)
	}

	logger.Info("Successfully fetched server detail", "serverID", serverID, "name", serverDetail.Name)
	return &serverDetail, nil
}

// addRemoteConfigurationToServerDetail adds a remote configuration to an existing server detail
func (r *MCPRegistryReconciler) addRemoteConfigurationToServerDetail(ctx context.Context, collection *mongo.Collection, serverDetail *mcpv1alpha1.ServerDetail, server *mcpv1alpha1.MCPServer) error {
	logger := log.FromContext(ctx)
	logger.Info("Adding remote configuration", "server", server.Name, "serverID", serverDetail.Server.ID)

	// Create the service URL for this MCPServer
	serviceName := fmt.Sprintf("mcp-%s-proxy", server.Name)
	serviceURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", serviceName, server.Namespace, server.Spec.Port)

	// Check if this remote URL already exists
	for _, remote := range serverDetail.Remotes {
		if remote.URL == serviceURL {
			logger.Info("Remote configuration already exists", "server", server.Name, "url", serviceURL)
			return nil
		}
	}

	// Add the new remote configuration
	newRemote := mcpv1alpha1.Remote{
		TransportType: "sse",
		URL:           serviceURL,
		Headers:       []mcpv1alpha1.Input{}, // Empty headers for now
	}

	// Update the server detail in MongoDB
	filter := bson.M{"id": serverDetail.Server.ID}
	update := bson.M{
		"$push": bson.M{
			"remotes": newRemote,
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update server detail with new remote: %w", err)
	}

	if result.ModifiedCount == 0 {
		return fmt.Errorf("no server detail was updated")
	}

	logger.Info("Successfully added remote configuration", "server", server.Name, "url", serviceURL)
	return nil
}
