-- Migration 301: Vehicle Tracking Performance Indexes
-- Purpose: Optimize bulk vehicle location queries, geofence spatial
--          checks, and history trail lookups for <2s P95 map loads.
-- Execution: Online (CONCURRENTLY) — no table locking.
-- Rollback: Drop indexes individually.

BEGIN;

-- 1. Bulk vehicle locations: speed up latest-position-per-vehicle query.
--    Supports: SELECT DISTINCT ON (vehicle_id) ... ORDER BY vehicle_id, time DESC
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_telemetry_vehicle_time_desc
    ON telemetry_points (vehicle_id, time DESC);

-- 2. Geofence spatial queries (polygon containment via PostGIS GIST index).
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofences_polygon_gist
    ON geofences USING GIST (polygon);

-- 3. Geofence lookup by org + status (list all active geofences for an org).
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofences_org_status
    ON geofences (org_id, status);

-- 4. Geofence event queries by vehicle (recent events for a vehicle).
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofence_events_vehicle_time
    ON geofence_events (vehicle_id, occurred_at DESC);

-- 5. Geofence event queries by geofence (recent events for a geofence).
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofence_events_geofence_time
    ON geofence_events (geofence_id, occurred_at DESC);

-- 6. Vehicle org lookup (should already exist — ensure it does).
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_org_status_deleted
    ON vehicles (org_id, status, deleted_at)
    WHERE deleted_at IS NULL;

-- 7. Add alert tracking columns to geofence_events (idempotent).
ALTER TABLE geofence_events
    ADD COLUMN IF NOT EXISTS alert_sent BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS alert_channels JSONB DEFAULT '["in_app"]';

COMMIT;

-- Rollback (if needed):
-- DROP INDEX CONCURRENTLY IF EXISTS idx_telemetry_vehicle_time_desc;
-- DROP INDEX CONCURRENTLY IF EXISTS idx_geofences_polygon_gist;
-- DROP INDEX CONCURRENTLY IF EXISTS idx_geofences_org_status;
-- DROP INDEX CONCURRENTLY IF EXISTS idx_geofence_events_vehicle_time;
-- DROP INDEX CONCURRENTLY IF EXISTS idx_geofence_events_geofence_time;
-- DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_org_status_deleted;
-- ALTER TABLE geofence_events DROP COLUMN IF EXISTS alert_sent, DROP COLUMN IF EXISTS alert_channels;
