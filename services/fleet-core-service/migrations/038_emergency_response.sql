-- Migration: 038_emergency_response
-- Description: Creates tables for SOS incidents, emergency contacts, and incident logs.

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Create sos_incidents table
CREATE TABLE IF NOT EXISTS sos_incidents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id),
    driver_id UUID REFERENCES drivers(id), -- Nullable if driver is unknown/not assigned
    trip_id UUID REFERENCES trips(id),     -- Nullable if not on an active trip
    
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    location_address TEXT,
    
    status TEXT NOT NULL DEFAULT 'TRIGGERED' CHECK (status IN ('TRIGGERED', 'ACKNOWLEDGED', 'RESOLVED', 'FALSE_ALARM')),
    severity TEXT NOT NULL DEFAULT 'CRITICAL' CHECK (severity IN ('LOW', 'MEDIUM', 'CRITICAL')),
    
    metadata JSONB DEFAULT '{}'::jsonb, -- Store raw device data, battery level, etc.
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for querying active incidents
CREATE INDEX IF NOT EXISTS idx_sos_incidents_status ON sos_incidents(status);
CREATE INDEX IF NOT EXISTS idx_sos_incidents_vehicle_id ON sos_incidents(vehicle_id);

-- 2. Create emergency_contacts table
CREATE TABLE IF NOT EXISTS emergency_contacts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    
    name TEXT NOT NULL,
    role TEXT NOT NULL, -- e.g., 'Safety Officer', 'Regional Manager'
    phone TEXT NOT NULL,
    email TEXT,
    
    priority INTEGER DEFAULT 1, -- 1 = First to call
    is_active BOOLEAN DEFAULT TRUE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_emergency_contacts_org_id ON emergency_contacts(organization_id);

-- 3. Create incident_logs table (Audit Trail)
CREATE TABLE IF NOT EXISTS incident_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    incident_id UUID NOT NULL REFERENCES sos_incidents(id),
    user_id UUID, -- Nullable if system action
    
    action TEXT NOT NULL, -- e.g., 'ACKNOWLEDGED', 'CALLED_POLICE', 'RESOLVED'
    notes TEXT,
    
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_incident_logs_incident_id ON incident_logs(incident_id);

-- Trigger to update updated_at
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_sos_incidents_modtime
BEFORE UPDATE ON sos_incidents
FOR EACH ROW
EXECUTE PROCEDURE update_timestamp();

CREATE TRIGGER update_emergency_contacts_modtime
BEFORE UPDATE ON emergency_contacts
FOR EACH ROW
EXECUTE PROCEDURE update_timestamp();
