-- Audit Module Expansion Migration
-- Created: 2026-02-08
-- Description: Expands audit_logs with severity and status, adds audit_settings table

-- 1. Expand audit_logs
ALTER TABLE audit_logs 
ADD COLUMN IF NOT EXISTS severity VARCHAR(20) DEFAULT 'info',
ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'success',
ADD COLUMN IF NOT EXISTS org_id UUID REFERENCES organizations(id) ON DELETE CASCADE;

-- Update existing logs with a default org_id (if possible, but usually logs are fresh)
-- For this project, we'll assume we can link them later or they are already linked via user_id

-- 2. Create Audit Settings table
CREATE TABLE IF NOT EXISTS audit_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
    retention_days_online INTEGER DEFAULT 90,
    retention_days_archival INTEGER DEFAULT 2555, -- 7 years
    masking_rules JSONB DEFAULT '["password", "secret", "token", "phone", "email_local"]',
    tracked_event_types JSONB DEFAULT '["MFA_POLICY_UPDATE", "MFA_EXCEPTION_GRANT", "MFA_OVERRIDE_TOGGLE", "LOGIN_SUCCESS", "LOGIN_FAILED"]',
    alert_on_critical BOOLEAN DEFAULT TRUE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexing for performance
CREATE INDEX IF NOT EXISTS idx_audit_logs_org_id ON audit_logs(org_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_severity ON audit_logs(severity);
