-- Migration: 228_fuel_vendor_management
-- Description: Create fuel_vendors table for vendor management

CREATE TABLE IF NOT EXISTS fuel_vendors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    gstin VARCHAR(15),
    address TEXT,
    contact_phone VARCHAR(20),
    email VARCHAR(255),
    status VARCHAR(20) DEFAULT 'ACTIVE', -- ACTIVE, INACTIVE, BLACKLISTED
    rating DECIMAL(3, 2) DEFAULT 0.0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Link fuel logs and transactions to vendors
ALTER TABLE fuel_logs ADD COLUMN IF NOT EXISTS vendor_id UUID;
ALTER TABLE fuel_transactions ADD COLUMN IF NOT EXISTS vendor_id UUID;

-- Indexes
CREATE INDEX IF NOT EXISTS idx_fuel_vendors_tenant ON fuel_vendors(tenant_id);
CREATE INDEX IF NOT EXISTS idx_fuel_vendors_gstin ON fuel_vendors(gstin);

-- Comments
COMMENT ON TABLE fuel_vendors IS 'Fuel vendors and stations for procurement management';
COMMENT ON COLUMN fuel_vendors.gstin IS 'GST Identification Number of the vendor';
