-- migration: 041_timescaledb_telemetry.sql
SET search_path TO public;

-- 1. Add org_id to telemetry tables for RLS and isolation
-- Note: These columns are required for the new ingestion layer and RLS policies
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'gps_locations' AND column_name = 'org_id') THEN
        ALTER TABLE gps_locations ADD COLUMN org_id UUID REFERENCES organizations(id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'obd_readings' AND column_name = 'org_id') THEN
        ALTER TABLE obd_readings ADD COLUMN org_id UUID REFERENCES organizations(id);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'fuel_readings' AND column_name = 'org_id') THEN
        ALTER TABLE fuel_readings ADD COLUMN org_id UUID REFERENCES organizations(id);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicle_telemetry' AND column_name = 'org_id') THEN
        ALTER TABLE vehicle_telemetry ADD COLUMN org_id UUID REFERENCES organizations(id);
    END IF;
END $$;

-- 2. Apply RLS policies to the new telemetry tables
ALTER TABLE gps_locations ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS gps_tenant_policy ON gps_locations;
CREATE POLICY gps_tenant_policy ON gps_locations
    FOR ALL TO flyfleet_app
    USING (org_id = get_current_org_id())
    WITH CHECK (org_id = get_current_org_id());

ALTER TABLE obd_readings ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS obd_tenant_policy ON obd_readings;
CREATE POLICY obd_tenant_policy ON obd_readings
    FOR ALL TO flyfleet_app
    USING (org_id = get_current_org_id())
    WITH CHECK (org_id = get_current_org_id());

ALTER TABLE fuel_readings ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS fuel_tenant_policy ON fuel_readings;
CREATE POLICY fuel_tenant_policy ON fuel_readings
    FOR ALL TO flyfleet_app
    USING (org_id = get_current_org_id())
    WITH CHECK (org_id = get_current_org_id());

-- 3. Convert to Hypertables (Optimized for Time-Series)
-- We drop existing PKs if they interfere with hypertable creation (hypertables must have time in PK)
-- In 032_telemetry_optimization.sql, PK was (time, vehicle_id). This is fine for Timescale.
SELECT create_hypertable('gps_locations', 'time', if_not_exists => TRUE);
SELECT create_hypertable('obd_readings', 'time', if_not_exists => TRUE);
SELECT create_hypertable('fuel_readings', 'time', if_not_exists => TRUE);
SELECT create_hypertable('vehicle_telemetry', 'time', if_not_exists => TRUE);

-- 4. Compression Policies (Reduce storage costs)
-- Compress chunks older than 7 days
ALTER TABLE gps_locations SET (timescaledb.compress, timescaledb.compress_segmentby = 'vehicle_id, org_id');
SELECT add_compression_policy('gps_locations', INTERVAL '7 days');

ALTER TABLE obd_readings SET (timescaledb.compress, timescaledb.compress_segmentby = 'vehicle_id, org_id');
SELECT add_compression_policy('obd_readings', INTERVAL '7 days');

ALTER TABLE fuel_readings SET (timescaledb.compress, timescaledb.compress_segmentby = 'vehicle_id, org_id');
SELECT add_compression_policy('fuel_readings', INTERVAL '7 days');

-- 5. Retention Policies (Automated Cleanup)
-- Drop raw data older than 90 days
SELECT add_retention_policy('gps_locations', INTERVAL '90 days');
SELECT add_retention_policy('obd_readings', INTERVAL '90 days');
SELECT add_retention_policy('fuel_readings', INTERVAL '90 days');

-- 6. Optimization: Global Index on vehicle_id + time
-- Hypertables automatically index 'time', but we often query by vehicle
CREATE INDEX IF NOT EXISTS idx_gps_vehicle_time_brin ON gps_locations USING BRIN (vehicle_id, time);
CREATE INDEX IF NOT EXISTS idx_obd_vehicle_time_brin ON obd_readings USING BRIN (vehicle_id, time);
