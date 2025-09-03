package controllers

import (
	"encoding/json"
	"testing"

	mcpv1alpha1 "github.com/stacklok/toolhive/cmd/thv-operator/api/v1alpha1"
	"github.com/stacklok/toolhive/pkg/registry"
)

func TestRegistryFormatConverter_SupportedFormats(t *testing.T) {
	converter := NewRegistryFormatConverter()
	formats := converter.SupportedFormats()
	
	expected := []string{
		mcpv1alpha1.RegistryFormatToolHive,
		mcpv1alpha1.RegistryFormatUpstream,
	}
	
	if len(formats) != len(expected) {
		t.Errorf("Expected %d supported formats, got %d", len(expected), len(formats))
	}
	
	for _, expectedFormat := range expected {
		found := false
		for _, format := range formats {
			if format == expectedFormat {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected format %s not found in supported formats", expectedFormat)
		}
	}
}

func TestRegistryFormatConverter_Convert_SameFormat(t *testing.T) {
	converter := NewRegistryFormatConverter()
	
	testData := []byte(`{"version": "1.0.0", "servers": {}}`)
	
	result, err := converter.Convert(testData, mcpv1alpha1.RegistryFormatToolHive, mcpv1alpha1.RegistryFormatToolHive)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if string(result) != string(testData) {
		t.Errorf("Expected no conversion when formats are the same")
	}
}

func TestRegistryFormatConverter_Convert_EmptyData(t *testing.T) {
	converter := NewRegistryFormatConverter()
	
	_, err := converter.Convert([]byte{}, mcpv1alpha1.RegistryFormatToolHive, mcpv1alpha1.RegistryFormatUpstream)
	if err == nil {
		t.Errorf("Expected error for empty data")
	}
}

func TestRegistryFormatConverter_Convert_UnsupportedFormat(t *testing.T) {
	converter := NewRegistryFormatConverter()
	
	testData := []byte(`{"version": "1.0.0", "servers": {}}`)
	
	_, err := converter.Convert(testData, "unsupported", mcpv1alpha1.RegistryFormatToolHive)
	if err == nil {
		t.Errorf("Expected error for unsupported source format")
	}
	
	_, err = converter.Convert(testData, mcpv1alpha1.RegistryFormatToolHive, "unsupported")
	if err == nil {
		t.Errorf("Expected error for unsupported target format")
	}
}

func TestRegistryFormatConverter_Validate_ToolHive(t *testing.T) {
	converter := NewRegistryFormatConverter()
	
	tests := []struct {
		name    string
		data    string
		wantErr bool
	}{
		{
			name: "valid ToolHive format",
			data: `{
				"version": "1.0.0",
				"last_updated": "2025-09-02T00:17:21Z",
				"servers": {
					"test-server": {
						"description": "Test server",
						"tier": "Official",
						"status": "Active",
						"transport": "stdio",
						"image": "test:latest"
					}
				}
			}`,
			wantErr: false,
		},
		{
			name: "missing version",
			data: `{
				"last_updated": "2025-09-02T00:17:21Z",
				"servers": {}
			}`,
			wantErr: true,
		},
		{
			name: "missing servers",
			data: `{
				"version": "1.0.0",
				"last_updated": "2025-09-02T00:17:21Z"
			}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			data:    `{invalid json}`,
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := converter.Validate([]byte(tt.data), mcpv1alpha1.RegistryFormatToolHive)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegistryFormatConverter_Validate_Upstream(t *testing.T) {
	converter := NewRegistryFormatConverter()
	
	tests := []struct {
		name    string
		data    string
		wantErr bool
	}{
		{
			name: "valid upstream format",
			data: `{
				"test-server": {
					"server": {
						"name": "test-server",
						"description": "Test server",
						"status": "active",
						"packages": [
							{
								"registry_name": "docker",
								"name": "test",
								"version": "latest"
							}
						]
					}
				}
			}`,
			wantErr: false,
		},
		{
			name:    "empty registry",
			data:    `{}`,
			wantErr: true,
		},
		{
			name: "server with empty name",
			data: `{
				"": {
					"server": {
						"name": "",
						"description": "Test server"
					}
				}
			}`,
			wantErr: true,
		},
		{
			name: "nil server detail",
			data: `{
				"test-server": null
			}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			data:    `{invalid json}`,
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := converter.Validate([]byte(tt.data), mcpv1alpha1.RegistryFormatUpstream)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegistryFormatConverter_DetectFormat(t *testing.T) {
	converter := NewRegistryFormatConverter()
	
	tests := []struct {
		name    string
		data    string
		want    string
		wantErr bool
	}{
		{
			name: "detect ToolHive format",
			data: `{
				"version": "1.0.0",
				"last_updated": "2025-09-02T00:17:21Z",
				"servers": {
					"test-server": {
						"description": "Test server",
						"image": "test:latest"
					}
				}
			}`,
			want:    mcpv1alpha1.RegistryFormatToolHive,
			wantErr: false,
		},
		{
			name: "detect upstream format",
			data: `{
				"test-server": {
					"server": {
						"name": "test-server",
						"description": "Test server",
						"packages": [
							{
								"registry_name": "docker",
								"name": "test",
								"version": "latest"
							}
						]
					}
				}
			}`,
			want:    mcpv1alpha1.RegistryFormatUpstream,
			wantErr: false,
		},
		{
			name:    "empty data",
			data:    ``,
			want:    "",
			wantErr: true,
		},
		{
			name:    "unknown format",
			data:    `{"unknown": "format"}`,
			want:    "",
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := converter.DetectFormat([]byte(tt.data))
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DetectFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistryFormatConverter_Convert_ToolHiveToUpstream(t *testing.T) {
	converter := NewRegistryFormatConverter()
	
	// Create a sample ToolHive registry
	toolhiveRegistry := &registry.Registry{
		Version:     "1.0.0",
		LastUpdated: "2025-09-02T00:17:21Z",
		Servers: map[string]*registry.ImageMetadata{
			"test-server": {
				BaseServerMetadata: registry.BaseServerMetadata{
					Name:        "test-server",
					Description: "Test server",
					Tier:        "Official",
					Status:      "Active",
					Transport:   "stdio",
					Tools:       []string{"test_tool"},
					Tags:        []string{"test"},
				},
				Image: "test:latest",
			},
		},
	}
	
	toolhiveData, err := json.Marshal(toolhiveRegistry)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}
	
	result, err := converter.Convert(toolhiveData, mcpv1alpha1.RegistryFormatToolHive, mcpv1alpha1.RegistryFormatUpstream)
	if err != nil {
		t.Errorf("Convert() error = %v", err)
		return
	}
	
	// Verify the result is valid upstream format
	err = converter.Validate(result, mcpv1alpha1.RegistryFormatUpstream)
	if err != nil {
		t.Errorf("Converted data is not valid upstream format: %v", err)
	}
	
	// Verify the structure
	var upstreamRegistry map[string]*registry.UpstreamServerDetail
	err = json.Unmarshal(result, &upstreamRegistry)
	if err != nil {
		t.Errorf("Failed to unmarshal converted data: %v", err)
		return
	}
	
	if _, exists := upstreamRegistry["test-server"]; !exists {
		t.Errorf("Expected server 'test-server' not found in converted registry")
	}
}

func TestNormalizeFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", mcpv1alpha1.RegistryFormatToolHive},
		{mcpv1alpha1.RegistryFormatToolHive, mcpv1alpha1.RegistryFormatToolHive},
		{mcpv1alpha1.RegistryFormatUpstream, mcpv1alpha1.RegistryFormatUpstream},
		{"custom", "custom"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeFormat(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeFormat(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsFormatSupported(t *testing.T) {
	supported := []string{mcpv1alpha1.RegistryFormatToolHive, mcpv1alpha1.RegistryFormatUpstream}
	
	tests := []struct {
		format   string
		expected bool
	}{
		{mcpv1alpha1.RegistryFormatToolHive, true},
		{mcpv1alpha1.RegistryFormatUpstream, true},
		{"unsupported", false},
		{"", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			result := isFormatSupported(tt.format, supported)
			if result != tt.expected {
				t.Errorf("isFormatSupported(%q) = %v, want %v", tt.format, result, tt.expected)
			}
		})
	}
}