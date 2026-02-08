-- Migration: 004_create_maintenance_settings.sql
-- Creates organization-level maintenance settings

CREATE TABLE IF NOT EXISTS maintenance_settings (
    organization_id UUID PRIMARY KEY,
    sla_p1_hours INT NOT NULL DEFAULT 4,
    sla_p2_hours INT NOT NULL DEFAULT 24,
    sla_p3_hours INT NOT NULL DEFAULT 72,
    low_stock_threshold INT NOT NULL DEFAULT 5,
    approval_threshold DECIMAL(10, 2) NOT NULL DEFAULT 500.00,
    auto_assign_enabled BOOLEAN NOT NULL DEFAULT false,
    notification_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Updated_at trigger
CREATE TRIGGER trigger_maintenance_settings_updated_at
    BEFORE UPDATE ON maintenance_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_work_orders_updated_at();

COMMENT ON TABLE maintenance_settings IS 'Organization maintenance configuration';
