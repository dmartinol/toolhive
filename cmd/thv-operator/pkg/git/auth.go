package git

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	ssh2 "golang.org/x/crypto/ssh"
)

// AuthManager handles authentication for Git operations
type AuthManager interface {
	// PrepareAuth prepares authentication method from config
	PrepareAuth(config AuthConfig) (transport.AuthMethod, error)
}

// DefaultAuthManager implements AuthManager
type DefaultAuthManager struct{}

// NewDefaultAuthManager creates a new DefaultAuthManager
func NewDefaultAuthManager() *DefaultAuthManager {
	return &DefaultAuthManager{}
}

// PrepareAuth prepares authentication method from config
func (*DefaultAuthManager) PrepareAuth(config AuthConfig) (transport.AuthMethod, error) {
	switch config.Type {
	case AuthTypeNone:
		return nil, nil

	case AuthTypeToken:
		if config.Token == "" {
			return nil, fmt.Errorf("token is required for token authentication")
		}
		return &http.BasicAuth{
			Username: "token", // GitHub/GitLab convention
			Password: config.Token,
		}, nil

	case AuthTypeBasic:
		if config.Username == "" || config.Password == "" {
			return nil, fmt.Errorf("username and password are required for basic authentication")
		}
		return &http.BasicAuth{
			Username: config.Username,
			Password: config.Password,
		}, nil

	case AuthTypeSSHKey:
		if len(config.SSHKey) == 0 {
			return nil, fmt.Errorf("SSH key is required for SSH authentication")
		}

		var signer ssh2.Signer
		var err error

		if config.SSHKeyPassword != "" {
			// Parse encrypted private key
			signer, err = ssh2.ParsePrivateKeyWithPassphrase(config.SSHKey, []byte(config.SSHKeyPassword))
		} else {
			// Parse unencrypted private key
			signer, err = ssh2.ParsePrivateKey(config.SSHKey)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to parse SSH key: %w", err)
		}

		return &ssh.PublicKeys{
			User:   "git",
			Signer: signer,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported authentication type: %s", config.Type)
	}
}
