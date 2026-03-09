-- Add organization_id to predictive_models to enforce multi-tenant isolation
ALTER TABLE predictive_models ADD COLUMN IF NOT EXISTS organization_id UUID;

-- Update existing models to a default system organization or keep as NULL if global
-- For now, we assume models should be isolated, but we might want some global ones.
-- However, the requirement is strict isolation.
-- If someone already created models, we might need a migration strategy, but this is a dev environment.
-- UPDATE predictive_models SET organization_id = '00000000-0000-0000-0000-000000000000' WHERE organization_id IS NULL;

-- Create index for organization_id
CREATE INDEX IF NOT EXISTS idx_predictive_models_org_id ON predictive_models(organization_id);
