package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MCPRegistrySpec defines the desired state of MCPRegistry
type MCPRegistrySpec struct {
	// RefreshInterval is the interval to refresh the registry data
	// +kubebuilder:default="1h"
	// +optional
	RefreshInterval string `json:"refreshInterval,omitempty"`

	// Timeout is the timeout for registry operations
	// +kubebuilder:default="30s"
	// +optional
	Timeout string `json:"timeout,omitempty"`
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

	// MongoDBServiceURL is the URL of the MongoDB service used by this registry
	// +optional
	MongoDBServiceURL string `json:"mongodbServiceURL,omitempty"`

	// MongoDBCollectionName is the name of the MongoDB collection used by this registry
	// +optional
	MongoDBCollectionName string `json:"mongodbCollectionName,omitempty"`
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
