-- Migration: 237_tds_records
-- Description: TDS (Tax Deducted at Source) records for vendor payments

CREATE TABLE IF NOT EXISTS tds_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(30) NOT NULL,               -- vendor_payment, expense_report
    entity_id UUID NOT NULL,
    vendor_id UUID,
    org_id UUID NOT NULL,
    pan_number VARCHAR(10),                         -- PAN of deductee
    tds_section VARCHAR(10) DEFAULT '194C',         -- TDS section (194C, 194I, etc.)
    gross_amount DECIMAL(15, 2) NOT NULL,
    tds_rate DECIMAL(5, 2) NOT NULL,
    tds_amount DECIMAL(15, 2) NOT NULL,
    net_amount DECIMAL(15, 2) NOT NULL,
    deducted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deposited BOOLEAN DEFAULT FALSE,
    deposit_challan VARCHAR(50),
    financial_year VARCHAR(10),                     -- e.g. '2024-25'
    quarter VARCHAR(5),                             -- Q1, Q2, Q3, Q4
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tds_records_entity ON tds_records(entity_type, entity_id);
CREATE INDEX idx_tds_records_org ON tds_records(org_id);
CREATE INDEX idx_tds_records_vendor ON tds_records(vendor_id) WHERE vendor_id IS NOT NULL;
CREATE INDEX idx_tds_records_fy ON tds_records(financial_year, quarter);

COMMENT ON TABLE tds_records IS 'TDS deduction records for vendor payments per Indian tax law';
