-- +goose Up
-- Video Module: dashcam ingestion, evidence management, safety analytics

DO $$ BEGIN
    CREATE TYPE video_type AS ENUM ('continuous', 'event', 'manual', 'live_capture');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE video_status AS ENUM ('uploading', 'processing', 'ready', 'failed', 'deleted');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE processing_status AS ENUM ('queued', 'transcoding', 'thumbnailing', 'analyzing', 'done', 'failed');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE camera_position AS ENUM ('front', 'driver', 'rear', 'cargo', 'side_left', 'side_right', 'panoramic');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE video_event_type AS ENUM (
        'collision', 'harsh_braking', 'harsh_acceleration', 'overspeed',
        'lane_departure', 'tailgating', 'panic_button', 'impact_sensor',
        'phone_usage', 'driver_distraction', 'seatbelt_violation',
        'smoking', 'drowsiness', 'camera_tamper', 'camera_offline'
    );
EXCEPTION WHEN duplicate_object THEN null;
END $$;

-- cameras: physical dashcam units registered per vehicle
CREATE TABLE IF NOT EXISTS cameras (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE cameras ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL;
    ALTER TABLE cameras ADD COLUMN IF NOT EXISTS vehicle_id UUID NOT NULL;
    ALTER TABLE cameras ADD COLUMN IF NOT EXISTS camera_position camera_position NOT NULL DEFAULT 'front';
    ALTER TABLE cameras ADD COLUMN IF NOT EXISTS serial_number VARCHAR(128);
    ALTER TABLE cameras ADD COLUMN IF NOT EXISTS firmware_version VARCHAR(64);
    ALTER TABLE cameras ADD COLUMN IF NOT EXISTS status VARCHAR(32) NOT NULL DEFAULT 'active';
    ALTER TABLE cameras ADD COLUMN IF NOT EXISTS installed_at TIMESTAMP WITH TIME ZONE;
    ALTER TABLE cameras ADD COLUMN IF NOT EXISTS last_heartbeat TIMESTAMP WITH TIME ZONE;
    ALTER TABLE cameras ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
    ALTER TABLE cameras ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
    
    -- Handle unique constraint separately
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_name = 'cameras' AND constraint_name = 'cameras_serial_number_key') THEN
        ALTER TABLE cameras ADD CONSTRAINT cameras_serial_number_key UNIQUE (serial_number);
    END IF;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_cameras_vehicle ON cameras(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_cameras_tenant ON cameras(tenant_id);
CREATE INDEX IF NOT EXISTS idx_cameras_heartbeat ON cameras(last_heartbeat);

-- videos: core clip metadata (all binary in object storage)
CREATE TABLE IF NOT EXISTS videos (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL;
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS vehicle_id UUID NOT NULL;
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS camera_id UUID REFERENCES cameras(id) ON DELETE SET NULL;
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS driver_id UUID;
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS trip_id UUID;
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS video_type video_type NOT NULL DEFAULT 'continuous';
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS status video_status NOT NULL DEFAULT 'uploading';
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS file_path TEXT NOT NULL DEFAULT '';
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS hls_playlist_path TEXT;
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS codec VARCHAR(32);
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS resolution VARCHAR(16);
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS file_size BIGINT;
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS duration_seconds INT;
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS start_time TIMESTAMP WITH TIME ZONE;
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS end_time TIMESTAMP WITH TIME ZONE;
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS evidence_locked BOOLEAN NOT NULL DEFAULT FALSE;
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS legal_hold BOOLEAN NOT NULL DEFAULT FALSE;
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS case_id UUID;
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS retention_override_days INT;
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS integrity_hash VARCHAR(64);
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS integrity_signature TEXT;
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS device_cert_serial VARCHAR(128);
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS privacy_blurred BOOLEAN NOT NULL DEFAULT FALSE;
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
    ALTER TABLE videos ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_videos_tenant ON videos(tenant_id);
CREATE INDEX IF NOT EXISTS idx_videos_vehicle ON videos(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_videos_driver ON videos(driver_id);
CREATE INDEX IF NOT EXISTS idx_videos_trip ON videos(trip_id);
CREATE INDEX IF NOT EXISTS idx_videos_camera ON videos(camera_id);
CREATE INDEX IF NOT EXISTS idx_videos_start_time ON videos(start_time DESC);
CREATE INDEX IF NOT EXISTS idx_videos_type ON videos(video_type);
CREATE INDEX IF NOT EXISTS idx_videos_evidence ON videos(evidence_locked) WHERE evidence_locked = TRUE;

-- video_events: AI-detected or sensor-triggered events linked to a video
CREATE TABLE IF NOT EXISTS video_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE video_events ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL;
    ALTER TABLE video_events ADD COLUMN IF NOT EXISTS vehicle_id UUID NOT NULL;
    ALTER TABLE video_events ADD COLUMN IF NOT EXISTS video_id UUID REFERENCES videos(id) ON DELETE CASCADE;
    ALTER TABLE video_events ADD COLUMN IF NOT EXISTS event_type video_event_type NOT NULL DEFAULT 'collision';
    ALTER TABLE video_events ADD COLUMN IF NOT EXISTS occurred_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
    ALTER TABLE video_events ADD COLUMN IF NOT EXISTS severity INT NOT NULL DEFAULT 1;
    ALTER TABLE video_events ADD COLUMN IF NOT EXISTS confidence_score FLOAT DEFAULT 0;
    ALTER TABLE video_events ADD COLUMN IF NOT EXISTS offset_seconds INT;
    ALTER TABLE video_events ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}';
    ALTER TABLE video_events ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_video_events_video ON video_events(video_id);
CREATE INDEX IF NOT EXISTS idx_video_events_vehicle ON video_events(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_video_events_type ON video_events(event_type);
CREATE INDEX IF NOT EXISTS idx_video_events_occurred ON video_events(occurred_at DESC);

-- video_thumbnails: representative frame URLs for UI previews
CREATE TABLE IF NOT EXISTS video_thumbnails (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE video_thumbnails ADD COLUMN IF NOT EXISTS video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE;
    ALTER TABLE video_thumbnails ADD COLUMN IF NOT EXISTS offset_seconds INT NOT NULL DEFAULT 0;
    ALTER TABLE video_thumbnails ADD COLUMN IF NOT EXISTS thumbnail_path TEXT NOT NULL DEFAULT '';
    ALTER TABLE video_thumbnails ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_video_thumbnails_video ON video_thumbnails(video_id);

-- video_processing_jobs: async processing queue
CREATE TABLE IF NOT EXISTS video_processing_jobs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE video_processing_jobs ADD COLUMN IF NOT EXISTS video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE;
    ALTER TABLE video_processing_jobs ADD COLUMN IF NOT EXISTS status processing_status NOT NULL DEFAULT 'queued';
    ALTER TABLE video_processing_jobs ADD COLUMN IF NOT EXISTS worker_id VARCHAR(64);
    ALTER TABLE video_processing_jobs ADD COLUMN IF NOT EXISTS attempts INT NOT NULL DEFAULT 0;
    ALTER TABLE video_processing_jobs ADD COLUMN IF NOT EXISTS last_heartbeat TIMESTAMP WITH TIME ZONE;
    ALTER TABLE video_processing_jobs ADD COLUMN IF NOT EXISTS error_message TEXT;
    ALTER TABLE video_processing_jobs ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
    ALTER TABLE video_processing_jobs ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_video_jobs_status ON video_processing_jobs(status);
CREATE INDEX IF NOT EXISTS idx_video_jobs_heartbeat ON video_processing_jobs(last_heartbeat);

-- NOTIFY trigger for processing workers
CREATE OR REPLACE FUNCTION video_job_notify() RETURNS trigger AS $$
BEGIN
    PERFORM pg_notify('video_processing_jobs', 'new');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_video_job_notify ON video_processing_jobs;
CREATE TRIGGER trg_video_job_notify
    AFTER INSERT ON video_processing_jobs
    FOR EACH ROW EXECUTE FUNCTION video_job_notify();

-- All tables and indexes handled defensively above.
