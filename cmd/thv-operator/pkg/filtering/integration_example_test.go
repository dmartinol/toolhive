package filtering_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mcpv1alpha1 "github.com/stacklok/toolhive/cmd/thv-operator/api/v1alpha1"
	"github.com/stacklok/toolhive/cmd/thv-operator/pkg/filtering"
	"github.com/stacklok/toolhive/pkg/registry"
)

// TestEndToEndFilteringIntegration demonstrates how filtering works end-to-end
// with various MCPRegistry filter configurations
func TestEndToEndFilteringIntegration(t *testing.T) {
	t.Parallel()

	// Create a comprehensive test registry similar to what would come from a real source
	testRegistry := createComprehensiveTestRegistry()
	filterService := filtering.NewDefaultFilterService()
	ctx := context.Background()

	t.Run("Production Only Filter", func(t *testing.T) {
		t.Parallel()

		// Filter for production servers only
		filter := &mcpv1alpha1.RegistryFilter{
			Tags: &mcpv1alpha1.TagFilter{
				Include: []string{"production"},
			},
		}

		result, err := filterService.ApplyFilters(ctx, testRegistry, filter)
		require.NoError(t, err)

		// Should only include servers with "production" tag
		assert.Len(t, result.Servers, 2) // filesystem, database
		assert.Contains(t, result.Servers, "filesystem")
		assert.Contains(t, result.Servers, "database")
		assert.Len(t, result.RemoteServers, 1) // api-gateway
		assert.Contains(t, result.RemoteServers, "api-gateway")
	})

	t.Run("Exclude Experimental and Deprecated", func(t *testing.T) {
		t.Parallel()

		// Exclude experimental and deprecated servers
		filter := &mcpv1alpha1.RegistryFilter{
			Tags: &mcpv1alpha1.TagFilter{
				Exclude: []string{"experimental", "deprecated"},
			},
		}

		result, err := filterService.ApplyFilters(ctx, testRegistry, filter)
		require.NoError(t, err)

		// Should exclude weather-experimental and legacy-tools
		assert.Len(t, result.Servers, 2) // filesystem, database
		assert.NotContains(t, result.Servers, "weather-experimental")
		assert.NotContains(t, result.Servers, "legacy-tools")
		assert.Len(t, result.RemoteServers, 1) // api-gateway
		assert.NotContains(t, result.RemoteServers, "old-service")
	})

	t.Run("Development Tools Pattern", func(t *testing.T) {
		t.Parallel()

		// Include only development tools with specific patterns
		filter := &mcpv1alpha1.RegistryFilter{
			NameFilters: &mcpv1alpha1.NameFilter{
				Include: []string{"*-tools", "weather-*"},
			},
			Tags: &mcpv1alpha1.TagFilter{
				Exclude: []string{"deprecated"},
			},
		}

		result, err := filterService.ApplyFilters(ctx, testRegistry, filter)
		require.NoError(t, err)

		// Should include weather-experimental (matches weather-* and not deprecated)
		// Should exclude legacy-tools (matches *-tools but has deprecated tag)
		assert.Len(t, result.Servers, 1) // weather-experimental
		assert.Contains(t, result.Servers, "weather-experimental")
		assert.Len(t, result.RemoteServers, 0)
	})

	t.Run("Complex Combined Filter", func(t *testing.T) {
		t.Parallel()

		// Complex filter: Include API servers, exclude legacy, require specific tags
		filter := &mcpv1alpha1.RegistryFilter{
			NameFilters: &mcpv1alpha1.NameFilter{
				Include: []string{"*api*", "database"},
				Exclude: []string{"*legacy*"},
			},
			Tags: &mcpv1alpha1.TagFilter{
				Include: []string{"api", "database"},
			},
		}

		result, err := filterService.ApplyFilters(ctx, testRegistry, filter)
		require.NoError(t, err)

		// Should include: database (matches name include + has database tag)
		// Should include: api-gateway (matches *api* name + has api tag)
		// Should exclude: old-service (doesn't match name patterns)
		assert.Len(t, result.Servers, 1) // database
		assert.Contains(t, result.Servers, "database")
		assert.Len(t, result.RemoteServers, 1) // api-gateway
		assert.Contains(t, result.RemoteServers, "api-gateway")
	})

	t.Run("No Matching Servers", func(t *testing.T) {
		t.Parallel()

		// Filter that matches no servers
		filter := &mcpv1alpha1.RegistryFilter{
			NameFilters: &mcpv1alpha1.NameFilter{
				Include: []string{"nonexistent-*"},
			},
		}

		result, err := filterService.ApplyFilters(ctx, testRegistry, filter)
		require.NoError(t, err)

		// Should result in empty registry
		assert.Len(t, result.Servers, 0)
		assert.Len(t, result.RemoteServers, 0)
		// But preserve metadata
		assert.Equal(t, testRegistry.Version, result.Version)
		assert.Equal(t, testRegistry.LastUpdated, result.LastUpdated)
	})

	t.Run("Metadata Preservation", func(t *testing.T) {
		t.Parallel()

		// Any filter should preserve registry metadata and groups
		filter := &mcpv1alpha1.RegistryFilter{
			Tags: &mcpv1alpha1.TagFilter{
				Include: []string{"production"},
			},
		}

		result, err := filterService.ApplyFilters(ctx, testRegistry, filter)
		require.NoError(t, err)

		// Metadata should be preserved
		assert.Equal(t, testRegistry.Version, result.Version)
		assert.Equal(t, testRegistry.LastUpdated, result.LastUpdated)
		assert.Equal(t, testRegistry.Groups, result.Groups)
	})
}

