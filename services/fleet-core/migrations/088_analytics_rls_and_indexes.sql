-- Migration: 088_analytics_rls_and_indexes.sql
-- Description: Enable RLS and add composite indexes for multi-tenant isolation.

-- 1. Enable Row Level Security on all Analytics tables
ALTER TABLE predictive_models ENABLE ROW LEVEL SECURITY;
ALTER TABLE prediction_results ENABLE ROW LEVEL SECURITY;
ALTER TABLE analytics_jobs ENABLE ROW LEVEL SECURITY;
ALTER TABLE vehicle_daily_stats ENABLE ROW LEVEL SECURITY;
ALTER TABLE driver_daily_stats ENABLE ROW LEVEL SECURITY;
ALTER TABLE fleet_hourly_stats ENABLE ROW LEVEL SECURITY;
ALTER TABLE analytics_query_logs ENABLE ROW LEVEL SECURITY;

-- 2. Create RLS Policies
-- We assume the database connection sets 'app.org_id' in the session context.
-- If not set, these policies will block all access, which is the secure default.

CREATE POLICY tenant_isolation_models ON predictive_models 
    USING (organization_id = current_setting('app.org_id', true)::UUID);

CREATE POLICY tenant_isolation_results ON prediction_results 
    USING (organization_id = current_setting('app.org_id', true)::UUID);

CREATE POLICY tenant_isolation_jobs ON analytics_jobs 
    USING (organization_id = current_setting('app.org_id', true)::UUID);

CREATE POLICY tenant_isolation_vehicle_stats ON vehicle_daily_stats 
    USING (organization_id = current_setting('app.org_id', true)::UUID);

CREATE POLICY tenant_isolation_driver_stats ON driver_daily_stats 
    USING (organization_id = current_setting('app.org_id', true)::UUID);

CREATE POLICY tenant_isolation_fleet_stats ON fleet_hourly_stats 
    USING (organization_id = current_setting('app.org_id', true)::UUID);

CREATE POLICY tenant_isolation_query_logs ON analytics_query_logs 
    USING (organization_id = current_setting('app.org_id', true)::UUID);

-- 3. Add Composite Indexes for Tenant Optimization
-- Note: Some basic indexes might already exist, but composite ones are better for common query patterns.

CREATE INDEX IF NOT EXISTS idx_predictive_models_org_status ON predictive_models(organization_id, status);

CREATE INDEX IF NOT EXISTS idx_prediction_results_org_model_time ON prediction_results(organization_id, model_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_analytics_jobs_org_status_time ON analytics_jobs(organization_id, status, created_at DESC);

-- Composite indexes for pre-aggregated stats (replacing/enhancing existing ones)
-- We already have (organization_id, date) for stats, but let's ensure they are optimal.
-- Redundancy check: 085 already had some.

-- Audit log optimization
CREATE INDEX IF NOT EXISTS idx_analytics_query_logs_org_type_time ON analytics_query_logs (organization_id, query_type, created_at DESC);
