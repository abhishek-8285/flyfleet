-- Migration: 068_route_optimization
-- Description: Creates tables for route optimization requests, results, and jobs

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";

-- 1. Create route_optimization_requests table
CREATE TABLE IF NOT EXISTS route_optimization_requests (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    start_time_window TIMESTAMPTZ,
    end_time_window TIMESTAMPTZ,
    constraints JSONB NOT NULL DEFAULT '{}',
    stops JSONB NOT NULL DEFAULT '[]',
    vehicles JSONB NOT NULL DEFAULT '[]',
    priorities JSONB NOT NULL DEFAULT '{}',
    optimization_goals JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    result JSONB,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 2. Create route_optimization_results table
CREATE TABLE IF NOT EXISTS route_optimization_results (
    id UUID PRIMARY KEY,
    request_id UUID NOT NULL REFERENCES route_optimization_requests(id) ON DELETE CASCADE,
    optimization_time TIMESTAMPTZ NOT NULL,
    total_distance DOUBLE PRECISION NOT NULL,
    total_time INTEGER NOT NULL,
    total_cost DOUBLE PRECISION NOT NULL,
    total_vehicles_used INTEGER NOT NULL,
    total_stops INTEGER NOT NULL,
    routes JSONB NOT NULL DEFAULT '[]',
    performance_metrics JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 3. Create route_optimization_jobs table
CREATE TABLE IF NOT EXISTS route_optimization_jobs (
    id UUID PRIMARY KEY,
    request_id UUID NOT NULL REFERENCES route_optimization_requests(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error TEXT,
    result_id UUID REFERENCES route_optimization_results(id),
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_route_requests_org ON route_optimization_requests(organization_id);
CREATE INDEX IF NOT EXISTS idx_route_requests_status ON route_optimization_requests(status);
CREATE INDEX IF NOT EXISTS idx_route_requests_created ON route_optimization_requests(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_route_results_request ON route_optimization_results(request_id);
CREATE INDEX IF NOT EXISTS idx_route_jobs_request ON route_optimization_jobs(request_id);
CREATE INDEX IF NOT EXISTS idx_route_jobs_status ON route_optimization_jobs(status);

-- Comments
COMMENT ON TABLE route_optimization_requests IS 'Route optimization requests with constraints and goals';
COMMENT ON TABLE route_optimization_results IS 'Optimized route results with performance metrics';
COMMENT ON TABLE route_optimization_jobs IS 'Background job tracking for route optimization';

COMMENT ON COLUMN route_optimization_requests.constraints IS 'Optimization constraints (max travel time, distance, etc.)';
COMMENT ON COLUMN route_optimization_requests.stops IS 'Array of stops with locations and time windows';
COMMENT ON COLUMN route_optimization_requests.vehicles IS 'Array of vehicle profiles with capabilities';
COMMENT ON COLUMN route_optimization_requests.priorities IS 'Priority weights for optimization objectives';
COMMENT ON COLUMN route_optimization_requests.optimization_goals IS 'Boolean flags for optimization goals';
COMMENT ON COLUMN route_optimization_results.routes IS 'Optimized routes for each vehicle';
COMMENT ON COLUMN route_optimization_results.performance_metrics IS 'Performance metrics (improvement, execution time, etc.)';

-- View for request statistics
CREATE OR REPLACE VIEW route_optimization_stats AS
SELECT
    organization_id,
    COUNT(*) as total_requests,
    COUNT(*) FILTER (WHERE status = 'completed') as completed_requests,
    COUNT(*) FILTER (WHERE status = 'failed') as failed_requests,
    COUNT(*) FILTER (WHERE status = 'processing') as processing_requests,
    AVG(EXTRACT(EPOCH FROM (updated_at - created_at))) FILTER (WHERE status = 'completed') as avg_processing_time_seconds,
    MAX(created_at) as last_request_at
FROM route_optimization_requests
GROUP BY organization_id;

-- View for job statistics
CREATE OR REPLACE VIEW route_optimization_job_stats AS
SELECT
    request_id,
    COUNT(*) as total_jobs,
    COUNT(*) FILTER (WHERE status = 'completed') as completed_jobs,
    COUNT(*) FILTER (WHERE status = 'failed') as failed_jobs,
    MAX(created_at) as last_job_at,
    MAX(completed_at) as last_completed_at
FROM route_optimization_jobs
GROUP BY request_id;

-- Function to cleanup old completed requests (keep last 90 days)
CREATE OR REPLACE FUNCTION cleanup_old_route_requests()
RETURNS void AS $$
BEGIN
    DELETE FROM route_optimization_requests
    WHERE status = 'completed'
      AND created_at < NOW() - INTERVAL '90 days';
END;
$$ LANGUAGE plpgsql;

-- Function to cleanup old jobs (keep last 30 days)
CREATE OR REPLACE FUNCTION cleanup_old_route_jobs()
RETURNS void AS $$
BEGIN
    DELETE FROM route_optimization_jobs
    WHERE completed_at < NOW() - INTERVAL '30 days';
END;
$$ LANGUAGE plpgsql;
