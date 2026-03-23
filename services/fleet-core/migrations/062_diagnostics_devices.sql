-- Migration: 062_diagnostics_devices.sql
-- Description: Create diagnostics device tables

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- devices
CREATE TABLE IF NOT EXISTS devices (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    vehicle_id UUID,
    device_imei TEXT UNIQUE,
    device_serial TEXT,
    device_model TEXT,
    firmware_version TEXT,
    hardware_version TEXT,
    status TEXT CHECK(status IN ('ACTIVE','INACTIVE','MAINTENANCE','ERROR','DECOMMISSIONED')),
    last_seen_at TIMESTAMP,
    registered_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_devices_imei ON devices(device_imei);
CREATE INDEX IF NOT EXISTS idx_devices_org_id ON devices(org_id);
CREATE INDEX IF NOT EXISTS idx_devices_vehicle_id ON devices(vehicle_id);

-- device_telemetry
CREATE TABLE IF NOT EXISTS device_telemetry (
    id BIGSERIAL PRIMARY KEY,
    device_id UUID REFERENCES devices(id),
    timestamp TIMESTAMP,
    gps_lat DOUBLE PRECISION,
    gps_lon DOUBLE PRECISION,
    speed DOUBLE PRECISION,
    battery_voltage DOUBLE PRECISION,
    signal_strength INTEGER,
    engine_rpm INTEGER,
    temperature DOUBLE PRECISION,
    payload_json JSONB
);

CREATE INDEX IF NOT EXISTS idx_device_telemetry_device_timestamp ON device_telemetry(device_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_device_telemetry_timestamp ON device_telemetry(timestamp);

-- device_commands
CREATE TABLE IF NOT EXISTS device_commands (
    id UUID PRIMARY KEY,
    device_id UUID REFERENCES devices(id),
    command_type TEXT,
    command_payload JSONB,
    status TEXT CHECK(status IN ('CREATED','QUEUED','DELIVERED','ACKED','APPLIED','FAILED')),
    created_at TIMESTAMP,
    delivered_at TIMESTAMP,
    acknowledged_at TIMESTAMP,
    applied_at TIMESTAMP
);

-- device_twins
CREATE TABLE IF NOT EXISTS device_twins (
    device_id UUID PRIMARY KEY REFERENCES devices(id),
    desired_state JSONB,
    reported_state JSONB,
    updated_at TIMESTAMP
);

-- device_diagnostics_logs
CREATE TABLE IF NOT EXISTS device_diagnostics_logs (
    id BIGSERIAL PRIMARY KEY,
    device_id UUID REFERENCES devices(id),
    log_level TEXT,
    log_message TEXT,
    created_at TIMESTAMP
);

-- firmware_packages
CREATE TABLE IF NOT EXISTS firmware_packages (
    id UUID PRIMARY KEY,
    firmware_version TEXT,
    device_model TEXT,
    checksum TEXT,
    file_url TEXT,
    size_bytes INTEGER,
    status TEXT CHECK(status IN ('DRAFT','PUBLISHED','DEPRECATED')),
    created_at TIMESTAMP
);

-- firmware_rollouts
CREATE TABLE IF NOT EXISTS firmware_rollouts (
    id UUID PRIMARY KEY,
    firmware_id UUID REFERENCES firmware_packages(id),
    org_id UUID,
    rollout_strategy TEXT,
    target_device_count INTEGER,
    status TEXT CHECK(status IN ('SCHEDULED','IN_PROGRESS','COMPLETED','FAILED','CANCELLED')),
    scheduled_at TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

-- firmware_rollout_events
CREATE TABLE IF NOT EXISTS firmware_rollout_events (
    id BIGSERIAL PRIMARY KEY,
    rollout_id UUID,
    device_id UUID,
    event_type TEXT,
    metadata JSONB,
    created_at TIMESTAMP
);
