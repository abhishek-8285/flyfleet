-- Migration: 103_driver_rls_guardrails
-- Description: Driver-specific RLS guardrails with standardized session context.

-- 1. Ensure assigned_driver_id exists (standardized name)
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'assigned_driver_id') THEN
        ALTER TABLE vehicles ADD COLUMN assigned_driver_id UUID REFERENCES drivers(id);
    END IF;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- 2. Enable RLS
ALTER TABLE vehicles ENABLE ROW LEVEL SECURITY;

-- 3. Policy: Tenant Admin Isolation (Can see all in THEIR tenant)
DROP POLICY IF EXISTS admin_bypass_rls ON vehicles;
CREATE POLICY admin_tenant_isolation ON vehicles
    FOR ALL
    USING (
        (current_setting('app.role', true) IN ('super_admin', 'org_admin', 'fleet_manager'))
        AND (tenant_id = current_setting('app.tenant_id', true)::UUID OR current_setting('app.role', true) = 'super_admin')
    );

-- 4. Policy: Driver Strict Isolation (Can only see assigned vehicle)
DROP POLICY IF EXISTS driver_strict_rls_policy ON vehicles;
CREATE POLICY driver_strict_rls_policy ON vehicles
    FOR SELECT
    USING (
        (current_setting('app.role', true) = 'DRIVER')
        AND (current_setting('app.user_id', true)::UUID = assigned_driver_id)
    );
