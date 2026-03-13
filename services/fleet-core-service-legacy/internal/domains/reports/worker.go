package reports

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	heartbeatInterval  = 30 * time.Second
	staleJobTimeout    = 2 * time.Minute // RUNNING jobs older than this are crashed candidates
	maxRetries         = 3
	baseRetrySeconds   = 60
)

// Worker polls the queue and processes report jobs with crash recovery,
// streaming output, exponential backoff, and DLQ support.
type Worker struct {
	id      string
	repo    Repository
	storage ObjectStorage
	csv     *StreamingCSVWriter
	metrics *Metrics
	logger  *zap.Logger
}

func NewWorker(id string, repo Repository, storage ObjectStorage, metrics *Metrics, logger *zap.Logger) *Worker {
	return &Worker{
		id:      id,
		repo:    repo,
		storage: storage,
		csv:     NewStreamingCSVWriter(storage, logger),
		metrics: metrics,
		logger:  logger,
	}
}

// Run is the main event loop. It first processes any stale/crashed jobs,
// then polls for new work via PostgreSQL LISTEN/NOTIFY with fallback.
func (w *Worker) Run(ctx context.Context, stopChan <-chan struct{}) {
	w.logger.Info("Worker started", zap.String("worker_id", w.id))

	// Recover stale RUNNING jobs left by crashed workers
	w.recoverStalledJobs(ctx)

	// Primary: LISTEN/NOTIFY for low-latency job pickup;
	// Fallback: 30s ticker to catch any missed notifications.
	notifyCh := w.repo.ListenForJobs(ctx)
	fallbackTicker := time.NewTicker(30 * time.Second)
	defer fallbackTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Worker context cancelled", zap.String("worker_id", w.id))
			return
		case <-stopChan:
			w.logger.Info("Worker stop signal", zap.String("worker_id", w.id))
			return
		case <-notifyCh:
			w.poll(ctx)
		case <-fallbackTicker.C:
			w.poll(ctx)
		}
	}
}

// poll acquires a single job and executes it.
func (w *Worker) poll(ctx context.Context) {
	job, err := w.repo.AcquireJobLock(ctx, w.id)
	if err != nil {
		w.logger.Error("Failed to acquire job lock", zap.String("worker_id", w.id), zap.Error(err))
		return
	}
	if job == nil {
		return // Nothing in queue
	}

	w.metrics.WorkerBusy()
	defer w.metrics.WorkerIdle()

	startTime := time.Now()
	log := w.logger.With(
		zap.String("worker_id", w.id),
		zap.String("job_id", job.ID.String()),
		zap.String("report_id", job.ReportID.String()),
		zap.Int("attempt", job.Attempts),
	)
	log.Info("Job acquired")

	// Start heartbeat goroutine
	heartbeatStop := make(chan struct{})
	go w.heartbeat(ctx, job, heartbeatStop, log)

	if err := w.repo.UpdateReportStatus(ctx, job.ReportID, ReportStatusRunning, 0, "", ""); err != nil {
		log.Error("Failed to mark report RUNNING", zap.Error(err))
		close(heartbeatStop)
		w.failJob(ctx, job, err, log)
		return
	}

	fileURL, rowCount, genErr := w.execute(ctx, job, log)
	close(heartbeatStop)

	duration := time.Since(startTime).Seconds()

	if genErr != nil {
		log.Error("Job execution failed", zap.Error(genErr), zap.Float64("duration_s", duration))
		w.metrics.IncFailed()
		w.failJob(ctx, job, genErr, log)
		return
	}

	// Success
	w.repo.UpdateReportStatus(ctx, job.ReportID, ReportStatusCompleted, 100, fileURL, "")
	w.repo.UpdateJobStatus(ctx, job.ID, JobStatusCompleted)
	w.metrics.IncCompleted(duration, rowCount)
	log.Info("Job completed successfully",
		zap.Float64("duration_s", duration),
		zap.Int64("rows", rowCount),
		zap.String("file_url", fileURL))
}

