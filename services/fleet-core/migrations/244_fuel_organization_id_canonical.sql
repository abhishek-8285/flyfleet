BEGIN;

-- Canonicalize fuel schema to tenant_id and remove legacy identity drift.

-- fuel_logs
ALTER TABLE fuel_logs
    ADD COLUMN IF NOT EXISTS tenant_id UUID;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'fuel_logs' AND column_name = 'org_id'
    ) THEN
        EXECUTE 'UPDATE fuel_logs SET tenant_id = COALESCE(tenant_id, org_id)';
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
    fallback_tenant UUID;
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'organizations'
    ) THEN
        SELECT id INTO fallback_tenant FROM organizations LIMIT 1;
        IF fallback_tenant IS NOT NULL THEN
            UPDATE fuel_logs
            SET tenant_id = fallback_tenant
            WHERE tenant_id IS NULL;
        END IF;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM fuel_logs WHERE tenant_id IS NULL) THEN
        ALTER TABLE fuel_logs ALTER COLUMN tenant_id SET NOT NULL;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'organizations'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_schema = 'public' AND table_name = 'fuel_logs' AND constraint_name = 'fuel_logs_tenant_id_fkey'
    ) THEN
        ALTER TABLE fuel_logs
            ADD CONSTRAINT fuel_logs_tenant_id_fkey
            FOREIGN KEY (tenant_id) REFERENCES organizations(id);
    END IF;
END $$;

-- fuel_transactions
ALTER TABLE fuel_transactions
    ADD COLUMN IF NOT EXISTS tenant_id UUID;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'fuel_transactions' AND column_name = 'org_id'
    ) THEN
        EXECUTE 'UPDATE fuel_transactions SET tenant_id = COALESCE(tenant_id, org_id)';
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
    fallback_tenant UUID;
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'organizations'
    ) THEN
        SELECT id INTO fallback_tenant FROM organizations LIMIT 1;
        IF fallback_tenant IS NOT NULL THEN
            UPDATE fuel_transactions
            SET tenant_id = fallback_tenant
            WHERE tenant_id IS NULL;
        END IF;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM fuel_transactions WHERE tenant_id IS NULL) THEN
        ALTER TABLE fuel_transactions ALTER COLUMN tenant_id SET NOT NULL;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'organizations'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_schema = 'public' AND table_name = 'fuel_transactions' AND constraint_name = 'fuel_transactions_tenant_id_fkey'
    ) THEN
        ALTER TABLE fuel_transactions
            ADD CONSTRAINT fuel_transactions_tenant_id_fkey
            FOREIGN KEY (tenant_id) REFERENCES organizations(id);
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
        UPDATE fuel_settings SET tenant_id = COALESCE(tenant_id, org_id);
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
        UPDATE fuel_audit_logs SET tenant_id = COALESCE(tenant_id, org_id);
        ALTER TABLE fuel_audit_logs DROP COLUMN org_id;
    END IF;
END $$;

-- Normalize indexes to canonical column names.
DROP INDEX IF EXISTS idx_fuel_logs_org_id;
CREATE INDEX IF NOT EXISTS idx_fuel_logs_tenant_id ON fuel_logs(tenant_id);

DROP INDEX IF EXISTS idx_fuel_transactions_org_id;
CREATE INDEX IF NOT EXISTS idx_fuel_transactions_tenant_id ON fuel_transactions(tenant_id);

DROP INDEX IF EXISTS idx_fuel_audit_logs_org_id;
CREATE INDEX IF NOT EXISTS idx_fuel_audit_logs_tenant_id ON fuel_audit_logs(tenant_id);

COMMIT;
