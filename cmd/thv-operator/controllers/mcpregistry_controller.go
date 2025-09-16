package controllers

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mcpv1alpha1 "github.com/stacklok/toolhive/cmd/thv-operator/api/v1alpha1"
	"github.com/stacklok/toolhive/cmd/thv-operator/pkg/sources"
	"github.com/stacklok/toolhive/cmd/thv-operator/pkg/sync"
)

// MCPRegistryReconciler reconciles a MCPRegistry object
type MCPRegistryReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	// Sync manager handles all sync operations
	syncManager sync.Manager
}

// NewMCPRegistryReconciler creates a new MCPRegistryReconciler with required dependencies
func NewMCPRegistryReconciler(k8sClient client.Client, scheme *runtime.Scheme) *MCPRegistryReconciler {
	sourceHandlerFactory := sources.NewSourceHandlerFactory(k8sClient)
	storageManager := sources.NewConfigMapStorageManager(k8sClient, scheme)
	syncManager := sync.NewDefaultSyncManager(k8sClient, scheme, sourceHandlerFactory, storageManager)

	return &MCPRegistryReconciler{
		Client:      k8sClient,
		Scheme:      scheme,
		syncManager: syncManager,
	}
}

// +kubebuilder:rbac:groups=toolhive.stacklok.dev,resources=mcpregistries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=toolhive.stacklok.dev,resources=mcpregistries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=toolhive.stacklok.dev,resources=mcpregistries/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *MCPRegistryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctxLogger := log.FromContext(ctx)

	// 1. Fetch MCPRegistry instance
	mcpRegistry := &mcpv1alpha1.MCPRegistry{}
	err := r.Get(ctx, req.NamespacedName, mcpRegistry)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			ctxLogger.Info("MCPRegistry resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		ctxLogger.Error(err, "Failed to get MCPRegistry")
		return ctrl.Result{}, err
	}
	ctxLogger.Info("Reconciling MCPRegistry", "MCPRegistry.Name", mcpRegistry.Name, "phase", mcpRegistry.Status.Phase)

	// 2. Handle deletion if DeletionTimestamp is set
	if mcpRegistry.GetDeletionTimestamp() != nil {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(mcpRegistry, "mcpregistry.toolhive.stacklok.dev/finalizer") {
			// Run finalization logic. If the finalization logic fails,
			// don't remove the finalizer so that we can retry during the next reconciliation.
			if err := r.finalizeMCPRegistry(ctx, mcpRegistry); err != nil {
				ctxLogger.Error(err, "Reconciliation completed with error while finalizing MCPRegistry",
					"MCPRegistry.Name", mcpRegistry.Name)
				return ctrl.Result{}, err
			}

			// Remove the finalizer. Once all finalizers have been removed, the object will be deleted.
			controllerutil.RemoveFinalizer(mcpRegistry, "mcpregistry.toolhive.stacklok.dev/finalizer")
			err := r.Update(ctx, mcpRegistry)
			if err != nil {
				ctxLogger.Error(err, "Reconciliation completed with error while removing finalizer",
					"MCPRegistry.Name", mcpRegistry.Name)
				return ctrl.Result{}, err
			}
		}
		ctxLogger.Info("Reconciliation of deleted MCPRegistry completed successfully",
			"MCPRegistry.Name", mcpRegistry.Name,
			"phase", mcpRegistry.Status.Phase)
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(mcpRegistry, "mcpregistry.toolhive.stacklok.dev/finalizer") {
		controllerutil.AddFinalizer(mcpRegistry, "mcpregistry.toolhive.stacklok.dev/finalizer")
		err = r.Update(ctx, mcpRegistry)
		if err != nil {
			ctxLogger.Error(err, "Reconciliation completed with error while adding finalizer",
				"MCPRegistry.Name", mcpRegistry.Name)
			return ctrl.Result{}, err
		}
		ctxLogger.Info("Reconciliation completed successfully after adding finalizer",
			"MCPRegistry.Name", mcpRegistry.Name)
		return ctrl.Result{}, nil
	}

	// 3. Check if sync is needed before performing it
	result, err := r.reconcileSync(ctx, mcpRegistry)

	// Refresh registry to get the actual current phase for logging
	var currentPhase mcpv1alpha1.MCPRegistryPhase
	if refreshErr := r.Get(ctx, client.ObjectKeyFromObject(mcpRegistry), mcpRegistry); refreshErr != nil {
		ctxLogger.V(1).Error(refreshErr, "Failed to refresh registry for final logging")
		currentPhase = "Unknown"
	} else {
		currentPhase = mcpRegistry.Status.Phase
	}

	// Log reconciliation completion with actual outcome
	if err != nil {
		ctxLogger.Error(err, "Reconciliation completed with error",
			"MCPRegistry.Name", mcpRegistry.Name,
			"phase", currentPhase)
	} else if currentPhase == mcpv1alpha1.MCPRegistryPhaseFailed {
		ctxLogger.Info("Reconciliation completed - sync failed, retry scheduled",
			"MCPRegistry.Name", mcpRegistry.Name,
			"phase", currentPhase,
			"requeueAfter", result.RequeueAfter)
	} else {
		ctxLogger.Info("Reconciliation completed successfully",
			"MCPRegistry.Name", mcpRegistry.Name,
			"phase", currentPhase,
			"requeueAfter", result.RequeueAfter)
	}

	return result, err
}

