-- MASTER IDENTITY HARMONIZATION MIGRATION
-- This migration removes NOT NULL constraints from all legacy identity columns across the entire database.
-- This is a required prerequisite for standardizing the Go backend to use tenant_id exclusively.

-- 1. Analytics Domain
DO $$ BEGIN ALTER TABLE analytics_jobs ALTER COLUMN organization_id DROP NOT NULL; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE predictive_models ALTER COLUMN organization_id DROP NOT NULL; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE ml_feature_store ALTER COLUMN organization_id DROP NOT NULL; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE driver_risk_events ALTER COLUMN organization_id DROP NOT NULL; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE analytics_quotas ALTER COLUMN organization_id DROP NOT NULL; EXCEPTION WHEN OTHERS THEN NULL; END $$;

-- 2. Governance Domain
DO $$ BEGIN ALTER TABLE governance_policies ALTER COLUMN organization_id DROP NOT NULL; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE audit_logs ALTER COLUMN organization_id DROP NOT NULL; EXCEPTION WHEN OTHERS THEN NULL; END $$;

-- 3. Emergency Domain
DO $$ BEGIN ALTER TABLE emergency_incidents ALTER COLUMN organization_id DROP NOT NULL; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE emergency_contacts ALTER COLUMN organization_id DROP NOT NULL; EXCEPTION WHEN OTHERS THEN NULL; END $$;

-- 4. Billing & Customer Domain
DO $$ BEGIN ALTER TABLE customers ALTER COLUMN org_id DROP NOT NULL; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE billing_subscriptions ALTER COLUMN org_id DROP NOT NULL; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE billing_device_assignments ALTER COLUMN org_id DROP NOT NULL; EXCEPTION WHEN OTHERS THEN NULL; END $$;

-- 5. Support & Admin Domain
DO $$ BEGIN ALTER TABLE support_tickets ALTER COLUMN organization_id DROP NOT NULL; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE system_audit_logs ALTER COLUMN organization_id DROP NOT NULL; EXCEPTION WHEN OTHERS THEN NULL; END $$;

-- 6. Core Operations (Additional tables)
DO $$ BEGIN ALTER TABLE geofences ALTER COLUMN org_id DROP NOT NULL; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE routes ALTER COLUMN organization_id DROP NOT NULL; EXCEPTION WHEN OTHERS THEN NULL; END $$;
DO $$ BEGIN ALTER TABLE maintenance_logs ALTER COLUMN organization_id DROP NOT NULL; EXCEPTION WHEN OTHERS THEN NULL; END $$;
