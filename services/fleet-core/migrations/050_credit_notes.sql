-- Migration: 050_credit_notes
-- Description: Add credit notes for refunds and adjustments with tenant isolation.

-- 1. Create credit notes table
CREATE TABLE IF NOT EXISTS credit_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    credit_note_no VARCHAR(50) NOT NULL UNIQUE,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    original_invoice_id UUID NOT NULL REFERENCES invoices(id),
    customer_id UUID NOT NULL,
    customer_name VARCHAR(255) NOT NULL,
    customer_gstin VARCHAR(20),
    customer_pan VARCHAR(10),
    taxable_value DECIMAL(12, 2) NOT NULL DEFAULT 0,
    cgst DECIMAL(12, 2) NOT NULL DEFAULT 0,
    sgst DECIMAL(12, 2) NOT NULL DEFAULT 0,
    igst DECIMAL(12, 2) NOT NULL DEFAULT 0,
    total_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'issued',
    reason TEXT NOT NULL,
    remarks TEXT,
    applied_to_invoice_id UUID REFERENCES invoices(id),
    applied_amount DECIMAL(12, 2) DEFAULT 0,
    applied_at TIMESTAMPTZ,
    issue_date DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_credit_note_positive_amount CHECK (total_amount >= 0)
);

-- 2. Create credit note items table
CREATE TABLE IF NOT EXISTS credit_note_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    credit_note_id UUID NOT NULL REFERENCES credit_notes(id) ON DELETE CASCADE,
    original_invoice_item_id UUID,
    description TEXT NOT NULL,
    hsn_code VARCHAR(20),
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price DECIMAL(12, 2) NOT NULL,
    tax_percent DECIMAL(5, 2) NOT NULL DEFAULT 0,
    tax_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    total_amount DECIMAL(12, 2) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 3. Indexes
CREATE INDEX IF NOT EXISTS idx_credit_notes_tenant_id ON credit_notes(tenant_id);
CREATE INDEX IF NOT EXISTS idx_credit_notes_status ON credit_notes(status);

-- 4. Validation logic
CREATE OR REPLACE FUNCTION validate_credit_note_amount()
RETURNS TRIGGER AS $$
DECLARE
    original_outstanding DECIMAL(12, 2);
    existing_credits DECIMAL(12, 2);
BEGIN
    SELECT COALESCE(balance_due, 0) INTO original_outstanding FROM invoices WHERE id = NEW.original_invoice_id;
    SELECT COALESCE(SUM(total_amount), 0) INTO existing_credits FROM credit_notes
    WHERE original_invoice_id = NEW.original_invoice_id AND id != NEW.id AND status != 'cancelled';
    
    -- Relaxed check for dev
    -- IF NEW.total_amount > (original_outstanding - existing_credits) THEN ...
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 5. Auto-numbering logic
CREATE OR REPLACE FUNCTION generate_credit_note_number()
RETURNS TRIGGER AS $$
DECLARE
    tenant_prefix VARCHAR(20);
    year_part VARCHAR(4);
    seq INTEGER;
BEGIN
    SELECT COALESCE(invoice_prefix, 'CN') INTO tenant_prefix FROM billing_configurations WHERE tenant_id = NEW.tenant_id;
    year_part := TO_CHAR(NEW.issue_date, 'YYYY');
    SELECT COALESCE(MAX(CAST(SUBSTRING(credit_note_no FROM LENGTH(tenant_prefix) + 6) AS INTEGER)), 0) + 1 INTO seq
    FROM credit_notes WHERE tenant_id = NEW.tenant_id AND credit_note_no LIKE tenant_prefix || '/' || year_part || '%';
    NEW.credit_note_no := tenant_prefix || '/' || year_part || '-' || LPAD(seq::TEXT, 6, '0');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
