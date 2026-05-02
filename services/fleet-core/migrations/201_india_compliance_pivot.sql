-- Migration: 201_india_compliance_pivot
-- Description: Additive India-specific compliance modules with tenant isolation.

-- 1. Trip Sheets (Consolidated regional workflow)
CREATE TABLE IF NOT EXISTS trip_sheets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id),
    driver_id UUID NOT NULL REFERENCES drivers(id),
    status VARCHAR(50) DEFAULT 'CREATED',
    origin TEXT,
    destination TEXT,
    distance_km DECIMAL(15,2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_trip_sheets_tenant ON trip_sheets(tenant_id);

-- 2. AIS-140 Device Metadata (Extensions for device_registry)
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'device_registry' AND column_name = 'imei') THEN
        ALTER TABLE device_registry ADD COLUMN imei TEXT UNIQUE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'device_registry' AND column_name = 'sim_number') THEN
        ALTER TABLE device_registry ADD COLUMN sim_number TEXT;
    END IF;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- 3. Panic Alerts
CREATE TABLE IF NOT EXISTS panic_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id),
    driver_id UUID REFERENCES drivers(id),
    status VARCHAR(50) DEFAULT 'ACTIVE',
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 4. Toll Transactions (FASTag)
CREATE TABLE IF NOT EXISTS toll_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id),
    fastag_id TEXT NOT NULL,
    plaza_name TEXT,
    amount DECIMAL(15,2) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_toll_tenant ON toll_transactions(tenant_id);
