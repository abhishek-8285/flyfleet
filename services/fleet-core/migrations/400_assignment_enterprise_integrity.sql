-- Migration: Enterprise Assignment Integrity V4
-- Description: Resilient snapshots, timeouts, and policy support.

BEGIN;

-- 1. Operation Lifecycle (Leases & Timeouts)
ALTER TABLE assignment_operations ADD COLUMN IF NOT EXISTS operation_timeout_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '30 seconds');
ALTER TABLE assignment_operations ADD COLUMN IF NOT EXISTS bypass_reason TEXT;

-- 2. Hybrid Audit Support
ALTER TABLE assignment_audit_logs ADD COLUMN IF NOT EXISTS is_full_snapshot BOOLEAN DEFAULT FALSE;
ALTER TABLE assignment_audit_logs ADD COLUMN IF NOT EXISTS checkpoint_id UUID;

-- 3. Compliance & Policy Tracking
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS compliance_status TEXT DEFAULT 'COMPLIANT'; -- COMPLIANT, NON_COMPLIANT, PENDING_REVIEW
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS last_full_snapshot_at TIMESTAMPTZ;

ALTER TABLE drivers ADD COLUMN IF NOT EXISTS compliance_status TEXT DEFAULT 'COMPLIANT';
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS last_full_snapshot_at TIMESTAMPTZ;

-- 4. Outbox Retry & Classification
ALTER TABLE vehicle_outbox ADD COLUMN IF NOT EXISTS retry_count INT DEFAULT 0;
ALTER TABLE vehicle_outbox ADD COLUMN IF NOT EXISTS error_type TEXT; -- TRANSIENT, PERMANENT
ALTER TABLE vehicle_outbox ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMPTZ DEFAULT NOW();

COMMIT;
