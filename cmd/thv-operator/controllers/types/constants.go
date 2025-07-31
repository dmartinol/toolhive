package types

// Registry constants
const (
	DefaultRepositoryUrl = "https://github.com/modelcontextprotocol/modelcontextprotocol"
	RegistryImage        = "quay.io/ecosystem-appeng/mcp-registry@sha256:244dc363914a30a9a24fd7ea6a0e709bb885b806ca401354450b479e13e1a16b"
	RegistryPort         = 8080
	RegistryUiImage      = "quay.io/maorfr/mcp-registry-ui:latest"
	RegistryUiPort       = 8080

	// MongoDB constants
	MongoDBDatabaseType   = "mongodb"
	MongoDBDatabaseName   = "mcp-registry"
	MongoDBCollectionName = "servers_v2"
	MongoDBLogLevel       = "debug"
	MongoDBSeedImport     = "false"
	MongoDBPort           = "27017"

	// Finalizer constants
	MCPRegistryFinalizer = "mcpregistry.toolhive.stacklok.dev/finalizer"
)
