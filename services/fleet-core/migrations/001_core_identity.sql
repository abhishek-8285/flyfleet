-- Migration: 001_core_identity
-- Description: Establishes a bulletproof identity foundation.

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

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

-- 1b. Self-healing: ensure tenants table has all required columns
-- (schema harmonizer may create it with only 'id')
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'tenants' AND column_name = 'name') THEN
        ALTER TABLE tenants ADD COLUMN name TEXT NOT NULL DEFAULT 'System';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'tenants' AND column_name = 'tax_id') THEN
        ALTER TABLE tenants ADD COLUMN tax_id TEXT UNIQUE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'tenants' AND column_name = 'status') THEN
        ALTER TABLE tenants ADD COLUMN status TEXT DEFAULT 'active';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'tenants' AND column_name = 'metadata') THEN
        ALTER TABLE tenants ADD COLUMN metadata JSONB DEFAULT '{}'::jsonb;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'tenants' AND column_name = 'created_at') THEN
        ALTER TABLE tenants ADD COLUMN created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'tenants' AND column_name = 'updated_at') THEN
        ALTER TABLE tenants ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
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
        -- If it exists but it's a table, check its columns
        IF (SELECT table_type FROM information_schema.tables WHERE table_name = 'organizations') = 'BASE TABLE' THEN
            -- Check if it has the required columns (name, status)
            IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'organizations' AND column_name = 'name') THEN
                -- Table exists but is incomplete (e.g. created by schema harmonizer with only id)
                -- Drop it and create as a view of tenants for consistency
                DROP TABLE organizations CASCADE;
                CREATE VIEW organizations AS SELECT * FROM tenants;
            ELSE
                -- Table has proper columns, ensure system row exists
                INSERT INTO organizations (id, name, status) 
                VALUES ('00000000-0000-0000-0000-000000000000', 'System', 'active')
                ON CONFLICT DO NOTHING;
            END IF;
        END IF;
    END IF;
END $$;

-- 3. Create basic users table if it doesn't exist (schema harmonizer may not create it)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4()
);

-- 3b. Seed System Tenant
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
-- 4. Self-healing: Users table must have tenant_id
DO $$ 
BEGIN 
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users') THEN
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'tenant_id') THEN
            ALTER TABLE users ADD COLUMN tenant_id UUID REFERENCES tenants(id);
        END IF;
    END IF;
END $$;
