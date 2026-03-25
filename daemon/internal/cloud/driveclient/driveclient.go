// Package driveclient provides Google Drive cloud storage implementation.
package driveclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/fulvian/verbalizer/daemon/internal/config"
	"github.com/fulvian/verbalizer/daemon/internal/secrets"
	"github.com/fulvian/verbalizer/daemon/internal/storage"
)

const (
	driveAPI    = "https://www.googleapis.com/drive/v3"
	driveUpload = "https://www.googleapis.com/upload/drive/v3"
)

// DriveClient implements CloudUploader for Google Drive.
type DriveClient struct {
	config  *config.CloudConfig
	secrets secrets.SecretStore
}

// UploadResult represents the result of a file upload operation.
type UploadResult struct {
	FileID      string `json:"id"`
	FileName    string `json:"name"`
	FileSize    int64  `json:"size"`
	MimeType    string `json:"mimeType"`
	WebViewLink string `json:"webViewLink"`
}

// Folder represents a folder in cloud storage.
type Folder struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ParentID string `json:"parents"`
}

// AboutResponse represents Drive about info.
type AboutResponse struct {
	User struct {
		EmailAddress string `json:"emailAddress"`
		DisplayName  string `json:"displayName"`
	} `json:"user"`
}

// NewDriveClient creates a new Google Drive client.
func NewDriveClient(cfg *config.CloudConfig, store secrets.SecretStore) *DriveClient {
	return &DriveClient{
		config:  cfg,
		secrets: store,
	}
}

// getAccessToken retrieves a valid access token.
func (c *DriveClient) getAccessToken() (string, error) {
	token, err := c.secrets.Get("google_access_token")
	if err != nil {
		return "", fmt.Errorf("access token not found: %w", err)
	}
	return string(token), nil
}

// Upload uploads a file to Google Drive using multipart upload.
func (c *DriveClient) Upload(path string, folderID string) (*UploadResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	token, err := c.getAccessToken()
	if err != nil {
		return nil, err
	}

	// Build multipart request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add metadata part
	metadata := map[string]interface{}{
		"name":    filepath.Base(path),
		"parents": []string{folderID},
	}
	metadataJSON, _ := json.Marshal(metadata)

	part, err := writer.CreateFormFile("metadata", "metadata.json")
	if err != nil {
		return nil, fmt.Errorf("failed to create metadata part: %w", err)
	}
	part.Write(metadataJSON)

	// Add media part
	part, err = writer.CreateFormFile("media", filepath.Base(path))
	if err != nil {
		return nil, fmt.Errorf("failed to create media part: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}
	writer.Close()

	// Create request
	req, err := http.NewRequest("POST", driveUpload+"/files?uploadType=multipart", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result UploadResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// CreateFolder creates a folder in Google Drive.
func (c *DriveClient) CreateFolder(name string, parentID string) (string, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return "", err
	}

	folderMetadata := map[string]interface{}{
		"name":     name,
		"parents":  []string{parentID},
		"mimeType": "application/vnd.google-apps.folder",
	}

	jsonData, err := json.Marshal(folderMetadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal folder metadata: %w", err)
	}

	req, err := http.NewRequest("POST", driveAPI+"/files", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("create folder request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create folder failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.ID, nil
}

// ListFolders lists folders in Google Drive.
func (c *DriveClient) ListFolders(parentID string) ([]Folder, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return nil, err
	}

	// Build query
	query := "mimeType='application/vnd.google-apps.folder' and trashed=false"
	if parentID != "" {
		query += fmt.Sprintf(" and '%s' in parents", parentID)
	}

	apiURL := driveAPI + "/files?" + url.Values{
		"q":        {query},
		"fields":   {"files(id,name,parents)"},
		"pageSize": {"100"},
	}.Encode()

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list folders request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list folders failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Files []Folder `json:"files"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Files, nil
}

// GetAbout returns account information.
func (c *DriveClient) GetAbout() (*AboutResponse, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", driveAPI+"/about?fields=user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("about request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("about failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var about AboutResponse
	if err := json.NewDecoder(resp.Body).Decode(&about); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &about, nil
}

// GetProvider returns the cloud provider type.
func (c *DriveClient) GetProvider() storage.CloudProvider {
	return storage.ProviderGoogleDrive
}

// NotImplementedError indicates a feature that is not yet implemented.
type NotImplementedError struct {
	Feature string
}

func (e *NotImplementedError) Error() string {
	return fmt.Sprintf("%s is not yet implemented", e.Feature)
}
