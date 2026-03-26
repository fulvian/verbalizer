package session

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/fulvian/verbalizer/daemon/internal/audio"
	"github.com/fulvian/verbalizer/daemon/internal/config"
	"github.com/fulvian/verbalizer/daemon/internal/formatter"
	"github.com/fulvian/verbalizer/daemon/internal/storage"
	"github.com/fulvian/verbalizer/daemon/internal/transcriber"
	"github.com/fulvian/verbalizer/daemon/pkg/api"
)

// Session represents an active recording session.
type Session struct {
	CallID         string
	Platform       api.Platform
	Title          string
	StartTime      time.Time
	AudioPath      string
	TranscriptPath string
}

// Manager manages active recording sessions.
type Manager struct {
	mu             sync.RWMutex
	currentSession *Session
	capture        audio.Capture
	transcriber    transcriber.Transcriber
	formatter      formatter.Formatter
	db             *storage.Database
	config         *config.Config
	isTranscribing bool
	syncQueue      SyncQueue
}

// SyncQueue defines the interface for sync job queuing.
type SyncQueue interface {
	Enqueue(callID string, localPath string, folderID string) error
}

// SetSyncQueue sets the sync queue for cloud uploads.
func (m *Manager) SetSyncQueue(q SyncQueue) {
	m.syncQueue = q
}

// NewManager creates a new session manager.
func NewManager(cfg *config.Config) (*Manager, error) {
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

	db, err := storage.NewDatabase(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Use config-driven paths for whisper binary and model
	// These should be absolute paths in production
	binaryPath := cfg.Transcription.Model
	modelPath := filepath.Join(cfg.DataDir, "models", "ggml-"+cfg.Transcription.Model+".bin")

	// Override with absolute path if not already absolute
	if !filepath.IsAbs(binaryPath) {
		binaryPath = filepath.Join(cfg.DataDir, binaryPath)
	}

	return &Manager{
		capture:     capturer,
		transcriber: transcriber.NewWhisper(binaryPath, modelPath),
		formatter:   formatter.NewMarkdownFormatter(),
		db:          db,
		config:      cfg,
	}, nil
}

// GetDatabase returns the database instance for use by other managers.
func (m *Manager) GetDatabase() *storage.Database {
	return m.db
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

	// Save session to database
	dbSess := &storage.Session{
		CallID:    m.currentSession.CallID,
		Platform:  string(m.currentSession.Platform),
		Title:     m.currentSession.Title,
		StartTime: m.currentSession.StartTime,
	}
	if err := m.db.SaveSession(dbSess); err != nil {
		fmt.Printf("Failed to save session to database: %v\n", err)
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

		// Generate Markdown output
		duration := time.Since(sess.StartTime)
		data := formatter.TranscriptData{
			Metadata: formatter.Metadata{
				Title:     sess.Title,
				Date:      sess.StartTime,
				Platform:  string(sess.Platform),
				Duration:  duration,
				AudioFile: filepath.Base(sess.AudioPath),
			},
			FullText: transcript.Text,
		}

		for _, s := range transcript.Segments {
			data.Segments = append(data.Segments, formatter.TranscriptSegment{
				Start: time.Duration(s.Start * float64(time.Second)),
				End:   time.Duration(s.End * float64(time.Second)),
				Text:  s.Text,
			})
		}

		markdown, err := m.formatter.Format(data)
		if err != nil {
			fmt.Printf("Failed to format transcript for session %s: %v\n", sess.CallID, err)
			return
		}

		// Save Markdown file to transcripts/ directory
		// Filename format: YYYY-MM-DD_HH-MM-SS_platform.md
		timestamp := sess.StartTime.Format("2006-01-02_15-04-05")
		mdFilename := fmt.Sprintf("%s_%s.md", timestamp, sess.Platform)
		mdPath := filepath.Join(m.config.TranscriptsDir, mdFilename)

		if err := os.MkdirAll(m.config.TranscriptsDir, 0755); err != nil {
			fmt.Printf("Failed to create transcripts directory: %v\n", err)
			return
		}

		if err := os.WriteFile(mdPath, []byte(markdown), 0644); err != nil {
			fmt.Printf("Failed to save Markdown transcript for session %s: %v\n", sess.CallID, err)
			return
		}

		// Update session in database
		endTime := time.Now()
		dbSess := &storage.Session{
			CallID:         sess.CallID,
			EndTime:        &endTime,
			AudioPath:      sess.AudioPath,
			TranscriptPath: mdPath,
		}
		if err := m.db.UpdateSession(dbSess); err != nil {
			fmt.Printf("Failed to update session in database: %v\n", err)
		}

		// Enqueue for cloud sync if configured
		if m.syncQueue != nil && m.config.IsCloudEnabled() && m.config.Cloud.TargetFolderID != "" {
			if err := m.syncQueue.Enqueue(sess.CallID, mdPath, m.config.Cloud.TargetFolderID); err != nil {
				fmt.Printf("Failed to enqueue cloud sync for session %s: %v\n", sess.CallID, err)
			} else {
				fmt.Printf("Enqueued cloud sync for session %s\n", sess.CallID)
			}
		}

		fmt.Printf("Transcription complete for session %s. Result saved to %s and %s\n", sess.CallID, transcriptPath, mdPath)
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
