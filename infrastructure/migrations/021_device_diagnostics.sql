-- Device Diagnostics Module Migration
-- Created: 2026-02-09
-- Description: Adds tables for device diagnostics, commands, firmware management

-- 1. Enhance devices table with additional health fields
ALTER TABLE devices ADD COLUMN IF NOT EXISTS firmware_version VARCHAR(50);
ALTER TABLE devices ADD COLUMN IF NOT EXISTS battery_level INTEGER;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS signal_strength INTEGER;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS hardware_revision VARCHAR(20);
ALTER TABLE devices ADD COLUMN IF NOT EXISTS last_heartbeat TIMESTAMP WITH TIME ZONE;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS model VARCHAR(50);
ALTER TABLE devices ADD COLUMN IF NOT EXISTS manufacturer VARCHAR(50);

-- 2. Device telemetry (raw packets from devices)
CREATE TABLE IF NOT EXISTS device_telemetry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID REFERENCES devices(id) ON DELETE CASCADE,
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    payload JSONB NOT NULL,
    packet_type VARCHAR(50), -- gps, obd, heartbeat, sensor
    raw_hex TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Device commands (remote control operations)
CREATE TABLE IF NOT EXISTS device_commands (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID REFERENCES devices(id) ON DELETE CASCADE,
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    command VARCHAR(100) NOT NULL, -- reboot, reset_modem, clear_cache, update_config
    category VARCHAR(50), -- power, network, config, diagnostic
    params JSONB,
    status VARCHAR(20) DEFAULT 'pending', -- pending, sent, executing, completed, failed, cancelled
    result JSONB,
    error_message TEXT,
    created_by UUID REFERENCES users(id),
    sent_at TIMESTAMP WITH TIME ZONE,
    executed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 4. Firmware packages repository
CREATE TABLE IF NOT EXISTS firmware_packages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version VARCHAR(50) NOT NULL,
    hardware_target VARCHAR(50) NOT NULL, -- HW-REV-04, HW-REV-03, etc.
    file_url TEXT NOT NULL,
    file_size INTEGER,
    checksum VARCHAR(64),
    release_notes TEXT,
    status VARCHAR(20) DEFAULT 'draft', -- draft, testing, active, deprecated
    is_critical BOOLEAN DEFAULT FALSE,
    released_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 5. Firmware rollouts (deployment campaigns)
CREATE TABLE IF NOT EXISTS firmware_rollouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    firmware_id UUID REFERENCES firmware_packages(id) ON DELETE CASCADE,
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(100),
    device_group VARCHAR(100), -- fleet, region, device_type filter
    target_filter JSONB, -- complex filters
    total_devices INTEGER DEFAULT 0,
    completed_devices INTEGER DEFAULT 0,
    failed_devices INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'scheduled', -- scheduled, in_progress, paused, completed, cancelled
    rollout_strategy VARCHAR(20) DEFAULT 'gradual', -- gradual, immediate, scheduled
    max_concurrent INTEGER DEFAULT 10,
    created_by UUID REFERENCES users(id),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 6. Device firmware status (per-device rollout tracking)
CREATE TABLE IF NOT EXISTS device_firmware_status (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID REFERENCES devices(id) ON DELETE CASCADE,
    rollout_id UUID REFERENCES firmware_rollouts(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'pending', -- pending, downloading, installing, completed, failed
    progress INTEGER DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(device_id, rollout_id)
);

-- 7. Device diagnostics logs (system health events)
CREATE TABLE IF NOT EXISTS device_diagnostics_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID REFERENCES devices(id) ON DELETE CASCADE,
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    log_type VARCHAR(50), -- connection, error, warning, info, debug
    source VARCHAR(50), -- modem, gps, sensor, system
    message TEXT,
    metadata JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_device_telemetry_device ON device_telemetry(device_id);
CREATE INDEX IF NOT EXISTS idx_device_telemetry_timestamp ON device_telemetry(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_device_commands_device ON device_commands(device_id);
CREATE INDEX IF NOT EXISTS idx_device_commands_status ON device_commands(status);
CREATE INDEX IF NOT EXISTS idx_firmware_rollouts_org ON firmware_rollouts(org_id);
CREATE INDEX IF NOT EXISTS idx_firmware_rollouts_status ON firmware_rollouts(status);
CREATE INDEX IF NOT EXISTS idx_device_firmware_status_rollout ON device_firmware_status(rollout_id);
CREATE INDEX IF NOT EXISTS idx_device_diagnostics_logs_device ON device_diagnostics_logs(device_id);
CREATE INDEX IF NOT EXISTS idx_device_diagnostics_logs_timestamp ON device_diagnostics_logs(timestamp DESC);
