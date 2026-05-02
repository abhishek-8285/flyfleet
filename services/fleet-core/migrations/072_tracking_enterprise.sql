-- Migration: 072_tracking_enterprise
-- Description: Enterprise Tracking Harmonization.

-- 1. Evolve telemetry_points (Additive only)
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'telemetry_points' AND column_name = 'vehicle_id') THEN
        ALTER TABLE telemetry_points ADD COLUMN vehicle_id UUID;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'telemetry_points' AND column_name = 'device_id') THEN
        ALTER TABLE telemetry_points ADD COLUMN device_id VARCHAR(100);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'telemetry_points' AND column_name = 'tenant_id') THEN
        ALTER TABLE telemetry_points ADD COLUMN tenant_id UUID REFERENCES tenants(id);
    END IF;
    -- PostGIS geo column skipped
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- 2. Device Registry
CREATE TABLE IF NOT EXISTS device_registry (
    device_id VARCHAR(100) PRIMARY KEY
);

DO $$ 
BEGIN 
    ALTER TABLE device_registry ADD COLUMN IF NOT EXISTS vehicle_id UUID REFERENCES vehicles(id);
    ALTER TABLE device_registry ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id);
    ALTER TABLE device_registry ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'active';
    ALTER TABLE device_registry ADD COLUMN IF NOT EXISTS last_heartbeat TIMESTAMPTZ;
    ALTER TABLE device_registry ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- 3. Live Location View
CREATE OR REPLACE VIEW v_live_locations AS
SELECT DISTINCT ON (vehicle_id)
    vehicle_id,
    device_id,
    lat,
    lng,
    speed,
    recorded_at AS last_seen
FROM telemetry_points
ORDER BY vehicle_id, recorded_at DESC;
