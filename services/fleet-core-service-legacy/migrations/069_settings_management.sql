-- Migration: 069_settings_management
-- Description: Creates tables for platform settings, feature flags, templates, notification channels, system limits, and audit logs

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Create platform_settings table (singleton configuration)
CREATE TABLE IF NOT EXISTS platform_settings (
    id UUID PRIMARY KEY DEFAULT '00000000-0000-0000-0000-000000000001',
    system_name VARCHAR(255) NOT NULL DEFAULT 'Enterprise Fleet Manager',
    support_email VARCHAR(255) NOT NULL,
    default_timezone VARCHAR(100) NOT NULL DEFAULT 'UTC',
    default_currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    default_language VARCHAR(10) NOT NULL DEFAULT 'en-US',
    white_label_enabled BOOLEAN DEFAULT FALSE,
    smtp_config_id UUID,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID NOT NULL,
    
    CONSTRAINT chk_platform_settings_single_row CHECK (id = '00000000-0000-0000-0000-000000000001')
);

-- 2. Create feature_flags table
CREATE TABLE IF NOT EXISTS feature_flags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    scope VARCHAR(50) NOT NULL DEFAULT 'GLOBAL', -- GLOBAL, ORG
    enabled BOOLEAN DEFAULT FALSE,
    tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID NOT NULL
);

-- 3. Create localization_settings table
CREATE TABLE IF NOT EXISTS localization_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    language_code VARCHAR(10) NOT NULL UNIQUE,
    is_active BOOLEAN DEFAULT TRUE,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID NOT NULL
);

-- 4. Create templates table (Email/SMS templates)
CREATE TABLE IF NOT EXISTS templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    template_key VARCHAR(100) NOT NULL UNIQUE,
    channel VARCHAR(50) NOT NULL, -- EMAIL, SMS
    subject VARCHAR(500),
    body TEXT NOT NULL,
    variables TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID NOT NULL
);

-- 5. Create notification_channels table
CREATE TABLE IF NOT EXISTS notification_channels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    channel_type VARCHAR(50) NOT NULL UNIQUE, -- EMAIL, SMS, SLACK
    provider VARCHAR(100) NOT NULL, -- SendGrid, Twilio, Slack
    config JSONB NOT NULL DEFAULT '{}', -- Encrypted config
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID NOT NULL
);

-- 6. Create system_limits table
CREATE TABLE IF NOT EXISTS system_limits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    api_rate_limit INTEGER NOT NULL DEFAULT 10000,
    burst_limit INTEGER NOT NULL DEFAULT 500,
    audit_retention VARCHAR(20) NOT NULL DEFAULT '1y', -- e.g. "1y", "180d"
    activity_retention VARCHAR(20) NOT NULL DEFAULT '180d',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID NOT NULL
);

-- 7. Create settings_audit_logs table
CREATE TABLE IF NOT EXISTS settings_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(100) NOT NULL, -- PLATFORM_SETTINGS, FEATURE_FLAG, TEMPLATE, etc.
    entity_id UUID,
    action VARCHAR(50) NOT NULL, -- CREATE, UPDATE, DELETE, TOGGLE, IMPORT
    actor_id UUID NOT NULL,
    actor_ip INET,
    changes JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_feature_flags_key ON feature_flags(key);