// createComprehensiveTestRegistry creates a test registry with various server types and tags
// that exercises different filtering scenarios
func createComprehensiveTestRegistry() *registry.Registry {
	return &registry.Registry{
		Version:     "1.0.0",
		LastUpdated: "2024-01-15T10:00:00Z",
		Servers: map[string]*registry.ImageMetadata{
			"filesystem": {
				BaseServerMetadata: registry.BaseServerMetadata{
					Name:        "filesystem",
					Description: "Filesystem operations server",
					Tags:        []string{"filesystem", "production", "stable"},
				},
				Image: "mcp/filesystem:latest",
			},
			"weather-experimental": {
				BaseServerMetadata: registry.BaseServerMetadata{
					Name:        "weather-experimental",
					Description: "Experimental weather server",
					Tags:        []string{"weather", "experimental", "beta"},
				},
				Image: "mcp/weather:experimental",
			},
			"database": {
				BaseServerMetadata: registry.BaseServerMetadata{
					Name:        "database",
					Description: "Database connectivity server",
					Tags:        []string{"database", "production", "sql"},
				},
				Image: "mcp/database:latest",
			},
			"legacy-tools": {
				BaseServerMetadata: registry.BaseServerMetadata{
					Name:        "legacy-tools",
					Description: "Legacy development tools",
					Tags:        []string{"tools", "deprecated", "legacy"},
				},
				Image: "mcp/legacy-tools:old",
			},
		},
		RemoteServers: map[string]*registry.RemoteServerMetadata{
			"api-gateway": {
				BaseServerMetadata: registry.BaseServerMetadata{
					Name:        "api-gateway",
					Description: "Remote API gateway service",
					Tags:        []string{"api", "production", "gateway"},
				},
				URL: "https://api.example.com/mcp",
			},
			"old-service": {
				BaseServerMetadata: registry.BaseServerMetadata{
					Name:        "old-service",
					Description: "Legacy remote service",
					Tags:        []string{"legacy", "deprecated", "remote"},
				},
				URL: "https://legacy.example.com/api",
			},
		},
		Groups: []*registry.Group{
			{
				Name:        "production-stack",
				Description: "Production-ready MCP servers",
			},
		},
	}
}
