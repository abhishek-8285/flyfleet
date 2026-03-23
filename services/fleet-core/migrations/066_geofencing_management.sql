-- Migration: 066_geofencing_management
-- Description: Creates tables for geofences, geofence events, and vehicle assignments

-- Enable PostGIS extension for geospatial queries
CREATE EXTENSION IF NOT EXISTS postgis;

-- 1. Create geofences table
CREATE TABLE IF NOT EXISTS geofences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('circle', 'polygon', 'route_corridor')),
    
    -- For circle geofences
    center GEOGRAPHY(POINT, 4326),
    radius_meters INTEGER CHECK (radius_meters >= 10 AND radius_meters <= 50000),
    
    -- For polygon geofences
    coordinates GEOGRAPHY(POLYGON, 4326),
    
    -- For route corridors
    route_polyline TEXT,
    buffer_meters INTEGER CHECK (buffer_meters >= 10 AND buffer_meters <= 1000),
    
    -- Alert settings
    entry_alert BOOLEAN DEFAULT FALSE,
    exit_alert BOOLEAN DEFAULT FALSE,
    speed_limit INTEGER CHECK (speed_limit >= 10 AND speed_limit <= 200),
    
    -- Target vehicles (empty = all vehicles)
    vehicles UUID[] DEFAULT '{}',
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    active_hours JSONB, -- {start: "08:00", end: "18:00", days: [1,2,3,4,5]}
    
    -- Status
    enabled BOOLEAN DEFAULT TRUE,
    
    -- Audit
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 2. Create geofence_assignments table (for vehicle-specific geofences)
CREATE TABLE IF NOT EXISTS geofence_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    geofence_id UUID NOT NULL REFERENCES geofences(id) ON DELETE CASCADE,
    vehicle_id UUID NOT NULL,
    assignment_type VARCHAR(20) NOT NULL CHECK (assignment_type IN ('enter', 'exit', 'dwell', 'both')),
    dwell_minutes INTEGER, -- For dwell alerts
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(geofence_id, vehicle_id)
);

