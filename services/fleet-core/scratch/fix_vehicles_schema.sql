DO $$
BEGIN
    -- 1. Ensure tenant_id exists
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'vehicles' AND column_name = 'tenant_id'
    ) THEN
        ALTER TABLE vehicles ADD COLUMN tenant_id UUID;
    END IF;

    -- 2. Backfill tenant_id if it's null
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'vehicles' AND column_name = 'org_id'
    ) THEN
        UPDATE vehicles SET tenant_id = org_id WHERE tenant_id IS NULL;
    END IF;
    
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'vehicles' AND column_name = 'organization_id'
    ) THEN
        UPDATE vehicles SET tenant_id = organization_id WHERE tenant_id IS NULL;
    END IF;

    -- 3. Drop NOT NULL constraint on org_id, then drop the column safely
    -- Drop org_id
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'vehicles' AND column_name = 'org_id'
    ) THEN
        ALTER TABLE vehicles ALTER COLUMN org_id DROP NOT NULL;
        ALTER TABLE vehicles DROP COLUMN org_id CASCADE;
    END IF;

    -- Drop organization_id if present
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'vehicles' AND column_name = 'organization_id'
    ) THEN
        ALTER TABLE vehicles ALTER COLUMN organization_id DROP NOT NULL;
        ALTER TABLE vehicles DROP COLUMN organization_id CASCADE;
    END IF;

    -- 4. Ensure tenant_id is NOT NULL (only if all are filled, else leave it for now)
    IF NOT EXISTS (SELECT 1 FROM vehicles WHERE tenant_id IS NULL) THEN
        ALTER TABLE vehicles ALTER COLUMN tenant_id SET NOT NULL;
    END IF;
END $$;
