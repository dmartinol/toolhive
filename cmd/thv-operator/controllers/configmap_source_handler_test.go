package controllers

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	mcpv1alpha1 "github.com/stacklok/toolhive/cmd/thv-operator/api/v1alpha1"
)

func TestConfigMapSourceHandler_Validate(t *testing.T) {
	handler := NewConfigMapSourceHandler(nil)

	tests := []struct {
		name    string
		source  mcpv1alpha1.MCPRegistrySource
		wantErr bool
	}{
		{
			name: "valid configmap source",
			source: mcpv1alpha1.MCPRegistrySource{
				Type: "configmap",
				ConfigMap: &mcpv1alpha1.ConfigMapRegistrySource{
					Name: "test-registry",
				},
			},
			wantErr: false,
		},
		{
			name: "missing configmap config",
			source: mcpv1alpha1.MCPRegistrySource{
				Type: "configmap",
			},
			wantErr: true,
		},
		{
			name: "missing configmap name",
			source: mcpv1alpha1.MCPRegistrySource{
				Type: "configmap",
				ConfigMap: &mcpv1alpha1.ConfigMapRegistrySource{},
			},
			wantErr: true,
		},
		{
			name: "wrong source type",
			source: mcpv1alpha1.MCPRegistrySource{
				Type: "url",
				ConfigMap: &mcpv1alpha1.ConfigMapRegistrySource{
					Name: "test-registry",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.Validate(&tt.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigMapSourceHandler_Sync(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = mcpv1alpha1.AddToScheme(scheme)

	validRegistryData := `{
		"version": "1.0.0",
		"last_updated": "2025-09-02T00:17:21Z",
		"servers": {
			"test-server": {
				"description": "Test MCP server",
				"tier": "Official",
				"status": "Active",
				"transport": "stdio",
				"image": "test/server:latest",
				"tools": ["test_tool"]
			}
		}
	}`

	tests := []struct {
		name        string
		registry    *mcpv1alpha1.MCPRegistry
		configMaps  []corev1.ConfigMap
		wantErr     bool
		wantServers int32
	}{
		{
			name: "successful sync",
			registry: &mcpv1alpha1.MCPRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-registry",
					Namespace: "default",
				},
				Spec: mcpv1alpha1.MCPRegistrySpec{
					Source: mcpv1alpha1.MCPRegistrySource{
						Type: "configmap",
						ConfigMap: &mcpv1alpha1.ConfigMapRegistrySource{
							Name: "test-registry-data",
							Key:  "registry.json",
						},
					},
				},
			},
			configMaps: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-registry-data",
						Namespace: "default",
					},
					Data: map[string]string{
						"registry.json": validRegistryData,
					},
				},
			},
			wantErr:     false,
			wantServers: 1,
		},
		{
			name: "configmap not found",
			registry: &mcpv1alpha1.MCPRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-registry",
					Namespace: "default",
				},
				Spec: mcpv1alpha1.MCPRegistrySpec{
					Source: mcpv1alpha1.MCPRegistrySource{
						Type: "configmap",
						ConfigMap: &mcpv1alpha1.ConfigMapRegistrySource{
							Name: "missing-registry",
							Key:  "registry.json",
						},
					},
				},
			},
			configMaps: []corev1.ConfigMap{},
			wantErr:    true,
		},
		{
			name: "key not found in configmap",
			registry: &mcpv1alpha1.MCPRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-registry",
					Namespace: "default",
				},
				Spec: mcpv1alpha1.MCPRegistrySpec{
					Source: mcpv1alpha1.MCPRegistrySource{
						Type: "configmap",
						ConfigMap: &mcpv1alpha1.ConfigMapRegistrySource{
							Name: "test-registry-data",
							Key:  "missing-key.json",
						},
					},
				},
			},
			configMaps: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-registry-data",
						Namespace: "default",
					},
					Data: map[string]string{
						"registry.json": validRegistryData,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid json data",
			registry: &mcpv1alpha1.MCPRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-registry",
					Namespace: "default",
				},
				Spec: mcpv1alpha1.MCPRegistrySpec{
					Source: mcpv1alpha1.MCPRegistrySource{
						Type: "configmap",
						ConfigMap: &mcpv1alpha1.ConfigMapRegistrySource{
							Name: "test-registry-data",
							Key:  "registry.json",
						},
					},
				},
			},
			configMaps: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-registry-data",
						Namespace: "default",
					},
					Data: map[string]string{
						"registry.json": "invalid json data",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client with test data
			objects := make([]runtime.Object, len(tt.configMaps))
			for i, cm := range tt.configMaps {
				objects[i] = &cm
			}
			
			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(objects...).
				Build()

			handler := NewConfigMapSourceHandler(client)

			result, err := handler.Sync(context.TODO(), tt.registry)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Sync() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != nil {
				if result.ServerCount != tt.wantServers {
					t.Errorf("Sync() serverCount = %v, want %v", result.ServerCount, tt.wantServers)
				}
				if result.Hash == "" {
					t.Error("Sync() hash should not be empty")
				}
				if len(result.Data) == 0 {
					t.Error("Sync() data should not be empty")
				}
			}
		})
	}
}

func TestConfigMapSourceHandler_GetConfigMapReference(t *testing.T) {
	handler := NewConfigMapSourceHandler(nil)

	tests := []struct {
		name      string
		registry  *mcpv1alpha1.MCPRegistry
		wantName  string
		wantNS    string
		wantKey   string
		wantErr   bool
	}{
		{
			name: "configmap in same namespace",
			registry: &mcpv1alpha1.MCPRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-registry",
					Namespace: "test-ns",
				},
				Spec: mcpv1alpha1.MCPRegistrySpec{
					Source: mcpv1alpha1.MCPRegistrySource{
						Type: "configmap",
						ConfigMap: &mcpv1alpha1.ConfigMapRegistrySource{
							Name: "registry-data",
							Key:  "data.json",
						},
					},
				},
			},
			wantName: "registry-data",
			wantNS:   "test-ns",
			wantKey:  "data.json",
			wantErr:  false,
		},
		{
			name: "configmap in different namespace",
			registry: &mcpv1alpha1.MCPRegistry{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-registry",
					Namespace: "test-ns",
				},
				Spec: mcpv1alpha1.MCPRegistrySpec{
					Source: mcpv1alpha1.MCPRegistrySource{
						Type: "configmap",
						ConfigMap: &mcpv1alpha1.ConfigMapRegistrySource{
							Name:      "registry-data",
							Namespace: "other-ns",
							Key:       "data.json",
						},
					},
				},
			},
			wantName: "registry-data",
			wantNS:   "other-ns",
			wantKey:  "data.json",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			namespacedName, key, err := handler.GetConfigMapReference(tt.registry)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfigMapReference() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if namespacedName.Name != tt.wantName {
					t.Errorf("GetConfigMapReference() name = %v, want %v", namespacedName.Name, tt.wantName)
				}
				if namespacedName.Namespace != tt.wantNS {
					t.Errorf("GetConfigMapReference() namespace = %v, want %v", namespacedName.Namespace, tt.wantNS)
				}
				if key != tt.wantKey {
					t.Errorf("GetConfigMapReference() key = %v, want %v", key, tt.wantKey)
				}
			}
		})
	}
}