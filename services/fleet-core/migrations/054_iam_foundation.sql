-- IAM User Management: Foundation Tables
-- Adds missing tables referenced in product spec but not yet created

-- ───────────────────────────────────────────────────────────────────
-- 1. ROLES TABLE
-- ───────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT false,
    permission_mask BIGINT DEFAULT 0,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(tenant_id, name)
);

-- ───────────────────────────────────────────────────────────────────
-- 2. PERMISSIONS TABLE (fine-grained permission catalog)
-- ───────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    bit_position INT UNIQUE NOT NULL,
    UNIQUE(resource, action)
);

-- ───────────────────────────────────────────────────────────────────
-- 3. ROLE_PERMISSIONS (many-to-many junction)
-- ───────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    UNIQUE(role_id, permission_id)
);

-- ───────────────────────────────────────────────────────────────────
-- 4. USER_ROLES (many-to-many: users can have multiple roles)
-- ───────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    assigned_by UUID REFERENCES users(id),
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, role_id, tenant_id)
);

-- ───────────────────────────────────────────────────────────────────
-- 5. MFA_POLICIES (per-org MFA configuration)
-- ───────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS mfa_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    require_mfa BOOLEAN DEFAULT false,
    allowed_methods TEXT[] DEFAULT '{totp,sms}',
    remember_device_days INT DEFAULT 30,
    enforced_for_roles TEXT[] DEFAULT '{org_admin,fleet_manager}',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id)
);

-- ───────────────────────────────────────────────────────────────────
-- 6. MFA_EVENTS (audit trail for MFA events)
-- ───────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS mfa_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    method VARCHAR(20),
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ───────────────────────────────────────────────────────────────────
-- INDEXES
-- ───────────────────────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_roles_tenant_deleted ON roles(tenant_id, deleted_at);
CREATE INDEX IF NOT EXISTS idx_permissions_resource_action ON permissions(resource, action);
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions(permission_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_tenant_id ON user_roles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_mfa_policies_tenant_id ON mfa_policies(tenant_id);
CREATE INDEX IF NOT EXISTS idx_mfa_events_user_id ON mfa_events(user_id);
CREATE INDEX IF NOT EXISTS idx_mfa_events_type ON mfa_events(event_type);

-- ───────────────────────────────────────────────────────────────────
-- RLS Policies
-- ───────────────────────────────────────────────────────────────────
ALTER TABLE roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_roles ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_roles ON roles
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid
           OR current_setting('app.is_super_admin', true)::boolean = true);

CREATE POLICY tenant_isolation_user_roles ON user_roles
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid
           OR current_setting('app.is_super_admin', true)::boolean = true);
