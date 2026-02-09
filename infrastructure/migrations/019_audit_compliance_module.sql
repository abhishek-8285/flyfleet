-- Audit & Compliance Module Expansion
-- Created: 2026-02-09
-- Description: Detailed entity history, evidence job queue, and policy management

-- 1. Align audit_logs
ALTER TABLE audit_logs 
ADD COLUMN IF NOT EXISTS module VARCHAR(100),
ADD COLUMN IF NOT EXISTS entity_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS before_state JSONB,
ADD COLUMN IF NOT EXISTS after_state JSONB,
ADD COLUMN IF NOT EXISTS metadata JSONB;

-- 2. Entity History Timeline
CREATE TABLE IF NOT EXISTS entity_audit_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(100) NOT NULL,
    change_data JSONB NOT NULL,
    actor UUID REFERENCES users(id) ON DELETE SET NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_entity_history_id ON entity_audit_history(entity_id);
CREATE INDEX IF NOT EXISTS idx_entity_history_type ON entity_audit_history(entity_type);

-- 3. Evidence Exports job queue
CREATE TABLE IF NOT EXISTS evidence_exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    module VARCHAR(50) NOT NULL,
    date_range_start TIMESTAMP WITH TIME ZONE,
    date_range_end TIMESTAMP WITH TIME ZONE,
    format VARCHAR(20) DEFAULT 'pdf', -- pdf, csv
    status VARCHAR(20) DEFAULT 'processing', -- processing, ready, failed
    requested_by UUID REFERENCES users(id) ON DELETE SET NULL,
    file_url TEXT,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 4. Audit Policies (Retention & Masking)
-- Extends/Replaces basic audit_settings
CREATE TABLE IF NOT EXISTS audit_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
    retention_years INTEGER DEFAULT 7,
    archive_policy VARCHAR(100) DEFAULT 'MOVE_TO_COLD_STORAGE_1Y',
    masking_rules JSONB DEFAULT '["credit_card", "ssn", "phone", "email"]',
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
