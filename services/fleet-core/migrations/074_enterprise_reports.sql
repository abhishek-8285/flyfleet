-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE IF NOT EXISTS reports (
    id UUID PRIMARY KEY
);

DO $$ 
BEGIN 
    ALTER TABLE reports ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL;
    ALTER TABLE reports ADD COLUMN IF NOT EXISTS report_type VARCHAR(100) NOT NULL DEFAULT '';
    ALTER TABLE reports ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'PENDING';
    ALTER TABLE reports ADD COLUMN IF NOT EXISTS parameters JSONB;
    ALTER TABLE reports ADD COLUMN IF NOT EXISTS requested_by UUID NOT NULL;
    ALTER TABLE reports ADD COLUMN IF NOT EXISTS requested_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
    ALTER TABLE reports ADD COLUMN IF NOT EXISTS started_at TIMESTAMP WITH TIME ZONE;
    ALTER TABLE reports ADD COLUMN IF NOT EXISTS completed_at TIMESTAMP WITH TIME ZONE;
    ALTER TABLE reports ADD COLUMN IF NOT EXISTS progress INT DEFAULT 0;
    ALTER TABLE reports ADD COLUMN IF NOT EXISTS file_url TEXT;
    ALTER TABLE reports ADD COLUMN IF NOT EXISTS error_message TEXT;
    ALTER TABLE reports ADD COLUMN IF NOT EXISTS retry_count INT DEFAULT 0;
    ALTER TABLE reports ADD COLUMN IF NOT EXISTS checksum TEXT;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS report_jobs (
    id UUID PRIMARY KEY
);

DO $$ 
BEGIN 
    ALTER TABLE report_jobs ADD COLUMN IF NOT EXISTS report_id UUID NOT NULL REFERENCES reports(id) ON DELETE CASCADE;
    ALTER TABLE report_jobs ADD COLUMN IF NOT EXISTS job_status VARCHAR(50) NOT NULL DEFAULT 'QUEUED';
    ALTER TABLE report_jobs ADD COLUMN IF NOT EXISTS worker_id VARCHAR(100);
    ALTER TABLE report_jobs ADD COLUMN IF NOT EXISTS attempts INT DEFAULT 0;
    ALTER TABLE report_jobs ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
    ALTER TABLE report_jobs ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS report_templates (
    id UUID PRIMARY KEY
);

DO $$ 
BEGIN 
    ALTER TABLE report_templates ADD COLUMN IF NOT EXISTS name VARCHAR(255) NOT NULL DEFAULT '';
    ALTER TABLE report_templates ADD COLUMN IF NOT EXISTS description TEXT;
    ALTER TABLE report_templates ADD COLUMN IF NOT EXISTS query_template TEXT NOT NULL DEFAULT '';
    ALTER TABLE report_templates ADD COLUMN IF NOT EXISTS allowed_filters JSONB;
    ALTER TABLE report_templates ADD COLUMN IF NOT EXISTS output_format VARCHAR(50) NOT NULL DEFAULT 'PDF';
    ALTER TABLE report_templates ADD COLUMN IF NOT EXISTS version INT DEFAULT 1;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS scheduled_reports (
    id UUID PRIMARY KEY
);

DO $$ 
BEGIN 
    ALTER TABLE scheduled_reports ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL;
    ALTER TABLE scheduled_reports ADD COLUMN IF NOT EXISTS report_template_id UUID NOT NULL REFERENCES report_templates(id);
    ALTER TABLE scheduled_reports ADD COLUMN IF NOT EXISTS cron_expression VARCHAR(100) NOT NULL DEFAULT '0 0 * * *';
    ALTER TABLE scheduled_reports ADD COLUMN IF NOT EXISTS parameters JSONB;
    ALTER TABLE scheduled_reports ADD COLUMN IF NOT EXISTS enabled BOOLEAN DEFAULT TRUE;
    ALTER TABLE scheduled_reports ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS report_audit_logs (
    id UUID PRIMARY KEY
);

DO $$ 
BEGIN 
    ALTER TABLE report_audit_logs ADD COLUMN IF NOT EXISTS report_id UUID REFERENCES reports(id) ON DELETE SET NULL;
    ALTER TABLE report_audit_logs ADD COLUMN IF NOT EXISTS user_id UUID NOT NULL;
    ALTER TABLE report_audit_logs ADD COLUMN IF NOT EXISTS action VARCHAR(100) NOT NULL DEFAULT '';
    ALTER TABLE report_audit_logs ADD COLUMN IF NOT EXISTS timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
    ALTER TABLE report_audit_logs ADD COLUMN IF NOT EXISTS ip_address TEXT;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

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
