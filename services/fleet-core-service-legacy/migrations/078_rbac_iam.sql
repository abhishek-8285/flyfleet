-- SQL Migration: 078_rbac_iam.sql
-- Description: Centralized Identity and Access Management (RBAC) tables

-- 1. Permissions table (Global and Custom)
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    category VARCHAR(50) NOT NULL,
    description TEXT,
    is_global BOOLEAN DEFAULT TRUE,
    organization_id UUID REFERENCES organizations(id), -- Null for global permissions
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Roles table (Global and Custom per Organization)
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    organization_id UUID REFERENCES organizations(id), -- Null for global system roles
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, organization_id)
);

-- 3. Role-Permission mapping
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- 4. User-Role mapping (Supports Multi-Role per Organization)
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id, organization_id)
);

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
ON CONFLICT (name, organization_id) DO NOTHING;
