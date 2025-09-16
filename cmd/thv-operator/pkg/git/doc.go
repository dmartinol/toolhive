// Package git provides Git repository operations for MCPRegistry sources.
//
// This package implements a thin wrapper around the go-git library to enable
// MCPRegistry resources to fetch registry data directly from Git repositories.
// It supports cloning repositories, checking out specific branches/tags/commits,
// and retrieving file contents from the repository.
//
// Key Components:
//
// # Client Interface
//
// The Client interface defines the core Git operations:
//   - Clone: Clone repositories with authentication support
//   - Pull: Update existing repositories (planned for future implementation)
//   - GetFileContent: Retrieve specific files from repositories
//   - GetCommitHash: Get current commit hash for change detection
//   - Cleanup: Remove local repository directories
//
// # Authentication
//
// The package includes authentication management for various Git hosting services:
//   - None: Public repositories (currently implemented)
//   - Token: Personal access tokens for GitHub/GitLab (planned)
//   - SSH: SSH key authentication (planned)
//   - Basic: Username/password authentication (planned)
//
// # Example Usage
//
//	client := git.NewDefaultGitClient()
//	config := &git.CloneConfig{
//	    URL:       "https://github.com/example/registry.git",
//	    Branch:    "main",
//	    Directory: "/tmp/repo",
//	    Auth: git.AuthConfig{
//	        Type: git.AuthTypeNone,
//	    },
//	}
//
//	repoInfo, err := client.Clone(ctx, config)
//	if err != nil {
//	    return err
//	}
//	defer client.Cleanup(repoInfo)
//
//	content, err := client.GetFileContent(repoInfo, "registry.json")
//	if err != nil {
//	    return err
//	}
//
// # Security Considerations
//
// This package is designed to be used within a Kubernetes operator environment
// where Git repositories contain MCP server registry data. Future versions will
// include security hardening such as:
//   - Repository URL validation to prevent SSRF attacks
//   - Sandboxed Git operations
//   - Secure credential management via Kubernetes secrets
//
// # Implementation Status
//
// Current implementation supports:
//   - Public repository access via HTTPS
//   - Branch, tag, and commit checkout
//   - File content retrieval
//   - Temporary directory management
//
// Planned features:
//   - Authentication for private repositories
//   - Repository caching for performance
//   - Webhook support for immediate sync triggers
//   - Git LFS support for large files
package git
