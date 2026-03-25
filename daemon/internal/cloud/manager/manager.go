// Package cloudmgr provides cloud service management for the daemon.
package cloudmgr

import (
	"fmt"
	"sync"

	"github.com/fulvian/verbalizer/daemon/internal/auth/googleoauth"
	"github.com/fulvian/verbalizer/daemon/internal/config"
	"github.com/fulvian/verbalizer/daemon/internal/secrets"
	"github.com/fulvian/verbalizer/daemon/internal/storage"
)

// Manager manages cloud services for the daemon.
type Manager struct {
	mu       sync.RWMutex
	config   *config.Config
	secrets  secrets.SecretStore
	database *storage.Database

	// OAuth provider
	oauth *googleoauth.GoogleOAuth
}

// NewManager creates a new cloud manager.
func NewManager(cfg *config.Config, db *storage.Database, secretsStore secrets.SecretStore) *Manager {
	return &Manager{
		config:   cfg,
		database: db,
		secrets:  secretsStore,
	}
}

// GoogleOAuth returns the Google OAuth provider, creating it if needed.
func (m *Manager) GoogleOAuth() *googleoauth.GoogleOAuth {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.oauth == nil {
		m.oauth = googleoauth.NewGoogleOAuth(&m.config.Cloud, m.secrets)
	}
	return m.oauth
}

// IsCloudEnabled returns true if cloud sync is enabled and configured.
func (m *Manager) IsCloudEnabled() bool {
	return m.config.IsCloudEnabled()
}

// GetAuthStatus returns the current authentication status.
func (m *Manager) GetAuthStatus() (*AuthStatus, error) {
	if !m.IsCloudEnabled() {
		return &AuthStatus{Enabled: false}, nil
	}

	oauth := m.GoogleOAuth()
	connected := oauth.IsAuthenticated()

	status := &AuthStatus{
		Enabled:   true,
		Connected: connected,
		Provider:  string(oauth.GetProvider()),
	}

	if connected {
		info, err := oauth.GetAccountInfo()
		if err == nil && info != nil {
			status.Email = info.GetEmail()
			status.Name = info.GetName()
		}
		status.Scopes = m.config.Cloud.Scope
	}

	return status, nil
}

// Disconnect disconnects the cloud account and revokes credentials.
func (m *Manager) Disconnect() error {
	if !m.IsCloudEnabled() {
		return fmt.Errorf("cloud sync is not enabled")
	}

	oauth := m.GoogleOAuth()
	if err := oauth.Revoke(); err != nil {
		return fmt.Errorf("failed to revoke OAuth: %w", err)
	}

	// Remove cloud account from database
	account, err := m.database.GetCloudAccount(storage.ProviderGoogleDrive)
	if err == nil && account != nil {
		m.database.RevokeCloudAccount(storage.ProviderGoogleDrive, account.AccountEmail)
	}

	return nil
}

// AuthStatus represents the authentication status.
type AuthStatus struct {
	Enabled   bool   `json:"enabled"`
	Connected bool   `json:"connected"`
	Provider  string `json:"provider,omitempty"`
	Email     string `json:"email,omitempty"`
	Name      string `json:"name,omitempty"`
	Scopes    string `json:"scopes,omitempty"`
}
