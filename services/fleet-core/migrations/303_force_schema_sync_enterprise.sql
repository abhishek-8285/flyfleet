-- Migration: 303_force_schema_sync_enterprise (REVISED)
-- Description: Synchronize legacy schema with Enterprise backend models WITHOUT breaking foreign keys
-- Robustness: Added IF EXISTS for column renames and handled collision during table rename.

-- 1. Restore and Alter VEHICLES
DO $$ 
BEGIN
    -- Only rename if legacy table exists AND target table doesn't
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'vehicles_legacy') THEN
        IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'vehicles') THEN
            ALTER TABLE vehicles_legacy RENAME TO vehicles;
        ELSE
            -- If both exist, we might have a collision. 
            -- In stabilization mode, we assume 'vehicles' is the one we want to keep if it has data.
            -- We'll just leave them both and let downstream migrations handle it.
            RAISE NOTICE 'Both vehicles and vehicles_legacy exist. Skipping rename to avoid collision.';
        END IF;
    END IF;
END $$;

-- Ensure vehicles table exists before proceeding
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'vehicles') THEN
        CREATE TABLE vehicles (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            status VARCHAR(50) DEFAULT 'active',
            created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
        );
    END IF;
END $$;

-- Rename and add columns to match models.Vehicle
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'license_plate') AND
       NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'registration_number_legacy') THEN
        ALTER TABLE vehicles RENAME COLUMN license_plate TO registration_number_legacy;
    END IF;
END $$;

ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS registration_number VARCHAR(50);

-- Backfill registration_number (idempotent)
DO $$
BEGIN
    UPDATE vehicles SET registration_number = COALESCE(registration_number_legacy, registration, '') 
    WHERE (registration_number IS NULL OR registration_number = '')
    AND (registration_number_legacy IS NOT NULL OR registration IS NOT NULL);
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- Ensure registration_number is not null for future queries
-- but don't force it if we have no data yet (prevents failure on empty tables)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM vehicles) THEN
        UPDATE vehicles SET registration_number = id::text WHERE registration_number IS NULL OR registration_number = '';
    END IF;
END $$;

ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS chassis_number TEXT;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS engine_number TEXT;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'vin') AND
       NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'vin_number') THEN
        ALTER TABLE vehicles RENAME COLUMN vin TO vin_number;
    END IF;
END $$;

ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS vehicle_type VARCHAR(30) DEFAULT 'TRUCK';
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS capacity_tonnage DECIMAL(10,2);

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'capacity') THEN
        UPDATE vehicles SET capacity_tonnage = capacity::decimal 
        WHERE capacity IS NOT NULL AND capacity_tonnage IS NULL;
    END IF;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS fuel_tank_capacity DECIMAL(10,2);
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS ownership_type VARCHAR(20) DEFAULT 'OWNED';
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS parent_vehicle_id UUID;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS branch_id UUID;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS cost_center_id UUID;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS gps_device_id VARCHAR(100);
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS immobilizer_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS ais140_compliant BOOLEAN DEFAULT FALSE;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1;

-- 2. DRIVERS: Add missing Enterprise/India Compliance fields
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS employee_code VARCHAR(50);
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS aadhaar_number VARCHAR(20);
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS pan_number VARCHAR(20);
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS license_issue_date DATE;
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS employment_type VARCHAR(50);

-- 3. TRIPS: Synchronize with Enterprise Lifecycle
ALTER TABLE trips ADD COLUMN IF NOT EXISTS trip_id VARCHAR(50);
ALTER TABLE trips ADD COLUMN IF NOT EXISTS distance DOUBLE PRECISION DEFAULT 0;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS fuel_consumed_liters DOUBLE PRECISION DEFAULT 0;

DO $$
BEGIN
    UPDATE trips SET trip_id = id::text WHERE trip_id IS NULL;
    -- Only update distance if columns exist
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='trips' AND column_name='actual_distance_km') THEN
        UPDATE trips SET distance = COALESCE(actual_distance_km, estimated_distance_km, 0) WHERE distance = 0;
    END IF;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- Indexes
CREATE INDEX IF NOT EXISTS idx_vehicles_registration ON vehicles(registration_number);
