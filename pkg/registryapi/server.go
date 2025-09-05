package registryapi

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	mcpv1alpha1 "github.com/stacklok/toolhive/cmd/thv-operator/api/v1alpha1"
	"github.com/stacklok/toolhive/cmd/thv-operator/controllers"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

//go:embed ../../../cmd/thv-registry-api/openapi.yaml
var openAPISpec []byte

// ServerConfig holds the configuration for the registry API server
type ServerConfig struct {
	Port                 int
	RegistryName         string
	RegistryNamespace    string
	MetricsAddr          string
	EnableLeaderElection bool
}

// Server represents the registry API server
type Server struct {
	config           *ServerConfig
	kubeClient       client.Client
	formatConverter  controllers.FormatConverter
	httpServer       *http.Server
}

// GetConfig returns the Kubernetes client configuration
func (c *ServerConfig) GetConfig() (*rest.Config, error) {
	return config.GetConfig()
}

// NewServer creates a new registry API server
func NewServer(config *ServerConfig) (*Server, error) {
	// Get Kubernetes client configuration
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	// Create Kubernetes client
	kubeClient, err := client.New(cfg, client.Options{
		Scheme: mcpv1alpha1.GetScheme(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Create format converter
	formatConverter := controllers.NewRegistryFormatConverter()

	server := &Server{
		config:          config,
		kubeClient:      kubeClient,
		formatConverter: formatConverter,
	}

	return server, nil
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	logger := log.FromContext(ctx).WithName("registry-api-server")

	// Create router
	r := s.setupRoutes()

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		logger.Info("Starting HTTP server", "port", s.config.Port)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrChan <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		logger.Info("Shutting down HTTP server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(shutdownCtx)
	case err := <-serverErrChan:
		return fmt.Errorf("HTTP server error: %w", err)
	}
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Health check
	r.Get("/health", s.handleHealth)
	r.Get("/readiness", s.handleReadiness)
	
	// OpenAPI documentation
	r.Get("/openapi.yaml", s.handleOpenAPI)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/registry", func(r chi.Router) {
			r.Get("/info", s.handleRegistryInfo)
			r.Get("/servers", s.handleListServers)
			r.Get("/servers/{name}", s.handleGetServer)
		})
	})

	return r
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleReadiness handles readiness check requests
func (s *Server) handleReadiness(w http.ResponseWriter, r *http.Request) {
	// Check if we can connect to Kubernetes API
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	registry := &mcpv1alpha1.MCPRegistry{}
	err := s.kubeClient.Get(ctx, client.ObjectKey{
		Name:      s.config.RegistryName,
		Namespace: s.config.RegistryNamespace,
	}, registry)

	if err != nil {
		http.Error(w, "Not ready: cannot access registry", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}

// handleOpenAPI serves the OpenAPI specification
func (s *Server) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	w.Write(openAPISpec)
}