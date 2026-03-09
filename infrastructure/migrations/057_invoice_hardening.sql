-- Update billing_invoice_documents to support reproducibility and integrity
ALTER TABLE billing_invoice_documents 
ADD COLUMN template_version INTEGER NOT NULL DEFAULT 1,
ADD COLUMN snapshot JSONB NOT NULL DEFAULT '{}',
ADD COLUMN checksum TEXT;

-- Index for snapshot querying (optional but good for auditing)
CREATE INDEX idx_billing_invoice_docs_snapshot ON billing_invoice_documents USING GIN (snapshot);

-- Add comments for documentation
COMMENT ON COLUMN billing_invoice_documents.snapshot IS 'JSON snapshot of all data used for rendering this specific PDF version';
COMMENT ON COLUMN billing_invoice_documents.checksum IS 'SHA256 checksum of the generated PDF file for integrity verification';
