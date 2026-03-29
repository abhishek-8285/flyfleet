-- Migration: 076_normalize_drivers_documents.sql
-- Description: Normalize drivers.documents JSONB array to driver_documents table
-- Created: 2026-03-29
-- Builds on: Migration 068 (normalized attendance_records and schedules)

SET search_path TO public;

-- =====================================================
-- PURPOSE: Better performance, queryability, and normalization
-- =====================================================

\echo 'Normalizing drivers.documents JSONB array...';

-- Enable extension if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ========================================
-- 1. Create driver_documents table
-- ========================================
CREATE TABLE IF NOT EXISTS driver_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Document metadata
    type VARCHAR(100) NOT NULL, -- e.g., 'license', 'pan_card', 'aadhaar', 'insurance', 'pollution_certificate'
    name VARCHAR(255) NOT NULL,
    document_number VARCHAR(100),
    issue_date DATE,
    expiry_date DATE,
    issuing_authority VARCHAR(255),

    -- File storage reference
    file_url TEXT,
    file_name VARCHAR(255),
    file_size_bytes BIGINT,
    mime_type VARCHAR(100),

    -- Verification status
    status VARCHAR(50) DEFAULT 'pending', -- pending, verified, rejected, expired
    verified_by UUID REFERENCES users(id) ON DELETE SET NULL,
    verified_at TIMESTAMPTZ,
    verification_notes TEXT,

    -- Compliance tracking
    is_mandatory BOOLEAN DEFAULT TRUE,
    reminder_days_before_expiry INTEGER DEFAULT 30,
    last_reminder_sent_at TIMESTAMPTZ,

    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Constraints
    UNIQUE(driver_id, type, document_number)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_driver_documents_driver ON driver_documents(driver_id);
CREATE INDEX IF NOT EXISTS idx_driver_documents_org ON driver_documents(org_id);
CREATE INDEX IF NOT EXISTS idx_driver_documents_type ON driver_documents(type);
CREATE INDEX IF NOT EXISTS idx_driver_documents_status ON driver_documents(status);
CREATE INDEX IF NOT EXISTS idx_driver_documents_expiry ON driver_documents(expiry_date) WHERE expiry_date IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_driver_documents_driver_type ON driver_documents(driver_id, type);
CREATE INDEX IF NOT EXISTS idx_driver_documents_org_status ON driver_documents(org_id, status);
CREATE INDEX IF NOT EXISTS idx_driver_documents_org_expiry ON driver_documents(org_id, expiry_date) WHERE expiry_date IS NOT NULL;

-- RLS for multi-tenancy
ALTER TABLE driver_documents ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS driver_documents_tenant_policy ON driver_documents;
CREATE POLICY driver_documents_tenant_policy ON driver_documents
    FOR ALL
    USING (org_id = get_current_org_id())
    WITH CHECK (org_id = get_current_org_id());

-- Trigger for updated_at
DROP TRIGGER IF EXISTS update_driver_documents_updated_at ON driver_documents;
CREATE TRIGGER update_driver_documents_updated_at
    BEFORE UPDATE ON driver_documents
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE driver_documents IS 'Normalized driver documents (license, PAN, Aadhaar, etc.). Replaces drivers.documents JSONB array.';
COMMENT ON COLUMN driver_documents.type IS 'Document type: license, pan_card, aadhaar, insurance, pollution_certificate, etc.';
COMMENT ON COLUMN driver_documents.status IS 'Verification status: pending, verified, rejected, expired';
COMMENT ON COLUMN driver_documents.is_mandatory IS 'Whether this document is required for the driver to be active';
COMMENT ON COLUMN driver_documents.reminder_days_before_expiry IS 'Days before expiry to send reminder notification';

-- ========================================
-- 2. Migrate data from drivers.documents JSONB
-- ========================================
-- Expected JSON structure per item:
-- {
--   "id": "uuid",
--   "type": "string",
--   "name": "string",
--   "document_number": "string",
--   "issue_date": "ISO8601",
--   "expiry_date": "ISO8601",
--   "issuing_authority": "string",
--   "file_url": "string",
--   "status": "pending|verified|rejected|expired",
--   "verified_by": "uuid",
--   "verified_at": "ISO8601",
--   "verification_notes": "string",
--   "is_mandatory": true,
--   "created_at": "ISO8601",
--   "updated_at": "ISO8601"
-- }

DO $$
DECLARE
    driver_record RECORD;
    doc_item JSONB;
    migrated_count INT := 0;
    skipped_null INT := 0;
    error_count INT := 0;
