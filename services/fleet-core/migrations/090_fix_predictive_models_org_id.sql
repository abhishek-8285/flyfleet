-- Migration: 090_fix_predictive_models_org_id
-- Description: Add organization_id to predictive_models for compatibility with Phase 8-11 migrations.

DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'predictive_models' AND column_name = 'organization_id') THEN
        ALTER TABLE predictive_models ADD COLUMN organization_id UUID;
    END IF;
END $$;
