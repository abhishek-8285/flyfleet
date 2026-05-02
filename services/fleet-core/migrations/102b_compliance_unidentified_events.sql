-- Migration: 102b_compliance_unidentified_events
-- Description: Create the missing unidentified_driver_events table for FMCSA compliance.

CREATE TABLE IF NOT EXISTS unidentified_driver_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id),
    event_time TIMESTAMP WITH TIME ZONE NOT NULL,
    location TEXT,
    odometer DOUBLE PRECISION,
    assigned_driver_id UUID REFERENCES drivers(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_unidentified_events_tenant ON unidentified_driver_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_unidentified_events_vehicle ON unidentified_driver_events(vehicle_id);
