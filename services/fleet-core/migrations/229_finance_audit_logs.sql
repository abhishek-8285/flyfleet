-- Migration: 229_finance_audit_logs
-- Description: Create finance_audit_logs table for general finance tracking

CREATE TABLE IF NOT EXISTS finance_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL,
    entity_id UUID NOT NULL, -- reports, transactions, budgets, etc.
    entity_type VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    actor_id UUID,
    old_values JSONB,
    new_values JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_finance_audit_org ON finance_audit_logs(org_id);
CREATE INDEX idx_finance_audit_entity ON finance_audit_logs(entity_id, entity_type);
CREATE INDEX idx_finance_audit_time ON finance_audit_logs(timestamp);

-- Comments
COMMENT ON TABLE finance_audit_logs IS 'Audit trail for all finance-related changes (expense reports, budgets, etc.)';
