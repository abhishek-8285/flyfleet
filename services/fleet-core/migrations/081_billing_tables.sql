-- Migration: 081_billing_tables
-- Description: Billing and Subscription Tables (Harmonized with Go repository).

-- 1. Billing Plans (Static IDs for consistency)
CREATE TABLE IF NOT EXISTS billing_plans (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Pricing Components
CREATE TABLE IF NOT EXISTS billing_pricing_components (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES billing_plans(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL, -- FLAT, PER_UNIT, USAGE
    unit_price DECIMAL(12, 4) NOT NULL DEFAULT 0,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    meter_event_name VARCHAR(100),
    frequency VARCHAR(50) NOT NULL DEFAULT 'MONTHLY',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 3. Billing Subscriptions
CREATE TABLE IF NOT EXISTS billing_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL REFERENCES tenants(id);
    ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS plan_id UUID NOT NULL REFERENCES billing_plans(id);
    ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'TRIAL';
    ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS current_period_start TIMESTAMPTZ NOT NULL DEFAULT NOW();
    ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS current_period_end TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '30 days';
    ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS trial_end TIMESTAMPTZ;
    ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE;
    ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}';
    ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
    ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
    
    -- Ensure unique constraint exists
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_name = 'billing_subscriptions' AND constraint_name = 'uq_billing_subscriptions_tenant') THEN
        ALTER TABLE billing_subscriptions ADD CONSTRAINT uq_billing_subscriptions_tenant UNIQUE (tenant_id);
    END IF;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_billing_subs_tenant ON billing_subscriptions(tenant_id);

-- 4. Seed Initial Plans
INSERT INTO billing_plans (id, name, description) VALUES
    ('11111111-1111-1111-1111-111111111111', 'Starter', 'Basic fleet tracking for up to 5 vehicles'),
    ('22222222-2222-2222-2222-222222222222', 'Growth', 'Advanced features for growing fleets'),
    ('33333333-3333-3333-3333-333333333333', 'Enterprise', 'Custom scale solutions with premium support')
ON CONFLICT (id) DO NOTHING;
