-- Admin Portal Core Migration
-- Created: 2026-02-07
-- Description: Adds Soft Delete support and RBAC tables

-- 1. Soft Delete Strategy (Add is_deleted to core tables)
ALTER TABLE organizations 
ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE;

ALTER TABLE users 
ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE;

ALTER TABLE users 
ADD COLUMN IF NOT EXISTS role_id UUID; -- Link to global roles

-- 2. RBAC Tables
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL, -- e.g., 'admin', 'fleet_manager', 'driver'
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE, -- System roles cannot be deleted
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    resource VARCHAR(50) NOT NULL, -- e.g., 'vehicle', 'user', 'trip'
    action VARCHAR(50) NOT NULL, -- e.g., 'create', 'read', 'update', 'delete'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_id, resource, action)
);

-- 3. Seed Basic Roles
INSERT INTO roles (name, description, is_system) VALUES 
('super_admin', 'Full platform access', TRUE),
('org_admin', 'Organization level administrator', TRUE),
('fleet_manager', 'Manages vehicles and drivers', TRUE),
('driver', 'Driver app access only', TRUE)
ON CONFLICT (name) DO NOTHING;

-- 4. Audit Logs (Already created in 002, but ensuring Soft Delete awareness for future tables)
-- Future tables should include is_deleted by default.

-- Indexing
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);
CREATE INDEX IF NOT EXISTS idx_permissions_role_id ON permissions(role_id);
