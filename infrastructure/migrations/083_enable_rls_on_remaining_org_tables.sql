-- Migration: 083_enable_rls_on_remaining_org_tables.sql
-- Description: Enable RLS on all org_id tables that were missed in migration 074
-- Created: 2026-03-29
-- Purpose: Achieve 100% RLS coverage on all multi-tenant tables

SET search_path TO public;

\echo 'Enabling RLS on remaining org_id tables...';

-- Count how many need RLS
SELECT COUNT(*) as tables_to_fix
FROM information_schema.columns c
LEFT JOIN pg_tables p ON c.table_name = p.tablename AND p.schemaname = 'public'
WHERE c.column_name = 'org_id'
  AND c.table_schema = 'public'
  AND (p.rowsecurity IS NULL OR p.rowsecurity = false);

-- Enable RLS on all tables with org_id that don't have it
DO $$
DECLARE
    r RECORD;
    policy_name TEXT;
BEGIN
    FOR r IN
        SELECT c.table_name
        FROM information_schema.columns c
        LEFT JOIN pg_tables p ON c.table_name = p.tablename AND p.schemaname = 'public'
        WHERE c.column_name = 'org_id'
          AND c.table_schema = 'public'
          AND (p.rowsecurity IS NULL OR p.rowsecurity = false)
        ORDER BY c.table_name
    LOOP
        -- Enable RLS
        EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', r.table_name);

        -- Create policy
        policy_name := r.table_name || '_tenant_policy';
        IF NOT EXISTS (
            SELECT 1 FROM pg_policies
            WHERE tablename = r.table_name AND policyname = policy_name
        ) THEN
            EXECUTE format(
                'CREATE POLICY %I ON %I FOR ALL USING (org_id = get_current_org_id()) WITH CHECK (org_id = get_current_org_id())',
                policy_name, r.table_name
            );
        END IF;

        RAISE NOTICE 'Enabled RLS on %', r.table_name;
    END LOOP;
END $$;

\echo '=====================================================';
\echo 'RLS enabled on all remaining org_id tables!';
\echo '=====================================================';

-- Verify RLS coverage
SELECT COUNT(*) as rls_tables
FROM pg_tables
WHERE rowsecurity = true AND schemaname = 'public'
  AND tablename IN (SELECT table_name FROM information_schema.columns WHERE column_name = 'org_id' AND table_schema = 'public');

\echo 'Total org_id tables with RLS: ';
SELECT COUNT(*) FROM pg_tables WHERE rowsecurity = true AND schemaname = 'public'
  AND tablename IN (SELECT table_name FROM information_schema.columns WHERE column_name = 'org_id' AND table_schema = 'public');
