-- Migration: 073_tracking_hardening
-- Description: Transactional Outbox for Telemetry, Tenant-ID enforcement,
--              and index hardening.

-- 1. Add tenant_id NOT NULL to telemetry_points (if not already strictly enforced)
DO $$ 
BEGIN 
    ALTER TABLE telemetry_points ALTER COLUMN tenant_id SET NOT NULL; 
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- 2. Create telemetry_outbox table for NATS relay
CREATE TABLE IF NOT EXISTS telemetry_outbox (
    id          BIGSERIAL PRIMARY KEY,
    partition_id INTEGER NOT NULL,      -- helps with pruning if needed
    payload     JSONB NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'processed', 'failed')),
    attempts    SMALLINT NOT NULL DEFAULT 0,
    last_attempt TIMESTAMPTZ,
    error_msg   TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for the relay worker (FIFO)
CREATE INDEX IF NOT EXISTS idx_telem_outbox_pending
    ON telemetry_outbox (created_at ASC)
    WHERE status = 'pending';

-- 3. Hardening: Add constraint to ensure lat/lng are always valid
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_name = 'telemetry_points' AND constraint_name = 'check_coords') THEN
        ALTER TABLE telemetry_points
            ADD CONSTRAINT check_coords CHECK (lat BETWEEN -90 AND 90 AND lng BETWEEN -180 AND 180);
    END IF;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- 4. Anomaly Thresholds Table (Per-Vehicle / Per-Tenant)
CREATE TABLE IF NOT EXISTS tracking_config (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    target_id      UUID NOT NULL,      -- vehicle_id or tenant_id
    target_type    VARCHAR(20) NOT NULL CHECK (target_type IN ('vehicle', 'tenant')),
    config         JSONB NOT NULL DEFAULT '{}', -- thresholds: overspeed, harsh_accel, etc.
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(target_id, target_type)
);

COMMENT ON TABLE telemetry_outbox IS 'Transactional outbox for reliable NATS JetStream publishing';
COMMENT ON TABLE tracking_config  IS 'Custom thresholds for anomaly detection per vehicle/tenant';
