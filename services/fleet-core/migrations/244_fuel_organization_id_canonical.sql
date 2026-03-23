BEGIN;

-- Canonicalize fuel schema to organization_id and remove legacy org_id drift.

-- fuel_logs
ALTER TABLE fuel_logs
    ADD COLUMN IF NOT EXISTS organization_id UUID;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'fuel_logs' AND column_name = 'org_id'
    ) THEN
        EXECUTE 'UPDATE fuel_logs SET organization_id = COALESCE(organization_id, org_id)';
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.table_constraints
        WHERE table_schema = 'public' AND table_name = 'fuel_logs' AND constraint_name = 'fuel_logs_org_id_fkey'
    ) THEN
        ALTER TABLE fuel_logs DROP CONSTRAINT fuel_logs_org_id_fkey;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'fuel_logs' AND column_name = 'org_id'
    ) THEN
        ALTER TABLE fuel_logs DROP COLUMN org_id;
    END IF;
END $$;

DO $$
DECLARE
    fallback_org UUID;
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'organizations'
    ) THEN
        SELECT id INTO fallback_org FROM organizations LIMIT 1;
        IF fallback_org IS NOT NULL THEN
            UPDATE fuel_logs
            SET organization_id = fallback_org
            WHERE organization_id IS NULL;
        END IF;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM fuel_logs WHERE organization_id IS NULL) THEN
        ALTER TABLE fuel_logs ALTER COLUMN organization_id SET NOT NULL;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'organizations'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_schema = 'public' AND table_name = 'fuel_logs' AND constraint_name = 'fuel_logs_organization_id_fkey'
    ) THEN
        ALTER TABLE fuel_logs
            ADD CONSTRAINT fuel_logs_organization_id_fkey
            FOREIGN KEY (organization_id) REFERENCES organizations(id);
    END IF;
END $$;

-- fuel_transactions
ALTER TABLE fuel_transactions
    ADD COLUMN IF NOT EXISTS organization_id UUID;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'fuel_transactions' AND column_name = 'org_id'
    ) THEN
        EXECUTE 'UPDATE fuel_transactions SET organization_id = COALESCE(organization_id, org_id)';
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'fuel_transactions' AND column_name = 'org_id'
    ) THEN
        ALTER TABLE fuel_transactions DROP COLUMN org_id;
    END IF;
END $$;

DO $$
DECLARE
    fallback_org UUID;
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'organizations'
    ) THEN
        SELECT id INTO fallback_org FROM organizations LIMIT 1;
        IF fallback_org IS NOT NULL THEN
            UPDATE fuel_transactions
            SET organization_id = fallback_org
            WHERE organization_id IS NULL;
        END IF;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM fuel_transactions WHERE organization_id IS NULL) THEN
        ALTER TABLE fuel_transactions ALTER COLUMN organization_id SET NOT NULL;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'organizations'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_schema = 'public' AND table_name = 'fuel_transactions' AND constraint_name = 'fuel_transactions_organization_id_fkey'
    ) THEN
        ALTER TABLE fuel_transactions
            ADD CONSTRAINT fuel_transactions_organization_id_fkey
            FOREIGN KEY (organization_id) REFERENCES organizations(id);
    END IF;
END $$;

-- fuel_settings already canonical; keep explicit guard for drifted environments.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'fuel_settings' AND column_name = 'org_id'
    ) THEN
        UPDATE fuel_settings SET organization_id = COALESCE(organization_id, org_id);
        ALTER TABLE fuel_settings DROP COLUMN org_id;
    END IF;
END $$;

-- fuel_audit_logs already canonical; keep explicit guard for drifted environments.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'fuel_audit_logs' AND column_name = 'org_id'
    ) THEN
        UPDATE fuel_audit_logs SET organization_id = COALESCE(organization_id, org_id);
        ALTER TABLE fuel_audit_logs DROP COLUMN org_id;
    END IF;
END $$;

-- Normalize indexes to canonical column names.
DROP INDEX IF EXISTS idx_fuel_logs_org_id;
CREATE INDEX IF NOT EXISTS idx_fuel_logs_organization_id ON fuel_logs(organization_id);

DROP INDEX IF EXISTS idx_fuel_transactions_org_id;
CREATE INDEX IF NOT EXISTS idx_fuel_transactions_organization_id ON fuel_transactions(organization_id);

DROP INDEX IF EXISTS idx_fuel_audit_logs_org_id;
CREATE INDEX IF NOT EXISTS idx_fuel_audit_logs_organization_id ON fuel_audit_logs(organization_id);

COMMIT;
