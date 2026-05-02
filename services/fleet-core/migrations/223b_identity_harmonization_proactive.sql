-- Migration: 223b_identity_harmonization_proactive
-- Description: Universal identity shield that ensures tenant_id, org_id, and organization_id exist on ALL public tables.
-- This prevents naming conflicts across all legacy and enterprise modules.

BEGIN;

DO $$ 
DECLARE
    r RECORD;
BEGIN
    -- Iterate over every base table in the public schema
    FOR r IN (
        SELECT table_name 
        FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_type = 'BASE TABLE'
        AND table_name NOT IN ('goose_db_version', 'spatial_ref_sys') -- Skip system tables
    )
    LOOP
        -- 1. Ensure tenant_id exists
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = r.table_name AND column_name = 'tenant_id') THEN
            EXECUTE format('ALTER TABLE %I ADD COLUMN tenant_id UUID', r.table_name);
            RAISE NOTICE 'Added tenant_id to %', r.table_name;
        END IF;
        
        -- 2. Ensure org_id exists
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = r.table_name AND column_name = 'org_id') THEN
            EXECUTE format('ALTER TABLE %I ADD COLUMN org_id UUID', r.table_name);
            RAISE NOTICE 'Added org_id to %', r.table_name;
        END IF;

        -- 3. Ensure organization_id exists
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = r.table_name AND column_name = 'organization_id') THEN
            EXECUTE format('ALTER TABLE %I ADD COLUMN organization_id UUID', r.table_name);
            RAISE NOTICE 'Added organization_id to %', r.table_name;
        END IF;

        -- 4. Cross-backfill values within the table (defensive)
        EXECUTE format('
            UPDATE %I 
            SET 
                tenant_id = COALESCE(tenant_id, org_id, organization_id),
                org_id = COALESCE(org_id, tenant_id, organization_id),
                organization_id = COALESCE(organization_id, tenant_id, org_id)
            WHERE tenant_id IS NULL OR org_id IS NULL OR organization_id IS NULL
        ', r.table_name);
    END LOOP;
END $$;

COMMIT;
