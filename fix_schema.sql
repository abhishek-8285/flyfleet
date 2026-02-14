-- Fix missing tables
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS ip_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID, -- Nullable for system-wide rules
    cidr VARCHAR(50) NOT NULL,
    action VARCHAR(20) NOT NULL,
    added_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS access_reviews (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   org_id UUID,
   title VARCHAR(255),
   status VARCHAR(50) DEFAULT 'pending',
   created_by UUID,
   created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
   updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS access_review_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id UUID REFERENCES access_reviews(id) ON DELETE CASCADE,
    user_id UUID,
    status VARCHAR(50) DEFAULT 'pending',
    reviewer_id UUID,
    notes TEXT,
    reviewed_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS login_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID,
    user_id UUID,
    ip_address VARCHAR(50),
    status VARCHAR(20), -- 'success', 'failure'
    os VARCHAR(50),
    browser VARCHAR(50),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS threat_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID,
    type VARCHAR(50),
    severity VARCHAR(20),
    description TEXT,
    status VARCHAR(20) DEFAULT 'open',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP WITH TIME ZONE
);

-- Missing columns in users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS failed_logins INT DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_until TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS mfa_enabled BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS mfa_secret VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS theme VARCHAR(50) DEFAULT 'light';
ALTER TABLE users ADD COLUMN IF NOT EXISTS theme_updated_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'active';
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT false;

-- Ensure users table exists (it should, but just in case)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    role VARCHAR(50) DEFAULT 'user',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Seed Admin User
ALTER TABLE ip_rules ADD COLUMN IF NOT EXISTS description TEXT;

-- Seed Admin User
-- Password is 'admin123'
INSERT INTO users (email, password_hash, full_name, role, is_active)
VALUES (
    'admin@flyfleet.com',
    '$2b$12$k2EpictQdcdbr7tR9N8LdeuosO8ZTzDwir2K3cOc7esxoKS5aFFzm', -- 'admin123'
    'System Admin',
    'admin',
    true
) ON CONFLICT (email) DO UPDATE SET 
    is_active = true,
    password_hash = EXCLUDED.password_hash;
