-- 080_iam_tables.sql
-- IAM Tables for Enterprise SaaS

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    status TEXT DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    organization_id UUID, -- NULL for system roles, otherwise scoped to org
    description TEXT,
    UNIQUE(name, organization_id)
);

CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY,
    resource TEXT NOT NULL,
    action TEXT NOT NULL,
    UNIQUE(resource, action)
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- Initial Permissions
INSERT INTO permissions (id, resource, action) VALUES 
    (gen_random_uuid(), 'vehicle', 'read'),
    (gen_random_uuid(), 'vehicle', 'write'),
    (gen_random_uuid(), 'driver', 'read'),
    (gen_random_uuid(), 'driver', 'write'),
    (gen_random_uuid(), 'billing', 'read'),
    (gen_random_uuid(), 'billing', 'write')
ON CONFLICT DO NOTHING;
