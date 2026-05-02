-- Migration: 034_maintenance_system
-- Description: Establishes the identity foundation (tenants) and maintenance system.

-- 1. Identity Foundation (Tenants)
-- This MUST be created before anything else that references it.
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    tax_id TEXT UNIQUE,
    status TEXT DEFAULT 'active',
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Seed System Tenant if not exists
INSERT INTO tenants (id, name, tax_id, status) 
VALUES ('00000000-0000-0000-0000-000000000000', 'System', 'SYSTEM-001', 'active')
ON CONFLICT DO NOTHING;

-- Create an alias view for 'organizations' to support legacy migrations
CREATE OR REPLACE VIEW organizations AS SELECT * FROM tenants;

-- 2. Maintenance System
CREATE TABLE IF NOT EXISTS work_orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    vehicle_id UUID NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    maintenance_type TEXT NOT NULL,
    priority TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'OPEN',
    assigned_to UUID,
    scheduled_date TIMESTAMP WITH TIME ZONE,
    completed_date TIMESTAMP WITH TIME ZONE,
    estimated_cost DECIMAL(10, 2) DEFAULT 0,
    actual_cost DECIMAL(10, 2) DEFAULT 0,
    vendor_id UUID,
    approval_status TEXT DEFAULT 'NOT_REQUIRED',
    cost_breakdown JSONB DEFAULT '{}'::jsonb,
    inspection_id UUID,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS inventory_parts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    part_number TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    stock_location TEXT,
    quantity_available INT NOT NULL DEFAULT 0,
    reorder_level INT NOT NULL DEFAULT 5,
    unit_cost DECIMAL(10, 2) NOT NULL DEFAULT 0,
    category TEXT,
    compatible_vehicles JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, part_number)
);

CREATE TABLE IF NOT EXISTS work_order_parts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    work_order_id UUID NOT NULL REFERENCES work_orders(id) ON DELETE CASCADE,
    part_id UUID NOT NULL REFERENCES inventory_parts(id),
    quantity INT NOT NULL DEFAULT 1,
    unit_cost DECIMAL(10, 2) NOT NULL,
    total_cost DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS maintenance_settings (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id),
    sla_p1_hours INT DEFAULT 4,
    sla_p2_hours INT DEFAULT 24,
    sla_p3_hours INT DEFAULT 72,
    low_stock_threshold INT DEFAULT 5,
    approval_threshold DECIMAL(10, 2) DEFAULT 5000.00,
    auto_assign_enabled BOOLEAN DEFAULT FALSE,
    notification_enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Indexes
CREATE INDEX IF NOT EXISTS idx_work_orders_tenant ON work_orders(tenant_id);
CREATE INDEX IF NOT EXISTS idx_work_orders_vehicle ON work_orders(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_work_orders_status ON work_orders(status);
CREATE INDEX IF NOT EXISTS idx_inventory_tenant ON inventory_parts(tenant_id);
