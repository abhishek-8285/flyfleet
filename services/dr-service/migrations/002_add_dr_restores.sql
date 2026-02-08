-- 002_add_dr_restores.sql

CREATE TABLE IF NOT EXISTS dr_restores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    backup_id UUID NOT NULL, -- Assuming foreign key check or loose check for now since backups are separate
    restore_type VARCHAR(20) NOT NULL,  -- FULL / INCREMENTAL
    target_env VARCHAR(20) NOT NULL,    -- PROD / STAGING / UAT
    status VARCHAR(20) NOT NULL,        -- PENDING / RUNNING / COMPLETED / FAILED / CANCELLED
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    initiated_by VARCHAR(255),
    duration_seconds INT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS dr_restore_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    restore_id UUID REFERENCES dr_restores(id),
    level VARCHAR(10),   -- INFO / WARN / ERROR
    module VARCHAR(50),
    message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_dr_restores_created_at ON dr_restores(created_at);
CREATE INDEX idx_dr_restore_logs_restore_id ON dr_restore_logs(restore_id);
CREATE INDEX idx_dr_restore_logs_created_at ON dr_restore_logs(created_at);
