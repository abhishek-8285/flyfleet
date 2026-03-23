-- +goose Up
-- +goose StatementBegin

-- 1. Idempotency store for already processed events
CREATE TABLE integration_processed_events (
    tenant_id UUID NOT NULL,
    event_id UUID NOT NULL,
    processor TEXT NOT NULL,
    processed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, event_id, processor)
);

-- 2. Webhook jobs (isolated retry queue from integration_events)
CREATE TABLE webhook_jobs (
    id UUID PRIMARY KEY,
    webhook_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    event_id UUID NOT NULL,
    payload JSONB,
    status TEXT NOT NULL,
    retry_count INT DEFAULT 0,
    next_attempt_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_webhook_jobs_status_next ON webhook_jobs(status, next_attempt_at) WHERE status = 'pending';

-- 3. Webhook delivery audit logs
CREATE TABLE webhook_deliveries (
    id UUID PRIMARY KEY,
    webhook_id UUID NOT NULL,
    job_id UUID NOT NULL,
    event_id UUID NOT NULL,
    status TEXT NOT NULL,
    response_code INT,
    latency_ms INT,
    error_message TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhook_deliveries_webhook ON webhook_deliveries(webhook_id, timestamp DESC);

-- 4. Integration Usage Counters (Quotas per integration)
CREATE TABLE integration_usage_counters (
    integration_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    period_start TIMESTAMP NOT NULL,
    events_count BIGINT DEFAULT 0,
    api_calls_count BIGINT DEFAULT 0,
    webhooks_count BIGINT DEFAULT 0,
    PRIMARY KEY (integration_id, period_start)
);

-- 5. Integration Audit Logs (Lifecycle and config changes)
CREATE TABLE integration_audit_logs (
    id UUID PRIMARY KEY,
    integration_id UUID,
    tenant_id UUID NOT NULL,
    action_type TEXT NOT NULL,
    actor_id UUID,
    details JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_integration_audit_tenant ON integration_audit_logs(tenant_id, created_at DESC);

-- Update credentials to support explicit versioning
ALTER TABLE integration_credentials ADD COLUMN credential_version INT DEFAULT 1;
ALTER TABLE integration_credentials ADD COLUMN active_version INT DEFAULT 1;

-- 6. Expand DLQ for event recovery
ALTER TABLE integration_dlq ADD COLUMN tenant_id UUID;
ALTER TABLE integration_dlq ADD COLUMN event_type TEXT;
ALTER TABLE integration_dlq ADD COLUMN payload JSONB;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE integration_credentials DROP COLUMN credential_version;
ALTER TABLE integration_credentials DROP COLUMN active_version;

ALTER TABLE integration_dlq DROP COLUMN tenant_id;
ALTER TABLE integration_dlq DROP COLUMN event_type;
ALTER TABLE integration_dlq DROP COLUMN payload;

DROP TABLE integration_audit_logs;
DROP TABLE integration_usage_counters;
DROP TABLE webhook_deliveries;
DROP TABLE webhook_jobs;
DROP TABLE integration_processed_events;

-- +goose StatementEnd
