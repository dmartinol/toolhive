package registryapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	mcpv1alpha1 "github.com/stacklok/toolhive/cmd/thv-operator/api/v1alpha1"
	"github.com/stacklok/toolhive/cmd/thv-operator/controllers"
	"github.com/stacklok/toolhive/pkg/registry"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// handleRegistryInfo handles GET /api/v1/registry/info
func (s *Server) handleRegistryInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.FromContext(ctx).WithName("registry-info")

	// Get the MCPRegistry resource
	registry := &mcpv1alpha1.MCPRegistry{}
	err := s.kubeClient.Get(ctx, client.ObjectKey{
		Name:      s.config.RegistryName,
		Namespace: s.config.RegistryNamespace,
	}, registry)
	if err != nil {
		logger.Error(err, "failed to get MCPRegistry")
		s.writeErrorResponse(w, "Registry not found", http.StatusNotFound)
		return
	}

	// Build response
	info := &RegistryInfo{
		Name:        registry.Name,
		DisplayName: registry.Spec.DisplayName,
		Format:      registry.Spec.Source.Format,
		Source: &RegistrySourceInfo{
			Type:   registry.Spec.Source.Type,
			Format: registry.Spec.Source.Format,
		},
		SyncPolicy: registry.Spec.SyncPolicy,
	}

	// Add status if available
	if registry.Status.Phase != "" {
		info.Status = &RegistryStatusInfo{
			Phase:       registry.Status.Phase,
			ServerCount: registry.Status.ServerCount,
			Message:     registry.Status.Message,
		}

		// Add last sync time if available
		if registry.Status.LastSyncTime != nil {
			info.Status.LastSyncTime = &registry.Status.LastSyncTime.Time
		}

		// Add last sync hash if available
		if registry.Status.LastSyncHash != "" {
			info.Status.LastSyncHash = registry.Status.LastSyncHash
		}
	}

	// Set default format if empty
	if info.Format == "" {
		info.Format = mcpv1alpha1.RegistryFormatToolHive
	}
	if info.Source.Format == "" {
		info.Source.Format = mcpv1alpha1.RegistryFormatToolHive
	}

	s.writeJSONResponse(w, info, http.StatusOK)
}

// handleListServers handles GET /api/v1/registry/servers
func (s *Server) handleListServers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.FromContext(ctx).WithName("list-servers")

	// Get format from query parameter
	requestedFormat := r.URL.Query().Get("format")
	if requestedFormat == "" {
		requestedFormat = mcpv1alpha1.RegistryFormatToolHive // Default
	}

	// Validate format
	supportedFormats := s.formatConverter.SupportedFormats()
	if !contains(supportedFormats, requestedFormat) {
		s.writeErrorResponse(w, fmt.Sprintf("Unsupported format: %s. Supported formats: %v", requestedFormat, supportedFormats), http.StatusBadRequest)
		return
	}

	// Get registry data
	registryData, err := s.getRegistryData(ctx)
	if err != nil {
		logger.Error(err, "failed to get registry data")
		s.writeErrorResponse(w, "Failed to retrieve registry data", http.StatusInternalServerError)
		return
	}

	// Convert format if needed
	convertedData, err := s.convertRegistryFormat(registryData, requestedFormat)
	if err != nil {
		logger.Error(err, "failed to convert registry format", "requested", requestedFormat)
		s.writeErrorResponse(w, "Failed to convert registry format", http.StatusInternalServerError)
		return
	}

	// Parse the converted data to extract servers
	var serversData map[string]interface{}
	if err := json.Unmarshal(convertedData, &serversData); err != nil {
		logger.Error(err, "failed to parse converted registry data")
		s.writeErrorResponse(w, "Failed to parse registry data", http.StatusInternalServerError)
		return
	}

	// Extract servers from the data
	servers, ok := serversData["servers"].(map[string]interface{})
	if !ok {
		servers = make(map[string]interface{})
	}

	response := &ServerListResponse{
		Servers: servers,
		Count:   len(servers),
		Format:  requestedFormat,
	}

	s.writeJSONResponse(w, response, http.StatusOK)
}

