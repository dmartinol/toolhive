package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Registry source types
const (
	// RegistrySourceTypeConfigMap is the type for registry data stored in ConfigMaps
	RegistrySourceTypeConfigMap = "configmap"

	// RegistrySourceTypeURL is the type for registry data fetched from HTTP/HTTPS endpoints
	RegistrySourceTypeURL = "url"

	// RegistrySourceTypeGit is the type for registry data fetched from Git repositories
	RegistrySourceTypeGit = "git"

	// RegistrySourceTypeRegistry is the type for registry data fetched from external registries
	RegistrySourceTypeRegistry = "registry"
)

// Registry formats
const (
	// RegistryFormatToolHive is the native ToolHive registry format
	RegistryFormatToolHive = "toolhive"

	// RegistryFormatUpstream is the upstream MCP registry format
	RegistryFormatUpstream = "upstream"
)

// Sync policy types
const (
	// SyncPolicyManual requires manual synchronization triggers
	SyncPolicyManual = "manual"

	// SyncPolicyAutomatic enables automatic synchronization at intervals
	SyncPolicyAutomatic = "automatic"
)

// MCPRegistrySpec defines the desired state of MCPRegistry
type MCPRegistrySpec struct {
	// DisplayName is a human-readable name for the registry
	// +optional
	DisplayName string `json:"displayName,omitempty"`

	// Source defines where to fetch registry data from
	// +kubebuilder:validation:Required
	Source MCPRegistrySource `json:"source"`

	// SyncPolicy defines the synchronization behavior
	// +optional
	SyncPolicy *MCPRegistrySyncPolicy `json:"syncPolicy,omitempty"`

	// Filter defines criteria for including/excluding servers
	// +optional
	Filter *MCPRegistryFilter `json:"filter,omitempty"`
}

// MCPRegistrySource defines the source configuration for registry data
type MCPRegistrySource struct {
	// Type specifies the source type
	// +kubebuilder:validation:Enum=configmap;url;git;registry
	// +kubebuilder:validation:Required
	Type string `json:"type"`

	// Format specifies the registry data format
	// +kubebuilder:validation:Enum=toolhive;upstream
	// +kubebuilder:default=toolhive
	// +optional
	Format string `json:"format,omitempty"`

	// ConfigMap references a ConfigMap containing registry data
	// Only used when Type is "configmap"
	// +optional
	ConfigMap *ConfigMapRegistrySource `json:"configmap,omitempty"`

	// URL references an HTTP/HTTPS endpoint serving registry data
	// Only used when Type is "url"
	// +optional
	URL *URLRegistrySource `json:"url,omitempty"`

	// Git references a Git repository containing registry data
	// Only used when Type is "git"
	// +optional
	Git *GitRegistrySource `json:"git,omitempty"`

	// Registry references an external registry
	// Only used when Type is "registry"
	// +optional
	Registry *RegistryRegistrySource `json:"registry,omitempty"`
}

// ConfigMapRegistrySource defines a ConfigMap source
type ConfigMapRegistrySource struct {
	// Name is the name of the ConfigMap
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace is the namespace of the ConfigMap
	// If not specified, defaults to the MCPRegistry's namespace
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Key is the key in the ConfigMap containing the registry data
	// +kubebuilder:default=registry.json
	// +optional
	Key string `json:"key,omitempty"`
}

// URLRegistrySource defines an HTTP/HTTPS source
type URLRegistrySource struct {
	// URL is the HTTP/HTTPS endpoint serving registry data
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="^https?://.*"
	URL string `json:"url"`

	// Headers contains optional HTTP headers for the request
	// +optional
	Headers map[string]string `json:"headers,omitempty"`

	// TLSConfig defines TLS configuration for HTTPS requests
	// +optional
	TLSConfig *TLSConfig `json:"tlsConfig,omitempty"`

	// Authentication defines authentication for the HTTP request
	// +optional
	Authentication *HTTPAuthentication `json:"authentication,omitempty"`
}

