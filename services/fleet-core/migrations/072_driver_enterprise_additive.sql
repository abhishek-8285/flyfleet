-- Migration: 072_driver_enterprise_additive
-- Description: Add enterprise columns to drivers table (complements 064_driver_management)
-- This migration is ADDITIVE only — adds columns that 064 doesn't have.

-- Add enterprise columns (IF NOT EXISTS to be safe)
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS gender VARCHAR(10);
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS address_line_1 TEXT;
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS address_line_2 TEXT;
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS city VARCHAR(100);
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS state VARCHAR(50);
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS pincode VARCHAR(10);
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS emergency_name VARCHAR(100);
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS emergency_phone VARCHAR(20);
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS blood_group VARCHAR(5);
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS shift_type VARCHAR(20) DEFAULT 'MORNING'
    CHECK (shift_type IN ('MORNING','AFTERNOON','NIGHT','ROTATING'));
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS behavior_monitoring BOOLEAN DEFAULT TRUE;
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS speed_limit_override INTEGER;
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS notes TEXT;
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS license_issue_date DATE;
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS puc_number VARCHAR(50);
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS puc_expiry_date DATE;
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS medical_cert_file TEXT;
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS medical_expiry_date DATE;
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS branch_id UUID;
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;

-- Add indexes for new columns (CONCURRENTLY for production safety)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_tenant_status ON drivers(org_id, status) WHERE deleted_at IS NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_license_expiry ON drivers(license_expiry) WHERE deleted_at IS NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_assigned_vehicle ON drivers(current_vehicle_id) WHERE current_vehicle_id IS NOT NULL AND deleted_at IS NULL;

-- Comments
COMMENT ON COLUMN drivers.gender IS 'Driver gender';
COMMENT ON COLUMN drivers.shift_type IS 'MORNING, AFTERNOON, NIGHT, ROTATING';
COMMENT ON COLUMN drivers.behavior_monitoring IS 'Enable behavior monitoring for this driver';
COMMENT ON COLUMN drivers.blood_group IS 'Driver blood group';
COMMENT ON COLUMN drivers.notes IS 'Admin notes about this driver';
