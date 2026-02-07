-- Migration to add video service tables
-- This adds video streaming, recording, and analytics functionality to the fleet platform

-- Table for storing video streams
CREATE TABLE IF NOT EXISTS video_streams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    camera_id VARCHAR(255) NOT NULL,
    vehicle_id UUID NOT NULL,
    driver_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    url TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, inactive, error
    resolution VARCHAR(20), -- e.g., "1080p", "720p"
    framerate INTEGER, -- frames per second
    bitrate INTEGER, -- kbps
    protocol VARCHAR(20), -- RTMP, HLS, WebRTC
    location JSONB NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ended_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    analytics_enabled BOOLEAN DEFAULT FALSE,
    CONSTRAINT fk_video_streams_vehicle FOREIGN KEY (vehicle_id) REFERENCES vehicles(id),
    CONSTRAINT fk_video_streams_driver FOREIGN KEY (driver_id) REFERENCES drivers(id),
    CONSTRAINT fk_video_streams_organization FOREIGN KEY (organization_id) REFERENCES organizations(id)
);

-- Table for storing video recordings
CREATE TABLE IF NOT EXISTS video_recordings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stream_id UUID NOT NULL,
    camera_id VARCHAR(255) NOT NULL,
    vehicle_id UUID NOT NULL,
    driver_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    filename VARCHAR(500) NOT NULL,
    filepath TEXT NOT NULL,
    file_size BIGINT NOT NULL, -- in bytes
    duration INTEGER NOT NULL, -- in seconds
    resolution VARCHAR(20), -- e.g., "1080p", "720p"
    framerate INTEGER, -- frames per second
    bitrate INTEGER, -- kbps
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    location_start JSONB NOT NULL,
    location_end JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'processing', -- recorded, processing, available, archived, deleted
    thumbnail_path TEXT,
    tags JSONB DEFAULT '[]',
    event_type VARCHAR(100), -- sudden_stop, harsh_acceleration, accident, etc.
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fk_video_recordings_stream FOREIGN KEY (stream_id) REFERENCES video_streams(id),
    CONSTRAINT fk_video_recordings_vehicle FOREIGN KEY (vehicle_id) REFERENCES vehicles(id),
    CONSTRAINT fk_video_recordings_driver FOREIGN KEY (driver_id) REFERENCES drivers(id),
    CONSTRAINT fk_video_recordings_organization FOREIGN KEY (organization_id) REFERENCES organizations(id)
);

-- Table for storing video analytics (AI-powered insights)
CREATE TABLE IF NOT EXISTS video_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recording_id UUID NOT NULL,
    stream_id UUID,
    objects_detected JSONB NOT NULL DEFAULT '[]',
    behavior_events JSONB NOT NULL DEFAULT '[]',
    traffic_signs JSONB NOT NULL DEFAULT '[]',
    incidents JSONB NOT NULL DEFAULT '[]',
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    confidence DECIMAL(3,2) NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fk_video_analytics_recording FOREIGN KEY (recording_id) REFERENCES video_recordings(id),
    CONSTRAINT fk_video_analytics_stream FOREIGN KEY (stream_id) REFERENCES video_streams(id)
);

-- Table for storing video retention policies
CREATE TABLE IF NOT EXISTS video_retention_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    event_type VARCHAR(100), -- Apply only to specific events (optional)
    retention_period INTEGER NOT NULL, -- in days
    auto_archive BOOLEAN DEFAULT FALSE,
    auto_delete BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fk_video_retention_policies_org FOREIGN KEY (organization_id) REFERENCES organizations(id)
);

-- Table for storing video thumbnails
CREATE TABLE IF NOT EXISTS video_thumbnails (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recording_id UUID NOT NULL,
    image_path TEXT NOT NULL,
    image_url TEXT NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL, -- Time offset in the video
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fk_video_thumbnails_recording FOREIGN KEY (recording_id) REFERENCES video_recordings(id)
);

-- Individual indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_video_streams_vehicle_id ON video_streams (vehicle_id);
CREATE INDEX IF NOT EXISTS idx_video_streams_status ON video_streams (status);
CREATE INDEX IF NOT EXISTS idx_video_streams_started_at ON video_streams (started_at);
CREATE INDEX IF NOT EXISTS idx_video_recordings_stream_id ON video_recordings (stream_id);
CREATE INDEX IF NOT EXISTS idx_video_recordings_vehicle_id ON video_recordings (vehicle_id);
CREATE INDEX IF NOT EXISTS idx_video_recordings_start_time ON video_recordings (start_time);
CREATE INDEX IF NOT EXISTS idx_video_recordings_status ON video_recordings (status);
CREATE INDEX IF NOT EXISTS idx_video_analytics_recording_id ON video_analytics (recording_id);
CREATE INDEX IF NOT EXISTS idx_video_analytics_timestamp ON video_analytics (timestamp);
CREATE INDEX IF NOT EXISTS idx_video_retention_policies_org_id ON video_retention_policies (organization_id);
CREATE INDEX IF NOT EXISTS idx_video_thumbnails_recording_id ON video_thumbnails (recording_id);

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_video_streams_vehicle_status ON video_streams (vehicle_id, status);
CREATE INDEX IF NOT EXISTS idx_video_recordings_vehicle_time ON video_recordings (vehicle_id, start_time);
CREATE INDEX IF NOT EXISTS idx_video_recordings_stream_status ON video_recordings (stream_id, status);

-- Add comments to describe the tables
COMMENT ON TABLE video_streams IS 'Stores information about live video streams from vehicles';
COMMENT ON COLUMN video_streams.status IS 'Status of the video stream (active, inactive, error)';
COMMENT ON COLUMN video_streams.analytics_enabled IS 'Flag to enable AI-powered analytics on this stream';

COMMENT ON TABLE video_recordings IS 'Stores information about recorded video clips';
COMMENT ON COLUMN video_recordings.status IS 'Status of the video recording (recorded, processing, available, archived, deleted)';
COMMENT ON COLUMN video_recordings.event_type IS 'Type of event that triggered the recording (sudden_stop, harsh_acceleration, etc.)';

COMMENT ON TABLE video_analytics IS 'Stores AI-powered insights extracted from video analysis';
COMMENT ON COLUMN video_analytics.confidence IS 'Confidence score for the analytics results (0-1)';
COMMENT ON COLUMN video_analytics.objects_detected IS 'List of objects detected in the video';
COMMENT ON COLUMN video_analytics.behavior_events IS 'List of behavioral events detected in the video';
COMMENT ON COLUMN video_analytics.incidents IS 'List of incidents detected in the video';

COMMENT ON TABLE video_retention_policies IS 'Stores retention policies for video data';
COMMENT ON COLUMN video_retention_policies.retention_period IS 'Number of days to retain videos before archiving/deletion';
COMMENT ON COLUMN video_retention_policies.auto_archive IS 'Flag to automatically archive videos after retention period';
COMMENT ON COLUMN video_retention_policies.auto_delete IS 'Flag to automatically delete videos after retention period';

COMMENT ON TABLE video_thumbnails IS 'Stores thumbnail images for video recordings';