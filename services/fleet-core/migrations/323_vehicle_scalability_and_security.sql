-- +goose Up
-- +goose StatementBegin
-- Migration: 323_vehicle_scalability_and_security.sql
-- Purpose: Implement partitioning for high-volume logs, tenant-scoped uniqueness, and compliance hard-blocks.

-- 1. Hard Compliance Layer
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS is_hard_blocked BOOLEAN DEFAULT FALSE;

-- 2. Tenant-Scoped Uniqueness Hardening
-- Drop previous global index if it exists
DROP INDEX IF EXISTS idx_unique_registration_active;
-- Enforce uniqueness per tenant for active registration numbers.
CREATE UNIQUE INDEX idx_unique_registration_tenant_active ON vehicles (tenant_id, registration_number) WHERE deleted_at IS NULL;

-- 3. High-Volume Log Partitioning (Vehicle Audit Logs)
-- Recreating vehicle_audit_logs as a partitioned table by range of timestamp.
DROP TABLE IF EXISTS vehicle_audit_logs;
CREATE TABLE vehicle_audit_logs (
    id UUID NOT NULL,
    vehicle_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    action TEXT NOT NULL,
    field_name TEXT,
    old_value TEXT,
    new_value TEXT,
    timestamp TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

-- Create initial partitions for current and next 2 months
CREATE TABLE vehicle_audit_logs_y2026m05 PARTITION OF vehicle_audit_logs FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE vehicle_audit_logs_y2026m06 PARTITION OF vehicle_audit_logs FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');

-- 4. Async Processing (Outbox Pattern)
CREATE TABLE vehicle_outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    processed_at TIMESTAMPTZ
);
CREATE INDEX idx_vehicle_outbox_unprocessed ON vehicle_outbox (created_at) WHERE processed_at IS NULL;

-- 5. Composite Indices Verification (Enforcing isolation)
CREATE INDEX IF NOT EXISTS idx_vehicles_tenant_status ON vehicles (tenant_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vehicles_tenant_branch ON vehicles (tenant_id, branch_id) WHERE deleted_at IS NULL;

-- +goose StatementEnd
