// Package controllers contains the reconciliation logic for the MCPRegistry custom resource.
// It handles the creation, update, and deletion of MCP registries in Kubernetes.
package controllers

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mcpv1alpha1 "github.com/stacklok/toolhive/cmd/thv-operator/api/v1alpha1"
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
	err := r.Get(ctx, req.NamespacedName, mcpRegistry)
	if err != nil {
		if errors.IsNotFound(err) {
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
	if !containsString(mcpRegistry.Finalizers, "mcpregistry.toolhive.stacklok.dev/finalizer") {
		controllerutil.AddFinalizer(mcpRegistry, "mcpregistry.toolhive.stacklok.dev/finalizer")
		if err := r.Update(ctx, mcpRegistry); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Update status to Syncing
	if mcpRegistry.Status.Phase != mcpv1alpha1.MCPRegistryPhaseSyncing {
		mcpRegistry.Status.Phase = mcpv1alpha1.MCPRegistryPhaseSyncing
		mcpRegistry.Status.Message = "Syncing registry data"
		if err := r.Status().Update(ctx, mcpRegistry); err != nil {
			logger.Error(err, "Failed to update MCPRegistry status")
			return ctrl.Result{}, err
		}
	}

	// Simulate registry sync (in a real implementation, this would fetch from the registry)
	logger.Info("Syncing MCP registry", "url", mcpRegistry.Spec.URL, "type", mcpRegistry.Spec.Type)

	// Update status to Ready
	now := metav1.Now()
	mcpRegistry.Status.Phase = mcpv1alpha1.MCPRegistryPhaseReady
	mcpRegistry.Status.Message = "Registry synced successfully"
	mcpRegistry.Status.LastSyncTime = &now
	mcpRegistry.Status.ServerCount = 0               // This would be populated from actual registry data
	mcpRegistry.Status.AvailableServers = []string{} // This would be populated from actual registry data

	// Update conditions
	meta.SetStatusCondition(&mcpRegistry.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "Synced",
		Message:            "Registry synced successfully",
		LastTransitionTime: now,
	})

	if err := r.Status().Update(ctx, mcpRegistry); err != nil {
		logger.Error(err, "Failed to update MCPRegistry status")
		return ctrl.Result{}, err
	}

	logger.Info("MCPRegistry reconciled successfully")

	// Requeue after the refresh interval
	refreshInterval, err := time.ParseDuration(mcpRegistry.Spec.RefreshInterval)
	if err != nil {
		// Default to 1 hour if parsing fails
		refreshInterval = time.Hour
	}

	return ctrl.Result{RequeueAfter: refreshInterval}, nil
}

// finalizeMCPRegistry handles the cleanup when an MCPRegistry is being deleted
func (r *MCPRegistryReconciler) finalizeMCPRegistry(ctx context.Context, mcpRegistry *mcpv1alpha1.MCPRegistry) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Perform any cleanup operations here
	logger.Info("Finalizing MCPRegistry", "name", mcpRegistry.Name)

	// Remove finalizer
	controllerutil.RemoveFinalizer(mcpRegistry, "mcpregistry.toolhive.stacklok.dev/finalizer")
	if err := r.Update(ctx, mcpRegistry); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MCPRegistryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcpv1alpha1.MCPRegistry{}).
		Complete(r)
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
