package session

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/fulvian/verbalizer/daemon/internal/audio"
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

	return &Manager{
		capture: capturer,
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
	m.currentSession = nil
	return nil
}

// GetStatus returns the current recording status.
func (m *Manager) GetStatus() api.StatusData {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := api.StatusData{
		IsRecording: m.currentSession != nil,
	}

	if m.currentSession != nil {
		status.CurrentCallID = m.currentSession.CallID
		status.Platform = m.currentSession.Platform
		status.AudioPath = m.currentSession.AudioPath
	}

	return status
}
