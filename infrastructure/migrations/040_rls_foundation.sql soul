-- migration: 040_rls_foundation.sql
SET search_path TO public;

-- 1. Create a function to get the current tenant ID from session settings
CREATE OR REPLACE FUNCTION get_current_org_id() RETURNS UUID AS $$
    SELECT current_setting('app.current_org_id', true)::uuid;
$$ LANGUAGE SQL STABLE;

-- 2. Define a common application role if it doesn't exist
-- Note: In some environments, roles are managed outside migrations, 
-- but we include it here for consistency in the transformation.
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'flyfleet_app') THEN
        CREATE ROLE flyfleet_app;
    END IF;
END $$;

-- 3. Enable RLS and define policies for core tables
-- Organizations: A user can only see their own organization
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS org_tenant_policy ON organizations;
CREATE POLICY org_tenant_policy ON organizations
    FOR ALL
    TO flyfleet_app
    USING (id = get_current_org_id())
    WITH CHECK (id = get_current_org_id());

-- Users: A user can only see users in their organization
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS user_tenant_policy ON users;
CREATE POLICY user_tenant_policy ON users
    FOR ALL
    TO flyfleet_app
    USING (org_id = get_current_org_id())
    WITH CHECK (org_id = get_current_org_id());

-- Vehicles: Isolation by org_id
ALTER TABLE vehicles ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS vehicle_tenant_policy ON vehicles;
CREATE POLICY vehicle_tenant_policy ON vehicles
    FOR ALL
    TO flyfleet_app
    USING (org_id = get_current_org_id())
    WITH CHECK (org_id = get_current_org_id());

-- Telemetry: High volume table isolation
ALTER TABLE telemetry ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS telemetry_tenant_policy ON telemetry;
CREATE POLICY telemetry_tenant_policy ON telemetry
    FOR ALL
    TO flyfleet_app
    USING (org_id = get_current_org_id())
    WITH CHECK (org_id = get_current_org_id());

-- Trips: Driver and Org isolation
ALTER TABLE trips ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS trip_tenant_policy ON trips;
CREATE POLICY trip_tenant_policy ON trips
    FOR ALL
    TO flyfleet_app
    USING (org_id = get_current_org_id())
    WITH CHECK (org_id = get_current_org_id());

-- 4. Grant necessary permissions to the app role
GRANT SELECT, INSERT, UPDATE, DELETE ON organizations TO flyfleet_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON users TO flyfleet_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON vehicles TO flyfleet_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON telemetry TO flyfleet_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON trips TO flyfleet_app;
soul
