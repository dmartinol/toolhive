package controllers

import (
	"encoding/json"
	"fmt"
	"time"

	mcpv1alpha1 "github.com/stacklok/toolhive/cmd/thv-operator/api/v1alpha1"
	"github.com/stacklok/toolhive/pkg/registry"
)

// FormatConverter defines the interface for converting between registry formats
type FormatConverter interface {
	// Convert transforms registry data from one format to another
	Convert(data []byte, fromFormat, toFormat string) ([]byte, error)
	
	// Validate ensures the provided data is valid for the specified format
	Validate(data []byte, format string) error
	
	// SupportedFormats returns the list of supported registry formats
	SupportedFormats() []string
	
	// DetectFormat attempts to automatically detect the format of the provided data
	DetectFormat(data []byte) (string, error)
}

// RegistryFormatConverter implements FormatConverter using existing registry conversion logic
type RegistryFormatConverter struct{}

// NewRegistryFormatConverter creates a new instance of RegistryFormatConverter
func NewRegistryFormatConverter() FormatConverter {
	return &RegistryFormatConverter{}
}

// Convert transforms registry data between formats
func (r *RegistryFormatConverter) Convert(data []byte, fromFormat, toFormat string) ([]byte, error) {
	// Validate input parameters
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}
	
	// Normalize format strings
	fromFormat = normalizeFormat(fromFormat)
	toFormat = normalizeFormat(toFormat)
	
	// No conversion needed if formats are the same
	if fromFormat == toFormat {
		return data, nil
	}
	
	// Validate that both formats are supported
	supported := r.SupportedFormats()
	if !isFormatSupported(fromFormat, supported) {
		return nil, fmt.Errorf("unsupported source format: %s", fromFormat)
	}
	if !isFormatSupported(toFormat, supported) {
		return nil, fmt.Errorf("unsupported target format: %s", toFormat)
	}
	
	// Perform format conversion
	switch {
	case fromFormat == mcpv1alpha1.RegistryFormatUpstream && toFormat == mcpv1alpha1.RegistryFormatToolHive:
		return r.convertUpstreamToToolhive(data)
	case fromFormat == mcpv1alpha1.RegistryFormatToolHive && toFormat == mcpv1alpha1.RegistryFormatUpstream:
		return r.convertToolhiveToUpstream(data)
	default:
		return nil, fmt.Errorf("unsupported conversion: %s -> %s", fromFormat, toFormat)
	}
}

// convertUpstreamToToolhive converts upstream MCP registry format to ToolHive format
func (r *RegistryFormatConverter) convertUpstreamToToolhive(data []byte) ([]byte, error) {
	// Parse upstream registry format
	// Note: Upstream format is a map of server names to server details
	var upstreamRegistry map[string]*registry.UpstreamServerDetail
	if err := json.Unmarshal(data, &upstreamRegistry); err != nil {
		return nil, fmt.Errorf("failed to parse upstream registry format: %w", err)
	}
	
	// Create ToolHive registry structure
	toolhiveRegistry := &registry.Registry{
		Version:     "1.0.0",
		LastUpdated: time.Now().Format(time.RFC3339),
		Servers:     make(map[string]*registry.ImageMetadata),
	}
	
	// Convert each server using existing conversion logic
	for serverName, upstreamServer := range upstreamRegistry {
		toolhiveServer, err := registry.ConvertUpstreamToToolhive(upstreamServer)
		if err != nil {
			return nil, fmt.Errorf("failed to convert server %s: %w", serverName, err)
		}
		
		// ToolHive registry currently supports only ImageMetadata servers
		if imageServer, ok := toolhiveServer.(*registry.ImageMetadata); ok {
			toolhiveRegistry.Servers[serverName] = imageServer
		} else {
			// For now, skip remote servers as ToolHive registry format doesn't support them
			// This could be enhanced in the future to support mixed server types
			continue
		}
	}
	
	// Marshal to JSON
	result, err := json.Marshal(toolhiveRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ToolHive registry: %w", err)
	}
	
	return result, nil
}