// GitRegistrySource defines a Git repository source
type GitRegistrySource struct {
	// Repository is the Git repository URL
	// +kubebuilder:validation:Required
	Repository string `json:"repository"`

	// Ref is the Git reference (branch, tag, or commit)
	// +kubebuilder:default=main
	// +optional
	Ref string `json:"ref,omitempty"`

	// Path is the path within the repository to the registry file
	// +kubebuilder:default=registry.json
	// +optional
	Path string `json:"path,omitempty"`

	// Authentication defines authentication for Git operations
	// +optional
	Authentication *GitAuthentication `json:"authentication,omitempty"`
}

// RegistryRegistrySource defines an external registry source
type RegistryRegistrySource struct {
	// URL is the base URL of the external registry
	// +kubebuilder:validation:Required
	URL string `json:"url"`

	// Authentication defines authentication for registry access
	// +optional
	Authentication *HTTPAuthentication `json:"authentication,omitempty"`
}

// TLSConfig defines TLS configuration
type TLSConfig struct {
	// InsecureSkipVerify skips TLS certificate verification
	// +kubebuilder:default=false
	// +optional
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`

	// CABundle is a PEM-encoded CA certificate bundle
	// +optional
	CABundle string `json:"caBundle,omitempty"`
}

// HTTPAuthentication defines HTTP authentication methods
type HTTPAuthentication struct {
	// BearerToken provides a bearer token for authentication
	// +optional
	BearerToken *BearerTokenAuth `json:"bearerToken,omitempty"`

	// BasicAuth provides basic authentication credentials
	// +optional
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`
}

// BearerTokenAuth defines bearer token authentication
type BearerTokenAuth struct {
	// Token is the bearer token value
	// For security, this should reference a Secret
	// +optional
	Token string `json:"token,omitempty"`

	// SecretRef references a Secret containing the bearer token
	// +optional
	SecretRef *SecretKeyRef `json:"secretRef,omitempty"`
}

// BasicAuth defines basic authentication credentials
type BasicAuth struct {
	// Username is the username for basic auth
	// +kubebuilder:validation:Required
	Username string `json:"username"`

	// Password is the password for basic auth
	// For security, this should reference a Secret
	// +optional
	Password string `json:"password,omitempty"`

	// SecretRef references a Secret containing the password
	// +optional
	SecretRef *SecretKeyRef `json:"secretRef,omitempty"`
}

// GitAuthentication defines Git authentication methods
type GitAuthentication struct {
	// SSHKey provides SSH key authentication for Git
	// +optional
	SSHKey *SSHKeyAuth `json:"sshKey,omitempty"`

	// Token provides token-based authentication for Git
	// +optional
	Token *TokenAuth `json:"token,omitempty"`
}

// SSHKeyAuth defines SSH key authentication
type SSHKeyAuth struct {
	// SecretRef references a Secret containing the SSH private key
	// +kubebuilder:validation:Required
	SecretRef SecretKeyRef `json:"secretRef"`
}

// TokenAuth defines token-based authentication
type TokenAuth struct {
	// Token is the authentication token
	// For security, this should reference a Secret
	// +optional
	Token string `json:"token,omitempty"`

	// SecretRef references a Secret containing the token
	// +optional
	SecretRef *SecretKeyRef `json:"secretRef,omitempty"`
}

