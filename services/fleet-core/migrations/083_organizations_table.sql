-- Migration: 083_organizations_table
-- Description: Protective no-op migration. Organization identity is managed via 'tenants' table (aliased as 'organizations' view in 001).

-- 1. Ensure System Tenant exists in the canonical table
INSERT INTO tenants (id, name, status)
VALUES ('00000000-0000-0000-0000-000000000000', 'System', 'active')
ON CONFLICT (id) DO NOTHING;

-- 2. Add foreign keys to the canonical 'tenants' table if missing
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'fk_user_tenant') THEN
        ALTER TABLE users ADD CONSTRAINT fk_user_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id);
    END IF;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;
