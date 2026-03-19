-- Replace the default 'user' role with 'driver', though the code enforces valid roles
ALTER TABLE users ALTER COLUMN role DROP DEFAULT;

-- We don't want to break existing rows, but the prompt says org_id MUST BE NOT NULL.
-- Since there might be admins without org_id, let's create a default mock organization 
-- if we are forcing it to NOT NULL, OR just enforce NOT NULL. The instructions say "org_id (NOT NULL)".
DO $$
DECLARE
    default_org uuid;
BEGIN
    SELECT id INTO default_org FROM organizations WHERE name = 'Default Organization' LIMIT 1;
    IF default_org IS NULL THEN
        -- If missing, we create a placeholder org to satisfy the NOT NULL constraint for legacy superadmins
        INSERT INTO organizations (name)
        VALUES ('Default Organization')
        RETURNING id INTO default_org;
    END IF;

    -- Update legacy rows with the default org
    UPDATE users SET org_id = default_org WHERE org_id IS NULL;
END
$$;

-- Now apply the NOT NULL constraint safely
ALTER TABLE users ALTER COLUMN org_id SET NOT NULL;


-- Create the invites table
CREATE TABLE IF NOT EXISTS invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    expiry TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    used BOOLEAN DEFAULT FALSE
);

-- Index for fast invite lookups
CREATE INDEX idx_invites_code ON invites(code);
