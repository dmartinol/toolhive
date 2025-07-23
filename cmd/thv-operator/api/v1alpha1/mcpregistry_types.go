package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MCPRegistrySpec defines the desired state of MCPRegistry
type MCPRegistrySpec struct {
	// URL is the URL of the MCP registry
	// +kubebuilder:validation:Required
	URL string `json:"url"`

	// Type is the type of registry (e.g., "http", "file", "embedded")
	// +kubebuilder:validation:Enum=http;file;embedded
	// +kubebuilder:default=http
	Type string `json:"type,omitempty"`

	// Authentication defines authentication configuration for the registry
	// +optional
	Authentication *RegistryAuthentication `json:"authentication,omitempty"`

	// RefreshInterval is the interval to refresh the registry data
	// +kubebuilder:default="1h"
	// +optional
	RefreshInterval string `json:"refreshInterval,omitempty"`

	// Timeout is the timeout for registry operations
	// +kubebuilder:default="30s"
	// +optional
	Timeout string `json:"timeout,omitempty"`

	// InsecureSkipVerify allows skipping TLS verification for HTTPS registries
	// +kubebuilder:default=false
	// +optional
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
}

// RegistryAuthentication defines authentication for registry access
type RegistryAuthentication struct {
	// Type is the type of authentication
	// +kubebuilder:validation:Enum=basic;bearer;none
	// +kubebuilder:default=none
	Type string `json:"type"`

	// Username is the username for basic authentication
	// +optional
	Username string `json:"username,omitempty"`

	// PasswordSecretRef references a secret containing the password
	// +optional
	PasswordSecretRef *SecretRef `json:"passwordSecretRef,omitempty"`

	// TokenSecretRef references a secret containing the bearer token
	// +optional
	TokenSecretRef *SecretRef `json:"tokenSecretRef,omitempty"`
}

// MCPRegistryStatus defines the observed state of MCPRegistry
type MCPRegistryStatus struct {
	// Conditions represent the latest available observations of the MCPRegistry's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Phase is the current phase of the MCPRegistry
	// +optional
	Phase MCPRegistryPhase `json:"phase,omitempty"`

	// Message provides additional information about the current phase
	// +optional
	Message string `json:"message,omitempty"`

	// LastSyncTime is the timestamp of the last successful sync
	// +optional
	LastSyncTime *metav1.Time `json:"lastSyncTime,omitempty"`

	// ServerCount is the number of MCP servers available in this registry
	// +optional
	ServerCount int32 `json:"serverCount,omitempty"`

	// AvailableServers is a list of available MCP server names
	// +optional
	AvailableServers []string `json:"availableServers,omitempty"`
}

// MCPRegistryPhase is the phase of the MCPRegistry
// +kubebuilder:validation:Enum=Pending;Ready;Failed;Syncing
type MCPRegistryPhase string

const (
	// MCPRegistryPhasePending means the MCPRegistry is being created
	MCPRegistryPhasePending MCPRegistryPhase = "Pending"

	// MCPRegistryPhaseReady means the MCPRegistry is ready and synced
	MCPRegistryPhaseReady MCPRegistryPhase = "Ready"

	// MCPRegistryPhaseFailed means the MCPRegistry failed to sync
	MCPRegistryPhaseFailed MCPRegistryPhase = "Failed"

	// MCPRegistryPhaseSyncing means the MCPRegistry is currently syncing
	MCPRegistryPhaseSyncing MCPRegistryPhase = "Syncing"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
//+kubebuilder:printcolumn:name="URL",type="string",JSONPath=".spec.url"
//+kubebuilder:printcolumn:name="Servers",type="integer",JSONPath=".status.serverCount"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// MCPRegistry is the Schema for the mcpregistries API
type MCPRegistry struct {
	metav1.TypeMeta   `json:",inline"` // nolint:revive
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MCPRegistrySpec   `json:"spec,omitempty"`
	Status MCPRegistryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MCPRegistryList contains a list of MCPRegistry
type MCPRegistryList struct {
	metav1.TypeMeta `json:",inline"` // nolint:revive
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MCPRegistry `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MCPRegistry{}, &MCPRegistryList{})
}
