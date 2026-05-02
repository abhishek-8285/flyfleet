-- Migration: 253_emergency_reliability_audit
-- Description: Authority delivery logs, timeline events, and geospatial authority routing

-- Phase 2: Authority delivery confirmation
CREATE TABLE IF NOT EXISTS authority_delivery_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id UUID NOT NULL REFERENCES sos_incidents(id) ON DELETE CASCADE,
    authority_type TEXT NOT NULL,
    delivery_status TEXT NOT NULL, -- SENT, CONFIRMED, FAILED
    response_code INTEGER,
    response_body TEXT,
    retry_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_authority_delivery_incident ON authority_delivery_logs(incident_id);

-- Phase 5: Incident timeline storage
CREATE TABLE IF NOT EXISTS incident_timeline_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id UUID NOT NULL REFERENCES sos_incidents(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL, -- SOS_TRIGGERED, CONTACT_NOTIFIED, ESCALATION_STARTED, AUTHORITY_ALERTED, ACKNOWLEDGED, RESOLVED
    event_data JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_incident_timeline_incident ON incident_timeline_events(incident_id);
CREATE INDEX IF NOT EXISTS idx_incident_timeline_type ON incident_timeline_events(event_type);

-- Phase 4: Geo failover authority routing
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='emergency_authority_zones' AND column_name='priority') THEN
        ALTER TABLE emergency_authority_zones ADD COLUMN priority INTEGER DEFAULT 1;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='emergency_authority_zones' AND column_name='zone_polygon') THEN
        -- Falling back to JSONB as PostGIS is not available in this environment
        ALTER TABLE emergency_authority_zones ADD COLUMN zone_polygon JSONB DEFAULT '{}'::jsonb;
    END IF;
END $$;

-- Seed some failover zones (placeholders using GeoJSON format)
INSERT INTO emergency_authority_zones (zone_name, authority_type, zone_polygon, contact_number, priority)
VALUES 
('Gurgaon Sector 45 Police Station (Backup)', 'POLICE', '{"type": "Polygon", "coordinates": [[[77.01, 28.44], [77.05, 28.44], [77.05, 28.47], [77.01, 28.47], [77.01, 28.44]]]}'::jsonb, '+911242385101', 2),
('Cyber City Fire Station (Backup)', 'FIRE', '{"type": "Polygon", "coordinates": [[[77.07, 28.48], [77.10, 28.48], [77.10, 28.51], [77.07, 28.51], [77.07, 28.48]]]}'::jsonb, '+911242388101', 2)
ON CONFLICT DO NOTHING;
