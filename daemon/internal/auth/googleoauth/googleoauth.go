// Package googleoauth provides Google OAuth 2.0 authentication implementation.
package googleoauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/fulvian/verbalizer/daemon/internal/config"
	"github.com/fulvian/verbalizer/daemon/internal/secrets"
	"github.com/fulvian/verbalizer/daemon/internal/storage"
)

// AccountInfo represents authenticated user account details.
type AccountInfo struct {
	Email string
	Name  string
}

// GetEmail returns the account email.
func (a *AccountInfo) GetEmail() string {
	return a.Email
}

// GetName returns the account name.
func (a *AccountInfo) GetName() string {
	return a.Name
}

const (
	authURL     = "https://accounts.google.com/o/oauth2/v2/auth"
	tokenURL    = "https://oauth2.googleapis.com/token"
	revokeURL   = "https://oauth2.googleapis.com/revoke"
	userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
)

// Keys for secret storage
const (
	secretKeyAccessToken  = "google_access_token"
	secretKeyRefreshToken = "google_refresh_token"
	secretKeyExpiry       = "google_token_expiry"
	secretKeyPKCEVerifier = "oauth_pkce_verifier"
	secretKeyOAuthState   = "oauth_state"
)

// GoogleOAuth implements AuthProvider for Google Drive.
type GoogleOAuth struct {
	config  *config.CloudConfig
	secrets secrets.SecretStore
	state   string
	mu      sync.RWMutex

	// HTTP server for loopback callback
	server     *http.Server
	serverAddr string
}

// TokenResponse represents Google's token endpoint response.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	IDToken      string `json:"id_token"`
}

// UserInfo represents user information from Google's userinfo endpoint.
type UserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
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
	if err := g.secrets.Save(secretKeyPKCEVerifier, []byte(verifier)); err != nil {
		return "", fmt.Errorf("failed to save PKCE verifier: %w", err)
	}

	// Store state for validation
	if err := g.secrets.Save(secretKeyOAuthState, []byte(state)); err != nil {
		return "", fmt.Errorf("failed to save OAuth state: %w", err)
	}

	// Build authorization URL
	params := url.Values{}
	params.Set("client_id", g.config.OAuthClientID)
	params.Set("redirect_uri", g.getRedirectURI())
	params.Set("response_type", "code")
	params.Set("scope", g.config.Scope)
	params.Set("state", state)
	params.Set("code_challenge", challenge)
	params.Set("code_challenge_method", "S256")

	return fmt.Sprintf("%s?%s", authURL, params.Encode()), nil
}

// CompleteAuth finishes the OAuth flow with the authorization code.
// It starts a loopback server, exchanges the code for tokens, and returns.
func (g *GoogleOAuth) CompleteAuth(authCode string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Validate state first
	storedState, err := g.secrets.Get(secretKeyOAuthState)
	if err != nil {
		return fmt.Errorf("OAuth state not found, call StartAuth first: %w", err)
	}
	if string(storedState) != g.state {
		return errors.New("OAuth state mismatch - possible CSRF attack")
	}

	// Get PKCE verifier
	verifier, err := g.secrets.Get(secretKeyPKCEVerifier)
	if err != nil {
		return fmt.Errorf("PKCE verifier not found: %w", err)
	}

	// Exchange code for tokens
	tokenResp, err := g.exchangeCodeForToken(authCode, string(verifier))
	if err != nil {
		return fmt.Errorf("token exchange failed: %w", err)
	}

	// Store tokens securely
	if err := g.secrets.Save(secretKeyAccessToken, []byte(tokenResp.AccessToken)); err != nil {
		return fmt.Errorf("failed to store access token: %w", err)
	}

	if tokenResp.RefreshToken != "" {
		if err := g.secrets.Save(secretKeyRefreshToken, []byte(tokenResp.RefreshToken)); err != nil {
			return fmt.Errorf("failed to store refresh token: %w", err)
		}
	}

	// Calculate and store expiry time
	expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	expiryBytes, _ := json.Marshal(expiry)
	if err := g.secrets.Save(secretKeyExpiry, expiryBytes); err != nil {
		return fmt.Errorf("failed to store token expiry: %w", err)
	}

	// Clean up temporary secrets
	g.secrets.Delete(secretKeyPKCEVerifier)
	g.secrets.Delete(secretKeyOAuthState)

	return nil
}

