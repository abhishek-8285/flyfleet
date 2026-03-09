-- migration: 056_invoice_rendering_infrastructure.sql
SET search_path TO public;

-- 1. Template Storage (Versioned)
CREATE TABLE IF NOT EXISTS billing_invoice_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    version INT NOT NULL,
    html_content TEXT NOT NULL,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, version)
);

-- 2. Document Registry (Tracks generated PDFs)
CREATE TABLE IF NOT EXISTS billing_invoice_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    template_id UUID NOT NULL REFERENCES billing_invoice_templates(id),
    file_url TEXT NOT NULL,
    storage_key TEXT NOT NULL,
    generated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Add Custom Fields/Metadata to Invoices if missing
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name='invoices' AND column_name='custom_fields') THEN
        ALTER TABLE invoices ADD COLUMN custom_fields JSONB DEFAULT '{}';
    END IF;
END $$;

-- 4. RLS Policies
ALTER TABLE billing_invoice_templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE billing_invoice_documents ENABLE ROW LEVEL SECURITY;

-- Templates are global/system-level usually, but can be org-scoped if needed. 
-- For now, let's keep them readable by all authenticated apps.
CREATE POLICY template_read_policy ON billing_invoice_templates
    FOR SELECT TO flyfleet_app
    USING (true);

CREATE POLICY document_tenant_policy ON billing_invoice_documents
    FOR ALL TO flyfleet_app
    USING (EXISTS (SELECT 1 FROM invoices i WHERE i.id = invoice_id AND i.org_id = get_current_org_id()));

-- 5. Indexes
CREATE INDEX IF NOT EXISTS idx_template_name ON billing_invoice_templates(name);
CREATE INDEX IF NOT EXISTS idx_doc_invoice_id ON billing_invoice_documents(invoice_id);

-- 6. Permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON billing_invoice_templates TO flyfleet_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON billing_invoice_documents TO flyfleet_app;
