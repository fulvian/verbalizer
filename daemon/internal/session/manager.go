package session

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/fulvian/verbalizer/daemon/internal/audio"
	"github.com/fulvian/verbalizer/daemon/internal/transcriber"
	"github.com/fulvian/verbalizer/daemon/pkg/api"
)

// Session represents an active recording session.
type Session struct {
	CallID    string
	Platform  api.Platform
	Title     string
	StartTime time.Time
	AudioPath string
}

// Manager manages active recording sessions.
type Manager struct {
	mu             sync.RWMutex
	currentSession *Session
	capture        audio.Capture
	transcriber    transcriber.Transcriber
	isTranscribing bool
}

// NewManager creates a new session manager.
func NewManager() (*Manager, error) {
	var capturer audio.Capture
	var err error

	switch runtime.GOOS {
	case "linux":
		capturer, err = audio.NewLinuxCapture()
	case "darwin":
		capturer, err = audio.NewDarwinCapture()
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize audio capture: %w", err)
	}

	// Initialize transcriber with default paths
	// In a real scenario, these would be configurable.
	binaryPath := "whisper/whisper.cpp/main"
	modelPath := "whisper/models/ggml-small.bin"

	return &Manager{
		capture:     capturer,
		transcriber: transcriber.NewWhisper(binaryPath, modelPath),
	}, nil
}

// StartRecording starts a new recording session.
func (m *Manager) StartRecording(payload api.StartRecordingPayload) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentSession != nil {
		return errors.New("recording already in progress")
	}

	if err := m.capture.Start(); err != nil {
		return fmt.Errorf("failed to start audio capture: %w", err)
	}

	m.currentSession = &Session{
		CallID:    payload.CallID,
		Platform:  payload.Platform,
		Title:     payload.Title,
		StartTime: time.Now(),
	}

	return nil
}

// StopRecording stops the current recording session.
func (m *Manager) StopRecording(callID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentSession == nil {
		return errors.New("no recording in progress")
	}

	if m.currentSession.CallID != callID {
		return errors.New("call ID mismatch")
	}

	audioPath, err := m.capture.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop audio capture: %w", err)
	}

	m.currentSession.AudioPath = audioPath

	// Trigger transcription in background
	go func(sess *Session) {
		m.mu.Lock()
		m.isTranscribing = true
		m.mu.Unlock()
		defer func() {
			m.mu.Lock()
			m.isTranscribing = false
			m.mu.Unlock()
		}()

		fmt.Printf("Starting transcription for session %s (%s)...\n", sess.CallID, sess.AudioPath)
		transcript, err := m.transcriber.Transcribe(sess.AudioPath)
		if err != nil {
			fmt.Printf("Transcription failed for session %s: %v\n", sess.CallID, err)
			return
		}

		// Save transcription to a file next to the audio
		transcriptPath := sess.AudioPath + ".txt"
		if err := os.WriteFile(transcriptPath, []byte(transcript.Text), 0644); err != nil {
			fmt.Printf("Failed to save transcript for session %s: %v\n", sess.CallID, err)
			return
		}

		fmt.Printf("Transcription complete for session %s. Result saved to %s\n", sess.CallID, transcriptPath)
	}(m.currentSession)

	m.currentSession = nil
	return nil
}

// GetStatus returns the current recording status.
func (m *Manager) GetStatus() api.StatusData {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := api.StatusData{
		IsRecording:    m.currentSession != nil,
		IsTranscribing: m.isTranscribing,
	}

	if m.currentSession != nil {
		status.CurrentCallID = m.currentSession.CallID
		status.Platform = m.currentSession.Platform
		status.AudioPath = m.currentSession.AudioPath
	}

	return status
}
