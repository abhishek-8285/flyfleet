-- Migration: 096_analytics_phase10_maturity.sql
-- Description: Implement query plan storage, feature versioning enforcement, and performance tuning.

-- 1. Query Plan Storage
CREATE TABLE IF NOT EXISTS analytics_query_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    query_type TEXT NOT NULL,
    query_params JSONB,
    execution_duration_ms BIGINT NOT NULL,
    query_plan TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Enable RLS on query plans
ALTER TABLE analytics_query_plans ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_query_plans ON analytics_query_plans;
CREATE POLICY tenant_isolation_query_plans ON analytics_query_plans 
    USING (organization_id = NULLIF(current_setting('app.org_id', true), '')::UUID);

-- 2. Feature Store Versioning Enforcement
-- Add a unique constraint to prevent duplicate feature/version pairs for the same entity
-- This ensures reproducibility.
CREATE UNIQUE INDEX IF NOT EXISTS idx_ml_feature_store_version_dedup 
ON ml_feature_store (organization_id, entity_type, entity_id, feature_name, feature_version);

-- 3. Adaptive Rate Limiting Config (Initial)
-- In a real scenario, this might be a dynamic table or Redis config.
-- For now, we ensure we have metadata to support it.
COMMENT ON COLUMN analytics_quotas.daily_limit IS 'Dynamic limit that can be adjusted via support module during high load';

-- 4. Streaming Ingestion Support
-- Ensure feature store has an index for the source to track Kafka vs Batch updates
CREATE INDEX IF NOT EXISTS idx_ml_feature_store_source 
ON ml_feature_store (feature_source, feature_timestamp DESC);