// convertToolhiveToUpstream converts ToolHive registry format to upstream MCP format
func (r *RegistryFormatConverter) convertToolhiveToUpstream(data []byte) ([]byte, error) {
	// Parse ToolHive registry format
	var toolhiveRegistry registry.Registry
	if err := json.Unmarshal(data, &toolhiveRegistry); err != nil {
		return nil, fmt.Errorf("failed to parse ToolHive registry format: %w", err)
	}
	
	// Create upstream registry structure
	upstreamRegistry := make(map[string]*registry.UpstreamServerDetail)
	
	// Convert each server using existing conversion logic
	for serverName, toolhiveServer := range toolhiveRegistry.Servers {
		upstreamServer, err := registry.ConvertToolhiveToUpstream(toolhiveServer)
		if err != nil {
			return nil, fmt.Errorf("failed to convert server %s: %w", serverName, err)
		}
		
		upstreamRegistry[serverName] = upstreamServer
	}
	
	// Marshal to JSON
	result, err := json.Marshal(upstreamRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal upstream registry: %w", err)
	}
	
	return result, nil
}

// Validate ensures the provided data is valid for the specified format
func (r *RegistryFormatConverter) Validate(data []byte, format string) error {
	if len(data) == 0 {
		return fmt.Errorf("data cannot be empty")
	}
	
	format = normalizeFormat(format)
	
	switch format {
	case mcpv1alpha1.RegistryFormatToolHive:
		var toolhiveRegistry registry.Registry
		if err := json.Unmarshal(data, &toolhiveRegistry); err != nil {
			return fmt.Errorf("invalid ToolHive registry format: %w", err)
		}
		
		// Basic validation
		if toolhiveRegistry.Version == "" {
			return fmt.Errorf("ToolHive registry missing required version field")
		}
		if toolhiveRegistry.Servers == nil {
			return fmt.Errorf("ToolHive registry missing servers field")
		}
		
	case mcpv1alpha1.RegistryFormatUpstream:
		var upstreamRegistry map[string]*registry.UpstreamServerDetail
		if err := json.Unmarshal(data, &upstreamRegistry); err != nil {
			return fmt.Errorf("invalid upstream registry format: %w", err)
		}
		
		// Basic validation - upstream format is a map of server names to details
		if len(upstreamRegistry) == 0 {
			return fmt.Errorf("upstream registry contains no servers")
		}
		
		// Validate each server has required fields
		for serverName, serverDetail := range upstreamRegistry {
			if serverName == "" {
				return fmt.Errorf("upstream registry contains server with empty name")
			}
			if serverDetail == nil {
				return fmt.Errorf("upstream registry server %s has nil details", serverName)
			}
			if serverDetail.Server.Name == "" {
				return fmt.Errorf("upstream registry server %s missing server.name field", serverName)
			}
		}
		
	default:
		return fmt.Errorf("unsupported format for validation: %s", format)
	}
	
	return nil
}

// SupportedFormats returns the list of supported registry formats
func (r *RegistryFormatConverter) SupportedFormats() []string {
	return []string{
		mcpv1alpha1.RegistryFormatToolHive,
		mcpv1alpha1.RegistryFormatUpstream,
	}
}

// DetectFormat attempts to automatically detect the format of the provided data
func (r *RegistryFormatConverter) DetectFormat(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("cannot detect format of empty data")
	}
	
	// Try to parse as ToolHive format first
	var toolhiveRegistry registry.Registry
	if err := json.Unmarshal(data, &toolhiveRegistry); err == nil {
		// Check for ToolHive-specific fields
		if toolhiveRegistry.Version != "" && toolhiveRegistry.Servers != nil {
			return mcpv1alpha1.RegistryFormatToolHive, nil
		}
	}
	
	// Try to parse as upstream format
	var upstreamRegistry map[string]*registry.UpstreamServerDetail
	if err := json.Unmarshal(data, &upstreamRegistry); err == nil {
		// Check for upstream-specific structure
		if len(upstreamRegistry) > 0 {
			// Look for upstream-specific fields in the first server
			for _, serverDetail := range upstreamRegistry {
				if serverDetail != nil && serverDetail.Server.Name != "" {
					return mcpv1alpha1.RegistryFormatUpstream, nil
				}
				break // Only check the first server
			}
		}
	}
	
	return "", fmt.Errorf("unable to detect registry format: data does not match any known format")
}

// Helper functions

// normalizeFormat ensures format strings are consistent
func normalizeFormat(format string) string {
	if format == "" {
		return mcpv1alpha1.RegistryFormatToolHive // Default format
	}
	return format
}

// isFormatSupported checks if a format is in the supported list
func isFormatSupported(format string, supported []string) bool {
	for _, s := range supported {
		if s == format {
			return true
		}
	}
	return false
}