-- Migration: 239_expense_items
-- Description: Granular expense line items for expense reports

CREATE TABLE IF NOT EXISTS expense_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    report_id UUID NOT NULL REFERENCES expense_reports(id) ON DELETE CASCADE,
    category VARCHAR(30) NOT NULL CHECK (category IN ('FUEL','MEALS','LODGING','TOLL','MAINTENANCE','OTHER')),
    amount DECIMAL(15, 2) NOT NULL,
    gst_amount DECIMAL(15, 2) DEFAULT 0,
    gst_rate DECIMAL(5, 2) DEFAULT 0,
    receipt_url TEXT,
    description TEXT,
    vendor_name VARCHAR(255),
    vendor_gstin VARCHAR(15),
    transaction_date DATE NOT NULL DEFAULT CURRENT_DATE,
    ocr_processed BOOLEAN DEFAULT FALSE,
    ocr_confidence DECIMAL(4, 2),                   -- 0.00–1.00
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_expense_items_report ON expense_items(report_id);
CREATE INDEX idx_expense_items_category ON expense_items(category);

COMMENT ON TABLE expense_items IS 'Granular line items for each expense report';
