-- Migration: 255_maintenance_enterprise_upgrade.sql
-- Description: Add enterprise maintenance capabilities including technician management, warranties, and procurement.

-- Pre-flight: Ensure maintenance_logs table exists (missed in early migrations)
CREATE TABLE IF NOT EXISTS maintenance_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    organization_id UUID, -- For enterprise support
    vehicle_id UUID NOT NULL,
    schedule_id TEXT,
    type TEXT,
    name TEXT,
    description TEXT,
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    duration_minutes INT,
    performed_by UUID,
    cost DECIMAL(15, 2) DEFAULT 0.00,
    total_cost DECIMAL(15, 2) DEFAULT 0.00,
    actual_cost DECIMAL(15, 2) DEFAULT 0.00,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    completion_date TIMESTAMPTZ
);

-- Phase 1: Technician Management
DO $$ BEGIN
    CREATE TYPE maintenance_technician_status AS ENUM ('AVAILABLE', 'BUSY', 'OFFLINE');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

CREATE TABLE IF NOT EXISTS maintenance_technicians (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    user_id UUID,
    name VARCHAR(255) NOT NULL,
    specialization VARCHAR(255),
    certifications JSONB DEFAULT '[]',
    hourly_rate DECIMAL(10, 2) DEFAULT 0.00,
    availability_status maintenance_technician_status DEFAULT 'OFFLINE',
    current_workload INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS technician_skills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    technician_id UUID NOT NULL REFERENCES maintenance_technicians(id) ON DELETE CASCADE,
    skill_name VARCHAR(255) NOT NULL,
    level INTEGER CHECK (level BETWEEN 1 AND 5),
    UNIQUE(technician_id, skill_name)
);

-- Phase 2: Warranty Management
CREATE TABLE IF NOT EXISTS component_warranties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    vehicle_id UUID NOT NULL,
    part_number VARCHAR(255) NOT NULL,
    warranty_provider VARCHAR(255),
    warranty_start DATE NOT NULL,
    warranty_end DATE NOT NULL,
    warranty_terms TEXT,
    claim_contact VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Phase 3: Procurement
DO $$ BEGIN
    CREATE TYPE purchase_order_status AS ENUM ('DRAFT', 'APPROVED', 'ORDERED', 'RECEIVED', 'CLOSED');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

CREATE TABLE IF NOT EXISTS purchase_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    vendor_id UUID,
    order_number VARCHAR(100) UNIQUE NOT NULL,
    status purchase_order_status DEFAULT 'DRAFT',
    total_cost DECIMAL(15, 2) DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS purchase_order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_order_id UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    part_id UUID,
    quantity INTEGER NOT NULL,
    unit_cost DECIMAL(10, 2) NOT NULL,
    total_cost DECIMAL(15, 2) NOT NULL
);

-- Phase 4: Downtime Tracking
CREATE TABLE IF NOT EXISTS vehicle_downtime (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    vehicle_id UUID NOT NULL,
    work_order_id UUID,
    downtime_start TIMESTAMP WITH TIME ZONE NOT NULL,
    downtime_end TIMESTAMP WITH TIME ZONE,
    downtime_minutes INTEGER DEFAULT 0,
    reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Phase 7: Lifecycle Metrics
CREATE TABLE IF NOT EXISTS vehicle_lifecycle_metrics (
    vehicle_id UUID PRIMARY KEY,
    total_maintenance_cost DECIMAL(15, 2) DEFAULT 0.00,
    maintenance_events INTEGER DEFAULT 0,
    downtime_hours DECIMAL(10, 2) DEFAULT 0.00,
    cost_per_km DECIMAL(10, 2) DEFAULT 0.00,
    replacement_recommendation TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Phase 5: Materialized view for cost analytics
DROP MATERIALIZED VIEW IF EXISTS maintenance_cost_summary;
CREATE MATERIALIZED VIEW maintenance_cost_summary AS
SELECT 
    COALESCE(organization_id, tenant_id) as organization_id,
    vehicle_id,
    DATE_TRUNC('month', created_at) as month,
    SUM(total_cost) as monthly_cost,
    COUNT(*) as maintenance_count
FROM (
    -- Combine maintenance_logs and potentially work_orders in future
    SELECT organization_id, tenant_id, vehicle_id, created_at, actual_cost as total_cost FROM maintenance_logs
) as all_costs
GROUP BY organization_id, tenant_id, vehicle_id, month;

CREATE UNIQUE INDEX IF NOT EXISTS idx_maintenance_cost_summary_org_veh_month ON maintenance_cost_summary(organization_id, vehicle_id, month);
