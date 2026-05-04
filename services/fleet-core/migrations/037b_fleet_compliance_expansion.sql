-- Migration: 037b_fleet_compliance_expansion
-- Goal: Upgrade vehicles and drivers with compliance and service tracking fields.
-- Note: Relocated from 037_fleet_compliance_expansion.sql to resolve numeric conflict.

-- VEHICLES: Compliance & Maintenance
ALTER TABLE vehicles 
ADD COLUMN IF NOT EXISTS insurance VARCHAR(100),
ADD COLUMN IF NOT EXISTS insurance_expiry_date TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS permit_number VARCHAR(100),
ADD COLUMN IF NOT EXISTS permit_expiry_date TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS pollution_expiry_date TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS fitness_expiry_date TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS tax_expiry_date TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS last_service_date TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS next_service_date TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS last_service_odometer BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}';

-- DRIVERS: Verification & Bio-Metrics
ALTER TABLE drivers 
ADD COLUMN IF NOT EXISTS last_medical_checkup_date TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS police_verification_date TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS police_verification_status VARCHAR(50) DEFAULT 'PENDING',
ADD COLUMN IF NOT EXISTS blood_group VARCHAR(10);

-- Ensure indexes for expiry tracking (Critical for alerts)
CREATE INDEX IF NOT EXISTS idx_vehicles_insurance_expiry ON vehicles(insurance_expiry_date);
CREATE INDEX IF NOT EXISTS idx_vehicles_permit_expiry ON vehicles(permit_expiry_date);
CREATE INDEX IF NOT EXISTS idx_vehicles_fitness_expiry ON vehicles(fitness_expiry_date);
