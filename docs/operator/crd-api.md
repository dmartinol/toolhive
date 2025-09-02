# API Reference

## Packages
- [toolhive.stacklok.dev/v1alpha1](#toolhivestacklokdevv1alpha1)


## toolhive.stacklok.dev/v1alpha1

Package v1alpha1 contains API Schema definitions for the toolhive v1alpha1 API group

### Resource Types
- [MCPRegistry](#mcpregistry)
- [MCPRegistryList](#mcpregistrylist)
- [MCPServer](#mcpserver)
- [MCPServerList](#mcpserverlist)



#### AuthzConfigRef



AuthzConfigRef defines a reference to authorization configuration



_Appears in:_
- [MCPServerSpec](#mcpserverspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _string_ | Type is the type of authorization configuration | configMap | Enum: [configMap inline] <br /> |
| `configMap` _[ConfigMapAuthzRef](#configmapauthzref)_ | ConfigMap references a ConfigMap containing authorization configuration<br />Only used when Type is "configMap" |  |  |
| `inline` _[InlineAuthzConfig](#inlineauthzconfig)_ | Inline contains direct authorization configuration<br />Only used when Type is "inline" |  |  |


#### BasicAuth



BasicAuth defines basic authentication credentials



_Appears in:_
- [HTTPAuthentication](#httpauthentication)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `username` _string_ | Username is the username for basic auth |  | Required: \{\} <br /> |
| `password` _string_ | Password is the password for basic auth<br />For security, this should reference a Secret |  |  |
| `secretRef` _[SecretKeyRef](#secretkeyref)_ | SecretRef references a Secret containing the password |  |  |


#### BearerTokenAuth



BearerTokenAuth defines bearer token authentication



_Appears in:_
- [HTTPAuthentication](#httpauthentication)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `token` _string_ | Token is the bearer token value<br />For security, this should reference a Secret |  |  |
| `secretRef` _[SecretKeyRef](#secretkeyref)_ | SecretRef references a Secret containing the bearer token |  |  |


#### ConfigMapAuthzRef



ConfigMapAuthzRef references a ConfigMap containing authorization configuration



_Appears in:_
- [AuthzConfigRef](#authzconfigref)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the ConfigMap |  | Required: \{\} <br /> |
| `key` _string_ | Key is the key in the ConfigMap that contains the authorization configuration | authz.json |  |


#### ConfigMapOIDCRef



ConfigMapOIDCRef references a ConfigMap containing OIDC configuration



_Appears in:_
- [OIDCConfigRef](#oidcconfigref)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the ConfigMap |  | Required: \{\} <br /> |
| `key` _string_ | Key is the key in the ConfigMap that contains the OIDC configuration | oidc.json |  |


#### ConfigMapReference



ConfigMapReference references a ConfigMap



_Appears in:_
- [StorageReference](#storagereference)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the ConfigMap |  |  |
| `namespace` _string_ | Namespace is the namespace of the ConfigMap |  |  |
| `key` _string_ | Key is the key in the ConfigMap |  |  |


#### ConfigMapRegistrySource



ConfigMapRegistrySource defines a ConfigMap source



_Appears in:_
- [MCPRegistrySource](#mcpregistrysource)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the ConfigMap |  | Required: \{\} <br /> |
| `namespace` _string_ | Namespace is the namespace of the ConfigMap<br />If not specified, defaults to the MCPRegistry's namespace |  |  |
| `key` _string_ | Key is the key in the ConfigMap containing the registry data | registry.json |  |


#### EnvVar



EnvVar represents an environment variable in a container



_Appears in:_
- [MCPServerSpec](#mcpserverspec)
- [ProxyDeploymentOverrides](#proxydeploymentoverrides)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name of the environment variable |  | Required: \{\} <br /> |
| `value` _string_ | Value of the environment variable |  | Required: \{\} <br /> |


#### GitAuthentication



GitAuthentication defines Git authentication methods



_Appears in:_
- [GitRegistrySource](#gitregistrysource)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sshKey` _[SSHKeyAuth](#sshkeyauth)_ | SSHKey provides SSH key authentication for Git |  |  |
| `token` _[TokenAuth](#tokenauth)_ | Token provides token-based authentication for Git |  |  |


#### GitRegistrySource



GitRegistrySource defines a Git repository source



_Appears in:_
- [MCPRegistrySource](#mcpregistrysource)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `repository` _string_ | Repository is the Git repository URL |  | Required: \{\} <br /> |
| `ref` _string_ | Ref is the Git reference (branch, tag, or commit) | main |  |
| `path` _string_ | Path is the path within the repository to the registry file | registry.json |  |
| `authentication` _[GitAuthentication](#gitauthentication)_ | Authentication defines authentication for Git operations |  |  |


#### HTTPAuthentication



HTTPAuthentication defines HTTP authentication methods



_Appears in:_
- [RegistryRegistrySource](#registryregistrysource)
- [URLRegistrySource](#urlregistrysource)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `bearerToken` _[BearerTokenAuth](#bearertokenauth)_ | BearerToken provides a bearer token for authentication |  |  |
| `basicAuth` _[BasicAuth](#basicauth)_ | BasicAuth provides basic authentication credentials |  |  |


#### InlineAuthzConfig



InlineAuthzConfig contains direct authorization configuration



_Appears in:_
- [AuthzConfigRef](#authzconfigref)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `policies` _string array_ | Policies is a list of Cedar policy strings |  | MinItems: 1 <br />Required: \{\} <br /> |
| `entitiesJson` _string_ | EntitiesJSON is a JSON string representing Cedar entities | [] |  |


#### InlineOIDCConfig



InlineOIDCConfig contains direct OIDC configuration



_Appears in:_
- [OIDCConfigRef](#oidcconfigref)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `issuer` _string_ | Issuer is the OIDC issuer URL |  | Required: \{\} <br /> |
| `audience` _string_ | Audience is the expected audience for the token |  |  |
| `jwksUrl` _string_ | JWKSURL is the URL to fetch the JWKS from |  |  |
| `introspectionUrl` _string_ | IntrospectionURL is the URL for token introspection endpoint |  |  |
| `clientId` _string_ | ClientID is the OIDC client ID |  |  |
| `clientSecret` _string_ | ClientSecret is the client secret for introspection (optional) |  |  |
| `thvCABundlePath` _string_ | ThvCABundlePath is the path to CA certificate bundle file for HTTPS requests<br />The file must be mounted into the pod (e.g., via ConfigMap or Secret volume) |  |  |
| `jwksAuthTokenPath` _string_ | JWKSAuthTokenPath is the path to file containing bearer token for JWKS/OIDC requests<br />The file must be mounted into the pod (e.g., via Secret volume) |  |  |
| `jwksAllowPrivateIP` _boolean_ | JWKSAllowPrivateIP allows JWKS/OIDC endpoints on private IP addresses<br />Use with caution - only enable for trusted internal IDPs | false |  |


#### KubernetesOIDCConfig



KubernetesOIDCConfig configures OIDC for Kubernetes service account token validation



_Appears in:_
- [OIDCConfigRef](#oidcconfigref)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `serviceAccount` _string_ | ServiceAccount is the name of the service account to validate tokens for<br />If empty, uses the pod's service account |  |  |
| `namespace` _string_ | Namespace is the namespace of the service account<br />If empty, uses the MCPServer's namespace |  |  |
| `audience` _string_ | Audience is the expected audience for the token | toolhive |  |
| `issuer` _string_ | Issuer is the OIDC issuer URL | https://kubernetes.default.svc |  |
| `jwksUrl` _string_ | JWKSURL is the URL to fetch the JWKS from<br />If empty, OIDC discovery will be used to automatically determine the JWKS URL |  |  |
| `introspectionUrl` _string_ | IntrospectionURL is the URL for token introspection endpoint<br />If empty, OIDC discovery will be used to automatically determine the introspection URL |  |  |
| `useClusterAuth` _boolean_ | UseClusterAuth enables using the Kubernetes cluster's CA bundle and service account token<br />When true, uses /var/run/secrets/kubernetes.io/serviceaccount/ca.crt for TLS verification<br />and /var/run/secrets/kubernetes.io/serviceaccount/token for bearer token authentication<br />Defaults to true if not specified |  |  |


#### MCPRegistry



MCPRegistry is the Schema for the mcpregistries API



_Appears in:_
- [MCPRegistryList](#mcpregistrylist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `toolhive.stacklok.dev/v1alpha1` | | |
| `kind` _string_ | `MCPRegistry` | | |
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br />Servers may infer this from the endpoint the client submits requests to.<br />Cannot be updated.<br />In CamelCase.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |  |  |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br />Servers should convert recognized schemas to the latest internal value, and<br />may reject unrecognized values.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |  |  |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[MCPRegistrySpec](#mcpregistryspec)_ |  |  |  |
| `status` _[MCPRegistryStatus](#mcpregistrystatus)_ |  |  |  |


#### MCPRegistryFilter



MCPRegistryFilter defines filtering criteria



_Appears in:_
- [MCPRegistrySpec](#mcpregistryspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `include` _string array_ | Include specifies patterns for servers to include |  |  |
| `exclude` _string array_ | Exclude specifies patterns for servers to exclude |  |  |
| `tags` _[TagFilter](#tagfilter)_ | Tags specifies tag-based filtering |  |  |


#### MCPRegistryList



MCPRegistryList contains a list of MCPRegistry





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `toolhive.stacklok.dev/v1alpha1` | | |
| `kind` _string_ | `MCPRegistryList` | | |
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br />Servers may infer this from the endpoint the client submits requests to.<br />Cannot be updated.<br />In CamelCase.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |  |  |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br />Servers should convert recognized schemas to the latest internal value, and<br />may reject unrecognized values.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |  |  |
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `items` _[MCPRegistry](#mcpregistry) array_ |  |  |  |


#### MCPRegistryPhase

_Underlying type:_ _string_

MCPRegistryPhase represents the lifecycle phase of an MCPRegistry

_Validation:_
- Enum: [Pending Syncing Ready Failed Updating]

_Appears in:_
- [MCPRegistryStatus](#mcpregistrystatus)

| Field | Description |
| --- | --- |
| `Pending` | MCPRegistryPhasePending means the MCPRegistry is being initialized<br /> |
| `Syncing` | MCPRegistryPhaseSyncing means the MCPRegistry is synchronizing data<br /> |
| `Ready` | MCPRegistryPhaseReady means the MCPRegistry is ready and up-to-date<br /> |
| `Failed` | MCPRegistryPhaseFailed means the MCPRegistry encountered an error<br /> |
| `Updating` | MCPRegistryPhaseUpdating means the MCPRegistry is updating its data<br /> |


#### MCPRegistrySource



MCPRegistrySource defines the source configuration for registry data



_Appears in:_
- [MCPRegistrySpec](#mcpregistryspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _string_ | Type specifies the source type |  | Enum: [configmap url git registry] <br />Required: \{\} <br /> |
| `configmap` _[ConfigMapRegistrySource](#configmapregistrysource)_ | ConfigMap references a ConfigMap containing registry data<br />Only used when Type is "configmap" |  |  |
| `url` _[URLRegistrySource](#urlregistrysource)_ | URL references an HTTP/HTTPS endpoint serving registry data<br />Only used when Type is "url" |  |  |
| `git` _[GitRegistrySource](#gitregistrysource)_ | Git references a Git repository containing registry data<br />Only used when Type is "git" |  |  |
| `registry` _[RegistryRegistrySource](#registryregistrysource)_ | Registry references an external registry<br />Only used when Type is "registry" |  |  |


#### MCPRegistrySpec



MCPRegistrySpec defines the desired state of MCPRegistry



_Appears in:_
- [MCPRegistry](#mcpregistry)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `displayName` _string_ | DisplayName is a human-readable name for the registry |  |  |
| `format` _string_ | Format specifies the registry data format | toolhive | Enum: [toolhive upstream] <br /> |
| `source` _[MCPRegistrySource](#mcpregistrysource)_ | Source defines where to fetch registry data from |  | Required: \{\} <br /> |
| `syncPolicy` _[MCPRegistrySyncPolicy](#mcpregistrysyncpolicy)_ | SyncPolicy defines the synchronization behavior |  |  |
| `filter` _[MCPRegistryFilter](#mcpregistryfilter)_ | Filter defines criteria for including/excluding servers |  |  |


#### MCPRegistryStatus



MCPRegistryStatus defines the observed state of MCPRegistry



_Appears in:_
- [MCPRegistry](#mcpregistry)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#condition-v1-meta) array_ | Conditions represent the latest available observations of the MCPRegistry's state |  |  |
| `phase` _[MCPRegistryPhase](#mcpregistryphase)_ | Phase is the current phase of the MCPRegistry |  | Enum: [Pending Syncing Ready Failed Updating] <br /> |
| `message` _string_ | Message provides additional information about the current phase |  |  |
| `lastSyncTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#time-v1-meta)_ | LastSyncTime is the timestamp of the last successful synchronization |  |  |
| `lastSyncHash` _string_ | LastSyncHash is a hash of the registry data from the last sync<br />Used to detect changes and avoid unnecessary updates |  |  |
| `serverCount` _integer_ | ServerCount is the number of servers currently in the registry |  |  |
| `syncAttempts` _integer_ | SyncAttempts tracks the number of sync attempts for the current operation |  |  |
| `storageRef` _[StorageReference](#storagereference)_ | StorageRef references the storage location for the registry data |  |  |


#### MCPRegistrySyncPolicy



MCPRegistrySyncPolicy defines synchronization behavior



_Appears in:_
- [MCPRegistrySpec](#mcpregistryspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _string_ | Type specifies the sync policy type | manual | Enum: [manual automatic] <br /> |
| `interval` _string_ | Interval specifies the sync interval for automatic synchronization<br />Only used when Type is "automatic" | 1h |  |
| `retryPolicy` _[RetryPolicy](#retrypolicy)_ | RetryPolicy defines retry behavior for failed syncs |  |  |


#### MCPServer



MCPServer is the Schema for the mcpservers API



_Appears in:_
- [MCPServerList](#mcpserverlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `toolhive.stacklok.dev/v1alpha1` | | |
| `kind` _string_ | `MCPServer` | | |
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br />Servers may infer this from the endpoint the client submits requests to.<br />Cannot be updated.<br />In CamelCase.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |  |  |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br />Servers should convert recognized schemas to the latest internal value, and<br />may reject unrecognized values.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |  |  |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[MCPServerSpec](#mcpserverspec)_ |  |  |  |
| `status` _[MCPServerStatus](#mcpserverstatus)_ |  |  |  |


#### MCPServerList



MCPServerList contains a list of MCPServer





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `toolhive.stacklok.dev/v1alpha1` | | |
| `kind` _string_ | `MCPServerList` | | |
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br />Servers may infer this from the endpoint the client submits requests to.<br />Cannot be updated.<br />In CamelCase.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |  |  |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br />Servers should convert recognized schemas to the latest internal value, and<br />may reject unrecognized values.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |  |  |
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `items` _[MCPServer](#mcpserver) array_ |  |  |  |


#### MCPServerPhase

_Underlying type:_ _string_

MCPServerPhase is the phase of the MCPServer

_Validation:_
- Enum: [Pending Running Failed Terminating]

_Appears in:_
- [MCPServerStatus](#mcpserverstatus)

| Field | Description |
| --- | --- |
| `Pending` | MCPServerPhasePending means the MCPServer is being created<br /> |
| `Running` | MCPServerPhaseRunning means the MCPServer is running<br /> |
| `Failed` | MCPServerPhaseFailed means the MCPServer failed to start<br /> |
| `Terminating` | MCPServerPhaseTerminating means the MCPServer is being deleted<br /> |


#### MCPServerSpec



MCPServerSpec defines the desired state of MCPServer



_Appears in:_
- [MCPServer](#mcpserver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `image` _string_ | Image is the container image for the MCP server |  | Required: \{\} <br /> |
| `transport` _string_ | Transport is the transport method for the MCP server (stdio, streamable-http or sse) | stdio | Enum: [stdio streamable-http sse] <br /> |
| `port` _integer_ | Port is the port to expose the MCP server on | 8080 | Maximum: 65535 <br />Minimum: 1 <br /> |
| `targetPort` _integer_ | TargetPort is the port that MCP server listens to |  | Maximum: 65535 <br />Minimum: 1 <br /> |
| `args` _string array_ | Args are additional arguments to pass to the MCP server |  |  |
| `env` _[EnvVar](#envvar) array_ | Env are environment variables to set in the MCP server container |  |  |
| `volumes` _[Volume](#volume) array_ | Volumes are volumes to mount in the MCP server container |  |  |
| `resources` _[ResourceRequirements](#resourcerequirements)_ | Resources defines the resource requirements for the MCP server container |  |  |
| `secrets` _[SecretRef](#secretref) array_ | Secrets are references to secrets to mount in the MCP server container |  |  |
| `serviceAccount` _string_ | ServiceAccount is the name of an already existing service account to use by the MCP server.<br />If not specified, a ServiceAccount will be created automatically and used by the MCP server. |  |  |
| `permissionProfile` _[PermissionProfileRef](#permissionprofileref)_ | PermissionProfile defines the permission profile to use |  |  |
| `podTemplateSpec` _[PodTemplateSpec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#podtemplatespec-v1-core)_ | PodTemplateSpec defines the pod template to use for the MCP server<br />This allows for customizing the pod configuration beyond what is provided by the other fields.<br />Note that to modify the specific container the MCP server runs in, you must specify<br />the `mcp` container name in the PodTemplateSpec. |  |  |
| `resourceOverrides` _[ResourceOverrides](#resourceoverrides)_ | ResourceOverrides allows overriding annotations and labels for resources created by the operator |  |  |
| `oidcConfig` _[OIDCConfigRef](#oidcconfigref)_ | OIDCConfig defines OIDC authentication configuration for the MCP server |  |  |
| `authzConfig` _[AuthzConfigRef](#authzconfigref)_ | AuthzConfig defines authorization policy configuration for the MCP server |  |  |
| `tools` _string array_ | ToolsFilter is the filter on tools applied to the MCP server |  |  |
| `telemetry` _[TelemetryConfig](#telemetryconfig)_ | Telemetry defines observability configuration for the MCP server |  |  |


#### MCPServerStatus



MCPServerStatus defines the observed state of MCPServer



_Appears in:_
- [MCPServer](#mcpserver)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#condition-v1-meta) array_ | Conditions represent the latest available observations of the MCPServer's state |  |  |
| `url` _string_ | URL is the URL where the MCP server can be accessed |  |  |
| `phase` _[MCPServerPhase](#mcpserverphase)_ | Phase is the current phase of the MCPServer |  | Enum: [Pending Running Failed Terminating] <br /> |
| `message` _string_ | Message provides additional information about the current phase |  |  |


#### NetworkPermissions



NetworkPermissions defines the network permissions for an MCP server



_Appears in:_
- [PermissionProfileSpec](#permissionprofilespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `outbound` _[OutboundNetworkPermissions](#outboundnetworkpermissions)_ | Outbound defines the outbound network permissions |  |  |


#### OIDCConfigRef



OIDCConfigRef defines a reference to OIDC configuration



_Appears in:_
- [MCPServerSpec](#mcpserverspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _string_ | Type is the type of OIDC configuration | kubernetes | Enum: [kubernetes configMap inline] <br /> |
| `resourceUrl` _string_ | ResourceURL is the explicit resource URL for OAuth discovery endpoint (RFC 9728)<br />If not specified, defaults to the in-cluster Kubernetes service URL |  |  |
| `kubernetes` _[KubernetesOIDCConfig](#kubernetesoidcconfig)_ | Kubernetes configures OIDC for Kubernetes service account token validation<br />Only used when Type is "kubernetes" |  |  |
| `configMap` _[ConfigMapOIDCRef](#configmapoidcref)_ | ConfigMap references a ConfigMap containing OIDC configuration<br />Only used when Type is "configmap" |  |  |
| `inline` _[InlineOIDCConfig](#inlineoidcconfig)_ | Inline contains direct OIDC configuration<br />Only used when Type is "inline" |  |  |


#### OpenTelemetryConfig



OpenTelemetryConfig defines pure OpenTelemetry configuration



_Appears in:_
- [TelemetryConfig](#telemetryconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled controls whether OpenTelemetry is enabled | false |  |
| `endpoint` _string_ | Endpoint is the OTLP endpoint URL for tracing and metrics |  |  |
| `serviceName` _string_ | ServiceName is the service name for telemetry<br />If not specified, defaults to the MCPServer name |  |  |
| `headers` _string array_ | Headers contains authentication headers for the OTLP endpoint<br />Specified as key=value pairs |  |  |
| `insecure` _boolean_ | Insecure indicates whether to use HTTP instead of HTTPS for the OTLP endpoint | false |  |
| `metrics` _[OpenTelemetryMetricsConfig](#opentelemetrymetricsconfig)_ | Metrics defines OpenTelemetry metrics-specific configuration |  |  |


#### OpenTelemetryMetricsConfig



OpenTelemetryMetricsConfig defines OpenTelemetry metrics configuration



_Appears in:_
- [OpenTelemetryConfig](#opentelemetryconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled controls whether OTLP metrics are sent | true |  |


#### OutboundNetworkPermissions



OutboundNetworkPermissions defines the outbound network permissions



_Appears in:_
- [NetworkPermissions](#networkpermissions)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `insecureAllowAll` _boolean_ | InsecureAllowAll allows all outbound network connections (not recommended) | false |  |
| `allowHost` _string array_ | AllowHost is a list of hosts to allow connections to |  |  |
| `allowPort` _integer array_ | AllowPort is a list of ports to allow connections to |  |  |


#### PermissionProfileRef



PermissionProfileRef defines a reference to a permission profile



_Appears in:_
- [MCPServerSpec](#mcpserverspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _string_ | Type is the type of permission profile reference | builtin | Enum: [builtin configmap] <br /> |
| `name` _string_ | Name is the name of the permission profile<br />If Type is "builtin", Name must be one of: "none", "network"<br />If Type is "configmap", Name is the name of the ConfigMap |  | Required: \{\} <br /> |
| `key` _string_ | Key is the key in the ConfigMap that contains the permission profile<br />Only used when Type is "configmap" |  |  |




#### PrometheusConfig



PrometheusConfig defines Prometheus-specific configuration



_Appears in:_
- [TelemetryConfig](#telemetryconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `enabled` _boolean_ | Enabled controls whether Prometheus metrics endpoint is exposed | false |  |


#### ProxyDeploymentOverrides



ProxyDeploymentOverrides defines overrides specific to the proxy deployment



_Appears in:_
- [ResourceOverrides](#resourceoverrides)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `annotations` _object (keys:string, values:string)_ | Annotations to add or override on the resource |  |  |
| `labels` _object (keys:string, values:string)_ | Labels to add or override on the resource |  |  |
| `env` _[EnvVar](#envvar) array_ | Env are environment variables to set in the proxy container (thv run process)<br />These affect the toolhive proxy itself, not the MCP server it manages |  |  |


#### RegistryRegistrySource



RegistryRegistrySource defines an external registry source



_Appears in:_
- [MCPRegistrySource](#mcpregistrysource)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `url` _string_ | URL is the base URL of the external registry |  | Required: \{\} <br /> |
| `authentication` _[HTTPAuthentication](#httpauthentication)_ | Authentication defines authentication for registry access |  |  |


#### ResourceList



ResourceList is a set of (resource name, quantity) pairs



_Appears in:_
- [ResourceRequirements](#resourcerequirements)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `cpu` _string_ | CPU is the CPU limit in cores (e.g., "500m" for 0.5 cores) |  |  |
| `memory` _string_ | Memory is the memory limit in bytes (e.g., "64Mi" for 64 megabytes) |  |  |


#### ResourceMetadataOverrides



ResourceMetadataOverrides defines metadata overrides for a resource



_Appears in:_
- [ProxyDeploymentOverrides](#proxydeploymentoverrides)
- [ResourceOverrides](#resourceoverrides)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `annotations` _object (keys:string, values:string)_ | Annotations to add or override on the resource |  |  |
| `labels` _object (keys:string, values:string)_ | Labels to add or override on the resource |  |  |


#### ResourceOverrides



ResourceOverrides defines overrides for annotations and labels on created resources



_Appears in:_
- [MCPServerSpec](#mcpserverspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `proxyDeployment` _[ProxyDeploymentOverrides](#proxydeploymentoverrides)_ | ProxyDeployment defines overrides for the Proxy Deployment resource (toolhive proxy) |  |  |
| `proxyService` _[ResourceMetadataOverrides](#resourcemetadataoverrides)_ | ProxyService defines overrides for the Proxy Service resource (points to the proxy deployment) |  |  |


#### ResourceRequirements



ResourceRequirements describes the compute resource requirements



_Appears in:_
- [MCPServerSpec](#mcpserverspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `limits` _[ResourceList](#resourcelist)_ | Limits describes the maximum amount of compute resources allowed |  |  |
| `requests` _[ResourceList](#resourcelist)_ | Requests describes the minimum amount of compute resources required |  |  |


#### RetryPolicy



RetryPolicy defines retry behavior



_Appears in:_
- [MCPRegistrySyncPolicy](#mcpregistrysyncpolicy)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `maxAttempts` _integer_ | MaxAttempts is the maximum number of retry attempts | 3 | Minimum: 1 <br /> |
| `backoffInterval` _string_ | BackoffInterval is the base interval between retries | 30s |  |
| `backoffMultiplier` _string_ | BackoffMultiplier is the multiplier for exponential backoff | 2.0 |  |


#### SSHKeyAuth



SSHKeyAuth defines SSH key authentication



_Appears in:_
- [GitAuthentication](#gitauthentication)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `secretRef` _[SecretKeyRef](#secretkeyref)_ | SecretRef references a Secret containing the SSH private key |  | Required: \{\} <br /> |


#### SecretKeyRef



SecretKeyRef references a key in a Secret



_Appears in:_
- [BasicAuth](#basicauth)
- [BearerTokenAuth](#bearertokenauth)
- [SSHKeyAuth](#sshkeyauth)
- [TokenAuth](#tokenauth)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the Secret |  | Required: \{\} <br /> |
| `key` _string_ | Key is the key in the Secret |  | Required: \{\} <br /> |
| `namespace` _string_ | Namespace is the namespace of the Secret<br />If not specified, defaults to the MCPRegistry's namespace |  |  |


#### SecretRef



SecretRef is a reference to a secret



_Appears in:_
- [MCPServerSpec](#mcpserverspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the secret |  | Required: \{\} <br /> |
| `key` _string_ | Key is the key in the secret itself |  | Required: \{\} <br /> |
| `targetEnvName` _string_ | TargetEnvName is the environment variable to be used when setting up the secret in the MCP server<br />If left unspecified, it defaults to the key |  |  |


#### StorageReference



StorageReference references where registry data is stored



_Appears in:_
- [MCPRegistryStatus](#mcpregistrystatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _string_ | Type specifies the storage type |  |  |
| `configMapRef` _[ConfigMapReference](#configmapreference)_ | ConfigMapRef references the ConfigMap storing the registry data |  |  |


#### TLSConfig



TLSConfig defines TLS configuration



_Appears in:_
- [URLRegistrySource](#urlregistrysource)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `insecureSkipVerify` _boolean_ | InsecureSkipVerify skips TLS certificate verification | false |  |
| `caBundle` _string_ | CABundle is a PEM-encoded CA certificate bundle |  |  |


#### TagFilter



TagFilter defines tag-based filtering



_Appears in:_
- [MCPRegistryFilter](#mcpregistryfilter)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `include` _string array_ | Include specifies tags that must be present |  |  |
| `exclude` _string array_ | Exclude specifies tags that must not be present |  |  |


#### TelemetryConfig



TelemetryConfig defines observability configuration for the MCP server



_Appears in:_
- [MCPServerSpec](#mcpserverspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `openTelemetry` _[OpenTelemetryConfig](#opentelemetryconfig)_ | OpenTelemetry defines OpenTelemetry configuration |  |  |
| `prometheus` _[PrometheusConfig](#prometheusconfig)_ | Prometheus defines Prometheus-specific configuration |  |  |


#### TokenAuth



TokenAuth defines token-based authentication



_Appears in:_
- [GitAuthentication](#gitauthentication)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `token` _string_ | Token is the authentication token<br />For security, this should reference a Secret |  |  |
| `secretRef` _[SecretKeyRef](#secretkeyref)_ | SecretRef references a Secret containing the token |  |  |


#### URLRegistrySource



URLRegistrySource defines an HTTP/HTTPS source



_Appears in:_
- [MCPRegistrySource](#mcpregistrysource)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `url` _string_ | URL is the HTTP/HTTPS endpoint serving registry data |  | Pattern: `^https?://.*` <br />Required: \{\} <br /> |
| `headers` _object (keys:string, values:string)_ | Headers contains optional HTTP headers for the request |  |  |
| `tlsConfig` _[TLSConfig](#tlsconfig)_ | TLSConfig defines TLS configuration for HTTPS requests |  |  |
| `authentication` _[HTTPAuthentication](#httpauthentication)_ | Authentication defines authentication for the HTTP request |  |  |


#### Volume



Volume represents a volume to mount in a container



_Appears in:_
- [MCPServerSpec](#mcpserverspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the volume |  | Required: \{\} <br /> |
| `hostPath` _string_ | HostPath is the path on the host to mount |  | Required: \{\} <br /> |
| `mountPath` _string_ | MountPath is the path in the container to mount to |  | Required: \{\} <br /> |
| `readOnly` _boolean_ | ReadOnly specifies whether the volume should be mounted read-only | false |  |


