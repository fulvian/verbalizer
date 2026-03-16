package main

import (
	"encoding/json"
	"fmt"

	"github.com/fulvian/verbalizer/daemon/internal/session"
	"github.com/fulvian/verbalizer/daemon/pkg/api"
)

// DaemonHandler connects IPC commands to the SessionManager.
type DaemonHandler struct {
	sessionMgr *session.Manager
}

// NewDaemonHandler creates a new DaemonHandler.
func NewDaemonHandler(sm *session.Manager) *DaemonHandler {
	return &DaemonHandler{
		sessionMgr: sm,
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
		return &api.Response{
			Success: true,
			Data: api.RecordingStartedData{
				CallID:        payload.CallID,
				RecordingPath: fmt.Sprintf("/tmp/%s.mp3", payload.CallID), // Placeholder
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

	default:
		return &api.Response{Success: false, Error: "unknown command"}, nil
	}
}
