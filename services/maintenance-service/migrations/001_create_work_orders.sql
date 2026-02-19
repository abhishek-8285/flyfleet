-- Migration: 001_create_work_orders.sql
-- Creates the work_orders table for maintenance module

CREATE TYPE maintenance_type AS ENUM ('preventive', 'corrective', 'inspection', 'breakdown');
CREATE TYPE work_order_priority AS ENUM ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL');
CREATE TYPE work_order_status AS ENUM ('OPEN', 'IN_PROGRESS', 'ON_HOLD', 'COMPLETED');

CREATE TABLE IF NOT EXISTS work_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    vehicle_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    maintenance_type maintenance_type NOT NULL DEFAULT 'corrective',
    priority work_order_priority NOT NULL DEFAULT 'MEDIUM',
    status work_order_status NOT NULL DEFAULT 'OPEN',
    assigned_to UUID,
    scheduled_date TIMESTAMP WITH TIME ZONE,
    completed_date TIMESTAMP WITH TIME ZONE,
    estimated_cost DECIMAL(10, 2) DEFAULT 0.00,
    actual_cost DECIMAL(10, 2) DEFAULT 0.00,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX idx_work_orders_organization ON work_orders(organization_id);
CREATE INDEX idx_work_orders_vehicle ON work_orders(vehicle_id);
CREATE INDEX idx_work_orders_status ON work_orders(status);
CREATE INDEX idx_work_orders_priority ON work_orders(priority);
CREATE INDEX idx_work_orders_assigned ON work_orders(assigned_to);
CREATE INDEX idx_work_orders_scheduled ON work_orders(scheduled_date);

-- Updated_at trigger
CREATE OR REPLACE FUNCTION update_work_orders_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_work_orders_updated_at
    BEFORE UPDATE ON work_orders
    FOR EACH ROW
    EXECUTE FUNCTION update_work_orders_updated_at();

COMMENT ON TABLE work_orders IS 'Fleet maintenance work orders';
