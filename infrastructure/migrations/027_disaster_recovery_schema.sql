-- Migration: 027_disaster_recovery_schema.sql
-- Description: Schema for Disaster Recovery (DR) module

CREATE TABLE IF NOT EXISTS backups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL, -- FULL, INCREMENTAL, SNAPSHOT
    status VARCHAR(50) NOT NULL, -- PENDING, IN_PROGRESS, COMPLETED, FAILED
    s3_key VARCHAR(255) NOT NULL,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    created_by UUID NOT NULL, -- User ID who initiated backup
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_backups_status ON backups(status);
CREATE INDEX idx_backups_created_at ON backups(created_at);

CREATE TABLE IF NOT EXISTS restore_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    backup_id UUID NOT NULL REFERENCES backups(id),
    status VARCHAR(50) NOT NULL, -- PENDING, IN_PROGRESS, COMPLETED, FAILED
    target_environment VARCHAR(50) NOT NULL, -- PRODUCTION, STAGING, DEV
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    performed_by UUID NOT NULL, -- User ID who initiated restore
    logs TEXT
);

CREATE INDEX idx_restore_jobs_status ON restore_jobs(status);
CREATE INDEX idx_restore_jobs_backup_id ON restore_jobs(backup_id);
