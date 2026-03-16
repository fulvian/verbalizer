// Package storage provides file and database storage.
package storage

import (
	"database/sql"
)

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
		CREATE TABLE IF NOT EXISTS recordings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			call_id TEXT NOT NULL UNIQUE,
			platform TEXT NOT NULL,
			title TEXT,
			audio_path TEXT NOT NULL,
			transcript_path TEXT,
			duration INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			transcribed_at DATETIME
		);
		
		CREATE INDEX IF NOT EXISTS idx_recordings_created ON recordings(created_at);
		CREATE INDEX IF NOT EXISTS idx_recordings_platform ON recordings(platform);
	`)
	return err
}
