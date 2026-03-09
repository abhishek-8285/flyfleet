package reports

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// queueDepthAlertMultiplier: alert fires when depth > workers × this value.
// Scales naturally with worker pool size.
const queueDepthAlertMultiplier = 10

type JobManager interface {
	Start(ctx context.Context)
	Stop()
	EnqueueReport(ctx context.Context, reportID uuid.UUID) error
}

type jobManager struct {
	repo        Repository
	storage     ObjectStorage
	logger      *zap.Logger
	workerCount int
	stopChan    chan struct{}
	wg          sync.WaitGroup
	workers     []*Worker
	metrics     *Metrics
	// semaphore limits max concurrent active jobs regardless of worker count
	semaphore chan struct{}
}

func NewJobManager(repo Repository, logger *zap.Logger, workerCount int) JobManager {
	storage := NewLocalStorage(logger)
	metrics := NewMetrics(logger, workerCount)

	jm := &jobManager{
		repo:        repo,
		storage:     storage,
		logger:      logger,
		workerCount: workerCount,
		stopChan:    make(chan struct{}),
		metrics:     metrics,
		semaphore:   make(chan struct{}, workerCount), // max_concurrent_workers = workerCount
	}

	for i := 0; i < workerCount; i++ {
		workerID := fmt.Sprintf("worker-%d", i)
		jm.workers = append(jm.workers, NewWorker(workerID, repo, storage, metrics, logger))
	}

	return jm
}

func (jm *jobManager) Start(ctx context.Context) {
	jm.logger.Info("Starting Report Job Manager",
		zap.Int("workers", jm.workerCount),
		zap.Int("max_concurrent", cap(jm.semaphore)))

	go jm.metrics.LogSnapshot(ctx)

	for _, w := range jm.workers {
		jm.wg.Add(1)
		go func(worker *Worker) {
			defer jm.wg.Done()
			worker.Run(ctx, jm.stopChan)
		}(w)
	}
}

func (jm *jobManager) Stop() {
	jm.logger.Info("Stopping Report Job Manager")
	close(jm.stopChan)
	jm.wg.Wait()
	jm.logger.Info("Report Job Manager stopped")
}

func (jm *jobManager) EnqueueReport(ctx context.Context, reportID uuid.UUID) error {
	job := &ReportJob{
		ID:        uuid.New(),
		ReportID:  reportID,
		JobStatus: JobStatusQueued,
		Attempts:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := jm.repo.EnqueueJob(ctx, job); err != nil {
		return err
	}

	newDepth := jm.metrics.queueDepth + 1
	jm.metrics.SetQueueDepth(newDepth)

	// Backpressure: alert scales with worker count
	alertThreshold := int64(jm.workerCount * queueDepthAlertMultiplier)
	if newDepth >= alertThreshold {
		jm.logger.Warn("Report queue depth high — consider scaling workers",
			zap.Int64("depth", newDepth),
			zap.Int64("alert_threshold", alertThreshold),
			zap.Int("workers", jm.workerCount))
	}

	jm.logger.Info("Report job enqueued",
		zap.String("job_id", job.ID.String()),
		zap.String("report_id", reportID.String()),
		zap.Int64("queue_depth", newDepth))
	return nil
}
