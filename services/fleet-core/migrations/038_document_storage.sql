-- 038_document_storage.sql
-- Centralized document management system for all platform entities

-- Idempotent table creation
CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

-- Idempotent column additions for existing or new table
DO $$ 
BEGIN 
    -- Core identity
    ALTER TABLE documents ADD COLUMN IF NOT EXISTS tenant_id UUID;
    ALTER TABLE documents ADD COLUMN IF NOT EXISTS entity_type VARCHAR(50);
    ALTER TABLE documents ADD COLUMN IF NOT EXISTS entity_id UUID;
    
    -- File details
    ALTER TABLE documents ADD COLUMN IF NOT EXISTS document_type VARCHAR(50);
    ALTER TABLE documents ADD COLUMN IF NOT EXISTS file_name VARCHAR(255);
    ALTER TABLE documents ADD COLUMN IF NOT EXISTS file_path TEXT;
    ALTER TABLE documents ADD COLUMN IF NOT EXISTS file_size_bytes BIGINT DEFAULT 0;
    ALTER TABLE documents ADD COLUMN IF NOT EXISTS mime_type VARCHAR(100) DEFAULT 'application/octet-stream';
    
    -- Metadata
    ALTER TABLE documents ADD COLUMN IF NOT EXISTS expiry_date TIMESTAMP WITH TIME ZONE;
    ALTER TABLE documents ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'PENDING';
    ALTER TABLE documents ADD COLUMN IF NOT EXISTS remarks TEXT;
    ALTER TABLE documents ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}';
    ALTER TABLE documents ADD COLUMN IF NOT EXISTS uploaded_by UUID;
    
    -- Timestamps
    ALTER TABLE documents ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
    ALTER TABLE documents ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
    ALTER TABLE documents ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;
END $$;

CREATE INDEX IF NOT EXISTS idx_documents_entity ON documents(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_documents_tenant ON documents(tenant_id);
CREATE INDEX IF NOT EXISTS idx_documents_expiry ON documents(expiry_date) WHERE expiry_date IS NOT NULL;

-- Add triggers for updated_at
DROP TRIGGER IF EXISTS set_updated_at_documents ON documents;
CREATE TRIGGER set_updated_at_documents
BEFORE UPDATE ON documents
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
