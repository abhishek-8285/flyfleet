-- Migration: 082_standardize_vehicles_tenant_id.sql
-- Description: Rename vehicles.tenant_id to org_id to match multi-tenancy standard
-- Created: 2026-03-29
-- Purpose: Complete org_id standardization - vehicles was the only base table using tenant_id
-- Dependencies: 071_standardize_organization_id_naming.sql, 077_finalize_org_id_standardization.sql

SET search_path TO public;

\echo 'Standardizing vehicles.tenant_id to org_id...';

-- ========================================
-- Step 1: Verify all tenant_id values are valid organization IDs
-- ========================================
DO $$
DECLARE
    orphan_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO orphan_count
    FROM vehicles v
    LEFT JOIN organizations o ON v.tenant_id = o.id
    WHERE v.tenant_id IS NOT NULL AND o.id IS NULL;

    IF orphan_count > 0 THEN
        RAISE EXCEPTION 'Found % vehicles with invalid tenant_id (no matching organization). Fix data first.', orphan_count;
    END IF;
END $$;

\echo '✅ All vehicles.tenant_id values reference valid organizations';

-- ========================================
-- Step 2: Rename the column tenant_id -> org_id
-- ========================================
ALTER TABLE vehicles RENAME COLUMN tenant_id TO org_id;

\echo '✅ Renamed vehicles.tenant_id to vehicles.org_id';

-- ========================================
-- Step 3: Add NOT NULL constraint (tenant_id was already NOT NULL)
-- ========================================
-- Already NOT NULL from schema, just ensuring it's documented
ALTER TABLE vehicles ALTER COLUMN org_id SET NOT NULL;

\echo '✅ vehicles.org_id is NOT NULL';

-- ========================================
-- Step 4: Add foreign key constraint (if not already present)
-- ========================================
-- Check if FK already exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints tc
        JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
        WHERE tc.constraint_type = 'FOREIGN KEY'
          AND tc.table_name = 'vehicles'
          AND kcu.column_name = 'org_id'
    ) THEN
        ALTER TABLE vehicles
        ADD CONSTRAINT fk_vehicles_organization
        FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE;
        \echo '✅ Added FK: vehicles.org_id → organizations.id';
    ELSE
        \echo '✅ FK already exists, skipping';
    END IF;
END $$;

-- ========================================
-- Step 5: Create standard indexes
-- ========================================

-- Create idx_vehicles_org_id if not exists
CREATE INDEX IF NOT EXISTS idx_vehicles_org_id ON vehicles(org_id);
\echo '✅ Created index: idx_vehicles_org_id';

-- Create composite index: org_id + status (common query pattern)
CREATE INDEX IF NOT EXISTS idx_vehicles_org_status ON vehicles(org_id, status);
\echo '✅ Created index: idx_vehicles_org_status';

-- Create composite index: org_id + created_at (for listing by org with time filtering)
CREATE INDEX IF NOT EXISTS idx_vehicles_org_created ON vehicles(org_id, created_at DESC);
\echo '✅ Created index: idx_vehicles_org_created';

-- ========================================
-- Step 6: Enable RLS if not already enabled
-- ========================================
DO $$
BEGIN
    -- Check if RLS is already enabled
    IF NOT EXISTS (
        SELECT 1 FROM pg_tables
        WHERE schemaname = 'public' AND tablename = 'vehicles' AND rowsecurity = true
    ) THEN
        ALTER TABLE vehicles ENABLE ROW LEVEL SECURITY;
        \echo '✅ Enabled RLS on vehicles';

        -- Create tenant isolation policy
        CREATE POLICY vehicles_tenant_policy ON vehicles
            FOR ALL
            USING (org_id = get_current_org_id())
            WITH CHECK (org_id = get_current_org_id());
        \echo '✅ Created RLS policy: vehicles_tenant_policy';
    ELSE
        \echo '✅ RLS already enabled on vehicles';

        -- Check if policy exists, if not create it
        IF NOT EXISTS (
            SELECT 1 FROM pg_policies WHERE tablename = 'vehicles' AND policyname = 'vehicles_tenant_policy'
        ) THEN
            CREATE POLICY vehicles_tenant_policy ON vehicles
                FOR ALL
                USING (org_id = get_current_org_id())
                WITH CHECK (org_id = get_current_org_id());
            \echo '✅ Created missing RLS policy: vehicles_tenant_policy';
        END IF;
    END IF;
END $$;

-- ========================================
-- Step 7: Verify all indexes on vehicles
-- ========================================
\echo 'Current indexes on vehicles:';
SELECT indexname, indexdef FROM pg_indexes
WHERE schemaname = 'public' AND tablename = 'vehicles'
ORDER BY indexname;

\echo '=====================================================';
\echo 'vehicles.tenant_id successfully standardized to org_id!';
\echo '=====================================================';
