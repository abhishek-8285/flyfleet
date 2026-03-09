-- 218_emergency_advanced_features.sql

-- Incident Media evidence
CREATE TABLE IF NOT EXISTS incident_media (
    id UUID PRIMARY KEY,
    incident_id UUID NOT NULL REFERENCES sos_incidents(id) ON DELETE CASCADE,
    media_type VARCHAR(20) NOT NULL, -- IMAGE, VIDEO, AUDIO
    file_url TEXT NOT NULL,
    uploaded_by UUID, -- Can be NULL if automated
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_incident_media_incident_id ON incident_media(incident_id);

-- Incident Vehicle Snapshot (Forensic Telemetry)
CREATE TABLE IF NOT EXISTS incident_vehicle_snapshot (
    incident_id UUID PRIMARY KEY REFERENCES sos_incidents(id) ON DELETE CASCADE,
    speed INTEGER,
    heading INTEGER,
    engine_status BOOLEAN,
    fuel_level INTEGER,
    battery_voltage FLOAT,
    gps_accuracy FLOAT,
    device_signal_strength INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add high_risk_location flag to sos_incidents
ALTER TABLE sos_incidents ADD COLUMN IF NOT EXISTS high_risk_location BOOLEAN DEFAULT FALSE;
