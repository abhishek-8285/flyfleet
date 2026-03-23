-- Migration: 086_analytics_audit_logging.sql
-- Description: Create audit log table for analytics queries.

CREATE TABLE IF NOT EXISTS analytics_query_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL,
    user_id UUID,
    query_type VARCHAR(50) NOT NULL,
    parameters JSONB,
    duration_ms INT NOT NULL,
    row_count INT,
    cache_hit BOOLEAN DEFAULT FALSE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_analytics_query_logs_org_created ON analytics_query_logs (organization_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_analytics_query_logs_type ON analytics_query_logs (query_type);

-- Comments
COMMENT ON TABLE analytics_query_logs IS 'Audit trail for all analytics query executions';
