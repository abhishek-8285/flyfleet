-- Migration: 214_driver_india_extensions
-- Description: Add Indian compliance and enterprise fields to the existing drivers table

ALTER TABLE drivers 
    ADD COLUMN IF NOT EXISTS employee_code VARCHAR(50),
    ADD COLUMN IF NOT EXISTS aadhaar_number VARCHAR(12),
    ADD COLUMN IF NOT EXISTS pan_number VARCHAR(10),
    ADD COLUMN IF NOT EXISTS license_issue_date DATE,
    ADD COLUMN IF NOT EXISTS employment_type VARCHAR(20) DEFAULT 'FULL_TIME' CHECK (employment_type IN ('FULL_TIME', 'CONTRACT'));

-- Add unique constraints
ALTER TABLE drivers
    ADD CONSTRAINT uq_drivers_employee_code UNIQUE (org_id, employee_code),
    ADD CONSTRAINT uq_drivers_aadhaar UNIQUE (aadhaar_number),
    ADD CONSTRAINT uq_drivers_pan UNIQUE (pan_number);

-- Indexes for new fields
CREATE INDEX IF NOT EXISTS idx_drivers_employee_code ON drivers(employee_code);
CREATE INDEX IF NOT EXISTS idx_drivers_aadhaar ON drivers(aadhaar_number);
CREATE INDEX IF NOT EXISTS idx_drivers_pan ON drivers(pan_number);

-- Comments
COMMENT ON COLUMN drivers.employee_code IS 'Internal employee identifier';
COMMENT ON COLUMN drivers.aadhaar_number IS 'Indian Aadhaar 12-digit UID';
COMMENT ON COLUMN drivers.pan_number IS 'Indian PAN Permanent Account Number';
COMMENT ON COLUMN drivers.employment_type IS 'FULL_TIME or CONTRACT employment status';
