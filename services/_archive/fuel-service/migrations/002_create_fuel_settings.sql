-- Create fuel settings table
CREATE TABLE IF NOT EXISTS fuel_settings (
    organization_id UUID PRIMARY KEY,
    default_currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    allowed_fuel_types JSONB NOT NULL DEFAULT '["diesel", "petrol-95", "adblue", "ev"]',
    validation_rules JSONB NOT NULL DEFAULT '{"odometer_check": true, "cost_calculation_verify": true}',
    auto_sync_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Trigger for updated_at
CREATE TRIGGER update_fuel_settings_updated_at
    BEFORE UPDATE ON fuel_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
