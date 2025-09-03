package controllers

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	mcpv1alpha1 "github.com/stacklok/toolhive/cmd/thv-operator/api/v1alpha1"
)

const (
	// MCPRegistryFinalizer is used to ensure cleanup of resources
	MCPRegistryFinalizer = "mcpregistry.toolhive.stacklok.dev/finalizer"
	
	// AnnotationSyncTrigger is used to trigger manual syncs
	AnnotationSyncTrigger = "toolhive.stacklok.dev/sync-trigger"
	
	// DefaultSyncInterval is the default automatic sync interval
	DefaultSyncInterval = 1 * time.Hour
)

// MCPRegistryReconciler reconciles MCPRegistry objects
type MCPRegistryReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	sourceHandlers map[string]SourceHandler
	storageManager StorageManager
	statusManager  *StatusManager
}

// NewMCPRegistryReconciler creates a new MCPRegistry reconciler
func NewMCPRegistryReconciler(client client.Client, scheme *runtime.Scheme) *MCPRegistryReconciler {
	statusManager := NewStatusManager(client)
	storageManager := NewConfigMapStorageManager(client)
	
	// Initialize source handlers
	sourceHandlers := map[string]SourceHandler{
		mcpv1alpha1.RegistrySourceTypeConfigMap: NewConfigMapSourceHandler(client),
		// Future handlers can be added here
	}
	
	return &MCPRegistryReconciler{
		Client:         client,
		Scheme:         scheme,
		sourceHandlers: sourceHandlers,
		storageManager: storageManager,
		statusManager:  statusManager,
	}
}

