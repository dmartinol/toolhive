package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mcpv1alpha1 "github.com/stacklok/toolhive/cmd/thv-operator/api/v1alpha1"
)

const (
	// ConfigMapSourceType is the identifier for ConfigMap sources
	ConfigMapSourceType = "configmap"
	// DefaultRegistryKey is the default key used in ConfigMaps for registry data
	DefaultRegistryKey = "registry.json"
)

// ConfigMapSourceHandler handles ConfigMap-based registry sources
type ConfigMapSourceHandler struct {
	client client.Client
}

// NewConfigMapSourceHandler creates a new ConfigMap source handler
func NewConfigMapSourceHandler(client client.Client) *ConfigMapSourceHandler {
	return &ConfigMapSourceHandler{
		client: client,
	}
}

// GetSourceType returns the source type this handler supports
func (h *ConfigMapSourceHandler) GetSourceType() string {
	return ConfigMapSourceType
}

// Validate validates the ConfigMap source configuration
func (h *ConfigMapSourceHandler) Validate(source *mcpv1alpha1.MCPRegistrySource) error {
	if source.Type != ConfigMapSourceType {
		return NewSourceHandlerError(ConfigMapSourceType, "validate", 
			"source type mismatch", fmt.Errorf("expected %s, got %s", ConfigMapSourceType, source.Type))
	}

	if source.ConfigMap == nil {
		return NewSourceHandlerError(ConfigMapSourceType, "validate", 
			"configmap configuration is required", ErrValidationFailed)
	}

	if source.ConfigMap.Name == "" {
		return NewSourceHandlerError(ConfigMapSourceType, "validate", 
			"configmap name is required", ErrValidationFailed)
	}

	// Set defaults
	if source.ConfigMap.Key == "" {
		source.ConfigMap.Key = DefaultRegistryKey
	}

	return nil
}

// Sync retrieves and processes data from a ConfigMap source
func (h *ConfigMapSourceHandler) Sync(ctx context.Context, registry *mcpv1alpha1.MCPRegistry) (*SyncResult, error) {
	logger := log.FromContext(ctx).WithValues("registry", registry.Name, "source", "configmap")
	
	// Validate source configuration
	if err := h.Validate(&registry.Spec.Source); err != nil {
		return nil, err
	}

	cmSource := registry.Spec.Source.ConfigMap
	
	// Determine namespace for ConfigMap lookup
	namespace := cmSource.Namespace
	if namespace == "" {
		namespace = registry.Namespace
	}

	// Fetch the ConfigMap
	configMap := &corev1.ConfigMap{}
	configMapKey := types.NamespacedName{
		Name:      cmSource.Name,
		Namespace: namespace,
	}

	logger.Info("Fetching ConfigMap", "configmap", configMapKey.String(), "key", cmSource.Key)
	
	if err := h.client.Get(ctx, configMapKey, configMap); err != nil {
		if errors.IsNotFound(err) {
			return nil, NewSourceHandlerError(ConfigMapSourceType, "sync", 
				fmt.Sprintf("ConfigMap '%s' not found in namespace '%s'", cmSource.Name, namespace), 
				ErrSourceNotFound)
		}
		return nil, NewSourceHandlerError(ConfigMapSourceType, "sync", 
			"failed to fetch ConfigMap", err)
	}

	// Extract data from the specified key
	data, exists := configMap.Data[cmSource.Key]
	if !exists {
		return nil, NewSourceHandlerError(ConfigMapSourceType, "sync", 
			fmt.Sprintf("key '%s' not found in ConfigMap '%s'", cmSource.Key, cmSource.Name), 
			ErrSourceNotFound)
	}

	// Validate JSON format
	var registryData RegistryData
	if err := json.Unmarshal([]byte(data), &registryData); err != nil {
		return nil, NewSourceHandlerError(ConfigMapSourceType, "sync", 
			fmt.Sprintf("invalid JSON in ConfigMap key '%s'", cmSource.Key), 
			ErrInvalidData)
	}

	// Calculate hash for change detection
	hash := h.calculateHash([]byte(data))

	// Get last modified time from ConfigMap
	var lastModified time.Time
	if configMap.CreationTimestamp.Time.After(lastModified) {
		lastModified = configMap.CreationTimestamp.Time
	}
	
	// Check for metadata managed fields to get a more accurate last modified time
	for _, managedField := range configMap.ManagedFields {
		if managedField.Time != nil && managedField.Time.Time.After(lastModified) {
			lastModified = managedField.Time.Time
		}
	}

	// Count servers
	serverCount := int32(len(registryData.Servers))

	// Apply format conversion if needed
	targetFormat := registry.Spec.Format
	if targetFormat == "" {
		targetFormat = mcpv1alpha1.RegistryFormatToolHive // Default format
	}
	
	// Detect source format and convert if necessary
	converter := NewRegistryFormatConverter()
	sourceFormat, err := converter.DetectFormat([]byte(data))
	if err != nil {
		// If detection fails, assume it's already in the target format
		sourceFormat = targetFormat
	}
	
	// Convert data if source and target formats differ
	convertedData := data
	if sourceFormat != targetFormat {
		convertedBytes, err := converter.Convert([]byte(data), sourceFormat, targetFormat)
		if err != nil {
			return nil, NewSourceHandlerError(ConfigMapSourceType, "conversion", 
				fmt.Sprintf("failed to convert from %s to %s format", sourceFormat, targetFormat), err)
		}
		convertedData = string(convertedBytes)
		logger.Info("Registry data converted", "from", sourceFormat, "to", targetFormat)
	}

	logger.Info("Successfully synced ConfigMap data", 
		"servers", serverCount, 
		"hash", hash[:8], 
		"lastModified", lastModified)

	return &SyncResult{
		Data:         []byte(convertedData),
		Hash:         hash,
		ServerCount:  serverCount,
		LastModified: lastModified,
		Format:       targetFormat,
	}, nil
}

