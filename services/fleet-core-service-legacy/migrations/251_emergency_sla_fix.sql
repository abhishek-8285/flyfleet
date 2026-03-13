-- Fix missing SLA violations table
CREATE TABLE IF NOT EXISTS incident_sla_violations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id UUID NOT NULL REFERENCES sos_incidents(id) ON DELETE CASCADE,
    severity TEXT NOT NULL,
    breach_type TEXT NOT NULL,
    breach_time TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sla_violations_incident_id ON incident_sla_violations(incident_id);
