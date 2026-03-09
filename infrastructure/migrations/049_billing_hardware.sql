-- 049_billing_hardware.sql
-- Migration to support hardware sales and lease-to-own models

CREATE TABLE IF NOT EXISTS billing_hardware_products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    model_code VARCHAR(100) UNIQUE NOT NULL,
    category VARCHAR(50) NOT NULL, -- DASHCAM, TRACKER, SENSOR
    one_time_price NUMERIC(15, 2) NOT NULL,
    lease_monthly_price NUMERIC(15, 2),
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS billing_device_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    device_id VARCHAR(100) NOT NULL, -- Links to devices.device_id
    product_id UUID NOT NULL REFERENCES billing_hardware_products(id),
    billing_type VARCHAR(20) NOT NULL, -- ONE_TIME, LEASE
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING_ACTIVATION', -- PENDING, ACTIVE, TERMINATED
    activated_at TIMESTAMPTZ,
    invoice_id UUID REFERENCES billing_invoices(id), -- Null if leased and billed in monthly runs
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(device_id)
);

-- Seed some hardware products
INSERT INTO billing_hardware_products (name, model_code, category, one_time_price, lease_monthly_price)
VALUES 
('FlyFleet Vision Pro Dashcam', 'FF-CAM-V2', 'DASHCAM', 399.00, 25.00),
('FlyFleet Standard GPS Tracker', 'FF-GPS-100', 'TRACKER', 99.00, 10.00),
('FlyFleet OBD-II Telematics Node', 'FF-OBD-V1', 'TRACKER', 149.00, 15.00)
ON CONFLICT (model_code) DO NOTHING;

-- Indexing for lookup
CREATE INDEX idx_device_assignments_org_id ON billing_device_assignments(org_id);
CREATE INDEX idx_device_assignments_device_id ON billing_device_assignments(device_id);

-- RLS
ALTER TABLE billing_hardware_products ENABLE ROW LEVEL SECURITY;
ALTER TABLE billing_device_assignments ENABLE ROW LEVEL SECURITY;

CREATE POLICY hardware_products_read_all ON billing_hardware_products FOR SELECT USING (true);
CREATE POLICY org_isolation_device_assignments ON billing_device_assignments 
    USING (org_id = CURRENT_SETTING('app.current_org_id')::UUID);
