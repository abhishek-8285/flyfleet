-- Migration: Device Management and Telemetry Tables
-- Created: 2026-02-15
-- Purpose: Add tables for device registration, telemetry, media uploads, attestations, and driver HOS

-- Devices table for device registration
CREATE TABLE IF NOT EXISTS devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id VARCHAR(255) UNIQUE NOT NULL,
    organization_id UUID REFERENCES organizations(id),
    vehicle_id UUID REFERENCES vehicles(id),
    hardware_attestation TEXT NOT NULL,
    android_version VARCHAR(50) NOT NULL,
    device_model VARCHAR(100) NOT NULL,
    app_version VARCHAR(50),
    status VARCHAR(20) DEFAULT 'active',
    jwt_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_devices_device_id ON devices(device_id);
CREATE INDEX idx_devices_organization_id ON devices(organization_id);
CREATE INDEX idx_devices_status ON devices(status);

-- Telemetry points table
CREATE TABLE IF NOT EXISTS telemetry_points (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id VARCHAR(255) NOT NULL REFERENCES devices(device_id),
    batch_id VARCHAR(255) NOT NULL,
    timestamp BIGINT NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    speed_kph DOUBLE PRECISION,
    heading DOUBLE PRECISION,
    altitude DOUBLE PRECISION,
    accuracy DOUBLE PRECISION,
    satellites INTEGER,
    can_bus_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_telemetry_device_id ON telemetry_points(device_id);
CREATE INDEX idx_telemetry_batch_id ON telemetry_points(batch_id);
CREATE INDEX idx_telemetry_timestamp ON telemetry_points(timestamp);
CREATE INDEX idx_telemetry_device_timestamp ON telemetry_points(device_id, timestamp DESC);

-- Media uploads table
CREATE TABLE IF NOT EXISTS media_uploads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id VARCHAR(255) NOT NULL REFERENCES devices(device_id),
    clip_id VARCHAR(255) UNIQUE NOT NULL,
    object_key TEXT NOT NULL,
    checksum_sha256 VARCHAR(64) NOT NULL,
    duration_seconds INTEGER,
    uploaded_bytes BIGINT,
    status VARCHAR(20) DEFAULT 'pending',
    storage_mode VARCHAR(20) DEFAULT 'local', -- 'local' or 'cloud'
    sync_status VARCHAR(20) DEFAULT 'pending_sync', -- 'pending_sync', 'syncing', 'synced', 'failed'
    local_path TEXT, -- Path on device for local storage
    cloud_url TEXT, -- Cloud URL if synced
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_media_uploads_device_id ON media_uploads(device_id);
CREATE INDEX idx_media_uploads_clip_id ON media_uploads(clip_id);
CREATE INDEX idx_media_uploads_status ON media_uploads(status);
CREATE INDEX idx_media_uploads_sync_status ON media_uploads(sync_status);
CREATE INDEX idx_media_uploads_storage_mode ON media_uploads(storage_mode);

-- Evidence attestations table
CREATE TABLE IF NOT EXISTS evidence_attestations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attestation_id VARCHAR(255) UNIQUE NOT NULL,
    device_id VARCHAR(255) NOT NULL REFERENCES devices(device_id),
    evidence_hash VARCHAR(64) NOT NULL,
    chain_hash VARCHAR(64) NOT NULL,
    signature TEXT NOT NULL,
    timestamp BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_evidence_attestations_device_id ON evidence_attestations(device_id);
CREATE INDEX idx_evidence_attestations_attestation_id ON evidence_attestations(attestation_id);

-- Device commands table
CREATE TABLE IF NOT EXISTS device_commands (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id VARCHAR(255) NOT NULL REFERENCES devices(device_id),
    type VARCHAR(50) NOT NULL,
    params JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'pending',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    priority INTEGER DEFAULT 3,
    created_at BIGINT NOT NULL,
    executed_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_device_commands_device_id ON device_commands(device_id);
CREATE INDEX idx_device_commands_status ON device_commands(status);
CREATE INDEX idx_device_commands_priority ON device_commands(priority, created_at);

-- Driver sessions table
CREATE TABLE IF NOT EXISTS driver_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id VARCHAR(255) NOT NULL REFERENCES devices(device_id),
    driver_id VARCHAR(255) NOT NULL,
    rfid_token VARCHAR(255),
    duty_status VARCHAR(20) NOT NULL,
    shift_start TIMESTAMP WITH TIME ZONE NOT NULL,
    shift_end TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(device_id, driver_id, status)
);

CREATE INDEX idx_driver_sessions_device_id ON driver_sessions(device_id);
CREATE INDEX idx_driver_sessions_driver_id ON driver_sessions(driver_id);
CREATE INDEX idx_driver_sessions_status ON driver_sessions(status);

-- Driver HOS logs table
CREATE TABLE IF NOT EXISTS driver_hos_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id VARCHAR(255) NOT NULL REFERENCES devices(device_id),
    driver_id VARCHAR(255) NOT NULL,
    duty_status VARCHAR(20) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    location GEOGRAPHY(POINT, 4326),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_driver_hos_logs_driver_id ON driver_hos_logs(driver_id);
CREATE INDEX idx_driver_hos_logs_timestamp ON driver_hos_logs(timestamp DESC);
CREATE INDEX idx_driver_hos_logs_driver_timestamp ON driver_hos_logs(driver_id, timestamp DESC);

-- Device configs table
CREATE TABLE IF NOT EXISTS device_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id VARCHAR(255) NOT NULL REFERENCES devices(device_id),
    config_version VARCHAR(20) NOT NULL DEFAULT 'v1',
    config_data JSONB NOT NULL,
    signature TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_device_configs_device_id ON device_configs(device_id);
CREATE INDEX idx_device_configs_version ON device_configs(config_version);

-- Insert default device config
INSERT INTO device_configs (device_id, config_version, config_data, signature)
VALUES (
    '*',
    'v1',
    '{
        "schema_version": "v1",
        "heartbeat_seconds": 45,
        "ring_buffer_minutes": 10,
        "gps": {
            "active_interval_seconds": 1,
            "idle_interval_seconds": 30
        },
        "ai": {
            "enabled": true,
            "frame_sample_stride": 6,
            "severity_thresholds": {
                "warn_min": 0.6,
                "critical_min": 0.85
            }
        },
        "retention": {
            "critical_clip_days": 14,
            "min_storage_reserve_mb": 512,
            "legal_hold": false
        },
        "feature_flags": {
            "camera_enabled": true,
            "audio_event_checks_enabled": false,
            "ai_enabled": true,
            "can_bus_enabled": true,
            "driver_lockout_enabled": true
        }
    }'::jsonb,
    'default-signature-placeholder'
)
ON CONFLICT DO NOTHING;
