-- Migration: 076_india_integration_audit.sql
-- Description: Creates table for India integration audit logs
-- Created: 2026-03-09

CREATE TABLE IF NOT EXISTS india_integration_logs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_type        VARCHAR(50) NOT NULL, -- fastag, ewaybill, vahan, sarathi, gst
    request_payload     JSONB,
    response_payload    JSONB,
    status              INTEGER NOT NULL, -- HTTP status or internal status
    error_message       TEXT,
    tenant_id           UUID,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_india_integration_service_type ON india_integration_logs(service_type);
CREATE INDEX IF NOT EXISTS idx_india_integration_tenant_id ON india_integration_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_india_integration_created_at ON india_integration_logs(created_at DESC);

COMMENT ON TABLE india_integration_logs IS 'Audit logs for India government system integrations';
