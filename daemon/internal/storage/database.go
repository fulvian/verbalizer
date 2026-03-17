// Package storage provides file and database storage.
package storage

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Session represents a recording session in the database.
type Session struct {
	CallID         string
	Platform       string
	Title          string
	StartTime      time.Time
	EndTime        *time.Time
	AudioPath      string
	TranscriptPath string
}

// Database handles SQLite operations.
type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection.
func NewDatabase(path string) (*Database, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	// Create tables
	if err := createTables(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Database{db: db}, nil
}

// Close closes the database connection.
func (d *Database) Close() error {
	return d.db.Close()
}

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			call_id TEXT PRIMARY KEY,
			platform TEXT NOT NULL,
			title TEXT,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			audio_path TEXT,
			transcript_path TEXT
		);
		
		CREATE INDEX IF NOT EXISTS idx_sessions_start_time ON sessions(start_time);
		CREATE INDEX IF NOT EXISTS idx_sessions_platform ON sessions(platform);
	`)
	return err
}

// SaveSession inserts a new session record.
func (d *Database) SaveSession(s *Session) error {
	_, err := d.db.Exec(`
		INSERT INTO sessions (call_id, platform, title, start_time, end_time, audio_path, transcript_path)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, s.CallID, s.Platform, s.Title, s.StartTime, s.EndTime, s.AudioPath, s.TranscriptPath)
	return err
}

// UpdateSession updates an existing session record.
func (d *Database) UpdateSession(s *Session) error {
	_, err := d.db.Exec(`
		UPDATE sessions 
		SET end_time = ?, audio_path = ?, transcript_path = ?
		WHERE call_id = ?
	`, s.EndTime, s.AudioPath, s.TranscriptPath, s.CallID)
	return err
}

// GetSession retrieves a session by call_id.
func (d *Database) GetSession(callID string) (*Session, error) {
	row := d.db.QueryRow(`
		SELECT call_id, platform, title, start_time, end_time, audio_path, transcript_path
		FROM sessions
		WHERE call_id = ?
	`, callID)

	var s Session
	err := row.Scan(&s.CallID, &s.Platform, &s.Title, &s.StartTime, &s.EndTime, &s.AudioPath, &s.TranscriptPath)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
