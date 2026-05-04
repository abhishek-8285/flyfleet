-- Migration: 066b_geofencing_management
-- Description: Creates tables for geofences with tenant isolation and PostGIS support.
-- Note: Relocated from 066_geofencing_management.sql to resolve numeric conflict.

-- 1. Create geofences table
CREATE TABLE IF NOT EXISTS geofences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('circle', 'polygon', 'route_corridor')),
    center_lat DOUBLE PRECISION,
    center_lng DOUBLE PRECISION,
    radius_meters INTEGER CHECK (radius_meters >= 10 AND radius_meters <= 50000),
    coordinates JSONB, -- Fallback for polygon/multipoint
    route_polyline TEXT,
    buffer_meters INTEGER CHECK (buffer_meters >= 10 AND buffer_meters <= 1000),
    entry_alert BOOLEAN DEFAULT FALSE,
    exit_alert BOOLEAN DEFAULT FALSE,
    speed_limit INTEGER,
    vehicles UUID[] DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    enabled BOOLEAN DEFAULT TRUE,
    created_by UUID,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_geofences_tenant ON geofences(tenant_id);

-- 3. Geofence assignments
CREATE TABLE IF NOT EXISTS geofence_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    geofence_id UUID NOT NULL REFERENCES geofences(id) ON DELETE CASCADE,
    vehicle_id UUID NOT NULL,
    assignment_type VARCHAR(20) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(geofence_id, vehicle_id)
);

-- 4. Geofence events
CREATE TABLE IF NOT EXISTS geofence_events (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    time TIMESTAMPTZ NOT NULL,
    vehicle_id UUID NOT NULL,
    geofence_id UUID NOT NULL REFERENCES geofences(id),
    event_type VARCHAR(20) NOT NULL,
    lat DOUBLE PRECISION,
    lng DOUBLE PRECISION,
    duration_minutes INTEGER,
    speed_kmh DECIMAL(5, 2),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 5. TimescaleDB Check
DO $$ 
BEGIN 
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'timescaledb') THEN
        PERFORM create_hypertable('geofence_events', 'time', if_not_exists => TRUE);
    END IF;
EXCEPTION WHEN OTHERS THEN 
    RAISE NOTICE 'TimescaleDB not available, skipping hypertable creation';
END $$;

-- 6. Stats View
CREATE OR REPLACE VIEW geofence_stats AS
SELECT
    g.id,
    g.tenant_id,
    g.name,
    g.type,
    g.enabled,
    COUNT(DISTINCT ga.vehicle_id) FILTER (WHERE ga.is_active = true) as assigned_vehicles
FROM geofences g
LEFT JOIN geofence_assignments ga ON ga.geofence_id = g.id
GROUP BY g.id, g.tenant_id, g.name, g.type, g.enabled;
