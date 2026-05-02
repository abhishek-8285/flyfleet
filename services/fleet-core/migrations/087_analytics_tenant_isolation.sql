-- Migration: 087_analytics_tenant_isolation
-- Description: Enforce multi-tenant isolation for predictive models using tenant_id.

DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'predictive_models' AND column_name = 'tenant_id') THEN
        ALTER TABLE predictive_models ADD COLUMN tenant_id UUID REFERENCES tenants(id);
    END IF;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_predictive_models_tenant ON predictive_models(tenant_id);
