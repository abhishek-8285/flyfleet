-- Migration: 089_analytics_ml_feature_store
-- Description: Create ML Feature Store table for versioned feature vectors.

CREATE TABLE IF NOT EXISTS ml_feature_store (
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    organization_id UUID NOT NULL, -- Legacy compatibility for migrations 094-096
    entity_type TEXT NOT NULL, -- 'VEHICLE', 'DRIVER'
    entity_id UUID NOT NULL,
    feature_name TEXT NOT NULL,
    feature_value DOUBLE PRECISION NOT NULL,
    feature_version TEXT DEFAULT 'v1',
    feature_source TEXT DEFAULT 'batch', -- 'stream', 'batch'
    feature_timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (tenant_id, entity_type, entity_id, feature_name, feature_version, feature_timestamp)
);

CREATE INDEX IF NOT EXISTS idx_ml_feature_store_lookup ON ml_feature_store (tenant_id, entity_type, entity_id, feature_name);
CREATE INDEX IF NOT EXISTS idx_ml_feature_store_org ON ml_feature_store (organization_id);
