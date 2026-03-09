-- migration: 053_billing_notifications_and_dunning.sql
SET search_path TO public;

-- 1. Payment Attempts (Dunning tracking)
CREATE TABLE IF NOT EXISTS billing_payment_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    invoice_id UUID NOT NULL REFERENCES billing_invoices(id) ON DELETE CASCADE,
    attempt_number INT NOT NULL DEFAULT 1,
    status VARCHAR(50) NOT NULL, -- e.g., SUCCESS, FAILED, PROCESSING
    failure_reason TEXT,
    processor_response JSONB,
    amount DECIMAL(15, 4) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- RLS Policies
ALTER TABLE billing_payment_attempts ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS payment_attempts_tenant_policy ON billing_payment_attempts;
CREATE POLICY payment_attempts_tenant_policy ON billing_payment_attempts
    FOR ALL TO flyfleet_app
    USING (org_id = get_current_org_id())
    WITH CHECK (org_id = get_current_org_id());

-- Indexes
CREATE INDEX IF NOT EXISTS idx_payment_attempts_invoice ON billing_payment_attempts(invoice_id);
CREATE INDEX IF NOT EXISTS idx_payment_attempts_status ON billing_payment_attempts(status);

-- Permissions
GRANT SELECT, INSERT, UPDATE ON billing_payment_attempts TO flyfleet_app;
