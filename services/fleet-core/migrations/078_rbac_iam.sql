-- SQL Migration: 078_rbac_iam.sql
-- Description: Centralized Identity and Access Management (RBAC) tables (Harmonized to Tenants)

-- 1. Permissions table (Global and Custom)
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    -- Harmonization
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'permissions' AND column_name = 'organization_id') 
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'permissions' AND column_name = 'tenant_id') THEN
        ALTER TABLE permissions RENAME COLUMN organization_id TO tenant_id;
    END IF;

    -- Ensure Columns
    ALTER TABLE permissions ADD COLUMN IF NOT EXISTS name VARCHAR(100);
    ALTER TABLE permissions ADD COLUMN IF NOT EXISTS category VARCHAR(50);
    ALTER TABLE permissions ADD COLUMN IF NOT EXISTS description TEXT;
    ALTER TABLE permissions ADD COLUMN IF NOT EXISTS is_global BOOLEAN DEFAULT TRUE;
    ALTER TABLE permissions ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id);
    ALTER TABLE permissions ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
    
    -- Handle unique constraint for name
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_name = 'permissions' AND constraint_name = 'permissions_name_key') THEN
        ALTER TABLE permissions ADD CONSTRAINT permissions_name_key UNIQUE (name);
    END IF;

    -- Type Safety
    IF (SELECT data_type FROM information_schema.columns WHERE table_name = 'permissions' AND column_name = 'tenant_id') = 'character varying' THEN
        ALTER TABLE permissions ALTER COLUMN tenant_id TYPE UUID USING tenant_id::uuid;
    END IF;
END $$;

-- 2. Roles table (Global and Custom per Organization)
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
    ALTER TABLE roles ADD COLUMN IF NOT EXISTS is_system BOOLEAN DEFAULT FALSE;
    ALTER TABLE roles ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
    ALTER TABLE roles ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
    
    -- Handle unique constraint
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_name = 'roles' AND constraint_name = 'roles_name_tenant_id_key') THEN
        ALTER TABLE roles ADD CONSTRAINT roles_name_tenant_id_key UNIQUE (name, tenant_id);
    END IF;

    -- Type Safety
    IF (SELECT data_type FROM information_schema.columns WHERE table_name = 'roles' AND column_name = 'tenant_id') = 'character varying' THEN
        ALTER TABLE roles ALTER COLUMN tenant_id TYPE UUID USING tenant_id::uuid;
    END IF;
END $$;

-- 3. Role-Permission mapping
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

-- 4. User-Role mapping (Supports Multi-Role per Organization)
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    -- Harmonization
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'user_roles' AND column_name = 'organization_id') 
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'user_roles' AND column_name = 'tenant_id') THEN
        ALTER TABLE user_roles RENAME COLUMN organization_id TO tenant_id;
    END IF;

    -- Ensure Columns
    ALTER TABLE user_roles ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE CASCADE;
    ALTER TABLE user_roles ADD COLUMN IF NOT EXISTS role_id UUID REFERENCES roles(id) ON DELETE CASCADE;
    ALTER TABLE user_roles ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
    ALTER TABLE user_roles ADD COLUMN IF NOT EXISTS granted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
    
    -- Type Safety
    IF (SELECT data_type FROM information_schema.columns WHERE table_name = 'user_roles' AND column_name = 'tenant_id') = 'character varying' THEN
        ALTER TABLE user_roles ALTER COLUMN tenant_id TYPE UUID USING tenant_id::uuid;
    END IF;
END $$;

-- 5. Seed default permissions
INSERT INTO permissions (name, category, description) VALUES
('fleet.vehicle.read', 'vehicle', 'Read vehicle data'),
('fleet.vehicle.write', 'vehicle', 'Create/Update vehicle data'),
('fleet.vehicle.delete', 'vehicle', 'Delete vehicle data'),
('fleet.driver.read', 'driver', 'Read driver data'),
('fleet.driver.write', 'driver', 'Create/Update driver data'),
('fleet.compliance.read', 'compliance', 'Read compliance data'),
('fleet.compliance.write', 'compliance', 'Update compliance status'),
('fleet.finance.read', 'finance', 'Read financial data'),
('fleet.admin.all', 'admin', 'Full administrative access')
ON CONFLICT (name) DO NOTHING;

-- 6. Seed system roles
INSERT INTO roles (name, is_system, description) VALUES
('SuperAdmin', TRUE, 'Global system administrator'),
('OrgAdmin', TRUE, 'Organization administrator'),
('FleetManager', TRUE, 'Manager of fleet operations'),
('Dispatcher', TRUE, 'Dispatcher and route planner'),
('Driver', TRUE, 'Standard driver access'),
('Viewer', TRUE, 'Read-only access')
ON CONFLICT (name, tenant_id) DO NOTHING;
