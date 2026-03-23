-- Migration: 084_alerts_v2_schema
-- Description: Core schema for the new Alerts & Notifications Module (alerts_v2)

-- 1. Notification Templates
CREATE TABLE IF NOT EXISTS alerts_v2_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    channel VARCHAR(50) NOT NULL, -- EMAIL, SMS, PUSH, WEBHOOK
    language VARCHAR(10) NOT NULL DEFAULT 'en',
    subject TEXT,
    body TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, channel, language)
);

-- 2. User Notification Preferences
CREATE TABLE IF NOT EXISTS alerts_v2_user_preferences (
    user_id UUID NOT NULL,
    channel VARCHAR(50) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, channel)
);

-- 3. Notification Logs (The central ledger)
CREATE TABLE IF NOT EXISTS alerts_v2_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,                     -- Optional if sent to an external recipient
    recipient TEXT NOT NULL,         -- phone, email, device token, or URL
    channel VARCHAR(50) NOT NULL,
    alert_type VARCHAR(100),
    subject TEXT,
    body TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'QUEUED', -- QUEUED, SENDING, SENT, FAILED
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3,
    provider VARCHAR(100),            -- twilio, sendgrid, etc.
    provider_response JSONB,
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    scheduled_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_alerts_v2_logs_status ON alerts_v2_logs(status) WHERE status = 'QUEUED';
CREATE INDEX idx_alerts_v2_logs_user_id ON alerts_v2_logs(user_id);

-- 4. Dead Letter Queue
CREATE TABLE IF NOT EXISTS alerts_v2_dlq (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    log_id UUID REFERENCES alerts_v2_logs(id),
    reason TEXT,
    payload JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
