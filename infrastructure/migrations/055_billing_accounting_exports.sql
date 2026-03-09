-- migration: 055_billing_accounting_exports.sql
SET search_path TO public;

-- 1. Accounting Exports (Log of all exports to external systems)
CREATE TABLE IF NOT EXISTS billing_accounting_exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    external_system VARCHAR(50) NOT NULL, -- QUICKBOOKS, XERO, SAP
    export_type VARCHAR(50) NOT NULL, -- LEDGER, INVOICES, PAYMENTS
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING', -- PENDING, COMPLETED, FAILED
    file_key TEXT, -- S3/Storage key of the exported file
    records_count INT NOT NULL DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_by UUID
);

-- RLS
ALTER TABLE billing_accounting_exports ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS accounting_tenant_policy ON billing_accounting_exports;
CREATE POLICY accounting_tenant_policy ON billing_accounting_exports
    FOR ALL TO flyfleet_app
    USING (org_id = get_current_org_id())
    WITH CHECK (org_id = get_current_org_id());

-- Indexes
CREATE INDEX IF NOT EXISTS idx_accounting_org_id ON billing_accounting_exports(org_id);
CREATE INDEX IF NOT EXISTS idx_accounting_status ON billing_accounting_exports(status);

-- Permissions
GRANT SELECT, INSERT, UPDATE ON billing_accounting_exports TO flyfleet_app;
