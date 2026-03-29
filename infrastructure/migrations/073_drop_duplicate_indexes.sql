-- Migration: 073_drop_duplicate_indexes.sql
-- Description: Remove duplicate and redundant indexes identified during analysis
-- Created: 2026-03-29

SET search_path TO public;

\echo 'Dropping duplicate and redundant indexes...';

-- ========================================
-- 1. DRIVERS TABLE: Duplicate org_id index
-- ========================================
-- Issue: Both idx_drivers_org and idx_drivers_org_id index the same column (org_id)
-- Keep: idx_drivers_org_id (more consistent with naming convention)
-- Drop: idx_drivers_org

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE schemaname = 'public'
          AND tablename = 'drivers'
          AND indexname = 'idx_drivers_org'
    ) THEN
        DROP INDEX idx_drivers_org;
        RAISE NOTICE 'Dropped duplicate index: idx_drivers_org (use idx_drivers_org_id instead)';
    ELSE
        RAISE NOTICE 'Index idx_drivers_org not found (already dropped or never existed)';
    END IF;
END $$;

-- ========================================
-- 2. Check for other potential duplicates
-- ========================================
-- Log any tables that have multiple indexes on the same column(s)

DO $$
DECLARE
    dup_record RECORD;
BEGIN
    FOR dup_record IN
        SELECT
            t.relname AS table_name,
            i.relname AS index_name,
            array_agg(a.attname ORDER BY array_position(i.indkey, a.attnum)) AS indexed_columns
        FROM pg_index idx
        JOIN pg_class i ON i.oid = idx.indexrelid
        JOIN pg_class t ON t.oid = idx.indrelid
        JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(idx.indkey)
        WHERE t.relkind = 'r' -- ordinary tables only
          AND t.relnamespace = 'public'::regnamespace
          AND array_length(idx.indkey, 1) = 1 -- single-column indexes only
        GROUP BY t.relname, i.relname
        HAVING COUNT(*) > 1
    LOOP
        -- This query finds tables with multiple single-column indexes on different columns
        -- But we want to find duplicates on the SAME column
        -- Let's do a different approach
    END LOOP;
END $$;

-- Better duplicate detection: find same column indexed multiple times
DO $$
DECLARE
    col_record RECORD;
    index_count INT;
BEGIN
    FOR col_record IN
        SELECT
            t.relname AS table_name,
            a.attname AS column_name,
            COUNT(*) AS index_count,
            string_agg(i.relname, ', ' ORDER BY i.relname) AS index_names
        FROM pg_index idx
        JOIN pg_class i ON i.oid = idx.indexrelid
        JOIN pg_class t ON t.oid = idx.indrelid
        JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(idx.indkey)
        WHERE t.relnamespace = 'public'::regnamespace
          AND array_length(idx.indkey, 1) = 1 -- single-column indexes
          AND a.attnum > 0
        GROUP BY t.relname, a.attname
        HAVING COUNT(*) > 1
        ORDER BY t.relname, a.attname
    LOOP
        RAISE WARNING 'Duplicate index on %.%: % indexes - %',
            col_record.table_name,
            col_record.column_name,
            col_record.index_count,
            col_record.index_names;
    END LOOP;
END $$;

-- ========================================
-- 3. Manually review and drop known redundant indexes
-- ========================================
-- Add additional DROP INDEX statements here as you identify them

-- Example pattern:
-- DROP INDEX IF EXISTS idx_table_column_old;

\echo '====================================================='
\echo 'Duplicate index cleanup complete!'
\echo '====================================================='
\echo 'Review warnings above for any additional duplicates'
\echo '====================================================='
