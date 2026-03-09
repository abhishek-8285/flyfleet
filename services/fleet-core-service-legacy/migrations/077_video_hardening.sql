-- +goose Up
-- Video Module Hardening: multipart upload, retry policy, DLQ, incident linking, indexes

-- 1. New tamper event types
ALTER TYPE video_event_type ADD VALUE IF NOT EXISTS 'camera_blocked';
ALTER TYPE video_event_type ADD VALUE IF NOT EXISTS 'camera_removed';
ALTER TYPE video_event_type ADD VALUE IF NOT EXISTS 'power_disconnect';
ALTER TYPE video_event_type ADD VALUE IF NOT EXISTS 'lens_covered';

-- 2. Multi-camera incident linking
ALTER TABLE videos ADD COLUMN IF NOT EXISTS incident_id UUID;
ALTER TABLE videos ADD COLUMN IF NOT EXISTS camera_group_id UUID;

CREATE INDEX IF NOT EXISTS idx_videos_incident ON videos(incident_id) WHERE incident_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_videos_camera_group ON videos(camera_group_id) WHERE camera_group_id IS NOT NULL;

-- 3. BRIN index for time-range queries (far more efficient than B-tree for append-only time data)
CREATE INDEX IF NOT EXISTS idx_videos_start_time_brin ON videos USING BRIN(start_time);

-- 4. Retry policy fields on processing jobs
ALTER TABLE video_processing_jobs ADD COLUMN IF NOT EXISTS retry_count  INT NOT NULL DEFAULT 0;
ALTER TABLE video_processing_jobs ADD COLUMN IF NOT EXISTS max_retries  INT NOT NULL DEFAULT 3;
ALTER TABLE video_processing_jobs ADD COLUMN IF NOT EXISTS retry_after  TIMESTAMP WITH TIME ZONE;
ALTER TABLE video_processing_jobs ADD COLUMN IF NOT EXISTS last_error   TEXT;

-- Index for retry scheduling: workers query status=queued AND (retry_after IS NULL OR retry_after <= NOW())
CREATE INDEX IF NOT EXISTS idx_video_jobs_retry ON video_processing_jobs(status, retry_after)
    WHERE status = 'queued';

-- 5. Dead Letter Queue for permanently failed video processing jobs
CREATE TABLE IF NOT EXISTS video_processing_dlq (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id          UUID NOT NULL,
    video_id        UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    attempts        INT NOT NULL,
    last_error      TEXT,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_video_dlq_video ON video_processing_dlq(video_id);

-- 6. Multipart upload sessions (S3-style chunked upload)
CREATE TABLE IF NOT EXISTS video_multipart_sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    vehicle_id      UUID NOT NULL,
    storage_key     TEXT NOT NULL,          -- final object storage key
    upload_id       TEXT NOT NULL UNIQUE,   -- opaque upload session identifier
    total_parts     INT,                    -- NULL until complete is called
    parts_received  INT NOT NULL DEFAULT 0,
    status          VARCHAR(32) NOT NULL DEFAULT 'active',  -- active | completed | aborted
    metadata        JSONB,                  -- original UploadVideoRequest
    expires_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW() + INTERVAL '24 hours',
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_multipart_tenant ON video_multipart_sessions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_multipart_upload_id ON video_multipart_sessions(upload_id);

-- 7. Multipart parts tracking
CREATE TABLE IF NOT EXISTS video_multipart_parts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id      UUID NOT NULL REFERENCES video_multipart_sessions(id) ON DELETE CASCADE,
    part_number     INT NOT NULL,
    size_bytes      BIGINT NOT NULL,
    etag            TEXT,                   -- storage backend ETag for verification
    received_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (session_id, part_number)
);

CREATE INDEX IF NOT EXISTS idx_multipart_parts_session ON video_multipart_parts(session_id);

-- +goose Down
DROP TABLE IF EXISTS video_multipart_parts;
DROP TABLE IF EXISTS video_multipart_sessions;
DROP TABLE IF EXISTS video_processing_dlq;
DROP INDEX IF EXISTS idx_video_jobs_retry;
DROP INDEX IF EXISTS idx_videos_start_time_brin;
DROP INDEX IF EXISTS idx_videos_camera_group;
DROP INDEX IF EXISTS idx_videos_incident;
