-- Migration: 220_emergency_schema_harmonization
-- Description: Align database schema with EmergencyRepository implementation

-- 1. Harmonize sos_incidents
ALTER TABLE sos_incidents 
ADD COLUMN IF NOT EXISTS org_id UUID,
ADD COLUMN IF NOT EXISTS acknowledged_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS resolved_at TIMESTAMP WITH TIME ZONE;

-- 2. Harmonize emergency_contacts
DO $$ 
BEGIN 
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'emergency_contacts' AND column_name = 'organization_id') THEN
        ALTER TABLE emergency_contacts RENAME COLUMN organization_id TO org_id;
    END IF;
END $$;

-- 3. Harmonize emergency_response_units
DO $$ 
BEGIN 
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'emergency_response_units' AND column_name = 'organization_id') THEN
        ALTER TABLE emergency_response_units RENAME COLUMN organization_id TO org_id;
    END IF;
END $$;

-- 4. Harmonize emergency_vehicle_state
DO $$ 
BEGIN 
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'emergency_vehicle_state' AND column_name = 'organization_id') THEN
        ALTER TABLE emergency_vehicle_state RENAME COLUMN organization_id TO org_id;
    END IF;
END $$;

-- 5. Add missing indexes for tenant isolation
CREATE INDEX IF NOT EXISTS idx_sos_incidents_org_id ON sos_incidents(org_id);
