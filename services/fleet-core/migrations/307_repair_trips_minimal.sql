-- +goose Up
-- +goose StatementBegin
-- Repairs the trips table by adding missing columns required by the repository scans
-- This fix addresses the 500 Internal Server Error (column "start_time" does not exist)

ALTER TABLE trips ADD COLUMN IF NOT EXISTS trip_type VARCHAR(50) DEFAULT 'standard';
ALTER TABLE trips ADD COLUMN IF NOT EXISTS start_time TIMESTAMP WITH TIME ZONE;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS end_time TIMESTAMP WITH TIME ZONE;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS planned_eta TIMESTAMP WITH TIME ZONE;

ALTER TABLE trips ADD COLUMN IF NOT EXISTS actual_route JSONB;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS cargo_details JSONB;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS start_location JSONB;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS end_location JSONB;

ALTER TABLE trips ADD COLUMN IF NOT EXISTS total_distance DOUBLE PRECISION;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS actual_duration_min INTEGER;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS start_odometer DOUBLE PRECISION;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS end_odometer DOUBLE PRECISION;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS fuel_level_snapshot DOUBLE PRECISION;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS fuel_used DOUBLE PRECISION;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS idling_time_minutes INTEGER;

ALTER TABLE trips ADD COLUMN IF NOT EXISTS eway_bill_no VARCHAR(100);
ALTER TABLE trips ADD COLUMN IF NOT EXISTS gst_classification VARCHAR(100);
ALTER TABLE trips ADD COLUMN IF NOT EXISTS permit_verified BOOLEAN DEFAULT FALSE;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS ais140_compliant BOOLEAN DEFAULT FALSE;

ALTER TABLE trips ADD COLUMN IF NOT EXISTS customer_name VARCHAR(255);
ALTER TABLE trips ADD COLUMN IF NOT EXISTS customer_phone VARCHAR(50);
ALTER TABLE trips ADD COLUMN IF NOT EXISTS customer_email VARCHAR(255);
ALTER TABLE trips ADD COLUMN IF NOT EXISTS cancellation_reason TEXT;

ALTER TABLE trips ADD COLUMN IF NOT EXISTS correlation_id UUID DEFAULT gen_random_uuid();
ALTER TABLE trips ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;

-- Lat/Lng fields for repository compliance (replaces/complements pickup_geo)
ALTER TABLE trips ADD COLUMN IF NOT EXISTS pickup_lat DOUBLE PRECISION;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS pickup_lng DOUBLE PRECISION;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS drop_lat DOUBLE PRECISION;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS drop_lng DOUBLE PRECISION;

-- Ensure constraints match expectations
ALTER TABLE trips DROP CONSTRAINT IF EXISTS chk_trips_trip_type;
ALTER TABLE trips ADD CONSTRAINT chk_trips_trip_type 
    CHECK (trip_type IN ('delivery', 'service', 'long_haul', 'standard', 'local', 'return'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- We don't drop columns in down to prevent data loss in repair scenarios
-- +goose StatementEnd
