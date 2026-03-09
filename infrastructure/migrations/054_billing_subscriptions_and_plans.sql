-- migration: 054_billing_subscriptions_and_plans.sql
SET search_path TO public;

-- 1. Billing Plans
CREATE TABLE IF NOT EXISTS billing_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Pricing Components (Atomic charges within a plan)
CREATE TABLE IF NOT EXISTS billing_pricing_components (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES billing_plans(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL, -- FLAT, PER_UNIT, USAGE
    unit_price DECIMAL(15, 4) NOT NULL DEFAULT 0.0000,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    meter_event_name VARCHAR(50), -- Only for USAGE type
    frequency VARCHAR(20) NOT NULL, -- MONTHLY, YEARLY, ONE_TIME
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Billing Subscriptions (Organization-Plan link)
CREATE TABLE IF NOT EXISTS billing_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    plan_id UUID NOT NULL REFERENCES billing_plans(id),
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE', -- ACTIVE, TRIAL, CANCELLED, PAST_DUE
    current_period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    current_period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    trial_end TIMESTAMP WITH TIME ZONE,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (org_id) -- One active subscription per org
);

-- RLS
ALTER TABLE billing_subscriptions ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS subs_tenant_policy ON billing_subscriptions;
CREATE POLICY subs_tenant_policy ON billing_subscriptions
    FOR ALL TO flyfleet_app
    USING (org_id = get_current_org_id())
    WITH CHECK (org_id = get_current_org_id());

-- Indexes
CREATE INDEX IF NOT EXISTS idx_subs_org_id ON billing_subscriptions(org_id);
CREATE INDEX IF NOT EXISTS idx_subs_plan_id ON billing_subscriptions(plan_id);
CREATE INDEX IF NOT EXISTS idx_components_plan_id ON billing_pricing_components(plan_id);

-- Permissions
GRANT SELECT, INSERT, UPDATE ON billing_plans TO flyfleet_app;
GRANT SELECT, INSERT, UPDATE ON billing_pricing_components TO flyfleet_app;
GRANT SELECT, INSERT, UPDATE ON billing_subscriptions TO flyfleet_app;
