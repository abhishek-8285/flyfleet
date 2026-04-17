-- Migration: 305_vehicle_operational_fields.sql
-- Add operational fields to vehicles table

ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS odometer bigint DEFAULT 0;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS current_driver_id uuid;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS speed_limit double precision DEFAULT 0.0;
