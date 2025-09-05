package registryapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServerConfig_GetConfig(t *testing.T) {
	config := &ServerConfig{
		Port:                 8080,
		RegistryName:         "test-registry",
		RegistryNamespace:    "test-namespace",
		MetricsAddr:          ":8081",
		EnableLeaderElection: false,
	}

	// Note: This will fail in unit tests because we're not in a Kubernetes cluster
	// In real usage, this would return a valid config
	cfg, err := config.GetConfig()
	
	// We expect an error in unit tests since we're not in a cluster
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestServer_SetupRoutes(t *testing.T) {
	server := createTestServer()
	
	router := server.setupRoutes()
	require.NotNil(t, router)
	
	// Test that the routes are properly configured by making requests
	tests := []struct {
		method     string
		path       string
		expectCode int
	}{
		{http.MethodGet, "/health", http.StatusOK},
		{http.MethodGet, "/openapi.yaml", http.StatusOK},
		{http.MethodPost, "/health", http.StatusMethodNotAllowed},
		{http.MethodGet, "/nonexistent", http.StatusNotFound},
	}
	
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)
			
			assert.Equal(t, tt.expectCode, w.Code)
		})
	}
}

func TestServer_StartAndShutdown(t *testing.T) {
	server := createTestServer()
	server.config.Port = 0 // Use random available port
	
	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	
	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		err := server.Start(ctx)
		errChan <- err
	}()
	
	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)
	
	// Cancel the context to trigger shutdown
	cancel()
	
	// Wait for server to shut down
	select {
	case err := <-errChan:
		// Server should shut down cleanly without error
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Server did not shut down within timeout")
	}
}

func TestNewServer(t *testing.T) {
	t.Run("invalid config", func(t *testing.T) {
		// This will fail because we're not in a Kubernetes cluster
		config := &ServerConfig{
			Port:              8080,
			RegistryName:      "test-registry",
			RegistryNamespace: "test-namespace",
		}
		
		server, err := NewServer(config)
		assert.Error(t, err)
		assert.Nil(t, server)
		assert.Contains(t, err.Error(), "failed to get kubeconfig")
	})
}

func TestServerIntegration(t *testing.T) {
	// This test demonstrates how the server components work together
	server := createTestServer()
	mockClient := server.kubeClient.(*MockClient)
	mockConverter := server.formatConverter.(*MockFormatConverter)
	
	// Set up mocks for a successful health check
	mockClient.On("Get", mock.Anything, mock.Anything, mock.AnythingOfType("*v1alpha1.MCPRegistry"), mock.Anything).Return(nil)
	
	router := server.setupRoutes()
	
	// Test health endpoint
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
	
	// Test readiness endpoint
	req = httptest.NewRequest(http.MethodGet, "/readiness", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Ready", w.Body.String())
	
	// Test OpenAPI endpoint
	req = httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/yaml", w.Header().Get("Content-Type"))
	
	mockClient.AssertExpectations(t)
	mockConverter.AssertExpectations(t)
}