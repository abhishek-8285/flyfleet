-- Standardizing identity model by removing strict constraints on legacy columns (org_id, organization_id).
-- This allows the application to transition fully to tenant_id without being blocked by old schema requirements.

BEGIN;

-- Vehicles
ALTER TABLE vehicles ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE vehicles ALTER COLUMN organization_id DROP NOT NULL;

-- Drivers
ALTER TABLE drivers ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE drivers ALTER COLUMN organization_id DROP NOT NULL;

-- Trips
ALTER TABLE trips ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE trips ALTER COLUMN organization_id DROP NOT NULL;

-- Fuel Logs
ALTER TABLE fuel_logs ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE fuel_logs ALTER COLUMN organization_id DROP NOT NULL;

-- Fuel Transactions
ALTER TABLE fuel_transactions ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE fuel_transactions ALTER COLUMN organization_id DROP NOT NULL;

-- Expense Policies
ALTER TABLE expense_policies ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE expense_policies ALTER COLUMN organization_id DROP NOT NULL;

-- Expense Reports
ALTER TABLE expense_reports ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE expense_reports ALTER COLUMN organization_id DROP NOT NULL;

-- Toll Transactions
ALTER TABLE toll_transactions ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE toll_transactions ALTER COLUMN organization_id DROP NOT NULL;

COMMIT;
