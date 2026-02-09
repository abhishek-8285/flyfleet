-- Integrations Management Module Schema
-- Integrations, API Clients, Webhooks, and Logs

-- 1. Integrations Table
CREATE TABLE IF NOT EXISTS integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- e.g., 'SMS', 'CLOUD', 'ERP'
    provider VARCHAR(100) NOT NULL, -- e.g., 'Twilio', 'AWS', 'SAP'
    status VARCHAR(20) DEFAULT 'active', -- active, inactive, error
    config JSONB DEFAULT '{}', -- Encrypted credentials and settings
    last_sync TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_integrations_org ON integrations(org_id);
CREATE INDEX idx_integrations_status ON integrations(status);

-- 2. API Clients Table
CREATE TABLE IF NOT EXISTS api_clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    client_id VARCHAR(100) UNIQUE NOT NULL,
    client_secret_hash TEXT NOT NULL,
    name VARCHAR(255) NOT NULL,
    scopes TEXT[] DEFAULT '{}',
    expiry TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) DEFAULT 'active', -- active, revoked, expired
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_api_clients_org ON api_clients(org_id);
CREATE INDEX idx_api_clients_client_id ON api_clients(client_id);

-- 3. Webhooks Table
CREATE TABLE IF NOT EXISTS webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    url TEXT NOT NULL,
    events TEXT[] DEFAULT '{}', -- e.g., ['TRIP_START', 'ALERT_TRIGGERED']
    secret VARCHAR(255), -- Signing secret
    status VARCHAR(20) DEFAULT 'active', -- active, inactive, failing
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_webhooks_org ON webhooks(org_id);

-- 4. Integration Logs Table
CREATE TABLE IF NOT EXISTS integration_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    integration_id UUID REFERENCES integrations(id) ON DELETE SET NULL,
    method VARCHAR(10) NOT NULL, -- GET, POST, etc.
    endpoint TEXT NOT NULL,
    request_body JSONB,
    response_body JSONB,
    status_code INTEGER,
    status VARCHAR(20), -- success, failure
    duration_ms INTEGER,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_integration_logs_org ON integration_logs(org_id);
CREATE INDEX idx_integration_logs_integration ON integration_logs(integration_id);
CREATE INDEX idx_integration_logs_timestamp ON integration_logs(timestamp DESC);
