-- Migration: 085_analytics_pre_aggregation
-- Description: Create pre-aggregated tables for high-performance analytics with tenant isolation.

-- 1. Vehicle Daily Stats
CREATE TABLE IF NOT EXISTS vehicle_daily_stats (
    vehicle_id UUID NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    date DATE NOT NULL,
    total_distance DECIMAL(15, 2) DEFAULT 0,
    total_fuel_consumed DECIMAL(15, 2) DEFAULT 0,
    total_engine_hours DECIMAL(15, 2) DEFAULT 0,
    harsh_braking_count INT DEFAULT 0,
    harsh_acceleration_count INT DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (vehicle_id, date)
);

CREATE INDEX IF NOT EXISTS idx_vehicle_daily_tenant ON vehicle_daily_stats (tenant_id, date);

-- 2. Driver Daily Stats
CREATE TABLE IF NOT EXISTS driver_daily_stats (
    driver_id UUID NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    date DATE NOT NULL,
    total_driving_hours DECIMAL(15, 2) DEFAULT 0,
    total_distance DECIMAL(15, 2) DEFAULT 0,
    safety_score DECIMAL(5, 2) DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (driver_id, date)
);

CREATE INDEX IF NOT EXISTS idx_driver_daily_tenant ON driver_daily_stats (tenant_id, date);

-- 3. Fleet Hourly Stats
CREATE TABLE IF NOT EXISTS fleet_hourly_stats (
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    active_vehicles INT DEFAULT 0,
    total_distance DECIMAL(15, 2) DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tenant_id, timestamp)
);
