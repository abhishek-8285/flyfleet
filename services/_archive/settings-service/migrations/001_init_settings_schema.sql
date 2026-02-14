-- Create extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Platform Settings Table
CREATE TABLE platform_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    system_name VARCHAR(255) NOT NULL DEFAULT 'Enterprise Fleet Manager',
    support_email VARCHAR(255) NOT NULL,
    default_timezone VARCHAR(100) NOT NULL DEFAULT 'UTC',
    default_currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    default_language VARCHAR(10) NOT NULL DEFAULT 'en-US',
    white_label_enabled BOOLEAN DEFAULT FALSE,
    smtp_config_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Feature Flags Table
CREATE TABLE feature_flags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    scope VARCHAR(50) NOT NULL DEFAULT 'GLOBAL', -- GLOBAL, ORG
    enabled BOOLEAN DEFAULT FALSE,
    tags TEXT[], -- Array of strings
    created_by UUID,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Localization Settings Table
CREATE TABLE localization_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    language_code VARCHAR(10) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    is_default BOOLEAN DEFAULT FALSE
);

-- Templates Table
CREATE TABLE templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    template_key VARCHAR(255) UNIQUE NOT NULL,
    channel VARCHAR(50) NOT NULL, -- EMAIL, SMS
    subject VARCHAR(255),
    body TEXT NOT NULL,
    variables TEXT[], -- Array of strings
    updated_by UUID,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Notification Channels Table
CREATE TABLE notification_channels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    channel_type VARCHAR(50) UNIQUE NOT NULL, -- EMAIL, SMS
    provider VARCHAR(100) NOT NULL,
    config JSONB, -- Encrypted configuration
    enabled BOOLEAN DEFAULT TRUE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- System Limits Table
CREATE TABLE system_limits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    api_rate_limit INT DEFAULT 10000,
    burst_limit INT DEFAULT 500,
    audit_retention VARCHAR(50) DEFAULT '1y',
    activity_retention VARCHAR(50) DEFAULT '180d'
);

-- Initial Seed Data (Optional, can be in a separate file or handled by app)
INSERT INTO platform_settings (id, support_email) VALUES (uuid_generate_v4(), 'support@fleet-system.com');
INSERT INTO system_limits (id) VALUES (uuid_generate_v4());
