-- Migration: 077_finalize_org_id_standardization.sql
-- Description: Final cleanup of organization_id → org_id for remaining base tables
-- Created: 2026-03-29
-- Critical: Must be run after all previous migrations

SET search_path TO public;

\echo 'Finalizing organization_id → org_id standardization...';

-- ========================================
-- 1. Identify tables that have BOTH columns (need data migration)
-- ========================================
DO $$
DECLARE
    tbl RECORD;
BEGIN
    FOR tbl IN
        SELECT table_name
        FROM information_schema.columns c
        WHERE c.table_schema = 'public'
          AND c.column_name = 'organization_id'
          AND EXISTS (
              SELECT 1 FROM information_schema.columns c2
              WHERE c2.table_schema = 'public'
                AND c2.table_name = c.table_name
                AND c2.column_name = 'org_id'
          )
    LOOP
        RAISE NOTICE 'Table % has BOTH org_id and organization_id - data migration needed', tbl.table_name;
    END LOOP;
END $$;

-- ========================================
-- 2. For tables with BOTH columns: Copy data from organization_id to org_id
-- ========================================
DO $$
DECLARE
    tbl RECORD;
    rows_updated INT;
BEGIN
    FOR tbl IN
        SELECT table_name
        FROM information_schema.columns c
        WHERE c.table_schema = 'public'
          AND c.column_name = 'organization_id'
          AND EXISTS (
              SELECT 1 FROM information_schema.columns c2
              WHERE c2.table_schema = 'public'
                AND c2.table_name = c.table_name
                AND c2.column_name = 'org_id'
          )
    LOOP
        RAISE NOTICE 'Migrating data in %: organization_id → org_id', tbl.table_name;

        -- Update rows where org_id is NULL but organization_id has value
        EXECUTE format(
            'UPDATE %I SET org_id = organization_id WHERE org_id IS NULL AND organization_id IS NOT NULL',
            tbl.table_name
        );
        GET DIAGNOSTICS rows_updated = ROW_COUNT;
        RAISE NOTICE '  Updated % rows', rows_updated;

        -- Verification: check for mismatches
        IF EXISTS (
            EXECUTE format(
                'SELECT 1 FROM %I WHERE org_id != organization_id AND organization_id IS NOT NULL LIMIT 1',
                tbl.table_name
            )
        ) THEN
            RAISE WARNING '  Table % still has mismatched values!', tbl.table_name;
        ELSE
            RAISE NOTICE '  Data migration successful for %', tbl.table_name;
        END IF;
    END LOOP;
END $$;

-- ========================================
-- 3. Drop organization_id column from all base tables (except views)
-- ========================================
DO $$
DECLARE
    tbl RECORD;
    colname TEXT;
BEGIN
    FOR tbl IN
        SELECT c.table_name
        FROM information_schema.columns c
        JOIN information_schema.tables t ON c.table_name = t.table_name AND c.table_schema = t.table_schema
        WHERE c.table_schema = 'public'
          AND c.column_name = 'organization_id'
          AND t.table_type = 'BASE TABLE'
          AND c.table_name NOT IN ('pg_statistic', 'pg_class')
    LOOP
        BEGIN
            -- Check if org_id also exists (should have been renamed already)
            IF EXISTS (
                SELECT 1 FROM information_schema.columns
                WHERE table_schema = 'public'
                  AND table_name = tbl.table_name
                  AND column_name = 'org_id'
            ) THEN
                -- org_id exists, safe to drop organization_id
                RAISE NOTICE 'Dropping organization_id from %', tbl.table_name;
                EXECUTE format('ALTER TABLE %I DROP COLUMN IF EXISTS organization_id', tbl.table_name);
            ELSE
                -- org_id doesn't exist, rename instead
                RAISE NOTICE 'Renaming organization_id → org_id on % (org_id missing)', tbl.table_name;
                EXECUTE format('ALTER TABLE %I RENAME COLUMN organization_id TO org_id', tbl.table_name);
            END IF;
        EXCEPTION
            WHEN others THEN
                RAISE WARNING 'Failed to drop organization_id from %: %', tbl.table_name, SQLERRM;
        END;
    END LOOP;
END $$;

-- ========================================
-- 4. Verify all base tables now use org_id (no organization_id)
-- ========================================
DO $$
DECLARE
    legacy_count INT;
    tbl_rec RECORD;
BEGIN
    SELECT COUNT(*) INTO legacy_count
    FROM information_schema.columns c
    JOIN information_schema.tables t ON c.table_name = t.table_name AND c.table_schema = t.table_schema
    WHERE c.table_schema = 'public'
      AND c.column_name = 'organization_id'
      AND t.table_type = 'BASE TABLE'
      AND c.table_name NOT IN ('pg_statistic', 'pg_class');

    IF legacy_count > 0 THEN
        RAISE WARNING 'Found % base tables still using organization_id:', legacy_count;
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
        RAISE NOTICE 'SUCCESS: All base tables now use org_id';
    END IF;
END $$;

-- ========================================
-- 5. Add/verify FK constraints on org_id for tables that might be missing them
-- ========================================
DO $$
DECLARE
    tbl RECORD;
    fk_name TEXT;
BEGIN
    FOR tbl IN
        SELECT DISTINCT table_name
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND column_name = 'org_id'
        ORDER BY table_name
    LOOP
        fk_name := 'fk_' || tbl.table_name || '_organization';

        IF NOT EXISTS (
            SELECT 1 FROM information_schema.table_constraints
            WHERE table_schema = 'public'
              AND table_name = tbl.table_name
              AND constraint_name = fk_name
        ) THEN
            BEGIN
                EXECUTE format(
                    'ALTER TABLE %I ADD CONSTRAINT %I FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE',
                    tbl.table_name,
                    fk_name
                );
                RAISE NOTICE 'Added FK: % on %', fk_name, tbl.table_name;
            EXCEPTION
                WHEN others THEN
                    -- Might fail due to data issues (NULLs, orphans). Log but continue.
                    RAISE WARNING 'Could not add FK % to %: %', fk_name, tbl.table_name, SQLERRM;
            END;
        END IF;
    END LOOP;
END $$;

\echo '====================================================='
\echo 'Organization ID standardization finalized!'
\echo '====================================================='
\echo 'Verification:'
\echo '  SELECT table_name FROM information_schema.columns WHERE column_name = ''organization_id'' AND table_schema = ''public'';'
\echo '  (Should return 0 rows for base tables)'
\echo '====================================================='
