-- YOLO FIX: Creating the full organizations table
-- YOLO FIX: Adding missing columns to the canonical tenants table
DO $$ 
BEGIN 
    ALTER TABLE tenants ADD COLUMN IF NOT EXISTS type TEXT DEFAULT 'fleet_owner';
    ALTER TABLE tenants ADD COLUMN IF NOT EXISTS logo_url TEXT DEFAULT '';
    ALTER TABLE tenants ADD COLUMN IF NOT EXISTS domain TEXT DEFAULT '';
    ALTER TABLE tenants ADD COLUMN IF NOT EXISTS default_theme TEXT DEFAULT 'system';
    -- status, max_users_limit, max_api_keys_limit are already handled in 098, but let's be safe
    ALTER TABLE tenants ADD COLUMN IF NOT EXISTS status TEXT DEFAULT 'active';
    ALTER TABLE tenants ADD COLUMN IF NOT EXISTS max_users_limit INT DEFAULT 10;
    ALTER TABLE tenants ADD COLUMN IF NOT EXISTS max_api_keys_limit INT DEFAULT 5;
    ALTER TABLE tenants ADD COLUMN IF NOT EXISTS plan TEXT DEFAULT 'enterprise';
    ALTER TABLE tenants ADD COLUMN IF NOT EXISTS contact_email TEXT DEFAULT '';
    ALTER TABLE tenants ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT false;
    ALTER TABLE tenants ADD COLUMN IF NOT EXISTS updated_by UUID;
    ALTER TABLE tenants ADD COLUMN IF NOT EXISTS created_by UUID;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- Ensure missing tables from startup logs exist
CREATE TABLE IF NOT EXISTS governance_settings (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    retention_days INT DEFAULT 30,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
