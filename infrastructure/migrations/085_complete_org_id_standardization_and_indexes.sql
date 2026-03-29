-- Migration: 085_complete_org_id_standardization_and_indexes.sql
-- Description: Complete org_id standardization and add missing FK indexes
-- Created: 2026-03-29
-- Dependencies: All previous migrations
-- This is a comprehensive cleanup migration

SET search_path TO public;

\echo '=====================================================';
\echo 'Starting comprehensive org_id standardization...';
\echo '=====================================================';

-- ========================
-- Step 1: Identify tables with only organization_id (need rename)
-- ========================
CREATE TEMP TABLE tmp_rename_tables (table_name text PRIMARY KEY);

INSERT INTO tmp_rename_tables
SELECT c.table_name
FROM information_schema.columns c
JOIN information_schema.tables t ON c.table_name = t.table_name AND c.table_schema = t.table_schema
WHERE c.table_schema = 'public'
  AND c.column_name = 'organization_id'
  AND t.table_type = 'BASE TABLE'
  AND NOT EXISTS (
    SELECT 1 FROM information_schema.columns c2
    WHERE c2.table_schema = 'public' AND c2.table_name = c.table_name AND c2.column_name = 'org_id'
  );

\echo 'Tables to rename (organization_id -> org_id): ' || (SELECT COUNT(*) FROM tmp_rename_tables);

-- ========================
-- Step 2: Identify tables with both columns (need drop organization_id)
-- ========================
CREATE TEMP TABLE tmp_drop_tables (table_name text PRIMARY KEY);

INSERT INTO tmp_drop_tables
SELECT c.table_name
FROM information_schema.columns c
WHERE c.table_schema = 'public'
  AND c.column_name = 'organization_id'
  AND EXISTS (
    SELECT 1 FROM information_schema.columns c2
    WHERE c2.table_schema = 'public' AND c2.table_name = c.table_name AND c2.column_name = 'org_id'
  );

\echo 'Tables to drop (redundant organization_id): ' || (SELECT COUNT(*) FROM tmp_drop_tables);

-- ========================
-- Step 3: Rename columns (organization_id -> org_id)
-- ========================
DO $$
DECLARE
    r RECORD;
    idx_count INT;
    pol_count INT;
BEGIN
    FOR r IN SELECT table_name FROM tmp_rename_tables ORDER BY table_name
    LOOP
        BEGIN
            -- Rename column
            EXECUTE format('ALTER TABLE %I RENAME COLUMN organization_id TO org_id', r.table_name);
            RAISE NOTICE 'Renamed column on %', r.table_name;

            -- After rename, ensure org_id is NOT NULL if it should be
            -- (skip for now - existing constraints should already be there or will be added later)

        EXCEPTION
            WHEN OTHERS THEN
                RAISE WARNING 'Failed to rename organization_id on %: %', r.table_name, SQLERRM;
        END;
    END LOOP;
END $$;

-- ========================
-- Step 4: Drop redundant organization_id columns (where org_id exists)
-- ========================
DO $$
DECLARE
    r RECORD;
BEGIN
    FOR r IN SELECT table_name FROM tmp_drop_tables ORDER BY table_name
    LOOP
        BEGIN
            -- First, ensure org_id has data (copy if needed)
            EXECUTE format(
                'UPDATE %I SET org_id = organization_id WHERE org_id IS NULL AND organization_id IS NOT NULL',
                r.table_name
            );

            -- Then drop organization_id
            EXECUTE format('ALTER TABLE %I DROP COLUMN IF EXISTS organization_id', r.table_name);
            RAISE NOTICE 'Dropped organization_id from %', r.table_name;

        EXCEPTION
            WHEN OTHERS THEN
                RAISE WARNING 'Failed to drop organization_id from %: %', r.table_name, SQLERRM;
        END;
    END LOOP;
END $$;

