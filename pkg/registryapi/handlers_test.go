package registryapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	mcpv1alpha1 "github.com/stacklok/toolhive/cmd/thv-operator/api/v1alpha1"
	"github.com/stacklok/toolhive/cmd/thv-operator/controllers"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MockClient implements a mock Kubernetes client for testing
type MockClient struct {
	mock.Mock
}

func (m *MockClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	args := m.Called(ctx, key, obj, opts)
	
	// If the test provides a registry object, populate it
	if registry, ok := obj.(*mcpv1alpha1.MCPRegistry); ok && args.Get(0) == nil {
		testRegistry := createTestRegistry()
		*registry = *testRegistry
	}
	
	// If the test provides a configmap object, populate it
	if configMap, ok := obj.(*corev1.ConfigMap); ok && args.Get(0) == nil {
		testConfigMap := createTestConfigMap()
		*configMap = *testConfigMap
	}
	
	return args.Error(0)
}

func (m *MockClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	args := m.Called(ctx, list, opts)
	return args.Error(0)
}

func (m *MockClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	args := m.Called(ctx, obj, opts)
	return args.Error(0)
}

func (m *MockClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	args := m.Called(ctx, obj, opts)
	return args.Error(0)
}

func (m *MockClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	args := m.Called(ctx, obj, opts)
	return args.Error(0)
}

func (m *MockClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	args := m.Called(ctx, obj, patch, opts)
	return args.Error(0)
}

func (m *MockClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	args := m.Called(ctx, obj, opts)
	return args.Error(0)
}

func (m *MockClient) Status() client.StatusWriter {
	args := m.Called()
	return args.Get(0).(client.StatusWriter)
}

func (m *MockClient) Scheme() *runtime.Scheme {
	args := m.Called()
	return args.Get(0).(*runtime.Scheme)
}

func (m *MockClient) RESTMapper() client.RESTMapper {
	args := m.Called()
	return args.Get(0).(client.RESTMapper)
}

func (m *MockClient) GroupVersionKindFor(obj runtime.Object) (*runtime.GroupVersionKind, error) {
	args := m.Called(obj)
	return args.Get(0).(*runtime.GroupVersionKind), args.Error(1)
}

func (m *MockClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	args := m.Called(obj)
	return args.Bool(0), args.Error(1)
}

// MockFormatConverter implements a mock format converter for testing
type MockFormatConverter struct {
	mock.Mock
}

func (m *MockFormatConverter) SupportedFormats() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockFormatConverter) DetectFormat(data []byte) (string, error) {
	args := m.Called(data)
	return args.String(0), args.Error(1)
}

func (m *MockFormatConverter) Convert(data []byte, fromFormat, toFormat string) ([]byte, error) {
	args := m.Called(data, fromFormat, toFormat)
	return args.Get(0).([]byte), args.Error(1)
}

// Test helper functions
func createTestRegistry() *mcpv1alpha1.MCPRegistry {
	now := metav1.NewTime(time.Now())
	return &mcpv1alpha1.MCPRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-registry",
			Namespace: "test-namespace",
		},
		Spec: mcpv1alpha1.MCPRegistrySpec{
			DisplayName: "Test Registry",
			Source: mcpv1alpha1.MCPRegistrySource{
				Type:   mcpv1alpha1.RegistrySourceTypeConfigMap,
				Format: mcpv1alpha1.RegistryFormatToolHive,
				ConfigMap: &mcpv1alpha1.ConfigMapRegistrySource{
					Name: "test-configmap",
				},
			},
		},
		Status: mcpv1alpha1.MCPRegistryStatus{
			Phase:        mcpv1alpha1.MCPRegistryPhaseReady,
			ServerCount:  2,
			LastSyncTime: &now,
			LastSyncHash: "test-hash",
			Message:      "Successfully synced 2 servers",
			StorageRef: &mcpv1alpha1.StorageReference{
				Type: "configmap",
				ConfigMapRef: &mcpv1alpha1.ConfigMapReference{
					Name:      "test-registry-storage",
					Namespace: "test-namespace",
					Key:       "data",
				},
			},
			ApiEndpoint: "http://test-registry-api.test-namespace.svc.cluster.local:80",
		},
	}
}

func createTestConfigMap() *corev1.ConfigMap {
	registryData := map[string]interface{}{
		"servers": map[string]interface{}{
			"filesystem": map[string]interface{}{
				"name":        "filesystem",
				"description": "Filesystem operations server",
				"docker": map[string]interface{}{
					"image": "mcp/filesystem:latest",
				},
			},
			"git": map[string]interface{}{
				"name":        "git",
				"description": "Git operations server",
				"docker": map[string]interface{}{
					"image": "mcp/git:latest",
				},
			},
		},
	}
	
	data, _ := json.Marshal(registryData)
	
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-registry-storage",
			Namespace: "test-namespace",
		},
		Data: map[string]string{
			controllers.StorageConfigMapKeyData: string(data),
		},
	}
}

