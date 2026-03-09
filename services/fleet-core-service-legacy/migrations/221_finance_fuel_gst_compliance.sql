-- Migration: 221_finance_fuel_gst_compliance
-- Description: Add GST support to finance and fuel modules

-- Modify fuel_transactions
ALTER TABLE fuel_transactions 
ADD COLUMN IF NOT EXISTS gst_number VARCHAR(15),
ADD COLUMN IF NOT EXISTS gst_rate DECIMAL(5,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS gst_amount DECIMAL(10,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS invoice_number VARCHAR(100),
ADD COLUMN IF NOT EXISTS vendor_gstin VARCHAR(15);

-- Modify fuel_logs
ALTER TABLE fuel_logs 
ADD COLUMN IF NOT EXISTS gst_number VARCHAR(15),
ADD COLUMN IF NOT EXISTS gst_rate DECIMAL(5,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS gst_amount DECIMAL(10,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS invoice_number VARCHAR(100),
ADD COLUMN IF NOT EXISTS vendor_gstin VARCHAR(15);

-- Modify expense_reports
ALTER TABLE expense_reports 
ADD COLUMN IF NOT EXISTS gst_number VARCHAR(15),
ADD COLUMN IF NOT EXISTS gst_rate DECIMAL(5,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS gst_amount DECIMAL(10,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS invoice_number VARCHAR(100),
ADD COLUMN IF NOT EXISTS vendor_gstin VARCHAR(15);

-- Create finance_audit_logs (Phase 10, but good to have early if needed for audit trail of GST changes)
-- Actually, the request wants it in Phase 10. I'll stick to the phases as much as possible, 
-- but I'll create the GST migration first.

-- Comments for documentation
COMMENT ON COLUMN fuel_transactions.gst_number IS 'GSTIN of the purchaser';
COMMENT ON COLUMN fuel_transactions.vendor_gstin IS 'GSTIN of the fuel station/vendor';
COMMENT ON COLUMN fuel_logs.gst_number IS 'GSTIN of the purchaser';
COMMENT ON COLUMN fuel_logs.vendor_gstin IS 'GSTIN of the fuel station/vendor';
COMMENT ON COLUMN expense_reports.gst_number IS 'GSTIN associated with the expense';
