package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	cloudmgr "github.com/fulvian/verbalizer/daemon/internal/cloud/manager"
	"github.com/fulvian/verbalizer/daemon/internal/session"
	"github.com/fulvian/verbalizer/daemon/pkg/api"
)

// DaemonHandler connects IPC commands to the SessionManager and CloudManager.
type DaemonHandler struct {
	sessionMgr *session.Manager
	cloudMgr   *cloudmgr.Manager
}

// NewDaemonHandler creates a new DaemonHandler.
func NewDaemonHandler(sm *session.Manager, cm *cloudmgr.Manager) *DaemonHandler {
	return &DaemonHandler{
		sessionMgr: sm,
		cloudMgr:   cm,
	}
}

// HandleCommand processes commands from the IPC server.
func (h *DaemonHandler) HandleCommand(cmd *api.Command) (*api.Response, error) {
	switch cmd.Type {
	case api.CmdStartRecording:
		var payload api.StartRecordingPayload
		if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
			return nil, fmt.Errorf("invalid payload: %w", err)
		}
		if err := h.sessionMgr.StartRecording(payload); err != nil {
			return &api.Response{Success: false, Error: err.Error()}, nil
		}
		// Get the actual recording path from session manager
		status := h.sessionMgr.GetStatus()
		return &api.Response{
			Success: true,
			Data: api.RecordingStartedData{
				CallID:        payload.CallID,
				RecordingPath: status.AudioPath, // Real path, not placeholder
			},
		}, nil

	case api.CmdStopRecording:
		var payload api.StopRecordingPayload
		if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
			return nil, fmt.Errorf("invalid payload: %w", err)
		}
		if err := h.sessionMgr.StopRecording(payload.CallID); err != nil {
			return &api.Response{Success: false, Error: err.Error()}, nil
		}
		return &api.Response{Success: true}, nil

	case api.CmdGetStatus:
		status := h.sessionMgr.GetStatus()
		return &api.Response{Success: true, Data: status}, nil

	case api.CmdGoogleAuthStart:
		return h.handleGoogleAuthStart()

	case api.CmdGoogleAuthStatus:
		return h.handleGoogleAuthStatus()

	case api.CmdGoogleAuthDisconnect:
		return h.handleGoogleAuthDisconnect()

	default:
		return &api.Response{Success: false, Error: "unknown command"}, nil
	}
}

// handleGoogleAuthStart initiates the Google OAuth flow.
func (h *DaemonHandler) handleGoogleAuthStart() (*api.Response, error) {
	if !h.cloudMgr.IsCloudEnabled() {
		return &api.Response{Success: false, Error: "cloud sync is not enabled"}, nil
	}

	oauth := h.cloudMgr.GoogleOAuth()

	// Start callback server
	_, err := oauth.StartCallbackServer()
	if err != nil {
		return &api.Response{Success: false, Error: fmt.Sprintf("failed to start callback server: %v", err)}, nil
	}

	// Get auth URL
	authURL, err := oauth.StartAuth()
	if err != nil {
		oauth.StopCallbackServer()
		return &api.Response{Success: false, Error: fmt.Sprintf("failed to start auth: %v", err)}, nil
	}

	// Open browser
	if err := openBrowser(authURL); err != nil {
		oauth.StopCallbackServer()
		return &api.Response{Success: false, Error: fmt.Sprintf("failed to open browser: %v", err)}, nil
	}

	// Wait for callback (with timeout)
	code, err := oauth.WaitForCallback(5 * time.Minute)
	oauth.StopCallbackServer()

	if err != nil {
		return &api.Response{Success: false, Error: fmt.Sprintf("OAuth callback failed: %v", err)}, nil
	}

	// Complete auth
	if err := oauth.CompleteAuth(code); err != nil {
		return &api.Response{Success: false, Error: fmt.Sprintf("failed to complete auth: %v", err)}, nil
	}

	return &api.Response{Success: true}, nil
}

// handleGoogleAuthStatus returns the current Google OAuth status.
func (h *DaemonHandler) handleGoogleAuthStatus() (*api.Response, error) {
	status, err := h.cloudMgr.GetAuthStatus()
	if err != nil {
		return &api.Response{Success: false, Error: err.Error()}, nil
	}
	return &api.Response{Success: true, Data: status}, nil
}

// handleGoogleAuthDisconnect revokes Google OAuth credentials.
func (h *DaemonHandler) handleGoogleAuthDisconnect() (*api.Response, error) {
	if err := h.cloudMgr.Disconnect(); err != nil {
		return &api.Response{Success: false, Error: err.Error()}, nil
	}
	return &api.Response{Success: true}, nil
}

// openBrowser opens a URL in the default browser.
func openBrowser(urlStr string) error {
	var cmd string
	switch runtime.GOOS {
	case "linux":
		cmd = "xdg-open"
	case "darwin":
		cmd = "open"
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return exec.Command(cmd, urlStr).Start()
}