DROP TABLE tmp_rename_tables;
DROP TABLE tmp_drop_tables;

\echo '=====================================================';
\echo 'Org_id column standardization complete!';
\echo '=====================================================';

-- ========================
-- Step 5: Add FK constraints on org_id for any tables missing them
-- ========================
DO $$
DECLARE
    r RECORD;
    fk_name TEXT;
BEGIN
    FOR r IN
        SELECT DISTINCT table_name
        FROM information_schema.columns
        WHERE table_schema = 'public' AND column_name = 'org_id'
          AND table_name IN (
            SELECT table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE' AND table_schema = 'public'
          )
        ORDER BY table_name
    LOOP
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
                    -- Could fail due to data issues (NULLs, orphans)
                    RAISE WARNING 'Could not add FK % to %: %', fk_name, r.table_name, SQLERRM;
            END;
        END IF;
    END LOOP;
END $$;

\echo 'Foreign key constraints ensured.';

-- ========================
-- Step 6: Create indexes on org_id where missing
-- ========================
DO $$
DECLARE
    r RECORD;
    idx_name TEXT;
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
    END LOOP;
END $$;

\echo 'org_id indexes ensured.';

-- ========================
-- Step 7: Ensure RLS is enabled and policy exists (standard naming)
-- ========================
DO $$
DECLARE
    r RECORD;
    policy_name TEXT;
    existing_policy TEXT;
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
        -- Enable RLS if not enabled
        IF NOT EXISTS (
            SELECT 1 FROM pg_tables WHERE schemaname = 'public' AND tablename = r.table_name AND rowsecurity = true
        ) THEN
            EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', r.table_name);
            RAISE NOTICE 'Enabled RLS on %', r.table_name;
        END IF;

        policy_name := r.table_name || '_tenant_policy';

        -- Check if a policy with standard name exists
        IF NOT EXISTS (
            SELECT 1 FROM pg_policies WHERE tablename = r.table_name AND policyname = policy_name
        ) THEN
            -- Drop any policies that might be using organization_id (old style)
            FOR existing_policy IN
                SELECT policyname FROM pg_policies WHERE tablename = r.table_name
            LOOP
                EXECUTE format('DROP POLICY %I ON %I', existing_policy, r.table_name);
            END LOOP;

            -- Create new standard policy
            EXECUTE format(
                'CREATE POLICY %I ON %I FOR ALL USING (org_id = get_current_org_id()) WITH CHECK (org_id = get_current_org_id())',
                policy_name, r.table_name
            );
            RAISE NOTICE 'Created RLS policy: % on %', policy_name, r.table_name;
        END IF;
    END LOOP;
END $$;

\echo 'RLS policies ensured.';

-- ========================
-- Step 8: Verification
-- ========================
\echo '=====================================================';
\echo 'Verification:';
SELECT COUNT(*) FROM information_schema.columns WHERE column_name = 'organization_id' AND table_schema = 'public' AND table_name IN (SELECT table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE') as remaining_org_id_columns;
\echo 'Remaining base tables with organization_id (should be 0): ';
SELECT COUNT(*) FROM information_schema.columns WHERE column_name = 'organization_id' AND table_schema = 'public' AND table_name IN (SELECT table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE');

SELECT COUNT(*) FROM pg_policies WHERE policyname NOT LIKE '%_tenant_policy' AND tablename IN (SELECT table_name FROM information_schema.columns WHERE column_name = 'org_id' AND table_schema = 'public') as non_standard_policies;
\echo 'Non-standard RLS policies (should be 0): ';
SELECT COUNT(*) FROM pg_policies WHERE policyname NOT LIKE '%_tenant_policy' AND tablename IN (SELECT table_name FROM information_schema.columns WHERE column_name = 'org_id' AND table_schema = 'public');

\echo '=====================================================';
\echo 'COMPREHENSIVE STANDARDIZATION COMPLETE!';
\echo '=====================================================';
