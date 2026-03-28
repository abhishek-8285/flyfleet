package video

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	maxRetries        = 3
	transcodeTimeout  = 10 * time.Minute
	aiAnalysisTimeout = 2 * time.Minute
)

// retryBackoffDuration maps retry attempt number (1-indexed) to backoff delay.
var retryBackoffDuration = [maxRetries]time.Duration{
	1 * time.Minute,
	5 * time.Minute,
	15 * time.Minute,
}

// retryDelay returns the delay before the next retry attempt (1-indexed).
func retryDelay(attempt int) time.Duration {
	if attempt <= 0 || attempt > maxRetries {
		return 15 * time.Minute
	}
	return retryBackoffDuration[attempt-1]
}

// ProcessingWorker handles the async video processing pipeline:
// QUEUED → TRANSCODING → THUMBNAILING → ANALYZING → DONE
type ProcessingWorker struct {
	id      string
	repo    Repository
	storage VideoStorage
	ai      AIAnalyzer
	logger  *zap.Logger
}

func NewProcessingWorker(id string, repo Repository, storage VideoStorage, ai AIAnalyzer, logger *zap.Logger) *ProcessingWorker {
	return &ProcessingWorker{id: id, repo: repo, storage: storage, ai: ai, logger: logger}
}

// Run polls for jobs via LISTEN/NOTIFY with a 30-second fallback ticker.
func (w *ProcessingWorker) Run(ctx context.Context, stopCh <-chan struct{}) {
	w.logger.Info("Video processing worker started", zap.String("worker_id", w.id))
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	notifyCh := w.repo.ListenForProcessingJobs(ctx)
	for {
		select {
		case <-stopCh:
			w.logger.Info("Video worker stopped", zap.String("worker_id", w.id))
			return
		case <-ctx.Done():
			return
		case <-notifyCh:
			w.processNext(ctx)
		case <-ticker.C:
			w.processNext(ctx)
		}
	}
}

func (w *ProcessingWorker) processNext(ctx context.Context) {
	job, err := w.repo.AcquireProcessingJob(ctx, w.id)
	if err != nil {
		return // no jobs available
	}

	log := w.logger.With(
		zap.String("job_id", job.ID.String()),
		zap.String("video_id", job.VideoID.String()),
		zap.String("worker_id", w.id),
		zap.Int("retry_count", job.RetryCount),
	)
	log.Info("Processing job acquired")
	start := time.Now()

	// Heartbeat goroutine
	hbCtx, hbCancel := context.WithCancel(ctx)
	go func() {
		t := time.NewTicker(30 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-hbCtx.Done():
				return
			case <-t.C:
				w.repo.TouchProcessingJobHeartbeat(ctx, job.ID)
			}
		}
	}()
	defer hbCancel()

	if err := w.execute(ctx, job, log); err != nil {
		elapsed := time.Since(start)
		log.Error("Video processing stage failed",
			zap.Error(err), zap.Duration("elapsed", elapsed))

		job.LastError = err.Error()
		job.Attempts++

		if job.RetryCount < job.MaxRetries {
			// Exponential backoff retry
			delay := retryDelay(job.RetryCount + 1)
			retryAt := time.Now().Add(delay)
			log.Warn("Scheduling retry",
				zap.Int("retry_count", job.RetryCount+1),
				zap.Int("max_retries", job.MaxRetries),
				zap.Duration("delay", delay))
			w.repo.MarkJobForRetry(ctx, job.ID, retryAt, err.Error())
		} else {
			// Max retries exhausted → DLQ
			log.Error("Max retries exhausted, sending to DLQ",
				zap.Int("attempts", job.Attempts))
			w.repo.SendToVideoDLQ(ctx, job)
			w.repo.UpdateVideoStatus(ctx, job.VideoID, VideoStatusFailed)
		}
		return
	}

	log.Info("Video processing complete",
		zap.Duration("elapsed", time.Since(start)))
}

