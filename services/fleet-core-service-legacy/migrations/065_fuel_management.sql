-- Migration: 065_fuel_management
-- Description: Creates tables for fuel logs, settings, and audit trail

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Create fuel_logs table
CREATE TABLE IF NOT EXISTS fuel_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL,
    vehicle_id UUID NOT NULL,
    driver_id UUID NOT NULL,
    station_name VARCHAR(255) NOT NULL,
    fuel_type VARCHAR(50) NOT NULL,
    quantity DECIMAL(10, 2) NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    total_cost DECIMAL(10, 2) NOT NULL,
    odometer DECIMAL(10, 2) NOT NULL,
    transaction_date TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED')),
    receipt_url TEXT,
    notes TEXT,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Create fuel_settings table
CREATE TABLE IF NOT EXISTS fuel_settings (
    organization_id UUID PRIMARY KEY,
    default_currency VARCHAR(3) DEFAULT 'USD',
    allowed_fuel_types JSONB DEFAULT '["diesel", "petrol-95", "adblue", "ev"]',
    validation_rules JSONB DEFAULT '{"odometer_check": true, "cost_calculation_verify": true}',
    auto_sync_enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Create fuel_audit_logs table for audit trail
CREATE TABLE IF NOT EXISTS fuel_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    fuel_log_id UUID NOT NULL REFERENCES fuel_logs(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL, -- created, updated, deleted, approved, rejected
    changed_by UUID NOT NULL,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    old_values JSONB,
    new_values JSONB,
    diff JSONB
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_fuel_logs_org ON fuel_logs(organization_id);
CREATE INDEX IF NOT EXISTS idx_fuel_logs_vehicle ON fuel_logs(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_fuel_logs_driver ON fuel_logs(driver_id);
CREATE INDEX IF NOT EXISTS idx_fuel_logs_status ON fuel_logs(status);
CREATE INDEX IF NOT EXISTS idx_fuel_logs_transaction_date ON fuel_logs(transaction_date);
CREATE INDEX IF NOT EXISTS idx_fuel_logs_org_status ON fuel_logs(organization_id, status);
CREATE INDEX IF NOT EXISTS idx_fuel_logs_org_vehicle ON fuel_logs(organization_id, vehicle_id);

CREATE INDEX IF NOT EXISTS idx_fuel_audit_log_fuel_log ON fuel_audit_logs(fuel_log_id);
CREATE INDEX IF NOT EXISTS idx_fuel_audit_log_action ON fuel_audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_fuel_audit_log_changed_at ON fuel_audit_logs(changed_at);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_fuel_logs_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_fuel_logs_modtime
BEFORE UPDATE ON fuel_logs
FOR EACH ROW
EXECUTE FUNCTION update_fuel_logs_timestamp();

CREATE TRIGGER update_fuel_settings_modtime
BEFORE UPDATE ON fuel_settings
FOR EACH ROW
EXECUTE FUNCTION update_fuel_logs_timestamp();

-- Comments
COMMENT ON TABLE fuel_logs IS 'Fuel transaction logs for fleet vehicles';
COMMENT ON TABLE fuel_settings IS 'Organization-level fuel management settings';
COMMENT ON TABLE fuel_audit_logs IS 'Audit trail for fuel log changes';

COMMENT ON COLUMN fuel_logs.status IS 'PENDING=awaiting approval, APPROVED=approved, REJECTED=rejected';
COMMENT ON COLUMN fuel_logs.odometer IS 'Vehicle odometer reading at time of fueling';
COMMENT ON COLUMN fuel_settings.validation_rules IS 'JSON validation rules for fuel transactions';
COMMENT ON COLUMN fuel_settings.allowed_fuel_types IS 'JSON array of allowed fuel types';

-- Function to get last odometer reading for a vehicle
CREATE OR REPLACE FUNCTION get_last_odometer(p_vehicle_id UUID, p_org_id UUID)
RETURNS DECIMAL AS $$
DECLARE
    last_odometer DECIMAL;
BEGIN
    SELECT odometer INTO last_odometer
    FROM fuel_logs
    WHERE vehicle_id = p_vehicle_id 
      AND organization_id = p_org_id
      AND status = 'APPROVED'
    ORDER BY transaction_date DESC, created_at DESC
    LIMIT 1;
    
    RETURN COALESCE(last_odometer, 0);
END;
$$ LANGUAGE plpgsql;

-- Function to calculate fuel efficiency (km per liter)
CREATE OR REPLACE FUNCTION calculate_fuel_efficiency(
    p_current_odometer DECIMAL,
    p_previous_odometer DECIMAL,
    p_fuel_quantity DECIMAL
)
RETURNS DECIMAL AS $$
BEGIN
    IF p_fuel_quantity <= 0 THEN
        RETURN 0;
    END IF;
    
    IF p_current_odometer <= p_previous_odometer THEN
        RETURN 0;
    END IF;
    
    RETURN ROUND((p_current_odometer - p_previous_odometer) / p_fuel_quantity, 2);
END;
$$ LANGUAGE plpgsql;

-- View for fuel log statistics
CREATE OR REPLACE VIEW fuel_log_stats AS
SELECT
    organization_id,
    COUNT(*) FILTER (WHERE status = 'PENDING') as pending_count,
    COUNT(*) FILTER (WHERE status = 'APPROVED') as approved_count,
    COUNT(*) FILTER (WHERE status = 'REJECTED') as rejected_count,
    SUM(quantity) FILTER (WHERE status = 'APPROVED') as total_consumption,
    SUM(total_cost) FILTER (WHERE status = 'APPROVED') as total_cost,
    AVG(total_cost / NULLIF(quantity, 0)) FILTER (WHERE status = 'APPROVED') as avg_price_per_unit,
    COUNT(DISTINCT vehicle_id) as unique_vehicles,
    COUNT(DISTINCT driver_id) as unique_drivers
FROM fuel_logs
GROUP BY organization_id;

-- View for fuel consumption by vehicle
CREATE OR REPLACE VIEW fuel_consumption_by_vehicle AS
SELECT
    organization_id,
    vehicle_id,
    COUNT(*) as transaction_count,
    SUM(quantity) as total_consumption,
    SUM(total_cost) as total_cost,
    AVG(total_cost / NULLIF(quantity, 0)) as avg_price,
    MAX(transaction_date) as last_fuel_date,
    MAX(odometer) as last_odometer
FROM fuel_logs
WHERE status = 'APPROVED'
GROUP BY organization_id, vehicle_id;

-- View for fuel consumption by driver
CREATE OR REPLACE VIEW fuel_consumption_by_driver AS
SELECT
    organization_id,
    driver_id,
    COUNT(*) as transaction_count,
    SUM(quantity) as total_consumption,
    SUM(total_cost) as total_cost,
    AVG(total_cost / NULLIF(quantity, 0)) as avg_price,
    MAX(transaction_date) as last_fuel_date
FROM fuel_logs
WHERE status = 'APPROVED'
GROUP BY organization_id, driver_id;

-- View for fuel consumption by fuel type
CREATE OR REPLACE VIEW fuel_consumption_by_type AS
SELECT
    organization_id,
    fuel_type,
    COUNT(*) as transaction_count,
    SUM(quantity) as total_consumption,
    SUM(total_cost) as total_cost,
    AVG(total_cost / NULLIF(quantity, 0)) as avg_price
FROM fuel_logs
WHERE status = 'APPROVED'
GROUP BY organization_id, fuel_type;

-- View for monthly fuel consumption
CREATE OR REPLACE VIEW fuel_consumption_monthly AS
SELECT
    organization_id,
    DATE_TRUNC('month', transaction_date) as month,
    COUNT(*) as transaction_count,
    SUM(quantity) as total_consumption,
    SUM(total_cost) as total_cost,
    AVG(total_cost / NULLIF(quantity, 0)) as avg_price
FROM fuel_logs
WHERE status = 'APPROVED'
GROUP BY organization_id, DATE_TRUNC('month', transaction_date)
ORDER BY month DESC;