// Reconcile handles MCPRegistry reconciliation
func (r *MCPRegistryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("registry", req.NamespacedName)
	
	// Fetch the MCPRegistry instance
	registry := &mcpv1alpha1.MCPRegistry{}
	if err := r.Get(ctx, req.NamespacedName, registry); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("MCPRegistry resource not found, ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get MCPRegistry")
		return ctrl.Result{}, err
	}

	// Handle deletion
	if registry.GetDeletionTimestamp() != nil {
		return r.handleDeletion(ctx, registry)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(registry, MCPRegistryFinalizer) {
		controllerutil.AddFinalizer(registry, MCPRegistryFinalizer)
		if err := r.Update(ctx, registry); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Initialize status if needed
	if registry.Status.Phase == "" {
		if err := r.statusManager.UpdatePhase(ctx, registry, mcpv1alpha1.MCPRegistryPhasePending, "Initializing registry"); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Perform sync operation
	return r.syncRegistry(ctx, registry)
}

// syncRegistry performs the main sync logic
func (r *MCPRegistryReconciler) syncRegistry(ctx context.Context, registry *mcpv1alpha1.MCPRegistry) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("registry", registry.Name)
	
	// Validate source configuration
	if err := r.validateSource(registry); err != nil {
		logger.Error(err, "Source validation failed")
		if updateErr := r.statusManager.UpdateSyncStatus(ctx, registry, nil, err); updateErr != nil {
			logger.Error(updateErr, "Failed to update status after validation error")
		}
		return ctrl.Result{RequeueAfter: r.getRetryDelay(registry)}, nil
	}

	// Check if sync is needed
	if !r.shouldSync(ctx, registry) {
		logger.V(1).Info("Sync not needed, skipping")
		return r.scheduleNextSync(registry), nil
	}

	// Update phase to syncing
	if err := r.statusManager.UpdatePhase(ctx, registry, mcpv1alpha1.MCPRegistryPhaseSyncing, "Syncing registry data"); err != nil {
		return ctrl.Result{}, err
	}

	// Get source handler
	sourceHandler, exists := r.sourceHandlers[registry.Spec.Source.Type]
	if !exists {
		err := fmt.Errorf("unsupported source type: %s", registry.Spec.Source.Type)
		if updateErr := r.statusManager.UpdateSyncStatus(ctx, registry, nil, err); updateErr != nil {
			logger.Error(updateErr, "Failed to update status after source handler error")
		}
		return ctrl.Result{}, err
	}

	// Perform sync operation
	logger.Info("Starting sync operation", "sourceType", registry.Spec.Source.Type)
	result, err := sourceHandler.Sync(ctx, registry)
	
	if err != nil {
		logger.Error(err, "Sync operation failed")
		if updateErr := r.statusManager.UpdateSyncStatus(ctx, registry, nil, err); updateErr != nil {
			logger.Error(updateErr, "Failed to update status after sync error")
		}
		
		// Determine if we should retry
		if r.statusManager.ShouldRetry(registry) {
			retryDelay := r.statusManager.GetRetryDelay(registry)
			logger.Info("Scheduling retry", "delay", retryDelay, "attempts", registry.Status.SyncAttempts)
			return ctrl.Result{RequeueAfter: retryDelay}, nil
		}
		
		return ctrl.Result{RequeueAfter: r.getRetryDelay(registry)}, nil
	}

	// Check if data has changed
	if result.Hash == registry.Status.LastSyncHash && registry.Status.Phase == mcpv1alpha1.MCPRegistryPhaseReady {
		logger.V(1).Info("Data unchanged, skipping storage update")
		return r.scheduleNextSync(registry), nil
	}

	// Store the data
	if err := r.storageManager.Store(ctx, registry, result.Data); err != nil {
		logger.Error(err, "Failed to store registry data")
		if updateErr := r.statusManager.UpdateSyncStatus(ctx, registry, nil, err); updateErr != nil {
			logger.Error(updateErr, "Failed to update status after storage error")
		}
		return ctrl.Result{RequeueAfter: r.getRetryDelay(registry)}, nil
	}

	// Update storage reference
	storageRef := r.storageManager.GetStorageReference(registry)
	if err := r.statusManager.UpdateStorageReference(ctx, registry, storageRef); err != nil {
		logger.Error(err, "Failed to update storage reference")
		return ctrl.Result{}, err
	}

	// Update sync status
	if err := r.statusManager.UpdateSyncStatus(ctx, registry, result, nil); err != nil {
		logger.Error(err, "Failed to update sync status")
		return ctrl.Result{}, err
	}

	logger.Info("Sync operation completed successfully", "servers", result.ServerCount, "hash", result.Hash[:8])
	
	return r.scheduleNextSync(registry), nil
}

// validateSource validates the source configuration
func (r *MCPRegistryReconciler) validateSource(registry *mcpv1alpha1.MCPRegistry) error {
	sourceHandler, exists := r.sourceHandlers[registry.Spec.Source.Type]
	if !exists {
		return fmt.Errorf("unsupported source type: %s", registry.Spec.Source.Type)
	}
	
	return sourceHandler.Validate(&registry.Spec.Source)
}

// shouldSync determines if a sync operation should be performed
func (r *MCPRegistryReconciler) shouldSync(ctx context.Context, registry *mcpv1alpha1.MCPRegistry) bool {
	logger := log.FromContext(ctx).WithValues("registry", registry.Name)
	
	// Force sync if phase is Pending or Failed
	if registry.Status.Phase == mcpv1alpha1.MCPRegistryPhasePending || registry.Status.Phase == mcpv1alpha1.MCPRegistryPhaseFailed {
		logger.V(1).Info("Force sync due to phase", "phase", registry.Status.Phase)
		return true
	}

	// Check for manual sync trigger
	if r.hasManualSyncTrigger(registry) {
		logger.Info("Manual sync triggered")
		return true
	}

	// Check automatic sync policy
	if registry.Spec.SyncPolicy != nil && registry.Spec.SyncPolicy.Type == mcpv1alpha1.SyncPolicyAutomatic {
		return r.shouldAutoSync(registry)
	}

	return false
}

// hasManualSyncTrigger checks if a manual sync has been triggered
func (r *MCPRegistryReconciler) hasManualSyncTrigger(registry *mcpv1alpha1.MCPRegistry) bool {
	if annotation, exists := registry.GetAnnotations()[AnnotationSyncTrigger]; exists {
		// Simple presence check - in a real implementation, you might want to 
		// track the last processed trigger value
		_ = annotation
		return true
	}
	return false
}

// shouldAutoSync determines if automatic sync should be performed
func (r *MCPRegistryReconciler) shouldAutoSync(registry *mcpv1alpha1.MCPRegistry) bool {
	if registry.Status.LastSyncTime == nil {
		return true
	}

	interval := r.getSyncInterval(registry)
	nextSync := registry.Status.LastSyncTime.Add(interval)
	
	return time.Now().After(nextSync)
}

// getSyncInterval gets the sync interval from the registry spec
func (r *MCPRegistryReconciler) getSyncInterval(registry *mcpv1alpha1.MCPRegistry) time.Duration {
	if registry.Spec.SyncPolicy == nil || registry.Spec.SyncPolicy.Interval == "" {
		return DefaultSyncInterval
	}
	
	if duration, err := time.ParseDuration(registry.Spec.SyncPolicy.Interval); err == nil {
		return duration
	}
	
	return DefaultSyncInterval
}

// getRetryDelay gets the retry delay for failed operations
func (r *MCPRegistryReconciler) getRetryDelay(registry *mcpv1alpha1.MCPRegistry) time.Duration {
	return r.statusManager.GetRetryDelay(registry)
}

// scheduleNextSync schedules the next automatic sync
func (r *MCPRegistryReconciler) scheduleNextSync(registry *mcpv1alpha1.MCPRegistry) ctrl.Result {
	if registry.Spec.SyncPolicy == nil || registry.Spec.SyncPolicy.Type != mcpv1alpha1.SyncPolicyAutomatic {
		return ctrl.Result{} // No automatic sync
	}
	
	interval := r.getSyncInterval(registry)
	return ctrl.Result{RequeueAfter: interval}
}

// handleDeletion handles the deletion of an MCPRegistry
func (r *MCPRegistryReconciler) handleDeletion(ctx context.Context, registry *mcpv1alpha1.MCPRegistry) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("registry", registry.Name)
	
	if controllerutil.ContainsFinalizer(registry, MCPRegistryFinalizer) {
		// Cleanup storage
		logger.Info("Cleaning up registry storage")
		if err := r.storageManager.Delete(ctx, registry); err != nil {
			logger.Error(err, "Failed to cleanup storage")
			return ctrl.Result{}, err
		}
		
		// Remove finalizer
		controllerutil.RemoveFinalizer(registry, MCPRegistryFinalizer)
		if err := r.Update(ctx, registry); err != nil {
			return ctrl.Result{}, err
		}
		
		logger.Info("Registry cleanup completed")
	}
	
	return ctrl.Result{}, nil
}