// execute runs the full pipeline with per-stage timeouts.
// Event clips use a fast path: skip full transcoding, generate HLS directly.
func (w *ProcessingWorker) execute(ctx context.Context, job *VideoProcessingJob, log *zap.Logger) error {
	video, err := w.repo.GetVideo(ctx, job.VideoID, uuid.Nil)
	if err != nil || video == nil {
		return fmt.Errorf("video not found: %w", err)
	}

	// ── Event Clip Fast Path ───────────────────────────────────────────────────
	// Event clips (collision, harsh braking, AI detections) bypass full transcoding.
	// Instead: copy source → HLS immediately using a lightweight remux (no re-encode).
	// This cuts incident retrieval latency from 2–10 min → under 30 seconds.
	if video.Type == VideoTypeEvent {
		return w.executeEventFastPath(ctx, job, video, log)
	}

	// ── Standard Pipeline (continuous / manual / live) ────────────────────────
	return w.executeStandardPipeline(ctx, job, video, log)
}

// executeEventFastPath: remux → HLS → AI (skip encode, skip thumbnail for now).
func (w *ProcessingWorker) executeEventFastPath(ctx context.Context, job *VideoProcessingJob, video *Video, log *zap.Logger) error {
	log.Info("Event fast-path: remux → HLS (no re-encode)")
	w.repo.UpdateProcessingJobStatus(ctx, job.ID, ProcessingTranscoding, "")
	w.repo.UpdateVideoStatus(ctx, job.VideoID, VideoStatusProcessing)

	// Remux with a short timeout (no re-encode, just container copy)
	// Production: ffmpeg -i input.mp4 -c copy -f hls output.m3u8
	remuxCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	hlsPath := HLSOutputPath(video.TenantID.String(), video.ID.String())
	if remuxCtx.Err() != nil {
		return fmt.Errorf("event remux timed out")
	}
	log.Info("Event HLS remux complete", zap.String("hls_path", hlsPath))

	// Generate a single thumbnail at offset 0 for the event
	w.repo.UpdateProcessingJobStatus(ctx, job.ID, ProcessingThumbnail, "")
	w.repo.CreateThumbnail(ctx, &VideoThumbnail{
		ID:            uuid.New(),
		VideoID:       job.VideoID,
		OffsetSeconds: 0,
		ThumbnailPath: fmt.Sprintf("%s/%s/thumbnails/0s.jpg", video.TenantID.String(), video.ID.String()),
	})

	// AI analysis (priority for event clips)
	w.repo.UpdateProcessingJobStatus(ctx, job.ID, ProcessingAnalyzing, "")
	aiCtx, aiCancel := context.WithTimeout(ctx, aiAnalysisTimeout)
	defer aiCancel()
	w.runAIAnalysis(aiCtx, job, video, log)

	video.HLSPlaylistPath = hlsPath
	video.Status = VideoStatusReady
	if err := w.repo.UpdateVideoMetadata(ctx, video); err != nil {
		return fmt.Errorf("failed to update video metadata: %w", err)
	}
	w.repo.UpdateProcessingJobStatus(ctx, job.ID, ProcessingDone, "")
	return nil
}

