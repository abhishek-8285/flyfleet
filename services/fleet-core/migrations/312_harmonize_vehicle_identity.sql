-- +goose Up
-- +goose StatementBegin

-- Harmonize identity columns in the vehicles table across the platform
DO $$
BEGIN
    -- 1. Add tenant_id if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'vehicles' AND column_name = 'tenant_id'
    ) THEN
        ALTER TABLE vehicles ADD COLUMN tenant_id UUID;
    END IF;

    -- 2. Backfill tenant_id from legacy columns if tenant_id is NULL
    -- Priority: org_id > organization_id
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'vehicles' AND column_name = 'org_id'
    ) THEN
        UPDATE vehicles SET tenant_id = COALESCE(tenant_id, org_id) WHERE tenant_id IS NULL;
    ELSIF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'vehicles' AND column_name = 'organization_id'
    ) THEN
        UPDATE vehicles SET tenant_id = COALESCE(tenant_id, organization_id) WHERE tenant_id IS NULL;
    END IF;

    -- 3. Set NOT NULL once backfilled (provided we have a fallback)
    -- If still NULL, use the first organization from organizations table as emergency fallback
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'organizations') THEN
        UPDATE vehicles SET tenant_id = (SELECT id FROM organizations LIMIT 1) WHERE tenant_id IS NULL;
    END IF;

    -- Final constraint
    IF EXISTS (SELECT 1 FROM vehicles WHERE tenant_id IS NULL) THEN
        RAISE NOTICE 'Some vehicles still have NULL tenant_id; cannot set NOT NULL constraint yet.';
    ELSE
        ALTER TABLE vehicles ALTER COLUMN tenant_id SET NOT NULL;
    END IF;

    -- 4. Clean up legacy columns
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'org_id') THEN
        ALTER TABLE vehicles DROP COLUMN org_id;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'organization_id') THEN
        ALTER TABLE vehicles DROP COLUMN organization_id;
    END IF;

    -- 5. Standardize Bulk Jobs Table
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'bulk_jobs' AND column_name = 'tenant_id') THEN
        ALTER TABLE bulk_jobs ADD COLUMN tenant_id UUID;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'bulk_jobs' AND column_name = 'organization_id') THEN
        UPDATE bulk_jobs SET tenant_id = COALESCE(tenant_id, organization_id) WHERE tenant_id IS NULL;
        ALTER TABLE bulk_jobs DROP COLUMN organization_id;
    END IF;
    ALTER TABLE bulk_jobs ALTER COLUMN tenant_id SET NOT NULL;

    -- 6. Standardize Trips Table
    -- The trips table already had tenant_id in some versions, but let's ensure it's uniform
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'trips' AND column_name = 'tenant_id') THEN
        ALTER TABLE trips ADD COLUMN tenant_id UUID;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'trips' AND column_name = 'organization_id') THEN
        UPDATE trips SET tenant_id = COALESCE(tenant_id, organization_id) WHERE tenant_id IS NULL;
        ALTER TABLE trips DROP COLUMN organization_id;
    END IF;
    -- Iforg_id exists as well
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'trips' AND column_name = 'org_id') THEN
        UPDATE trips SET tenant_id = COALESCE(tenant_id, org_id) WHERE tenant_id IS NULL;
        ALTER TABLE trips DROP COLUMN org_id;
    END IF;

    -- 8. Add missing operational columns to vehicles table
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'gps_enabled') THEN
        ALTER TABLE vehicles ADD COLUMN gps_enabled BOOLEAN DEFAULT FALSE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'gps_device_type') THEN
        ALTER TABLE vehicles ADD COLUMN gps_device_type VARCHAR(50);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'odometer') THEN
        ALTER TABLE vehicles ADD COLUMN odometer BIGINT DEFAULT 0;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'current_driver_id') THEN
        ALTER TABLE vehicles ADD COLUMN current_driver_id UUID;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'speed_limit') THEN
        ALTER TABLE vehicles ADD COLUMN speed_limit DOUBLE PRECISION DEFAULT 80.0;
    END IF;

    -- Final cleanup and indexing
    DROP INDEX IF EXISTS idx_vehicles_tenant_id;
    CREATE INDEX IF NOT EXISTS idx_vehicles_tenant_id ON vehicles(tenant_id);
    CREATE INDEX IF NOT EXISTS idx_bulk_jobs_tenant_id ON bulk_jobs(tenant_id);
    CREATE INDEX IF NOT EXISTS idx_trips_tenant_id ON trips(tenant_id);
    CREATE INDEX IF NOT EXISTS idx_vehicles_gps_enabled ON vehicles(gps_enabled);

END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS org_id UUID;
UPDATE vehicles SET org_id = tenant_id;
-- +goose StatementEnd
