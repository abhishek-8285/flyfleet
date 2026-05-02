-- Migration: 091_driver_vehicle_schema_harmonization
-- Description: Ensure critical columns for drivers and vehicles exist to prevent runtime errors.

-- 1. DRIVERS HARMONIZATION
DO $$ 
BEGIN 
    -- Ensure India Compliance fields exist
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
        ALTER TABLE drivers ADD COLUMN tenant_id UUID REFERENCES tenants(id);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='drivers' AND column_name='organization_id') THEN
        ALTER TABLE drivers ADD COLUMN organization_id UUID;
    END IF;

    -- Sync identities if possible
    UPDATE drivers SET tenant_id = organization_id WHERE tenant_id IS NULL AND organization_id IS NOT NULL;
    UPDATE drivers SET organization_id = tenant_id WHERE organization_id IS NULL AND tenant_id IS NOT NULL;
END $$;

-- 2. VEHICLES HARMONIZATION
DO $$ 
BEGIN 
    -- Ensure registration number consistency
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vehicles' AND column_name='registration_number') THEN
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vehicles' AND column_name='registration') THEN
            ALTER TABLE vehicles RENAME COLUMN registration TO registration_number;
        ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vehicles' AND column_name='license_plate') THEN
            ALTER TABLE vehicles RENAME COLUMN license_plate TO registration_number;
        ELSE
            ALTER TABLE vehicles ADD COLUMN registration_number VARCHAR(50);
        END IF;
    END IF;

    -- Identity Harmonization
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vehicles' AND column_name='tenant_id') THEN
        ALTER TABLE vehicles ADD COLUMN tenant_id UUID REFERENCES tenants(id);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vehicles' AND column_name='org_id') THEN
        ALTER TABLE vehicles ADD COLUMN org_id UUID;
    END IF;

    -- Sync identities
    UPDATE vehicles SET tenant_id = org_id WHERE tenant_id IS NULL AND org_id IS NOT NULL;
    UPDATE vehicles SET org_id = tenant_id WHERE org_id IS NULL AND tenant_id IS NOT NULL;

    -- Handle NOT NULL constraint on org_id if it exists
    -- If org_id is still null, we might need a default to satisfy constraints
    UPDATE vehicles SET org_id = '00000000-0000-0000-0000-000000000000' WHERE org_id IS NULL;
END $$;
