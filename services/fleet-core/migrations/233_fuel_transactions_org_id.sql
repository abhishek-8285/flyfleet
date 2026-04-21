-- Migration: 233_fuel_transactions_tenant_id
-- Description: Add tenant_id to fuel_transactions for partitioning and multi-tenancy

ALTER TABLE fuel_transactions ADD COLUMN IF NOT EXISTS tenant_id UUID;

-- Index
CREATE INDEX idx_fuel_transactions_tenant_id ON fuel_transactions(tenant_id);

-- Comment
COMMENT ON COLUMN fuel_transactions.tenant_id IS 'Tenant ID for multi-tenant isolation';
