package git

import (
	"github.com/go-git/go-git/v5"
)

// CloneConfig contains configuration for cloning a repository
type CloneConfig struct {
	// URL is the repository URL to clone
	URL string

	// Branch is the specific branch to clone (optional)
	Branch string

	// Tag is the specific tag to clone (optional)
	Tag string

	// Commit is the specific commit to clone (optional)
	Commit string

	// Directory is the local directory to clone into
	Directory string

	// Auth contains authentication configuration
	Auth AuthConfig
}

// AuthConfig contains authentication configuration for Git operations
type AuthConfig struct {
	// Type is the authentication type (none, token, ssh-key, basic-auth)
	Type AuthType

	// Token is used for token-based authentication (GitHub, GitLab tokens)
	Token string

	// Username is used for basic authentication
	Username string

	// Password is used for basic authentication
	Password string

	// SSHKey contains SSH private key data for SSH authentication
	SSHKey []byte

	// SSHKeyPassword is the password for encrypted SSH keys
	SSHKeyPassword string
}

// AuthType represents the type of authentication to use
type AuthType string

const (
	// AuthTypeNone means no authentication (public repositories)
	AuthTypeNone AuthType = "none"

	// AuthTypeToken means token-based authentication (Personal Access Tokens)
	AuthTypeToken AuthType = "token"

	// AuthTypeSSHKey means SSH key authentication
	AuthTypeSSHKey AuthType = "ssh-key"

	// AuthTypeBasic means basic username/password authentication
	AuthTypeBasic AuthType = "basic-auth"
)

// RepositoryInfo contains information about a Git repository
type RepositoryInfo struct {
	// Repository is the go-git repository instance
	Repository *git.Repository

	// CurrentCommit is the current commit hash
	CurrentCommit string

	// Branch is the current branch name
	Branch string

	// RemoteURL is the remote repository URL
	RemoteURL string
}