// calculateHash calculates SHA256 hash of the data for change detection
func (h *ConfigMapSourceHandler) calculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// GetConfigMapReference returns the ConfigMap reference for the given registry
func (h *ConfigMapSourceHandler) GetConfigMapReference(registry *mcpv1alpha1.MCPRegistry) (types.NamespacedName, string, error) {
	if err := h.Validate(&registry.Spec.Source); err != nil {
		return types.NamespacedName{}, "", err
	}

	cmSource := registry.Spec.Source.ConfigMap
	namespace := cmSource.Namespace
	if namespace == "" {
		namespace = registry.Namespace
	}

	return types.NamespacedName{
		Name:      cmSource.Name,
		Namespace: namespace,
	}, cmSource.Key, nil
}

// validateRegistryData validates the structure of registry data
func (h *ConfigMapSourceHandler) validateRegistryData(data []byte) error {
	var registryData RegistryData
	if err := json.Unmarshal(data, &registryData); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	// Basic validation
	if registryData.Version == "" {
		return fmt.Errorf("version field is required")
	}

	if registryData.Servers == nil {
		return fmt.Errorf("servers field is required")
	}

	// Validate each server
	for serverName, server := range registryData.Servers {
		if err := h.validateServer(serverName, server); err != nil {
			return fmt.Errorf("invalid server '%s': %w", serverName, err)
		}
	}

	return nil
}

// validateServer validates a single server configuration
func (h *ConfigMapSourceHandler) validateServer(name string, server Server) error {
	if server.Description == "" {
		return fmt.Errorf("description is required")
	}

	if server.Image == "" {
		return fmt.Errorf("image is required")
	}

	if server.Transport == "" {
		return fmt.Errorf("transport is required")
	}

	// Validate transport type
	validTransports := []string{"stdio", "http", "sse", "streamable-http"}
	validTransport := false
	for _, valid := range validTransports {
		if server.Transport == valid {
			validTransport = true
			break
		}
	}
	if !validTransport {
		return fmt.Errorf("invalid transport '%s', must be one of: %v", server.Transport, validTransports)
	}

	return nil
}