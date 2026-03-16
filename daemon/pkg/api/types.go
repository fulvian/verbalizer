// Package api provides shared types and interfaces for the Verbalizer daemon.
package api

import "encoding/json"

// Platform represents the meeting platform.
type Platform string

const (
	PlatformGoogleMeet Platform = "google-meet"
	PlatformMSTeams    Platform = "ms-teams"
)

// CommandType represents the type of command sent to the daemon.
type CommandType string

const (
	CmdStartRecording CommandType = "START_RECORDING"
	CmdStopRecording  CommandType = "STOP_RECORDING"
	CmdGetStatus      CommandType = "GET_STATUS"
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
	CurrentCallID  string   `json:"currentCallId,omitempty"`
	Platform       Platform `json:"platform,omitempty"`
	RecordingsDir  string   `json:"recordingsDir"`
	TranscriptsDir string   `json:"transcriptsDir"`
}

// RecordingStartedData contains data returned when recording starts.
type RecordingStartedData struct {
	CallID        string `json:"callId"`
	RecordingPath string `json:"recordingPath"`
}
