-- 083_admin_support_phase3.sql
-- Description: Adds schema for two-person approvals and search indexing metadata.

BEGIN;

CREATE TABLE IF NOT EXISTS admin_operation_approvals (
    id UUID PRIMARY KEY,
    operation_type TEXT NOT NULL,
    target_id TEXT,
    payload JSONB,
    requested_by UUID NOT NULL,
    status TEXT NOT NULL,
    approved_by UUID,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_op_approvals_status ON admin_operation_approvals(status);
CREATE INDEX IF NOT EXISTS idx_admin_op_approvals_req_by ON admin_operation_approvals(requested_by);

-- Dummy table to track sync state for the admin search index
CREATE TABLE IF NOT EXISTS admin_search_sync_state (
    domain TEXT PRIMARY KEY,
    last_synced_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO admin_search_sync_state (domain) VALUES ('organizations'), ('users'), ('tickets'), ('incidents')
ON CONFLICT (domain) DO NOTHING;

COMMIT;
