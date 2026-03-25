// Package api provides shared types and interfaces for the Verbalizer daemon.
package api

import "encoding/json"

// Platform represents the meeting platform.
type Platform string

const (
	PlatformGoogleMeet Platform = "google-meet"
	PlatformMSTeams    Platform = "ms-teams"
)

// CloudProvider represents a cloud storage provider.
type CloudProvider string

const (
	CloudProviderGoogleDrive CloudProvider = "google_drive"
)

// CommandType represents the type of command sent to the daemon.
type CommandType string

const (
	CmdStartRecording CommandType = "START_RECORDING"
	CmdStopRecording  CommandType = "STOP_RECORDING"
	CmdGetStatus      CommandType = "GET_STATUS"
	// Cloud commands
	CmdGoogleAuthStart       CommandType = "GOOGLE_AUTH_START"
	CmdGoogleAuthStatus      CommandType = "GOOGLE_AUTH_STATUS"
	CmdGoogleAuthDisconnect  CommandType = "GOOGLE_AUTH_DISCONNECT"
	CmdGoogleDriveSetFolder  CommandType = "GOOGLE_DRIVE_SET_FOLDER"
	CmdGoogleDriveGetFolder  CommandType = "GOOGLE_DRIVE_GET_FOLDER"
	CmdGoogleDriveSyncStatus CommandType = "GOOGLE_DRIVE_SYNC_STATUS"
	CmdGoogleDriveSyncRetry  CommandType = "GOOGLE_DRIVE_SYNC_RETRY"
)

// Command represents a command sent to the daemon.
type Command struct {
	Type    CommandType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// StartRecordingPayload contains data for starting a recording.
type StartRecordingPayload struct {
	Platform Platform `json:"platform"`
	CallID   string   `json:"callId"`
	Title    string   `json:"title,omitempty"`
}

// StopRecordingPayload contains data for stopping a recording.
type StopRecordingPayload struct {
	CallID string `json:"callId"`
}

// Response represents a response from the daemon.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// StatusData contains the current daemon status.
type StatusData struct {
	IsRecording    bool     `json:"isRecording"`
	IsTranscribing bool     `json:"isTranscribing"`
	CurrentCallID  string   `json:"currentCallId,omitempty"`
	Platform       Platform `json:"platform,omitempty"`
	AudioPath      string   `json:"audioPath,omitempty"`
	RecordingsDir  string   `json:"recordingsDir,omitempty"`
	TranscriptsDir string   `json:"transcriptsDir,omitempty"`
}

// RecordingStartedData contains data returned when recording starts.
type RecordingStartedData struct {
	CallID        string `json:"callId"`
	RecordingPath string `json:"recordingPath"`
}

// GoogleAuthStatusData contains Google OAuth status information.
type GoogleAuthStatusData struct {
	Connected        bool   `json:"connected"`
	Email            string `json:"email,omitempty"`
	Scopes           string `json:"scopes,omitempty"`
	ConnectedAt      string `json:"connectedAt,omitempty"`
	TargetFolderID   string `json:"targetFolderId,omitempty"`
	TargetFolderName string `json:"targetFolderName,omitempty"`
}

// GoogleDriveFolder represents a Google Drive folder.
type GoogleDriveFolder struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ParentID string `json:"parentId,omitempty"`
}

// GoogleDriveSetFolderPayload contains data for setting the target folder.
type GoogleDriveSetFolderPayload struct {
	FolderID string `json:"folderId"`
}

// GoogleDriveSyncStatusData contains sync status for a session.
type GoogleDriveSyncStatusData struct {
	CallID       string `json:"callId"`
	State        string `json:"state"` // pending, uploading, synced, failed, permanent_failed
	AttemptCount int    `json:"attemptCount"`
	LastError    string `json:"lastError,omitempty"`
	RemoteFileID string `json:"remoteFileId,omitempty"`
	UpdatedAt    string `json:"updatedAt,omitempty"`
}

// GoogleDriveSyncStats contains overall sync queue statistics.
type GoogleDriveSyncStats struct {
	Pending         int `json:"pending"`
	Uploading       int `json:"uploading"`
	Synced          int `json:"synced"`
	Failed          int `json:"failed"`
	PermanentFailed int `json:"permanentFailed"`
	Total           int `json:"total"`
}
