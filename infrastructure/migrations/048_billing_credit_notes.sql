-- 048_billing_credit_notes.sql
-- Migration to support financial adjustments and refunds

CREATE TABLE IF NOT EXISTS billing_credit_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    original_invoice_id UUID REFERENCES billing_invoices(id),
    credit_note_no VARCHAR(50) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'issued', -- issued, void, applied
    total_amount NUMERIC(15, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL,
    reason TEXT,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS billing_credit_note_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    credit_note_id UUID NOT NULL REFERENCES billing_credit_notes(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    amount NUMERIC(15, 2) NOT NULL,
    tax_percent NUMERIC(5, 2) DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexing for performance
CREATE INDEX idx_credit_notes_org_id ON billing_credit_notes(org_id);
CREATE INDEX idx_credit_notes_invoice_id ON billing_credit_notes(invoice_id);

-- Enable RLS
ALTER TABLE billing_credit_notes ENABLE ROW LEVEL SECURITY;
ALTER TABLE billing_credit_note_items ENABLE ROW LEVEL SECURITY;

CREATE POLICY org_isolation_credit_notes ON billing_credit_notes 
    USING (org_id = CURRENT_SETTING('app.current_org_id')::UUID);

CREATE POLICY org_isolation_credit_note_items ON billing_credit_note_items 
    USING (credit_note_id IN (SELECT id FROM billing_credit_notes WHERE org_id = CURRENT_SETTING('app.current_org_id')::UUID));
