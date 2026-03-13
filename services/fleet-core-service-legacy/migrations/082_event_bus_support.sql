-- 082_event_bus_support.sql
-- Support tables for Event Bus standardization

-- Optional: If we need to persist events for replay or auditing
CREATE TABLE IF NOT EXISTS platform_events (
    id UUID PRIMARY KEY,
    event_type TEXT NOT NULL,
    organization_id UUID,
    source TEXT NOT NULL,
    payload JSONB NOT NULL,
    timestamp BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_events_org ON platform_events(organization_id);
CREATE INDEX idx_events_type ON platform_events(event_type);
CREATE INDEX idx_events_timestamp ON platform_events(timestamp);
