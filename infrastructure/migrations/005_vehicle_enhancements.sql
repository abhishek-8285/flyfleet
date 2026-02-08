-- Vehicle Enhancements Migration
-- Created: 2026-02-08
-- Description: Adds VIN, Odometer, Service Due, etc. to vehicles table

ALTER TABLE vehicles 
ADD COLUMN IF NOT EXISTS vehicle_no VARCHAR(50) UNIQUE,
ADD COLUMN IF NOT EXISTS vin VARCHAR(50) UNIQUE,
ADD COLUMN IF NOT EXISTS odometer BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS service_due TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS fuel_type VARCHAR(50),
ADD COLUMN IF NOT EXISTS notes TEXT;

CREATE INDEX IF NOT EXISTS idx_vehicles_vin ON vehicles(vin);
CREATE INDEX IF NOT EXISTS idx_vehicles_vehicle_no ON vehicles(vehicle_no);