// findMCPRegistriesForConfigMap finds MCPRegistries that reference a given ConfigMap
func (r *MCPRegistryReconciler) findMCPRegistriesForConfigMap(ctx context.Context, configMap client.Object) []reconcile.Request {
	
	registryList := &mcpv1alpha1.MCPRegistryList{}
	if err := r.List(ctx, registryList); err != nil {
		return nil
	}
	
	var requests []reconcile.Request
	
	for _, registry := range registryList.Items {
		if registry.Spec.Source.Type == mcpv1alpha1.RegistrySourceTypeConfigMap &&
		   registry.Spec.Source.ConfigMap != nil {
			
			// Check if this registry references the changed ConfigMap
			cmSource := registry.Spec.Source.ConfigMap
			namespace := cmSource.Namespace
			if namespace == "" {
				namespace = registry.Namespace
			}
			
			if cmSource.Name == configMap.GetName() && namespace == configMap.GetNamespace() {
				requests = append(requests, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      registry.Name,
						Namespace: registry.Namespace,
					},
				})
			}
		}
	}
	
	return requests
}

// SetupWithManager sets up the controller with the Manager
func (r *MCPRegistryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcpv1alpha1.MCPRegistry{}).
		Owns(&corev1.ConfigMap{}). // Watch owned ConfigMaps (storage)
		Watches(
			&corev1.ConfigMap{}, // Watch referenced ConfigMaps (sources)
			handler.EnqueueRequestsFromMapFunc(r.findMCPRegistriesForConfigMap),
		).
		Complete(r)
}