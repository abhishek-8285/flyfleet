-- Firmware rollout event history / audit trail
-- Created: 2026-02-19

CREATE TABLE IF NOT EXISTS firmware_rollout_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rollout_id UUID NOT NULL REFERENCES firmware_rollouts(id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    event_type VARCHAR(64) NOT NULL,
    from_status VARCHAR(32),
    to_status VARCHAR(32),
    actor_user_id UUID REFERENCES users(id),
    message TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_firmware_rollout_events_rollout_created
    ON firmware_rollout_events(rollout_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_firmware_rollout_events_org_created
    ON firmware_rollout_events(org_id, created_at DESC);
