-- Migration: 071_standardize_organization_id_naming.sql
-- Description: Rename organization_id to org_id for consistency across all tables
-- Created: 2026-03-29
-- Builds on: Migration 067 (standardized trips and work_orders)
-- Extends to: 29 remaining tables using organization_id

SET search_path TO public;

-- =====================================================
-- PURPOSE: Ensure consistent org_id naming convention
-- =====================================================

\echo 'Standardizing organization_id to org_id...'

-- ========================================
-- Helper function to safely rename column and recreate indexes
-- ========================================
DO $$
DECLARE
    table_record RECORD;
    old_column TEXT := 'organization_id';
    new_column TEXT := 'org_id';
    index_record RECORD;
    idx_name_old TEXT;
    idx_name_new TEXT;
BEGIN
    -- Find all BASE TABLES (not views) with organization_id column
    FOR table_record IN
        SELECT c.table_name
        FROM information_schema.columns c
        JOIN information_schema.tables t ON c.table_name = t.table_name AND c.table_schema = t.table_schema
        WHERE c.table_schema = 'public'
          AND c.column_name = old_column
          AND t.table_type = 'BASE TABLE'  -- Skip views and materialized views
          AND c.table_name NOT IN ('pg_statistic', 'pg_class')
    LOOP
        RAISE NOTICE 'Processing table: %', table_record.table_name;

        -- 1. Drop indexes that reference the old column (will recreate after rename)
        FOR index_record IN
            SELECT indexname
            FROM pg_indexes
            WHERE schemaname = 'public'
              AND tablename = table_record.table_name
              AND indexdef LIKE '%' || old_column || '%'
        LOOP
            idx_name_old := index_record.indexname;
            -- Build new index name by replacing organization_id with org_id
            idx_name_new := REPLACE(idx_name_old, old_column, new_column);
            -- Also fix any double underscores from the replacement
            idx_name_new := REGEXP_REPLACE(idx_name_new, '_+', '_', 'g);

            EXECUTE format('DROP INDEX IF EXISTS %I', idx_name_old);
            RAISE NOTICE '  Dropped index: %', idx_name_old;
        END LOOP;

        -- 2. Rename the column
        BEGIN
            EXECUTE format('ALTER TABLE %I RENAME COLUMN %I TO %I',
                           table_record.table_name, old_column, new_column);
            RAISE NOTICE '  Renamed column: % -> %', old_column, new_column;
        EXCEPTION
            WHEN others THEN
                RAISE WARNING '  Failed to rename column on %: %', table_record.table_name, SQLERRM;
        END;

        -- 3. Recreate indexes that were on the new column (based on pattern)
        -- Basic org index
        BEGIN
            EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_%s ON %I(%I)',
                           table_record.table_name, new_column, table_record.table_name, new_column);
        EXCEPTION
            WHEN others THEN
                RAISE WARNING '  Failed to create basic index on %.%: %', table_record.table_name, new_column, SQLERRM;
        END;

        -- Composite with status (if status column exists)
        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public'
              AND table_name = table_record.table_name
              AND column_name = 'status'
        ) THEN
            BEGIN
                EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_%s_status ON %I(%I, status)',
                               table_record.table_name, new_column, table_record.table_name, new_column);
            EXCEPTION
                WHEN others THEN
                    RAISE WARNING '  Failed to create status composite index on %: %', table_record.table_name, SQLERRM;
            END;
        END IF;

        -- Composite with created_at (for time-based queries)
        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public'
              AND table_name = table_record.table_name
              AND column_name = 'created_at'
        ) THEN
            BEGIN
                EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%s_%s_created ON %I(%I, created_at DESC)',
                               table_record.table_name, new_column, table_record.table_name, new_column);
            EXCEPTION
                WHEN others THEN
                    RAISE WARNING '  Failed to create created_at composite index on %: %', table_record.table_name, SQLERRM;
            END;
        END IF;

        RAISE NOTICE '  Recreated indexes for %', table_record.table_name;

        -- 4. Update column comment
        BEGIN
            EXECUTE format('COMMENT ON COLUMN %I.%I IS ''Organization UUID - foreign key to organizations(id)''',
                           table_record.table_name, new_column);
        EXCEPTION
            WHEN others THEN
                -- Comment failure is not critical
                NULL;
        END;
    END LOOP;
END $$;