CREATE INDEX IF NOT EXISTS idx_feature_flags_enabled ON feature_flags(enabled);
CREATE INDEX IF NOT EXISTS idx_localization_settings_code ON localization_settings(language_code);
CREATE INDEX IF NOT EXISTS idx_templates_key ON templates(template_key);
CREATE INDEX IF NOT EXISTS idx_notification_channels_type ON notification_channels(channel_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON settings_audit_logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON settings_audit_logs(created_at DESC);

-- Comments
COMMENT ON TABLE platform_settings IS 'Global platform configuration (singleton record)';
COMMENT ON TABLE feature_flags IS 'Feature flag toggles for gradual rollouts';
COMMENT ON TABLE localization_settings IS 'Language/localization configuration';
COMMENT ON TABLE templates IS 'Communication templates (Email/SMS)';
COMMENT ON TABLE notification_channels IS 'External notification provider configurations';
COMMENT ON TABLE system_limits IS 'System-wide rate limits and retention policies';
COMMENT ON TABLE settings_audit_logs IS 'Audit trail for settings changes';

COMMENT ON COLUMN platform_settings.id IS 'Singleton ID - only one row should exist';
COMMENT ON COLUMN notification_channels.config IS 'Encrypted provider configuration (API keys, etc.)';
COMMENT ON COLUMN settings_audit_logs.changes IS 'JSON diff of changes made';

-- Views

-- View for active feature flags
CREATE OR REPLACE VIEW active_feature_flags AS
SELECT id, key, description, scope, enabled, tags
FROM feature_flags
WHERE enabled = TRUE
ORDER BY key;

-- View for active notification channels
CREATE OR REPLACE VIEW active_notification_channels AS
SELECT id, channel_type, provider, enabled
FROM notification_channels
WHERE enabled = TRUE
ORDER BY channel_type;

-- View for recent audit logs (last 30 days)
CREATE OR REPLACE VIEW recent_audit_logs AS
SELECT id, entity_type, entity_id, action, actor_id, actor_ip, changes, created_at
FROM settings_audit_logs
WHERE created_at >= NOW() - INTERVAL '30 days'
ORDER BY created_at DESC;

-- Functions

-- Function to cleanup old audit logs (keep last 1 year)
CREATE OR REPLACE FUNCTION cleanup_old_audit_logs()
RETURNS void AS $$
BEGIN
    DELETE FROM settings_audit_logs
    WHERE created_at < NOW() - INTERVAL '1 year';
END;
$$ LANGUAGE plpgsql;

-- Function to get active feature flag by key
CREATE OR REPLACE FUNCTION get_feature_flag(p_key VARCHAR)
RETURNS BOOLEAN AS $$
DECLARE
    v_enabled BOOLEAN;
BEGIN
    SELECT enabled INTO v_enabled
    FROM feature_flags
    WHERE key = p_key
    LIMIT 1;
    
    RETURN COALESCE(v_enabled, FALSE);
END;
$$ LANGUAGE plpgsql;

-- Insert default platform settings if not exists
INSERT INTO platform_settings (id, system_name, support_email, default_timezone, default_currency, default_language, created_by, updated_by)
VALUES ('00000000-0000-0000-0000-000000000001', 'Enterprise Fleet Manager', 'support@fleet-system.com', 'UTC', 'USD', 'en-US', '00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000')
ON CONFLICT (id) DO NOTHING;

-- Insert default system limits if not exists
INSERT INTO system_limits (id, api_rate_limit, burst_limit, audit_retention, activity_retention, created_by, updated_by)
VALUES (uuid_generate_v4(), 10000, 500, '1y', '180d', '00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000')
ON CONFLICT (id) DO NOTHING;

-- Insert default feature flags if not exists
INSERT INTO feature_flags (id, key, description, scope, enabled, created_by, updated_by)
VALUES 
    (uuid_generate_v4(), 'dark_mode', 'Enable dark mode theme', 'GLOBAL', FALSE, '00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000'),
    (uuid_generate_v4(), 'advanced_analytics', 'Enable advanced analytics dashboard', 'GLOBAL', FALSE, '00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000'),
    (uuid_generate_v4(), 'export_import', 'Enable settings export/import functionality', 'GLOBAL', TRUE, '00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000')
ON CONFLICT (key) DO NOTHING;

-- Insert default templates if not exists
INSERT INTO templates (id, template_key, channel, subject, body, variables, created_by, updated_by)
VALUES 
    (uuid_generate_v4(), 'welcome_email', 'EMAIL', 'Welcome to {{system_name}}', 
     'Welcome {{user_name}}! Your account has been created. Please verify your email address.', 
     '{user_name,verification_link}', '00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000'),
    (uuid_generate_v4(), 'password_reset_email', 'EMAIL', 'Password Reset Request', 
     'You requested a password reset. Click here to reset: {{reset_link}}', 
     '{reset_link,expiry_time}', '00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000')
ON CONFLICT (template_key) DO NOTHING;
