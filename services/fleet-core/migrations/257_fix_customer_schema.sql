-- Migration 257: Fix Customer Schema & API Key Enforcements
-- Target to ensure the customers table handles programmatic access at scale

-- 1. Ensure extensions for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 2. Create the customers table if it doesn't exist
CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(50),
    api_key VARCHAR(255) UNIQUE,
    api_secret_hash VARCHAR(255),
    address JSONB,
    billing_info JSONB,
    status VARCHAR(50) DEFAULT 'active', -- active, inactive, suspended
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- 3. Add columns if table already exists but columns were missing
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='customers' AND column_name='api_key') THEN
        ALTER TABLE customers ADD COLUMN api_key VARCHAR(255) UNIQUE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='customers' AND column_name='tenant_id') THEN
        ALTER TABLE customers ADD COLUMN tenant_id UUID;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='customers' AND column_name='api_secret_hash') THEN
        ALTER TABLE customers ADD COLUMN api_secret_hash VARCHAR(255);
    END IF;
END $$;

-- 4. Critical: Ensure Indexes for high-throughput authentication lookups
CREATE INDEX IF NOT EXISTS idx_customers_tenant_id ON customers(tenant_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_customers_api_key ON customers(api_key);
