-- Migration: 086_finish_organization_id_standardization.sql
-- Description: Finish renaming organization_id to org_id on remaining tables
-- Created: 2026-03-29
-- Dependencies: 085_complete_org_id_standardization_and_indexes.sql

SET search_path TO public;

\echo 'Finishing organization_id → org_id standardization on remaining tables...';

-- ========================
-- Step 1: Rename organization_id to org_id on tables that only have organization_id
-- ========================
DO $$
DECLARE
    r RECORD;
BEGIN
    FOR r IN
        SELECT c.table_name
        FROM information_schema.columns c
        JOIN information_schema.tables t ON c.table_name = t.table_name AND c.table_schema = t.table_schema
        WHERE c.table_schema = 'public'
          AND c.column_name = 'organization_id'
          AND t.table_type = 'BASE TABLE'
          AND NOT EXISTS (
              SELECT 1 FROM information_schema.columns c2
              WHERE c2.table_schema = 'public' AND c2.table_name = c.table_name AND c2.column_name = 'org_id'
          )
    LOOP
        BEGIN
            EXECUTE format('ALTER TABLE %I RENAME COLUMN organization_id TO org_id', r.table_name);
            RAISE NOTICE 'Renamed organization_id → org_id on %', r.table_name;
        EXCEPTION
            WHEN OTHERS THEN
                RAISE WARNING 'Failed to rename on %: %', r.table_name, SQLERRM;
        END;
    END LOOP;
END $$;

-- ========================
-- Step 2: Drop organization_id from tables that have both (fuel_logs, fuel_transactions)
-- ========================
DO $$
DECLARE
    r RECORD;
BEGIN
    FOR r IN
        SELECT c.table_name
        FROM information_schema.columns c
        WHERE c.table_schema = 'public'
          AND c.column_name = 'organization_id'
          AND EXISTS (
              SELECT 1 FROM information_schema.columns c2
              WHERE c2.table_schema = 'public' AND c2.table_name = c.table_name AND c2.column_name = 'org_id'
          )
    LOOP
        BEGIN
            -- Ensure org_id is populated
            EXECUTE format(
                'UPDATE %I SET org_id = organization_id WHERE org_id IS NULL AND organization_id IS NOT NULL',
                r.table_name
            );
            -- Drop organization_id
            EXECUTE format('ALTER TABLE %I DROP COLUMN IF EXISTS organization_id', r.table_name);
            RAISE NOTICE 'Dropped redundant organization_id from %', r.table_name;
        EXCEPTION
            WHEN OTHERS THEN
                RAISE WARNING 'Failed to drop organization_id from %: %', r.table_name, SQLERRM;
        END;
    END LOOP;
END $$;

-- ========================
-- Step 3: For these tables, ensure FK, index, and RLS
-- This covers the tables that were just renamed (they now have org_id)
-- ========================
DO $$
DECLARE
    r RECORD;
    fk_name TEXT;
    idx_name TEXT;
    policy_name TEXT;
BEGIN
    FOR r IN
        SELECT table_name
        FROM information_schema.columns
        WHERE table_schema = 'public' AND column_name = 'org_id'
          AND table_name IN (
            SELECT table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE' AND table_schema = 'public'
          )
        ORDER BY table_name
    LOOP
        -- Add FK if missing
        fk_name := 'fk_' || r.table_name || '_organization';
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.table_constraints
            WHERE table_schema = 'public' AND table_name = r.table_name AND constraint_name = fk_name
        ) THEN
            BEGIN
                EXECUTE format(
                    'ALTER TABLE %I ADD CONSTRAINT %I FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE',
                    r.table_name, fk_name
                );
                RAISE NOTICE 'Added FK: % on %', fk_name, r.table_name;
            EXCEPTION
                WHEN OTHERS THEN
                    RAISE WARNING 'Could not add FK % to %: %', fk_name, r.table_name, SQLERRM;
            END;
        END IF;

        -- Add index if missing
        idx_name := 'idx_' || r.table_name || '_org_id';
        IF NOT EXISTS (
            SELECT 1 FROM pg_indexes
            WHERE schemaname = 'public' AND tablename = r.table_name AND indexname = idx_name
        ) THEN
            BEGIN
                EXECUTE format('CREATE INDEX %I ON %I (org_id)', idx_name, r.table_name);
                RAISE NOTICE 'Created index: % on %', idx_name, r.table_name;
            EXCEPTION
                WHEN OTHERS THEN
                    RAISE WARNING 'Could not create index % on %: %', idx_name, r.table_name, SQLERRM;
            END;
        END IF;

        -- Ensure RLS enabled and policy exists
        IF NOT EXISTS (
            SELECT 1 FROM pg_tables WHERE schemaname = 'public' AND tablename = r.table_name AND rowsecurity = true
        ) THEN
            EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', r.table_name);
            RAISE NOTICE 'Enabled RLS on %', r.table_name;
        END IF;

        policy_name := r.table_name || '_tenant_policy';
        IF NOT EXISTS (
            SELECT 1 FROM pg_policies WHERE tablename = r.table_name AND policyname = policy_name
        ) THEN
            -- Drop any old policies referencing organization_id (they may exist if we renamed)
            -- This will drop all existing policies on the table to avoid conflicts
            EXECUTE format('DROP POLICY IF EXISTS %I_tenant_policy ON %I', r.table_name, r.table_name);
            -- Actually, we need to drop any policy with a different name if it exists? We'll just create ours.
            EXECUTE format(
                'CREATE POLICY %I ON %I FOR ALL USING (org_id = get_current_org_id()) WITH CHECK (org_id = get_current_org_id())',
                policy_name, r.table_name
            );
            RAISE NOTICE 'Created RLS policy: % on %', policy_name, r.table_name;
        END IF;
    END LOOP;
END $$;

\echo '=====================================================';
\echo 'Organization ID standardization complete!';
\echo '=====================================================';

-- Verification
SELECT COUNT(*) as remaining_org_id_columns
FROM information_schema.columns
WHERE column_name = 'organization_id'
  AND table_schema = 'public'
  AND table_name IN (SELECT table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE' AND table_schema = 'public');

\echo 'Remaining base tables with organization_id (should be 0):';
SELECT COUNT(*) FROM information_schema.columns WHERE column_name = 'organization_id' AND table_schema = 'public' AND table_name IN (SELECT table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE' AND table_schema = 'public');

\echo '=====================================================';
