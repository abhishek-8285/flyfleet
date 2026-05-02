-- Migration: 094b_analytics_schema_harmonization
-- Description: Ensure organization_id exists on all analytics tables used in RLS policies.

DO $$ 
BEGIN 
    -- 1. predictive_models
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='predictive_models' AND column_name='organization_id') THEN
        ALTER TABLE predictive_models ADD COLUMN organization_id UUID;
    END IF;

    -- 2. prediction_results
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='prediction_results' AND column_name='organization_id') THEN
        ALTER TABLE prediction_results ADD COLUMN organization_id UUID;
    END IF;

    -- 3. analytics_jobs
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='analytics_jobs' AND column_name='organization_id') THEN
        ALTER TABLE analytics_jobs ADD COLUMN organization_id UUID;
    END IF;

    -- 4. fleet_analytics
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='fleet_analytics' AND column_name='organization_id') THEN
        ALTER TABLE fleet_analytics ADD COLUMN organization_id UUID;
    END IF;

    -- 5. driver_behavior_analytics
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='driver_behavior_analytics' AND column_name='organization_id') THEN
        ALTER TABLE driver_behavior_analytics ADD COLUMN organization_id UUID;
    END IF;

    -- 6. predictive_maintenance_analytics
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='predictive_maintenance_analytics' AND column_name='organization_id') THEN
        ALTER TABLE predictive_maintenance_analytics ADD COLUMN organization_id UUID;
    END IF;

    -- 7. analytics_query_logs
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='analytics_query_logs' AND column_name='organization_id') THEN
        ALTER TABLE analytics_query_logs ADD COLUMN organization_id UUID;
    END IF;

    -- 8. ml_feature_store
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='ml_feature_store' AND column_name='organization_id') THEN
        ALTER TABLE ml_feature_store ADD COLUMN organization_id UUID;
    END IF;

    -- 9. analytics_quotas
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='analytics_quotas' AND column_name='organization_id') THEN
        ALTER TABLE analytics_quotas ADD COLUMN organization_id UUID;
    END IF;

    -- 10. tenant_id synchronization (Best practice for FlyFleet)
    -- If organization_id exists but tenant_id is missing, add and sync.
    -- This helps later migrations that expect tenant_id.
    
    -- We'll do this for the main ones.
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='predictive_models' AND column_name='organization_id') AND 
       NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='predictive_models' AND column_name='tenant_id') THEN
        ALTER TABLE predictive_models ADD COLUMN tenant_id UUID;
    END IF;
    UPDATE predictive_models SET tenant_id = organization_id WHERE tenant_id IS NULL AND organization_id IS NOT NULL;

END $$;
