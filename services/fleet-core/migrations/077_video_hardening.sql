-- +goose Up
-- Video Module Hardening: multipart upload, retry policy, DLQ, incident linking, indexes

-- 1. New tamper event types
DO $$ 
BEGIN
    -- Ensure the type exists
    IF NOT EXISTS (SELECT 1 FROM pg_type t JOIN pg_namespace n ON n.oid = t.typnamespace WHERE t.typname = 'video_event_type') THEN
        CREATE TYPE video_event_type AS ENUM ('collision', 'harsh_braking');
    END IF;
END $$;

ALTER TYPE video_event_type ADD VALUE IF NOT EXISTS 'camera_blocked';
ALTER TYPE video_event_type ADD VALUE IF NOT EXISTS 'camera_removed';
ALTER TYPE video_event_type ADD VALUE IF NOT EXISTS 'power_disconnect';
ALTER TYPE video_event_type ADD VALUE IF NOT EXISTS 'lens_covered';

-- 2. Multi-camera incident linking
CREATE TABLE IF NOT EXISTS videos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS incident_id UUID;
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS camera_group_id UUID;
    -- Also ensure standard columns exist in case 076 was partial
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS start_time TIMESTAMP WITH TIME ZONE;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_videos_incident ON videos(incident_id) WHERE incident_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_videos_camera_group ON videos(camera_group_id) WHERE camera_group_id IS NOT NULL;

-- 3. BRIN index for time-range queries (far more efficient than B-tree for append-only time data)
CREATE INDEX IF NOT EXISTS idx_videos_start_time_brin ON videos USING BRIN(start_time);

-- 4. Retry policy fields on processing jobs
CREATE TABLE IF NOT EXISTS video_processing_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE video_processing_jobs ADD COLUMN IF NOT EXISTS status VARCHAR(32) NOT NULL DEFAULT 'queued';
    ALTER TABLE video_processing_jobs ADD COLUMN IF NOT EXISTS retry_count INT NOT NULL DEFAULT 0;
    ALTER TABLE video_processing_jobs ADD COLUMN IF NOT EXISTS max_retries INT NOT NULL DEFAULT 3;
    ALTER TABLE video_processing_jobs ADD COLUMN IF NOT EXISTS retry_after TIMESTAMP WITH TIME ZONE;
    ALTER TABLE video_processing_jobs ADD COLUMN IF NOT EXISTS last_error TEXT;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- Index for retry scheduling
CREATE INDEX IF NOT EXISTS idx_video_jobs_retry ON video_processing_jobs(status, retry_after)
    WHERE status = 'queued';

-- 5. Dead Letter Queue for permanently failed video processing jobs
CREATE TABLE IF NOT EXISTS video_processing_dlq (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE video_processing_dlq ADD COLUMN IF NOT EXISTS job_id UUID NOT NULL;
    ALTER TABLE video_processing_dlq ADD COLUMN IF NOT EXISTS video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE;
    ALTER TABLE video_processing_dlq ADD COLUMN IF NOT EXISTS attempts INT NOT NULL DEFAULT 0;
    ALTER TABLE video_processing_dlq ADD COLUMN IF NOT EXISTS last_error TEXT;
    ALTER TABLE video_processing_dlq ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_video_dlq_video ON video_processing_dlq(video_id);

-- 6. Multipart upload sessions (S3-style chunked upload)
CREATE TABLE IF NOT EXISTS video_multipart_sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE video_multipart_sessions ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL;
    ALTER TABLE video_multipart_sessions ADD COLUMN IF NOT EXISTS vehicle_id UUID NOT NULL;
    ALTER TABLE video_multipart_sessions ADD COLUMN IF NOT EXISTS storage_key TEXT NOT NULL DEFAULT '';
    ALTER TABLE video_multipart_sessions ADD COLUMN IF NOT EXISTS upload_id TEXT;
    ALTER TABLE video_multipart_sessions ADD COLUMN IF NOT EXISTS total_parts INT;
    ALTER TABLE video_multipart_sessions ADD COLUMN IF NOT EXISTS parts_received INT NOT NULL DEFAULT 0;
    ALTER TABLE video_multipart_sessions ADD COLUMN IF NOT EXISTS status VARCHAR(32) NOT NULL DEFAULT 'active';
    ALTER TABLE video_multipart_sessions ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}';
    ALTER TABLE video_multipart_sessions ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW() + INTERVAL '24 hours';
    ALTER TABLE video_multipart_sessions ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
    ALTER TABLE video_multipart_sessions ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_name = 'video_multipart_sessions' AND constraint_name = 'uq_video_multipart_upload_id') THEN
        ALTER TABLE video_multipart_sessions ADD CONSTRAINT uq_video_multipart_upload_id UNIQUE (upload_id);
    END IF;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_multipart_tenant ON video_multipart_sessions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_multipart_upload_id ON video_multipart_sessions(upload_id);

-- 7. Multipart parts tracking
CREATE TABLE IF NOT EXISTS video_multipart_parts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE video_multipart_parts ADD COLUMN IF NOT EXISTS session_id UUID NOT NULL REFERENCES video_multipart_sessions(id) ON DELETE CASCADE;
    ALTER TABLE video_multipart_parts ADD COLUMN IF NOT EXISTS part_number INT NOT NULL DEFAULT 1;
    ALTER TABLE video_multipart_parts ADD COLUMN IF NOT EXISTS size_bytes BIGINT NOT NULL DEFAULT 0;
    ALTER TABLE video_multipart_parts ADD COLUMN IF NOT EXISTS etag TEXT;
    ALTER TABLE video_multipart_parts ADD COLUMN IF NOT EXISTS received_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_name = 'video_multipart_parts' AND constraint_name = 'uq_video_multipart_parts_session_part') THEN
        ALTER TABLE video_multipart_parts ADD CONSTRAINT uq_video_multipart_parts_session_part UNIQUE (session_id, part_number);
    END IF;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_multipart_parts_session ON video_multipart_parts(session_id);

-- All tables and indexes handled defensively above.
