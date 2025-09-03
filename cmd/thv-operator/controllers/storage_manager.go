package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mcpv1alpha1 "github.com/stacklok/toolhive/cmd/thv-operator/api/v1alpha1"
)

const (
	// StorageConfigMapKeyData is the key used to store registry data in ConfigMaps
	StorageConfigMapKeyData = "registry.json"
	// StorageConfigMapKeyMetadata is the key used to store metadata in ConfigMaps
	StorageConfigMapKeyMetadata = "metadata.json"
)

// StorageManager defines the interface for storing and retrieving registry data
type StorageManager interface {
	// Store saves registry data to persistent storage
	Store(ctx context.Context, registry *mcpv1alpha1.MCPRegistry, data []byte) error
	
	// Get retrieves registry data from persistent storage
	Get(ctx context.Context, registry *mcpv1alpha1.MCPRegistry) ([]byte, error)
	
	// Delete removes registry data from persistent storage
	Delete(ctx context.Context, registry *mcpv1alpha1.MCPRegistry) error
	
	// GetStorageReference returns a reference to where the data is stored
	GetStorageReference(registry *mcpv1alpha1.MCPRegistry) *mcpv1alpha1.StorageReference
}

// ConfigMapStorageManager implements StorageManager using Kubernetes ConfigMaps
type ConfigMapStorageManager struct {
	client client.Client
	scheme *metav1.GroupVersionKind
}

// NewConfigMapStorageManager creates a new ConfigMap-based storage manager
func NewConfigMapStorageManager(client client.Client) *ConfigMapStorageManager {
	return &ConfigMapStorageManager{
		client: client,
	}
}

// Store saves registry data to a ConfigMap
func (s *ConfigMapStorageManager) Store(ctx context.Context, registry *mcpv1alpha1.MCPRegistry, data []byte) error {
	logger := log.FromContext(ctx).WithValues("registry", registry.Name, "storage", "configmap")
	
	configMapName := s.getStorageConfigMapName(registry)
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: registry.Namespace,
			Labels:    s.getStorageLabels(registry),
			Annotations: map[string]string{
				"toolhive.stacklok.dev/registry-name":   registry.Name,
				"toolhive.stacklok.dev/registry-format": registry.Spec.Format,
				"toolhive.stacklok.dev/storage-type":    "registry-data",
			},
		},
		Data: map[string]string{
			StorageConfigMapKeyData: string(data),
		},
	}

	// Set owner reference so the ConfigMap is cleaned up when the registry is deleted
	if err := controllerutil.SetControllerReference(registry, configMap, s.client.Scheme()); err != nil {
		return fmt.Errorf("failed to set controller reference: %w", err)
	}

	// Check if ConfigMap already exists
	existing := &corev1.ConfigMap{}
	err := s.client.Get(ctx, types.NamespacedName{
		Name:      configMapName,
		Namespace: registry.Namespace,
	}, existing)

	if err != nil && errors.IsNotFound(err) {
		// Create new ConfigMap
		logger.Info("Creating storage ConfigMap", "configmap", configMapName)
		if err := s.client.Create(ctx, configMap); err != nil {
			return fmt.Errorf("failed to create storage ConfigMap: %w", err)
		}
		logger.Info("Successfully created storage ConfigMap")
	} else if err != nil {
		return fmt.Errorf("failed to get existing ConfigMap: %w", err)
	} else {
		// Update existing ConfigMap
		logger.Info("Updating storage ConfigMap", "configmap", configMapName)
		existing.Data = configMap.Data
		existing.Labels = configMap.Labels
		existing.Annotations = configMap.Annotations
		
		if err := s.client.Update(ctx, existing); err != nil {
			return fmt.Errorf("failed to update storage ConfigMap: %w", err)
		}
		logger.Info("Successfully updated storage ConfigMap")
	}

	return nil
}

// Get retrieves registry data from a ConfigMap
func (s *ConfigMapStorageManager) Get(ctx context.Context, registry *mcpv1alpha1.MCPRegistry) ([]byte, error) {
	logger := log.FromContext(ctx).WithValues("registry", registry.Name, "storage", "configmap")
	
	configMapName := s.getStorageConfigMapName(registry)
	configMap := &corev1.ConfigMap{}
	
	err := s.client.Get(ctx, types.NamespacedName{
		Name:      configMapName,
		Namespace: registry.Namespace,
	}, configMap)
	
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Storage ConfigMap not found", "configmap", configMapName)
			return nil, fmt.Errorf("storage ConfigMap not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get storage ConfigMap: %w", err)
	}

	data, exists := configMap.Data[StorageConfigMapKeyData]
	if !exists {
		return nil, fmt.Errorf("data key '%s' not found in storage ConfigMap", StorageConfigMapKeyData)
	}

	logger.Info("Successfully retrieved data from storage ConfigMap")
	return []byte(data), nil
}

// Delete removes the storage ConfigMap
func (s *ConfigMapStorageManager) Delete(ctx context.Context, registry *mcpv1alpha1.MCPRegistry) error {
	logger := log.FromContext(ctx).WithValues("registry", registry.Name, "storage", "configmap")
	
	configMapName := s.getStorageConfigMapName(registry)
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: registry.Namespace,
		},
	}

	logger.Info("Deleting storage ConfigMap", "configmap", configMapName)
	if err := s.client.Delete(ctx, configMap); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Storage ConfigMap already deleted", "configmap", configMapName)
			return nil
		}
		return fmt.Errorf("failed to delete storage ConfigMap: %w", err)
	}

	logger.Info("Successfully deleted storage ConfigMap")
	return nil
}

// GetStorageReference returns a reference to the storage location
func (s *ConfigMapStorageManager) GetStorageReference(registry *mcpv1alpha1.MCPRegistry) *mcpv1alpha1.StorageReference {
	configMapName := s.getStorageConfigMapName(registry)
	
	return &mcpv1alpha1.StorageReference{
		Type: "configmap",
		ConfigMapRef: &mcpv1alpha1.ConfigMapReference{
			Name:      configMapName,
			Namespace: registry.Namespace,
			Key:       StorageConfigMapKeyData,
		},
	}
}

// getStorageConfigMapName generates the name for the storage ConfigMap
func (s *ConfigMapStorageManager) getStorageConfigMapName(registry *mcpv1alpha1.MCPRegistry) string {
	return fmt.Sprintf("%s-registry-storage", registry.Name)
}

// getStorageLabels returns labels for the storage ConfigMap
func (s *ConfigMapStorageManager) getStorageLabels(registry *mcpv1alpha1.MCPRegistry) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "toolhive-registry",
		"app.kubernetes.io/component":  "storage",
		"app.kubernetes.io/part-of":    "toolhive",
		"app.kubernetes.io/managed-by": "toolhive-operator",
		"toolhive.stacklok.dev/registry": registry.Name,
		"toolhive.stacklok.dev/type":     "registry-storage",
	}
}

// StorageError represents storage-related errors
type StorageError struct {
	Operation string
	Registry  string
	Reason    string
	Err       error
}

func (e *StorageError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("storage %s failed for registry '%s': %s: %v", e.Operation, e.Registry, e.Reason, e.Err)
	}
	return fmt.Sprintf("storage %s failed for registry '%s': %s", e.Operation, e.Registry, e.Reason)
}

func (e *StorageError) Unwrap() error {
	return e.Err
}

// NewStorageError creates a new storage error
func NewStorageError(operation, registry, reason string, err error) *StorageError {
	return &StorageError{
		Operation: operation,
		Registry:  registry,
		Reason:    reason,
		Err:       err,
	}
}