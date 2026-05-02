-- Migration: 084_alerts_v2_schema
-- Description: Core schema for Alerts Module with mandatory tenant isolation.

-- 1. Notification Templates
CREATE TABLE IF NOT EXISTS alerts_v2_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    channel VARCHAR(50) NOT NULL,
    language VARCHAR(10) NOT NULL DEFAULT 'en',
    subject TEXT,
    body TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, name, channel, language)
);

CREATE INDEX IF NOT EXISTS idx_alerts_templates_tenant ON alerts_v2_templates(tenant_id);

-- 2. User Notification Preferences
CREATE TABLE IF NOT EXISTS alerts_v2_user_preferences (
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    user_id UUID NOT NULL,
    channel VARCHAR(50) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (tenant_id, user_id, channel)
);

-- 3. Notification Logs
CREATE TABLE IF NOT EXISTS alerts_v2_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    user_id UUID,
    recipient TEXT NOT NULL,
    channel VARCHAR(50) NOT NULL,
    alert_type VARCHAR(100),
    subject TEXT,
    body TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'QUEUED',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alerts_logs_tenant ON alerts_v2_logs(tenant_id);
