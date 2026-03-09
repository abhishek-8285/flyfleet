-- Migration: 227_driver_fuel_limits
-- Description: Create fuel_limits table and update driver metadata

CREATE TABLE IF NOT EXISTS driver_fuel_limits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL,
    driver_id UUID NOT NULL, -- Specifically for individual drivers
    daily_liters_limit DECIMAL(10, 2),
    monthly_liters_limit DECIMAL(10, 2),
    daily_amount_limit DECIMAL(10, 2),
    monthly_amount_limit DECIMAL(10, 2),
    currency VARCHAR(3) DEFAULT 'INR',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, driver_id)
);

-- Organization level defaults
CREATE TABLE IF NOT EXISTS org_fuel_settings (
    organization_id UUID PRIMARY KEY,
    default_daily_liters_limit DECIMAL(10, 2),
    default_monthly_liters_limit DECIMAL(10, 2),
    default_daily_amount_limit DECIMAL(10, 2),
    default_monthly_amount_limit DECIMAL(10, 2),
    currency VARCHAR(3) DEFAULT 'INR',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_driver_fuel_limits_driver ON driver_fuel_limits(driver_id);

-- Comments
COMMENT ON TABLE driver_fuel_limits IS 'Daily and monthly fuel spending and volume limits per driver';
COMMENT ON TABLE org_fuel_settings IS 'Organization-wide default fuel limits';
