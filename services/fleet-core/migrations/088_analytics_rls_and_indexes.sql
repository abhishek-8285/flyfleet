-- Migration: 088_analytics_rls_and_indexes
-- Description: Enable RLS and add composite indexes for tenant isolation.

-- 1. Enable RLS
ALTER TABLE predictive_models ENABLE ROW LEVEL SECURITY;
ALTER TABLE vehicle_daily_stats ENABLE ROW LEVEL SECURITY;
ALTER TABLE driver_daily_stats ENABLE ROW LEVEL SECURITY;
ALTER TABLE fleet_hourly_stats ENABLE ROW LEVEL SECURITY;

-- 2. Create RLS Policies using app.tenant_id
DO $$ 
BEGIN 
    DROP POLICY IF EXISTS tenant_isolation_models ON predictive_models;
    CREATE POLICY tenant_isolation_models ON predictive_models 
        USING (tenant_id = current_setting('app.tenant_id', true)::UUID);

    DROP POLICY IF EXISTS tenant_isolation_vehicle_stats ON vehicle_daily_stats;
    CREATE POLICY tenant_isolation_vehicle_stats ON vehicle_daily_stats 
        USING (tenant_id = current_setting('app.tenant_id', true)::UUID);

    DROP POLICY IF EXISTS tenant_isolation_driver_stats ON driver_daily_stats;
    CREATE POLICY tenant_isolation_driver_stats ON driver_daily_stats 
        USING (tenant_id = current_setting('app.tenant_id', true)::UUID);
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- 3. Composite Indexes
CREATE INDEX IF NOT EXISTS idx_pred_models_tenant_status ON predictive_models(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_vehicle_stats_tenant_date ON vehicle_daily_stats(tenant_id, date);
