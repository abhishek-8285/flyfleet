-- Migration to add route optimization tables
-- This adds route optimization functionality to the fleet platform

-- Table for storing route optimization requests
CREATE TABLE IF NOT EXISTS route_optimization_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    start_time_window TIMESTAMP WITH TIME ZONE,
    end_time_window TIMESTAMP WITH TIME ZONE,
    constraints JSONB NOT NULL,
    stops JSONB NOT NULL,
    vehicles JSONB NOT NULL,
    priorities JSONB NOT NULL,
    optimization_goals JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    result JSONB,
    CONSTRAINT fk_route_optimization_requests_org FOREIGN KEY (organization_id) REFERENCES organizations(id)
);

-- Table for storing route optimization results
CREATE TABLE IF NOT EXISTS route_optimization_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL,
    optimization_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    total_distance DECIMAL(10,2) NOT NULL,
    total_time INTEGER NOT NULL, -- in minutes
    total_cost DECIMAL(12,2) NOT NULL,
    total_vehicles_used INTEGER NOT NULL,
    total_stops INTEGER NOT NULL,
    routes JSONB NOT NULL,
    performance_metrics JSONB NOT NULL,
    CONSTRAINT fk_route_optimization_results_request FOREIGN KEY (request_id) REFERENCES route_optimization_requests(id)
);

-- Table for storing optimization jobs
CREATE TABLE IF NOT EXISTS optimization_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error TEXT,
    result_id UUID,
    CONSTRAINT fk_optimization_jobs_request FOREIGN KEY (request_id) REFERENCES route_optimization_requests(id),
    CONSTRAINT fk_optimization_jobs_result FOREIGN KEY (result_id) REFERENCES route_optimization_results(id)
);

-- Individual indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_route_optimization_requests_org ON route_optimization_requests (organization_id);
CREATE INDEX IF NOT EXISTS idx_route_optimization_requests_status ON route_optimization_requests (status);
CREATE INDEX IF NOT EXISTS idx_route_optimization_requests_created_at ON route_optimization_requests (created_at);
CREATE INDEX IF NOT EXISTS idx_route_optimization_results_request ON route_optimization_results (request_id);
CREATE INDEX IF NOT EXISTS idx_optimization_jobs_request ON optimization_jobs (request_id);
CREATE INDEX IF NOT EXISTS idx_optimization_jobs_status ON optimization_jobs (status);
CREATE INDEX IF NOT EXISTS idx_optimization_jobs_created_at ON optimization_jobs (created_at);

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_route_optimization_requests_org_status ON route_optimization_requests (organization_id, status);
CREATE INDEX IF NOT EXISTS idx_optimization_jobs_request_status ON optimization_jobs (request_id, status);

-- Add comments to describe the tables
COMMENT ON TABLE route_optimization_requests IS 'Stores route optimization requests with all parameters and constraints';
COMMENT ON COLUMN route_optimization_requests.status IS 'Status of the optimization request (pending, processing, completed, failed)';
COMMENT ON COLUMN route_optimization_requests.result IS 'Stores the result of the optimization if completed';

COMMENT ON TABLE route_optimization_results IS 'Stores the results of route optimization computations';
COMMENT ON COLUMN route_optimization_results.total_distance IS 'Total distance of all routes in kilometers';
COMMENT ON COLUMN route_optimization_results.total_time IS 'Total time of all routes in minutes';
COMMENT ON COLUMN route_optimization_results.total_cost IS 'Total cost of all routes';

COMMENT ON TABLE optimization_jobs IS 'Tracks background jobs for route optimization processing';
COMMENT ON COLUMN optimization_jobs.status IS 'Status of the optimization job (pending, running, completed, failed)';