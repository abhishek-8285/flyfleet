-- Up
CREATE TABLE scheduled_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,

    name VARCHAR(255) NOT NULL,
    report_type VARCHAR(100) NOT NULL,

    format VARCHAR(20) NOT NULL,        -- PDF, CSV, XLSX

    frequency VARCHAR(50) NOT NULL,     -- DAILY, WEEKLY, MONTHLY
    cron_expression VARCHAR(100) NOT NULL,

    delivery_time TIME NOT NULL,
    timezone VARCHAR(50) NOT NULL,

    filters JSONB NOT NULL DEFAULT '{}'::jsonb,

    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',

    created_by UUID NOT NULL,
    updated_by UUID,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE report_recipients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID REFERENCES scheduled_reports(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE report_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID REFERENCES scheduled_reports(id),

    started_at TIMESTAMP WITH TIME ZONE,
    finished_at TIMESTAMP WITH TIME ZONE,

    duration_seconds INT,

    status VARCHAR(20),  -- SUCCESS / FAILED
    error_message TEXT,

    file_url TEXT,
    file_size BIGINT,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE report_global_settings (
    id SERIAL PRIMARY KEY,

    max_concurrent_jobs INT DEFAULT 50,
    max_retry_attempts INT DEFAULT 3,
    retention_days INT DEFAULT 90,

    smtp_host VARCHAR(255),
    smtp_port INT,
    smtp_user VARCHAR(255),

    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

/*
-- Down
DROP TABLE report_global_settings;
DROP TABLE report_executions;
DROP TABLE report_recipients;
DROP TABLE scheduled_reports;
*/
