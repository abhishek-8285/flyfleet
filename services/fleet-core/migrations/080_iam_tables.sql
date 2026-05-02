-- 080_iam_tables.sql
-- IAM Tables for Enterprise SaaS

-- 1. Users table (Ensure baseline columns)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE users ADD COLUMN IF NOT EXISTS organization_id UUID;
    ALTER TABLE users ADD COLUMN IF NOT EXISTS email TEXT;
    ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT;
    ALTER TABLE users ADD COLUMN IF NOT EXISTS status TEXT DEFAULT 'active';
    ALTER TABLE users ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
    
    -- Handle unique email constraint
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_name = 'users' AND constraint_name = 'users_email_key') THEN
        ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
    END IF;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- 2. Roles table (Handled defensively in 078, but ensuring baseline here too)
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE roles ADD COLUMN IF NOT EXISTS name VARCHAR(100);
    ALTER TABLE roles ADD COLUMN IF NOT EXISTS organization_id UUID;
    ALTER TABLE roles ADD COLUMN IF NOT EXISTS description TEXT;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- 3. Permissions table (Handled defensively, adding resource/action if missing)
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE permissions ADD COLUMN IF NOT EXISTS resource TEXT;
    ALTER TABLE permissions ADD COLUMN IF NOT EXISTS action TEXT;
    
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'permissions' AND column_name = 'action') THEN
        ALTER TABLE permissions ALTER COLUMN action DROP NOT NULL;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'permissions' AND column_name = 'bit_position') THEN
        ALTER TABLE permissions ALTER COLUMN bit_position DROP NOT NULL;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_name = 'permissions' AND constraint_name = 'uq_permissions_resource_action') THEN
        ALTER TABLE permissions ADD CONSTRAINT uq_permissions_resource_action UNIQUE (resource, action);
    END IF;
EXCEPTION WHEN OTHERS THEN NULL;
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
EXCEPTION WHEN OTHERS THEN NULL;
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
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- Initial Permissions
INSERT INTO permissions (id, resource, action) VALUES 
    (gen_random_uuid(), 'vehicle', 'read'),
    (gen_random_uuid(), 'vehicle', 'write'),
    (gen_random_uuid(), 'driver', 'read'),
    (gen_random_uuid(), 'driver', 'write'),
    (gen_random_uuid(), 'billing', 'read'),
    (gen_random_uuid(), 'billing', 'write')
ON CONFLICT DO NOTHING;
