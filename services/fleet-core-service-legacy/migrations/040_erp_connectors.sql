CREATE TABLE IF NOT EXISTS erp_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL, -- Multi-tenant support
    name TEXT NOT NULL, -- e.g., "SAP Production", "SFDC Sandbox"
    type TEXT NOT NULL, -- 'SAP', 'ORACLE', 'SALESFORCE', 'DYNAMICS'
    endpoint_url TEXT NOT NULL,
    auth_config JSONB NOT NULL, -- Encrypted credentials (API Key, OAuth tokens)
    settings JSONB DEFAULT '{}'::jsonb, -- Connector specific settings (e.g., SOAP version)
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sync_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL,
    erp_config_id UUID REFERENCES erp_configs(id),
    entity_type TEXT NOT NULL, -- 'CUSTOMER', 'VENDOR', 'ORDER', 'INVOICE'
    direction TEXT NOT NULL, -- 'INBOUND' (ERP -> Fleet), 'OUTBOUND' (Fleet -> ERP)
    status TEXT DEFAULT 'PENDING', -- 'PENDING', 'IN_PROGRESS', 'COMPLETED', 'FAILED'
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    records_processed INT DEFAULT 0,
    error_log TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS id_mappings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL,
    erp_config_id UUID REFERENCES erp_configs(id),
    entity_type TEXT NOT NULL,
    local_id UUID NOT NULL, -- FlyFleet ID
    remote_id TEXT NOT NULL, -- ERP ID
    last_synced_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(erp_config_id, entity_type, local_id),
    UNIQUE(erp_config_id, entity_type, remote_id)
);

-- Index for fast lookups during sync
CREATE INDEX idx_id_mappings_local ON id_mappings(local_id);
CREATE INDEX idx_id_mappings_remote ON id_mappings(remote_id);
