-- Invitation Onboarding (Standardized to Tenants)
CREATE TABLE IF NOT EXISTS invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

-- Defensive setup for all columns
DO $$ 
BEGIN 
    -- Column Harmonization (Legacy org_id -> tenant_id)
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'invitations' AND column_name = 'org_id') 
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'invitations' AND column_name = 'tenant_id') THEN
        ALTER TABLE invitations RENAME COLUMN org_id TO tenant_id;
    END IF;

    -- Ensure all columns exist
    ALTER TABLE invitations ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
    ALTER TABLE invitations ADD COLUMN IF NOT EXISTS email VARCHAR(255);
    ALTER TABLE invitations ADD COLUMN IF NOT EXISTS invited_by UUID REFERENCES users(id);
    ALTER TABLE invitations ADD COLUMN IF NOT EXISTS token VARCHAR(512);
    ALTER TABLE invitations ADD COLUMN IF NOT EXISTS role_ids UUID[] DEFAULT '{}';
    ALTER TABLE invitations ADD COLUMN IF NOT EXISTS message TEXT;
    ALTER TABLE invitations ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'pending';
    ALTER TABLE invitations ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP WITH TIME ZONE;
    ALTER TABLE invitations ADD COLUMN IF NOT EXISTS accepted_at TIMESTAMP WITH TIME ZONE;
    ALTER TABLE invitations ADD COLUMN IF NOT EXISTS accepted_user_id UUID REFERENCES users(id);
    ALTER TABLE invitations ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
    ALTER TABLE invitations ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

    -- Constraints (Defensive)
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_name = 'invitations' AND constraint_type = 'UNIQUE' AND constraint_name = 'invitations_token_key') THEN
        -- Check if there are nulls before adding unique
        UPDATE invitations SET token = gen_random_uuid()::text WHERE token IS NULL;
        ALTER TABLE invitations ALTER COLUMN token SET NOT NULL;
        ALTER TABLE invitations ADD CONSTRAINT invitations_token_key UNIQUE (token);
    END IF;
    -- Type Safety (Crucial for RLS)
    IF (SELECT data_type FROM information_schema.columns WHERE table_name = 'invitations' AND column_name = 'tenant_id') IN ('character varying', 'text') THEN
        ALTER TABLE invitations ALTER COLUMN tenant_id TYPE UUID USING tenant_id::uuid;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_invitations_tenant_id ON invitations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_invitations_email ON invitations(email);
CREATE INDEX IF NOT EXISTS idx_invitations_token ON invitations(token);
CREATE INDEX IF NOT EXISTS idx_invitations_status ON invitations(status);
CREATE INDEX IF NOT EXISTS idx_invitations_expires_at ON invitations(expires_at);

ALTER TABLE invitations ENABLE ROW LEVEL SECURITY;

-- Drop legacy policy if exists
DROP POLICY IF EXISTS org_isolation_invitations ON invitations;
DROP POLICY IF EXISTS tenant_isolation_invitations ON invitations;

CREATE POLICY tenant_isolation_invitations ON invitations
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid
           OR current_setting('app.is_super_admin', true)::boolean = true);
