-- Migration: 233_fuel_transactions_org_id
-- Description: Add org_id to fuel_transactions for partitioning and multi-tenancy

ALTER TABLE fuel_transactions ADD COLUMN IF NOT EXISTS org_id UUID;

-- Index
CREATE INDEX idx_fuel_transactions_org_id ON fuel_transactions(org_id);

-- Comment
COMMENT ON COLUMN fuel_transactions.org_id IS 'Organization ID for multi-tenant isolation';
