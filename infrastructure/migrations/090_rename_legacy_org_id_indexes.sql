-- Migration: 090_rename_legacy_org_id_indexes.sql
-- Description: Rename indexes with 'organization_id' in name to standard 'org_id' naming
-- Created: 2026-03-29
-- Purpose: Complete naming consistency (indexes contain org_id but names say organization_id)
-- Note: These indexes are functionally correct (they index org_id column), just legacy names

SET search_path TO public;

\echo 'Renaming legacy indexes to standard org_id naming...';

-- List of indexes to rename: old_name -> new_name
-- All these indexes already index the org_id column, just have legacy naming
-- Format: table_name, old_index_name, new_index_name

DO $$
DECLARE
    rename_map RECORD;
    mappings JSONB := '[
        {"table": "driver_fuel_limits", "old": "driver_fuel_limits_organization_id_driver_id_key", "new": "idx_driver_fuel_limits_org_id_driver_id"},
        {"table": "driver_verifications", "old": "idx_driver_organization_id", "new": "idx_driver_verifications_org_id"},
        {"table": "fastag_transactions", "old": "idx_fastag_organization_id", "new": "idx_fastag_transactions_org_id"},
        {"table": "gst_compliances", "old": "idx_gst_organization_id", "new": "idx_gst_compliances_org_id"},
        {"table": "integration_configs", "old": "idx_config_organization_id", "new": "idx_integration_configs_org_id"},
        {"table": "inventory_parts", "old": "inventory_parts_organization_id_part_number_key", "new": "uq_inventory_parts_org_id_part_number"},
        {"table": "local_permits", "old": "idx_permit_organization_id", "new": "idx_local_permits_org_id"},
        {"table": "org_budgets", "old": "org_budgets_organization_id_category_period_start_key", "new": "uq_org_budgets_org_id_category_period_start"},
        {"table": "organization_users", "old": "organization_users_organization_id_email_key", "new": "uq_organization_users_org_id_email"},
        {"table": "trip_settings", "old": "trip_settings_organization_id_key", "new": "uq_trip_settings_org_id"},
        {"table": "vehicle_verifications", "old": "idx_vehicle_organization_id", "new": "idx_vehicle_verifications_org_id"},
        {"table": "video_recordings", "old": "idx_video_recordings_organization_id", "new": "idx_video_recordings_org_id"},
        {"table": "video_streams", "old": "idx_video_streams_organization_id", "new": "idx_video_streams_org_id"}
    ]';
BEGIN
    FOR rename_map IN SELECT * FROM jsonb_to_recordset(mappings) AS x("table" text, "old" text, "new" text)
    LOOP
        BEGIN
            -- Check if old index exists
            IF EXISTS (
                SELECT 1 FROM pg_indexes
                WHERE schemaname = 'public' AND tablename = rename_map.table AND indexname = rename_map.old
            ) THEN
                -- Check if new index name already exists (skip if it does)
                IF NOT EXISTS (
                    SELECT 1 FROM pg_indexes
                    WHERE schemaname = 'public' AND tablename = rename_map.table AND indexname = rename_map.new
                ) THEN
                    EXECUTE format('ALTER INDEX %I RENAME TO %I', rename_map.old, rename_map.new);
                    RAISE NOTICE 'Renamed index on %: % -> %', rename_map.table, rename_map.old, rename_map.new;
                ELSE
                    RAISE NOTICE 'Index % already exists on %, dropping old: %', rename_map.new, rename_map.table, rename_map.old;
                    EXECUTE format('DROP INDEX %I', rename_map.old);
                END IF;
            END IF;
        EXCEPTION
            WHEN OTHERS THEN
                RAISE WARNING 'Failed to rename index % on %: %', rename_map.old, rename_map.table, SQLERRM;
        END;
    END LOOP;
END $$;

\echo '=====================================================';
\echo 'Legacy index names standardized!';
\echo '=====================================================';

-- Verify remaining indexes with old naming
SELECT COUNT(*) as remaining_legacy_named_indexes
FROM pg_indexes
WHERE schemaname = 'public'
  AND indexname LIKE '%organization_id%'
  AND indexdef NOT LIKE '%org_id%';
\echo 'Remaining legacy-named indexes (should be 0): ';
SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public' AND indexname LIKE '%organization_id%' AND indexdef NOT LIKE '%org_id%';
