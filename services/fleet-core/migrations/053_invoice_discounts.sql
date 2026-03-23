-- Migration: 053_invoice_discounts
-- Description: Add discount support for invoices

-- Add discount columns to invoices table
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS discount_type VARCHAR(20) DEFAULT 'none'; -- none, percentage, fixed
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS discount_value DECIMAL(12, 2) DEFAULT 0;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS discount_reason TEXT;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS subtotal DECIMAL(12, 2) DEFAULT 0;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS discount_amount DECIMAL(12, 2) DEFAULT 0;

-- Add constraint for valid discount type
ALTER TABLE invoices ADD CONSTRAINT chk_invoice_discount_type 
    CHECK (discount_type IN ('none', 'percentage', 'fixed'));

-- Add constraint for discount value range
ALTER TABLE invoices ADD CONSTRAINT chk_invoice_discount_value 
    CHECK (discount_value >= 0 AND (discount_type = 'percentage' AND discount_value <= 100 OR discount_type != 'percentage'));

-- Add constraint for discount amount
ALTER TABLE invoices ADD CONSTRAINT chk_invoice_discount_amount 
    CHECK (discount_amount >= 0 AND discount_amount <= subtotal);

-- Comments
COMMENT ON COLUMN invoices.discount_type IS 'Type of discount: none, percentage, or fixed';
COMMENT ON COLUMN invoices.discount_value IS 'Discount value (percentage or fixed amount)';
COMMENT ON COLUMN invoices.discount_reason IS 'Reason for applying discount';
COMMENT ON COLUMN invoices.subtotal IS 'Total before discount';
COMMENT ON COLUMN invoices.discount_amount IS 'Actual discount amount applied';

-- Index for querying discounted invoices
CREATE INDEX IF NOT EXISTS idx_invoices_discount ON invoices(discount_type, discount_value);
