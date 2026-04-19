-- +goose Up
-- +goose StatementBegin
-- Repairs the vehicles table to match Enterprise Repository expectations

-- 1. Identity & Registration
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS registration_number VARCHAR(50);
UPDATE vehicles SET registration_number = license_plate WHERE registration_number IS NULL;

-- 2. Compliance & Encryption (stubs for AES-256)
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS chassis_number TEXT;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS engine_number TEXT;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS vin_number VARCHAR(100);
UPDATE vehicles SET vin_number = vin WHERE vin_number IS NULL AND vin IS NOT NULL;

-- 3. Enterprise Metadata
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS vehicle_type VARCHAR(50) DEFAULT 'TRUCK';
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS capacity_tonnage DOUBLE PRECISION;
UPDATE vehicles SET capacity_tonnage = capacity WHERE capacity_tonnage IS NULL;

ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS fuel_tank_capacity DOUBLE PRECISION;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS ownership_type VARCHAR(50) DEFAULT 'OWNED';
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS parent_vehicle_id UUID;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS branch_id UUID;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS cost_center_id UUID;

-- 4. IoT & GPS
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS gps_device_id VARCHAR(100);
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS gps_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS gps_device_type VARCHAR(50);
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS immobilizer_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS ais140_compliant BOOLEAN DEFAULT FALSE;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS speed_limit DOUBLE PRECISION DEFAULT 80.0;

-- 5. Audit & Optimistic Locking
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;

-- 6. Ensure status enum compatibility
ALTER TABLE vehicles DROP CONSTRAINT IF EXISTS chk_vehicles_status;
ALTER TABLE vehicles ADD CONSTRAINT chk_vehicles_status 
    CHECK (status IN ('ACTIVE', 'MAINTENANCE', 'SUSPENDED', 'DECOMMISSIONED', 'active', 'inactive', 'maintenance', 'retired', 'out_of_service'));

-- +goose StatementEnd
