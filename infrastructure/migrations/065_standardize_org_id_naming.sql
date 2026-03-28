-- Migration: 067_standardize_org_id_naming.sql
-- Description: Standardize organization_id column name to org_id across all tables
-- Created: 2026-03-28

SET search_path TO public;

-- ========================================
-- 1. Rename organization_id → org_id in trips table
-- ========================================
-- First, we must drop any constraints/indexes that reference the old column name

-- Drop indexes that reference organization_id (will recreate after rename)
DROP INDEX IF EXISTS idx_trips_org_id;
DROP INDEX IF EXISTS idx_trips_org_id_status;  -- if exists from previous versions

-- Rename the column
ALTER TABLE trips RENAME COLUMN organization_id TO org_id;

-- Recreate indexes with org_id naming (consistent with other tables)
CREATE INDEX idx_trips_org_id ON trips(org_id);
CREATE INDEX idx_trips_org_status ON trips(org_id, status);

-- Update any comments
COMMENT ON COLUMN trips.org_id IS 'Organization UUID - foreign key to organizations(id)';

-- ========================================
-- 2. Rename organization_id → org_id in work_orders table
-- ========================================

-- First, check if work_orders has organization_id column
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'work_orders' AND column_name = 'organization_id'
    ) THEN
        -- Drop indexes on old column
        DROP INDEX IF EXISTS idx_work_orders_org;
        DROP INDEX IF EXISTS idx_work_orders_org_status;
        DROP INDEX IF EXISTS idx_work_orders_org_priority;

        -- Rename column
        ALTER TABLE work_orders RENAME COLUMN organization_id TO org_id;

        -- Recreate indexes with consistent naming
        CREATE INDEX idx_work_orders_org ON work_orders (org_id);
        CREATE INDEX idx_work_orders_org_status ON work_orders (org_id, status);
        CREATE INDEX idx_work_orders_org_priority ON work_orders (org_id, priority);

        COMMENT ON COLUMN work_orders.org_id IS 'Organization UUID - foreign key to organizations(id)';

        RAISE NOTICE 'Renamed organization_id to org_id in work_orders';
    END IF;
END $$;

-- ========================================
-- 3. Verification: Ensure no tables still use organization_id
-- ========================================
DO $$
DECLARE
    legacy_count INT;
BEGIN
    SELECT COUNT(*) INTO legacy_count
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND column_name = 'organization_id'
      AND table_name NOT IN ('pg_statistic', 'pg_class');  -- exclude system tables

    IF legacy_count > 0 THEN
        RAISE WARNING 'Found % tables still using organization_id column name. Consider manual review.', legacy_count;
    ELSE
        RAISE NOTICE 'All tables now use org_id (no organization_id columns found)';
    END IF;
END $$;

-- ========================================
-- 4. Apply foreign key constraints to org_id columns (if missing)
-- ========================================

-- For trips: ensure FK exists (should already be added in migration 064)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_name = 'trips' AND constraint_name = 'fk_trips_organization'
    ) THEN
        ALTER TABLE trips
        ADD CONSTRAINT fk_trips_organization
        FOREIGN KEY (org_id) REFERENCES organizations(id)
        ON DELETE CASCADE;
        RAISE NOTICE 'Added FK fk_trips_organization to trips.org_id';
    END IF;
END $$;

-- For work_orders: ensure FK exists
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'work_orders' AND column_name = 'org_id'
    ) THEN
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.table_constraints
            WHERE table_name = 'work_orders' AND constraint_name = 'fk_work_orders_organization'
        ) THEN
            ALTER TABLE work_orders
            ADD CONSTRAINT fk_work_orders_organization
            FOREIGN KEY (org_id) REFERENCES organizations(id)
            ON DELETE CASCADE;
            RAISE NOTICE 'Added FK fk_work_orders_organization to work_orders.org_id';
        END IF;
    END IF;
END $$;

COMMENT ON CONSTRAINT fk_trips_organization ON trips IS 'Ensures trip belongs to valid organization';
COMMENT ON CONSTRAINT fk_work_orders_organization ON work_orders IS 'Ensures work order belongs to valid organization';
