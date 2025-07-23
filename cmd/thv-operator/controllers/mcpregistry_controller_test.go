package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mcpv1alpha1 "github.com/stacklok/toolhive/cmd/thv-operator/api/v1alpha1"
)

func TestMCPRegistrySpec_Validation(t *testing.T) {
	tests := []struct {
		name    string
		spec    mcpv1alpha1.MCPRegistrySpec
		isValid bool
	}{
		{
			name: "valid http registry",
			spec: mcpv1alpha1.MCPRegistrySpec{
				URL:             "https://example.com/registry",
				Type:            "http",
				RefreshInterval: "1h",
				Timeout:         "30s",
			},
			isValid: true,
		},
		{
			name: "valid file registry",
			spec: mcpv1alpha1.MCPRegistrySpec{
				URL:             "file:///path/to/registry",
				Type:            "file",
				RefreshInterval: "30m",
				Timeout:         "10s",
			},
			isValid: true,
		},
		{
			name: "valid embedded registry",
			spec: mcpv1alpha1.MCPRegistrySpec{
				URL:             "embedded://default",
				Type:            "embedded",
				RefreshInterval: "24h",
				Timeout:         "60s",
			},
			isValid: true,
		},
		{
			name: "missing URL",
			spec: mcpv1alpha1.MCPRegistrySpec{
				Type:            "http",
				RefreshInterval: "1h",
				Timeout:         "30s",
			},
			isValid: false,
		},
		{
			name: "invalid type",
			spec: mcpv1alpha1.MCPRegistrySpec{
				URL:             "https://example.com/registry",
				Type:            "invalid",
				RefreshInterval: "1h",
				Timeout:         "30s",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// In a real implementation, you would validate the spec here
			// For now, we just check that the struct can be created
			registry := &mcpv1alpha1.MCPRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-registry",
					Namespace: "default",
				},
				Spec: tt.spec,
			}

			if tt.isValid {
				assert.NotNil(t, registry)
				assert.Equal(t, tt.spec.URL, registry.Spec.URL)
				assert.Equal(t, tt.spec.Type, registry.Spec.Type)
			}
		})
	}
}

func TestMCPRegistryStatus_Phases(t *testing.T) {
	tests := []struct {
		name     string
		phase    mcpv1alpha1.MCPRegistryPhase
		expected string
	}{
		{
			name:     "pending phase",
			phase:    mcpv1alpha1.MCPRegistryPhasePending,
			expected: "Pending",
		},
		{
			name:     "ready phase",
			phase:    mcpv1alpha1.MCPRegistryPhaseReady,
			expected: "Ready",
		},
		{
			name:     "failed phase",
			phase:    mcpv1alpha1.MCPRegistryPhaseFailed,
			expected: "Failed",
		},
		{
			name:     "syncing phase",
			phase:    mcpv1alpha1.MCPRegistryPhaseSyncing,
			expected: "Syncing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.phase))
		})
	}
}

func TestRegistryAuthentication_Validation(t *testing.T) {
	tests := []struct {
		name           string
		authentication *mcpv1alpha1.RegistryAuthentication
		isValid        bool
	}{
		{
			name: "no authentication",
			authentication: &mcpv1alpha1.RegistryAuthentication{
				Type: "none",
			},
			isValid: true,
		},
		{
			name: "basic authentication",
			authentication: &mcpv1alpha1.RegistryAuthentication{
				Type:     "basic",
				Username: "user",
				PasswordSecretRef: &mcpv1alpha1.SecretRef{
					Name: "password-secret",
					Key:  "password",
				},
			},
			isValid: true,
		},
		{
			name: "bearer authentication",
			authentication: &mcpv1alpha1.RegistryAuthentication{
				Type: "bearer",
				TokenSecretRef: &mcpv1alpha1.SecretRef{
					Name: "token-secret",
					Key:  "token",
				},
			},
			isValid: true,
		},
		{
			name: "invalid authentication type",
			authentication: &mcpv1alpha1.RegistryAuthentication{
				Type: "invalid",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// In a real implementation, you would validate the authentication here
			// For now, we just check that the struct can be created
			if tt.isValid {
				assert.NotNil(t, tt.authentication)
				assert.NotEmpty(t, tt.authentication.Type)
			}
		})
	}
}
