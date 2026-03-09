-- migration: 052_billing_tax_and_sequencing.sql
SET search_path TO public;

-- 1. Tax Regions (e.g., India, EU, US-California)
CREATE TABLE IF NOT EXISTS billing_tax_regions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    country_code VARCHAR(2) NOT NULL,
    state_code VARCHAR(10), -- Optional
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Tax Rates
CREATE TABLE IF NOT EXISTS billing_tax_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    region_id UUID NOT NULL REFERENCES billing_tax_regions(id),
    name VARCHAR(100) NOT NULL, -- e.g., GST, VAT, Sales Tax
    percentage DECIMAL(5, 2) NOT NULL,
    tax_type VARCHAR(50) NOT NULL, -- e.g., CGST, SGST, IGST, VAT
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Invoice Tax Lines (Breakdown per invoice)
CREATE TABLE IF NOT EXISTS billing_invoice_tax_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES billing_invoices(id) ON DELETE CASCADE,
    tax_rate_id UUID NOT NULL REFERENCES billing_tax_rates(id),
    name VARCHAR(100) NOT NULL,
    amount DECIMAL(15, 4) NOT NULL,
    percentage DECIMAL(5, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 4. Invoice Sequencing (Ensures gapless, sequential numbering per Org)
CREATE TABLE IF NOT EXISTS billing_invoice_sequences (
    org_id UUID PRIMARY KEY REFERENCES organizations(id),
    prefix VARCHAR(20) NOT NULL DEFAULT 'INV-',
    next_value BIGINT NOT NULL DEFAULT 1,
    suffix VARCHAR(20) DEFAULT '',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- RLS Policies
ALTER TABLE billing_invoice_tax_lines ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tax_lines_tenant_policy ON billing_invoice_tax_lines;
CREATE POLICY tax_lines_tenant_policy ON billing_invoice_tax_lines
    FOR ALL TO flyfleet_app
    USING (EXISTS (SELECT 1 FROM billing_invoices WHERE id = invoice_id AND org_id = get_current_org_id()));

ALTER TABLE billing_invoice_sequences ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS sequences_tenant_policy ON billing_invoice_sequences;
CREATE POLICY sequences_tenant_policy ON billing_invoice_sequences
    FOR ALL TO flyfleet_app
    USING (org_id = get_current_org_id())
    WITH CHECK (org_id = get_current_org_id());

-- Note: billing_tax_regions and billing_tax_rates are system-wide, no RLS needed.

-- Indexes
CREATE INDEX IF NOT EXISTS idx_tax_lines_invoice ON billing_invoice_tax_lines(invoice_id);
CREATE INDEX IF NOT EXISTS idx_tax_rates_region ON billing_tax_rates(region_id);

-- Permissions
GRANT SELECT ON billing_tax_regions TO flyfleet_app;
GRANT SELECT ON billing_tax_rates TO flyfleet_app;
GRANT SELECT, INSERT ON billing_invoice_tax_lines TO flyfleet_app;
GRANT SELECT, INSERT, UPDATE ON billing_invoice_sequences TO flyfleet_app;