// exchangeCodeForToken exchanges an authorization code for tokens.
func (g *GoogleOAuth) exchangeCodeForToken(code, verifier string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", g.config.OAuthClientID)
	data.Set("code", code)
	data.Set("code_verifier", verifier)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", g.getRedirectURI())

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token endpoint returned status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

// refreshAccessToken refreshes the access token using the refresh token.
func (g *GoogleOAuth) refreshAccessToken() error {
	refreshToken, err := g.secrets.Get(secretKeyRefreshToken)
	if err != nil {
		return fmt.Errorf("refresh token not found: %w", err)
	}

	data := url.Values{}
	data.Set("client_id", g.config.OAuthClientID)
	data.Set("refresh_token", string(refreshToken))
	data.Set("grant_type", "refresh_token")

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token refresh returned status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	// Store new access token
	if err := g.secrets.Save(secretKeyAccessToken, []byte(tokenResp.AccessToken)); err != nil {
		return fmt.Errorf("failed to store access token: %w", err)
	}

	// Update expiry
	if tokenResp.ExpiresIn > 0 {
		expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
		expiryBytes, _ := json.Marshal(expiry)
		if err := g.secrets.Save(secretKeyExpiry, expiryBytes); err != nil {
			return fmt.Errorf("failed to store token expiry: %w", err)
		}
	}

	return nil
}

// GetAccessToken returns a valid access token, refreshing if necessary.
func (g *GoogleOAuth) GetAccessToken() (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Check if token is expired
	expiryBytes, err := g.secrets.Get(secretKeyExpiry)
	if err == nil {
		var expiry time.Time
		if json.Unmarshal(expiryBytes, &expiry) == nil && time.Now().After(expiry) {
			// Token expired, refresh
			if err := g.refreshAccessToken(); err != nil {
				return "", fmt.Errorf("failed to refresh access token: %w", err)
			}
		}
	}

	// Get current access token
	tokenBytes, err := g.secrets.Get(secretKeyAccessToken)
	if err != nil {
		return "", errors.New("not authenticated, call StartAuth first")
	}

	return string(tokenBytes), nil
}

// GetAccountInfo returns the authenticated user's account information.
func (g *GoogleOAuth) GetAccountInfo() (*AccountInfo, error) {
	token, err := g.GetAccessToken()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo endpoint returned status %d", resp.StatusCode)
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse userinfo response: %w", err)
	}

	return &AccountInfo{
		Email: userInfo.Email,
		Name:  userInfo.Name,
	}, nil
}

// Revoke revokes the current authentication.
func (g *GoogleOAuth) Revoke() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Try to revoke refresh token if we have one
	refreshToken, err := g.secrets.Get(secretKeyRefreshToken)
	if err == nil && len(refreshToken) > 0 {
		// Revoke on Google's servers (best effort)
		data := url.Values{}
		data.Set("token", string(refreshToken))
		http.PostForm(revokeURL, data)
	}

	// Clean up all stored secrets
	g.secrets.Delete(secretKeyAccessToken)
	g.secrets.Delete(secretKeyRefreshToken)
	g.secrets.Delete(secretKeyExpiry)
	g.secrets.Delete(secretKeyPKCEVerifier)
	g.secrets.Delete(secretKeyOAuthState)

	return nil
}

// IsAuthenticated returns true if there's a valid authentication.
func (g *GoogleOAuth) IsAuthenticated() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Check if we have a refresh token (primary indicator)
	exists, err := g.secrets.Exists(secretKeyRefreshToken)
	if err == nil && exists {
		return true
	}

	// Fallback: check access token
	exists, err = g.secrets.Exists(secretKeyAccessToken)
	return err == nil && exists
}

// GetProvider returns the cloud provider type.
func (g *GoogleOAuth) GetProvider() storage.CloudProvider {
	return storage.ProviderGoogleDrive
}

// StartCallbackServer starts an HTTP server on a random port from the configured range.
func (g *GoogleOAuth) StartCallbackServer() (int, error) {
	// Parse port range
	portRange := g.config.OAuthRedirectPort
	if portRange == "" {
		portRange = "49152-65535"
	}

	// Try to find an available port
	portStart, portEnd := 49152, 65535
	if _, err := fmt.Sscanf(portRange, "%d-%d", &portStart, &portEnd); err != nil {
		portStart, portEnd = 49152, 65535
	}

	// Find available port
	var port int
	for p := portStart; p <= portEnd; p++ {
		addr := fmt.Sprintf("%s:%d", g.config.OAuthRedirectHost, p)
		if listener, err := net.Listen("tcp", addr); err == nil {
			listener.Close()
			port = p
			break
		}
	}
	if port == 0 {
		return 0, errors.New("no available port in range")
	}

	g.serverAddr = fmt.Sprintf("%s:%d", g.config.OAuthRedirectHost, port)
	return port, nil
}

// WaitForCallback waits for OAuth callback and returns the auth code.
func (g *GoogleOAuth) WaitForCallback(timeout time.Duration) (string, error) {
	if g.serverAddr == "" {
		return "", errors.New("callback server not started")
	}

	// Start server
	server := &http.Server{
		Addr:         g.serverAddr,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}
	g.server = server

	// Channel to receive the code
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		params := r.URL.Query()
		errorParam := params.Get("error")
		if errorParam != "" {
			errCh <- fmt.Errorf("OAuth error: %s - %s", errorParam, params.Get("error_description"))
			return
		}

		code := params.Get("code")
		state := params.Get("state")

		// Validate state
		if state != g.state {
			errCh <- errors.New("state mismatch in callback")
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}

		codeCh <- code
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body><h1>Authentication successful!</h1><p>You can close this window.</p></body></html>"))
	})

	go server.ListenAndServe()

	select {
	case code := <-codeCh:
		return code, nil
	case err := <-errCh:
		return "", err
	case <-time.After(timeout):
		return "", errors.New("OAuth callback timeout")
	}
}

// StopCallbackServer stops the callback server.
func (g *GoogleOAuth) StopCallbackServer() {
	if g.server != nil {
		g.server.Close()
		g.server = nil
	}
}

func (g *GoogleOAuth) getRedirectURI() string {
	return fmt.Sprintf("http://%s", g.serverAddr)
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

// NormalizeEmail normalizes Google email addresses.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
