-- Migration 301: Vehicle Tracking Performance Indexes
-- Purpose: Optimize bulk vehicle location queries, geofence checks, and history trail lookups.
-- Execution: Fixed for transaction compatibility and schema alignment.

-- 1. Bulk vehicle locations: speed up latest-position-per-vehicle query.
-- Schema alignment: recorded_at instead of time for telemetry_points
CREATE INDEX IF NOT EXISTS idx_telemetry_vehicle_time_desc
    ON telemetry_points (vehicle_id, recorded_at DESC);

-- 2. Geofence spatial queries (GIN index for JSONB coordinates as PostGIS is unavailable).
CREATE INDEX IF NOT EXISTS idx_geofences_coordinates_gin
    ON geofences USING GIN (coordinates);

-- 3. Geofence lookup by org + status.
-- Checking if it's org_id or tenant_id
DO $$ 
BEGIN 
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='geofences' AND column_name='org_id') THEN
        CREATE INDEX IF NOT EXISTS idx_geofences_org_status ON geofences (org_id, enabled);
    ELSE
        CREATE INDEX IF NOT EXISTS idx_geofences_tenant_status ON geofences (tenant_id, enabled);
    END IF;
END $$;

-- 4. Geofence event queries by vehicle (recent events for a vehicle).
-- Schema alignment: time instead of occurred_at
CREATE INDEX IF NOT EXISTS idx_geofence_events_vehicle_time
    ON geofence_events (vehicle_id, time DESC);

-- 5. Geofence event queries by geofence (recent events for a geofence).
CREATE INDEX IF NOT EXISTS idx_geofence_events_geofence_time
    ON geofence_events (geofence_id, time DESC);

-- 6. Vehicle lookup (standardizing on available identity column).
DO $$ 
BEGIN 
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vehicles' AND column_name='org_id') THEN
        CREATE INDEX IF NOT EXISTS idx_vehicles_org_status_deleted
            ON vehicles (org_id, status)
            WHERE deleted_at IS NULL;
    ELSE
        CREATE INDEX IF NOT EXISTS idx_vehicles_tenant_status_deleted
            ON vehicles (tenant_id, status)
            WHERE deleted_at IS NULL;
    END IF;
END $$;

-- 7. Add alert tracking columns to geofence_events (idempotent).
ALTER TABLE geofence_events
    ADD COLUMN IF NOT EXISTS alert_sent BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS alert_channels JSONB DEFAULT '["in_app"]';
