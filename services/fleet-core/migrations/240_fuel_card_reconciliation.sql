-- Migration: 240_fuel_card_reconciliation
-- Description: Fuel card reconciliation tracking

CREATE TABLE IF NOT EXISTS fuel_card_reconciliation (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    card_transaction_id UUID,                       -- from fuel_card_transactions
    internal_transaction_id UUID,                   -- from fuel_logs or fuel_transactions
    org_id UUID NOT NULL,
    status VARCHAR(20) DEFAULT 'PENDING',           -- PENDING, MATCHED, UNMATCHED, DISPUTED
    variance DECIMAL(15, 2) DEFAULT 0,              -- difference in amounts
    card_amount DECIMAL(15, 2),
    internal_amount DECIMAL(15, 2),
    reconciled_by UUID,
    reconciled_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_card_recon_org ON fuel_card_reconciliation(org_id);
CREATE INDEX idx_card_recon_status ON fuel_card_reconciliation(status);
CREATE INDEX idx_card_recon_card_tx ON fuel_card_reconciliation(card_transaction_id) WHERE card_transaction_id IS NOT NULL;
CREATE INDEX idx_card_recon_internal_tx ON fuel_card_reconciliation(internal_transaction_id) WHERE internal_transaction_id IS NOT NULL;

COMMENT ON TABLE fuel_card_reconciliation IS 'Fuel card reconciliation - matches external card charges with internal fuel logs';
