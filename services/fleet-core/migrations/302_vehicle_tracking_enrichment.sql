-- Migration 302: Vehicle Tracking Enrichment
-- Adds performance indexes, geofence tables, and helper functions
-- for the enhanced vehicle tracking & dispatcher feature.

-- 1. Performance index on vehicles for bulk org query
CREATE INDEX IF NOT EXISTS idx_vehicles_org_status_active
    ON vehicles (tenant_id, status)
    WHERE deleted_at IS NULL AND status = 'active';

-- 2. Covering index for driver batch lookup
-- Schema alignment: full_name instead of name
CREATE INDEX IF NOT EXISTS idx_drivers_id_lookup
    ON drivers (id) INCLUDE (full_name, phone);

-- 3. Telemetry batch latest optimization
-- Schema alignment: recorded_at instead of time
CREATE INDEX IF NOT EXISTS idx_telemetry_vehicle_time_desc
    ON telemetry_points (vehicle_id, recorded_at DESC);

-- 4. Ensure geofence_assignments has proper index for vehicle-active lookup
CREATE INDEX IF NOT EXISTS idx_geofence_assignments_vehicle_active
    ON geofence_assignments (vehicle_id, is_active);

-- 5. Geofence event query optimization (org-scoped)
-- Guarded: tenant_id might not exist yet (added in 316)
DO $$ 
BEGIN 
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='geofence_events' AND column_name='tenant_id') THEN
        CREATE INDEX IF NOT EXISTS idx_geofence_events_org_time
            ON geofence_events (tenant_id, time DESC)
            WHERE tenant_id IS NOT NULL;
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='geofence_events' AND column_name='org_id') THEN
        CREATE INDEX IF NOT EXISTS idx_geofence_events_org_time
            ON geofence_events (org_id, time DESC)
            WHERE org_id IS NOT NULL;
    END IF;
END $$;