// SecretKeyRef references a key in a Secret
type SecretKeyRef struct {
	// Name is the name of the Secret
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Key is the key in the Secret
	// +kubebuilder:validation:Required
	Key string `json:"key"`

	// Namespace is the namespace of the Secret
	// If not specified, defaults to the MCPRegistry's namespace
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// MCPRegistrySyncPolicy defines synchronization behavior
type MCPRegistrySyncPolicy struct {
	// Type specifies the sync policy type
	// +kubebuilder:validation:Enum=manual;automatic
	// +kubebuilder:default=manual
	// +optional
	Type string `json:"type,omitempty"`

	// Interval specifies the sync interval for automatic synchronization using Go duration format.
	// Valid units: s (seconds), m (minutes), h (hours)
	// Examples: "3m", "1h", "24h" (1 day), "168h" (1 week), "720h" (1 month)
	// Only used when Type is "automatic"
	// +kubebuilder:default="1h"
	// +optional
	Interval string `json:"interval,omitempty"`

	// RetryPolicy defines retry behavior for failed syncs
	// +optional
	RetryPolicy *RetryPolicy `json:"retryPolicy,omitempty"`
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	// MaxAttempts is the maximum number of retry attempts
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=3
	// +optional
	MaxAttempts int32 `json:"maxAttempts,omitempty"`

	// BackoffInterval is the base interval between retries
	// +kubebuilder:default="30s"
	// +optional
	BackoffInterval string `json:"backoffInterval,omitempty"`

	// BackoffMultiplier is the multiplier for exponential backoff
	// +kubebuilder:default="2.0"
	// +optional
	BackoffMultiplier string `json:"backoffMultiplier,omitempty"`
}

// MCPRegistryFilter defines filtering criteria
type MCPRegistryFilter struct {
	// Include specifies patterns for servers to include
	// +optional
	Include []string `json:"include,omitempty"`

	// Exclude specifies patterns for servers to exclude
	// +optional
	Exclude []string `json:"exclude,omitempty"`

	// Tags specifies tag-based filtering
	// +optional
	Tags *TagFilter `json:"tags,omitempty"`
}

// TagFilter defines tag-based filtering
type TagFilter struct {
	// Include specifies tags that must be present
	// +optional
	Include []string `json:"include,omitempty"`

	// Exclude specifies tags that must not be present
	// +optional
	Exclude []string `json:"exclude,omitempty"`
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

	// LastSyncTime is the timestamp of the last successful synchronization
	// +optional
	LastSyncTime *metav1.Time `json:"lastSyncTime,omitempty"`

	// LastSyncHash is a hash of the registry data from the last sync
	// Used to detect changes and avoid unnecessary updates
	// +optional
	LastSyncHash string `json:"lastSyncHash,omitempty"`

	// ServerCount is the number of servers currently in the registry
	// +optional
	ServerCount int32 `json:"serverCount,omitempty"`

	// SyncAttempts tracks the number of sync attempts for the current operation
	// +optional
	SyncAttempts int32 `json:"syncAttempts,omitempty"`

	// StorageRef references the storage location for the registry data
	// +optional
	StorageRef *StorageReference `json:"storageRef,omitempty"`
}

// MCPRegistryPhase represents the lifecycle phase of an MCPRegistry
// +kubebuilder:validation:Enum=Pending;Syncing;Ready;Failed;Updating
type MCPRegistryPhase string

const (
	// MCPRegistryPhasePending means the MCPRegistry is being initialized
	MCPRegistryPhasePending MCPRegistryPhase = "Pending"

	// MCPRegistryPhaseSyncing means the MCPRegistry is synchronizing data
	MCPRegistryPhaseSyncing MCPRegistryPhase = "Syncing"

	// MCPRegistryPhaseReady means the MCPRegistry is ready and up-to-date
	MCPRegistryPhaseReady MCPRegistryPhase = "Ready"

	// MCPRegistryPhaseFailed means the MCPRegistry encountered an error
	MCPRegistryPhaseFailed MCPRegistryPhase = "Failed"

	// MCPRegistryPhaseUpdating means the MCPRegistry is updating its data
	MCPRegistryPhaseUpdating MCPRegistryPhase = "Updating"
)

// StorageReference references where registry data is stored
type StorageReference struct {
	// Type specifies the storage type
	// +optional
	Type string `json:"type,omitempty"`

	// ConfigMapRef references the ConfigMap storing the registry data
	// +optional
	ConfigMapRef *ConfigMapReference `json:"configMapRef,omitempty"`
}

// ConfigMapReference references a ConfigMap
type ConfigMapReference struct {
	// Name is the name of the ConfigMap
	// +optional
	Name string `json:"name,omitempty"`

	// Namespace is the namespace of the ConfigMap
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Key is the key in the ConfigMap
	// +optional
	Key string `json:"key,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
//+kubebuilder:printcolumn:name="Source",type="string",JSONPath=".spec.source.type"
//+kubebuilder:printcolumn:name="Format",type="string",JSONPath=".spec.format"
//+kubebuilder:printcolumn:name="Servers",type="integer",JSONPath=".status.serverCount"
//+kubebuilder:printcolumn:name="Last Sync",type="date",JSONPath=".status.lastSyncTime"
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