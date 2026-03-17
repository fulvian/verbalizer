package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fulvian/verbalizer/daemon/internal/audio"
	"github.com/fulvian/verbalizer/daemon/internal/config"
	"github.com/fulvian/verbalizer/daemon/internal/formatter"
	"github.com/fulvian/verbalizer/daemon/internal/storage"
	"github.com/fulvian/verbalizer/daemon/internal/transcriber"
	"github.com/fulvian/verbalizer/daemon/pkg/api"
)

func TestSessionManager(t *testing.T) {
	// Setup temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "verbalizer-session-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := config.DefaultConfig()
	cfg.DataDir = tmpDir
	cfg.RecordingsDir = filepath.Join(tmpDir, "recordings")
	cfg.TranscriptsDir = filepath.Join(tmpDir, "transcripts")
	cfg.DBPath = filepath.Join(tmpDir, "test.db")

	if err := cfg.EnsureDirs(); err != nil {
		t.Fatalf("Failed to ensure dirs: %v", err)
	}

	db, err := storage.NewDatabase(cfg.DBPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	mockCapture := audio.NewMockCapture()
	mockTranscriber := transcriber.NewMockTranscriber()
	sm := &Manager{
		capture:     mockCapture,
		transcriber: mockTranscriber,
		formatter:   formatter.NewMarkdownFormatter(),
		db:          db,
		config:      cfg,
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

		// Wait a bit for background transcription
		time.Sleep(100 * time.Millisecond)

		if mockTranscriber.TranscribeCount != 1 {
			t.Errorf("Expected Transcribe to be called once, got %d", mockTranscriber.TranscribeCount)
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
