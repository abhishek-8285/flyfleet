-- Schema Synchronization Script for Auth Service
-- Aligning DB with Go models/repository

-- 1. Organizations Hardening
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS default_theme VARCHAR(50) DEFAULT 'light';

-- 2. Roles Hardening
ALTER TABLE roles ADD COLUMN IF NOT EXISTS organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE;

-- 3. Permissions Table Revision
-- Drop existing unique constraint to allow modifications
ALTER TABLE permissions DROP CONSTRAINT IF EXISTS permissions_role_id_resource_action_key;
ALTER TABLE permissions DROP COLUMN IF EXISTS role_id;
ALTER TABLE permissions RENAME COLUMN resource TO module;
ALTER TABLE permissions ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE permissions ADD CONSTRAINT permissions_module_action_key UNIQUE (module, action);

-- 4. Many-to-Many Linking Tables
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    assigned_by UUID, -- Can be NULL for system assignment
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id)
);

-- 5. User Organizations Enhancements
ALTER TABLE user_organizations ADD COLUMN IF NOT EXISTS role VARCHAR(50) DEFAULT 'member';
