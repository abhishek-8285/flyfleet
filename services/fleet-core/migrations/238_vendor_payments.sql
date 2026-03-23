-- Migration: 238_vendor_payments
-- Description: Vendor payment ledger for fuel stations and service providers

CREATE TABLE IF NOT EXISTS vendor_payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vendor_id UUID NOT NULL REFERENCES fuel_vendors(id) ON DELETE RESTRICT,
    org_id UUID NOT NULL,
    invoice_number VARCHAR(50) NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    tds_deducted DECIMAL(15, 2) DEFAULT 0,
    net_payable DECIMAL(15, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'INR',
    payment_status VARCHAR(20) DEFAULT 'PENDING',   -- PENDING, PROCESSING, PAID, FAILED, REVERSED
    payment_mode VARCHAR(30),                        -- NEFT, RTGS, IMPS, CHEQUE, CASH
    payment_reference VARCHAR(100),                  -- UTR, cheque number, etc.
    bank_account VARCHAR(30),
    ifsc_code VARCHAR(11),
    payment_date DATE,
    paid_at TIMESTAMP WITH TIME ZONE,
    gst_invoice_id UUID REFERENCES gst_invoices(id),
    tds_record_id UUID REFERENCES tds_records(id),
    notes TEXT,
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_vendor_payments_vendor ON vendor_payments(vendor_id);
CREATE INDEX idx_vendor_payments_org ON vendor_payments(org_id);
CREATE INDEX idx_vendor_payments_status ON vendor_payments(payment_status);
CREATE INDEX idx_vendor_payments_date ON vendor_payments(payment_date);

COMMENT ON TABLE vendor_payments IS 'Vendor payment ledger tracking all payments to fuel stations and providers';
