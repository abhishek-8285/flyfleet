-- Migration: 247_finance_aggregation_and_banking
-- Description: Add bank_transactions and dashboard aggregation view

-- Pre-flight: Ensure identity columns exist for core tables used in this script
DO $$
DECLARE
    r RECORD;
BEGIN
    -- fuel_logs, fuel_transactions, expense_items, tds_records
    FOR r IN (SELECT table_name FROM (VALUES ('fuel_logs'), ('fuel_transactions'), ('expense_items'), ('tds_records')) AS t(table_name))
    LOOP
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = r.table_name) THEN
            IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = r.table_name AND column_name = 'org_id') THEN
                -- Try to find a source to backfill from (tenant_id or organization_id)
                IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = r.table_name AND column_name = 'tenant_id') THEN
                    EXECUTE format('ALTER TABLE %I ADD COLUMN org_id UUID', r.table_name);
                    EXECUTE format('UPDATE %I SET org_id = tenant_id', r.table_name);
                ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = r.table_name AND column_name = 'organization_id') THEN
                    EXECUTE format('ALTER TABLE %I ADD COLUMN org_id UUID', r.table_name);
                    EXECUTE format('UPDATE %I SET org_id = organization_id', r.table_name);
                ELSE
                    EXECUTE format('ALTER TABLE %I ADD COLUMN org_id UUID', r.table_name);
                END IF;
            END IF;
        END IF;
    END LOOP;
END $$;

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

DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'bank_transactions' AND column_name = 'org_id') THEN CREATE INDEX IF NOT EXISTS idx_bank_transactions_org_id ON bank_transactions(org_id); END IF; END $$;
CREATE INDEX IF NOT EXISTS idx_bank_transactions_date ON bank_transactions(transaction_date);
CREATE INDEX IF NOT EXISTS idx_bank_transactions_status ON bank_transactions(status);

-- 2. Create finance_dashboard_summary materialized view
DROP MATERIALIZED VIEW IF EXISTS finance_dashboard_summary;

CREATE MATERIALIZED VIEW finance_dashboard_summary AS
WITH monthly_fuel AS (
    SELECT
        org_id,
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

DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'finance_dashboard_summary' AND column_name = 'org_id') THEN CREATE UNIQUE INDEX IF NOT EXISTS idx_finance_summary_org_month ON finance_dashboard_summary(org_id, month); END IF; END $$;

-- 3. Trigger for updated_at on bank_transactions
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_bank_transactions_updated_at') THEN
        CREATE TRIGGER trg_bank_transactions_updated_at BEFORE UPDATE ON bank_transactions
        FOR EACH ROW EXECUTE FUNCTION set_finance_fuel_updated_at();
    END IF;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;
