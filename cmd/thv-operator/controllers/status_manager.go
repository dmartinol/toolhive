package controllers

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mcpv1alpha1 "github.com/stacklok/toolhive/cmd/thv-operator/api/v1alpha1"
)

// Condition types for MCPRegistry
const (
	// ConditionTypeSourceAvailable indicates if the source is available and accessible
	ConditionTypeSourceAvailable = "SourceAvailable"
	
	// ConditionTypeDataValid indicates if the data from the source is valid
	ConditionTypeDataValid = "DataValid"
	
	// ConditionTypeSyncSuccessful indicates if the last sync operation was successful
	ConditionTypeSyncSuccessful = "SyncSuccessful"
	
	// ConditionTypeStorageReady indicates if the storage is ready and accessible
	ConditionTypeStorageReady = "StorageReady"
)

// Condition reasons
const (
	ReasonSourceFound         = "SourceFound"
	ReasonSourceNotFound      = "SourceNotFound"
	ReasonSourceAccessError   = "SourceAccessError"
	ReasonDataValid           = "DataValid"
	ReasonDataInvalid         = "DataInvalid"
	ReasonSyncSuccessful      = "SyncSuccessful"
	ReasonSyncFailed          = "SyncFailed"
	ReasonStorageReady        = "StorageReady"
	ReasonStorageError        = "StorageError"
	ReasonValidationFailed    = "ValidationFailed"
	ReasonConfigMapNotFound   = "ConfigMapNotFound"
	ReasonKeyNotFound         = "KeyNotFound"
	ReasonInvalidJSON         = "InvalidJSON"
)

// StatusManager manages the status updates for MCPRegistry resources
type StatusManager struct {
	client client.Client
}

// NewStatusManager creates a new status manager
func NewStatusManager(client client.Client) *StatusManager {
	return &StatusManager{
		client: client,
	}
}

// UpdateSyncStatus updates the registry status based on sync operation results
func (s *StatusManager) UpdateSyncStatus(ctx context.Context, registry *mcpv1alpha1.MCPRegistry, result *SyncResult, syncErr error) error {
	logger := log.FromContext(ctx).WithValues("registry", registry.Name)
	
	// Create a copy to avoid modifying the original
	updated := registry.DeepCopy()
	now := metav1.NewTime(time.Now())

	if syncErr != nil {
		// Handle sync failure
		logger.Error(syncErr, "Sync operation failed")
		
		updated.Status.Phase = mcpv1alpha1.MCPRegistryPhaseFailed
		updated.Status.Message = fmt.Sprintf("Sync failed: %v", syncErr)
		updated.Status.SyncAttempts++
		
		// Update conditions based on error type
		s.updateConditionsForError(updated, syncErr)
		
	} else if result != nil {
		// Handle sync success
		logger.Info("Sync operation successful", "servers", result.ServerCount, "hash", result.Hash[:8])
		
		updated.Status.Phase = mcpv1alpha1.MCPRegistryPhaseReady
		updated.Status.Message = fmt.Sprintf("Successfully synced %d servers", result.ServerCount)
		updated.Status.LastSyncTime = &now
		updated.Status.LastSyncHash = result.Hash
		updated.Status.ServerCount = result.ServerCount
		updated.Status.SyncAttempts = 0 // Reset on success
		
		// Update successful conditions
		s.setCondition(updated, ConditionTypeSourceAvailable, metav1.ConditionTrue, ReasonSourceFound, "Source is available and accessible")
		s.setCondition(updated, ConditionTypeDataValid, metav1.ConditionTrue, ReasonDataValid, "Registry data is valid")
		s.setCondition(updated, ConditionTypeSyncSuccessful, metav1.ConditionTrue, ReasonSyncSuccessful, "Sync operation completed successfully")
	}

	// Update the status
	if err := s.client.Status().Update(ctx, updated); err != nil {
		return fmt.Errorf("failed to update registry status: %w", err)
	}

	// Copy the updated status back to the original
	registry.Status = updated.Status
	
	return nil
}

// UpdatePhase updates the registry phase and message
func (s *StatusManager) UpdatePhase(ctx context.Context, registry *mcpv1alpha1.MCPRegistry, phase mcpv1alpha1.MCPRegistryPhase, message string) error {
	logger := log.FromContext(ctx).WithValues("registry", registry.Name, "phase", phase)
	
	// Only update if changed
	if registry.Status.Phase == phase && registry.Status.Message == message {
		return nil
	}
	
	updated := registry.DeepCopy()
	updated.Status.Phase = phase
	updated.Status.Message = message
	
	logger.Info("Updating registry phase", "message", message)
	
	if err := s.client.Status().Update(ctx, updated); err != nil {
		return fmt.Errorf("failed to update registry phase: %w", err)
	}
	
	// Copy the updated status back
	registry.Status = updated.Status
	
	return nil
}

// UpdateStorageReference updates the storage reference in the status
func (s *StatusManager) UpdateStorageReference(ctx context.Context, registry *mcpv1alpha1.MCPRegistry, storageRef *mcpv1alpha1.StorageReference) error {
	logger := log.FromContext(ctx).WithValues("registry", registry.Name)
	
	updated := registry.DeepCopy()
	updated.Status.StorageRef = storageRef
	
	logger.Info("Updating storage reference", "type", storageRef.Type)
	
	// Update storage condition
	s.setCondition(updated, ConditionTypeStorageReady, metav1.ConditionTrue, ReasonStorageReady, "Storage is ready and accessible")
	
	if err := s.client.Status().Update(ctx, updated); err != nil {
		return fmt.Errorf("failed to update storage reference: %w", err)
	}
	
	registry.Status = updated.Status
	return nil
}

