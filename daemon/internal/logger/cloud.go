// Package logger provides structured logging for the Verbalizer daemon.
package logger

// CloudLogger provides cloud-specific logging helpers.
type CloudLogger struct{}

// Cloud returns a cloud logger.
func Cloud() *CloudLogger {
	return &CloudLogger{}
}

// OAuth operations
func (c *CloudLogger) OAuthStart(email string) {
	Info("OAuth flow started", String("provider", "google"), String("email", email))
}

func (c *CloudLogger) OAuthComplete(email string, success bool) {
	level := LevelInfo
	if !success {
		level = LevelError
	}
	entry := WithFields(
		String("provider", "google"),
		String("email", email),
		Bool("success", success),
	)
	switch level {
	case LevelError:
		entry.error("OAuth flow completed")
	default:
		entry.info("OAuth flow completed")
	}
}

func (c *CloudLogger) OAuthError(err error, stage string) {
	WithFields(
		String("provider", "google"),
		String("stage", stage),
		Err(err),
	).error("OAuth error")
}

// Upload operations
func (c *CloudLogger) UploadStart(callID, localPath string) {
	Info("Upload started",
		String("call_id", callID),
		String("local_path", localPath),
	)
}

func (c *CloudLogger) UploadComplete(callID, remoteID string, size int64) {
	Info("Upload completed",
		String("call_id", callID),
		String("remote_id", remoteID),
		Int64("size_bytes", size),
	)
}

func (c *CloudLogger) UploadError(callID string, err error, retryable bool) {
	WithFields(
		String("call_id", callID),
		Err(err),
		Bool("retryable", retryable),
	).error("Upload failed")
}

func (c *CloudLogger) UploadRetry(callID string, attempt, maxAttempts int, delaySeconds int) {
	Warn("Upload retry scheduled",
		String("call_id", callID),
		Int("attempt", attempt),
		Int("max_attempts", maxAttempts),
		Int("delay_seconds", delaySeconds),
	)
}

// Sync queue operations
func (c *CloudLogger) SyncEnqueue(callID, folderID string) {
	Info("Sync job enqueued",
		String("call_id", callID),
		String("folder_id", folderID),
	)
}

func (c *CloudLogger) SyncDequeue(callID string) {
	Debug("Sync job dequeued",
		String("call_id", callID),
	)
}

func (c *CloudLogger) SyncComplete(callID string) {
	Info("Sync completed",
		String("call_id", callID),
	)
}

func (c *CloudLogger) SyncPermanentFail(callID string, err error) {
	WithFields(
		String("call_id", callID),
		Err(err),
	).error("Sync permanently failed")
}