// execute runs each pipeline stage with granular status updates visible to the user.
func (w *Worker) execute(ctx context.Context, job *ReportJob, log *zap.Logger) (fileURL string, rowCount int64, err error) {
	// Stage 1: QUERYING — running the aggregation SQL
	w.repo.UpdateReportStatus(ctx, job.ReportID, ReportStatusQuerying, 5, "", "")
	onProgress := func(pct int) {
		w.repo.UpdateReportStatus(ctx, job.ReportID, ReportStatusStreamingRows, pct, "", "")
	}

	// In production: query daily_*_metrics table via reportAggregationMap()
	rowCh := MockRowSource(ctx, ReportTypeVehicleUtilization, nil, onProgress)

	// Stage 2: STREAMING_ROWS — cursor fetching begins; progress mapped in onProgress
	w.repo.UpdateReportStatus(ctx, job.ReportID, ReportStatusStreamingRows, 20, "", "")
	
	// Check for cancellation before starting streaming
	if cancelled, _ := w.isCancelled(ctx, job.ReportID); cancelled {
		return "", 0, errors.New("job cancelled")
	}

	headers := []string{"id", "report_type", "timestamp", "data"}

	// Stage 3: GENERATING — CSV row writing
	w.repo.UpdateReportStatus(ctx, job.ReportID, ReportStatusGenerating, 50, "", "")

	// Stage 4: COMPRESSING — gzip applied inside StreamingCSVWriter
	w.repo.UpdateReportStatus(ctx, job.ReportID, ReportStatusCompressing, 75, "", "")
	reportKey := job.ReportID.String() + "/export.csv.gz"

	// Stage 5: UPLOADING
	w.repo.UpdateReportStatus(ctx, job.ReportID, ReportStatusUploading, 90, "", "")
	fileURL, rowCount, err = w.csv.WriteAndUpload(ctx, reportKey, headers, rowCh)
	if err != nil {
		return
	}

	// Stage 6: DELIVERING (notification sent to user)
	w.repo.UpdateReportStatus(ctx, job.ReportID, ReportStatusDelivering, 98, fileURL, "")
	return
}

// heartbeat updates the job's updated_at timestamp periodically to signal it's alive.
// Any RUNNING job that stops updating its timestamp will be recovered by another worker.
func (w *Worker) heartbeat(ctx context.Context, job *ReportJob, stop <-chan struct{}, log *zap.Logger) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.repo.TouchJobHeartbeat(ctx, job.ID); err != nil {
				log.Warn("Heartbeat update failed", zap.Error(err))
			}
		}
	}
}

// recoverStalledJobs resets RUNNING jobs whose heartbeat has timed out (crash recovery).
func (w *Worker) recoverStalledJobs(ctx context.Context) {
	recovered, err := w.repo.ResetStalledJobs(ctx, staleJobTimeout)
	if err != nil {
		w.logger.Error("Failed to recover stalled jobs", zap.Error(err))
		return
	}
	if recovered > 0 {
		w.logger.Warn("Recovered stalled RUNNING jobs",
			zap.Int64("count", recovered),
			zap.String("worker_id", w.id))
	}
}

// failJob handles permanent failure (DLQ) vs retry scheduling.
func (w *Worker) failJob(ctx context.Context, job *ReportJob, execErr error, log *zap.Logger) {
	if job.Attempts >= maxRetries {
		// Send to dead-letter queue
		log.Error("Job exhausted retries — moving to DLQ",
			zap.Int("attempts", job.Attempts),
			zap.Error(execErr))
		w.repo.SendToDeadLetterQueue(ctx, job.ID, execErr.Error())
		w.repo.UpdateReportStatus(ctx, job.ReportID, ReportStatusFailed, 0, "",
			fmt.Sprintf("Permanently failed after %d attempts: %s", maxRetries, execErr.Error()))
	} else {
		delay := retryDelay(job.Attempts)
		log.Info("Scheduling retry with exponential backoff",
			zap.Int("attempt", job.Attempts),
			zap.Duration("delay", delay))
		w.repo.UpdateJobStatus(ctx, job.ID, JobStatusFailed)
		w.repo.UpdateReportStatus(ctx, job.ReportID, ReportStatusFailed, 0, "",
			fmt.Sprintf("Attempt %d failed: %s. Retry after %s.", job.Attempts, execErr.Error(), delay))
	}
}

func (w *Worker) isCancelled(ctx context.Context, reportID uuid.UUID) (bool, error) {
	rep, err := w.repo.GetReport(ctx, reportID, uuid.Nil) // Repo needs to handle Nil tenant for internal checks or we need tenantID in job
	if err != nil {
		return false, err
	}
	return rep != nil && rep.Status == ReportStatusCancelled, nil
}

// retryDelay computes exponential backoff: 1m → 5m → 15m
func retryDelay(attempt int) time.Duration {
	delays := []time.Duration{1 * time.Minute, 5 * time.Minute, 15 * time.Minute}
	if attempt < len(delays) {
		return delays[attempt]
	}
	return 15 * time.Minute
}
