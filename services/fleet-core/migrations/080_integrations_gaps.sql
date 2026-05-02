-- +goose Up
-- +goose StatementBegin

-- 1. Idempotency store for already processed events
CREATE TABLE IF NOT EXISTS integration_processed_events (
    tenant_id UUID NOT NULL,
    event_id UUID NOT NULL,
    processor TEXT NOT NULL,
    PRIMARY KEY (tenant_id, event_id, processor)
);

DO $$ 
BEGIN 
    ALTER TABLE integration_processed_events ADD COLUMN IF NOT EXISTS processed_at TIMESTAMP NOT NULL DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- 2. Webhook jobs (isolated retry queue from integration_events)
CREATE TABLE IF NOT EXISTS webhook_jobs (
    id UUID PRIMARY KEY
);

DO $$ 
BEGIN 
    ALTER TABLE webhook_jobs ADD COLUMN IF NOT EXISTS webhook_id UUID NOT NULL;
    ALTER TABLE webhook_jobs ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL;
    ALTER TABLE webhook_jobs ADD COLUMN IF NOT EXISTS event_id UUID NOT NULL;
    ALTER TABLE webhook_jobs ADD COLUMN IF NOT EXISTS payload JSONB;
    ALTER TABLE webhook_jobs ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'pending';
    ALTER TABLE webhook_jobs ADD COLUMN IF NOT EXISTS retry_count INT DEFAULT 0;
    ALTER TABLE webhook_jobs ADD COLUMN IF NOT EXISTS next_attempt_at TIMESTAMP;
    ALTER TABLE webhook_jobs ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT NOW();
    ALTER TABLE webhook_jobs ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_webhook_jobs_status_next ON webhook_jobs(status, next_attempt_at) WHERE status = 'pending';

-- 3. Webhook delivery audit logs
CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id UUID PRIMARY KEY
);

DO $$ 
BEGIN 
    ALTER TABLE webhook_deliveries ADD COLUMN IF NOT EXISTS webhook_id UUID NOT NULL;
    ALTER TABLE webhook_deliveries ADD COLUMN IF NOT EXISTS job_id UUID NOT NULL;
    ALTER TABLE webhook_deliveries ADD COLUMN IF NOT EXISTS event_id UUID NOT NULL;
    ALTER TABLE webhook_deliveries ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'success';
    ALTER TABLE webhook_deliveries ADD COLUMN IF NOT EXISTS response_code INT;
    ALTER TABLE webhook_deliveries ADD COLUMN IF NOT EXISTS latency_ms INT;
    ALTER TABLE webhook_deliveries ADD COLUMN IF NOT EXISTS error_message TEXT;
    ALTER TABLE webhook_deliveries ADD COLUMN IF NOT EXISTS timestamp TIMESTAMP NOT NULL DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_webhook ON webhook_deliveries(webhook_id, timestamp DESC);

-- 4. Integration Usage Counters (Quotas per integration)
CREATE TABLE IF NOT EXISTS integration_usage_counters (
    integration_id UUID NOT NULL,
    period_start TIMESTAMP NOT NULL,
    PRIMARY KEY (integration_id, period_start)
);

DO $$ 
BEGIN 
    ALTER TABLE integration_usage_counters ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL;
    ALTER TABLE integration_usage_counters ADD COLUMN IF NOT EXISTS events_count BIGINT DEFAULT 0;
    ALTER TABLE integration_usage_counters ADD COLUMN IF NOT EXISTS api_calls_count BIGINT DEFAULT 0;
    ALTER TABLE integration_usage_counters ADD COLUMN IF NOT EXISTS webhooks_count BIGINT DEFAULT 0;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- 5. Integration Audit Logs (Lifecycle and config changes)
CREATE TABLE IF NOT EXISTS integration_audit_logs (
    id UUID PRIMARY KEY
);

DO $$ 
BEGIN 
    ALTER TABLE integration_audit_logs ADD COLUMN IF NOT EXISTS integration_id UUID;
    ALTER TABLE integration_audit_logs ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL;
    ALTER TABLE integration_audit_logs ADD COLUMN IF NOT EXISTS action_type TEXT NOT NULL DEFAULT '';
    ALTER TABLE integration_audit_logs ADD COLUMN IF NOT EXISTS actor_id UUID;
    ALTER TABLE integration_audit_logs ADD COLUMN IF NOT EXISTS details JSONB;
    ALTER TABLE integration_audit_logs ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_integration_audit_tenant ON integration_audit_logs(tenant_id, created_at DESC);

-- Update credentials to support explicit versioning
DO $$ BEGIN ALTER TABLE integration_credentials ADD COLUMN IF NOT EXISTS credential_version INT DEFAULT 1; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE integration_credentials ADD COLUMN IF NOT EXISTS active_version INT DEFAULT 1; EXCEPTION WHEN OTHERS THEN NULL; END $$;

-- 6. Expand DLQ for event recovery
DO $$ BEGIN ALTER TABLE integration_dlq ADD COLUMN IF NOT EXISTS tenant_id UUID; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE integration_dlq ADD COLUMN IF NOT EXISTS event_type TEXT; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE integration_dlq ADD COLUMN IF NOT EXISTS payload JSONB; EXCEPTION WHEN OTHERS THEN NULL; END $$;

-- +goose StatementEnd

-- All tables and indexes handled defensively above.