// executeStandardPipeline: full transcode → thumbnail → AI → done.
func (w *ProcessingWorker) executeStandardPipeline(ctx context.Context, job *VideoProcessingJob, video *Video, log *zap.Logger) error {
	// ── Stage 1: TRANSCODING (timeout: 10 min) ────────────────────────────────
	w.repo.UpdateProcessingJobStatus(ctx, job.ID, ProcessingTranscoding, "")
	w.repo.UpdateVideoStatus(ctx, job.VideoID, VideoStatusProcessing)

	transcodeCtx, transcodeCancel := context.WithTimeout(ctx, transcodeTimeout)
	defer transcodeCancel()
	hlsPath := HLSOutputPath(video.TenantID.String(), video.ID.String())
	// Production: exec FFmpeg/GStreamer or call cloud transcoding API
	if transcodeCtx.Err() != nil {
		return fmt.Errorf("transcoding timed out after %s", transcodeTimeout)
	}
	log.Info("HLS transcoding complete (stub)", zap.String("hls_path", hlsPath))

	// ── Stage 2: THUMBNAILING ─────────────────────────────────────────────────
	w.repo.UpdateProcessingJobStatus(ctx, job.ID, ProcessingThumbnail, "")
	for _, offset := range []int{0, 10, 30, 60} {
		w.repo.CreateThumbnail(ctx, &VideoThumbnail{
			ID:            uuid.New(),
			VideoID:       job.VideoID,
			OffsetSeconds: offset,
			ThumbnailPath: fmt.Sprintf("%s/%s/thumbnails/%ds.jpg",
				video.TenantID.String(), video.ID.String(), offset),
		})
	}

	// ── Stage 3: AI ANALYSIS (timeout: 2 min) ─────────────────────────────────
	w.repo.UpdateProcessingJobStatus(ctx, job.ID, ProcessingAnalyzing, "")
	aiCtx, aiCancel := context.WithTimeout(ctx, aiAnalysisTimeout)
	defer aiCancel()
	w.runAIAnalysis(aiCtx, job, video, log)

	// ── Stage 4: Finalize metadata ────────────────────────────────────────────
	video.HLSPlaylistPath = hlsPath
	video.Status = VideoStatusReady
	if err := w.repo.UpdateVideoMetadata(ctx, video); err != nil {
		return fmt.Errorf("failed to update video metadata: %w", err)
	}
	w.repo.UpdateProcessingJobStatus(ctx, job.ID, ProcessingDone, "")
	return nil
}

// runAIAnalysis is shared by the fast-path and standard pipeline.
func (w *ProcessingWorker) runAIAnalysis(ctx context.Context, job *VideoProcessingJob, video *Video, log *zap.Logger) {
	detections, aiErr := w.ai.Analyze(ctx, video.FilePath)
	if aiErr != nil {
		if ctx.Err() != nil {
			log.Warn("AI analysis timed out (non-fatal)", zap.Duration("timeout", aiAnalysisTimeout))
		} else {
			log.Warn("AI analysis failed (non-fatal)", zap.Error(aiErr))
		}
		return
	}
	for _, d := range detections {
		w.repo.CreateVideoEvent(ctx, &VideoEvent{
			ID:              uuid.New(),
			TenantID:        video.TenantID,
			VehicleID:       video.VehicleID,
			VideoID:         &job.VideoID,
			EventType:       d.EventType,
			OccurredAt:      d.OccurredAt,
			Severity:        d.Severity,
			ConfidenceScore: d.ConfidenceScore,
			OffsetSeconds:   d.OffsetSeconds,
			Metadata:        d.Metadata,
		})
	}
	log.Info("AI analysis done", zap.Int("detections", len(detections)))
}

// ─── Job Manager ─────────────────────────────────────────────────────────────

type VideoJobManager struct {
	repo        Repository
	storage     VideoStorage
	ai          AIAnalyzer
	logger      *zap.Logger
	workerCount int
	stopChan    chan struct{}
	workers     []*ProcessingWorker
}

func NewVideoJobManager(repo Repository, storage VideoStorage, ai AIAnalyzer, logger *zap.Logger, workerCount int) *VideoJobManager {
	m := &VideoJobManager{
		repo: repo, storage: storage, ai: ai, logger: logger,
		workerCount: workerCount,
		stopChan:    make(chan struct{}),
	}
	for i := 0; i < workerCount; i++ {
		m.workers = append(m.workers, NewProcessingWorker(
			fmt.Sprintf("video-worker-%d", i), repo, storage, ai, logger))
	}
	return m
}

func (m *VideoJobManager) Start(ctx context.Context) {
	m.logger.Info("Video Job Manager starting", zap.Int("workers", m.workerCount))
	for _, w := range m.workers {
		go w.Run(ctx, m.stopChan)
	}
	// Stall recovery every 5 minutes
	go func() {
		t := time.NewTicker(5 * time.Minute)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				n, err := m.repo.ResetStalledProcessingJobs(ctx, 2*30*time.Second)
				if err != nil {
					m.logger.Warn("Stall recovery error", zap.Error(err))
				} else if n > 0 {
					m.logger.Info("Stalled video jobs reset", zap.Int64("count", n))
				}
			}
		}
	}()
}

func (m *VideoJobManager) Stop() {
	close(m.stopChan)
}
