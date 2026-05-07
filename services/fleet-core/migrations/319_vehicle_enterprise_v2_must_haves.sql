-- +goose Up
-- +goose StatementBegin
-- Migration: 319_vehicle_enterprise_v2_must_haves.sql
-- Purpose: Add mandatory operational metrics and harden schema constraints.

-- 1. Add missing enterprise fields
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS risk_score DOUBLE PRECISION DEFAULT 0.0;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS is_blacklisted BOOLEAN DEFAULT FALSE;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS make VARCHAR(100);
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS model VARCHAR(100);
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS year INTEGER;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS color VARCHAR(50);

-- 2. Harden constraints (Odometer non-nullable)
UPDATE vehicles SET odometer = 0 WHERE odometer IS NULL;
ALTER TABLE vehicles ALTER COLUMN odometer SET NOT NULL;
ALTER TABLE vehicles ALTER COLUMN odometer SET DEFAULT 0;

-- 3. Normalize status validation
ALTER TABLE vehicles DROP CONSTRAINT IF EXISTS chk_vehicles_status;
ALTER TABLE vehicles ADD CONSTRAINT chk_vehicles_status 
    CHECK (status IN ('ACTIVE', 'MAINTENANCE', 'SUSPENDED', 'DECOMMISSIONED', 'IDLE', 'PROVISIONING', 'RETIRED'));

-- 4. Capacity & Fuel Tank constraints
UPDATE vehicles SET capacity_tonnage = 1.0 WHERE capacity_tonnage IS NULL OR capacity_tonnage <= 0;
ALTER TABLE vehicles ALTER COLUMN capacity_tonnage SET NOT NULL;
ALTER TABLE vehicles ALTER COLUMN capacity_tonnage SET DEFAULT 1.0;

UPDATE vehicles SET fuel_tank_capacity = 50.0 WHERE fuel_tank_capacity IS NULL OR fuel_tank_capacity <= 0;
ALTER TABLE vehicles ALTER COLUMN fuel_tank_capacity SET NOT NULL;
ALTER TABLE vehicles ALTER COLUMN fuel_tank_capacity SET DEFAULT 50.0;

-- +goose StatementEnd
