-- Migration: 028_route_optimization_schema.sql
-- Description: Schema for Route Optimization module (Missing in original migration)

CREATE TABLE IF NOT EXISTS route_optimization_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, processing, completed, failed
    constraints JSONB DEFAULT '{}'::jsonb,
    stops JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_by UUID, -- Optional user ID
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_route_opt_req_org ON route_optimization_requests(organization_id);
CREATE INDEX idx_route_opt_req_status ON route_optimization_requests(status);

CREATE TABLE IF NOT EXISTS route_optimization_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES route_optimization_requests(id) ON DELETE CASCADE,
    routes JSONB NOT NULL DEFAULT '[]'::jsonb,
    total_distance DECIMAL(10, 2) DEFAULT 0,
    total_duration DECIMAL(10, 2) DEFAULT 0, -- in minutes
    total_cost DECIMAL(10, 2) DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_route_opt_res_req ON route_optimization_results(request_id);

CREATE TABLE IF NOT EXISTS route_optimization_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES route_optimization_requests(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'queued', -- queued, processing, completed, failed
    worker_id VARCHAR(100),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_route_opt_jobs_status ON route_optimization_jobs(status);
