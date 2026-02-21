-- Fix missing schema pieces for the entire platform
-- Targets: organizations, users, telemetry, fuel_readings, fuel_logs, etc.

-- 0. Organizations table enrichment
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS created_by UUID;
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS updated_by UUID;
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT false;

-- 1. Users table enrichment
ALTER TABLE users ADD COLUMN IF NOT EXISTS org_id UUID REFERENCES organizations(id);
ALTER TABLE users ADD COLUMN IF NOT EXISTS role_id UUID;
ALTER TABLE users ADD COLUMN IF NOT EXISTS verified BOOLEAN DEFAULT true;
ALTER TABLE users ADD COLUMN IF NOT EXISTS first_name VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_name VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS created_by UUID;
ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_by UUID;

-- 2. Telemetry and related tables
CREATE TABLE IF NOT EXISTS telemetry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID,
    org_id UUID REFERENCES organizations(id),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    data JSONB
);

CREATE TABLE IF NOT EXISTS fuel_readings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID,
    org_id UUID REFERENCES organizations(id),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    level DOUBLE PRECISION
);

CREATE TABLE IF NOT EXISTS fuel_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID,
    org_id UUID REFERENCES organizations(id),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    amount DOUBLE PRECISION,
    cost DOUBLE PRECISION
);

-- 3. Ensure role_id exists if used in indices
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='role_id') THEN
        ALTER TABLE users ADD COLUMN role_id UUID;
    END IF;
END $$;

-- 4. Create common function if missing
CREATE OR REPLACE FUNCTION get_current_org_id() RETURNS UUID AS $$
    SELECT current_setting('app.current_org_id', true)::uuid;
$$ LANGUAGE SQL STABLE;
