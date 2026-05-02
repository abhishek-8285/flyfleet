-- Migration: 038_emergency_response
-- Description: Creates tables for SOS incidents, emergency contacts, and incident logs.

-- 1. Create sos_incidents table
CREATE TABLE IF NOT EXISTS sos_incidents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id UUID NOT NULL, -- references vehicles(id)
    driver_id UUID, -- references drivers(id)
    trip_id UUID,     -- references trips(id)
    
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    location_address TEXT,
    
    status TEXT NOT NULL DEFAULT 'TRIGGERED' CHECK (status IN ('TRIGGERED', 'ACKNOWLEDGED', 'RESOLVED', 'FALSE_ALARM')),
    severity TEXT NOT NULL DEFAULT 'CRITICAL' CHECK (severity IN ('LOW', 'MEDIUM', 'CRITICAL')),
    
    metadata JSONB DEFAULT '{}'::jsonb,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for querying active incidents
CREATE INDEX IF NOT EXISTS idx_sos_incidents_status ON sos_incidents(status);
CREATE INDEX IF NOT EXISTS idx_sos_incidents_vehicle_id ON sos_incidents(vehicle_id);

-- 2. Create emergency_contacts table
CREATE TABLE IF NOT EXISTS emergency_contacts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    
    name TEXT NOT NULL,
    role TEXT NOT NULL,
    phone TEXT NOT NULL,
    email TEXT,
    
    priority INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT TRUE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_emergency_contacts_tenant_id ON emergency_contacts(tenant_id);

-- 3. Create incident_logs table
CREATE TABLE IF NOT EXISTS incident_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    incident_id UUID NOT NULL REFERENCES sos_incidents(id),
    user_id UUID,
    
    action TEXT NOT NULL,
    notes TEXT,
    
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_incident_logs_incident_id ON incident_logs(incident_id);

-- 4. Cleanup/Resilience: Add tenant_id to sos_incidents if it's missing (it was missing in original)
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'sos_incidents' AND column_name = 'tenant_id') THEN
        ALTER TABLE sos_incidents ADD COLUMN tenant_id UUID REFERENCES tenants(id);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_sos_incidents_tenant_id ON sos_incidents(tenant_id);
