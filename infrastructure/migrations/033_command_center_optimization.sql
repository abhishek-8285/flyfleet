-- Command Center Optimization Migration
-- Creates materialized view for live fleet status and optimized indexes

-- 1. Create Materialized View for Live Fleet Status
-- This view aggregates the latest known state for all active vehicles
CREATE MATERIALIZED VIEW IF NOT EXISTS vehicle_status AS
WITH latest_telemetry AS (
    SELECT DISTINCT ON (vehicle_id)
        vehicle_id,
        time,
        latitude,
        longitude,
        speed_kph,
        heading,
        odometer_km,
        fuel_level_percent,
        engine_temp_celsius,
        alerts
    FROM vehicle_telemetry
    ORDER BY vehicle_id, time DESC
)
SELECT 
    v.id AS vehicle_id,
    v.registration_number,
    v.status AS vehicle_registry_status,
    lt.time AS last_update,
    lt.latitude,
    lt.longitude,
    lt.speed_kph,
    lt.heading,
    lt.odometer_km,
    lt.fuel_level_percent,
    lt.engine_temp_celsius,
    lt.alerts,
    CASE 
        WHEN lt.time < NOW() - INTERVAL '5 minutes' THEN 'offline'
        WHEN lt.speed_kph > 5 THEN 'moving'
        ELSE 'idle'
    END AS operational_status,
    d.name AS driver_name,
    d.id AS driver_id
FROM vehicles v
LEFT JOIN latest_telemetry lt ON v.id = lt.vehicle_id
LEFT JOIN drivers d ON v.current_driver_id = d.id
WHERE v.is_deleted = FALSE;

-- 2. Create Unique Index for Materialized View (Required for Concurrent Refresh)
CREATE UNIQUE INDEX IF NOT EXISTS idx_vehicle_status_vehicle_id ON vehicle_status (vehicle_id);

-- 3. Additional Performance Indexes for Command Center
-- Optimized history lookups
CREATE INDEX IF NOT EXISTS idx_gps_locations_vehicle_time_desc ON gps_locations (vehicle_id, time DESC);

-- 4. Helper function to refresh the view (can be called by a trigger or cron)
CREATE OR REPLACE FUNCTION refresh_vehicle_status()
RETURNS trigger AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY vehicle_status;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- 5. Comment on the view
COMMENT ON MATERIALIZED VIEW vehicle_status IS 'Aggregated real-time status of all active vehicles for the Command Center';
