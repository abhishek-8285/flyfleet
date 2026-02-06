-- Enable UUID extension if not enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

---------------------------------------------------
-- TRIPS & LOGISTICS
---------------------------------------------------

CREATE TABLE IF NOT EXISTS trips (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    vehicle_id UUID REFERENCES vehicles(id),
    driver_id UUID REFERENCES drivers(id),
    
    status VARCHAR(50) NOT NULL DEFAULT 'planned', -- planned, in_progress, completed, cancelled
    title VARCHAR(255),
    description TEXT,
    
    -- Route & Location Data (JSONB for flexibility)
    planned_route JSONB,  -- Array of Waypoints
    actual_route JSONB,   -- Array of Waypoints/Breadcrumbs
    waypoints JSONB,      -- Array of defined stops
    
    start_location JSONB, -- {lat, lng, address}
    end_location JSONB,   -- {lat, lng, address}
    
    -- Metrics
    estimated_distance_km DECIMAL(10, 2),
    actual_distance_km DECIMAL(10, 2),
    estimated_duration_min INTEGER,
    actual_duration_min INTEGER,
    
    fuel_consumed_liters DECIMAL(10, 2) DEFAULT 0,
    idling_time_minutes INTEGER DEFAULT 0,
    
    -- Logistics
    cargo_details JSONB, -- {type, weight, volume, ...}
    pod_documents JSONB, -- Proof of Delivery [{url, type, signature}]
    
    -- Scheduling
    scheduled_start TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    
    -- Customer Info
    customer_name VARCHAR(255),
    customer_phone VARCHAR(50),
    customer_email VARCHAR(255),
    
    cancellation_reason TEXT,
    events JSONB, -- Array of TripEvents [Breakdown, Delay, etc]

    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_trips_org_id ON trips(org_id);
CREATE INDEX idx_trips_vehicle_id ON trips(vehicle_id);
CREATE INDEX idx_trips_driver_id ON trips(driver_id);
CREATE INDEX idx_trips_status ON trips(status);

---------------------------------------------------
-- ALERTS & MONITORING
---------------------------------------------------

-- Alert Rules (Configuration)
CREATE TABLE IF NOT EXISTS alert_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL, -- overspeed, geofence, harsh_braking
    description TEXT,
    severity VARCHAR(20) DEFAULT 'warning', -- info, warning, critical
    
    conditions JSONB NOT NULL, -- {threshold: 80, duration: 5}
    actions JSONB,             -- [{type: "email", target: "admin@fleet.com"}]
    
    -- Target Scope (If null, applies to all)
    vehicles JSONB,   -- Array of Vehicle UUIDs
    driver_ids JSONB,  -- Array of Driver UUIDs
    
    enabled BOOLEAN DEFAULT TRUE,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Geofences
CREATE TABLE IF NOT EXISTS geofences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL, -- circle, polygon, route
    
    -- Geometry Definition
    center JSONB,        -- {lat, lng} for Circle
    radius_meters INTEGER, -- for Circle
    coordinates JSONB,   -- Array of {lat, lng} for Polygon
    route_polyline TEXT, -- Encoded polyline for Route
    buffer_meters INTEGER, -- Corridor width
    
    -- Triggers
    entry_alert BOOLEAN DEFAULT TRUE,
    exit_alert BOOLEAN DEFAULT TRUE,
    speed_limit INTEGER, -- Optional speed limit inside fence
    
    vehicles JSONB, -- Array of Vehicle UUIDs restricted to this fence
    
    enabled BOOLEAN DEFAULT TRUE,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Active Alerts (Incidents)
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    rule_id UUID REFERENCES alert_rules(id), -- Optional link to rule
    
    type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) DEFAULT 'warning',
    
    vehicle_id UUID REFERENCES vehicles(id),
    driver_id UUID REFERENCES drivers(id),
    
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    location JSONB, -- {lat, lng, address}
    
    details JSONB, -- {speed: 120, limit: 80}
    
    status VARCHAR(20) DEFAULT 'new', -- new, acknowledged, resolved, escalated
    notes TEXT,
    
    acknowledged_by UUID REFERENCES users(id),
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    
    resolved_by UUID REFERENCES users(id),
    resolved_at TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_alerts_org_id ON alerts(org_id);
CREATE INDEX idx_alerts_vehicle_id ON alerts(vehicle_id);
CREATE INDEX idx_alerts_status ON alerts(status);
CREATE INDEX idx_alerts_timestamp ON alerts(timestamp);
