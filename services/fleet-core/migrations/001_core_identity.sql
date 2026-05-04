-- Migration: 001_core_identity
-- Description: Establishes a bulletproof identity foundation.

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 0. Global Utilities
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 1. Create 'tenants' table if it doesn't exist
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    tax_id TEXT UNIQUE,
    status TEXT DEFAULT 'active',
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Ensure tax_id exists (Self-healing for pre-created tables)
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'tenants' AND column_name = 'tax_id') THEN
        ALTER TABLE tenants ADD COLUMN tax_id TEXT UNIQUE;
    END IF;
END $$;

-- 2. Handle 'organizations' table/view conflict
-- We use DO block to check if it's a table or view
DO $$ 
BEGIN 
    -- If organizations doesn't exist at all, create it as a view of tenants
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'organizations') THEN
        CREATE VIEW organizations AS SELECT * FROM tenants;
    ELSE
        -- If it exists but it's a table, we just ensure it has a 'System' row to pass early migrations
        IF (SELECT table_type FROM information_schema.tables WHERE table_name = 'organizations') = 'BASE TABLE' THEN
            -- No-op: It's a real table, we let it be. 
            -- But we ensure it has the system ID to satisfy FKs
            INSERT INTO organizations (id, name, status) 
            VALUES ('00000000-0000-0000-0000-000000000000', 'System', 'active')
            ON CONFLICT DO NOTHING;
        END IF;
    END IF;
END $$;

-- 3. Seed System Tenant
INSERT INTO tenants (id, name, tax_id, status) 
VALUES ('00000000-0000-0000-0000-000000000000', 'System', 'SYSTEM-001', 'active')
ON CONFLICT DO NOTHING;

-- 4. Self-healing: Users table must have tenant_id
DO $$ 
BEGIN 
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users') THEN
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'tenant_id') THEN
            ALTER TABLE users ADD COLUMN tenant_id UUID REFERENCES tenants(id);
        END IF;
    END IF;
END $$;
