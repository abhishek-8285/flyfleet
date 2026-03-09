-- +goose Up
-- Integrations Module (Domain 17) — Full DDL

-- Integration status enum
CREATE TYPE integration_status AS ENUM (
    'created', 'configured', 'active', 'paused', 'failed', 'disabled'
);

-- Core integrations registry
CREATE TABLE integrations (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        UUID NOT NULL,
    name             TEXT NOT NULL,
    provider         TEXT NOT NULL,           -- sap, salesforce, stripe, etc.
    integration_type TEXT NOT NULL,           -- erp, insurance, telematics, payment, etc.
    status           integration_status NOT NULL DEFAULT 'created',
    config_json      JSONB,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_integrations_tenant    ON integrations(tenant_id);
CREATE INDEX idx_integrations_status    ON integrations(status);
CREATE INDEX idx_integrations_provider  ON integrations(tenant_id, provider);

-- Credential vault (AES-256-GCM encrypted)
-- encrypted_secret contains: nonce (12B) || ciphertext
CREATE TABLE integration_credentials (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_id   UUID NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    credential_type  TEXT NOT NULL,           -- api_key, oauth_token, basic_auth, mtls_cert
    encrypted_secret BYTEA NOT NULL,          -- AES-256-GCM: nonce(12B)||ciphertext
    key_version      TEXT NOT NULL DEFAULT 'v1', -- for KMS key rotation
    expires_at       TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_credentials_integration ON integration_credentials(integration_id);

-- Webhook subscriptions
CREATE TABLE webhook_subscriptions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL,
    event_type   TEXT NOT NULL,
    target_url   TEXT NOT NULL,
    secret       TEXT NOT NULL,               -- HMAC-SHA256 signing secret
    status       TEXT NOT NULL DEFAULT 'active',
    retry_policy JSONB NOT NULL DEFAULT '{"max_retries":5,"backoff_seconds":[60,300,900,3600,21600]}',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhooks_tenant     ON webhook_subscriptions(tenant_id);
CREATE INDEX idx_webhooks_event_type ON webhook_subscriptions(event_type);
CREATE INDEX idx_webhooks_status     ON webhook_subscriptions(tenant_id, status);

-- Integration event queue
CREATE TABLE integration_events (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_id UUID REFERENCES integrations(id) ON DELETE SET NULL,
    tenant_id      UUID NOT NULL,
    event_type     TEXT NOT NULL,
    schema_version TEXT NOT NULL DEFAULT '1.0',
    trace_id       TEXT,
    payload        JSONB NOT NULL,
    status         TEXT NOT NULL DEFAULT 'pending', -- pending, processing, delivered, failed
    retry_count    INT NOT NULL DEFAULT 0,
    retry_after    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at   TIMESTAMPTZ
);

CREATE INDEX idx_events_tenant     ON integration_events(tenant_id);
CREATE INDEX idx_events_status     ON integration_events(status, retry_after) WHERE status IN ('pending','failed');
CREATE INDEX idx_events_type       ON integration_events(event_type);
-- BRIN index for time-range queries on large event tables
CREATE INDEX idx_events_created_brin ON integration_events USING BRIN(created_at);

-- Dead Letter Queue
CREATE TABLE integration_dlq (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_id UUID REFERENCES integrations(id) ON DELETE SET NULL,
    event_id       UUID REFERENCES integration_events(id) ON DELETE SET NULL,
    error_message  TEXT,
    retry_count    INT NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_dlq_integration ON integration_dlq(integration_id);
CREATE INDEX idx_dlq_created     ON integration_dlq(created_at);

-- Sync jobs (scheduler)
CREATE TABLE integration_sync_jobs (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_id UUID REFERENCES integrations(id) ON DELETE CASCADE,
    job_type       TEXT NOT NULL,             -- erp_daily, vehicle_hourly, compliance_weekly
    status         TEXT NOT NULL DEFAULT 'scheduled', -- scheduled, running, done, failed
    scheduled_at   TIMESTAMPTZ NOT NULL,
    started_at     TIMESTAMPTZ,
    completed_at   TIMESTAMPTZ,
    error_message  TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sync_jobs_integration ON integration_sync_jobs(integration_id);
CREATE INDEX idx_sync_jobs_status      ON integration_sync_jobs(status, scheduled_at) WHERE status IN ('scheduled','running');

-- +goose Down
DROP TABLE IF EXISTS integration_sync_jobs;
DROP TABLE IF EXISTS integration_dlq;
DROP TABLE IF EXISTS integration_events;
DROP TABLE IF EXISTS webhook_subscriptions;
DROP TABLE IF EXISTS integration_credentials;
DROP TABLE IF EXISTS integrations;
DROP TYPE IF EXISTS integration_status;
