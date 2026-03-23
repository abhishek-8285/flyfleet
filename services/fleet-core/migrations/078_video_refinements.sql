-- +goose Up
-- Video Refinements: processing priority queue

-- Add priority to processing jobs: higher = processed sooner
-- Default 10 = continuous, 50 = event, 100 = manual/panic
ALTER TABLE video_processing_jobs ADD COLUMN IF NOT EXISTS priority INT NOT NULL DEFAULT 10;

-- Re-index for priority-aware job acquisition
DROP INDEX IF EXISTS idx_video_jobs_retry;
CREATE INDEX IF NOT EXISTS idx_video_jobs_acquire ON video_processing_jobs(priority DESC, created_at ASC, retry_after)
    WHERE status = 'queued';

-- +goose Down
DROP INDEX IF EXISTS idx_video_jobs_acquire;
CREATE INDEX IF NOT EXISTS idx_video_jobs_retry ON video_processing_jobs(status, retry_after) WHERE status = 'queued';
ALTER TABLE video_processing_jobs DROP COLUMN IF EXISTS priority;
