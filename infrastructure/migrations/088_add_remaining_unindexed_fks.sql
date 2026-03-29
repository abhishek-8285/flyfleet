-- Migration: 088_add_remaining_unindexed_fks.sql
-- Description: Add indexes for all remaining unindexed foreign key constraints
-- Created: 2026-03-29
-- Dependencies: 087_finalize_fuel_logs_standardization.sql

SET search_path TO public;

\echo 'Adding indexes for remaining unindexed foreign keys...';

DO $$
DECLARE
    r RECORD;
    idx_name TEXT;
    total INT;
BEGIN
    -- Count total unindexed FKs
    SELECT COUNT(DISTINCT tc.table_name || '.' || kcu.column_name) INTO total
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
    WHERE tc.constraint_type = 'FOREIGN KEY'
      AND tc.table_schema = 'public'
      AND NOT EXISTS (
          SELECT 1 FROM pg_indexes pi
          WHERE pi.schemaname = 'public'
            AND pi.tablename = tc.table_name
            AND pi.indexdef LIKE '%' || kcu.column_name || '%'
      )
      AND tc.table_name NOT IN ('telemetry_events')
      AND tc.table_name NOT LIKE 'telemetry_events_p%';

    RAISE NOTICE 'Total unindexed FKs to index: %', total;

    -- Create indexes
    FOR r IN
        SELECT DISTINCT tc.table_name, kcu.column_name
        FROM information_schema.table_constraints tc
        JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
        WHERE tc.constraint_type = 'FOREIGN KEY'
          AND tc.table_schema = 'public'
          AND NOT EXISTS (
              SELECT 1 FROM pg_indexes pi
              WHERE pi.schemaname = 'public'
                AND pi.tablename = tc.table_name
                AND pi.indexdef LIKE '%' || kcu.column_name || '%'
          )
          AND tc.table_name NOT IN ('telemetry_events')
          AND tc.table_name NOT LIKE 'telemetry_events_p%'
        ORDER BY tc.table_name, kcu.column_name
    LOOP
        idx_name := 'idx_' || r.table_name || '_' || r.column_name;

        IF NOT EXISTS (
            SELECT 1 FROM pg_indexes
            WHERE schemaname = 'public' AND tablename = r.table_name AND indexname = idx_name
        ) THEN
            BEGIN
                EXECUTE format('CREATE INDEX %I ON %I (%I)', idx_name, r.table_name, r.column_name);
                RAISE NOTICE 'Created index: % on %.%', idx_name, r.table_name, r.column_name;
            EXCEPTION
                WHEN OTHERS THEN
                    RAISE WARNING 'Failed to create index % on %: %', idx_name, r.table_name, SQLERRM;
            END;
        END IF;
    END LOOP;
END $$;

\echo '=====================================================';
\echo 'Remaining unindexed foreign keys have been indexed!';
\echo '=====================================================';

-- Final verification
WITH final_unindexed AS (
  SELECT DISTINCT tc.table_name, kcu.column_name
  FROM information_schema.table_constraints tc
  JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
  WHERE tc.constraint_type = 'FOREIGN KEY'
    AND tc.table_schema = 'public'
    AND NOT EXISTS (
        SELECT 1 FROM pg_indexes pi
        WHERE pi.schemaname = 'public'
          AND pi.tablename = tc.table_name
          AND pi.indexdef LIKE '%' || kcu.column_name || '%'
    )
    AND tc.table_name NOT IN ('telemetry_events')
    AND tc.table_name NOT LIKE 'telemetry_events_p%'
)
SELECT COUNT(*) as remaining_unindexed FROM final_unindexed;
\echo 'Remaining unindexed FKs (should be 0): ';
SELECT COUNT(*) FROM final_unindexed;
