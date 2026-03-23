-- Migration: 254_emergency_production_lockdown
-- Description: Adds tables for incident snapshots and archival.

-- 1. Create emergency_incident_snapshots table
CREATE TABLE IF NOT EXISTS emergency_incident_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id UUID NOT NULL,
    vehicle_id UUID NOT NULL,
    org_id UUID NOT NULL,
    incident_type TEXT NOT NULL,
    severity TEXT NOT NULL,
    status TEXT NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    snapshot_data JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_incident_snapshots_org ON emergency_incident_snapshots(org_id);
CREATE INDEX IF NOT EXISTS idx_incident_snapshots_incident ON emergency_incident_snapshots(incident_id);

-- 2. Create emergency_incidents_archive table (replica of sos_incidents)
CREATE TABLE IF NOT EXISTS emergency_incidents_archive (
    id UUID PRIMARY KEY,
    vehicle_id UUID NOT NULL,
    driver_id UUID,
    trip_id UUID,
    device_id UUID,
    timestamp TIMESTAMPTZ,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    location_address TEXT,
    status TEXT NOT NULL,
    severity TEXT NOT NULL,
    incident_type TEXT NOT NULL,
    panic_source TEXT NOT NULL,
    speed INTEGER,
    heading INTEGER,
    metadata JSONB,
    acknowledged_by UUID,
    resolved_by UUID,
    org_id UUID NOT NULL,
    high_risk_location BOOLEAN DEFAULT FALSE,
    simulation_mode BOOLEAN DEFAULT FALSE,
    driver_feedback TEXT,
    feedback_notes TEXT,
    acknowledged_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    archived_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_incidents_archive_org ON emergency_incidents_archive(org_id);
CREATE INDEX IF NOT EXISTS idx_incidents_archive_vehicle ON emergency_incidents_archive(vehicle_id);
