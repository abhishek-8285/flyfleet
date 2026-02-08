CREATE TABLE IF NOT EXISTS backups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    backup_type VARCHAR(20) NOT NULL, -- FULL, INCREMENTAL
    status VARCHAR(20) NOT NULL DEFAULT 'SCHEDULED', -- SCHEDULED, RUNNING, COMPLETED, FAILED, VERIFIED
    start_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,
    duration_seconds INTEGER,
    size_bytes BIGINT,
    location VARCHAR(255),
    retention_days INTEGER DEFAULT 30,
    checksum VARCHAR(255),
    initiated_by VARCHAR(255),
    remarks TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS backup_logs (
    id SERIAL PRIMARY KEY,
    backup_id UUID REFERENCES backups(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    result VARCHAR(20) NOT NULL, -- SUCCESS, FAILURE
    actor VARCHAR(255),
    message TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS backup_storage_pools (
    id SERIAL PRIMARY KEY,
    provider VARCHAR(50) NOT NULL, -- AWS, AZURE, ONPREM
    bucket VARCHAR(255) NOT NULL,
    region VARCHAR(50),
    tier VARCHAR(50) DEFAULT 'STANDARD',
    encryption_enabled BOOLEAN DEFAULT TRUE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_backups_status ON backups(status);
CREATE INDEX idx_backups_created_at ON backups(created_at);
CREATE INDEX idx_backup_logs_backup_id ON backup_logs(backup_id);
