-- Migration: 303_force_schema_sync_enterprise (REVISED)
-- Description: Synchronize legacy schema with Enterprise backend models WITHOUT breaking foreign keys

-- 1. Restore and Alter VEHICLES
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'vehicles_legacy') THEN
        ALTER TABLE vehicles_legacy RENAME TO vehicles;
    END IF;
END $$;

-- Rename and add columns to match models.Vehicle
ALTER TABLE vehicles RENAME COLUMN license_plate TO registration_number_legacy;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS registration_number VARCHAR(50);
UPDATE vehicles SET registration_number = COALESCE(registration_number_legacy, registration, '');
ALTER TABLE vehicles ALTER COLUMN registration_number SET NOT NULL;

ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS chassis_number TEXT;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS engine_number TEXT;
ALTER TABLE vehicles RENAME COLUMN vin TO vin_number;

ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS vehicle_type VARCHAR(30) DEFAULT 'TRUCK';
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS capacity_tonnage DECIMAL(10,2);
UPDATE vehicles SET capacity_tonnage = capacity::decimal WHERE capacity IS NOT NULL;

ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS fuel_tank_capacity DECIMAL(10,2);
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS ownership_type VARCHAR(20) DEFAULT 'OWNED';
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS parent_vehicle_id UUID;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS branch_id UUID;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS cost_center_id UUID;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS gps_device_id VARCHAR(100);
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS immobilizer_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS ais140_compliant BOOLEAN DEFAULT FALSE;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1;

-- Clean up legacy name
-- ALTER TABLE vehicles DROP COLUMN registration_number_legacy; 

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

UPDATE trips SET trip_id = id::text WHERE trip_id IS NULL;
UPDATE trips SET distance = COALESCE(actual_distance_km, estimated_distance_km, 0) WHERE distance = 0;

-- Indexes
CREATE INDEX IF NOT EXISTS idx_vehicles_registration ON vehicles(registration_number);
