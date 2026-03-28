-- Migration: 066_normalize_updated_at_triggers.sql
-- Description: Ensure all tables with updated_at column have automatic trigger
-- Created: 2026-03-28

SET search_path TO public;

-- ========================================
-- 1. Create or replace the updated_at trigger function (if not exists)
-- ========================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ========================================
-- 2. Add triggers to tables with updated_at but no trigger
-- ========================================

-- List of tables that SHOULD have the trigger based on schema review
-- We'll add triggers only if they don't already exist

DO $$
DECLARE
    table_record RECORD;
    expected_trigger_name TEXT;
BEGIN
    -- Find all tables with an updated_at column that are user tables (not system)
    FOR table_record IN
        SELECT DISTINCT table_name
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND column_name = 'updated_at'
          AND table_name NOT LIKE 'pg_%'
          AND table_name NOT LIKE 'sql_%'
    LOOP
        expected_trigger_name := 'update_' || table_record.table_name || '_updated_at';

        -- Check if trigger already exists for this table
        IF NOT EXISTS (
            SELECT 1
            FROM information_schema.triggers
            WHERE trigger_schema = 'public'
              AND event_object_table = table_record.table_name
              AND trigger_name = expected_trigger_name
        ) THEN
            -- Build and execute CREATE TRIGGER statement dynamically
            EXECUTE format(
                'CREATE TRIGGER %I ' ||
                'BEFORE UPDATE ON %I ' ||
                'FOR EACH ROW ' ||
                'EXECUTE FUNCTION update_updated_at_column()',
                expected_trigger_name,
                table_record.table_name
            );

            RAISE NOTICE 'Added trigger % to table %', expected_trigger_name, table_record.table_name;
        END IF;
    END LOOP;
END $$;

-- ========================================
-- 3. Verify key tables have the trigger
-- ========================================

-- List of critical tables that MUST have the trigger
-- These are tables we know have updated_at columns from migrations
CREATE TEMP TABLE critical_tables_with_updated_at AS
SELECT unnest(ARRAY[
    'vehicles',
    'drivers',
    'organizations',
    'users',
    'support_tickets',
    'incident_comments',
    'billing_customer_balances',
    'billing_ledger_entries',
    'billing_revenue_schedules',
    'org_support_metrics',
    'user_sessions',
    'support_audit_log'
]) AS table_name;

-- Check that all critical tables now have triggers
DO $$
DECLARE
    missing_count INT;
BEGIN
    SELECT COUNT(*) INTO missing_count
    FROM critical_tables_with_updated_at ct
    WHERE NOT EXISTS (
        SELECT 1
        FROM information_schema.triggers tr
        WHERE tr.event_object_table = ct.table_name
          AND tr.trigger_schema = 'public'
          AND tr.trigger_name = 'update_' || ct.table_name || '_updated_at'
    );

    IF missing_count > 0 THEN
        RAISE WARNING 'Critical tables missing updated_at triggers: %', missing_count;
    ELSE
        RAISE NOTICE 'All % critical tables have updated_at triggers', (SELECT COUNT(*) FROM critical_tables_with_updated_at);
    END IF;
END $$;

DROP TABLE critical_tables_with_updated_at;

-- ========================================
-- 4. Comment on function (documentation)
-- ========================================
COMMENT ON FUNCTION update_updated_at_column() IS 'Automatically sets updated_at = NOW() on row updates. Applied to all tables with updated_at column via trigger.';
