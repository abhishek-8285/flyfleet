-- Migration: 246_geofence_events_retention
-- Description: Add TimescaleDB retention policy for geofence events

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'timescaledb') THEN
        BEGIN
            PERFORM add_retention_policy('geofence_events', INTERVAL '180 days');
        EXCEPTION
            WHEN duplicate_object THEN NULL;
            WHEN undefined_function THEN NULL;
        END;
    END IF;
END $$;
