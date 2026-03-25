// Package storage provides file and database storage.
package storage

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// CloudProvider represents a cloud storage provider.
type CloudProvider string

const (
	ProviderGoogleDrive CloudProvider = "google_drive"
)

// CloudAccount represents a connected cloud account.
type CloudAccount struct {
	ID           int64
	Provider     CloudProvider
	AccountEmail string
	Scopes       string
	ConnectedAt  time.Time
	Status       string // "active", "revoked", "expired"
}

// CloudSyncJob represents a cloud sync operation.
type CloudSyncJob struct {
	ID               int64
	SessionCallID    string
	LocalPath        string
	Provider         CloudProvider
	TargetFolderID   string
	State            string // "pending", "uploading", "synced", "failed", "permanent_failed"
	AttemptCount     int
	NextRetryAt      *time.Time
	LastErrorCode    int
	LastErrorMessage string
	RemoteFileID     string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Session represents a recording session in the database.
type Session struct {
	CallID            string
	Platform          string
	Title             string
	StartTime         time.Time
	EndTime           *time.Time
	AudioPath         string
	TranscriptPath    string
	CloudSyncState    string // "none", "pending", "synced", "failed"
	CloudRemoteFileID string // ID of file in cloud storage
	CloudLastSyncAt   *time.Time
}

// Database handles SQLite operations.
type Database struct {
	db *sql.DB
}

// migrations defines the database schema migrations.
// Each migration runs in order and is idempotent.
var migrations = []string{
	// Migration 1: Initial schema (sessions table)
	`CREATE TABLE IF NOT EXISTS sessions (
		call_id TEXT PRIMARY KEY,
		platform TEXT NOT NULL,
		title TEXT,
		start_time DATETIME NOT NULL,
		end_time DATETIME,
		audio_path TEXT,
		transcript_path TEXT,
		cloud_sync_state TEXT DEFAULT 'none',
		cloud_remote_file_id TEXT,
		cloud_last_sync_at DATETIME
	);
	CREATE INDEX IF NOT EXISTS idx_sessions_start_time ON sessions(start_time);
	CREATE INDEX IF NOT EXISTS idx_sessions_platform ON sessions(platform);
	CREATE INDEX IF NOT EXISTS idx_sessions_cloud_sync ON sessions(cloud_sync_state);`,

	// Migration 2: Cloud accounts table
	`CREATE TABLE IF NOT EXISTS cloud_accounts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		provider TEXT NOT NULL,
		account_email TEXT NOT NULL,
		scopes TEXT NOT NULL,
		connected_at DATETIME NOT NULL,
		status TEXT NOT NULL DEFAULT 'active',
		UNIQUE(provider, account_email)
	);
	CREATE INDEX IF NOT EXISTS idx_cloud_accounts_provider ON cloud_accounts(provider);
	CREATE INDEX IF NOT EXISTS idx_cloud_accounts_status ON cloud_accounts(status);`,

	// Migration 3: Cloud sync jobs table
	`CREATE TABLE IF NOT EXISTS cloud_sync_jobs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_call_id TEXT NOT NULL,
		local_path TEXT NOT NULL,
		provider TEXT NOT NULL,
		target_folder_id TEXT NOT NULL,
		state TEXT NOT NULL DEFAULT 'pending',
		attempt_count INTEGER NOT NULL DEFAULT 0,
		next_retry_at DATETIME,
		last_error_code INTEGER DEFAULT 0,
		last_error_message TEXT,
		remote_file_id TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY(session_call_id) REFERENCES sessions(call_id)
	);
	CREATE INDEX IF NOT EXISTS idx_cloud_sync_jobs_state ON cloud_sync_jobs(state);
	CREATE INDEX IF NOT EXISTS idx_cloud_sync_jobs_next_retry ON cloud_sync_jobs(next_retry_at);
	CREATE INDEX IF NOT EXISTS idx_cloud_sync_jobs_session ON cloud_sync_jobs(session_call_id);`,

	// Migration 4: Schema version tracking
	`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at DATETIME NOT NULL
	);`,
}

// NewDatabase creates a new database connection.
func NewDatabase(path string) (*Database, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Database{db: db}, nil
}

// runMigrations executes all pending database migrations.
func runMigrations(db *sql.DB) error {
	// Create migrations table if not exists
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME NOT NULL
		);
	`)
	if err != nil {
		return err
	}

	// Get current version
	var currentVersion int
	row := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations")
	if err := row.Scan(&currentVersion); err != nil {
		return err
	}

	// Apply pending migrations
	for i := currentVersion; i < len(migrations); i++ {
		if _, err := db.Exec(migrations[i]); err != nil {
			return err
		}
		_, err = db.Exec("INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)", i+1, time.Now())
		if err != nil {
			return err
		}
	}

	return nil
}

// Close closes the database connection.
func (d *Database) Close() error {
	return d.db.Close()
}

// SaveSession inserts a new session record.
func (d *Database) SaveSession(s *Session) error {
	_, err := d.db.Exec(`
		INSERT INTO sessions (call_id, platform, title, start_time, end_time, audio_path, transcript_path, cloud_sync_state)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, s.CallID, s.Platform, s.Title, s.StartTime, s.EndTime, s.AudioPath, s.TranscriptPath, s.CloudSyncState)
	return err
}