BEGIN
    FOR driver_record IN
        SELECT id, org_id, documents
        FROM drivers
        WHERE documents IS NOT NULL
          AND jsonb_array_length(documents) > 0
    LOOP
        FOR doc_item IN SELECT jsonb_array_elements(driver_record.documents)
        LOOP
            BEGIN
                INSERT INTO driver_documents (
                    id,
                    driver_id,
                    org_id,
                    type,
                    name,
                    document_number,
                    issue_date,
                    expiry_date,
                    issuing_authority,
                    file_url,
                    file_name,
                    file_size_bytes,
                    mime_type,
                    status,
                    verified_by,
                    verified_at,
                    verification_notes,
                    is_mandatory,
                    reminder_days_before_expiry,
                    created_at,
                    updated_at,
                    created_by,
                    updated_by
                ) VALUES (
                    COALESCE((doc_item->>'id')::UUID, gen_random_uuid()),
                    driver_record.id,
                    driver_record.org_id,
                    doc_item->>'type',
                    doc_item->>'name',
                    doc_item->>'document_number',
                    (doc_item->>'issue_date')::TIMESTAMPTZ,
                    (doc_item->>'expiry_date')::TIMESTAMPTZ,
                    doc_item->>'issuing_authority',
                    doc_item->>'file_url',
                    doc_item->>'file_name',
                    (doc_item->>'file_size_bytes')::BIGINT,
                    doc_item->>'mime_type',
                    COALESCE(doc_item->>'status', 'pending'),
                    (doc_item->>'verified_by')::UUID,
                    (doc_item->>'verified_at')::TIMESTAMPTZ,
                    doc_item->>'verification_notes',
                    COALESCE((doc_item->>'is_mandatory')::BOOLEAN, TRUE),
                    COALESCE((doc_item->>'reminder_days_before_expiry')::INTEGER, 30),
                    COALESCE((doc_item->>'created_at')::TIMESTAMPTZ, NOW()),
                    COALESCE((doc_item->>'updated_at')::TIMESTAMPTZ, NOW()),
                    (doc_item->>'created_by')::UUID,
                    (doc_item->>'updated_by')::UUID
                )
                ON CONFLICT (driver_id, type, document_number) DO NOTHING;

                migrated_count := migrated_count + 1;
            EXCEPTION
                WHEN others THEN
                    error_count := error_count + 1;
                    RAISE WARNING 'Failed to migrate document for driver %: %', driver_record.id, SQLERRM;
            END;
        END LOOP;
    END LOOP;

    -- Count skipped drivers
    SELECT COUNT(*) INTO skipped_null
    FROM drivers
    WHERE documents IS NULL OR jsonb_array_length(documents) = 0;

    RAISE NOTICE 'Data migration complete:';
    RAISE NOTICE '  Migrated document records: %', migrated_count;
    RAISE NOTICE '  Skipped drivers (no documents): %', skipped_null;
    RAISE NOTICE '  Errors encountered: %', error_count;

    IF error_count > 0 THEN
        RAISE WARNING 'Some documents failed to migrate. Review warnings above.';
    END IF;
END $$;

-- ========================================
-- 3. Add deprecation comment to drivers.documents
-- ========================================
COMMENT ON COLUMN drivers.documents IS
    'DEPRECATED: Use driver_documents table instead. Kept for backward compatibility during migration.';

-- ========================================
-- 4. Verification
-- ========================================
DO $$
DECLARE
    new_table_count INT;
    total_docs_new INT;
    total_docs_old INT;
BEGIN
    SELECT COUNT(*) INTO new_table_count
    FROM information_schema.tables
    WHERE table_schema = 'public'
      AND table_name = 'driver_documents';

    IF new_table_count = 1 THEN
        RAISE NOTICE 'driver_documents table exists';

        SELECT COUNT(*) INTO total_docs_new
        FROM driver_documents;

        SELECT COUNT(*) INTO total_docs_old
        FROM drivers,
             jsonb_array_elements(documents)
        WHERE documents IS NOT NULL;

        RAISE NOTICE 'Documents in new table: %', total_docs_new;
        RAISE NOTICE 'Documents in old JSONB (estimate): %', COALESCE(total_docs_old, 0);

        IF total_docs_new >= COALESCE(total_docs_old, 0) * 0.95 THEN
            RAISE NOTICE 'Data migration appears successful (>95%% coverage)';
        END IF;
    ELSE
        RAISE WARNING 'driver_documents table not found';
    END IF;
END $$;

\echo '====================================================='
\echo 'drivers.documents normalization complete!'
\echo '====================================================='
\echo 'Summary:'
\echo '  - Created driver_documents table with proper FKs and indexes'
\echo '  - Migrated data from drivers.documents JSONB'
\echo '  - RLS enabled for multi-tenancy'
\echo '  - Deprecation comment added to drivers.documents'
\echo ''
\echo 'Next steps:'
\echo '  1. Update application code to use driver_documents table'
\echo '  2. After confirming migration success, consider:'
\echo '     - Removing drivers.documents column in future migration (after deprecation period)'
\echo '     - Adding triggers to keep backward compatibility if needed'
\echo '====================================================='
