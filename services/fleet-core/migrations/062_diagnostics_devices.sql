-- Migration: 062_diagnostics_devices
-- Description: Create diagnostics device tables with tenant isolation.

-- 1. Devices table
CREATE TABLE IF NOT EXISTS devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    vehicle_id UUID,
    device_imei TEXT UNIQUE,
    device_serial TEXT,
    device_model TEXT,
    firmware_version TEXT,
    hardware_version TEXT,
    status TEXT CHECK(status IN ('ACTIVE','INACTIVE','MAINTENANCE','ERROR','DECOMMISSIONED')),
    last_seen_at TIMESTAMP WITH TIME ZONE,
    registered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_devices_tenant_id ON devices(tenant_id);

-- 2. Device telemetry
CREATE TABLE IF NOT EXISTS device_telemetry (
    id BIGSERIAL PRIMARY KEY,
    device_id UUID REFERENCES devices(id),
    timestamp TIMESTAMP WITH TIME ZONE
);

DO $$ 
BEGIN 
    ALTER TABLE device_telemetry ADD COLUMN IF NOT EXISTS gps_lat DOUBLE PRECISION;
    ALTER TABLE device_telemetry ADD COLUMN IF NOT EXISTS gps_lon DOUBLE PRECISION;
    ALTER TABLE device_telemetry ADD COLUMN IF NOT EXISTS speed DOUBLE PRECISION;
    ALTER TABLE device_telemetry ADD COLUMN IF NOT EXISTS battery_voltage DOUBLE PRECISION;
    ALTER TABLE device_telemetry ADD COLUMN IF NOT EXISTS signal_strength INTEGER;
    ALTER TABLE device_telemetry ADD COLUMN IF NOT EXISTS engine_rpm INTEGER;
    ALTER TABLE device_telemetry ADD COLUMN IF NOT EXISTS temperature DOUBLE PRECISION;
    ALTER TABLE device_telemetry ADD COLUMN IF NOT EXISTS payload_json JSONB;
END $$;

CREATE INDEX IF NOT EXISTS idx_device_telemetry_device_ts ON device_telemetry(device_id, timestamp);

-- 3. Device commands
CREATE TABLE IF NOT EXISTS device_commands (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id UUID REFERENCES devices(id),
    command_type TEXT,
    command_payload JSONB,
    status TEXT CHECK(status IN ('CREATED','QUEUED','DELIVERED','ACKED','APPLIED','FAILED')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    delivered_at TIMESTAMP WITH TIME ZONE,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    applied_at TIMESTAMP WITH TIME ZONE
);

-- 4. Device twins
CREATE TABLE IF NOT EXISTS device_twins (
    device_id UUID PRIMARY KEY REFERENCES devices(id),
    desired_state JSONB,
    reported_state JSONB,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 5. Device diagnostics logs
CREATE TABLE IF NOT EXISTS device_diagnostics_logs (
    id BIGSERIAL PRIMARY KEY,
    device_id UUID REFERENCES devices(id),
    log_level TEXT,
    log_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 6. Firmware management
CREATE TABLE IF NOT EXISTS firmware_packages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    firmware_version TEXT,
    device_model TEXT,
    checksum TEXT,
    file_url TEXT,
    size_bytes INTEGER,
    status TEXT CHECK(status IN ('DRAFT','PUBLISHED','DEPRECATED')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS firmware_rollouts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    firmware_id UUID REFERENCES firmware_packages(id),
    tenant_id UUID REFERENCES tenants(id),
    rollout_strategy TEXT,
    target_device_count INTEGER,
    status TEXT CHECK(status IN ('SCHEDULED','IN_PROGRESS','COMPLETED','FAILED','CANCELLED')),
    scheduled_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);
