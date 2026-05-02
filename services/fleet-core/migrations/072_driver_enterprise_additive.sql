-- Migration: 072_driver_enterprise_additive
-- Description: Add enterprise columns to drivers table with tenant isolation.

-- 1. Add enterprise columns
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS gender VARCHAR(10);
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS shift_type VARCHAR(20) DEFAULT 'MORNING';
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS behavior_monitoring BOOLEAN DEFAULT TRUE;
ALTER TABLE drivers ADD COLUMN IF NOT EXISTS branch_id UUID;

-- 2. Standardize Indexes
CREATE INDEX IF NOT EXISTS idx_drivers_tenant_status ON drivers(tenant_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_drivers_license_expiry ON drivers(license_expiry) WHERE deleted_at IS NULL;
