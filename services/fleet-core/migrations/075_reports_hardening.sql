-- +goose Up
-- Phase 2: DLQ table, heartbeat column, NOTIFY trigger for LISTEN/NOTIFY queue

-- Add heartbeat tracking to report_jobs (used for crash recovery)
ALTER TABLE report_jobs ADD COLUMN IF NOT EXISTS last_heartbeat TIMESTAMP WITH TIME ZONE;

-- Dead-letter queue for permanently failed jobs
CREATE TABLE IF NOT EXISTS report_dlq (
    id UUID PRIMARY KEY,
    job_id UUID NOT NULL UNIQUE REFERENCES report_jobs(id) ON DELETE CASCADE,
    reason TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_report_dlq_created ON report_dlq(created_at);

-- NOTIFY trigger: fires pg_notify('report_jobs','new') whenever a job is inserted,
-- enabling workers to use LISTEN/NOTIFY instead of polling every 5 seconds.
CREATE OR REPLACE FUNCTION report_job_notify() RETURNS trigger AS $$
BEGIN
    PERFORM pg_notify('report_jobs', 'new');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_report_job_notify ON report_jobs;
CREATE TRIGGER trg_report_job_notify
    AFTER INSERT ON report_jobs
    FOR EACH ROW EXECUTE FUNCTION report_job_notify();

-- +goose Down
DROP TRIGGER IF EXISTS trg_report_job_notify ON report_jobs;
DROP FUNCTION IF EXISTS report_job_notify();
DROP TABLE IF EXISTS report_dlq;
ALTER TABLE report_jobs DROP COLUMN IF EXISTS last_heartbeat;
