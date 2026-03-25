// Package cloud provides interfaces and implementations for cloud storage operations.
package cloud

import (
	"fmt"

	"github.com/fulvian/verbalizer/daemon/internal/config"
	"github.com/fulvian/verbalizer/daemon/internal/storage"
)

// AuthProvider defines the interface for cloud authentication providers.
type AuthProvider interface {
	// StartAuth initiates the OAuth flow and returns the auth URL to open in browser.
	StartAuth() (string, error)

	// CompleteAuth finishes the OAuth flow with the authorization code.
	CompleteAuth(code string) error

	// GetAccountInfo returns the authenticated user's account information.
	GetAccountInfo() (*AccountInfo, error)

	// Revoke revokes the current authentication.
	Revoke() error

	// IsAuthenticated returns true if there's a valid authentication.
	IsAuthenticated() bool

	// GetProvider returns the cloud provider type.
	GetProvider() storage.CloudProvider
}

// AccountInfo represents authenticated user account details.
type AccountInfo struct {
	Email string
	Name  string
}

// CloudUploader defines the interface for uploading files to cloud storage.
type CloudUploader interface {
	// Upload uploads a file to the cloud storage.
	Upload(path string, folderID string) (*UploadResult, error)

	// CreateFolder creates a new folder in the cloud storage.
	CreateFolder(name string, parentID string) (string, error)

	// ListFolders lists folders in the cloud storage.
	ListFolders(parentID string) ([]Folder, error)

	// GetProvider returns the cloud provider type.
	GetProvider() storage.CloudProvider
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

// SyncQueue defines the interface for managing sync job queues.
type SyncQueue interface {
	// Enqueue adds a new sync job to the queue.
	Enqueue(callID string, localPath string, folderID string) error

	// ProcessNext processes the next job in the queue.
	ProcessNext() error

	// GetJobStatus returns the status of a sync job.
	GetJobStatus(callID string) (*JobStatus, error)

	// RetryFailed retries all failed jobs that are ready for retry.
	RetryFailed() error

	// GetStats returns queue statistics.
	GetStats() (*QueueStats, error)
}

// JobStatus represents the status of a sync job.
type JobStatus struct {
	CallID       string
	State        string
	AttemptCount int
	LastError    string
	RemoteFileID string
	CreatedAt    string
	UpdatedAt    string
}

// QueueStats represents queue statistics.
type QueueStats struct {
	Pending         int
	Uploading       int
	Synced          int
	Failed          int
	PermanentFailed int
	Total           int
}

// NewCloudUploader creates a new cloud uploader based on configuration.
func NewCloudUploader(provider storage.CloudProvider, cfg *config.CloudConfig, store interface{}) (CloudUploader, error) {
	// Placeholder - actual implementation in Milestone C
	return nil, &NotImplementedError{Feature: "NewCloudUploader"}
}

// NewAuthProvider creates a new auth provider based on configuration.
func NewAuthProvider(provider storage.CloudProvider, cfg *config.CloudConfig, store interface{}) (AuthProvider, error) {
	// Placeholder - actual implementation in Milestone B
	return nil, &NotImplementedError{Feature: "NewAuthProvider"}
}

// UnsupportedProviderError represents an unsupported cloud provider error.
type UnsupportedProviderError struct {
	Provider storage.CloudProvider
}

func (e *UnsupportedProviderError) Error() string {
	return "unsupported cloud provider: " + string(e.Provider)
}

// NotImplementedError indicates a feature that is not yet implemented.
type NotImplementedError struct {
	Feature string
}

func (e *NotImplementedError) Error() string {
	return fmt.Sprintf("%s is not yet implemented", e.Feature)
}

// SecretStore defines the interface for secure credential storage.
type SecretStore interface {
	// Save saves a secret value.
	Save(key string, value []byte) error

	// Get retrieves a secret value.
	Get(key string) ([]byte, error)

	// Delete removes a secret.
	Delete(key string) error

	// Exists checks if a secret exists.
	Exists(key string) (bool, error)
}