// UpdateSession updates an existing session record.
func (d *Database) UpdateSession(s *Session) error {
	_, err := d.db.Exec(`
		UPDATE sessions 
		SET end_time = ?, audio_path = ?, transcript_path = ?, cloud_sync_state = ?, cloud_remote_file_id = ?, cloud_last_sync_at = ?
		WHERE call_id = ?
	`, s.EndTime, s.AudioPath, s.TranscriptPath, s.CloudSyncState, s.CloudRemoteFileID, s.CloudLastSyncAt, s.CallID)
	return err
}

// GetSession retrieves a session by call_id.
func (d *Database) GetSession(callID string) (*Session, error) {
	row := d.db.QueryRow(`
		SELECT call_id, platform, title, start_time, end_time, audio_path, transcript_path, 
		       COALESCE(cloud_sync_state, 'none'), COALESCE(cloud_remote_file_id, ''), cloud_last_sync_at
		FROM sessions
		WHERE call_id = ?
	`, callID)

	var s Session
	err := row.Scan(&s.CallID, &s.Platform, &s.Title, &s.StartTime, &s.EndTime, &s.AudioPath, &s.TranscriptPath, &s.CloudSyncState, &s.CloudRemoteFileID, &s.CloudLastSyncAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// ListSessions retrieves all sessions with optional limit.
func (d *Database) ListSessions(limit int) ([]*Session, error) {
	rows, err := d.db.Query(`
		SELECT call_id, platform, title, start_time, end_time, audio_path, transcript_path,
		       COALESCE(cloud_sync_state, 'none'), COALESCE(cloud_remote_file_id, ''), cloud_last_sync_at
		FROM sessions
		ORDER BY start_time DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.CallID, &s.Platform, &s.Title, &s.StartTime, &s.EndTime, &s.AudioPath, &s.TranscriptPath, &s.CloudSyncState, &s.CloudRemoteFileID, &s.CloudLastSyncAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, &s)
	}
	return sessions, nil
}

// SaveCloudAccount inserts or updates a cloud account.
func (d *Database) SaveCloudAccount(a *CloudAccount) error {
	_, err := d.db.Exec(`
		INSERT INTO cloud_accounts (provider, account_email, scopes, connected_at, status)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(provider, account_email) DO UPDATE SET
			scopes = excluded.scopes,
			status = excluded.status,
			connected_at = excluded.connected_at
	`, a.Provider, a.AccountEmail, a.Scopes, a.ConnectedAt, a.Status)
	return err
}

// GetCloudAccount retrieves an active cloud account by provider.
func (d *Database) GetCloudAccount(provider CloudProvider) (*CloudAccount, error) {
	row := d.db.QueryRow(`
		SELECT id, provider, account_email, scopes, connected_at, status
		FROM cloud_accounts
		WHERE provider = ? AND status = 'active'
		LIMIT 1
	`, provider)

	var a CloudAccount
	err := row.Scan(&a.ID, &a.Provider, &a.AccountEmail, &a.Scopes, &a.ConnectedAt, &a.Status)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// RevokeCloudAccount marks a cloud account as revoked.
func (d *Database) RevokeCloudAccount(provider CloudProvider, email string) error {
	_, err := d.db.Exec(`
		UPDATE cloud_accounts SET status = 'revoked' WHERE provider = ? AND account_email = ?
	`, provider, email)
	return err
}

// SaveCloudSyncJob inserts a new cloud sync job.
func (d *Database) SaveCloudSyncJob(j *CloudSyncJob) error {
	_, err := d.db.Exec(`
		INSERT INTO cloud_sync_jobs (session_call_id, local_path, provider, target_folder_id, state, attempt_count, next_retry_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, j.SessionCallID, j.LocalPath, j.Provider, j.TargetFolderID, j.State, j.AttemptCount, j.NextRetryAt, j.CreatedAt, j.UpdatedAt)
	return err
}

// UpdateCloudSyncJob updates a cloud sync job.
func (d *Database) UpdateCloudSyncJob(j *CloudSyncJob) error {
	_, err := d.db.Exec(`
		UPDATE cloud_sync_jobs 
		SET state = ?, attempt_count = ?, next_retry_at = ?, last_error_code = ?, last_error_message = ?, remote_file_id = ?, updated_at = ?
		WHERE id = ?
	`, j.State, j.AttemptCount, j.NextRetryAt, j.LastErrorCode, j.LastErrorMessage, j.RemoteFileID, j.UpdatedAt, j.ID)
	return err
}

// GetCloudSyncJob retrieves a sync job by ID.
func (d *Database) GetCloudSyncJob(id int64) (*CloudSyncJob, error) {
	row := d.db.QueryRow(`
		SELECT id, session_call_id, local_path, provider, target_folder_id, state, 
		       attempt_count, next_retry_at, last_error_code, last_error_message, 
		       remote_file_id, created_at, updated_at
		FROM cloud_sync_jobs
		WHERE id = ?
	`, id)

	var j CloudSyncJob
	var nextRetryAt sql.NullTime
	var lastErrorCode sql.NullInt64
	var lastErrorMessage sql.NullString
	var remoteFileID sql.NullString

	err := row.Scan(&j.ID, &j.SessionCallID, &j.LocalPath, &j.Provider, &j.TargetFolderID, &j.State,
		&j.AttemptCount, &nextRetryAt, &lastErrorCode, &lastErrorMessage, &remoteFileID, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if nextRetryAt.Valid {
		j.NextRetryAt = &nextRetryAt.Time
	}
	if lastErrorCode.Valid {
		j.LastErrorCode = int(lastErrorCode.Int64)
	}
	if lastErrorMessage.Valid {
		j.LastErrorMessage = lastErrorMessage.String
	}
	if remoteFileID.Valid {
		j.RemoteFileID = remoteFileID.String
	}

	return &j, nil
}

// GetPendingCloudSyncJobs retrieves jobs ready for retry.
func (d *Database) GetPendingCloudSyncJobs(limit int) ([]*CloudSyncJob, error) {
	rows, err := d.db.Query(`
		SELECT id, session_call_id, local_path, provider, target_folder_id, state,
		       attempt_count, next_retry_at, last_error_code, last_error_message,
		       remote_file_id, created_at, updated_at
		FROM cloud_sync_jobs
		WHERE state IN ('pending', 'failed') 
		  AND (next_retry_at IS NULL OR next_retry_at <= ?)
		ORDER BY created_at ASC
		LIMIT ?
	`, time.Now(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*CloudSyncJob
	for rows.Next() {
		var j CloudSyncJob
		var nextRetryAt sql.NullTime
		var lastErrorCode sql.NullInt64
		var lastErrorMessage sql.NullString
		var remoteFileID sql.NullString

		if err := rows.Scan(&j.ID, &j.SessionCallID, &j.LocalPath, &j.Provider, &j.TargetFolderID, &j.State,
			&j.AttemptCount, &nextRetryAt, &lastErrorCode, &lastErrorMessage, &remoteFileID, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, err
		}

		if nextRetryAt.Valid {
			j.NextRetryAt = &nextRetryAt.Time
		}
		if lastErrorCode.Valid {
			j.LastErrorCode = int(lastErrorCode.Int64)
		}
		if lastErrorMessage.Valid {
			j.LastErrorMessage = lastErrorMessage.String
		}
		if remoteFileID.Valid {
			j.RemoteFileID = remoteFileID.String
		}

		jobs = append(jobs, &j)
	}
	return jobs, nil
}

// GetCloudSyncJobBySessionCallID retrieves a sync job by session call ID.
func (d *Database) GetCloudSyncJobBySessionCallID(callID string) (*CloudSyncJob, error) {
	row := d.db.QueryRow(`
		SELECT id, session_call_id, local_path, provider, target_folder_id, state,
		       attempt_count, next_retry_at, last_error_code, last_error_message,
		       remote_file_id, created_at, updated_at
		FROM cloud_sync_jobs
		WHERE session_call_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, callID)

	var j CloudSyncJob
	var nextRetryAt sql.NullTime
	var lastErrorCode sql.NullInt64
	var lastErrorMessage sql.NullString
	var remoteFileID sql.NullString

	err := row.Scan(&j.ID, &j.SessionCallID, &j.LocalPath, &j.Provider, &j.TargetFolderID, &j.State,
		&j.AttemptCount, &nextRetryAt, &lastErrorCode, &lastErrorMessage, &remoteFileID, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if nextRetryAt.Valid {
		j.NextRetryAt = &nextRetryAt.Time
	}
	if lastErrorCode.Valid {
		j.LastErrorCode = int(lastErrorCode.Int64)
	}
	if lastErrorMessage.Valid {
		j.LastErrorMessage = lastErrorMessage.String
	}
	if remoteFileID.Valid {
		j.RemoteFileID = remoteFileID.String
	}

	return &j, nil
}
