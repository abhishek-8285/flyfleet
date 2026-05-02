-- Migration: 220_repair_vehicle_foreign_keys
-- Description: Updates all legacy foreign key constraints to point to the new enterprise vehicles table.
-- Robustness: Uses DO blocks to handle missing tables and cleans up orphaned records.

BEGIN;

-- 1. DRIVERS
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'drivers') THEN
        ALTER TABLE drivers DROP CONSTRAINT IF EXISTS drivers_current_vehicle_id_fkey;
        ALTER TABLE drivers ADD CONSTRAINT drivers_current_vehicle_id_fkey 
            FOREIGN KEY (current_vehicle_id) REFERENCES vehicles(id) ON DELETE SET NULL;
    END IF;
END $$;

-- 2. TRIPS
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'trips') THEN
        -- Relax NOT NULL if it exists to allow SET NULL
        ALTER TABLE trips ALTER COLUMN vehicle_id DROP NOT NULL;
        -- Cleanup orphaned records
        UPDATE trips SET vehicle_id = NULL 
        WHERE vehicle_id IS NOT NULL 
        AND NOT EXISTS (SELECT 1 FROM vehicles WHERE id = trips.vehicle_id);

        ALTER TABLE trips DROP CONSTRAINT IF EXISTS trips_vehicle_id_fkey;
        ALTER TABLE trips ADD CONSTRAINT trips_vehicle_id_fkey 
            FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE SET NULL;
    END IF;
END $$;

-- 3. ALERTS
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'alerts') THEN
        ALTER TABLE alerts DROP CONSTRAINT IF EXISTS alerts_vehicle_id_fkey;
        ALTER TABLE alerts ADD CONSTRAINT alerts_vehicle_id_fkey 
            FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE;
    END IF;
END $$;

-- 4. TELEMETRY
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'telemetry_analytics') THEN
        ALTER TABLE telemetry_analytics DROP CONSTRAINT IF EXISTS fk_telemetry_analytics_vehicle;
        ALTER TABLE telemetry_analytics ADD CONSTRAINT fk_telemetry_analytics_vehicle 
            FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'telemetry_anomalies') THEN
        ALTER TABLE telemetry_anomalies DROP CONSTRAINT IF EXISTS fk_telemetry_anomalies_vehicle;
        ALTER TABLE telemetry_anomalies ADD CONSTRAINT fk_telemetry_anomalies_vehicle 
            FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE;
    END IF;
END $$;

-- 5. VIDEO SERVICES
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'video_recordings') THEN
        ALTER TABLE video_recordings DROP CONSTRAINT IF EXISTS fk_video_recordings_vehicle;
        ALTER TABLE video_recordings ADD CONSTRAINT fk_video_recordings_vehicle 
            FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'video_streams') THEN
        ALTER TABLE video_streams DROP CONSTRAINT IF EXISTS fk_video_streams_vehicle;
        ALTER TABLE video_streams ADD CONSTRAINT fk_video_streams_vehicle 
            FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE;
    END IF;
END $$;

-- 6. COMPLIANCE & DOCUMENTS
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'local_permits') THEN
        ALTER TABLE local_permits DROP CONSTRAINT IF EXISTS local_permits_vehicle_id_fkey;
        ALTER TABLE local_permits ADD CONSTRAINT local_permits_vehicle_id_fkey 
            FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'vehicle_verifications') THEN
        ALTER TABLE vehicle_verifications DROP CONSTRAINT IF EXISTS vehicle_verifications_vehicle_id_fkey;
        ALTER TABLE vehicle_verifications ADD CONSTRAINT vehicle_verifications_vehicle_id_fkey 
            FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE;
    END IF;
END $$;

-- 7. CONSUMPTION & TELEMETRY EVENTS
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'consumption_events') THEN
        ALTER TABLE consumption_events DROP CONSTRAINT IF EXISTS consumption_events_vehicle_id_fkey;
        ALTER TABLE consumption_events ADD CONSTRAINT consumption_events_vehicle_id_fkey 
            FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'telemetry_events') THEN
        ALTER TABLE telemetry_events DROP CONSTRAINT IF EXISTS telemetry_events_vehicle_id_fkey;
        ALTER TABLE telemetry_events ADD CONSTRAINT telemetry_events_vehicle_id_fkey 
            FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE;
    END IF;
END $$;

COMMIT;
