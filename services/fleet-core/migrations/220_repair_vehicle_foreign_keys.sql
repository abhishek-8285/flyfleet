-- Migration: 220_repair_vehicle_foreign_keys
-- Description: Updates all legacy foreign key constraints to point to the new enterprise vehicles table.

BEGIN;

-- 1. DRIVERS
ALTER TABLE drivers DROP CONSTRAINT IF EXISTS drivers_current_vehicle_id_fkey;
ALTER TABLE drivers ADD CONSTRAINT drivers_current_vehicle_id_fkey 
    FOREIGN KEY (current_vehicle_id) REFERENCES vehicles(id) ON DELETE SET NULL;

-- 2. TRIPS
ALTER TABLE trips DROP CONSTRAINT IF EXISTS trips_vehicle_id_fkey;
ALTER TABLE trips ADD CONSTRAINT trips_vehicle_id_fkey 
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE SET NULL;

-- 3. ALERTS
ALTER TABLE alerts DROP CONSTRAINT IF EXISTS alerts_vehicle_id_fkey;
ALTER TABLE alerts ADD CONSTRAINT alerts_vehicle_id_fkey 
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE;

-- 4. TELEMETRY
ALTER TABLE telemetry_analytics DROP CONSTRAINT IF EXISTS fk_telemetry_analytics_vehicle;
ALTER TABLE telemetry_analytics ADD CONSTRAINT fk_telemetry_analytics_vehicle 
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE;

ALTER TABLE telemetry_anomalies DROP CONSTRAINT IF EXISTS fk_telemetry_anomalies_vehicle;
ALTER TABLE telemetry_anomalies ADD CONSTRAINT fk_telemetry_anomalies_vehicle 
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE;

-- 5. VIDEO SERVICES
ALTER TABLE video_recordings DROP CONSTRAINT IF EXISTS fk_video_recordings_vehicle;
ALTER TABLE video_recordings ADD CONSTRAINT fk_video_recordings_vehicle 
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE;

ALTER TABLE video_streams DROP CONSTRAINT IF EXISTS fk_video_streams_vehicle;
ALTER TABLE video_streams ADD CONSTRAINT fk_video_streams_vehicle 
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE;

-- 6. COMPLIANCE & DOCUMENTS
ALTER TABLE local_permits DROP CONSTRAINT IF EXISTS local_permits_vehicle_id_fkey;
ALTER TABLE local_permits ADD CONSTRAINT local_permits_vehicle_id_fkey 
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE;

ALTER TABLE vehicle_verifications DROP CONSTRAINT IF EXISTS vehicle_verifications_vehicle_id_fkey;
ALTER TABLE vehicle_verifications ADD CONSTRAINT vehicle_verifications_vehicle_id_fkey 
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE;

-- 7. CONSUMPTION & TELEMETRY EVENTS
ALTER TABLE consumption_events DROP CONSTRAINT IF EXISTS consumption_events_vehicle_id_fkey;
ALTER TABLE consumption_events ADD CONSTRAINT consumption_events_vehicle_id_fkey 
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE;

-- Note: Partitions inherit constraints from the parent table in many setups, 
-- but we explicitly handle the main table telemetry_events.
ALTER TABLE telemetry_events DROP CONSTRAINT IF EXISTS telemetry_events_vehicle_id_fkey;
ALTER TABLE telemetry_events ADD CONSTRAINT telemetry_events_vehicle_id_fkey 
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE;

COMMIT;
