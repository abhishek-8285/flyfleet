-- Migration: 095_analytics_phase9_hardening.sql
-- Description: Implement RLS fail-fases, job deduplication, and feature store caching.

-- 1. RLS Fail-Safe Update
-- We strictly enforce that app.org_id must be present for any data access on analytics tables.
-- Update existing policies to ensure they return false (reject) if current_setting is empty.

DROP POLICY IF EXISTS tenant_isolation_models ON predictive_models;
CREATE POLICY tenant_isolation_models ON predictive_models 
    USING (organization_id = NULLIF(current_setting('app.org_id', true), '')::UUID);

DROP POLICY IF EXISTS tenant_isolation_results ON prediction_results;
CREATE POLICY tenant_isolation_results ON prediction_results 
    USING (organization_id = NULLIF(current_setting('app.org_id', true), '')::UUID);

DROP POLICY IF EXISTS tenant_isolation_jobs ON analytics_jobs;
CREATE POLICY tenant_isolation_jobs ON analytics_jobs 
    USING (organization_id = NULLIF(current_setting('app.org_id', true), '')::UUID);

-- 2. Job Deduplication (Idempotency)
-- Add idempotency_key column to analytics_jobs if not exists
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='analytics_jobs' AND column_name='idempotency_key') THEN
        ALTER TABLE analytics_jobs ADD COLUMN idempotency_key TEXT;
    END IF;
END $$;

-- Add unique constraint to prevent duplicate jobs for the same org + idempotency key
-- We use a partial index to allow multiple NULL idempotency keys (if not provided)
CREATE UNIQUE INDEX IF NOT EXISTS idx_analytics_jobs_dedup 
ON analytics_jobs (organization_id, idempotency_key) 
WHERE idempotency_key IS NOT NULL;

-- 3. Feature Store Partition Optimization
-- Ensure latest features are easily accessible
CREATE INDEX IF NOT EXISTS idx_ml_feature_store_latest 
ON ml_feature_store (organization_id, entity_type, entity_id, feature_timestamp DESC);

-- 4. Audit Log Hardening
ALTER TABLE analytics_query_logs ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_query_logs ON analytics_query_logs;
CREATE POLICY tenant_isolation_query_logs ON analytics_query_logs 
    USING (organization_id = NULLIF(current_setting('app.org_id', true), '')::UUID);
