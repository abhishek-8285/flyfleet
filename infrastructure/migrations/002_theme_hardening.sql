-- Enterprise Theme Hardening Migration
-- Created: 2026-02-07
-- Description: Adds theme timestamps, org defaults, and audit logging support

-- Add default_theme to organizations
ALTER TABLE organizations 
ADD COLUMN IF NOT EXISTS default_theme VARCHAR(20) DEFAULT 'light' CHECK (default_theme IN ('light', 'dark', 'system'));

-- Add theme_updated_at to users
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS theme_updated_at TIMESTAMP WITH TIME ZONE;

-- Create Audit Logs table (required for logging theme changes)
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL, -- e.g., USER_THEME_UPDATE
    resource VARCHAR(100) NOT NULL, -- e.g., user_profile
    details JSONB, -- Flexible storage for change details (old_value, new_value)
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for audit logs performance
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);
