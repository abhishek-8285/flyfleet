-- Migration: 094_analytics_phase8_hardening.sql
-- Description: Add Phase 8 tables and extend RLS policies.

-- 1. Create Model Deployments Table
CREATE TABLE IF NOT EXISTS model_deployments (
    id UUID PRIMARY KEY,
    model_id TEXT NOT NULL,
    organization_id UUID NOT NULL,
    deployment_strategy TEXT NOT NULL, -- 'ab', 'shadow', 'canary'
    traffic_split DOUBLE PRECISION DEFAULT 0.0,
    is_active BOOLEAN DEFAULT true,
    deployed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 2. Create Driver Risk Events Table
CREATE TABLE IF NOT EXISTS driver_risk_events (
    id UUID PRIMARY KEY,
    driver_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    risk_type TEXT NOT NULL,
    score_delta DOUBLE PRECISION NOT NULL,
    metadata JSONB,
    event_time TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Create Analytics Quotas Table
CREATE TABLE IF NOT EXISTS analytics_quotas (
    organization_id UUID PRIMARY KEY,
    daily_limit BIGINT DEFAULT 10000, -- Default ACU limit
    used_credits BIGINT DEFAULT 0,
    last_reset_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 4. Enable RLS on new tables and Feature Store
ALTER TABLE ml_feature_store ENABLE ROW LEVEL SECURITY;
ALTER TABLE model_deployments ENABLE ROW LEVEL SECURITY;
ALTER TABLE driver_risk_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE analytics_quotas ENABLE ROW LEVEL SECURITY;

-- 5. Create RLS Policies for new tables
CREATE POLICY tenant_isolation_features ON ml_feature_store 
    USING (organization_id = current_setting('app.org_id', true)::UUID);

CREATE POLICY tenant_isolation_deployments ON model_deployments 
    USING (organization_id = current_setting('app.org_id', true)::UUID);

CREATE POLICY tenant_isolation_risk_events ON driver_risk_events 
    USING (organization_id = current_setting('app.org_id', true)::UUID);

CREATE POLICY tenant_isolation_quotas ON analytics_quotas 
    USING (organization_id = current_setting('app.org_id', true)::UUID);

-- 6. Indexes for Phase 8 Performance
CREATE INDEX IF NOT EXISTS idx_ml_features_version_org ON ml_feature_store (organization_id, feature_version, feature_timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_model_deployments_active ON model_deployments (organization_id, model_id) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_driver_risk_events_time ON driver_risk_events (driver_id, organization_id, event_time DESC);