-- ========================================
-- 2. Add foreign key constraints if missing (only on base tables)
-- ========================================
DO $$
DECLARE
    table_record RECORD;
BEGIN
    FOR table_record IN
        SELECT c.table_name
        FROM information_schema.columns c
        JOIN information_schema.tables t ON c.table_name = t.table_name AND c.table_schema = t.table_schema
        WHERE c.table_schema = 'public'
          AND c.column_name = 'org_id'
          AND t.table_type = 'BASE TABLE'
          AND c.table_name NOT IN ('pg_statistic', 'pg_class')
    LOOP
        -- Check if FK exists
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.table_constraints
            WHERE table_name = table_record.table_name
              AND constraint_name = 'fk_' || table_record.table_name || '_organization'
        ) THEN
            BEGIN
                EXECUTE format(
                    'ALTER TABLE %I ADD CONSTRAINT fk_%s_organization FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE',
                    table_record.table_name,
                    table_record.table_name
                );
                EXECUTE format(
                    'COMMENT ON CONSTRAINT fk_%s_organization ON %I IS ''Ensures %s belongs to valid organization''',
                    table_record.table_name,
                    table_record.table_name,
                    table_record.table_name
                );
                RAISE NOTICE 'Added FK: fk_%s_organization', table_record.table_name;
            EXCEPTION
                WHEN others THEN
                    RAISE WARNING 'Could not add FK to %: %', table_record.table_name, SQLERRM;
            END;
        END IF;
    END LOOP;
END $$;

-- ========================================
-- 3. Verification: Ensure no BASE TABLES still use organization_id
-- ========================================
DO $$
DECLARE
    legacy_count INT;
    tbl_rec RECORD;  -- Renamed from legacy_count to avoid conflict
BEGIN
    SELECT COUNT(*) INTO legacy_count
    FROM information_schema.columns c
    JOIN information_schema.tables t ON c.table_name = t.table_name AND c.table_schema = t.table_schema
    WHERE c.table_schema = 'public'
      AND c.column_name = 'organization_id'
      AND t.table_type = 'BASE TABLE'
      AND c.table_name NOT IN ('pg_statistic', 'pg_class');

    IF legacy_count > 0 THEN
        RAISE WARNING 'Found % BASE TABLES still using organization_id column name. Manual review needed:', legacy_count;
        -- List them
        FOR tbl_rec IN
            SELECT c.table_name
            FROM information_schema.columns c
            JOIN information_schema.tables t ON c.table_name = t.table_name AND c.table_schema = t.table_schema
            WHERE c.table_schema = 'public'
              AND c.column_name = 'organization_id'
              AND t.table_type = 'BASE TABLE'
              AND c.table_name NOT IN ('pg_statistic', 'pg_class')
            ORDER BY c.table_name
        LOOP
            RAISE WARNING '  - %', tbl_rec.table_name;
        END LOOP;
    ELSE
        RAISE NOTICE 'SUCCESS: All BASE TABLES now use org_id (no organization_id columns found)';
    END IF;

    -- Also report any views that still use organization_id (expected)
    DECLARE
        view_count INT;
    BEGIN
        SELECT COUNT(*) INTO view_count
        FROM information_schema.columns c
        JOIN information_schema.views v ON c.table_name = v.table_name AND c.table_schema = v.table_schema
        WHERE c.table_schema = 'public'
          AND c.column_name = 'organization_id';

        IF view_count > 0 THEN
            RAISE NOTICE 'Note: % VIEWS still use organization_id (expected - columns cannot be renamed on views)', view_count;
        END IF;
    END;
END $$;

-- ========================================
-- 4. Comment on schema (documentation)
-- ========================================
COMMENT ON COLUMN organizations.id IS 'UUID primary key - referenced by org_id throughout the schema';

\echo '====================================================='
\echo 'organization_id → org_id standardization complete!'
\echo '====================================================='
\echo 'Next steps:'
\echo '  1. Verify: SELECT table_name, column_name FROM information_schema.columns WHERE column_name = ''organization_id'' OR column_name = ''org_id'' ORDER BY table_name;'
\echo '  2. Check indexes: SELECT tablename, indexname FROM pg_indexes WHERE indexname LIKE ''%organization_id%'' OR indexname LIKE ''%org_id%'';'
\echo '  3. Run application tests to ensure no broken queries'
\echo '  4. If any views use organization_id in their definition, consider updating view definitions to use org_id'
