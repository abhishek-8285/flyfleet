-- +goose Up
-- +goose StatementBegin
-- Migration: 324_vehicle_infra_lifecycle.sql
-- Purpose: Implement TTL-aware outbox partitioning and lifecycle management for high-volume data.

-- 1. TTL-Aware Outbox Partitioning
-- Recreating vehicle_outbox as a partitioned table by range of created_at.
DROP TABLE IF EXISTS vehicle_outbox;
CREATE TABLE vehicle_outbox (
    id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    processed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Create initial outbox partitions
CREATE TABLE vehicle_outbox_y2026m05 PARTITION OF vehicle_outbox FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE vehicle_outbox_y2026m06 PARTITION OF vehicle_outbox FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');

-- 2. Lifecycle Indexes (TTL)
CREATE INDEX idx_vehicle_outbox_ttl ON vehicle_outbox (expires_at) WHERE processed_at IS NOT NULL;
CREATE INDEX idx_vehicle_audit_logs_ttl ON vehicle_audit_logs (timestamp);

-- 3. RBAC Schema Initialization
-- Ensure fleet-specific permissions exist in the IAM layer
-- Note: Assuming iam_permissions table exists as per earlier IAM conversations
INSERT INTO iam_permissions (id, code, description, category)
VALUES 
    (gen_random_uuid(), 'fleet.vehicle.block', 'Permission to hard-block assets due to compliance', 'FLEET'),
    (gen_random_uuid(), 'fleet.vehicle.unblock', 'Permission to lift hard-blocks from assets', 'FLEET'),
    (gen_random_uuid(), 'fleet.vehicle.audit', 'Permission to view sensitive vehicle audit trails', 'FLEET')
ON CONFLICT (code) DO NOTHING;

-- +goose StatementEnd