func createTestServer() *Server {
	mockClient := &MockClient{}
	mockConverter := &MockFormatConverter{}
	
	return &Server{
		config: &ServerConfig{
			RegistryName:      "test-registry",
			RegistryNamespace: "test-namespace",
		},
		kubeClient:      mockClient,
		formatConverter: mockConverter,
	}
}

func TestHandleRegistryInfo(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*MockClient, *MockFormatConverter)
		expectedStatus int
		expectedBody   func(t *testing.T, body []byte)
	}{
		{
			name: "successful registry info",
			setupMocks: func(mc *MockClient, mfc *MockFormatConverter) {
				mc.On("Get", mock.Anything, mock.Anything, mock.AnythingOfType("*v1alpha1.MCPRegistry"), mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body []byte) {
				var info RegistryInfo
				require.NoError(t, json.Unmarshal(body, &info))
				assert.Equal(t, "test-registry", info.Name)
				assert.Equal(t, "Test Registry", info.DisplayName)
				assert.Equal(t, mcpv1alpha1.RegistryFormatToolHive, info.Format)
				assert.NotNil(t, info.Source)
				assert.Equal(t, mcpv1alpha1.RegistrySourceTypeConfigMap, info.Source.Type)
				assert.NotNil(t, info.Status)
				assert.Equal(t, "Ready", string(info.Status.Phase))
				assert.Equal(t, int32(2), info.Status.ServerCount)
			},
		},
		{
			name: "registry not found",
			setupMocks: func(mc *MockClient, mfc *MockFormatConverter) {
				mc.On("Get", mock.Anything, mock.Anything, mock.AnythingOfType("*v1alpha1.MCPRegistry"), mock.Anything).Return(assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: func(t *testing.T, body []byte) {
				var errorResp ErrorResponse
				require.NoError(t, json.Unmarshal(body, &errorResp))
				assert.Equal(t, "Registry not found", errorResp.Message)
				assert.Equal(t, http.StatusNotFound, errorResp.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServer()
			mockClient := server.kubeClient.(*MockClient)
			mockConverter := server.formatConverter.(*MockFormatConverter)
			
			tt.setupMocks(mockClient, mockConverter)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/registry/info", nil)
			w := httptest.NewRecorder()

			server.handleRegistryInfo(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.expectedBody(t, w.Body.Bytes())

			mockClient.AssertExpectations(t)
			mockConverter.AssertExpectations(t)
		})
	}
}

func TestHandleListServers(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setupMocks     func(*MockClient, *MockFormatConverter)
		expectedStatus int
		expectedBody   func(t *testing.T, body []byte)
	}{
		{
			name:        "successful server list - default format",
			queryParams: "",
			setupMocks: func(mc *MockClient, mfc *MockFormatConverter) {
				mc.On("Get", mock.Anything, mock.Anything, mock.AnythingOfType("*v1alpha1.MCPRegistry"), mock.Anything).Return(nil)
				mc.On("Get", mock.Anything, mock.Anything, mock.AnythingOfType("*v1.ConfigMap"), mock.Anything).Return(nil)
				mfc.On("SupportedFormats").Return([]string{mcpv1alpha1.RegistryFormatToolHive, mcpv1alpha1.RegistryFormatUpstream})
				mfc.On("DetectFormat", mock.Anything).Return(mcpv1alpha1.RegistryFormatToolHive, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body []byte) {
				var response ServerListResponse
				require.NoError(t, json.Unmarshal(body, &response))
				assert.Equal(t, 2, response.Count)
				assert.Equal(t, mcpv1alpha1.RegistryFormatToolHive, response.Format)
				assert.Contains(t, response.Servers, "filesystem")
				assert.Contains(t, response.Servers, "git")
			},
		},
		{
			name:        "successful server list - upstream format",
			queryParams: "?format=upstream",
			setupMocks: func(mc *MockClient, mfc *MockFormatConverter) {
				mc.On("Get", mock.Anything, mock.Anything, mock.AnythingOfType("*v1alpha1.MCPRegistry"), mock.Anything).Return(nil)
				mc.On("Get", mock.Anything, mock.Anything, mock.AnythingOfType("*v1.ConfigMap"), mock.Anything).Return(nil)
				mfc.On("SupportedFormats").Return([]string{mcpv1alpha1.RegistryFormatToolHive, mcpv1alpha1.RegistryFormatUpstream})
				mfc.On("DetectFormat", mock.Anything).Return(mcpv1alpha1.RegistryFormatToolHive, nil)
				mfc.On("Convert", mock.Anything, mcpv1alpha1.RegistryFormatToolHive, mcpv1alpha1.RegistryFormatUpstream).Return(func(data []byte, from, to string) []byte {
					// Return the same data for test purposes
					return data
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body []byte) {
				var response ServerListResponse
				require.NoError(t, json.Unmarshal(body, &response))
				assert.Equal(t, 2, response.Count)
				assert.Equal(t, mcpv1alpha1.RegistryFormatUpstream, response.Format)
			},
		},
		{
			name:        "unsupported format",
			queryParams: "?format=invalid",
			setupMocks: func(mc *MockClient, mfc *MockFormatConverter) {
				mfc.On("SupportedFormats").Return([]string{mcpv1alpha1.RegistryFormatToolHive, mcpv1alpha1.RegistryFormatUpstream})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body []byte) {
				var errorResp ErrorResponse
				require.NoError(t, json.Unmarshal(body, &errorResp))
				assert.Contains(t, errorResp.Message, "Unsupported format: invalid")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServer()
			mockClient := server.kubeClient.(*MockClient)
			mockConverter := server.formatConverter.(*MockFormatConverter)
			
			tt.setupMocks(mockClient, mockConverter)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/registry/servers"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			server.handleListServers(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.expectedBody(t, w.Body.Bytes())

			mockClient.AssertExpectations(t)
			mockConverter.AssertExpectations(t)
		})
	}
}

func TestHandleGetServer(t *testing.T) {
	tests := []struct {
		name           string
		serverName     string
		queryParams    string
		setupMocks     func(*MockClient, *MockFormatConverter)
		expectedStatus int
		expectedBody   func(t *testing.T, body []byte)
	}{
		{
			name:        "successful get server",
			serverName:  "filesystem",
			queryParams: "",
			setupMocks: func(mc *MockClient, mfc *MockFormatConverter) {
				mc.On("Get", mock.Anything, mock.Anything, mock.AnythingOfType("*v1alpha1.MCPRegistry"), mock.Anything).Return(nil)
				mc.On("Get", mock.Anything, mock.Anything, mock.AnythingOfType("*v1.ConfigMap"), mock.Anything).Return(nil)
				mfc.On("SupportedFormats").Return([]string{mcpv1alpha1.RegistryFormatToolHive, mcpv1alpha1.RegistryFormatUpstream})
				mfc.On("DetectFormat", mock.Anything).Return(mcpv1alpha1.RegistryFormatToolHive, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body []byte) {
				var response ServerResponse
				require.NoError(t, json.Unmarshal(body, &response))
				assert.Equal(t, "filesystem", response.Name)
				assert.Equal(t, mcpv1alpha1.RegistryFormatToolHive, response.Format)
				assert.NotNil(t, response.Server)
			},
		},
		{
			name:        "server not found",
			serverName:  "nonexistent",
			queryParams: "",
			setupMocks: func(mc *MockClient, mfc *MockFormatConverter) {
				mc.On("Get", mock.Anything, mock.Anything, mock.AnythingOfType("*v1alpha1.MCPRegistry"), mock.Anything).Return(nil)
				mc.On("Get", mock.Anything, mock.Anything, mock.AnythingOfType("*v1.ConfigMap"), mock.Anything).Return(nil)
				mfc.On("SupportedFormats").Return([]string{mcpv1alpha1.RegistryFormatToolHive, mcpv1alpha1.RegistryFormatUpstream})
				mfc.On("DetectFormat", mock.Anything).Return(mcpv1alpha1.RegistryFormatToolHive, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: func(t *testing.T, body []byte) {
				var errorResp ErrorResponse
				require.NoError(t, json.Unmarshal(body, &errorResp))
				assert.Contains(t, errorResp.Message, "Server 'nonexistent' not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServer()
			mockClient := server.kubeClient.(*MockClient)
			mockConverter := server.formatConverter.(*MockFormatConverter)
			
			tt.setupMocks(mockClient, mockConverter)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/registry/servers/"+tt.serverName+tt.queryParams, nil)
			w := httptest.NewRecorder()

			// Set up chi router context with URL parameters
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("name", tt.serverName)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			server.handleGetServer(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.expectedBody(t, w.Body.Bytes())

			mockClient.AssertExpectations(t)
			mockConverter.AssertExpectations(t)
		})
	}
}

func TestHealthAndReadiness(t *testing.T) {
	t.Run("health check", func(t *testing.T) {
		server := createTestServer()
		
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		server.handleHealth(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "OK", w.Body.String())
	})

	t.Run("readiness check - success", func(t *testing.T) {
		server := createTestServer()
		mockClient := server.kubeClient.(*MockClient)
		
		mockClient.On("Get", mock.Anything, mock.Anything, mock.AnythingOfType("*v1alpha1.MCPRegistry"), mock.Anything).Return(nil)

		req := httptest.NewRequest(http.MethodGet, "/readiness", nil)
		w := httptest.NewRecorder()

		server.handleReadiness(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Ready", w.Body.String())
		mockClient.AssertExpectations(t)
	})

	t.Run("readiness check - failure", func(t *testing.T) {
		server := createTestServer()
		mockClient := server.kubeClient.(*MockClient)
		
		mockClient.On("Get", mock.Anything, mock.Anything, mock.AnythingOfType("*v1alpha1.MCPRegistry"), mock.Anything).Return(assert.AnError)

		req := httptest.NewRequest(http.MethodGet, "/readiness", nil)
		w := httptest.NewRecorder()

		server.handleReadiness(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Contains(t, w.Body.String(), "Not ready: cannot access registry")
		mockClient.AssertExpectations(t)
	})
}

func TestOpenAPIEndpoint(t *testing.T) {
	server := createTestServer()
	
	req := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	w := httptest.NewRecorder()

	server.handleOpenAPI(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/yaml", w.Header().Get("Content-Type"))
	assert.NotEmpty(t, w.Body.String())
}