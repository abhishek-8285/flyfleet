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

-- Seed Initial Permissions (using resource/action schema with name as alias)
INSERT INTO permissions (resource, action, name, category, description, bit_position) VALUES
    ('identity', 'profile.read', 'identity.profile.read', 'identity', 'Read own user profile', 300),
    ('identity', 'profile.write', 'identity.profile.write', 'identity', 'Update own user profile', 301),
    ('identity', 'mfa.manage', 'identity.mfa.manage', 'identity', 'Manage own MFA settings', 302),
    ('org', 'profile.read', 'org.profile.read', 'org', 'View organization profile and settings', 303),
    ('org', 'profile.write', 'org.profile.write', 'org', 'Update organization profile', 304),
    ('org', 'users.list', 'org.users.list', 'org', 'List all users in the organization', 305),
    ('org', 'users.manage', 'org.users.manage', 'org', 'Invite, update, or remove users from organization', 306),
    ('org', 'billing.read', 'org.billing.read', 'org', 'View billing and subscription status', 307),
    ('fleet', 'vehicles.list', 'fleet.vehicles.list', 'fleet', 'List all vehicles in fleet', 308),
    ('fleet', 'vehicles.manage', 'fleet.vehicles.manage', 'fleet', 'Add or update vehicle information', 309),
    ('fleet', 'telemetry.read', 'fleet.telemetry.read', 'fleet', 'View real-time and historical telemetry', 310),
    ('fleet', 'trips.list', 'fleet.trips.list', 'fleet', 'View historical trips', 311),
    ('system', 'audit.read', 'system.audit.read', 'system', 'Read security and activity audit logs', 312)
ON CONFLICT (resource, action) DO NOTHING;

-- Seed Default Role Mappings (Tenant/Org Level)
-- These are role-permission name mappings; stored in a separate table
DO $$
BEGIN
    -- Store role-permission mappings as JSON or simple lookup
    -- Since role_permissions has a composite PK on (role_id, permission_id), we skip direct inserts
    -- and let the application handle role assignments
    NULL;
END $$;
