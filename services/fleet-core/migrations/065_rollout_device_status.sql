-- Migration: 065_rollout_device_status.sql
-- Description: Track firmware rollout per device and add safety controls

-- Add safety controls to firmware_rollouts
ALTER TABLE firmware_rollouts
ADD COLUMN IF NOT EXISTS failure_threshold INTEGER DEFAULT 5,
ADD COLUMN IF NOT EXISTS batch_size INTEGER DEFAULT 100,
ADD COLUMN IF NOT EXISTS rollback_enabled BOOLEAN DEFAULT false;

-- Create rollout device status tracking
CREATE TABLE IF NOT EXISTS rollout_device_status (
    id BIGSERIAL PRIMARY KEY,
    rollout_id UUID REFERENCES firmware_rollouts(id),
    device_id UUID REFERENCES devices(id),
    status TEXT CHECK(status IN ('PENDING', 'UPDATING', 'SUCCESS', 'FAILED')),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_rollout_device_status_rollout ON rollout_device_status(rollout_id);
CREATE INDEX idx_rollout_device_status_device ON rollout_device_status(device_id);
CREATE INDEX idx_rollout_device_status_state ON rollout_device_status(rollout_id, status);
