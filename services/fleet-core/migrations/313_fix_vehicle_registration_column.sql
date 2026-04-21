-- +goose Up
-- +goose StatementBegin

-- Ensure the vehicles table has the expected schema for the Enterprise Repository
DO $$
BEGIN
    -- 1. Ensure registration_number exists
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'vehicles' AND column_name = 'registration_number'
    ) THEN
        ALTER TABLE vehicles ADD COLUMN registration_number VARCHAR(50);
        
        -- Backfill from license_plate if available
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'license_plate') THEN
            UPDATE vehicles SET registration_number = license_plate WHERE registration_number IS NULL;
        END IF;
    END IF;

    -- 2. Ensure missing operational columns from previous migrations are present
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'vin_number') THEN
        ALTER TABLE vehicles ADD COLUMN vin_number VARCHAR(100);
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'vin') THEN
            UPDATE vehicles SET vin_number = vin WHERE vin_number IS NULL;
        END IF;
    END IF;

    -- 3. Ensure NOT NULL constraint for registration_number (with fallback)
    UPDATE vehicles SET registration_number = 'UNKNOWN-' || id::text WHERE registration_number IS NULL OR registration_number = '';
    ALTER TABLE vehicles ALTER COLUMN registration_number SET NOT NULL;

    -- 4. Create missing indexes for performance
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_vehicles_registration_number') THEN
        CREATE INDEX idx_vehicles_registration_number ON vehicles(registration_number);
    END IF;

END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- No destructive down migration for core schema fixes
-- +goose StatementEnd
