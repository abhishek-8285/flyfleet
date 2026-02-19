CREATE TABLE IF NOT EXISTS historical_demand (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    route_id TEXT NOT NULL, -- Logical route identifier (e.g., "BLR-mysore")
    date DATE NOT NULL,
    hour INT NOT NULL CHECK (hour >= 0 AND hour <= 23),
    demand_count INT NOT NULL DEFAULT 0, -- Number of bookings/searches
    avg_price DECIMAL(10, 2),
    weather_condition TEXT, -- Optional: "Rainy", "Sunny"
    is_holiday BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(route_id, date, hour)
);

CREATE TABLE IF NOT EXISTS model_predictions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    model_version TEXT NOT NULL,
    route_id TEXT NOT NULL,
    forecast_date DATE NOT NULL,
    forecast_hour INT NOT NULL,
    predicted_demand DECIMAL(10, 2) NOT NULL,
    confidence_interval_lower DECIMAL(10, 2),
    confidence_interval_upper DECIMAL(10, 2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pricing_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    description TEXT,
    condition_json JSONB NOT NULL, -- e.g., {"min_demand": 50, "max_utilization": 0.8}
    action_json JSONB NOT NULL,    -- e.g., {"multiplier": 1.1, "type": "SURGE"}
    priority INT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ab_test_experiments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    status TEXT DEFAULT 'DRAFT', -- DRAFT, RUNNING, COMPLETED
    variants JSONB NOT NULL, -- e.g., ["A", "B", "Control"]
    metrics JSONB DEFAULT '{}'::jsonb, -- Store results like conversion_rate, revenue_lift
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
