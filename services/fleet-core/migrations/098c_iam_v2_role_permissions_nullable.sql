-- Migration: 098c_iam_v2_role_permissions_nullable
-- Description: Prepare role_permissions for auth_v2 by adding text columns and relaxing UUID constraints.

DO $$ 
BEGIN 
    -- 1. Harmonize role_permissions table
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'role_permissions') THEN
        
        -- Add v2 text columns if they don't exist
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'role_permissions' AND column_name = 'role') THEN
            ALTER TABLE role_permissions ADD COLUMN role TEXT;
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'role_permissions' AND column_name = 'permission_name') THEN
            ALTER TABLE role_permissions ADD COLUMN permission_name TEXT;
        END IF;

        -- Relax role_id and permission_id constraints (allow legacy UUIDs to be NULL)
        IF (SELECT is_nullable FROM information_schema.columns WHERE table_name = 'role_permissions' AND column_name = 'role_id') = 'NO' THEN
            ALTER TABLE role_permissions ALTER COLUMN role_id DROP NOT NULL;
        END IF;

        IF (SELECT is_nullable FROM information_schema.columns WHERE table_name = 'role_permissions' AND column_name = 'permission_id') = 'NO' THEN
            ALTER TABLE role_permissions ALTER COLUMN permission_id DROP NOT NULL;
        END IF;

        -- Add unique index to support ON CONFLICT in migration 099
        IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'role_permissions' AND indexname = 'idx_role_permissions_v2_unique') THEN
            CREATE UNIQUE INDEX idx_role_permissions_v2_unique ON role_permissions (role, permission_name);
        END IF;
    END IF;

    -- 2. Ensure roles table has organization_id/tenant_id harmonization
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'roles') THEN
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'roles' AND column_name = 'tenant_id') THEN
            ALTER TABLE roles ADD COLUMN tenant_id UUID;
        END IF;
    END IF;

END $$;
