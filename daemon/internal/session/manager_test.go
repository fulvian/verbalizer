package session

import (
	"github.com/fulvian/verbalizer/daemon/internal/audio"
	"github.com/fulvian/verbalizer/daemon/pkg/api"
	"testing"
)

func TestSessionManager(t *testing.T) {
	mockCapture := audio.NewMockCapture()
	sm := &Manager{
		capture: mockCapture,
	}

	t.Run("InitialState", func(t *testing.T) {
		status := sm.GetStatus()
		if status.IsRecording {
			t.Errorf("Expected IsRecording to be false, got true")
		}
	})

	t.Run("StartRecording", func(t *testing.T) {
		payload := api.StartRecordingPayload{
			Platform: api.PlatformGoogleMeet,
			CallID:   "call-123",
			Title:    "Test Meeting",
		}
		err := sm.StartRecording(payload)
		if err != nil {
			t.Fatalf("Failed to start recording: %v", err)
		}

		status := sm.GetStatus()
		if !status.IsRecording {
			t.Errorf("Expected IsRecording to be true, got false")
		}
		if status.CurrentCallID != "call-123" {
			t.Errorf("Expected CurrentCallID to be 'call-123', got %s", status.CurrentCallID)
		}
		if status.Platform != api.PlatformGoogleMeet {
			t.Errorf("Expected Platform to be 'google-meet', got %s", status.Platform)
		}
	})

	t.Run("StartRecordingAlreadyRecording", func(t *testing.T) {
		payload := api.StartRecordingPayload{
			Platform: api.PlatformGoogleMeet,
			CallID:   "call-456",
		}
		err := sm.StartRecording(payload)
		if err == nil {
			t.Errorf("Expected error when starting recording while already recording, got nil")
		}
	})

	t.Run("StopRecording", func(t *testing.T) {
		err := sm.StopRecording("call-123")
		if err != nil {
			t.Fatalf("Failed to stop recording: %v", err)
		}

		status := sm.GetStatus()
		if status.IsRecording {
			t.Errorf("Expected IsRecording to be false, got true")
		}
	})

	t.Run("StopRecordingWrongCallID", func(t *testing.T) {
		// Restart recording first
		sm.StartRecording(api.StartRecordingPayload{CallID: "call-789"})

		err := sm.StopRecording("wrong-id")
		if err == nil {
			t.Errorf("Expected error when stopping recording with wrong CallID, got nil")
		}
	})
}
