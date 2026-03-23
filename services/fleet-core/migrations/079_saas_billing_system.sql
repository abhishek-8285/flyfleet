-- Create billing system tables for enterprise SaaS
CREATE TABLE billing_plans (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    base_price_monthly DECIMAL(12, 2) NOT NULL,
    base_price_yearly DECIMAL(12, 2) NOT NULL,
    features JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE billing_subscriptions (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL,
    plan_id UUID REFERENCES billing_plans(id),
    status VARCHAR(50) NOT NULL, -- active, trialing, past_due, canceled
    billing_cycle VARCHAR(20) NOT NULL, -- monthly, yearly
    current_period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    current_period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(organization_id)
);

CREATE TABLE billing_usage_records (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL,
    usage_type VARCHAR(100) NOT NULL, -- active_vehicles, tracking_pings, api_calls
    quantity INTEGER NOT NULL DEFAULT 0,
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE billing_invoices (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL,
    subscription_id UUID REFERENCES billing_subscriptions(id),
    amount DECIMAL(12, 2) NOT NULL,
    status VARCHAR(50) NOT NULL, -- draft, open, paid, void, uncollectible
    due_date TIMESTAMP WITH TIME ZONE NOT NULL,
    paid_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_billing_usage_org ON billing_usage_records(organization_id, usage_type, recorded_at);
CREATE INDEX idx_billing_invoices_org ON billing_invoices(organization_id, status);

-- Insert demo plans
INSERT INTO billing_plans (id, name, description, base_price_monthly, base_price_yearly, features) VALUES
('b3b3b3b3-b3b3-b3b3-b3b3-b3b3b3b3b3b1', 'Startup', 'Perfect for small fleets up to 10 vehicles', 49.00, 490.00, '{"max_vehicles": 10, "tracking": true, "basic_reports": true}'),
('b3b3b3b3-b3b3-b3b3-b3b3-b3b3b3b3b3b2', 'Growth', 'Scale your fleet with advanced analytics', 199.00, 1990.00, '{"max_vehicles": 50, "tracking": true, "advanced_reports": true, "api_access": true}'),
('b3b3b3b3-b3b3-b3b3-b3b3-b3b3b3b3b3b3', 'Enterprise', 'Complete control and premium support', 999.00, 9990.00, '{"max_vehicles": -1, "tracking": true, "custom_reports": true, "api_access": true, "rbac": true}');
