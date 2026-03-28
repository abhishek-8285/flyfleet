-- Migration: 064_fix_trips_foreign_keys.sql
-- Description: Add missing foreign key constraints to trips table
-- Created: 2026-03-28

SET search_path TO public;

-- ========================================
-- Add missing foreign key: driver_id → drivers.id
-- ========================================
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_name = 'trips' AND constraint_name = 'fk_trips_driver'
    ) THEN
        ALTER TABLE trips
        ADD CONSTRAINT fk_trips_driver
        FOREIGN KEY (driver_id) REFERENCES drivers(id)
        ON DELETE SET NULL;

        RAISE NOTICE 'Added constraint fk_trips_driver';
    ELSE
        RAISE NOTICE 'Constraint fk_trips_driver already exists, skipping';
    END IF;
END $$;

-- ========================================
-- Add missing foreign key: org_id → organizations.id
-- Supports both org_id (post-067) and organization_id (pre-067)
-- ========================================
DO $$
BEGIN
    -- Check which column exists
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'trips' AND column_name = 'org_id'
    ) THEN
        -- org_id exists (after migration 067)
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.table_constraints
            WHERE table_name = 'trips' AND constraint_name = 'fk_trips_organization'
        ) THEN
            ALTER TABLE trips
            ADD CONSTRAINT fk_trips_organization
            FOREIGN KEY (org_id) REFERENCES organizations(id)
            ON DELETE CASCADE;

            RAISE NOTICE 'Added constraint fk_trips_organization on trips.org_id';
        END IF;
    ELSIF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'trips' AND column_name = 'organization_id'
    ) THEN
        -- organization_id exists (before migration 067)
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.table_constraints
            WHERE table_name = 'trips' AND constraint_name = 'fk_trips_organization_legacy'
        ) THEN
            ALTER TABLE trips
            ADD CONSTRAINT fk_trips_organization_legacy
            FOREIGN KEY (organization_id) REFERENCES organizations(id)
            ON DELETE CASCADE;

            RAISE NOTICE 'Added constraint fk_trips_organization_legacy on trips.organization_id';
        END IF;
    ELSE
        RAISE WARNING 'Neither org_id nor organization_id found in trips table - cannot add FK';
    END IF;
END $$;

-- ========================================
-- Add comments to constraints if they exist
-- ========================================
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_name = 'trips' AND constraint_name = 'fk_trips_driver'
    ) THEN
        COMMENT ON CONSTRAINT fk_trips_driver ON trips IS 'Ensures trip driver references valid driver in same organization';
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_name = 'trips' AND constraint_name = 'fk_trips_organization'
    ) THEN
        COMMENT ON CONSTRAINT fk_trips_organization ON trips IS 'Ensures trip belongs to valid organization';
    END IF;
END $$;