-- 3. Create geofence_events table (TimescaleDB hypertable for time-series data)
CREATE TABLE IF NOT EXISTS geofence_events (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    time TIMESTAMPTZ NOT NULL,
    vehicle_id UUID NOT NULL,
    geofence_id UUID NOT NULL REFERENCES geofences(id),
    event_type VARCHAR(20) NOT NULL CHECK (event_type IN ('enter', 'exit', 'dwell', 'speed')),
    location GEOGRAPHY(POINT, 4326),
    duration_minutes INTEGER, -- For exit/dwell events
    speed_kmh DECIMAL(5, 2), -- For speed events
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Create hypertable for time-series data (requires TimescaleDB)
SELECT create_hypertable('geofence_events', 'time', if_not_exists => TRUE);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_geofences_org ON geofences(org_id);
CREATE INDEX IF NOT EXISTS idx_geofences_type ON geofences(type);
CREATE INDEX IF NOT EXISTS idx_geofences_enabled ON geofences(enabled);
CREATE INDEX IF NOT EXISTS idx_geofences_center ON geofences USING GIST (center);
CREATE INDEX IF NOT EXISTS idx_geofences_coordinates ON geofences USING GIST (coordinates);

CREATE INDEX IF NOT EXISTS idx_geofence_assignments_geofence ON geofence_assignments(geofence_id);
CREATE INDEX IF NOT EXISTS idx_geofence_assignments_vehicle ON geofence_assignments(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_geofence_assignments_active ON geofence_assignments(is_active);

-- TimescaleDB indexes
CREATE INDEX IF NOT EXISTS idx_geofence_events_vehicle_time ON geofence_events(vehicle_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_geofence_events_geofence_time ON geofence_events(geofence_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_geofence_events_type_time ON geofence_events(event_type, time DESC);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_geofencing_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_geofences_modtime
BEFORE UPDATE ON geofences
FOR EACH ROW
EXECUTE FUNCTION update_geofencing_timestamp();

CREATE TRIGGER update_geofence_assignments_modtime
BEFORE UPDATE ON geofence_assignments
FOR EACH ROW
EXECUTE FUNCTION update_geofencing_timestamp();

-- Comments
COMMENT ON TABLE geofences IS 'Geographic boundaries for vehicle monitoring';
COMMENT ON TABLE geofence_assignments IS 'Vehicle-specific geofence assignments';
COMMENT ON TABLE geofence_events IS 'Geofence entry/exit/dwell events (time-series)';

COMMENT ON COLUMN geofences.type IS 'circle=point+radius, polygon=polygon, route_corridor=polyline+buffer';
COMMENT ON COLUMN geofences.center IS 'Center point for circle geofences (longitude, latitude)';
COMMENT ON COLUMN geofences.coordinates IS 'Polygon coordinates for polygon geofences';
COMMENT ON COLUMN geofences.route_polyline IS 'Encoded polyline for route corridors';
COMMENT ON COLUMN geofence_assignments.assignment_type IS 'enter=alert on entry, exit=alert on exit, dwell=alert after dwell time, both=enter+exit';

-- View for active geofences with vehicle count
CREATE OR REPLACE VIEW geofence_stats AS
SELECT
    g.id,
    g.org_id,
    g.name,
    g.type,
    g.enabled,
    COUNT(DISTINCT ga.vehicle_id) FILTER (WHERE ga.is_active = true) as assigned_vehicles,
    COUNT(ge.id) FILTER (WHERE ge.time >= NOW() - INTERVAL '24 hours') as events_24h,
    COUNT(ge.id) FILTER (WHERE ge.event_type = 'enter' AND ge.time >= NOW() - INTERVAL '24 hours') as entries_24h,
    COUNT(ge.id) FILTER (WHERE ge.event_type = 'exit' AND ge.time >= NOW() - INTERVAL '24 hours') as exits_24h
FROM geofences g
LEFT JOIN geofence_assignments ga ON ga.geofence_id = g.id
LEFT JOIN geofence_events ge ON ge.geofence_id = g.id
GROUP BY g.id, g.org_id, g.name, g.type, g.enabled;

-- View for recent geofence events
CREATE OR REPLACE VIEW recent_geofence_events AS
SELECT
    ge.time,
    ge.vehicle_id,
    v.registration_number as vehicle_reg_number,
    ge.geofence_id,
    g.name as geofence_name,
    ge.event_type,
    ST_X(ge.location::geometry) as longitude,
    ST_Y(ge.location::geometry) as latitude,
    ge.duration_minutes,
    ge.speed_kmh,
    ge.metadata
FROM geofence_events ge
JOIN geofences g ON ge.geofence_id = g.id
LEFT JOIN vehicles v ON ge.vehicle_id = v.id
ORDER BY ge.time DESC
LIMIT 1000;

-- Function to check if point is inside geofence
CREATE OR REPLACE FUNCTION is_point_inside_geofence(
    p_longitude DOUBLE PRECISION,
    p_latitude DOUBLE PRECISION,
    p_geofence_id UUID
)
RETURNS BOOLEAN AS $$
DECLARE
    v_inside BOOLEAN;
    v_geofence_type VARCHAR;
    v_center GEOGRAPHY;
    v_radius INTEGER;
    v_polygon GEOGRAPHY;
    v_point GEOGRAPHY;
BEGIN
    v_point := ST_SetSRID(ST_MakePoint(p_longitude, p_latitude), 4326)::geography;
    
    -- Get geofence details
    SELECT type, center, radius_meters, coordinates
    INTO v_geofence_type, v_center, v_radius, v_polygon
    FROM geofences
    WHERE id = p_geofence_id AND enabled = true;
    
    IF NOT FOUND THEN
        RETURN FALSE;
    END IF;
    
    -- Check based on geofence type
    IF v_geofence_type = 'circle' THEN
        SELECT ST_DWithin(v_point, v_center, v_radius) INTO v_inside;
    ELSIF v_geofence_type = 'polygon' THEN
        SELECT ST_Within(v_point, v_polygon) INTO v_inside;
    ELSE
        RETURN FALSE;
    END IF;
    
    RETURN v_inside;
END;
$$ LANGUAGE plpgsql;

-- Function to get geofences for a vehicle
CREATE OR REPLACE FUNCTION get_geofences_for_vehicle(p_vehicle_id UUID)
RETURNS TABLE (
    geofence_id UUID,
    name VARCHAR,
    type VARCHAR,
    assignment_type VARCHAR,
    dwell_minutes INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        g.id,
        g.name,
        g.type,
        ga.assignment_type,
        ga.dwell_minutes
    FROM geofences g
    LEFT JOIN geofence_assignments ga ON ga.geofence_id = g.id AND ga.is_active = true
    WHERE g.enabled = true
      AND (ga.vehicle_id = p_vehicle_id OR array_length(g.vehicles, 1) IS NULL OR g.vehicles @> ARRAY[p_vehicle_id]);
END;
$$ LANGUAGE plpgsql;
