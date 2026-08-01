CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ------------------------------------------------------------
-- Tenants
-- ------------------------------------------------------------

CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'active',
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER set_tenants_updated_at
BEFORE UPDATE ON tenants
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- ------------------------------------------------------------
-- Users / recipients
-- ------------------------------------------------------------

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    external_user_id TEXT NOT NULL,
    email TEXT,
    phone TEXT,
    locale TEXT,
    timezone TEXT,
    attributes JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, external_user_id)
);

CREATE INDEX idx_users_tenant_external_id ON users (tenant_id, external_user_id);
CREATE INDEX idx_users_email ON users (tenant_id, email);
CREATE INDEX idx_users_phone ON users (tenant_id, phone);

CREATE TRIGGER set_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- ------------------------------------------------------------
-- API keys
-- ------------------------------------------------------------

CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name TEXT NOT NULL,
    key_prefix TEXT NOT NULL,
    key_hash TEXT NOT NULL UNIQUE,
    scopes TEXT[] NOT NULL DEFAULT '{}',
    environment TEXT NOT NULL DEFAULT 'development',
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_api_keys_tenant ON api_keys (tenant_id);

-- ------------------------------------------------------------
-- Events
-- ------------------------------------------------------------

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    event_type TEXT NOT NULL,
    user_id TEXT,
    idempotency_key TEXT,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL DEFAULT 'received',
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ,
    UNIQUE (tenant_id, idempotency_key)
);

CREATE INDEX idx_events_tenant_type ON events (tenant_id, event_type);
CREATE INDEX idx_events_status ON events (status);
CREATE INDEX idx_events_received_at ON events (received_at DESC);

-- ------------------------------------------------------------
-- Transactional outbox
-- ------------------------------------------------------------

CREATE TABLE outbox_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    aggregate_type TEXT NOT NULL,
    aggregate_id TEXT NOT NULL,
    topic TEXT NOT NULL,
    payload JSONB NOT NULL,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_outbox_unpublished
ON outbox_messages (created_at ASC)
WHERE published_at IS NULL;

-- ------------------------------------------------------------
-- Templates
-- ------------------------------------------------------------

CREATE TABLE templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    key TEXT NOT NULL,
    name TEXT NOT NULL,
    channel TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'transactional',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, key, channel)
);

CREATE TRIGGER set_templates_updated_at
BEFORE UPDATE ON templates
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE template_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES templates(id),
    version INT NOT NULL,
    subject TEXT,
    html_body TEXT,
    text_body TEXT,
    json_body JSONB,
    variables_schema JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (template_id, version)
);

CREATE TABLE template_localizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_version_id UUID NOT NULL REFERENCES template_versions(id),
    locale TEXT NOT NULL,
    subject TEXT,
    html_body TEXT,
    text_body TEXT,
    json_body JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (template_version_id, locale)
);

-- ------------------------------------------------------------
-- Providers
-- ------------------------------------------------------------

CREATE TABLE providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type TEXT NOT NULL,
    channel TEXT NOT NULL,
    name TEXT NOT NULL,
    capabilities JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (type, channel)
);

CREATE TABLE provider_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    provider_id UUID NOT NULL REFERENCES providers(id),
    name TEXT NOT NULL,
    credentials_encrypted BYTEA,
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT true,
    priority INT NOT NULL DEFAULT 100,
    weight INT NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, provider_id, name)
);

CREATE TRIGGER set_provider_configs_updated_at
BEFORE UPDATE ON provider_configs
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE provider_health (
    provider_config_id UUID PRIMARY KEY REFERENCES provider_configs(id),
    status TEXT NOT NULL DEFAULT 'unknown',
    latency_ms INT,
    error_rate REAL,
    last_success_at TIMESTAMPTZ,
    last_failure_at TIMESTAMPTZ,
    last_error TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ------------------------------------------------------------
-- Notifications
-- ------------------------------------------------------------

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    event_id UUID REFERENCES events(id),
    template_id UUID REFERENCES templates(id),
    channel TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'transactional',
    status TEXT NOT NULL DEFAULT 'queued',
    recipient_user_id UUID REFERENCES users(id),
    to_address TEXT,
    subject TEXT,
    body TEXT,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    error TEXT,
    scheduled_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    opened_at TIMESTAMPTZ,
    clicked_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notifications_tenant_status ON notifications (tenant_id, status);
CREATE INDEX idx_notifications_tenant_channel ON notifications (tenant_id, channel);
CREATE INDEX idx_notifications_recipient ON notifications (tenant_id, recipient_user_id);
CREATE INDEX idx_notifications_created_at ON notifications (created_at DESC);

CREATE TRIGGER set_notifications_updated_at
BEFORE UPDATE ON notifications
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE delivery_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID NOT NULL REFERENCES notifications(id),
    provider_config_id UUID REFERENCES provider_configs(id),
    status TEXT NOT NULL,
    attempt INT NOT NULL,
    request JSONB,
    response JSONB,
    error TEXT,
    latency_ms INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_delivery_attempts_notification ON delivery_attempts (notification_id);

-- ------------------------------------------------------------
-- Inbox
-- ------------------------------------------------------------

CREATE TABLE inbox_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    user_id UUID NOT NULL REFERENCES users(id),
    notification_id UUID REFERENCES notifications(id),
    title TEXT NOT NULL,
    body TEXT,
    category TEXT,
    action_url TEXT,
    image_url TEXT,
    data JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_read BOOLEAN NOT NULL DEFAULT false,
    is_archived BOOLEAN NOT NULL DEFAULT false,
    is_pinned BOOLEAN NOT NULL DEFAULT false,
    read_at TIMESTAMPTZ,
    archived_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_inbox_user_unread
ON inbox_items (user_id, is_read, created_at DESC);

CREATE INDEX idx_inbox_user_archived
ON inbox_items (user_id, is_archived, created_at DESC);

-- ------------------------------------------------------------
-- Preferences
-- ------------------------------------------------------------

CREATE TABLE user_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    user_id UUID NOT NULL REFERENCES users(id),
    category TEXT NOT NULL,
    channel TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, category, channel)
);

CREATE TABLE quiet_hours (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    enabled BOOLEAN NOT NULL DEFAULT false,
    start_time TEXT NOT NULL DEFAULT '22:00',
    end_time TEXT NOT NULL DEFAULT '08:00',
    timezone TEXT NOT NULL DEFAULT 'UTC',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE suppressions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    channel TEXT NOT NULL,
    address TEXT NOT NULL,
    reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, channel, address)
);

-- ------------------------------------------------------------
-- Audit logs
-- ------------------------------------------------------------

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    actor_type TEXT NOT NULL,
    actor_id TEXT,
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_tenant_resource
ON audit_logs (tenant_id, resource_type, resource_id);

CREATE INDEX idx_audit_created_at
ON audit_logs (created_at DESC);
