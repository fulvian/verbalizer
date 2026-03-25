// Package driveclient provides Google Drive cloud storage implementation.
package driveclient

import (
	"fmt"

	"github.com/fulvian/verbalizer/daemon/internal/config"
	"github.com/fulvian/verbalizer/daemon/internal/secrets"
	"github.com/fulvian/verbalizer/daemon/internal/storage"
)

// DriveClient implements CloudUploader for Google Drive.
type DriveClient struct {
	config  *config.CloudConfig
	secrets secrets.SecretStore
}

// UploadResult represents the result of a file upload operation.
type UploadResult struct {
	FileID      string
	FileName    string
	FileSize    int64
	MimeType    string
	WebViewLink string
}

// Folder represents a folder in cloud storage.
type Folder struct {
	ID       string
	Name     string
	ParentID string
}

// NewDriveClient creates a new Google Drive client.
func NewDriveClient(cfg *config.CloudConfig, store secrets.SecretStore) *DriveClient {
	return &DriveClient{
		config:  cfg,
		secrets: store,
	}
}

// Upload uploads a file to Google Drive.
func (c *DriveClient) Upload(path string, folderID string) (*UploadResult, error) {
	return nil, &NotImplementedError{Feature: "DriveClient.Upload"}
}

// CreateFolder creates a folder in Google Drive.
func (c *DriveClient) CreateFolder(name string, parentID string) (string, error) {
	return "", &NotImplementedError{Feature: "DriveClient.CreateFolder"}
}

// ListFolders lists folders in Google Drive.
func (c *DriveClient) ListFolders(parentID string) ([]Folder, error) {
	return nil, &NotImplementedError{Feature: "DriveClient.ListFolders"}
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
