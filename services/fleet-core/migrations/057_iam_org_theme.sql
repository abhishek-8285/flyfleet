-- Migration: 057_iam_org_theme
-- Description: Organization theming for org switch UI (Harmonized with tenants structure).

-- Add theme columns to tenants (organizations view will pick them up)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'tenants') THEN
        ALTER TABLE tenants ADD COLUMN IF NOT EXISTS theme_color VARCHAR(7);
        ALTER TABLE tenants ADD COLUMN IF NOT EXISTS theme_logo_url VARCHAR(512);
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'roles') THEN
            ALTER TABLE tenants ADD COLUMN IF NOT EXISTS default_role_id UUID REFERENCES roles(id);
        END IF;
    END IF;

    -- user_organizations
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'user_organizations') THEN
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'roles') THEN
            ALTER TABLE user_organizations ADD COLUMN IF NOT EXISTS role_id UUID REFERENCES roles(id);
        END IF;
    END IF;
END $$;

-- Org switch audit trail (Use tenants/organizations consistently)
CREATE TABLE IF NOT EXISTS org_switch_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    from_tenant_id UUID REFERENCES tenants(id),
    to_tenant_id UUID NOT NULL REFERENCES tenants(id),
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_org_switch_user_id ON org_switch_events(user_id);
CREATE INDEX IF NOT EXISTS idx_org_switch_created ON org_switch_events(created_at DESC);
