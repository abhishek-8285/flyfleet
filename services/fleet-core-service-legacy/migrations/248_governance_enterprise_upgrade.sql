-- Migration: 248_governance_enterprise_upgrade.sql
-- Description: Upgrade Governance module with export storage and download security
-- Created: 2026-03-08

-- Governance Exports Table
CREATE TABLE IF NOT EXISTS governance_exports (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id  UUID NOT NULL REFERENCES governance_requests(id) ON DELETE CASCADE,
    file_path   TEXT NOT NULL,
    file_size   BIGINT,
    checksum    TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_governance_exports_request
ON governance_exports(request_id);

-- Governance Download Tokens Table
CREATE TABLE IF NOT EXISTS governance_download_tokens (
    token       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    export_id   UUID NOT NULL REFERENCES governance_exports(id) ON DELETE CASCADE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_governance_download_tokens_export
ON governance_download_tokens(export_id);

-- Add comments
COMMENT ON TABLE governance_exports IS 'Stores metadata for generated data exports';
COMMENT ON TABLE governance_download_tokens IS 'Secure short-lived tokens for downloading exports';
