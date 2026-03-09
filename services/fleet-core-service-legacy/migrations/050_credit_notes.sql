-- Migration: 050_credit_notes
-- Description: Add credit notes for refunds and adjustments

-- Create credit notes table
CREATE TABLE IF NOT EXISTS credit_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    credit_note_no VARCHAR(50) NOT NULL UNIQUE,
    org_id UUID NOT NULL,
    
    -- Reference to original invoice
    original_invoice_id UUID NOT NULL REFERENCES invoices(id),
    
    -- Customer details (denormalized for historical accuracy)
    customer_id UUID NOT NULL,
    customer_name VARCHAR(255) NOT NULL,
    customer_gstin VARCHAR(20),
    customer_pan VARCHAR(10),
    
    -- Amounts
    taxable_value DECIMAL(12, 2) NOT NULL DEFAULT 0,
    cgst DECIMAL(12, 2) NOT NULL DEFAULT 0,
    sgst DECIMAL(12, 2) NOT NULL DEFAULT 0,
    igst DECIMAL(12, 2) NOT NULL DEFAULT 0,
    total_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'issued', -- draft, issued, applied, refunded, cancelled
    reason TEXT NOT NULL,
    remarks TEXT,
    
    -- Application tracking
    applied_to_invoice_id UUID REFERENCES invoices(id),
    applied_amount DECIMAL(12, 2) DEFAULT 0,
    applied_at TIMESTAMPTZ,
    
    -- Dates
    issue_date DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    
    -- Constraints
    CONSTRAINT chk_credit_note_positive_amount CHECK (total_amount >= 0),
    CONSTRAINT chk_credit_note_status CHECK (status IN ('draft', 'issued', 'applied', 'refunded', 'cancelled')),
    CONSTRAINT chk_credit_note_issue_date CHECK (issue_date >= (SELECT invoice_date FROM invoices WHERE id = original_invoice_id))
);

-- Create credit note items table
CREATE TABLE IF NOT EXISTS credit_note_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    credit_note_id UUID NOT NULL REFERENCES credit_notes(id) ON DELETE CASCADE,
    original_invoice_item_id UUID, -- Reference to original invoice item if applicable
    
    description TEXT NOT NULL,
    hsn_code VARCHAR(20),
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price DECIMAL(12, 2) NOT NULL,
    tax_percent DECIMAL(5, 2) NOT NULL DEFAULT 0,
    tax_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    total_amount DECIMAL(12, 2) NOT NULL,
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_credit_note_item_positive CHECK (total_amount >= 0)
);

-- Indexes
CREATE INDEX idx_credit_notes_org_id ON credit_notes(org_id);
CREATE INDEX idx_credit_notes_original_invoice ON credit_notes(original_invoice_id);
CREATE INDEX idx_credit_notes_customer ON credit_notes(customer_id);
CREATE INDEX idx_credit_notes_status ON credit_notes(status);
CREATE INDEX idx_credit_notes_issue_date ON credit_notes(issue_date);
CREATE INDEX idx_credit_note_items_credit_note ON credit_note_items(credit_note_id);

-- Comments
COMMENT ON TABLE credit_notes IS 'Credit notes for refunds, adjustments, and returns';
COMMENT ON COLUMN credit_notes.original_invoice_id IS 'Original invoice being credited';
COMMENT ON COLUMN credit_notes.status IS 'draft=editable, issued=final, applied=used against invoice, refunded=cash refunded, cancelled=voided';
COMMENT ON COLUMN credit_notes.applied_to_invoice_id IS 'Invoice this credit was applied to (if different from original)';
COMMENT ON COLUMN credit_notes.reason IS 'Reason for credit note (required for compliance)';

-- Function to validate credit note amount doesn't exceed original invoice
CREATE OR REPLACE FUNCTION validate_credit_note_amount()
RETURNS TRIGGER AS $$
DECLARE
    original_outstanding DECIMAL(12, 2);
    existing_credits DECIMAL(12, 2);
BEGIN
    -- Get original invoice outstanding
    SELECT balance_due INTO original_outstanding
    FROM invoices
    WHERE id = NEW.original_invoice_id;
    
    -- Get sum of existing credit notes for this invoice
    SELECT COALESCE(SUM(total_amount), 0) INTO existing_credits
    FROM credit_notes
    WHERE original_invoice_id = NEW.original_invoice_id
      AND id != NEW.id
      AND status != 'cancelled';
    
    -- Validate
    IF NEW.total_amount > (original_outstanding - existing_credits) THEN
        RAISE EXCEPTION 'Credit note amount exceeds outstanding invoice balance';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add trigger for validation
DROP TRIGGER IF EXISTS trg_validate_credit_note_amount ON credit_notes;
CREATE TRIGGER trg_validate_credit_note_amount
    BEFORE INSERT OR UPDATE ON credit_notes
    FOR EACH ROW
    EXECUTE FUNCTION validate_credit_note_amount();

-- Function to auto-generate credit note number
CREATE OR REPLACE FUNCTION generate_credit_note_number()
RETURNS TRIGGER AS $$
DECLARE
    org_prefix VARCHAR(20);
    year_part VARCHAR(4);
    seq INTEGER;
BEGIN
    -- Get org prefix (or use default)
    SELECT COALESCE(invoice_prefix, 'CN') INTO org_prefix
    FROM billing_configurations
    WHERE org_id = NEW.org_id;
    
    year_part := TO_CHAR(NEW.issue_date, 'YYYY');
    
    -- Get next sequence number
    SELECT COALESCE(MAX(CAST(SUBSTRING(credit_note_no FROM LENGTH(org_prefix) + 6) AS INTEGER)), 0) + 1 INTO seq
    FROM credit_notes
    WHERE org_id = NEW.org_id
      AND credit_note_no LIKE org_prefix || '/' || year_part || '%';
    
    NEW.credit_note_no := org_prefix || '/' || year_part || '-' || LPAD(seq::TEXT, 6, '0');
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add trigger for auto-numbering
DROP TRIGGER IF EXISTS trg_generate_credit_note_number ON credit_notes;
CREATE TRIGGER trg_generate_credit_note_number
    BEFORE INSERT ON credit_notes
    FOR EACH ROW
    WHEN (NEW.credit_note_no IS NULL)
    EXECUTE FUNCTION generate_credit_note_number();
