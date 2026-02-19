-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Work Orders Table
CREATE TABLE IF NOT EXISTS work_orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL,
    vehicle_id UUID NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    maintenance_type TEXT NOT NULL, -- preventive, corrective, inspection, breakdown
    priority TEXT NOT NULL, -- LOW, MEDIUM, HIGH, CRITICAL
    status TEXT NOT NULL DEFAULT 'OPEN', -- OPEN, IN_PROGRESS, ON_HOLD, COMPLETED
    assigned_to UUID, -- User ID (Mechanic)
    scheduled_date TIMESTAMP WITH TIME ZONE,
    completed_date TIMESTAMP WITH TIME ZONE,
    estimated_cost DECIMAL(10, 2) DEFAULT 0,
    actual_cost DECIMAL(10, 2) DEFAULT 0,
    vendor_id UUID, -- Link to maintenance_vendors
    approval_status TEXT DEFAULT 'NOT_REQUIRED', -- PENDING_MANAGER, APPROVED, etc.
    cost_breakdown JSONB DEFAULT '{}'::jsonb, -- Store labor, parts, external costs
    inspection_id UUID, -- Link to vehicle_inspections if generated from one
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Inventory Parts Table
CREATE TABLE IF NOT EXISTS inventory_parts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL,
    part_number TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    stock_location TEXT,
    quantity_available INT NOT NULL DEFAULT 0,
    reorder_level INT NOT NULL DEFAULT 5,
    unit_cost DECIMAL(10, 2) NOT NULL DEFAULT 0,
    category TEXT,
    compatible_vehicles JSONB DEFAULT '[]'::jsonb, -- Array of vehicle types/models
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(organization_id, part_number)
);

-- Work Order Parts Join Table
CREATE TABLE IF NOT EXISTS work_order_parts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    work_order_id UUID NOT NULL REFERENCES work_orders(id) ON DELETE CASCADE,
    part_id UUID NOT NULL REFERENCES inventory_parts(id),
    quantity INT NOT NULL DEFAULT 1,
    unit_cost DECIMAL(10, 2) NOT NULL, -- Snapshot cost at time of assignment
    total_cost DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Maintenance Settings Table
CREATE TABLE IF NOT EXISTS maintenance_settings (
    organization_id UUID PRIMARY KEY,
    sla_p1_hours INT DEFAULT 4, -- Critical
    sla_p2_hours INT DEFAULT 24, -- High
    sla_p3_hours INT DEFAULT 72, -- Medium
    low_stock_threshold INT DEFAULT 5,
    approval_threshold DECIMAL(10, 2) DEFAULT 5000.00,
    auto_assign_enabled BOOLEAN DEFAULT FALSE,
    notification_enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_work_orders_org ON work_orders(organization_id);
CREATE INDEX idx_work_orders_vehicle ON work_orders(vehicle_id);
CREATE INDEX idx_work_orders_status ON work_orders(status);
CREATE INDEX idx_inventory_org ON inventory_parts(organization_id);
