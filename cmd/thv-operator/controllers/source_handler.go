package controllers

import (
	"context"
	"fmt"
	"time"

	mcpv1alpha1 "github.com/stacklok/toolhive/cmd/thv-operator/api/v1alpha1"
)

// SourceHandler defines the interface for handling different registry source types
type SourceHandler interface {
	// Sync retrieves and processes data from the source
	Sync(ctx context.Context, registry *mcpv1alpha1.MCPRegistry) (*SyncResult, error)
	
	// Validate validates the source configuration
	Validate(source *mcpv1alpha1.MCPRegistrySource) error
	
	// GetSourceType returns the source type this handler supports
	GetSourceType() string
}

// SyncResult contains the result of a sync operation
type SyncResult struct {
	// Data is the raw registry data
	Data []byte
	
	// Hash is a hash of the data for change detection
	Hash string
	
	// ServerCount is the number of servers in the registry
	ServerCount int32
	
	// LastModified is when the source data was last modified
	LastModified time.Time
	
	// Format is the detected format of the data
	Format string
}

// RegistryData represents the structure of registry data
type RegistryData struct {
	Schema      string            `json:"$schema,omitempty"`
	Version     string            `json:"version"`
	LastUpdated time.Time         `json:"last_updated"`
	Servers     map[string]Server `json:"servers"`
}

// Server represents an MCP server in the registry
type Server struct {
	Description   string                 `json:"description"`
	Tier          string                 `json:"tier"`
	Status        string                 `json:"status"`
	Transport     string                 `json:"transport"`
	Tools         []string               `json:"tools,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	RepositoryURL string                 `json:"repository_url,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	Image         string                 `json:"image"`
	Command       []string               `json:"command,omitempty"`
	Args          []string               `json:"args,omitempty"`
	Environments  []Environment          `json:"environments,omitempty"`
}

// Environment represents an environment configuration for a server
type Environment struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Config      EnvConfig   `json:"config"`
}

// EnvConfig represents environment-specific configuration
type EnvConfig struct {
	Env  map[string]string `json:"env,omitempty"`
	Args []string          `json:"args,omitempty"`
}

// SourceHandlerError represents errors that occur during source handling
type SourceHandlerError struct {
	SourceType string
	Operation  string
	Reason     string
	Err        error
}

func (e *SourceHandlerError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s source %s failed: %s: %v", e.SourceType, e.Operation, e.Reason, e.Err)
	}
	return fmt.Sprintf("%s source %s failed: %s", e.SourceType, e.Operation, e.Reason)
}

func (e *SourceHandlerError) Unwrap() error {
	return e.Err
}

// Common error types
var (
	ErrSourceNotFound     = fmt.Errorf("source not found")
	ErrInvalidData        = fmt.Errorf("invalid data format")
	ErrValidationFailed   = fmt.Errorf("validation failed")
	ErrUnsupportedFormat  = fmt.Errorf("unsupported format")
)

// NewSourceHandlerError creates a new source handler error
func NewSourceHandlerError(sourceType, operation, reason string, err error) *SourceHandlerError {
	return &SourceHandlerError{
		SourceType: sourceType,
		Operation:  operation,
		Reason:     reason,
		Err:        err,
	}
}