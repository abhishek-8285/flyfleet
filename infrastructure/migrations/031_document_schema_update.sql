-- Migration: Add missing columns to documents table for full document management
-- This adds org_id, title, file_name, file_size_bytes, mime_type, storage_path,
-- reminder_days, uploaded_by, and soft delete support.

-- Add org_id column (nullable initially, then backfill from vehicle's org)
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'documents' AND column_name = 'org_id') THEN
        ALTER TABLE documents ADD COLUMN org_id UUID REFERENCES organizations(id);
    END IF;
END $$;

-- Add title column
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'documents' AND column_name = 'title') THEN
        ALTER TABLE documents ADD COLUMN title VARCHAR(255);
    END IF;
END $$;

-- Add file_name column
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'documents' AND column_name = 'file_name') THEN
        ALTER TABLE documents ADD COLUMN file_name VARCHAR(255) DEFAULT '';
    END IF;
END $$;

-- Add file_size_bytes column
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'documents' AND column_name = 'file_size_bytes') THEN
        ALTER TABLE documents ADD COLUMN file_size_bytes INTEGER DEFAULT 0;
    END IF;
END $$;

-- Add mime_type column
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'documents' AND column_name = 'mime_type') THEN
        ALTER TABLE documents ADD COLUMN mime_type VARCHAR(100) DEFAULT 'application/octet-stream';
    END IF;
END $$;

-- Add storage_path column (internal only, not exposed via API)
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'documents' AND column_name = 'storage_path') THEN
        ALTER TABLE documents ADD COLUMN storage_path VARCHAR(500) DEFAULT '';
    END IF;
END $$;

-- Add reminder_days column
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'documents' AND column_name = 'reminder_days') THEN
        ALTER TABLE documents ADD COLUMN reminder_days INTEGER DEFAULT 30;
    END IF;
END $$;

-- Add uploaded_by column (FK to users)
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'documents' AND column_name = 'uploaded_by') THEN
        ALTER TABLE documents ADD COLUMN uploaded_by UUID REFERENCES users(id);
    END IF;
END $$;

-- Add soft delete column
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'documents' AND column_name = 'deleted_at') THEN
        ALTER TABLE documents ADD COLUMN deleted_at TIMESTAMPTZ;
    END IF;
END $$;

-- Backfill org_id from vehicle's organization
UPDATE documents d
SET org_id = v.org_id
FROM vehicles v
WHERE d.vehicle_id = v.id AND d.org_id IS NULL;

-- Backfill title from doc_number where title is null
UPDATE documents SET title = doc_number WHERE title IS NULL;

-- Add indexes for new query patterns
CREATE INDEX IF NOT EXISTS idx_documents_org_id ON documents(org_id);
CREATE INDEX IF NOT EXISTS idx_documents_org_expiry ON documents(org_id, expiry_date) WHERE deleted_at IS NULL AND status != 'expired';
CREATE INDEX IF NOT EXISTS idx_documents_vehicle_status ON documents(vehicle_id, status) WHERE deleted_at IS NULL;
