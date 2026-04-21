-- Migration: Standardize Identity Schema (org_id -> tenant_id) - V3 (Final)
-- Date: 2026-04-19

BEGIN;

-- 1. Helper Function
CREATE OR REPLACE FUNCTION get_current_tenant_id() RETURNS uuid AS $$
BEGIN
    RETURN current_setting('flyfleet.current_org_id')::uuid;
EXCEPTION
    WHEN OTHERS THEN RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- 2. Drop All Policies Mentioned to Avoid Dependency Issues
DROP POLICY IF EXISTS audit_logs_tenant_policy ON audit_logs;
DROP POLICY IF EXISTS documents_tenant_policy ON documents;
DROP POLICY IF EXISTS drivers_tenant_policy ON drivers;
DROP POLICY IF EXISTS invoice_items_tenant_policy ON invoice_items;
DROP POLICY IF EXISTS vehicle_tenant_policy ON vehicles;
DROP POLICY IF EXISTS vehicles_tenant_policy ON vehicles;
DROP POLICY IF EXISTS vehicles_legacy_tenant_policy ON vehicles_legacy;
DROP POLICY IF EXISTS vehicles_tenant_policy ON vehicles_legacy;
DROP POLICY IF EXISTS work_orders_tenant_policy ON work_orders;

-- 3. Robust Column Adjustments
DO $$ 
BEGIN
    -- Vehicles
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vehicles' AND column_name='org_id') AND 
       NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vehicles' AND column_name='tenant_id') THEN
        ALTER TABLE vehicles RENAME COLUMN org_id TO tenant_id;
    END IF;

    -- Drivers
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='drivers' AND column_name='org_id') AND 
       NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='drivers' AND column_name='tenant_id') THEN
        ALTER TABLE drivers RENAME COLUMN org_id TO tenant_id;
    END IF;

    -- Drop legacy org_id where tenant_id already exists
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='audit_logs' AND column_name='org_id') THEN
        ALTER TABLE audit_logs DROP COLUMN org_id;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='documents' AND column_name='org_id') THEN
        ALTER TABLE documents DROP COLUMN org_id;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='invoice_items' AND column_name='org_id') THEN
        ALTER TABLE invoice_items DROP COLUMN org_id;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vehicles_legacy' AND column_name='org_id') THEN
        ALTER TABLE vehicles_legacy DROP COLUMN org_id;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='work_orders' AND column_name='org_id') THEN
        ALTER TABLE work_orders DROP COLUMN org_id;
    END IF;
END $$;

-- 4. Rebuild Constraints and Indices
ALTER INDEX IF EXISTS idx_vehicles_org_id RENAME TO idx_vehicles_tenant_id;
ALTER TABLE vehicles DROP CONSTRAINT IF EXISTS fk_vehicles_organization;
ALTER TABLE vehicles ADD CONSTRAINT fk_vehicles_organization FOREIGN KEY (tenant_id) REFERENCES organizations(id) ON DELETE CASCADE;

ALTER INDEX IF EXISTS idx_drivers_org_id RENAME TO idx_drivers_tenant_id;
ALTER TABLE drivers DROP CONSTRAINT IF EXISTS fk_drivers_organization;
ALTER TABLE drivers ADD CONSTRAINT fk_drivers_organization FOREIGN KEY (tenant_id) REFERENCES organizations(id) ON DELETE CASCADE;

-- 5. Recreate All Policies using tenant_id
CREATE POLICY vehicles_tenant_policy ON vehicles USING (tenant_id = get_current_tenant_id()) WITH CHECK (tenant_id = get_current_tenant_id());
CREATE POLICY drivers_tenant_policy ON drivers USING (tenant_id = get_current_tenant_id()) WITH CHECK (tenant_id = get_current_tenant_id());
CREATE POLICY audit_logs_tenant_policy ON audit_logs USING (tenant_id = get_current_tenant_id()) WITH CHECK (tenant_id = get_current_tenant_id());
CREATE POLICY documents_tenant_policy ON documents USING (tenant_id = get_current_tenant_id()) WITH CHECK (tenant_id = get_current_tenant_id());
CREATE POLICY invoice_items_tenant_policy ON invoice_items USING (tenant_id = get_current_tenant_id()) WITH CHECK (tenant_id = get_current_tenant_id());
CREATE POLICY vehicles_legacy_tenant_policy ON vehicles_legacy USING (tenant_id = get_current_tenant_id()) WITH CHECK (tenant_id = get_current_tenant_id());
CREATE POLICY work_orders_tenant_policy ON work_orders USING (tenant_id = get_current_tenant_id()) WITH CHECK (tenant_id = get_current_tenant_id());

COMMIT;
