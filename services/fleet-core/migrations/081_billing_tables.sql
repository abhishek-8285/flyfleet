-- 081_billing_tables.sql
-- Billing and Subscription Tables

CREATE TABLE IF NOT EXISTS plans (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    monthly_price DECIMAL(12, 2) NOT NULL,
    max_vehicles INTEGER NOT NULL,
    max_drivers INTEGER NOT NULL,
    features JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL,
    plan_id UUID REFERENCES plans(id),
    status TEXT NOT NULL,
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS usage_records (
    id BIGSERIAL PRIMARY KEY,
    organization_id UUID NOT NULL,
    metric TEXT NOT NULL,
    value INTEGER NOT NULL,
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL,
    amount DECIMAL(12, 2) NOT NULL,
    currency TEXT DEFAULT 'USD',
    status TEXT NOT NULL,
    issued_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    paid_at TIMESTAMP WITH TIME ZONE
);

-- Seed Initial Plans
INSERT INTO plans (id, name, monthly_price, max_vehicles, max_drivers, features) VALUES
    (gen_random_uuid(), 'Starter', 49.00, 10, 5, '{"api_access": true, "reports": false}'),
    (gen_random_uuid(), 'Professional', 149.00, 50, 25, '{"api_access": true, "reports": true, "geo_fencing": true}'),
    (gen_random_uuid(), 'Enterprise', 499.00, 1000, 500, '{"api_access": true, "reports": true, "geo_fencing": true, "dedicated_support": true, "custom_integrations": true}')
ON CONFLICT DO NOTHING;
