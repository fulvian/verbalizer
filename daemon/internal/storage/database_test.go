package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDatabase(t *testing.T) {
	// Setup temporary database path
	tmpDir, err := os.MkdirTemp("", "verbalizer-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	// Test NewDatabase
	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test SaveSession
	startTime := time.Now().Round(time.Second)
	s := &Session{
		CallID:    "test-call-1",
		Platform:  "google-meet",
		Title:     "Test Meeting",
		StartTime: startTime,
	}

	if err := db.SaveSession(s); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Test GetSession
	saved, err := db.GetSession("test-call-1")
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if saved.CallID != s.CallID {
		t.Errorf("Expected CallID %s, got %s", s.CallID, saved.CallID)
	}
	if saved.Platform != s.Platform {
		t.Errorf("Expected Platform %s, got %s", s.Platform, saved.Platform)
	}
	if !saved.StartTime.Equal(startTime) {
		t.Errorf("Expected StartTime %v, got %v", startTime, saved.StartTime)
	}

	// Test UpdateSession
	endTime := startTime.Add(time.Hour)
	s.EndTime = &endTime
	s.AudioPath = "/path/to/audio.mp3"
	s.TranscriptPath = "/path/to/transcript.md"

	if err := db.UpdateSession(s); err != nil {
		t.Fatalf("Failed to update session: %v", err)
	}

	// Verify update
	updated, err := db.GetSession("test-call-1")
	if err != nil {
		t.Fatalf("Failed to get updated session: %v", err)
	}

	if updated.EndTime == nil || !updated.EndTime.Equal(endTime) {
		t.Errorf("Expected EndTime %v, got %v", endTime, updated.EndTime)
	}
	if updated.AudioPath != s.AudioPath {
		t.Errorf("Expected AudioPath %s, got %s", s.AudioPath, updated.AudioPath)
	}
	if updated.TranscriptPath != s.TranscriptPath {
		t.Errorf("Expected TranscriptPath %s, got %s", s.TranscriptPath, updated.TranscriptPath)
	}
}
