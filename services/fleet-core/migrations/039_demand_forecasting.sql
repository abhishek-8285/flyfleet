-- Migration: 039_demand_forecasting
-- Description: Creates tables for demand forecasting, pricing rules, and experiments.

-- 1. Create historical_demand table
CREATE TABLE IF NOT EXISTS historical_demand (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    route_id TEXT NOT NULL,
    date DATE NOT NULL,
    hour INT NOT NULL CHECK (hour >= 0 AND hour <= 23),
    demand_count INT NOT NULL DEFAULT 0,
    avg_price DECIMAL(10, 2),
    weather_condition TEXT,
    is_holiday BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, route_id, date, hour)
);

CREATE INDEX IF NOT EXISTS idx_historical_demand_tenant_route ON historical_demand(tenant_id, route_id);

-- 2. Create model_predictions table
CREATE TABLE IF NOT EXISTS model_predictions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    model_version TEXT NOT NULL,
    route_id TEXT NOT NULL,
    forecast_date DATE NOT NULL,
    forecast_hour INT NOT NULL,
    predicted_demand DECIMAL(10, 2) NOT NULL,
    confidence_interval_lower DECIMAL(10, 2),
    confidence_interval_upper DECIMAL(10, 2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_model_predictions_tenant ON model_predictions(tenant_id);

-- 3. Create pricing_rules table
CREATE TABLE IF NOT EXISTS pricing_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name TEXT NOT NULL,
    description TEXT,
    condition_json JSONB NOT NULL,
    action_json JSONB NOT NULL,
    priority INT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pricing_rules_tenant ON pricing_rules(tenant_id);

-- 4. Create ab_test_experiments table
CREATE TABLE IF NOT EXISTS ab_test_experiments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name TEXT NOT NULL,
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    status TEXT DEFAULT 'DRAFT',
    variants JSONB NOT NULL,
    metrics JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ab_test_experiments_tenant ON ab_test_experiments(tenant_id);
