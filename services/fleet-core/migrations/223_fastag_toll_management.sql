-- Migration: 223_fastag_toll_management
-- Description: Create toll_transactions table for FASTag integration

CREATE TABLE IF NOT EXISTS toll_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id UUID NOT NULL,
    fastag_id VARCHAR(100) NOT NULL,
    toll_plaza VARCHAR(255) NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'INR',
    transaction_time TIMESTAMP WITH TIME ZONE NOT NULL,
    external_transaction_id VARCHAR(255) UNIQUE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_toll_vehicle_id ON toll_transactions(vehicle_id);
CREATE INDEX idx_toll_fastag_id ON toll_transactions(fastag_id);
CREATE INDEX idx_toll_time ON toll_transactions(transaction_time);

-- Comments
COMMENT ON TABLE toll_transactions IS 'Toll transactions ingested from FASTag providers';
