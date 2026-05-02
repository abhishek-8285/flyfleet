BEGIN;

-- Pre-flight: Ensure identity columns exist for core tables used in this script
DO $$
DECLARE
    r RECORD;
BEGIN
    -- fuel_logs, fuel_transactions, fuel_audit_logs, fuel_settings
    FOR r IN (SELECT table_name FROM (VALUES ('fuel_logs'), ('fuel_transactions'), ('fuel_audit_logs'), ('fuel_settings')) AS t(table_name))
    LOOP
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = r.table_name) THEN
            IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = r.table_name AND column_name = 'tenant_id') THEN
                EXECUTE format('ALTER TABLE %I ADD COLUMN tenant_id UUID', r.table_name);
            END IF;
            IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = r.table_name AND column_name = 'org_id') THEN
                EXECUTE format('ALTER TABLE %I ADD COLUMN org_id UUID', r.table_name);
            END IF;
        END IF;
    END LOOP;
END $$;

-- -----------------------------------------------------------------------------
-- fuel_logs
-- -----------------------------------------------------------------------------
-- Redundant: Handled by 223b Universal Shield
-- DO $$ BEGIN EXECUTE 'UPDATE fuel_logs SET tenant_id = COALESCE(tenant_id, org_id)'; END $$;

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

-- DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'fuel_logs' AND column_name = 'org_id') THEN ALTER TABLE fuel_logs DROP COLUMN org_id; END IF; END $$;

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
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- -----------------------------------------------------------------------------
-- fuel_transactions
-- -----------------------------------------------------------------------------
-- Redundant: Handled by 223b Universal Shield
-- DO $$ BEGIN EXECUTE 'UPDATE fuel_transactions SET tenant_id = COALESCE(tenant_id, org_id)'; END $$;

-- DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'fuel_transactions' AND column_name = 'org_id') THEN ALTER TABLE fuel_transactions DROP COLUMN org_id; END IF; END $$;

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
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- -----------------------------------------------------------------------------
-- fuel_settings & fuel_audit_logs
-- -----------------------------------------------------------------------------
-- DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'fuel_settings' AND column_name = 'org_id') THEN ALTER TABLE fuel_settings DROP COLUMN org_id; END IF; END $$;
-- DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'fuel_audit_logs' AND column_name = 'org_id') THEN ALTER TABLE fuel_audit_logs DROP COLUMN org_id; END IF; END $$;

-- -----------------------------------------------------------------------------
-- Normalize indexes
-- -----------------------------------------------------------------------------
DROP INDEX IF EXISTS idx_fuel_logs_org_id;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'fuel_logs' AND column_name = 'tenant_id') THEN CREATE INDEX IF NOT EXISTS idx_fuel_logs_tenant_id ON fuel_logs(tenant_id); END IF; END $$;

DROP INDEX IF EXISTS idx_fuel_transactions_org_id;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'fuel_transactions' AND column_name = 'tenant_id') THEN CREATE INDEX IF NOT EXISTS idx_fuel_transactions_tenant_id ON fuel_transactions(tenant_id); END IF; END $$;

DROP INDEX IF EXISTS idx_fuel_audit_logs_org_id;
DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'fuel_audit_logs' AND column_name = 'tenant_id') THEN CREATE INDEX IF NOT EXISTS idx_fuel_audit_logs_tenant_id ON fuel_audit_logs(tenant_id); END IF; END $$;

COMMIT;
