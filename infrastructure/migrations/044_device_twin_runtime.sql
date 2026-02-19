-- Device twin runtime table for desired/reported convergence.
CREATE TABLE IF NOT EXISTS device_twins (
    device_id UUID PRIMARY KEY REFERENCES devices(id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    desired_state JSONB NOT NULL DEFAULT '{}'::jsonb,
    reported_state JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_seen TIMESTAMPTZ,
    drift_status VARCHAR(20) NOT NULL DEFAULT 'unknown',
    drift_reason TEXT,
    last_command_id UUID REFERENCES device_commands(id) ON DELETE SET NULL,
    last_command_status VARCHAR(32),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_device_twins_drift_status CHECK (drift_status IN ('in_sync', 'drifted', 'unknown'))
);

CREATE INDEX IF NOT EXISTS idx_device_twins_org ON device_twins(org_id);
CREATE INDEX IF NOT EXISTS idx_device_twins_drift_status ON device_twins(drift_status);
CREATE INDEX IF NOT EXISTS idx_device_twins_last_seen ON device_twins(last_seen DESC);
