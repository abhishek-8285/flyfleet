-- YOLO FIX: Creating the full organizations table
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT DEFAULT 'fleet_owner',
    logo_url TEXT DEFAULT '',
    domain TEXT DEFAULT '',
    default_theme TEXT DEFAULT 'system',
    status TEXT DEFAULT 'active',
    max_users_limit INT DEFAULT 10,
    max_api_keys_limit INT DEFAULT 5,
    plan TEXT DEFAULT 'enterprise',
    contact_email TEXT DEFAULT '',
    is_deleted BOOLEAN DEFAULT false,
    tax_id TEXT UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID
);

-- Ensure missing tables from startup logs exist
CREATE TABLE IF NOT EXISTS governance_settings (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    retention_days INT DEFAULT 30,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
