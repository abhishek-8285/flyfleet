-- 003_update_users_schema.sql
-- Add new columns for enhanced user management
ALTER TABLE users ADD COLUMN IF NOT EXISTS first_name VARCHAR(100);
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_name VARCHAR(100);
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS org_id UUID REFERENCES organizations(id);

-- Migration logic for existing data
UPDATE users 
SET 
    first_name = CASE 
        WHEN position(' ' in name) > 0 THEN substring(name from 1 for position(' ' in name) - 1)
        ELSE name 
    END,
    last_name = CASE 
        WHEN position(' ' in name) > 0 THEN substring(name from position(' ' in name) + 1)
        ELSE '' 
    END
WHERE (first_name IS NULL OR first_name = '') AND name IS NOT NULL AND name <> '';

-- Link users to their primary organization from mapping table
UPDATE users u
SET org_id = (
    SELECT org_id 
    FROM user_organizations uo 
    WHERE uo.user_id = u.id 
    ORDER BY created_at ASC 
    LIMIT 1
)
WHERE u.org_id IS NULL;

-- Sync deleted_at with is_deleted if exists
UPDATE users SET deleted_at = NOW() WHERE is_deleted = TRUE AND deleted_at IS NULL;

-- Add index for organization search
CREATE INDEX IF NOT EXISTS idx_users_org_id ON users(org_id);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
