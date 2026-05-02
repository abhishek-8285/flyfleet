-- Create billing system tables for enterprise SaaS
CREATE TABLE IF NOT EXISTS billing_plans (
    id UUID PRIMARY KEY
);

DO $$ 
BEGIN 
    ALTER TABLE billing_plans ADD COLUMN IF NOT EXISTS name VARCHAR(255) NOT NULL DEFAULT '';
    ALTER TABLE billing_plans ADD COLUMN IF NOT EXISTS description TEXT;
    ALTER TABLE billing_plans ADD COLUMN IF NOT EXISTS base_price_monthly DECIMAL(12, 2) NOT NULL DEFAULT 0;
    ALTER TABLE billing_plans ADD COLUMN IF NOT EXISTS base_price_yearly DECIMAL(12, 2) NOT NULL DEFAULT 0;
    ALTER TABLE billing_plans ADD COLUMN IF NOT EXISTS features JSONB DEFAULT '{}';
    ALTER TABLE billing_plans ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT TRUE;
    ALTER TABLE billing_plans ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();
    ALTER TABLE billing_plans ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS billing_subscriptions (
    id UUID PRIMARY KEY
);

DO $$ 
BEGIN 
    ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL REFERENCES tenants(id);
    ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS plan_id UUID REFERENCES billing_plans(id);
    ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'active';
    ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS billing_cycle VARCHAR(20) NOT NULL DEFAULT 'monthly';
    ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS current_period_start TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
    ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS current_period_end TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW() + INTERVAL '30 days';
    ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS cancel_at_period_end BOOLEAN DEFAULT FALSE;
    ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();
    ALTER TABLE billing_subscriptions ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();
    
    -- Add unique constraint if missing
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_name = 'billing_subscriptions' AND constraint_name = 'uq_billing_subscriptions_tenant') THEN
        ALTER TABLE billing_subscriptions ADD CONSTRAINT uq_billing_subscriptions_tenant UNIQUE (tenant_id);
    END IF;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS billing_usage_records (
    id UUID PRIMARY KEY
);

DO $$ 
BEGIN 
    ALTER TABLE billing_usage_records ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL REFERENCES tenants(id);
    ALTER TABLE billing_usage_records ADD COLUMN IF NOT EXISTS usage_type VARCHAR(100) NOT NULL DEFAULT '';
    ALTER TABLE billing_usage_records ADD COLUMN IF NOT EXISTS quantity INTEGER NOT NULL DEFAULT 0;
    ALTER TABLE billing_usage_records ADD COLUMN IF NOT EXISTS recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS billing_invoices (
    id UUID PRIMARY KEY
);

DO $$ 
BEGIN 
    ALTER TABLE billing_invoices ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL REFERENCES tenants(id);
    ALTER TABLE billing_invoices ADD COLUMN IF NOT EXISTS subscription_id UUID REFERENCES billing_subscriptions(id);
    ALTER TABLE billing_invoices ADD COLUMN IF NOT EXISTS amount DECIMAL(12, 2) NOT NULL DEFAULT 0;
    ALTER TABLE billing_invoices ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'draft';
    ALTER TABLE billing_invoices ADD COLUMN IF NOT EXISTS due_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW() + INTERVAL '7 days';
    ALTER TABLE billing_invoices ADD COLUMN IF NOT EXISTS paid_at TIMESTAMP WITH TIME ZONE;
    ALTER TABLE billing_invoices ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();
    ALTER TABLE billing_invoices ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_billing_usage_tenant ON billing_usage_records(tenant_id, usage_type, recorded_at);
CREATE INDEX IF NOT EXISTS idx_billing_invoices_tenant ON billing_invoices(tenant_id, status);

-- Insert demo plans
INSERT INTO billing_plans (id, name, description, base_price_monthly, base_price_yearly, features) VALUES
('b3b3b3b3-b3b3-b3b3-b3b3-b3b3b3b3b3b1', 'Startup', 'Perfect for small fleets up to 10 vehicles', 49.00, 490.00, '{"max_vehicles": 10, "tracking": true, "basic_reports": true}'),
('b3b3b3b3-b3b3-b3b3-b3b3-b3b3b3b3b3b2', 'Growth', 'Scale your fleet with advanced analytics', 199.00, 1990.00, '{"max_vehicles": 50, "tracking": true, "advanced_reports": true, "api_access": true}'),
('b3b3b3b3-b3b3-b3b3-b3b3-b3b3b3b3b3b3', 'Enterprise', 'Complete control and premium support', 999.00, 9990.00, '{"max_vehicles": -1, "tracking": true, "custom_reports": true, "api_access": true, "rbac": true}')
ON CONFLICT (id) DO NOTHING;
