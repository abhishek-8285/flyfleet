-- Migration: 247_finance_aggregation_and_banking
-- Description: Add bank_transactions and dashboard aggregation view

-- 1. Create bank_transactions table
CREATE TABLE IF NOT EXISTS bank_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL,
    bank_account_id VARCHAR(50),
    transaction_id VARCHAR(100) UNIQUE,
    amount DECIMAL(15, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'INR',
    transaction_date DATE NOT NULL,
    description TEXT,
    category VARCHAR(50),
    status VARCHAR(20) DEFAULT 'UNRECONCILED', -- UNRECONCILED, RECONCILED, IGNORED
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bank_transactions_org_id ON bank_transactions(org_id);
CREATE INDEX IF NOT EXISTS idx_bank_transactions_date ON bank_transactions(transaction_date);
CREATE INDEX IF NOT EXISTS idx_bank_transactions_status ON bank_transactions(status);

-- 2. Create finance_dashboard_summary materialized view
-- Aggregates metrics per org/month
CREATE MATERIALIZED VIEW IF NOT EXISTS finance_dashboard_summary AS
SELECT
    org_id,
    DATE_TRUNC('month', transaction_date)::DATE as month,
    SUM(amount) as total_fuel_cost,
    0.0 as total_other_expenses, -- Placeholder for union/join below
    COUNT(*) as transaction_count
FROM fuel_transactions
GROUP BY org_id, month

UNION ALL

SELECT
    org_id,
    DATE_TRUNC('month', transaction_date)::DATE as month,
    0.0 as total_fuel_cost,
    SUM(amount) as total_other_expenses,
    COUNT(*) as transaction_count
FROM expense_items
GROUP BY org_id, month;

-- Actually, a better materialized view would join them or use a common CTE.
-- Let's redefine it more robustly.
DROP MATERIALIZED VIEW IF EXISTS finance_dashboard_summary;

CREATE MATERIALIZED VIEW finance_dashboard_summary AS
WITH monthly_fuel AS (
    SELECT
        organization_id as org_id,
        DATE_TRUNC('month', transaction_date)::DATE as month,
        SUM(total_cost) as fuel_cost,
        COUNT(*) as fuel_count
    FROM fuel_logs
    GROUP BY 1, 2
),
monthly_expenses AS (
    SELECT
        org_id,
        DATE_TRUNC('month', transaction_date)::DATE as month,
        SUM(amount) as expense_cost,
        COUNT(*) as expense_count
    FROM expense_items
    GROUP BY 1, 2
),
monthly_tds AS (
    SELECT
        org_id,
        DATE_TRUNC('month', deducted_at)::DATE as month,
        SUM(tds_amount) as tds_cost
    FROM tds_records
    GROUP BY 1, 2
)
SELECT
    COALESCE(f.org_id, e.org_id, t.org_id) as org_id,
    COALESCE(f.month, e.month, t.month) as month,
    COALESCE(f.fuel_cost, 0) as total_fuel_cost,
    COALESCE(e.expense_cost, 0) as total_expense_cost,
    COALESCE(t.tds_cost, 0) as total_tds_deducted,
    (COALESCE(f.fuel_cost, 0) + COALESCE(e.expense_cost, 0)) as total_cost,
    COALESCE(f.fuel_count, 0) as fuel_transaction_count,
    COALESCE(e.expense_count, 0) as expense_count,
    NOW() as last_updated
FROM monthly_fuel f
FULL OUTER JOIN monthly_expenses e ON f.org_id = e.org_id AND f.month = e.month
FULL OUTER JOIN monthly_tds t ON COALESCE(f.org_id, e.org_id) = t.org_id AND COALESCE(f.month, e.month) = t.month;

CREATE UNIQUE INDEX IF NOT EXISTS idx_finance_summary_org_month ON finance_dashboard_summary(org_id, month);

-- 3. Trigger for updated_at on bank_transactions
CREATE TRIGGER trg_bank_transactions_updated_at BEFORE UPDATE ON bank_transactions
FOR EACH ROW EXECUTE FUNCTION set_finance_fuel_updated_at();
