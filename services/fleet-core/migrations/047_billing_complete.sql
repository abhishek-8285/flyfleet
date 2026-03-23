-- Migration: 047_billing_complete
-- Description: Complete billing module schema with invoices, items, settings, and payments

-- Create invoices table (aligned with backend models)
CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    invoice_number VARCHAR(100) NOT NULL UNIQUE,
    customer_name VARCHAR(255) NOT NULL,
    customer_id UUID,
    shipment_id UUID,
    invoice_date DATE NOT NULL,
    due_date DATE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING', -- PENDING, PAID, OVERDUE, PARTIAL
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    total_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    amount_paid DECIMAL(12, 2) NOT NULL DEFAULT 0,
    balance_due DECIMAL(12, 2) NOT NULL DEFAULT 0,
    notes TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID,
    deleted_at TIMESTAMPTZ
);

-- Create invoice_items table
CREATE TABLE IF NOT EXISTS invoice_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price DECIMAL(12, 2) NOT NULL,
    tax_percent DECIMAL(5, 2) NOT NULL DEFAULT 0,
    total DECIMAL(12, 2) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Create billing_settings table
CREATE TABLE IF NOT EXISTS billing_settings (
    org_id UUID PRIMARY KEY,
    default_tax_rate DECIMAL(5, 2) NOT NULL DEFAULT 18.0,
    base_currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    grace_period_days INTEGER NOT NULL DEFAULT 14,
    auto_invoicing_enabled BOOLEAN NOT NULL DEFAULT false,
    invoice_prefix VARCHAR(20) NOT NULL DEFAULT 'INV',
    next_invoice_number INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID
);

-- Create payment_records table for audit trail
CREATE TABLE IF NOT EXISTS payment_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    amount DECIMAL(12, 2) NOT NULL,
    payment_method VARCHAR(50), -- CASH, CARD, TRANSFER, CHECK
    payment_reference VARCHAR(255),
    payment_date TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID
);

-- Create invoice_sequence table for auto-numbering
CREATE TABLE IF NOT EXISTS invoice_sequence (
    id INTEGER PRIMARY KEY DEFAULT 1,
    current_value INTEGER NOT NULL DEFAULT 1,
    prefix VARCHAR(20) NOT NULL DEFAULT 'INV',
    year INTEGER NOT NULL DEFAULT EXTRACT(YEAR FROM CURRENT_DATE),
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT single_row CHECK (id = 1)
);

-- Initialize sequence
INSERT INTO invoice_sequence (id, current_value, prefix, year) 
VALUES (1, 1, 'INV', EXTRACT(YEAR FROM CURRENT_DATE))
ON CONFLICT (id) DO NOTHING;

-- Indexes for invoices
CREATE INDEX IF NOT EXISTS idx_invoices_org_id ON invoices(org_id);
CREATE INDEX IF NOT EXISTS idx_invoices_customer_id ON invoices(customer_id);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status);
CREATE INDEX IF NOT EXISTS idx_invoices_due_date ON invoices(due_date);
CREATE INDEX IF NOT EXISTS idx_invoices_invoice_date ON invoices(invoice_date);
CREATE INDEX IF NOT EXISTS idx_invoices_deleted_at ON invoices(deleted_at);
CREATE INDEX IF NOT EXISTS idx_invoices_invoice_number ON invoices(invoice_number);

-- Indexes for invoice_items
CREATE INDEX IF NOT EXISTS idx_invoice_items_invoice_id ON invoice_items(invoice_id);

-- Indexes for payment_records
CREATE INDEX IF NOT EXISTS idx_payment_records_invoice_id ON payment_records(invoice_id);
CREATE INDEX IF NOT EXISTS idx_payment_records_payment_date ON payment_records(payment_date);

-- Function to generate invoice number
CREATE OR REPLACE FUNCTION generate_invoice_number()
RETURNS TRIGGER AS $$
DECLARE
    seq_record RECORD;
    new_number INTEGER;
BEGIN
    -- Get and increment sequence
    SELECT * INTO seq_record FROM invoice_sequence WHERE id = 1 FOR UPDATE;
    
    -- Reset sequence if year changed
    IF seq_record.year != EXTRACT(YEAR FROM CURRENT_DATE) THEN
        new_number := 1;
        UPDATE invoice_sequence 
        SET current_value = 2, 
            year = EXTRACT(YEAR FROM CURRENT_DATE),
            updated_at = CURRENT_TIMESTAMP
        WHERE id = 1;
    ELSE
        new_number := seq_record.current_value;
        UPDATE invoice_sequence 
        SET current_value = current_value + 1,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = 1;
    END IF;
    
    -- Set invoice number
    NEW.invoice_number := format('%s-%s-%06d', seq_record.prefix, EXTRACT(YEAR FROM CURRENT_DATE), new_number);
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-generate invoice number
CREATE TRIGGER trg_generate_invoice_number
    BEFORE INSERT ON invoices
    FOR EACH ROW
    WHEN (NEW.invoice_number IS NULL)
    EXECUTE FUNCTION generate_invoice_number();

-- Function to update overdue invoices
CREATE OR REPLACE FUNCTION update_overdue_invoices()
RETURNS void AS $$
BEGIN
    UPDATE invoices 
    SET status = 'OVERDUE', 
        updated_at = CURRENT_TIMESTAMP,
        metadata = jsonb_set(metadata, '{overdue_updated_at}', to_jsonb(CURRENT_TIMESTAMP))
    WHERE status IN ('PENDING', 'PARTIAL')
      AND due_date < CURRENT_DATE
      AND id NOT IN (
          SELECT id FROM invoices 
          WHERE status = 'OVERDUE' 
            AND metadata->>'overdue_updated_at' IS NOT NULL
            AND (metadata->>'overdue_updated_at')::TIMESTAMPTZ > CURRENT_DATE - INTERVAL '1 day'
      );
END;
$$ LANGUAGE plpgsql;

-- Function to calculate invoice totals
CREATE OR REPLACE FUNCTION calculate_invoice_total(p_invoice_id UUID)
RETURNS TABLE(total_amount DECIMAL, balance_due DECIMAL) AS $$
DECLARE
    v_total DECIMAL := 0;
    v_paid DECIMAL;
BEGIN
    -- Sum all line items
    SELECT COALESCE(SUM(total), 0) INTO v_total
    FROM invoice_items
    WHERE invoice_id = p_invoice_id;
    
    -- Get amount paid
    SELECT COALESCE(amount_paid, 0) INTO v_paid
    FROM invoices
    WHERE id = p_invoice_id;
    
    RETURN QUERY SELECT v_total, (v_total - v_paid);
END;
$$ LANGUAGE plpgsql;

-- Comments
COMMENT ON TABLE invoices IS 'Stores customer invoices for billing';
COMMENT ON TABLE invoice_items IS 'Line items for invoices';
COMMENT ON TABLE billing_settings IS 'Organization-level billing configuration';
COMMENT ON TABLE payment_records IS 'Audit trail for all payment transactions';
COMMENT ON TABLE invoice_sequence IS 'Auto-increment sequence for invoice numbers';
COMMENT ON FUNCTION generate_invoice_number() IS 'Auto-generates invoice numbers in format INV-YYYY-NNNNNN';
COMMENT ON FUNCTION update_overdue_invoices() IS 'Marks invoices as OVERDUE when past due date';
