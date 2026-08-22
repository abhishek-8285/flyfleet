-- 080_iam_tables.sql
-- IAM Tables for Enterprise SaaS (Harmonized to Tenants)

-- 1. Users table (Ensure baseline columns)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    -- Harmonization
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'organization_id') 
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'tenant_id') THEN
        ALTER TABLE users RENAME COLUMN organization_id TO tenant_id;
    END IF;

    -- Ensure Columns
    ALTER TABLE users ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id);
    ALTER TABLE users ADD COLUMN IF NOT EXISTS email TEXT;
    ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT;
    ALTER TABLE users ADD COLUMN IF NOT EXISTS status TEXT DEFAULT 'active';
    ALTER TABLE users ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
    
    -- Handle unique email constraint
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_name = 'users' AND constraint_name = 'users_email_key') THEN
        ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
    END IF;

    -- Type Safety
    IF (SELECT data_type FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'tenant_id') = 'character varying' THEN
        ALTER TABLE users ALTER COLUMN tenant_id TYPE UUID USING tenant_id::uuid;
    END IF;
END $$;

-- 2. Roles table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    -- Harmonization
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'roles' AND column_name = 'organization_id') 
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'roles' AND column_name = 'tenant_id') THEN
        ALTER TABLE roles RENAME COLUMN organization_id TO tenant_id;
    END IF;

    -- Ensure Columns
    ALTER TABLE roles ADD COLUMN IF NOT EXISTS name VARCHAR(100);
    ALTER TABLE roles ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id);
    ALTER TABLE roles ADD COLUMN IF NOT EXISTS description TEXT;

    -- Type Safety
    IF (SELECT data_type FROM information_schema.columns WHERE table_name = 'roles' AND column_name = 'tenant_id') = 'character varying' THEN
        ALTER TABLE roles ALTER COLUMN tenant_id TYPE UUID USING tenant_id::uuid;
    END IF;
END $$;

-- 3. Permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE permissions ADD COLUMN IF NOT EXISTS resource TEXT;
    ALTER TABLE permissions ADD COLUMN IF NOT EXISTS action TEXT;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_name = 'permissions' AND constraint_name = 'uq_permissions_resource_action') THEN
        ALTER TABLE permissions ADD CONSTRAINT uq_permissions_resource_action UNIQUE (resource, action);
    END IF;
END $$;

-- 4. Role-Permission mapping
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE
);

DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_name = 'role_permissions' AND constraint_name = 'role_permissions_pkey') THEN
        ALTER TABLE role_permissions ADD PRIMARY KEY (role_id, permission_id);
    END IF;
END $$;

-- 5. User-Role mapping
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE
);

DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_name = 'user_roles' AND constraint_name = 'user_roles_pkey') THEN
        ALTER TABLE user_roles ADD PRIMARY KEY (user_id, role_id);
    END IF;
END $$;

-- Initial Permissions (with bit_position to satisfy NOT NULL constraint)
INSERT INTO permissions (resource, action, description, bit_position) VALUES 
    ('vehicle', 'read', 'Read vehicle data', 200),
    ('vehicle', 'write', 'Write vehicle data', 201),
    ('driver', 'read', 'Read driver data', 202),
    ('driver', 'write', 'Write driver data', 203),
    ('billing', 'read', 'Read billing data', 204),
    ('billing', 'write', 'Write billing data', 205)
ON CONFLICT (resource, action) DO NOTHING;
