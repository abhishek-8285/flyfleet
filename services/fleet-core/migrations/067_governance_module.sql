-- Migration: 067_governance_module.sql
-- Description: Creates tables for governance requests, settings, and audit logging
-- Created: 2026-03-03

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create governance request status enum
DO $$ BEGIN
    CREATE TYPE governance_request_status AS ENUM (
        'PENDING',
        'APPROVED',
        'REJECTED',
        'EXPIRED',
        'COMPLETED'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create governance request type enum
DO $$ BEGIN
    CREATE TYPE governance_request_type AS ENUM (
        'EXPORT',
        'DELETION',
        'ACCESS_REVIEW'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Governance Requests Table
-- Stores data governance requests (export, deletion, access review)
CREATE TABLE IF NOT EXISTS governance_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    requester_id    UUID NOT NULL,
    request_type    governance_request_type NOT NULL,
    data_scope      TEXT NOT NULL,
    purpose         TEXT NOT NULL,
    format          TEXT NOT NULL DEFAULT 'CSV',
    delivery_method TEXT NOT NULL DEFAULT 'DIRECT',
    status          governance_request_status NOT NULL DEFAULT 'PENDING',
    expiry_date     TIMESTAMPTZ,
    comments        TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Indexes for common queries
    CONSTRAINT fk_governance_requests_org FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- Indexes for governance_requests
CREATE INDEX IF NOT EXISTS idx_governance_requests_org_id ON governance_requests(organization_id);
CREATE INDEX IF NOT EXISTS idx_governance_requests_status ON governance_requests(status);
CREATE INDEX IF NOT EXISTS idx_governance_requests_type ON governance_requests(request_type);
CREATE INDEX IF NOT EXISTS idx_governance_requests_requester ON governance_requests(requester_id);
CREATE INDEX IF NOT EXISTS idx_governance_requests_created_at ON governance_requests(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_governance_requests_expiry ON governance_requests(expiry_date) WHERE status = 'APPROVED';

-- Governance Settings Table
-- Stores organization-level governance configuration
CREATE TABLE IF NOT EXISTS governance_settings (
    organization_id     UUID PRIMARY KEY,
    retention_policies  JSONB NOT NULL DEFAULT '{}',
    masking_rules       JSONB NOT NULL DEFAULT '{}',
    auto_expiry_enabled BOOLEAN NOT NULL DEFAULT true,
    download_ttl_hours  INTEGER NOT NULL DEFAULT 48,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by          UUID,
    updated_by          UUID,
    
    -- Indexes for common queries
    CONSTRAINT fk_governance_settings_org FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- Governance Audit Logs Table
-- Tracks all governance-related actions for compliance
CREATE TABLE IF NOT EXISTS governance_audit_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    user_id         UUID NOT NULL,
    action          TEXT NOT NULL,
    module          TEXT NOT NULL,
    change_detail   TEXT,
    ip_address      INET,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Indexes for common queries
    CONSTRAINT fk_governance_audit_org FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- Indexes for governance_audit_logs
CREATE INDEX IF NOT EXISTS idx_governance_audit_org_id ON governance_audit_logs(organization_id);
CREATE INDEX IF NOT EXISTS idx_governance_audit_user_id ON governance_audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_governance_audit_action ON governance_audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_governance_audit_created_at ON governance_audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_governance_audit_module ON governance_audit_logs(module);

-- Insert default settings for existing organizations (if any)
-- This will be handled by the application on first access

-- Add trigger to auto-update updated_at timestamp for governance_requests
CREATE OR REPLACE FUNCTION update_governance_requests_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_governance_requests ON governance_requests;
CREATE TRIGGER trigger_update_governance_requests
    BEFORE UPDATE ON governance_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_governance_requests_updated_at();

-- Add trigger to auto-update updated_at timestamp for governance_settings
CREATE OR REPLACE FUNCTION update_governance_settings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_governance_settings ON governance_settings;
CREATE TRIGGER trigger_update_governance_settings
    BEFORE UPDATE ON governance_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_governance_settings_updated_at();

-- Comments for documentation
COMMENT ON TABLE governance_requests IS 'Stores data governance requests for export, deletion, and access review workflows';
COMMENT ON TABLE governance_settings IS 'Organization-level governance configuration including retention policies and masking rules';
COMMENT ON TABLE governance_audit_logs IS 'Audit trail for all governance-related actions for compliance reporting';

COMMENT ON COLUMN governance_requests.request_type IS 'Type of governance request: EXPORT, DELETION, or ACCESS_REVIEW';
COMMENT ON COLUMN governance_requests.status IS 'Current status: PENDING, APPROVED, REJECTED, EXPIRED, or COMPLETED';
COMMENT ON COLUMN governance_requests.expiry_date IS 'Date when approved request expires and data should be deleted';
COMMENT ON COLUMN governance_requests.data_scope IS 'Description of data covered by the request';
COMMENT ON COLUMN governance_requests.purpose IS 'Business justification for the request';

COMMENT ON COLUMN governance_settings.retention_policies IS 'JSON configuration for data retention rules';
COMMENT ON COLUMN governance_settings.masking_rules IS 'JSON configuration for data masking requirements';
COMMENT ON COLUMN governance_settings.download_ttl_hours IS 'Hours before approved export links expire (default: 48)';

COMMENT ON COLUMN governance_audit_logs.action IS 'Action performed: CREATE_REQUEST, APPROVE_REQUEST, REJECT_REQUEST, UPDATE_SETTINGS, etc.';
COMMENT ON COLUMN governance_audit_logs.module IS 'Module affected: DATA_REQUESTS, GOVERNANCE_SETTINGS, etc.';
COMMENT ON COLUMN governance_audit_logs.change_detail IS 'Details of what changed (e.g., approval notes)';
