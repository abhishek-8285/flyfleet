-- Migration: 249_governance_hardening.sql
-- Description: Hardening Governance module with legal holds, verification, and lineage tracking
-- Created: 2026-03-08

-- Phase 2: Legal Hold System
CREATE TABLE IF NOT EXISTS legal_holds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    entity_type TEXT NOT NULL,
    entity_id UUID NOT NULL,
    reason TEXT,
    created_by UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_legal_holds_entity
ON legal_holds(entity_type, entity_id);

-- Phase 3: Identity Verification
ALTER TABLE governance_requests
ADD COLUMN IF NOT EXISTS verification_status TEXT DEFAULT 'PENDING',
ADD COLUMN IF NOT EXISTS verified_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS verified_by UUID;

-- Phase 5: Export Data Lineage
CREATE TABLE IF NOT EXISTS governance_export_lineage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID REFERENCES governance_requests(id) ON DELETE CASCADE,
    table_name TEXT NOT NULL,
    record_count INTEGER,
    export_size_bytes BIGINT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Phase 6: Compliance Report Storage
CREATE TABLE IF NOT EXISTS governance_compliance_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID REFERENCES governance_requests(id) ON DELETE CASCADE,
    export_checksum TEXT,
    export_size_bytes BIGINT,
    processing_time_ms INTEGER,
    tables_accessed JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Phase 8: Extended Status Workflow
-- Note: ALTER TYPE ... ADD VALUE cannot be executed inside a transaction block in some Postgres versions.
-- However, we'll try it here.
ALTER TYPE governance_request_status ADD VALUE IF NOT EXISTS 'PROCESSING';
ALTER TYPE governance_request_status ADD VALUE IF NOT EXISTS 'FAILED';
