CREATE TABLE IF NOT EXISTS bulk_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('processing', 'completed', 'failed', 'partial_success')),
    total_records INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    failed_count INTEGER DEFAULT 0,
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS bulk_job_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES bulk_jobs(id) ON DELETE CASCADE,
    row_number INTEGER NOT NULL,
    record_data JSONB,
    status VARCHAR(50) NOT NULL CHECK (status IN ('success', 'failed')),
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS bulk_config (
    organization_id UUID PRIMARY KEY,
    max_records_per_file INTEGER DEFAULT 5000,
    concurrent_threads INTEGER DEFAULT 4,
    retention_months INTEGER DEFAULT 6,
    mandatory_fields JSONB DEFAULT '["vin", "registration_no", "make", "model_year"]'::jsonb,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_bulk_jobs_org ON bulk_jobs(organization_id);
CREATE INDEX idx_bulk_jobs_status ON bulk_jobs(status);
CREATE INDEX idx_bulk_job_records_job_id ON bulk_job_records(job_id);