// reconcileSync checks if sync is needed and performs it if necessary
func (r *MCPRegistryReconciler) reconcileSync(ctx context.Context, mcpRegistry *mcpv1alpha1.MCPRegistry) (ctrl.Result, error) {
	ctxLogger := log.FromContext(ctx)

	// Refresh the object to get latest status for accurate timing calculations
	if err := r.Get(ctx, client.ObjectKeyFromObject(mcpRegistry), mcpRegistry); err != nil {
		ctxLogger.Error(err, "Failed to refresh MCPRegistry object for sync check")
		return ctrl.Result{}, err
	}

	// Check if sync is needed
	syncNeeded, syncReason, nextSyncTime, err := r.syncManager.ShouldSync(ctx, mcpRegistry)
	if err != nil {
		ctxLogger.Error(err, "Failed to determine if sync is needed")
		// Proceed with sync on error to be safe
		syncNeeded = true
		syncReason = sync.ReasonErrorCheckingSyncNeed
	}

	if !syncNeeded {
		ctxLogger.Info("Sync not needed", "reason", syncReason)
		// Schedule next reconciliation if we have a sync policy
		if nextSyncTime != nil {
			requeueAfter := time.Until(*nextSyncTime)
			ctxLogger.Info("Scheduling next automatic sync", "requeueAfter", requeueAfter)
			return ctrl.Result{RequeueAfter: requeueAfter}, nil
		}
		return ctrl.Result{}, nil
	}

	ctxLogger.Info("Sync needed", "reason", syncReason)

	// Handle manual sync with no data changes - update trigger tracking only
	if syncReason == sync.ReasonManualNoChanges {
		return r.syncManager.UpdateManualSyncTriggerOnly(ctx, mcpRegistry)
	}

	result, err := r.syncManager.PerformSync(ctx, mcpRegistry)

	if err != nil {
		// Sync failed - schedule retry with exponential backoff
		ctxLogger.Error(err, "Sync failed, scheduling retry")
		// Use a shorter retry interval instead of the full sync interval
		retryAfter := time.Minute * 5 // Default retry interval
		if result.RequeueAfter > 0 {
			// If PerformSync already set a retry interval, use it
			retryAfter = result.RequeueAfter
		}
		return ctrl.Result{RequeueAfter: retryAfter}, err
	}

	// Refresh object to check if sync actually succeeded or if failure was recorded
	// Add retry logic to handle potential caching delays
	var refreshedRegistry *mcpv1alpha1.MCPRegistry
	for i := 0; i < 3; i++ {
		freshRegistry := &mcpv1alpha1.MCPRegistry{}
		if err := r.Get(ctx, client.ObjectKeyFromObject(mcpRegistry), freshRegistry); err != nil {
			ctxLogger.Error(err, "Failed to refresh MCPRegistry after sync", "attempt", i+1)
			return ctrl.Result{}, err
		}

		// If this is a retry and we see an updated syncAttempts, we got the fresh object
		if i > 0 && freshRegistry.Status.SyncAttempts > mcpRegistry.Status.SyncAttempts {
			ctxLogger.Info("Got fresh object after retry", "attempt", i+1, "syncAttempts", freshRegistry.Status.SyncAttempts)
			refreshedRegistry = freshRegistry
			break
		}

		// On first attempt, or if syncAttempts hasn't changed, use what we got
		refreshedRegistry = freshRegistry
		if i < 2 {
			// Brief pause to let cache update (except on final attempt)
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Log the actual phase we see for debugging
	ctxLogger.V(1).Info(
		"Post-sync phase check",
		"phase",
		refreshedRegistry.Status.Phase,
		"message", refreshedRegistry.Status.Message,
		"syncAttempts",
		refreshedRegistry.Status.SyncAttempts)

	// Check if sync actually succeeded by looking at the phase
	if refreshedRegistry.Status.Phase == mcpv1alpha1.MCPRegistryPhaseFailed {
		ctxLogger.Info("Sync failed but failure was properly recorded, scheduling retry")
		// Sync failed but failure was properly recorded - use sync manager's retry interval
		retryAfter := time.Minute * 5 // Default fallback
		if result.RequeueAfter > 0 {
			retryAfter = result.RequeueAfter
		}
		return ctrl.Result{RequeueAfter: retryAfter}, nil
	}

	// Sync actually succeeded - schedule next automatic sync only if this was an automatic sync (not manual)
	if mcpRegistry.Spec.SyncPolicy != nil && !sync.IsManualSync(syncReason) {
		interval, parseErr := time.ParseDuration(mcpRegistry.Spec.SyncPolicy.Interval)
		if parseErr == nil {
			result.RequeueAfter = interval
			ctxLogger.Info("Automatic sync successful, scheduled next automatic sync", "interval", interval)
		} else {
			ctxLogger.Error(parseErr, "Invalid sync interval in policy", "interval", mcpRegistry.Spec.SyncPolicy.Interval)
		}
	} else if sync.IsManualSync(syncReason) {
		ctxLogger.Info("Manual sync successful, automatic sync schedule unchanged")
	} else {
		ctxLogger.Info("Sync successful, no automatic sync policy configured")
	}

	return result, err
}

// finalizeMCPRegistry performs the finalizer logic for the MCPRegistry
func (r *MCPRegistryReconciler) finalizeMCPRegistry(ctx context.Context, registry *mcpv1alpha1.MCPRegistry) error {
	ctxLogger := log.FromContext(ctx)

	// Update the MCPRegistry status to indicate termination
	registry.Status.Phase = mcpv1alpha1.MCPRegistryPhaseTerminating
	registry.Status.Message = "MCPRegistry is being terminated"
	if err := r.Status().Update(ctx, registry); err != nil {
		ctxLogger.Error(err, "Failed to update MCPRegistry status during finalization")
		return err
	}

	// Clean up internal storage ConfigMaps
	if err := r.syncManager.Delete(ctx, registry); err != nil {
		ctxLogger.Error(err, "Failed to delete storage during finalization")
		// Continue with finalization even if storage cleanup fails
	}

	// TODO: Add additional cleanup logic when other features are implemented:
	// - Clean up Registry API service
	// - Cancel any running sync operations

	ctxLogger.Info("MCPRegistry finalization completed", "registry", registry.Name)
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MCPRegistryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcpv1alpha1.MCPRegistry{}).
		Complete(r)
}
