-- +goose Up
-- Integrations Module (Domain 17) — Full DDL

-- Integration status enum
DO $$ BEGIN
    CREATE TYPE integration_status AS ENUM (
        'created', 'configured', 'active', 'paused', 'failed', 'disabled'
    );
EXCEPTION WHEN duplicate_object THEN null;
END $$;

-- Core integrations registry
CREATE TABLE IF NOT EXISTS integrations (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE integrations ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL;
    ALTER TABLE integrations ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT '';
    ALTER TABLE integrations ADD COLUMN IF NOT EXISTS provider TEXT NOT NULL DEFAULT '';
    ALTER TABLE integrations ADD COLUMN IF NOT EXISTS integration_type TEXT NOT NULL DEFAULT '';
    ALTER TABLE integrations ADD COLUMN IF NOT EXISTS status integration_status NOT NULL DEFAULT 'created';
    ALTER TABLE integrations ADD COLUMN IF NOT EXISTS config_json JSONB;
    ALTER TABLE integrations ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
    ALTER TABLE integrations ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_integrations_tenant    ON integrations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_integrations_status    ON integrations(status);
CREATE INDEX IF NOT EXISTS idx_integrations_provider  ON integrations(tenant_id, provider);

-- Credential vault (AES-256-GCM encrypted)
CREATE TABLE IF NOT EXISTS integration_credentials (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE integration_credentials ADD COLUMN IF NOT EXISTS integration_id UUID NOT NULL REFERENCES integrations(id) ON DELETE CASCADE;
    ALTER TABLE integration_credentials ADD COLUMN IF NOT EXISTS credential_type TEXT NOT NULL DEFAULT '';
    ALTER TABLE integration_credentials ADD COLUMN IF NOT EXISTS encrypted_secret BYTEA NOT NULL DEFAULT '\x'::bytea;
    ALTER TABLE integration_credentials ADD COLUMN IF NOT EXISTS key_version TEXT NOT NULL DEFAULT 'v1';
    ALTER TABLE integration_credentials ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ;
    ALTER TABLE integration_credentials ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_credentials_integration ON integration_credentials(integration_id);

-- Webhook subscriptions
CREATE TABLE IF NOT EXISTS webhook_subscriptions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE webhook_subscriptions ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL;
    ALTER TABLE webhook_subscriptions ADD COLUMN IF NOT EXISTS event_type TEXT NOT NULL DEFAULT '';
    ALTER TABLE webhook_subscriptions ADD COLUMN IF NOT EXISTS target_url TEXT NOT NULL DEFAULT '';
    ALTER TABLE webhook_subscriptions ADD COLUMN IF NOT EXISTS secret TEXT NOT NULL DEFAULT '';
    ALTER TABLE webhook_subscriptions ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active';
    ALTER TABLE webhook_subscriptions ADD COLUMN IF NOT EXISTS retry_policy JSONB NOT NULL DEFAULT '{"max_retries":5,"backoff_seconds":[60,300,900,3600,21600]}';
    ALTER TABLE webhook_subscriptions ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_webhooks_tenant     ON webhook_subscriptions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_webhooks_event_type ON webhook_subscriptions(event_type);
CREATE INDEX IF NOT EXISTS idx_webhooks_status     ON webhook_subscriptions(tenant_id, status);

-- Integration event queue
CREATE TABLE IF NOT EXISTS integration_events (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE integration_events ADD COLUMN IF NOT EXISTS integration_id UUID REFERENCES integrations(id) ON DELETE SET NULL;
    ALTER TABLE integration_events ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL;
    ALTER TABLE integration_events ADD COLUMN IF NOT EXISTS event_type TEXT NOT NULL DEFAULT '';
    ALTER TABLE integration_events ADD COLUMN IF NOT EXISTS schema_version TEXT NOT NULL DEFAULT '1.0';
    ALTER TABLE integration_events ADD COLUMN IF NOT EXISTS trace_id TEXT;
    ALTER TABLE integration_events ADD COLUMN IF NOT EXISTS payload JSONB NOT NULL DEFAULT '{}';
    ALTER TABLE integration_events ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'pending';
    ALTER TABLE integration_events ADD COLUMN IF NOT EXISTS retry_count INT NOT NULL DEFAULT 0;
    ALTER TABLE integration_events ADD COLUMN IF NOT EXISTS retry_after TIMESTAMPTZ;
    ALTER TABLE integration_events ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
    ALTER TABLE integration_events ADD COLUMN IF NOT EXISTS processed_at TIMESTAMPTZ;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_events_tenant     ON integration_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_events_status     ON integration_events(status, retry_after) WHERE status IN ('pending','failed');
CREATE INDEX IF NOT EXISTS idx_events_type       ON integration_events(event_type);
CREATE INDEX IF NOT EXISTS idx_events_created_brin ON integration_events USING BRIN(created_at);

-- Dead Letter Queue
CREATE TABLE IF NOT EXISTS integration_dlq (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE integration_dlq ADD COLUMN IF NOT EXISTS integration_id UUID REFERENCES integrations(id) ON DELETE SET NULL;
    ALTER TABLE integration_dlq ADD COLUMN IF NOT EXISTS event_id UUID REFERENCES integration_events(id) ON DELETE SET NULL;
    ALTER TABLE integration_dlq ADD COLUMN IF NOT EXISTS error_message TEXT;
    ALTER TABLE integration_dlq ADD COLUMN IF NOT EXISTS retry_count INT NOT NULL DEFAULT 0;
    ALTER TABLE integration_dlq ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_dlq_integration ON integration_dlq(integration_id);
CREATE INDEX IF NOT EXISTS idx_dlq_created     ON integration_dlq(created_at);

-- Sync jobs (scheduler)
CREATE TABLE IF NOT EXISTS integration_sync_jobs (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE integration_sync_jobs ADD COLUMN IF NOT EXISTS integration_id UUID REFERENCES integrations(id) ON DELETE CASCADE;
    ALTER TABLE integration_sync_jobs ADD COLUMN IF NOT EXISTS job_type TEXT NOT NULL DEFAULT '';
    ALTER TABLE integration_sync_jobs ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'scheduled';
    ALTER TABLE integration_sync_jobs ADD COLUMN IF NOT EXISTS scheduled_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
    ALTER TABLE integration_sync_jobs ADD COLUMN IF NOT EXISTS started_at TIMESTAMPTZ;
    ALTER TABLE integration_sync_jobs ADD COLUMN IF NOT EXISTS completed_at TIMESTAMPTZ;
    ALTER TABLE integration_sync_jobs ADD COLUMN IF NOT EXISTS error_message TEXT;
    ALTER TABLE integration_sync_jobs ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_sync_jobs_integration ON integration_sync_jobs(integration_id);
CREATE INDEX IF NOT EXISTS idx_sync_jobs_status      ON integration_sync_jobs(status, scheduled_at) WHERE status IN ('scheduled','running');

-- All tables and indexes handled defensively above.
