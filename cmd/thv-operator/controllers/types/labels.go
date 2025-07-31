package types

// Label keys and values used by the controllers
const (
	// AppLabel is the standard app label key
	AppLabel = "app"

	// ComponentLabel is the component label key
	ComponentLabel = "component"

	// MongoDBComponent is the component value for MongoDB resources
	MongoDBComponent = "mongodb"

	// RegistryComponent is the component value for registry resources

	RegistryComponent = "mcp-registry"
	// RegistryNameLabel is the annotation key for associating MCPServers with MCPRegistries
	RegistryNameLabel = "toolhive.stacklok.dev/registry-name"

	// RegistryNamespaceLabel is the annotation key for specifying the registry namespace
	RegistryNamespaceLabel = "toolhive.stacklok.dev/registry-namespace"

	// RegisteredServerIDLabel is the label key for linking MCPServers to pre-registered server entries
	RegisteredServerIDLabel = "toolhive.stacklok.dev/registered-server-id"
)
