-- Migration: 071_vehicle_enterprise
-- Description: Enterprise-grade Vehicle Asset & Compliance Management with tenant isolation.

-- 1. Resilient Vehicle Table (Ensuring all columns exist)
DO $$ 
BEGIN 
    -- Add enterprise columns if they don't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'vehicle_type') THEN
        ALTER TABLE vehicles ADD COLUMN vehicle_type VARCHAR(30) DEFAULT 'TRUCK';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'fuel_type') THEN
        ALTER TABLE vehicles ADD COLUMN fuel_type VARCHAR(20) DEFAULT 'DIESEL';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'ownership_type') THEN
        ALTER TABLE vehicles ADD COLUMN ownership_type VARCHAR(20) DEFAULT 'OWNED';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'gps_device_id') THEN
        ALTER TABLE vehicles ADD COLUMN gps_device_id VARCHAR(100);
    END IF;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- 2. Compliance table
CREATE TABLE IF NOT EXISTS vehicle_compliance (
    vehicle_id UUID PRIMARY KEY REFERENCES vehicles(id) ON DELETE CASCADE,
    rc_expiry_date DATE,
    insurance_expiry_date DATE,
    puc_expiry_date DATE,
    fitness_expiry_date DATE,
    permit_expiry_date DATE,
    permit_type VARCHAR(50),
    is_blacklisted BOOLEAN DEFAULT FALSE,
    legal_hold BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. Financials table
CREATE TABLE IF NOT EXISTS vehicle_financials (
    vehicle_id UUID PRIMARY KEY REFERENCES vehicles(id) ON DELETE CASCADE,
    acquisition_cost DECIMAL(15,2),
    total_maintenance_cost DECIMAL(15,2) DEFAULT 0,
    total_fuel_cost DECIMAL(15,2) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 4. Documents table
CREATE TABLE IF NOT EXISTS vehicle_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    document_type VARCHAR(30) NOT NULL,
    file_url TEXT NOT NULL,
    uploaded_at TIMESTAMPTZ DEFAULT NOW()
);

-- 5. Indexes (Non-concurrent for migration compatibility)
CREATE INDEX IF NOT EXISTS idx_vehicles_tenant ON vehicles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_compliance_rc ON vehicle_compliance(rc_expiry_date);
CREATE INDEX IF NOT EXISTS idx_vdoc_vehicle ON vehicle_documents(vehicle_id);

-- 6. Compliance View
CREATE OR REPLACE VIEW vehicles_due_for_compliance AS
SELECT
    v.id AS vehicle_id,
    v.tenant_id,
    v.registration_number,
    v.status,
    c.rc_expiry_date,
    c.insurance_expiry_date
FROM vehicles v
JOIN vehicle_compliance c ON c.vehicle_id = v.id
WHERE v.deleted_at IS NULL;
