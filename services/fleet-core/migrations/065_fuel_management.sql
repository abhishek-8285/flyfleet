-- Migration: 065_fuel_management
-- Description: Creates tables for fuel logs with tenant isolation.

-- 1. Fuel logs
CREATE TABLE IF NOT EXISTS fuel_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4()
);

DO $$ 
BEGIN 
    ALTER TABLE fuel_logs ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL REFERENCES tenants(id);
    ALTER TABLE fuel_logs ADD COLUMN IF NOT EXISTS vehicle_id UUID NOT NULL;
    ALTER TABLE fuel_logs ADD COLUMN IF NOT EXISTS driver_id UUID NOT NULL;
    ALTER TABLE fuel_logs ADD COLUMN IF NOT EXISTS station_name VARCHAR(255) NOT NULL DEFAULT '';
    ALTER TABLE fuel_logs ADD COLUMN IF NOT EXISTS fuel_type VARCHAR(50) NOT NULL DEFAULT 'diesel';
    ALTER TABLE fuel_logs ADD COLUMN IF NOT EXISTS quantity DECIMAL(10, 2) NOT NULL DEFAULT 0;
    ALTER TABLE fuel_logs ADD COLUMN IF NOT EXISTS unit_price DECIMAL(10, 2) NOT NULL DEFAULT 0;
    ALTER TABLE fuel_logs ADD COLUMN IF NOT EXISTS total_cost DECIMAL(10, 2) NOT NULL DEFAULT 0;
    ALTER TABLE fuel_logs ADD COLUMN IF NOT EXISTS odometer DECIMAL(10, 2) NOT NULL DEFAULT 0;
    ALTER TABLE fuel_logs ADD COLUMN IF NOT EXISTS transaction_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
    ALTER TABLE fuel_logs ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'PENDING';
    ALTER TABLE fuel_logs ADD COLUMN IF NOT EXISTS receipt_url TEXT;
    ALTER TABLE fuel_logs ADD COLUMN IF NOT EXISTS notes TEXT;
    ALTER TABLE fuel_logs ADD COLUMN IF NOT EXISTS created_by UUID;
    ALTER TABLE fuel_logs ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
    ALTER TABLE fuel_logs ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_fuel_logs_tenant ON fuel_logs(tenant_id);

-- 2. Fuel settings
CREATE TABLE IF NOT EXISTS fuel_settings (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id),
    default_currency VARCHAR(3) DEFAULT 'USD',
    allowed_fuel_types JSONB DEFAULT '["diesel", "petrol-95", "adblue", "ev"]',
    validation_rules JSONB DEFAULT '{"odometer_check": true, "cost_calculation_verify": true}',
    auto_sync_enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Fuel audit logs
CREATE TABLE IF NOT EXISTS fuel_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    fuel_log_id UUID NOT NULL REFERENCES fuel_logs(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    changed_by UUID NOT NULL,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    old_values JSONB,
    new_values JSONB
);

-- 4. Views
CREATE OR REPLACE VIEW fuel_log_stats AS
SELECT
    tenant_id,
    COUNT(*) FILTER (WHERE status = 'PENDING') as pending_count,
    COUNT(*) FILTER (WHERE status = 'APPROVED') as approved_count,
    SUM(quantity) FILTER (WHERE status = 'APPROVED') as total_consumption,
    SUM(total_cost) FILTER (WHERE status = 'APPROVED') as total_cost
FROM fuel_logs
GROUP BY tenant_id;
