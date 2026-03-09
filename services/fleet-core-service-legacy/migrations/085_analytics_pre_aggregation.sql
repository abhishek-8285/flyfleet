-- Migration: 085_analytics_pre_aggregation.sql
-- Description: Create pre-aggregated tables for high-performance analytics queries.

CREATE TABLE IF NOT EXISTS vehicle_daily_stats (
    vehicle_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    date DATE NOT NULL,
    total_distance NUMERIC(15, 2) DEFAULT 0,
    total_fuel_consumed NUMERIC(15, 2) DEFAULT 0,
    total_engine_hours NUMERIC(15, 2) DEFAULT 0,
    total_idle_hours NUMERIC(15, 2) DEFAULT 0,
    max_speed NUMERIC(10, 2) DEFAULT 0,
    avg_speed NUMERIC(10, 2) DEFAULT 0,
    harsh_braking_count INT DEFAULT 0,
    rapid_acceleration_count INT DEFAULT 0,
    sharp_turn_count INT DEFAULT 0,
    overspeeding_count INT DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (vehicle_id, date)
);

CREATE INDEX IF NOT EXISTS idx_vehicle_daily_stats_org_date ON vehicle_daily_stats (organization_id, date);

CREATE TABLE IF NOT EXISTS driver_daily_stats (
    driver_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    date DATE NOT NULL,
    total_driving_hours NUMERIC(15, 2) DEFAULT 0,
    total_distance NUMERIC(15, 2) DEFAULT 0,
    total_idle_hours NUMERIC(15, 2) DEFAULT 0,
    safety_score NUMERIC(5, 2) DEFAULT 0,
    eco_score NUMERIC(5, 2) DEFAULT 0,
    violation_count INT DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (driver_id, date)
);

CREATE INDEX IF NOT EXISTS idx_driver_daily_stats_org_date ON driver_daily_stats (organization_id, date);

CREATE TABLE IF NOT EXISTS fleet_hourly_stats (
    organization_id UUID NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    active_vehicles INT DEFAULT 0,
    total_distance NUMERIC(15, 2) DEFAULT 0,
    total_fuel_consumed NUMERIC(15, 2) DEFAULT 0,
    average_utilization NUMERIC(5, 2) DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (organization_id, timestamp)
);

CREATE INDEX IF NOT EXISTS idx_fleet_hourly_stats_timestamp ON fleet_hourly_stats (timestamp);

-- Comments
COMMENT ON TABLE vehicle_daily_stats IS 'Pre-aggregated daily metrics per vehicle';
COMMENT ON TABLE driver_daily_stats IS 'Pre-aggregated daily metrics per driver';
COMMENT ON TABLE fleet_hourly_stats IS 'Pre-aggregated hourly metrics for the entire fleet';
