-- Migration: 064b_driver_management
-- Description: Complete driver management schema with tenant isolation and behavior scoring.
-- Note: Relocated from 064_driver_management.sql to resolve numeric conflict.

-- 1. Create drivers table
DROP VIEW IF EXISTS drivers CASCADE; -- Safety check for view-to-table transitions
CREATE TABLE IF NOT EXISTS drivers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4()
);

DO $$ 
BEGIN 
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL REFERENCES tenants(id);
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS first_name VARCHAR(50) NOT NULL DEFAULT '';
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS last_name VARCHAR(50) NOT NULL DEFAULT '';
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS full_name VARCHAR(100) NOT NULL DEFAULT '';
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS email VARCHAR(255) NOT NULL DEFAULT '';
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS phone VARCHAR(20) NOT NULL DEFAULT '';
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS license_number VARCHAR(50) NOT NULL DEFAULT '';
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS license_type VARCHAR(20);
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS license_expiry DATE NOT NULL DEFAULT CURRENT_DATE;
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS license_authority VARCHAR(100) NOT NULL DEFAULT '';
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS date_of_birth DATE NOT NULL DEFAULT '1970-01-01';
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS address JSONB NOT NULL DEFAULT '{}';
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS emergency_contact JSONB NOT NULL DEFAULT '{}';
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active';
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS current_vehicle_id UUID;
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS current_trip_id UUID;
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS joining_date DATE NOT NULL DEFAULT CURRENT_DATE;
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS resignation_date DATE;
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS experience_years INTEGER DEFAULT 0;
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS rating DECIMAL(3, 2) DEFAULT 0.00;
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS total_trips INTEGER DEFAULT 0;
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS total_distance_km DECIMAL(10, 2) DEFAULT 0.00;
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS total_hours_driven DECIMAL(10, 2) DEFAULT 0.00;
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS behavior_score INTEGER DEFAULT 100;
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS documents JSONB DEFAULT '[]'::jsonb;
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS attendance_records JSONB DEFAULT '[]'::jsonb;
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS schedules JSONB DEFAULT '[]'::jsonb;
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS last_known_location JSONB;
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS last_location_update TIMESTAMP WITH TIME ZONE;
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}'::jsonb;
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
    ALTER TABLE drivers ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- 2. Driver behavior analytics
CREATE TABLE IF NOT EXISTS driver_behavior (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    score DECIMAL(5, 2) DEFAULT 0.00 CHECK (score >= 0 AND score <= 100),
    safety_score DECIMAL(5, 2) DEFAULT 0.00,
    efficiency_score DECIMAL(5, 2) DEFAULT 0.00,
    risk_level VARCHAR(20) DEFAULT 'low' CHECK (risk_level IN ('low', 'medium', 'high', 'critical')),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 3. Driver audit logs
CREATE TABLE IF NOT EXISTS driver_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    old_value JSONB,
    new_value JSONB,
    user_id UUID,
    user_type VARCHAR(20),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 4. Indexes
CREATE INDEX IF NOT EXISTS idx_drivers_tenant ON drivers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_driver_behavior_tenant ON driver_behavior(tenant_id);
CREATE INDEX IF NOT EXISTS idx_driver_audit_tenant ON driver_audit_logs(tenant_id);

-- 5. Views
CREATE OR REPLACE VIEW driver_performance_summary AS
SELECT 
    d.id,
    d.tenant_id,
    d.full_name,
    d.status,
    d.rating,
    d.behavior_score,
    d.total_trips,
    d.total_distance_km
FROM drivers d
WHERE d.deleted_at IS NULL;
