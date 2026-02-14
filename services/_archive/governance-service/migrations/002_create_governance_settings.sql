-- Create governance settings table
CREATE TABLE IF NOT EXISTS governance_settings (
    organization_id UUID PRIMARY KEY,
    retention_policies JSONB NOT NULL DEFAULT '{}',
    masking_rules JSONB NOT NULL DEFAULT '{}',
    auto_expiry_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    download_ttl_hours INTEGER NOT NULL DEFAULT 48,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Audit logs for governance actions
CREATE TABLE IF NOT EXISTS governance_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    user_id UUID NOT NULL,
    action TEXT NOT NULL,
    module TEXT NOT NULL,
    change_detail TEXT,
    ip_address TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_gov_audit_org ON governance_audit_logs(organization_id);
