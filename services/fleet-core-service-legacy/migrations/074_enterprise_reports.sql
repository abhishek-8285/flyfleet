-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE IF NOT EXISTS reports (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    report_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL,
    parameters JSONB,
    requested_by UUID NOT NULL,
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    progress INT DEFAULT 0,
    file_url TEXT,
    error_message TEXT,
    retry_count INT DEFAULT 0,
    checksum TEXT
);

CREATE TABLE IF NOT EXISTS report_jobs (
    id UUID PRIMARY KEY,
    report_id UUID NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    job_status VARCHAR(50) NOT NULL,
    worker_id VARCHAR(100),
    attempts INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE IF NOT EXISTS report_templates (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    query_template TEXT NOT NULL,
    allowed_filters JSONB,
    output_format VARCHAR(50) NOT NULL,
    version INT DEFAULT 1
);

CREATE TABLE IF NOT EXISTS scheduled_reports (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    report_template_id UUID NOT NULL REFERENCES report_templates(id),
    cron_expression VARCHAR(100) NOT NULL,
    parameters JSONB,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE IF NOT EXISTS report_audit_logs (
    id UUID PRIMARY KEY,
    report_id UUID REFERENCES reports(id) ON DELETE SET NULL,
    user_id UUID NOT NULL,
    action VARCHAR(100) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    ip_address TEXT
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_reports_tenant_id ON reports(tenant_id);
CREATE INDEX IF NOT EXISTS idx_reports_status ON reports(status);
CREATE INDEX IF NOT EXISTS idx_report_jobs_status ON report_jobs(job_status);
CREATE INDEX IF NOT EXISTS idx_scheduled_reports_tenant_id ON scheduled_reports(tenant_id);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE IF EXISTS report_audit_logs;
DROP TABLE IF EXISTS scheduled_reports;
DROP TABLE IF EXISTS report_templates;
DROP TABLE IF EXISTS report_jobs;
DROP TABLE IF EXISTS reports;
