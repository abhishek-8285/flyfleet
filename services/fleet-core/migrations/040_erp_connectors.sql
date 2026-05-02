-- Migration: 040_erp_connectors
-- Description: Creates tables for ERP integration configurations, sync jobs, and ID mappings.

-- 1. Create erp_configs table
CREATE TABLE IF NOT EXISTS erp_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    endpoint_url TEXT NOT NULL,
    auth_config JSONB NOT NULL,
    settings JSONB DEFAULT '{}'::jsonb,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_erp_configs_tenant ON erp_configs(tenant_id);

-- 2. Create sync_jobs table
CREATE TABLE IF NOT EXISTS sync_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    erp_config_id UUID REFERENCES erp_configs(id),
    entity_type TEXT NOT NULL,
    direction TEXT NOT NULL,
    status TEXT DEFAULT 'PENDING',
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    records_processed INT DEFAULT 0,
    error_log TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sync_jobs_tenant ON sync_jobs(tenant_id);

-- 3. Create id_mappings table
CREATE TABLE IF NOT EXISTS id_mappings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    erp_config_id UUID REFERENCES erp_configs(id),
    entity_type TEXT NOT NULL,
    local_id UUID NOT NULL,
    remote_id TEXT NOT NULL,
    last_synced_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(erp_config_id, entity_type, local_id),
    UNIQUE(erp_config_id, entity_type, remote_id)
);

-- Index for fast lookups during sync
CREATE INDEX IF NOT EXISTS idx_id_mappings_local ON id_mappings(local_id);
CREATE INDEX IF NOT EXISTS idx_id_mappings_remote ON id_mappings(remote_id);
