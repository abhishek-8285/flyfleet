-- Migration: 245_geofence_event_idempotency
-- Description: Add idempotency guard for geofence events at 1-minute bucket granularity

CREATE UNIQUE INDEX IF NOT EXISTS uniq_geofence_event
ON geofence_events (
    time,
    vehicle_id,
    geofence_id,
    event_type
);
