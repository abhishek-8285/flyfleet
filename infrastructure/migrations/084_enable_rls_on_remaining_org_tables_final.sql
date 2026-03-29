-- Migration: 084_enable_rls_on_remaining_org_tables_final.sql
-- Description: Enable RLS on remaining org_id tables (excludes views, handles errors)
-- Created: 2026-03-29
-- Dependencies: 083_enable_rls_on_remaining_org_tables.sql

SET search_path TO public;

\echo 'Enabling RLS on final remaining org_id tables (excluding views)...';

-- Create a temporary table of tables that need RLS
CREATE TEMP TABLE temp_rls_tables AS
SELECT c.table_name
FROM information_schema.columns c
LEFT JOIN pg_tables p ON c.table_name = p.tablename AND p.schemaname = 'public'
WHERE c.column_name = 'org_id'
  AND c.table_schema = 'public'
  AND (p.rowsecurity IS NULL OR p.rowsecurity = false)
  AND NOT EXISTS (
      SELECT 1 FROM information_schema.views v
      WHERE v.table_name = c.table_name AND v.table_schema = 'public'
  );

-- Count how many need fixing
SELECT COUNT(*) as tables_to_fix FROM temp_rls_tables;

-- Process each table using a DO block with exception handling
DO $$
DECLARE
    r RECORD;
    policy_name TEXT;
BEGIN
    FOR r IN SELECT table_name FROM temp_rls_tables ORDER BY table_name
    LOOP
        BEGIN
            EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', r.table_name);
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
        EXCEPTION
            WHEN OTHERS THEN
                RAISE WARNING 'Failed on %: %', r.table_name, SQLERRM;
        END;
    END LOOP;
END $$;

DROP TABLE temp_rls_tables;

\echo '=====================================================';
\echo 'RLS enablement complete on all applicable tables!';
\echo '=====================================================';

-- Final verification: count RLS-enabled org_id tables
SELECT COUNT(*) as rls_enabled_org_tables
FROM pg_tables p
JOIN information_schema.columns c ON p.tablename = c.table_name AND c.table_schema = 'public'
WHERE p.schemaname = 'public'
  AND c.column_name = 'org_id'
  AND p.rowsecurity = true;
