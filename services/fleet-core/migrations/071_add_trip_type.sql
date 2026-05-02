-- +goose Up
-- +goose StatementBegin

-- Add trip_type column for categorizing trips
ALTER TABLE trips ADD COLUMN IF NOT EXISTS trip_type VARCHAR(50) DEFAULT 'standard';

-- Add check constraint for valid trip types
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.constraint_column_usage WHERE table_name = 'trips' AND constraint_name = 'chk_trips_trip_type') THEN
        ALTER TABLE trips ADD CONSTRAINT chk_trips_trip_type 
            CHECK (trip_type IN ('delivery', 'service', 'long_haul', 'standard', 'local', 'return'));
    END IF;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- Add index for filtering by trip_type
CREATE INDEX IF NOT EXISTS idx_trips_type_status ON trips(trip_type, status);
CREATE INDEX IF NOT EXISTS idx_trips_tenant_type ON trips(tenant_id, trip_type);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_trips_tenant_type;
DROP INDEX IF EXISTS idx_trips_type_status;
ALTER TABLE trips DROP CONSTRAINT IF EXISTS chk_trips_trip_type;
ALTER TABLE trips DROP COLUMN IF EXISTS trip_type;
-- +goose StatementEnd
