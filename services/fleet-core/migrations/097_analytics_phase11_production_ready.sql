-- Migration: 097_analytics_phase11_production_ready.sql
-- Description: Implement Access Logs, Feature Lineage, Materialized Views, and Capacity Monitoring.

-- 1. Analytics Access Logs (Compliance & Forensic Audit)
CREATE TABLE IF NOT EXISTS analytics_access_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    user_id UUID NOT NULL,
    resource_id TEXT NOT NULL,
    resource_type TEXT NOT NULL, -- MODEL, REPORT, KPI
    access_context JSONB,
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_analytics_access_logs_org ON analytics_access_logs (organization_id, created_at DESC);
CREATE INDEX idx_analytics_access_logs_user ON analytics_access_logs (user_id, created_at DESC);

-- Enable RLS for Access Logs
ALTER TABLE analytics_access_logs ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_access_logs ON analytics_access_logs 
    USING (organization_id = NULLIF(current_setting('app.org_id', true), '')::UUID);

-- 2. Feature Lineage (Schema Update)
ALTER TABLE ml_feature_store ADD COLUMN IF NOT EXISTS feature_lineage TEXT;

-- 3. Materialized Fleet Reports (Executive Summary)
CREATE TABLE IF NOT EXISTS materialized_fleet_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    report_type TEXT NOT NULL, -- MONTHLY_SUMMARY, QUARTERLY_RISK
    metric_key TEXT NOT NULL,
    metric_value DOUBLE PRECISION NOT NULL,
    period TEXT NOT NULL, -- e.g., '2026-03'
    generated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_materialized_fleet_reports_unique 
ON materialized_fleet_reports (organization_id, report_type, metric_key, period);

-- Enable RLS for Materialized Reports
ALTER TABLE materialized_fleet_reports ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_materialized_reports ON materialized_fleet_reports 
    USING (organization_id = NULLIF(current_setting('app.org_id', true), '')::UUID);

-- 4. Capacity planning view (Internal Monitor)
CREATE OR REPLACE VIEW v_analytics_capacity_stats AS
SELECT 
    schemaname,
    relname as table_name,
    pg_size_pretty(pg_total_relation_size(relid)) as total_size,
    pg_relation_size(relid) as data_size,
    (SELECT count(*) FROM pg_inherits WHERE inhparent = relid) as partition_count
FROM pg_stat_user_tables
WHERE relname IN ('vehicle_daily_stats', 'prediction_results', 'ml_feature_store', 'analytics_query_logs');
