-- 099_auth_v2_rbac.sql
-- RBAC Authorization: Role-to-Permission mapping for auth_v2

-- Core Permissions Table
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE permissions ADD COLUMN IF NOT EXISTS name TEXT UNIQUE;
    ALTER TABLE permissions ADD COLUMN IF NOT EXISTS category TEXT;
    ALTER TABLE permissions ADD COLUMN IF NOT EXISTS description TEXT;
    ALTER TABLE permissions ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- Role to Permission Mapping
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID,
    permission_id UUID
);

DO $$ 
BEGIN 
    -- Add columns if they don't exist
    ALTER TABLE role_permissions ADD COLUMN IF NOT EXISTS role TEXT;
    ALTER TABLE role_permissions ADD COLUMN IF NOT EXISTS permission_name TEXT;
    
    -- Handle constraints if possible
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'role_permissions' AND column_name = 'permission_name') THEN
        ALTER TABLE role_permissions DROP CONSTRAINT IF EXISTS role_permissions_permission_name_fkey;
        -- We don't enforce REFERENCES here if the table is shared with UUID version
    END IF;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- Seed Initial Permissions
INSERT INTO permissions (name, category, description) VALUES
    ('identity.profile.read', 'identity', 'Read own user profile'),
    ('identity.profile.write', 'identity', 'Update own user profile'),
    ('identity.mfa.manage', 'identity', 'Manage own MFA settings'),
    
    ('org.profile.read', 'org', 'View organization profile and settings'),
    ('org.profile.write', 'org', 'Update organization profile'),
    ('org.users.list', 'org', 'List all users in the organization'),
    ('org.users.manage', 'org', 'Invite, update, or remove users from organization'),
    ('org.billing.read', 'org', 'View billing and subscription status'),
    
    ('fleet.vehicles.list', 'fleet', 'List all vehicles in fleet'),
    ('fleet.vehicles.manage', 'fleet', 'Add or update vehicle information'),
    ('fleet.telemetry.read', 'fleet', 'View real-time and historical telemetry'),
    ('fleet.trips.list', 'fleet', 'View historical trips'),
    
    ('system.audit.read', 'system', 'Read security and activity audit logs')
ON CONFLICT (name) DO NOTHING;

-- Seed Default Role Mappings (Tenant/Org Level)
-- OWNER: Full control within organization
INSERT INTO role_permissions (role, permission_name) VALUES
    ('OWNER', 'identity.profile.read'),
    ('OWNER', 'identity.profile.write'),
    ('OWNER', 'identity.mfa.manage'),
    ('OWNER', 'org.profile.read'),
    ('OWNER', 'org.profile.write'),
    ('OWNER', 'org.users.list'),
    ('OWNER', 'org.users.manage'),
    ('OWNER', 'org.billing.read'),
    ('OWNER', 'fleet.vehicles.list'),
    ('OWNER', 'fleet.vehicles.manage'),
    ('OWNER', 'fleet.telemetry.read'),
    ('OWNER', 'fleet.trips.list'),
    ('OWNER', 'system.audit.read')
ON CONFLICT DO NOTHING;

-- ADMIN: General administration, no billing
INSERT INTO role_permissions (role, permission_name) VALUES
    ('ADMIN', 'identity.profile.read'),
    ('ADMIN', 'identity.profile.write'),
    ('ADMIN', 'identity.mfa.manage'),
    ('ADMIN', 'org.profile.read'),
    ('ADMIN', 'org.users.list'),
    ('ADMIN', 'org.users.manage'),
    ('ADMIN', 'fleet.vehicles.list'),
    ('ADMIN', 'fleet.vehicles.manage'),
    ('ADMIN', 'fleet.telemetry.read'),
    ('ADMIN', 'fleet.trips.list')
ON CONFLICT DO NOTHING;

-- DISPATCHER: Fleet operations
INSERT INTO role_permissions (role, permission_name) VALUES
    ('DISPATCHER', 'identity.profile.read'),
    ('DISPATCHER', 'identity.profile.write'),
    ('DISPATCHER', 'fleet.vehicles.list'),
    ('DISPATCHER', 'fleet.telemetry.read'),
    ('DISPATCHER', 'fleet.trips.list')
ON CONFLICT DO NOTHING;

-- VIEWER: Read-only access
INSERT INTO role_permissions (role, permission_name) VALUES
    ('VIEWER', 'identity.profile.read'),
    ('VIEWER', 'org.profile.read'),
    ('VIEWER', 'fleet.vehicles.list'),
    ('VIEWER', 'fleet.telemetry.read'),
    ('VIEWER', 'fleet.trips.list')
ON CONFLICT DO NOTHING;
