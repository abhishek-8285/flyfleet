-- Add GPS tracking control columns to vehicles table
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS gps_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS gps_device_type TEXT;

-- Create an index to support uniqueness check (optional but good for performance)
CREATE INDEX IF NOT EXISTS idx_vehicles_gps_device_id ON vehicles (gps_device_id) WHERE gps_enabled = TRUE AND deleted_at IS NULL;

-- Note: We don't use a strict UNIQUE constraint because multiple vehicles might have had 
-- this device in the past (historical records). The uniqueness is enforced by business logic 
-- on ENABLED devices.
