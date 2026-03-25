// Package syncqueue provides a persistent sync queue with retry logic.
package syncqueue

import (
	"math/rand"
	"time"

	"github.com/fulvian/verbalizer/daemon/internal/cloud/driveclient"
	"github.com/fulvian/verbalizer/daemon/internal/config"
	"github.com/fulvian/verbalizer/daemon/internal/logger"
	"github.com/fulvian/verbalizer/daemon/internal/storage"
)

// Worker processes sync jobs from the queue.
type Worker struct {
	db           *storage.Database
	driveClient  *driveclient.DriveClient
	config       *config.RetryConfig
	stopCh       chan struct{}
	processingCh chan int64
}

// NewWorker creates a new sync queue worker.
func NewWorker(db *storage.Database, dc *driveclient.DriveClient, cfg *config.RetryConfig) *Worker {
	return &Worker{
		db:           db,
		driveClient:  dc,
		config:       cfg,
		stopCh:       make(chan struct{}),
		processingCh: make(chan int64, 10),
	}
}

// Start begins processing jobs in the queue.
func (w *Worker) Start() {
	go w.processLoop()
}

// Stop gracefully stops the worker.
func (w *Worker) Stop() {
	close(w.stopCh)
}

// processLoop continuously polls for and processes jobs.
func (w *Worker) processLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.processBatch()
		}
	}
}

// processBatch processes a batch of pending jobs.
func (w *Worker) processBatch() {
	jobs, err := w.db.GetPendingCloudSyncJobs(5)
	if err != nil {
		logger.Error("Failed to get pending jobs", logger.Err(err))
		return
	}

	for _, job := range jobs {
		select {
		case <-w.stopCh:
			return
		case w.processingCh <- job.ID:
			go w.processJob(job)
		}
	}
}

// processJob processes a single sync job.
func (w *Worker) processJob(job *storage.CloudSyncJob) {
	defer func() {
		<-w.processingCh
	}()

	// Mark as uploading
	job.State = "uploading"
	job.AttemptCount++
	job.UpdatedAt = time.Now()
	if err := w.db.UpdateCloudSyncJob(job); err != nil {
		logger.Error("Failed to update job", logger.Int64("job_id", job.ID), logger.Err(err))
		return
	}

	// Attempt upload
	result, err := w.driveClient.Upload(job.LocalPath, job.TargetFolderID)
	if err != nil {
		w.handleUploadError(job, err)
		return
	}

	// Success
	job.State = "synced"
	job.RemoteFileID = result.FileID
	job.UpdatedAt = time.Now()
	job.LastErrorCode = 0
	job.LastErrorMessage = ""
	if err := w.db.UpdateCloudSyncJob(job); err != nil {
		logger.Error("Failed to update job on success", logger.Int64("job_id", job.ID), logger.Err(err))
	}

	// Update session
	if err := w.updateSessionCloudState(job.SessionCallID, "synced", result.FileID); err != nil {
		logger.Error("Failed to update session cloud state", logger.Err(err))
	}

	logger.Cloud().SyncComplete(job.SessionCallID)
}

// handleUploadError handles a failed upload attempt.
func (w *Worker) handleUploadError(job *storage.CloudSyncJob, uploadErr error) {
	cat := logger.CategorizeError(uploadErr)
	retryable := cat.IsRetryable()

	if !retryable || job.AttemptCount >= w.config.MaxAttempts {
		job.State = "permanent_failed"
		logger.Cloud().SyncPermanentFail(job.SessionCallID, uploadErr)
	} else {
		job.State = "failed"
		delay := w.calculateBackoff(job.AttemptCount)
		delaySec := int(delay.Seconds())
		nextRetry := time.Now().Add(delay)
		job.NextRetryAt = &nextRetry
		logger.Cloud().UploadRetry(job.SessionCallID, job.AttemptCount, w.config.MaxAttempts, delaySec)
	}

	job.LastErrorMessage = uploadErr.Error()
	job.UpdatedAt = time.Now()

	if err := w.db.UpdateCloudSyncJob(job); err != nil {
		logger.Error("Failed to update job", logger.Int64("job_id", job.ID), logger.Err(err))
	}
}

// calculateBackoff calculates exponential backoff with jitter.
func (w *Worker) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff: base * 2^attempt
	multiplier := 1 << attempt
	delay := w.config.BaseDelaySeconds * multiplier

	// Cap at max delay
	if delay > w.config.MaxDelaySeconds {
		delay = w.config.MaxDelaySeconds
	}

	// Add jitter (±25%)
	jitter := float64(delay) * 0.25
	delay = delay + int(jitter*(rand.Float64()*2-1))

	if delay < 1 {
		delay = 1
	}

	return time.Duration(delay) * time.Second
}

// updateSessionCloudState updates the session's cloud sync state.
func (w *Worker) updateSessionCloudState(callID, state, remoteFileID string) error {
	sess, err := w.db.GetSession(callID)
	if err != nil {
		return err
	}
	sess.CloudSyncState = state
	sess.CloudRemoteFileID = remoteFileID
	now := time.Now()
	sess.CloudLastSyncAt = &now
	return w.db.UpdateSession(sess)
}

// Enqueue adds a new sync job to the queue.
func (w *Worker) Enqueue(callID, localPath, folderID string) error {
	job := &storage.CloudSyncJob{
		SessionCallID:  callID,
		LocalPath:      localPath,
		Provider:       storage.ProviderGoogleDrive,
		TargetFolderID: folderID,
		State:          "pending",
		AttemptCount:   0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	return w.db.SaveCloudSyncJob(job)
}

// GetStats returns queue statistics.
func (w *Worker) GetStats() (*Stats, error) {
	jobs, err := w.db.GetPendingCloudSyncJobs(1000)
	if err != nil {
		return nil, err
	}

	stats := &Stats{
		Total: len(jobs),
	}
	for _, job := range jobs {
		switch job.State {
		case "pending":
			stats.Pending++
		case "uploading":
			stats.Uploading++
		case "synced":
			stats.Synced++
		case "failed":
			stats.Failed++
		case "permanent_failed":
			stats.PermanentFailed++
		}
	}

	return stats, nil
}

// Stats represents queue statistics.
type Stats struct {
	Pending         int
	Uploading       int
	Synced          int
	Failed          int
	PermanentFailed int
	Total           int
}

// Queue provides the interface for the sync queue.
type Queue interface {
	Enqueue(callID, localPath, folderID string) error
	ProcessNext() error
	GetJobStatus(callID string) (*storage.CloudSyncJob, error)
	RetryFailed() error
	GetStats() (*Stats, error)
}

// Ensure Worker implements Queue
var _ Queue = (*Worker)(nil)

// GetJobStatus returns the sync job for a session.
func (w *Worker) GetJobStatus(callID string) (*storage.CloudSyncJob, error) {
	return w.db.GetCloudSyncJobBySessionCallID(callID)
}

// ProcessNext processes the next pending job.
func (w *Worker) ProcessNext() error {
	jobs, err := w.db.GetPendingCloudSyncJobs(1)
	if err != nil {
		return err
	}
	if len(jobs) == 0 {
		return nil
	}
	w.processJob(jobs[0])
	return nil
}

// RetryFailed resets failed jobs for retry.
func (w *Worker) RetryFailed() error {
	jobs, err := w.db.GetPendingCloudSyncJobs(100)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if job.State == "failed" || job.State == "permanent_failed" {
			if job.AttemptCount < w.config.MaxAttempts {
				job.State = "pending"
				job.NextRetryAt = nil
				job.UpdatedAt = time.Now()
				if err := w.db.UpdateCloudSyncJob(job); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
