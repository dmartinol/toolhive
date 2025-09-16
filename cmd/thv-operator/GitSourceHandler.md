# GitSourceHandler Implementation Plan

This document outlines the implementation plan for adding Git repository support as a source for MCPRegistry data.

## Overview

The GitSourceHandler will enable MCPRegistry resources to fetch registry data directly from Git repositories, supporting both public and private repositories with various authentication methods.

## Functional Requirements

### Core Functionality
1. **Repository Access**: Support for Git repositories (GitHub, GitLab, Bitbucket, generic Git)
2. **Authentication**: Support multiple authentication methods (SSH keys, HTTPS tokens, basic auth)
3. **Branch/Tag Support**: Ability to specify specific branches, tags, or commit SHAs
4. **File Path Selection**: Specify the path to the registry file within the repository
5. **Change Detection**: Efficient detection of repository changes for sync triggering
6. **Caching**: Local caching for performance optimization
7. **Security**: Secure credential management via Kubernetes secrets

### Configuration Example
```yaml
apiVersion: toolhive.stacklok.dev/v1alpha1
kind: MCPRegistry
metadata:
  name: git-registry
spec:
  source:
    type: git
    format: toolhive
    git:
      repository: "https://github.com/example/mcp-registry.git"
      branch: "main"
      path: "registry.json"
      auth:
        type: token
        secretRef:
          name: git-credentials
          key: token
```

## Implementation Tasks

### Phase 1: API Design & CRD Updates

#### Task 1.1: Extend MCPRegistry CRD for Basic Git Sources
- **Description**: Add basic Git source configuration to the MCPRegistry CRD (without authentication)
- **Files to Create/Modify**:
  - `api/v1alpha1/mcpregistry_types.go` - Add GitSource struct and constants
  - `api/v1alpha1/zz_generated.deepcopy.go` - Regenerate deep copy methods
- **Key Components**:
  ```go
  const (
      RegistrySourceTypeGit = "git"
  )
  
  type GitSource struct {
      Repository string `json:"repository"`
      Branch     string `json:"branch,omitempty"`
      Tag        string `json:"tag,omitempty"`
      Commit     string `json:"commit,omitempty"`
      Path       string `json:"path,omitempty"`
      // Auth field will be added in Task 2.3
  }
  ```

#### Task 1.2: Update CRD Validation Rules for Basic Git Sources
- **Description**: Add validation for basic Git source configurations (authentication validation in Task 2.3)
- **Requirements**:
  - Repository URL validation (HTTP/HTTPS/SSH formats)
  - Mutual exclusivity of branch/tag/commit
  - Path validation (must end with .json)
  - Basic Git source type validation

### Phase 2: Core Implementation

#### Task 2.1: Create Git Client Wrapper
- **Description**: Create a thin wrapper around existing Git library (go-git/go-git)
- **Files to Create**:
  - `pkg/git/client.go` - Git client wrapper interface
  - `pkg/git/types.go` - Git-related types and configuration structs
  - `pkg/git/auth.go` - Authentication helpers for go-git
- **Dependencies**: 
  - `github.com/go-git/go-git/v5` - Pure Go Git implementation
  - `github.com/go-git/go-git/v5/plumbing/transport/http` - HTTP auth
  - `github.com/go-git/go-git/v5/plumbing/transport/ssh` - SSH auth
- **Key Components**:
  ```go
  type GitClient interface {
      Clone(ctx context.Context, config *CloneConfig) (*git.Repository, error)
      Pull(ctx context.Context, repo *git.Repository) error
      GetFileContent(repo *git.Repository, path string) ([]byte, error)
      GetCommitHash(repo *git.Repository) (string, error)
  }
  ```

#### Task 2.2: Implement GitSourceHandler
- **Description**: Create the main Git source handler implementing SourceHandler interface
- **Files to Create**:
  - `pkg/sources/git.go` - GitSourceHandler implementation
  - `pkg/sources/git_test.go` - Comprehensive unit tests
- **Key Methods**:
  - `FetchRegistry()` - Clone/pull repository and extract registry data
  - `CurrentHash()` - Get current commit hash for change detection
  - `Validate()` - Validate Git source configuration

#### Task 2.3: Authentication Management & CRD Extension
- **Description**: Implement secure authentication for Git repositories and extend CRD with Auth fields
- **Files to Create/Modify**:
  - `api/v1alpha1/mcpregistry_types.go` - Add GitAuth types and extend GitSource
  - `pkg/git/auth.go` - Authentication manager implementation
  - `pkg/git/types.go` - Authentication-related types
- **CRD Extensions**:
  ```go
  type GitSource struct {
      Repository string   `json:"repository"`
      Branch     string   `json:"branch,omitempty"`
      Tag        string   `json:"tag,omitempty"`
      Commit     string   `json:"commit,omitempty"`
      Path       string   `json:"path,omitempty"`
      Auth       *GitAuth `json:"auth,omitempty"` // Added in this task
  }
  
  type GitAuth struct {
      Type      string              `json:"type"`
      SecretRef *SecretKeyReference `json:"secretRef,omitempty"`
  }
  
  type SecretKeyReference struct {
      Name string `json:"name"`
      Key  string `json:"key"`
  }
  ```
