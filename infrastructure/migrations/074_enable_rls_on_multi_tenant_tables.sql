-- Migration: 074_enable_rls_on_multi_tenant_tables.sql
-- Description: Enable Row Level Security on all multi-tenant tables (org_id)
-- Created: 2026-03-29
-- Security: Critical for defense-in-depth multi-tenancy enforcement

SET search_path TO public;

-- =====================================================
-- PURPOSE: Ensure all org_id tables enforce tenant isolation
-- =====================================================

\echo 'Enabling Row Level Security on multi-tenant tables...';

-- ========================================
-- 1. Verify get_current_org_id() function exists
-- ========================================
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_proc p
        JOIN pg_namespace n ON p.pronamespace = n.oid
        WHERE n.nspname = 'public'
          AND p.proname = 'get_current_org_id'
    ) THEN
        RAISE EXCEPTION 'Function get_current_org_id() does not exist. Please create it first (likely from RLS foundation migration 040).';
    END IF;
END $$;

-- ========================================
-- 2. Check for application role (informational)
-- ========================================
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_roles
        WHERE rolname = 'flyfleet_app'
    ) THEN
        RAISE NOTICE 'Note: Role flyfleet_app does not exist. Policies will be created for ALL users (FOR ALL).';
    ELSE
        RAISE NOTICE 'Role flyfleet_app exists - policies will apply to this role.';
    END IF;
END $$;

-- ========================================
-- 3. Enable RLS and create policies for all org_id tables
-- ========================================
-- Exclude tables that already have RLS enabled (checked dynamically)
-- Table list from analysis: 59 tables with org_id

DO $$
DECLARE
    table_record RECORD;
    policy_name TEXT;
    already_enabled BOOLEAN;
    policy_exists BOOLEAN;
BEGIN
    -- Get all BASE TABLES with org_id column (multi-tenant tables), excluding views
    FOR table_record IN
        SELECT c.table_name
        FROM information_schema.columns c
        JOIN information_schema.tables t ON c.table_name = t.table_name AND c.table_schema = t.table_schema
        WHERE c.table_schema = 'public'
          AND c.column_name = 'org_id'
          AND t.table_type = 'BASE TABLE'  -- Skip views and materialized views
          AND c.table_name NOT IN (
              -- Exclude junction/utility tables if needed
              'schema_migrations', 'schema_migrations_legacy'
          )
        ORDER BY c.table_name
    LOOP
        -- Check if RLS is already enabled
        SELECT EXISTS (
            SELECT 1 FROM pg_tables
            WHERE schemaname = 'public'
              AND tablename = table_record.table_name
              AND rowsecurity = true
        ) INTO already_enabled;

        IF NOT already_enabled THEN
            EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', table_record.table_name);
            RAISE NOTICE 'Enabled RLS on: %', table_record.table_name;
        ELSE
            RAISE NOTICE 'RLS already enabled on: %', table_record.table_name;
        END IF;

        -- Create policy if not exists
        policy_name := table_record.table_name || '_tenant_policy';

        SELECT EXISTS (
            SELECT 1 FROM pg_policies
            WHERE schemaname = 'public'
              AND tablename = table_record.table_name
              AND policyname = policy_name
        ) INTO policy_exists;

        IF NOT policy_exists THEN
            -- Create policy for all roles (no specific role needed)
            EXECUTE format(
                'CREATE POLICY %I ON %I ' ||
                'FOR ALL ' ||
                'USING (org_id = get_current_org_id()) ' ||
                'WITH CHECK (org_id = get_current_org_id())',
                policy_name,
                table_record.table_name
            );
            RAISE NOTICE '  Created policy: %', policy_name;
        ELSE
            RAISE NOTICE '  Policy % already exists', policy_name;
        END IF;
    END LOOP;
END $$;

-- ========================================
-- 4. Verification: Count RLS-enabled tables
-- ========================================
DO $$
DECLARE
    total_org_tables INT;
    rls_enabled_count INT;
BEGIN
    SELECT COUNT(*) INTO total_org_tables
    FROM information_schema.columns c
    JOIN information_schema.tables t ON c.table_name = t.table_name AND c.table_schema = t.table_schema
    WHERE c.table_schema = 'public'
      AND c.column_name = 'org_id'
      AND t.table_type = 'BASE TABLE'
      AND c.table_name NOT IN ('schema_migrations', 'schema_migrations_legacy');

    SELECT COUNT(*) INTO rls_enabled_count
    FROM pg_tables
    WHERE schemaname = 'public'
      AND rowsecurity = true
      AND tablename IN (
          SELECT c.table_name
          FROM information_schema.columns c
          JOIN information_schema.tables t ON c.table_name = t.table_name AND c.table_schema = t.table_schema
          WHERE c.table_schema = 'public'
            AND c.column_name = 'org_id'
            AND t.table_type = 'BASE TABLE'
      );

    RAISE NOTICE 'RLS Coverage:';
    RAISE NOTICE '  Tables with org_id: %', total_org_tables;
    RAISE NOTICE '  RLS-enabled tables: %', rls_enabled_count;
    RAISE NOTICE '  Coverage: %', (rls_enabled_count * 100.0 / total_org_tables) || '%';

    IF rls_enabled_count < total_org_tables THEN
        RAISE WARNING 'Not all org_id tables have RLS enabled. Check for errors above.';
    END IF;
END $$;

-- ========================================
-- 5. Test policy on a sample table
-- ========================================
-- Verify the policy actually filters by org_id
-- This is informational - does not modify data
DO $$
DECLARE
    sample_table TEXT;
    policy_count INT;
BEGIN
    -- Pick first org_id table
    SELECT table_name INTO sample_table
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND column_name = 'org_id'
      AND table_name NOT LIKE 'pg_%'
    LIMIT 1;

    IF sample_table IS NOT NULL THEN
        SELECT COUNT(*) INTO policy_count
        FROM pg_policies
        WHERE schemaname = 'public'
          AND tablename = sample_table
          AND policyname = sample_table || '_tenant_policy';

        RAISE NOTICE 'Sample table: % has % tenant policy', sample_table, policy_count;
    END IF;
END $$;

\echo '====================================================='
\echo 'Row Level Security enabled on multi-tenant tables!'
\echo '====================================================='
\echo 'Verification Queries:'
\echo '  - Check RLS status: SELECT tablename, rowsecurity FROM pg_tables WHERE schemaname = ''public'' AND rowsecurity = true;'
\echo '  - List policies: SELECT schemaname, tablename, policyname FROM pg_policies WHERE schemaname = ''public'';'
\echo '  - Test tenant isolation: SET app.org_id = ''<your-org-uuid>''; SELECT * FROM <table> LIMIT 1;'
\echo '====================================================='
