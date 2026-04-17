-- Migration: 073_vehicle_enhancements
-- Description: Add odometer and current_driver_id to vehicles table for enterprise lifecycle management
-- Spec: Enterprise Vehicle Module

-- Add odometer column
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS odometer BIGINT DEFAULT 0;
COMMENT ON COLUMN vehicles.odometer IS 'Current vehicle odometer reading in meters';

-- Add current_driver_id column (to unify tracking and asset domains)
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS current_driver_id UUID REFERENCES drivers(id) ON DELETE SET NULL;
COMMENT ON COLUMN vehicles.current_driver_id IS 'ID of the driver currently assigned to this vehicle';

-- Add speed column if missing (often requested with odometer)
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS speed_limit DECIMAL(5,2) DEFAULT 0;
COMMENT ON COLUMN vehicles.speed_limit IS 'Configured speed limit for this vehicle (0 = no limit)';

-- Create index for driver lookup
CREATE INDEX IF NOT EXISTS idx_vehicles_current_driver ON vehicles(current_driver_id) WHERE current_driver_id IS NOT NULL;
