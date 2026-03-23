-- Migration: 216_emergency_india_extensions
-- Description: Upgrade emergency module for AIS-140 / India compliance

-- 1. Extend sos_incidents table
ALTER TABLE sos_incidents 
ADD COLUMN IF NOT EXISTS incident_type VARCHAR(50) DEFAULT 'SOS_PANIC',
ADD COLUMN IF NOT EXISTS device_id UUID,
ADD COLUMN IF NOT EXISTS panic_source VARCHAR(50) DEFAULT 'DRIVER_MOBILE',
ADD COLUMN IF NOT EXISTS speed INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS heading INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS acknowledged_by UUID REFERENCES users(id),
ADD COLUMN IF NOT EXISTS resolved_by UUID REFERENCES users(id);

-- Add check constraint for incident_type
DO $$ 
BEGIN 
    ALTER TABLE sos_incidents ADD CONSTRAINT check_incident_type 
    CHECK (incident_type IN ('SOS_PANIC', 'VEHICLE_ACCIDENT', 'VEHICLE_BREAKDOWN', 'MEDICAL_EMERGENCY', 'SECURITY_THREAT', 'FIRE_ALERT'));
EXCEPTION 
    WHEN duplicate_object THEN NULL; 
END $$;

-- Add check constraint for panic_source
DO $$ 
BEGIN 
    ALTER TABLE sos_incidents ADD CONSTRAINT check_panic_source 
    CHECK (panic_source IN ('AIS140_DEVICE', 'DRIVER_MOBILE', 'VEHICLE_TELEMATICS', 'DISPATCHER'));
EXCEPTION 
    WHEN duplicate_object THEN NULL; 
END $$;

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_sos_trip ON sos_incidents(trip_id);
CREATE INDEX IF NOT EXISTS idx_sos_device ON sos_incidents(device_id);
CREATE INDEX IF NOT EXISTS idx_sos_incident_type ON sos_incidents(incident_type);

-- 2. Create emergency_escalations table
CREATE TABLE IF NOT EXISTS emergency_escalations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id UUID NOT NULL REFERENCES sos_incidents(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL REFERENCES emergency_contacts(id),
    notification_channel VARCHAR(50) NOT NULL CHECK (notification_channel IN ('SMS', 'PHONE', 'WHATSAPP', 'PUSH')),
    status VARCHAR(20) NOT NULL DEFAULT 'SENT' CHECK (status IN ('SENT', 'DELIVERED', 'FAILED', 'RESPONDED')),
    attempted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    responded_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_escalations_incident_id ON emergency_escalations(incident_id);

-- 3. Create emergency_response_units table
CREATE TABLE IF NOT EXISTS emergency_response_units (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    unit_type VARCHAR(50) NOT NULL CHECK (unit_type IN ('POLICE', 'AMBULANCE', 'FIRE', 'ROAD_ASSISTANCE')),
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    coverage_city TEXT,
    coverage_state TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_eru_org_id ON emergency_response_units(organization_id);

-- 4. Create emergency_vehicle_state table
CREATE TABLE IF NOT EXISTS emergency_vehicle_state (
    vehicle_id UUID PRIMARY KEY REFERENCES vehicles(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id),
    halted BOOLEAN DEFAULT FALSE,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_evs_org_id ON emergency_vehicle_state(organization_id);
