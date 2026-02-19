-- migration: 042_billing_metering.sql
SET search_path TO public;

-- 1. Usage Meters: Definitions of billable metrics
CREATE TABLE IF NOT EXISTS usage_meters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL, -- e.g., 'DISTANCE_KM', 'API_CALL', 'DT_DOWNLOAD'
    name VARCHAR(100) NOT NULL,
    description TEXT,
    unit VARCHAR(20) NOT NULL, -- e.g., 'km', 'count', 'gb'
    unit_price DECIMAL(15, 4) NOT NULL DEFAULT 0.0000,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Seed basic meters
INSERT INTO usage_meters (code, name, unit, unit_price)
VALUES 
('DISTANCE_KM', 'Mileage', 'km', 0.15),
('HEARTBEAT_COUNT', 'Device Connectivity', 'count', 0.01)
ON CONFLICT (code) DO NOTHING;

-- 2. Consumption Events: High-volume usage logging
CREATE TABLE IF NOT EXISTS consumption_events (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    org_id UUID NOT NULL REFERENCES organizations(id),
    vehicle_id UUID REFERENCES vehicles(id),
    meter_code VARCHAR(50) NOT NULL REFERENCES usage_meters(code),
    amount DECIMAL(15, 4) NOT NULL,
    metadata JSONB,
    processed_invoice_id UUID -- Null until billed
);

-- 3. Optimization: Convert to Hypertable
SELECT create_hypertable('consumption_events', 'time', if_not_exists => TRUE);

-- 4. Enable RLS for tenant isolation
ALTER TABLE consumption_events ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS consumption_tenant_policy ON consumption_events;
CREATE POLICY consumption_tenant_policy ON consumption_events
    FOR ALL TO flyfleet_app
    USING (org_id = get_current_org_id())
    WITH CHECK (org_id = get_current_org_id());

-- 5. Indexes for aggregation
CREATE INDEX IF NOT EXISTS idx_consumption_aggregation 
ON consumption_events (org_id, meter_code, time DESC) 
WHERE processed_invoice_id IS NULL;

-- 6. Grant permissions
GRANT SELECT, INSERT, UPDATE ON usage_meters TO flyfleet_app;
GRANT SELECT, INSERT, UPDATE ON consumption_events TO flyfleet_app;
