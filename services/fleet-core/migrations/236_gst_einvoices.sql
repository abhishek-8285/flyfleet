-- Migration: 236_gst_einvoices
-- Description: GST e-invoice table for IRN generation and digital signing

CREATE TABLE IF NOT EXISTS gst_invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_number VARCHAR(50) NOT NULL UNIQUE,
    irn VARCHAR(64),                                -- Invoice Reference Number from IRP
    ack_number VARCHAR(30),                         -- Acknowledgement Number from IRP
    ack_date TIMESTAMP WITH TIME ZONE,
    qr_code TEXT,                                    -- Base64 encoded QR code
    signed_invoice TEXT,                             -- Signed JSON payload from IRP
    status VARCHAR(20) DEFAULT 'PENDING',           -- PENDING, GENERATED, CANCELLED, ERROR
    entity_type VARCHAR(30),                         -- fuel_transaction, expense_report
    entity_id UUID,
    org_id UUID NOT NULL,
    gstin VARCHAR(15) NOT NULL,
    buyer_gstin VARCHAR(15),
    total_amount DECIMAL(15, 2) NOT NULL DEFAULT 0,
    taxable_amount DECIMAL(15, 2) NOT NULL DEFAULT 0,
    cgst_amount DECIMAL(15, 2) DEFAULT 0,
    sgst_amount DECIMAL(15, 2) DEFAULT 0,
    igst_amount DECIMAL(15, 2) DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_gst_invoices_org ON gst_invoices(org_id);
CREATE INDEX idx_gst_invoices_entity ON gst_invoices(entity_type, entity_id);
CREATE INDEX idx_gst_invoices_status ON gst_invoices(status);
CREATE INDEX idx_gst_invoices_irn ON gst_invoices(irn) WHERE irn IS NOT NULL;

COMMENT ON TABLE gst_invoices IS 'GST e-invoices with IRN from Invoice Registration Portal';
