-- Migration: 222_fuel_card_integration
-- Description: Create fuel_card_transactions table and integration layer support

CREATE TABLE IF NOT EXISTS fuel_card_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id UUID,
    card_id VARCHAR(100) NOT NULL,
    provider VARCHAR(50) NOT NULL, -- IOCL, HPCL, BPCL
    station VARCHAR(255),
    liters DECIMAL(10, 2) NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'INR',
    transaction_time TIMESTAMP WITH TIME ZONE NOT NULL,
    synced_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    external_transaction_id VARCHAR(255) UNIQUE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_fuel_card_tx_card_id ON fuel_card_transactions(card_id);
CREATE INDEX idx_fuel_card_tx_vehicle_id ON fuel_card_transactions(vehicle_id);
CREATE INDEX idx_fuel_card_tx_time ON fuel_card_transactions(transaction_time);
CREATE INDEX idx_fuel_card_tx_provider ON fuel_card_transactions(provider);

-- Comments
COMMENT ON TABLE fuel_card_transactions IS 'Transactions synced from external fuel card providers (IOCL, HPCL, BPCL)';
COMMENT ON COLUMN fuel_card_transactions.provider IS 'Indian fuel card providers: IOCL, HPCL, BPCL';
