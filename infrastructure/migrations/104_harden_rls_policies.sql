-- 1. Create a secure getter to handle missing/malformed session variables
CREATE OR REPLACE FUNCTION get_current_org_id() RETURNS uuid AS $$
BEGIN
  RETURN current_setting('flyfleet.current_org_id')::uuid;
EXCEPTION
  WHEN OTHERS THEN RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

-- 2. Apply the "Fail-Closed" Policy to Vehicles
ALTER TABLE vehicles ENABLE ROW LEVEL SECURITY;

-- Drop any existing permissive or legacy policies
DROP POLICY IF EXISTS vehicle_tenant_isolation ON vehicles;
DROP POLICY IF EXISTS vehicle_tenant_isolation_legacy ON vehicles;

CREATE POLICY vehicle_tenant_isolation ON vehicles
    FOR ALL
    TO flyfleet_app_user
    USING (org_id = get_current_org_id());

-- 3. Apply the "Fail-Closed" Policy to Trips
ALTER TABLE trips ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS trip_tenant_isolation ON trips;

CREATE POLICY trip_tenant_isolation ON trips
    FOR ALL
    TO flyfleet_app_user
    USING (org_id = get_current_org_id());

-- 4. Apply the "Fail-Closed" Policy to GPS Pings (Telemetry)
ALTER TABLE telemetry_events ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS telemetry_tenant_isolation ON telemetry_events;

CREATE POLICY telemetry_tenant_isolation ON telemetry_events
    FOR ALL
    TO flyfleet_app_user
    USING (org_id = get_current_org_id());
