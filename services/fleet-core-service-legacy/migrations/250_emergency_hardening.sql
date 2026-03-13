-- Phase 1: Contact Acknowledgement Tracking
CREATE TABLE IF NOT EXISTS emergency_contact_responses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id UUID NOT NULL REFERENCES sos_incidents(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL REFERENCES emergency_contacts(id),
    response_status TEXT NOT NULL DEFAULT 'NOTIFIED',
    responded_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for fast lookup by incident
CREATE INDEX IF NOT EXISTS idx_emergency_contact_responses_incident ON emergency_contact_responses(incident_id);

-- Phase 9: Incident Dashboard View
CREATE OR REPLACE VIEW emergency_incident_stats AS
SELECT
    DATE(created_at) as day,
    org_id,
    COUNT(*) as incidents,
    AVG(EXTRACT(EPOCH FROM resolved_at - created_at)) FILTER (WHERE status = 'RESOLVED') as avg_response_time
FROM sos_incidents
GROUP BY day, org_id;

-- Ensure indices for analytics
CREATE INDEX IF NOT EXISTS idx_sos_incidents_org_created ON sos_incidents(org_id, created_at);
