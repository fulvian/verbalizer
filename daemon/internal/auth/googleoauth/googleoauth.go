// Package googleoauth provides Google OAuth 2.0 authentication implementation.
package googleoauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"

	"github.com/fulvian/verbalizer/daemon/internal/config"
	"github.com/fulvian/verbalizer/daemon/internal/secrets"
	"github.com/fulvian/verbalizer/daemon/internal/storage"
)

const (
	authURL  = "https://accounts.google.com/o/oauth2/v2/auth"
	tokenURL = "https://oauth2.googleapis.com/token"
)

// GoogleOAuth implements AuthProvider for Google Drive.
type GoogleOAuth struct {
	config  *config.CloudConfig
	secrets secrets.SecretStore
	state   string
}

// NewGoogleOAuth creates a new Google OAuth authenticator.
func NewGoogleOAuth(cfg *config.CloudConfig, store secrets.SecretStore) *GoogleOAuth {
	return &GoogleOAuth{
		config:  cfg,
		secrets: store,
	}
}

// StartAuth initiates the OAuth flow and returns the auth URL.
func (g *GoogleOAuth) StartAuth() (string, error) {
	// Generate PKCE verifier and challenge
	verifier, err := generatePKCEVerifier()
	if err != nil {
		return "", fmt.Errorf("failed to generate PKCE verifier: %w", err)
	}
	challenge := generatePKCEChallenge(verifier)

	// Generate state for CSRF protection
	state, err := generateState()
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	g.state = state

	// Store verifier for later use
	if err := g.secrets.Save("oauth_pkce_verifier", []byte(verifier)); err != nil {
		return "", fmt.Errorf("failed to save PKCE verifier: %w", err)
	}

	// Build authorization URL
	params := url.Values{}
	params.Set("client_id", g.config.OAuthClientID)
	params.Set("redirect_uri", fmt.Sprintf("http://%s:%s", g.config.OAuthRedirectHost, g.config.OAuthRedirectPort))
	params.Set("response_type", "code")
	params.Set("scope", g.config.Scope)
	params.Set("state", state)
	params.Set("code_challenge", challenge)
	params.Set("code_challenge_method", "S256")

	return fmt.Sprintf("%s?%s", authURL, params.Encode()), nil
}

// CompleteAuth finishes the OAuth flow with the authorization code.
func (g *GoogleOAuth) CompleteAuth(code string) error {
	// Placeholder - actual implementation in Milestone B
	return &NotImplementedError{Feature: "GoogleOAuth.CompleteAuth"}
}

// GetAccountInfo returns the authenticated user's account information.
func (g *GoogleOAuth) GetAccountInfo() (*AccountInfo, error) {
	// Placeholder
	return nil, &NotImplementedError{Feature: "GoogleOAuth.GetAccountInfo"}
}

// Revoke revokes the current authentication.
func (g *GoogleOAuth) Revoke() error {
	// Placeholder
	return &NotImplementedError{Feature: "GoogleOAuth.Revoke"}
}

// IsAuthenticated returns true if there's a valid authentication.
func (g *GoogleOAuth) IsAuthenticated() bool {
	// Placeholder
	return false
}

// GetProvider returns the cloud provider type.
func (g *GoogleOAuth) GetProvider() storage.CloudProvider {
	return storage.ProviderGoogleDrive
}

// generatePKCEVerifier generates a PKCE code verifier.
func generatePKCEVerifier() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// generatePKCEChallenge generates a PKCE code challenge from verifier using S256.
func generatePKCEChallenge(verifier string) string {
	h := sha256.New()
	h.Write([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// generateState generates a random state parameter for CSRF protection.
func generateState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// AccountInfo represents authenticated user account details.
type AccountInfo struct {
	Email string
	Name  string
}

// NotImplementedError indicates a feature that is not yet implemented.
type NotImplementedError struct {
	Feature string
}

func (e *NotImplementedError) Error() string {
	return fmt.Sprintf("%s is not yet implemented", e.Feature)
}
