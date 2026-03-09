-- Migration: 235_currency_rates
-- Description: Create currency_rates table for multi-currency support

CREATE TABLE IF NOT EXISTS currency_rates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_currency VARCHAR(3) NOT NULL,
    to_currency VARCHAR(3) NOT NULL,
    rate DECIMAL(18, 6) NOT NULL,
    provider VARCHAR(50),
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(from_currency, to_currency)
);

-- Index
CREATE INDEX idx_currency_rates_pair ON currency_rates(from_currency, to_currency);

-- Seed data for INR/USD (Stub)
INSERT INTO currency_rates (from_currency, to_currency, rate, provider) 
VALUES ('USD', 'INR', 83.00, 'MOCK')
ON CONFLICT (from_currency, to_currency) DO UPDATE SET rate = 83.00, last_updated = CURRENT_TIMESTAMP;

INSERT INTO currency_rates (from_currency, to_currency, rate, provider) 
VALUES ('INR', 'USD', 0.012, 'MOCK')
ON CONFLICT (from_currency, to_currency) DO UPDATE SET rate = 0.012, last_updated = CURRENT_TIMESTAMP;
