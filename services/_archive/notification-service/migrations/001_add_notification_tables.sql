-- Notifications Module Schema

CREATE TABLE IF NOT EXISTS notification_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    correlation_id VARCHAR(100),
    user_id UUID,
    user_email VARCHAR(255),
    channel VARCHAR(20) NOT NULL,       -- EMAIL, SMS, PUSH, IN_APP
    type VARCHAR(50) NOT NULL,          -- SECURITY, MAINTENANCE, SYSTEM
    subject TEXT,
    body TEXT,
    status VARCHAR(20) NOT NULL,        -- QUEUED, SENT, FAILED, READ
    provider VARCHAR(50),               -- SES, TWILIO, FCM
    provider_response TEXT,
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notif_user ON notification_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_notif_status ON notification_logs(status);
CREATE INDEX IF NOT EXISTS idx_notif_channel ON notification_logs(channel);
CREATE INDEX IF NOT EXISTS idx_notif_created ON notification_logs(created_at);

CREATE TABLE IF NOT EXISTS notification_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID NOT NULL REFERENCES notification_logs(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,    -- QUEUED, SENT, RETRY, FAILED, BOUNCED
    message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS notification_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel VARCHAR(20) UNIQUE NOT NULL, -- EMAIL, SMS, PUSH, IN_APP
    enabled BOOLEAN DEFAULT true,
    max_retries INT DEFAULT 3,
    priority VARCHAR(20) DEFAULT 'STANDARD', -- HIGH, STANDARD, LOW
    retention_months INT DEFAULT 6,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS notification_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    language VARCHAR(10) NOT NULL DEFAULT 'en',
    subject TEXT,
    body TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, channel, language)
);
