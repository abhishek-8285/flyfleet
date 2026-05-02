-- Migration: 067_governance_module
-- Description: Creates tables for governance requests with tenant isolation.

-- 1. Governance request type enum
DO $$ BEGIN
    CREATE TYPE governance_request_status AS ENUM ('PENDING', 'APPROVED', 'REJECTED', 'EXPIRED', 'COMPLETED');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE governance_request_type AS ENUM ('EXPORT', 'DELETION', 'ACCESS_REVIEW');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

-- 2. Governance Requests Table
CREATE TABLE IF NOT EXISTS governance_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE governance_requests ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE;
    ALTER TABLE governance_requests ADD COLUMN IF NOT EXISTS requester_id UUID NOT NULL;
    ALTER TABLE governance_requests ADD COLUMN IF NOT EXISTS request_type governance_request_type NOT NULL DEFAULT 'ACCESS_REVIEW';
    ALTER TABLE governance_requests ADD COLUMN IF NOT EXISTS data_scope TEXT NOT NULL DEFAULT '';
    ALTER TABLE governance_requests ADD COLUMN IF NOT EXISTS purpose TEXT NOT NULL DEFAULT '';
    ALTER TABLE governance_requests ADD COLUMN IF NOT EXISTS format TEXT NOT NULL DEFAULT 'CSV';
    ALTER TABLE governance_requests ADD COLUMN IF NOT EXISTS delivery_method TEXT NOT NULL DEFAULT 'DIRECT';
    ALTER TABLE governance_requests ADD COLUMN IF NOT EXISTS status governance_request_status NOT NULL DEFAULT 'PENDING';
    ALTER TABLE governance_requests ADD COLUMN IF NOT EXISTS expiry_date TIMESTAMPTZ;
    ALTER TABLE governance_requests ADD COLUMN IF NOT EXISTS comments TEXT;
    ALTER TABLE governance_requests ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
    ALTER TABLE governance_requests ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_governance_requests_tenant ON governance_requests(tenant_id);

-- 3. Governance Settings Table
CREATE TABLE IF NOT EXISTS governance_settings (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE
);

DO $$ 
BEGIN 
    ALTER TABLE governance_settings ADD COLUMN IF NOT EXISTS retention_policies JSONB NOT NULL DEFAULT '{}';
    ALTER TABLE governance_settings ADD COLUMN IF NOT EXISTS masking_rules JSONB NOT NULL DEFAULT '{}';
    ALTER TABLE governance_settings ADD COLUMN IF NOT EXISTS auto_expiry_enabled BOOLEAN NOT NULL DEFAULT true;
    ALTER TABLE governance_settings ADD COLUMN IF NOT EXISTS download_ttl_hours INTEGER NOT NULL DEFAULT 48;
    ALTER TABLE governance_settings ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
    ALTER TABLE governance_settings ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- 4. Governance Audit Logs Table
CREATE TABLE IF NOT EXISTS governance_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE governance_audit_logs ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE;
    ALTER TABLE governance_audit_logs ADD COLUMN IF NOT EXISTS user_id UUID NOT NULL;
    ALTER TABLE governance_audit_logs ADD COLUMN IF NOT EXISTS action TEXT NOT NULL;
    ALTER TABLE governance_audit_logs ADD COLUMN IF NOT EXISTS module TEXT NOT NULL;
    ALTER TABLE governance_audit_logs ADD COLUMN IF NOT EXISTS change_detail TEXT;
    ALTER TABLE governance_audit_logs ADD COLUMN IF NOT EXISTS ip_address INET;
    ALTER TABLE governance_audit_logs ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_governance_audit_tenant ON governance_audit_logs(tenant_id);
