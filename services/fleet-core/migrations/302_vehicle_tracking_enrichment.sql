-- Migration 302: Vehicle Tracking Enrichment
-- Adds performance indexes, geofence tables, and helper functions
-- for the enhanced vehicle tracking & dispatcher feature.
-- Safe to run online; indexes use CONCURRENTLY to avoid locking.

-- 1. Performance index on vehicles for bulk org query
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_org_status_active
    ON vehicles (tenant_id, status)
    WHERE deleted_at IS NULL AND status = 'active';

-- 2. Covering index for driver batch lookup
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_id_lookup
    ON drivers (id) INCLUDE (name, phone);

-- 3. Telemetry batch latest optimization
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_telemetry_vehicle_time_desc
    ON telemetry_points (vehicle_id, time DESC);

-- 4. Ensure geofence_assignments has proper index for vehicle-active lookup
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofence_assignments_vehicle_active
    ON geofence_assignments (vehicle_id, is_active);

-- 5. Geofence event query optimization (org-scoped)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofence_events_org_time
    ON geofence_events (tenant_id, time DESC)
    WHERE tenant_id IS NOT NULL;
