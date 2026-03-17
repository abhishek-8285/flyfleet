-- +goose Up
-- +goose StatementBegin

-- 1. Ensure `assigned_driver_id` exists on `vehicles` exactly as requested.
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS assigned_driver_id UUID;

-- 2. Enable Row Level Security (RLS) on vehicles table
ALTER TABLE vehicles ENABLE ROW LEVEL SECURITY;

-- 3. Policy: SUPER_ADMIN and ORG_ADMIN bypassing RLS (they see all or all in tenant)
CREATE POLICY admin_bypass_rls ON vehicles
    FOR ALL
    USING (
        current_setting('flyfleet.current_role', true) IN ('super_admin', 'org_admin', 'fleet_manager')
    );

-- 4. Policy: Standard driver can only see the vehicle assigned to them.
CREATE POLICY driver_strict_rls_policy ON vehicles
    FOR SELECT
    USING (
        current_setting('flyfleet.current_role', true) IN ('DRIVER', 'STAFF') 
        AND current_setting('flyfleet.current_user_id', true) = assigned_driver_id::text
    );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP POLICY IF EXISTS driver_strict_rls_policy ON vehicles;
DROP POLICY IF EXISTS admin_bypass_rls ON vehicles;
ALTER TABLE vehicles DISABLE ROW LEVEL SECURITY;
ALTER TABLE vehicles DROP COLUMN IF EXISTS assigned_driver_id;

-- +goose StatementEnd
