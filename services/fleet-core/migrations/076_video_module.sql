-- +goose Up
-- Video Module: dashcam ingestion, evidence management, safety analytics

CREATE TYPE video_type AS ENUM ('continuous', 'event', 'manual', 'live_capture');
CREATE TYPE video_status AS ENUM ('uploading', 'processing', 'ready', 'failed', 'deleted');
CREATE TYPE processing_status AS ENUM ('queued', 'transcoding', 'thumbnailing', 'analyzing', 'done', 'failed');
CREATE TYPE camera_position AS ENUM ('front', 'driver', 'rear', 'cargo', 'side_left', 'side_right', 'panoramic');
CREATE TYPE video_event_type AS ENUM (
    'collision', 'harsh_braking', 'harsh_acceleration', 'overspeed',
    'lane_departure', 'tailgating', 'panic_button', 'impact_sensor',
    'phone_usage', 'driver_distraction', 'seatbelt_violation',
    'smoking', 'drowsiness', 'camera_tamper', 'camera_offline'
);

-- cameras: physical dashcam units registered per vehicle
CREATE TABLE IF NOT EXISTS cameras (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL,
    vehicle_id          UUID NOT NULL,
    camera_position     camera_position NOT NULL,
    serial_number       VARCHAR(128) UNIQUE,
    firmware_version    VARCHAR(64),
    status              VARCHAR(32) NOT NULL DEFAULT 'active',
    installed_at        TIMESTAMP WITH TIME ZONE,
    last_heartbeat      TIMESTAMP WITH TIME ZONE,
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cameras_vehicle ON cameras(vehicle_id);
CREATE INDEX idx_cameras_tenant ON cameras(tenant_id);
CREATE INDEX idx_cameras_heartbeat ON cameras(last_heartbeat);

-- videos: core clip metadata (all binary in object storage)
CREATE TABLE IF NOT EXISTS videos (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id               UUID NOT NULL,
    vehicle_id              UUID NOT NULL,
    camera_id               UUID REFERENCES cameras(id) ON DELETE SET NULL,
    driver_id               UUID,
    trip_id                 UUID,
    video_type              video_type NOT NULL DEFAULT 'continuous',
    status                  video_status NOT NULL DEFAULT 'uploading',
    file_path               TEXT NOT NULL,        -- object storage key
    hls_playlist_path       TEXT,                 -- HLS m3u8 key after transcoding
    codec                   VARCHAR(32),
    resolution              VARCHAR(16),
    file_size               BIGINT,
    duration_seconds        INT,
    start_time              TIMESTAMP WITH TIME ZONE,
    end_time                TIMESTAMP WITH TIME ZONE,
    -- Evidence management
    evidence_locked         BOOLEAN NOT NULL DEFAULT FALSE,
    legal_hold              BOOLEAN NOT NULL DEFAULT FALSE,
    case_id                 UUID,
    retention_override_days INT,                  -- NULL = use default policy
    -- Integrity
    integrity_hash          VARCHAR(64),          -- SHA256 hex
    integrity_signature     TEXT,                 -- device-signed hash
    device_cert_serial      VARCHAR(128),
    -- Privacy
    privacy_blurred         BOOLEAN NOT NULL DEFAULT FALSE,
    created_at              TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_videos_tenant ON videos(tenant_id);
CREATE INDEX idx_videos_vehicle ON videos(vehicle_id);
CREATE INDEX idx_videos_driver ON videos(driver_id);
CREATE INDEX idx_videos_trip ON videos(trip_id);
CREATE INDEX idx_videos_camera ON videos(camera_id);
CREATE INDEX idx_videos_start_time ON videos(start_time DESC);
CREATE INDEX idx_videos_type ON videos(video_type);
CREATE INDEX idx_videos_evidence ON videos(evidence_locked) WHERE evidence_locked = TRUE;

-- video_events: AI-detected or sensor-triggered events linked to a video
CREATE TABLE IF NOT EXISTS video_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    vehicle_id      UUID NOT NULL,
    video_id        UUID REFERENCES videos(id) ON DELETE CASCADE,
    event_type      video_event_type NOT NULL,
    occurred_at     TIMESTAMP WITH TIME ZONE NOT NULL,
    severity        INT NOT NULL DEFAULT 1 CHECK (severity BETWEEN 1 AND 5),
    confidence_score FLOAT CHECK (confidence_score BETWEEN 0 AND 1),
    offset_seconds  INT,         -- seconds into the video clip
    metadata        JSONB,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_video_events_video ON video_events(video_id);
CREATE INDEX idx_video_events_vehicle ON video_events(vehicle_id);
CREATE INDEX idx_video_events_type ON video_events(event_type);
CREATE INDEX idx_video_events_occurred ON video_events(occurred_at DESC);

-- video_thumbnails: representative frame URLs for UI previews
CREATE TABLE IF NOT EXISTS video_thumbnails (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id        UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    offset_seconds  INT NOT NULL,
    thumbnail_path  TEXT NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_video_thumbnails_video ON video_thumbnails(video_id);

-- video_processing_jobs: async processing queue
CREATE TABLE IF NOT EXISTS video_processing_jobs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id        UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    status          processing_status NOT NULL DEFAULT 'queued',
    worker_id       VARCHAR(64),
    attempts        INT NOT NULL DEFAULT 0,
    last_heartbeat  TIMESTAMP WITH TIME ZONE,
    error_message   TEXT,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_video_jobs_status ON video_processing_jobs(status);
CREATE INDEX idx_video_jobs_heartbeat ON video_processing_jobs(last_heartbeat);

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

-- +goose Down
DROP TRIGGER IF EXISTS trg_video_job_notify ON video_processing_jobs;
DROP FUNCTION IF EXISTS video_job_notify();
DROP TABLE IF EXISTS video_processing_jobs;
DROP TABLE IF EXISTS video_thumbnails;
DROP TABLE IF EXISTS video_events;
DROP TABLE IF EXISTS videos;
DROP TABLE IF EXISTS cameras;
DROP TYPE IF EXISTS video_event_type;
DROP TYPE IF EXISTS camera_position;
DROP TYPE IF EXISTS processing_status;
DROP TYPE IF EXISTS video_status;
DROP TYPE IF EXISTS video_type;