- **Authentication Requirements**:
  - Support for Personal Access Tokens (GitHub, GitLab)
  - SSH key authentication
  - Basic authentication (username/password)
  - Integration with Kubernetes secrets
  - Secure credential management
- **CRD Validation Extensions**:
  - Authentication type validation (token, ssh-key, basic-auth, none)
  - SecretRef validation when auth is specified
  - Namespace security (secrets must be in same namespace)

#### Task 2.4: Repository Caching System
- **Description**: Implement local repository caching for performance
- **Requirements**:
  - Persistent local cache directory
  - Cache invalidation based on commit changes
  - Configurable cache size limits
  - Cleanup of stale repositories

### Phase 3: Integration & Factory Updates

#### Task 3.1: Update SourceHandlerFactory
- **Description**: Add Git handler support to the factory
- **Files to Modify**:
  - `pkg/sources/factory.go` - Add Git handler creation
  - `pkg/sources/factory_test.go` - Add Git handler tests

#### Task 3.2: Error Handling & Logging
- **Description**: Implement comprehensive error handling and logging
- **Requirements**:
  - Detailed error messages for common Git issues
  - Structured logging for debugging
  - Retry logic for transient network failures
  - Timeout configuration for Git operations

### Phase 4: Security & Configuration

#### Task 4.1: Security Hardening
- **Description**: Implement security best practices
- **Requirements**:
  - Validate repository URLs (prevent SSRF attacks)
  - Sandbox Git operations
  - Secure temporary file handling
  - Input validation and sanitization

#### Task 4.2: Configuration Management
- **Description**: Add configuration options for Git operations
- **Files to Create/Modify**:
  - Operator configuration for Git settings
  - Configurable timeouts, retry policies
  - Cache directory configuration
  - Resource limits for Git operations

### Phase 5: Testing & Documentation

#### Task 5.1: Comprehensive Testing
- **Description**: Create thorough test coverage
- **Test Types**:
  - Unit tests for all Git components
  - Integration tests with real Git repositories
  - Mock Git server for controlled testing
  - Security testing (malicious repositories)
  - Performance testing (large repositories)

#### Task 5.2: End-to-End Testing
- **Description**: E2E tests with real MCPRegistry resources
- **Requirements**:
  - Test with public repositories
  - Test with private repositories (authentication)
  - Test branch/tag switching
  - Test error scenarios and recovery

#### Task 5.3: Documentation
- **Description**: Create comprehensive documentation
- **Files to Create**:
  - `docs/git-source-handler.md` - User guide
  - Code documentation and examples
  - Security considerations document
  - Troubleshooting guide

### Phase 6: Advanced Features

#### Task 6.1: Webhook Support (Future)
- **Description**: Support for Git webhooks to trigger immediate syncs
- **Requirements**:
  - Webhook endpoint in operator
  - Signature verification
  - Rate limiting and security

#### Task 6.2: Submodule Support (Future)
- **Description**: Support for Git repositories with submodules
- **Requirements**:
  - Recursive submodule fetching
  - Authentication for submodules
  - Performance considerations

#### Task 6.3: Large File Support (Future)
- **Description**: Support for Git LFS (Large File Storage)
- **Requirements**:
  - LFS client integration
  - Bandwidth considerations
  - Storage management

## Technical Considerations

### Dependencies
- **Git Library**: `github.com/go-git/go-git/v5` (pure Go, no CGO dependencies)
  - Advantages: Pure Go, easy deployment, good performance
  - Authentication support for HTTP(S) and SSH
  - In-memory and filesystem storage options
- **Authentication**: Kubernetes client for secret management
- **File System**: Secure temporary directory management  
- **Networking**: Built-in HTTP client from go-git with timeout support

### Performance Considerations
- **Caching Strategy**: Local repository cache with configurable retention
- **Incremental Updates**: Use Git's incremental fetch capabilities
- **Concurrent Operations**: Proper synchronization for parallel Git operations
- **Resource Limits**: Memory and disk usage limits for Git operations

### Security Considerations
- **Repository Validation**: Prevent access to internal/private networks
- **Credential Security**: Secure handling of authentication secrets
- **File System Security**: Sandboxed Git operations
- **Input Validation**: Thorough validation of all user inputs

## Success Criteria

- [ ] Git repositories can be used as MCPRegistry sources
- [ ] Support for public and private repositories
- [ ] Multiple authentication methods working
- [ ] Efficient change detection and caching
- [ ] Comprehensive error handling and logging
- [ ] Full test coverage (>90%)
- [ ] Security review passed
- [ ] Documentation complete
- [ ] Performance benchmarks met

## Estimated Timeline

- **Phase 1**: API Design & CRD Updates - 2-3 days
- **Phase 2**: Core Implementation - 5-7 days (using go-git library)
- **Phase 3**: Integration & Factory Updates - 2-3 days
- **Phase 4**: Security & Configuration - 3-4 days
- **Phase 5**: Testing & Documentation - 5-7 days
- **Phase 6**: Advanced Features - 7-10 days (optional)

**Total Estimated Time**: 3-4 weeks for core functionality, additional 1-2 weeks for advanced features

## Dependencies & Prerequisites

- Existing SourceHandler interface and factory pattern
- Kubernetes secret management system
- Registry data validation framework
- Comprehensive testing infrastructure
- Security review process