// handleGetServer handles GET /api/v1/registry/servers/{name}
func (s *Server) handleGetServer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.FromContext(ctx).WithName("get-server")

	// Get server name from URL
	serverName := chi.URLParam(r, "name")
	if serverName == "" {
		s.writeErrorResponse(w, "Server name is required", http.StatusBadRequest)
		return
	}

	// Get format from query parameter
	requestedFormat := r.URL.Query().Get("format")
	if requestedFormat == "" {
		requestedFormat = mcpv1alpha1.RegistryFormatToolHive // Default
	}

	// Validate format
	supportedFormats := s.formatConverter.SupportedFormats()
	if !contains(supportedFormats, requestedFormat) {
		s.writeErrorResponse(w, fmt.Sprintf("Unsupported format: %s. Supported formats: %v", requestedFormat, supportedFormats), http.StatusBadRequest)
		return
	}

	// Get registry data
	registryData, err := s.getRegistryData(ctx)
	if err != nil {
		logger.Error(err, "failed to get registry data")
		s.writeErrorResponse(w, "Failed to retrieve registry data", http.StatusInternalServerError)
		return
	}

	// Convert format if needed
	convertedData, err := s.convertRegistryFormat(registryData, requestedFormat)
	if err != nil {
		logger.Error(err, "failed to convert registry format", "requested", requestedFormat)
		s.writeErrorResponse(w, "Failed to convert registry format", http.StatusInternalServerError)
		return
	}

	// Parse the converted data to extract servers
	var serversData map[string]interface{}
	if err := json.Unmarshal(convertedData, &serversData); err != nil {
		logger.Error(err, "failed to parse converted registry data")
		s.writeErrorResponse(w, "Failed to parse registry data", http.StatusInternalServerError)
		return
	}

	// Extract servers from the data
	servers, ok := serversData["servers"].(map[string]interface{})
	if !ok {
		servers = make(map[string]interface{})
	}

	// Find the specific server
	server, exists := servers[serverName]
	if !exists {
		s.writeErrorResponse(w, fmt.Sprintf("Server '%s' not found", serverName), http.StatusNotFound)
		return
	}

	response := &ServerResponse{
		Name:   serverName,
		Server: server,
		Format: requestedFormat,
	}

	s.writeJSONResponse(w, response, http.StatusOK)
}

// getRegistryData retrieves the raw registry data from the storage ConfigMap
func (s *Server) getRegistryData(ctx context.Context) ([]byte, error) {
	// Get the MCPRegistry to find the storage reference
	registry := &mcpv1alpha1.MCPRegistry{}
	err := s.kubeClient.Get(ctx, client.ObjectKey{
		Name:      s.config.RegistryName,
		Namespace: s.config.RegistryNamespace,
	}, registry)
	if err != nil {
		return nil, fmt.Errorf("failed to get MCPRegistry: %w", err)
	}

	// Get storage reference from status
	if registry.Status.StorageRef == nil {
		return nil, fmt.Errorf("no storage reference found in registry status")
	}

	// Get the storage ConfigMap
	configMap := &corev1.ConfigMap{}
	err = s.kubeClient.Get(ctx, types.NamespacedName{
		Name:      registry.Status.StorageRef.Name,
		Namespace: registry.Status.StorageRef.Namespace,
	}, configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage ConfigMap: %w", err)
	}

	// Extract data
	data, exists := configMap.Data[controllers.StorageConfigMapKeyData]
	if !exists {
		return nil, fmt.Errorf("no registry data found in storage ConfigMap")
	}

	return []byte(data), nil
}

// convertRegistryFormat converts registry data to the requested format
func (s *Server) convertRegistryFormat(data []byte, targetFormat string) ([]byte, error) {
	// Detect current format
	currentFormat, err := s.formatConverter.DetectFormat(data)
	if err != nil {
		// If detection fails, assume it's already in the target format
		currentFormat = targetFormat
	}

	// Convert if needed
	if currentFormat != targetFormat {
		return s.formatConverter.Convert(data, currentFormat, targetFormat)
	}

	return data, nil
}

// writeJSONResponse writes a JSON response
func (s *Server) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If we fail to encode, there's not much we can do at this point
		// since we've already written the status code
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// writeErrorResponse writes an error response
func (s *Server) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	errorResp := &ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
		Code:    statusCode,
	}
	s.writeJSONResponse(w, errorResp, statusCode)
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}