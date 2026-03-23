-- Migration: 101_billing_subscriptions
-- Description: Add infrastructure for subscription plans and organization subscriptions

-- Create billing_plans table
CREATE TABLE IF NOT EXISTS billing_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Create billing_pricing_components table
CREATE TABLE IF NOT EXISTS billing_pricing_components (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES billing_plans(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL, -- FLAT, PER_UNIT, USAGE
    unit_price DECIMAL(12, 4) NOT NULL DEFAULT 0,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    meter_event_name VARCHAR(100), -- for type=USAGE
    frequency VARCHAR(50) NOT NULL DEFAULT 'MONTHLY', -- MONTHLY, YEARLY, ONE_TIME
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Create billing_subscriptions table
CREATE TABLE IF NOT EXISTS billing_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL UNIQUE,
    plan_id UUID NOT NULL REFERENCES billing_plans(id),
    status VARCHAR(50) NOT NULL DEFAULT 'TRIAL', -- ACTIVE, TRIAL, CANCELLED, PAST_DUE
    current_period_start TIMESTAMPTZ NOT NULL,
    current_period_end TIMESTAMPTZ NOT NULL,
    trial_end TIMESTAMPTZ,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT false,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Add subscription_id to invoices
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS subscription_id UUID REFERENCES billing_subscriptions(id);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_billing_plans_is_active ON billing_plans(is_active);
CREATE INDEX IF NOT EXISTS idx_billing_pricing_components_plan_id ON billing_pricing_components(plan_id);
CREATE INDEX IF NOT EXISTS idx_billing_subscriptions_org_id ON billing_subscriptions(org_id);
CREATE INDEX IF NOT EXISTS idx_billing_subscriptions_plan_id ON billing_subscriptions(plan_id);
CREATE INDEX IF NOT EXISTS idx_billing_subscriptions_status ON billing_subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_invoices_subscription_id ON invoices(subscription_id);

-- Seed initial plans
INSERT INTO billing_plans (name, description) VALUES 
('Starter', 'Basic fleet tracking for up to 5 vehicles'),
('Growth', 'Advanced features for growing fleets'),
('Enterprise', 'Custom scale solutions with premium support');

-- Seed pricing components for Starter Plan (ID will need to be queried)
DO $$
DECLARE
    starter_id UUID;
    growth_id UUID;
    enterprise_id UUID;
BEGIN
    SELECT id INTO starter_id FROM billing_plans WHERE name = 'Starter';
    SELECT id INTO growth_id FROM billing_plans WHERE name = 'Growth';
    SELECT id INTO enterprise_id FROM billing_plans WHERE name = 'Enterprise';

    -- Starter Components
    INSERT INTO billing_pricing_components (plan_id, name, type, unit_price, frequency)
    VALUES (starter_id, 'Platform Fee', 'FLAT', 49.00, 'MONTHLY');

    -- Growth Components
    INSERT INTO billing_pricing_components (plan_id, name, type, unit_price, frequency)
    VALUES (growth_id, 'Base Platform Fee', 'FLAT', 199.00, 'MONTHLY');
    INSERT INTO billing_pricing_components (plan_id, name, type, unit_price, frequency)
    VALUES (growth_id, 'Vehicle Fee', 'PER_UNIT', 15.00, 'MONTHLY');
    INSERT INTO billing_pricing_components (plan_id, name, type, unit_price, meter_event_name, frequency)
    VALUES (growth_id, 'Telemetry Ingestion', 'USAGE', 0.01, 'telemetry_event', 'MONTHLY');

    -- Enterprise Components
    INSERT INTO billing_pricing_components (plan_id, name, type, unit_price, frequency)
    VALUES (enterprise_id, 'Custom Platform Fee', 'FLAT', 999.00, 'MONTHLY');
END $$;
