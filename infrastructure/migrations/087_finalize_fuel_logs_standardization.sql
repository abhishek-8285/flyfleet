-- Migration: 087_finalize_fuel_logs_standardization.sql
-- Description: Drop organization_id from fuel_logs/fuel_transactions and recreate dependent materialized view
-- Created: 2026-03-29
-- Dependencies: 086_finish_organization_id_standardization.sql

SET search_path TO public;

\echo 'Finalizing fuel_logs and fuel_transactions standardization...';

-- ========================
-- Step 1: Drop dependent materialized view
-- ========================
DROP MATERIALIZED VIEW IF EXISTS finance_dashboard_summary;
\echo 'Dropped materialized view finance_dashboard_summary';

-- ========================
-- Step 2: Drop indexes on organization_id
-- ========================
DROP INDEX IF EXISTS idx_fuel_logs_org_id;
\echo 'Dropped index idx_fuel_logs_org_id';

-- ========================
-- Step 3: Drop organization_id column from fuel_logs
-- ========================
ALTER TABLE fuel_logs DROP COLUMN IF EXISTS organization_id;
\echo 'Dropped organization_id from fuel_logs';

-- ========================
-- Step 4: Drop organization_id column from fuel_transactions (if exists)
-- ========================
ALTER TABLE fuel_transactions DROP COLUMN IF EXISTS organization_id;
\echo 'Dropped organization_id from fuel_transactions';

-- ========================
-- Step 5: Recreate materialized view using org_id
-- ========================
CREATE MATERIALIZED VIEW finance_dashboard_summary AS
WITH monthly_fuel AS (
    SELECT fuel_logs.org_id AS org_id,
           date_trunc('month', fuel_logs.transaction_date)::date AS month,
           sum(fuel_logs.total_cost) AS fuel_cost,
           count(*) AS fuel_count
    FROM fuel_logs
    GROUP BY fuel_logs.org_id, date_trunc('month', fuel_logs.transaction_date)::date
),
monthly_expenses AS (
    SELECT expense_items.org_id,
           date_trunc('month', expense_items.transaction_date::timestamp with time zone)::date AS month,
           sum(expense_items.amount) AS expense_cost,
           count(*) AS expense_count
    FROM expense_items
    GROUP BY expense_items.org_id, date_trunc('month', expense_items.transaction_date::timestamp with time zone)::date
),
monthly_tds AS (
    SELECT tds_records.org_id,
           date_trunc('month', tds_records.deducted_at)::date AS month,
           sum(tds_records.tds_amount) AS tds_cost
    FROM tds_records
    GROUP BY tds_records.org_id, date_trunc('month', tds_records.deducted_at)::date
)
SELECT COALESCE(f.org_id, e.org_id, t.org_id) AS org_id,
       COALESCE(f.month, e.month, t.month) AS month,
       COALESCE(f.fuel_cost, 0) AS total_fuel_cost,
       COALESCE(e.expense_cost, 0) AS total_expense_cost,
       COALESCE(t.tds_cost, 0) AS total_tds_deducted,
       COALESCE(f.fuel_cost, 0) + COALESCE(e.expense_cost, 0) AS total_cost,
       COALESCE(f.fuel_count, 0) AS fuel_transaction_count,
       COALESCE(e.expense_count, 0) AS expense_count,
       now() AS last_updated
FROM monthly_fuel f
FULL JOIN monthly_expenses e ON f.org_id = e.org_id AND f.month = e.month
FULL JOIN monthly_tds t ON COALESCE(f.org_id, e.org_id) = t.org_id AND COALESCE(f.month, e.month) = t.month;

\echo 'Recreated materialized view finance_dashboard_summary';

-- ========================
-- Step 6: Recreate index on the materialized view
-- ========================
CREATE UNIQUE INDEX idx_finance_summary_org_month ON finance_dashboard_summary (org_id, month);
\echo 'Created index idx_finance_summary_org_month on finance_dashboard_summary';

-- ========================
-- Step 7: Refresh the materialized view to populate data
-- ========================
REFRESH MATERIALIZED VIEW finance_dashboard_summary;
\echo 'Refreshed finance_dashboard_summary';

-- Verify no organization_id remains on fuel logs
SELECT 'fuel_logs organization_id exists: ' ||
  CASE WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'fuel_logs' AND column_name = 'organization_id')
    THEN 'YES' ELSE 'NO' END as status;
SELECT 'fuel_transactions organization_id exists: ' ||
  CASE WHEN EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'fuel_transactions' AND column_name = 'organization_id')
    THEN 'YES' ELSE 'NO' END as status;

\echo '=====================================================';
\echo 'Fuel logs standardization complete!';
\echo '=====================================================';
