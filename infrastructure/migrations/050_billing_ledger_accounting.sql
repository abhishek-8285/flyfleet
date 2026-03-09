-- migration: 050_billing_ledger_accounting.sql
SET search_path TO public;

-- 1. Customer Credit Balance (Wallet)
CREATE TABLE IF NOT EXISTS billing_customer_balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    balance DECIMAL(15, 4) NOT NULL DEFAULT 0.0000,
    credit_limit DECIMAL(15, 4) NOT NULL DEFAULT 0.0000,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE', -- ACTIVE, SUSPENDED
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (org_id, currency)
);

-- 2. Credit Transactions (Granular history of the balance)
CREATE TABLE IF NOT EXISTS billing_credit_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    balance_id UUID NOT NULL REFERENCES billing_customer_balances(id),
    transaction_type VARCHAR(50) NOT NULL, -- REFUND, MANUAL_CREDIT, DEBIT_FOR_INVOICE, PROMOTIONAL
    amount DECIMAL(15, 4) NOT NULL, -- Can be positive or negative
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    reference_type VARCHAR(50), -- e.g., INVOICE, CREDIT_NOTE
    reference_id UUID,          -- link to the source document
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID
);

-- 3. Financial Ledger Entries (Immutable general ledger)
CREATE TABLE IF NOT EXISTS billing_ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    entry_type VARCHAR(50) NOT NULL, -- INVOICE_ISSUED, PAYMENT_RECEIVED, CREDIT_APPLIED, REFUND_PROCESSED, REVENUE_RECOGNIZED
    amount DECIMAL(15, 4) NOT NULL, -- Positive for money in, Negative for money out / obligations
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    invoice_id UUID REFERENCES billing_invoices(id),
    payment_id UUID, -- References payment gateway / external tx
    credit_note_id UUID REFERENCES billing_credit_notes(id),
    ledger_account VARCHAR(50) NOT NULL, -- e.g., 'ACCOUNTS_RECEIVABLE', 'DEFERRED_REVENUE', 'CASH', 'REVENUE'
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 4. Revenue Recognition Schedules
CREATE TABLE IF NOT EXISTS billing_revenue_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    invoice_line_item_id UUID NOT NULL, -- Can point to an invoice item or subscription
    total_amount DECIMAL(15, 4) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING', -- PENDING, IN_PROGRESS, COMPLETED
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 5. Recognized Revenue Entries (amortization of the schedule over time)
CREATE TABLE IF NOT EXISTS billing_recognized_revenue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id UUID NOT NULL REFERENCES billing_revenue_schedules(id),
    amount DECIMAL(15, 4) NOT NULL,
    recognized_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- RLS Policies For Multi-Tenancy
ALTER TABLE billing_customer_balances ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS balances_tenant_policy ON billing_customer_balances;
CREATE POLICY balances_tenant_policy ON billing_customer_balances
    FOR ALL TO flyfleet_app
    USING (org_id = get_current_org_id())
    WITH CHECK (org_id = get_current_org_id());

ALTER TABLE billing_ledger_entries ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS ledger_tenant_policy ON billing_ledger_entries;
CREATE POLICY ledger_tenant_policy ON billing_ledger_entries
    FOR ALL TO flyfleet_app
    USING (org_id = get_current_org_id())
    WITH CHECK (org_id = get_current_org_id());

ALTER TABLE billing_revenue_schedules ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS revenue_schedule_tenant_policy ON billing_revenue_schedules;
CREATE POLICY revenue_schedule_tenant_policy ON billing_revenue_schedules
    FOR ALL TO flyfleet_app
    USING (org_id = get_current_org_id())
    WITH CHECK (org_id = get_current_org_id());

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_ledger_org_id ON billing_ledger_entries(org_id);
CREATE INDEX IF NOT EXISTS idx_ledger_invoice_id ON billing_ledger_entries(invoice_id);
CREATE INDEX IF NOT EXISTS idx_credit_txn_balance_id ON billing_credit_transactions(balance_id);
CREATE INDEX IF NOT EXISTS idx_revenue_sch_org ON billing_revenue_schedules(org_id);

-- Permissions
GRANT SELECT, INSERT, UPDATE ON billing_customer_balances TO flyfleet_app;
GRANT SELECT, INSERT ON billing_credit_transactions TO flyfleet_app;
GRANT SELECT, INSERT ON billing_ledger_entries TO flyfleet_app;
GRANT SELECT, INSERT, UPDATE ON billing_revenue_schedules TO flyfleet_app;
GRANT SELECT, INSERT ON billing_recognized_revenue TO flyfleet_app;
