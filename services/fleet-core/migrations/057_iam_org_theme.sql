-- Organization theming for org switch UI

ALTER TABLE organizations ADD COLUMN IF NOT EXISTS theme_color VARCHAR(7);
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS theme_logo_url VARCHAR(512);
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS default_role_id UUID REFERENCES roles(id);

-- Add theme to user_organizations for per-user role in org
ALTER TABLE user_organizations ADD COLUMN IF NOT EXISTS role_id UUID REFERENCES roles(id);

-- Org switch audit trail
CREATE TABLE IF NOT EXISTS org_switch_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    from_org_id UUID REFERENCES organizations(id),
    to_org_id UUID NOT NULL REFERENCES organizations(id),
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_org_switch_user_id ON org_switch_events(user_id);
CREATE INDEX IF NOT EXISTS idx_org_switch_created ON org_switch_events(created_at DESC);

ALTER TABLE org_switch_events ENABLE ROW LEVEL SECURITY;

CREATE POLICY org_isolation_org_switch ON org_switch_events
    USING (to_org_id = current_setting('app.current_org_id', true)::uuid
           OR current_setting('app.is_super_admin', true)::boolean = true);
