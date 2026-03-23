-- Migration: 234_finance_retention
-- Description: Create finance_retention_policies table

CREATE TABLE IF NOT EXISTS finance_retention_policies (
    organization_id UUID PRIMARY KEY,
    fuel_logs_retention_days INTEGER DEFAULT 2555, -- Default 7 years for Indian tax compliance
    expense_reports_retention_days INTEGER DEFAULT 2555,
    audit_logs_retention_days INTEGER DEFAULT 365,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Comment
COMMENT ON TABLE finance_retention_policies IS 'Confirms how long financial records should be kept per organization';
