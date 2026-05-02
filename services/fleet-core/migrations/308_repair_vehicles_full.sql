-- +goose Up
-- +goose StatementBegin
-- Repairs the vehicles table to match Enterprise Repository expectations
-- Hardened: Use multi-column fallback for backfills

-- 1. Identity & Registration
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS registration_number VARCHAR(50);

DO $$
BEGIN
    -- Backfill registration_number from any available legacy column
    UPDATE vehicles 
    SET registration_number = COALESCE(
        registration_number,
        (SELECT c.column_name FROM information_schema.columns c WHERE c.table_name='vehicles' AND c.column_name='registration_number_legacy' LIMIT 1),
        (SELECT c.column_name FROM information_schema.columns c WHERE c.table_name='vehicles' AND c.column_name='license_plate' LIMIT 1),
        (SELECT c.column_name FROM information_schema.columns c WHERE c.table_name='vehicles' AND c.column_name='registration' LIMIT 1)
    )
    WHERE registration_number IS NULL OR registration_number = '';
    
    -- Fallback to ID if still null
    UPDATE vehicles SET registration_number = id::text WHERE registration_number IS NULL OR registration_number = '';
EXCEPTION WHEN OTHERS THEN 
    -- Manual approach if subqueries are tricky in UPDATE
    UPDATE vehicles SET registration_number = id::text WHERE registration_number IS NULL;
END $$;

-- 2. Compliance & Encryption (stubs for AES-256)
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS chassis_number TEXT;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS engine_number TEXT;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS vin_number VARCHAR(100);

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vehicles' AND column_name='vin') THEN
        UPDATE vehicles SET vin_number = vin WHERE vin_number IS NULL AND vin IS NOT NULL;
    END IF;
END $$;

-- 3. Enterprise Metadata
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS vehicle_type VARCHAR(50) DEFAULT 'TRUCK';
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS capacity_tonnage DOUBLE PRECISION;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vehicles' AND column_name='capacity') THEN
        UPDATE vehicles SET capacity_tonnage = capacity::double precision WHERE capacity_tonnage IS NULL;
    END IF;
END $$;

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
