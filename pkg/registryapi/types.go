package registryapi

import (
	"time"

	mcpv1alpha1 "github.com/stacklok/toolhive/cmd/thv-operator/api/v1alpha1"
)

// RegistryInfo represents the response for /api/v1/registry/info
type RegistryInfo struct {
	Name        string                            `json:"name"`
	DisplayName string                            `json:"displayName,omitempty"`
	Format      string                            `json:"format"`
	Source      *RegistrySourceInfo               `json:"source"`
	Status      *RegistryStatusInfo               `json:"status"`
	SyncPolicy  *mcpv1alpha1.MCPRegistrySyncPolicy `json:"syncPolicy,omitempty"`
}

// RegistrySourceInfo represents source information
type RegistrySourceInfo struct {
	Type   string `json:"type"`
	Format string `json:"format"`
	// Additional source-specific fields can be added here if needed
}

// RegistryStatusInfo represents status information
type RegistryStatusInfo struct {
	Phase        string     `json:"phase"`
	ServerCount  int32      `json:"serverCount"`
	LastSyncTime *time.Time `json:"lastSyncTime,omitempty"`
	LastSyncHash string     `json:"lastSyncHash,omitempty"`
	Message      string     `json:"message,omitempty"`
}

// ServerListResponse represents the response for /api/v1/registry/servers
type ServerListResponse struct {
	Servers map[string]interface{} `json:"servers"`
	Count   int                    `json:"count"`
	Format  string                 `json:"format"`
}

// ServerResponse represents the response for /api/v1/registry/servers/{name}
type ServerResponse struct {
	Name   string      `json:"name"`
	Server interface{} `json:"server"`
	Format string      `json:"format"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}