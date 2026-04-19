-- Add category to invoices
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS category text DEFAULT 'General';

-- Update status constraint to be case-insensitive or allow uppercase
ALTER TABLE invoices DROP CONSTRAINT IF EXISTS chk_invoices_status;
ALTER TABLE invoices ADD CONSTRAINT chk_invoices_status CHECK (
    status = ANY (ARRAY['DRAFT', 'PENDING', 'PAID', 'OVERDUE', 'PARTIAL', 'VOID', 'REFUNDED', 'draft', 'paid', 'sent', 'viewed', 'overdue', 'void', 'refunded'])
);

-- Ensure billing_settings defaults
ALTER TABLE billing_settings ALTER COLUMN default_tax_rate SET DEFAULT 18.00;
ALTER TABLE billing_settings ALTER COLUMN base_currency SET DEFAULT 'USD';