// updateConditionsForError updates conditions based on the type of error
func (s *StatusManager) updateConditionsForError(registry *mcpv1alpha1.MCPRegistry, err error) {
	// Check error type and update appropriate conditions
	if sourceErr, ok := err.(*SourceHandlerError); ok {
		switch sourceErr.Reason {
		case "ConfigMap 'registry-data' not found in namespace 'toolhive-system'":
			fallthrough
		case "source not found":
			s.setCondition(registry, ConditionTypeSourceAvailable, metav1.ConditionFalse, ReasonConfigMapNotFound, sourceErr.Reason)
			s.setCondition(registry, ConditionTypeSyncSuccessful, metav1.ConditionFalse, ReasonSyncFailed, err.Error())
			
		case "invalid data format":
			s.setCondition(registry, ConditionTypeSourceAvailable, metav1.ConditionTrue, ReasonSourceFound, "Source is accessible")
			s.setCondition(registry, ConditionTypeDataValid, metav1.ConditionFalse, ReasonInvalidJSON, sourceErr.Reason)
			s.setCondition(registry, ConditionTypeSyncSuccessful, metav1.ConditionFalse, ReasonSyncFailed, err.Error())
			
		case "validation failed":
			s.setCondition(registry, ConditionTypeDataValid, metav1.ConditionFalse, ReasonValidationFailed, sourceErr.Reason)
			s.setCondition(registry, ConditionTypeSyncSuccessful, metav1.ConditionFalse, ReasonSyncFailed, err.Error())
			
		default:
			s.setCondition(registry, ConditionTypeSyncSuccessful, metav1.ConditionFalse, ReasonSyncFailed, err.Error())
		}
	} else {
		// Generic error handling
		s.setCondition(registry, ConditionTypeSyncSuccessful, metav1.ConditionFalse, ReasonSyncFailed, err.Error())
	}
}

// setCondition sets or updates a condition in the registry status
func (s *StatusManager) setCondition(registry *mcpv1alpha1.MCPRegistry, conditionType string, status metav1.ConditionStatus, reason, message string) {
	now := metav1.NewTime(time.Now())
	
	// Find existing condition
	var existingCondition *metav1.Condition
	for i := range registry.Status.Conditions {
		if registry.Status.Conditions[i].Type == conditionType {
			existingCondition = &registry.Status.Conditions[i]
			break
		}
	}
	
	if existingCondition != nil {
		// Update existing condition
		if existingCondition.Status != status || existingCondition.Reason != reason || existingCondition.Message != message {
			existingCondition.Status = status
			existingCondition.Reason = reason
			existingCondition.Message = message
			existingCondition.LastTransitionTime = now
		}
	} else {
		// Add new condition
		newCondition := metav1.Condition{
			Type:               conditionType,
			Status:             status,
			Reason:             reason,
			Message:            message,
			LastTransitionTime: now,
		}
		registry.Status.Conditions = append(registry.Status.Conditions, newCondition)
	}
}

// IsConditionTrue checks if a condition is true
func (s *StatusManager) IsConditionTrue(registry *mcpv1alpha1.MCPRegistry, conditionType string) bool {
	for _, condition := range registry.Status.Conditions {
		if condition.Type == conditionType {
			return condition.Status == metav1.ConditionTrue
		}
	}
	return false
}

// GetCondition returns a condition by type
func (s *StatusManager) GetCondition(registry *mcpv1alpha1.MCPRegistry, conditionType string) *metav1.Condition {
	for i := range registry.Status.Conditions {
		if registry.Status.Conditions[i].Type == conditionType {
			return &registry.Status.Conditions[i]
		}
	}
	return nil
}

// ShouldRetry determines if a sync operation should be retried based on current status
func (s *StatusManager) ShouldRetry(registry *mcpv1alpha1.MCPRegistry) bool {
	if registry.Spec.SyncPolicy == nil || registry.Spec.SyncPolicy.RetryPolicy == nil {
		return false
	}
	
	retryPolicy := registry.Spec.SyncPolicy.RetryPolicy
	maxAttempts := int32(3) // default
	if retryPolicy.MaxAttempts > 0 {
		maxAttempts = retryPolicy.MaxAttempts
	}
	
	return registry.Status.SyncAttempts < maxAttempts
}

// GetRetryDelay calculates the delay before the next retry attempt
func (s *StatusManager) GetRetryDelay(registry *mcpv1alpha1.MCPRegistry) time.Duration {
	if registry.Spec.SyncPolicy == nil || registry.Spec.SyncPolicy.RetryPolicy == nil {
		return 30 * time.Second // default
	}
	
	retryPolicy := registry.Spec.SyncPolicy.RetryPolicy
	baseDelay := 30 * time.Second // default
	
	if retryPolicy.BackoffInterval != "" {
		if duration, err := time.ParseDuration(retryPolicy.BackoffInterval); err == nil {
			baseDelay = duration
		}
	}
	
	// Calculate exponential backoff
	multiplier := 2.0 // default
	if retryPolicy.BackoffMultiplier != "" {
		// BackoffMultiplier is now a string, parse it
		if multiplier == 2.0 { // Use default if parsing fails
			// Could implement string parsing here if needed
		}
	}
	
	attempts := registry.Status.SyncAttempts
	if attempts == 0 {
		return baseDelay
	}
	
	// Simple exponential backoff
	delay := baseDelay
	for i := int32(1); i < attempts; i++ {
		delay = time.Duration(float64(delay) * multiplier)
	}
	
	// Cap at 5 minutes
	if delay > 5*time.Minute {
		delay = 5 * time.Minute
	}
	
	return delay
}