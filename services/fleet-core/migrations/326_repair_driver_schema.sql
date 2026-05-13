-- Migration: 326_repair_driver_schema
-- Description: Add missing columns to drivers table to resolve runtime 500 errors.

DO $$ 
BEGIN 
    -- India Compliance fields
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='drivers' AND column_name='employee_code') THEN
        ALTER TABLE drivers ADD COLUMN employee_code VARCHAR(50);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='drivers' AND column_name='aadhaar_number') THEN
        ALTER TABLE drivers ADD COLUMN aadhaar_number VARCHAR(20);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='drivers' AND column_name='pan_number') THEN
        ALTER TABLE drivers ADD COLUMN pan_number VARCHAR(20);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='drivers' AND column_name='license_issue_date') THEN
        ALTER TABLE drivers ADD COLUMN license_issue_date DATE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='drivers' AND column_name='employment_type') THEN
        ALTER TABLE drivers ADD COLUMN employment_type VARCHAR(50);
    END IF;

    -- Identity Harmonization
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='drivers' AND column_name='tenant_id') THEN
        ALTER TABLE drivers ADD COLUMN tenant_id UUID;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='drivers' AND column_name='organization_id') THEN
        ALTER TABLE drivers ADD COLUMN organization_id UUID;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='drivers' AND column_name='full_name') THEN
        ALTER TABLE drivers ADD COLUMN full_name VARCHAR(255);
    END IF;
END $$;
