-- Migration: 002_create_inventory_parts.sql
-- Creates the inventory_parts table for parts management

CREATE TABLE IF NOT EXISTS inventory_parts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    part_number VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    stock_location VARCHAR(100),
    quantity_available INT NOT NULL DEFAULT 0 CHECK (quantity_available >= 0),
    reorder_level INT NOT NULL DEFAULT 5,
    unit_cost DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT unique_part_per_org UNIQUE (organization_id, part_number)
);

-- Indexes
CREATE INDEX idx_inventory_organization ON inventory_parts(organization_id);
CREATE INDEX idx_inventory_part_number ON inventory_parts(part_number);
CREATE INDEX idx_inventory_low_stock ON inventory_parts(organization_id, quantity_available) 
    WHERE quantity_available <= reorder_level;

-- Updated_at trigger
CREATE TRIGGER trigger_inventory_parts_updated_at
    BEFORE UPDATE ON inventory_parts
    FOR EACH ROW
    EXECUTE FUNCTION update_work_orders_updated_at();

COMMENT ON TABLE inventory_parts IS 'Maintenance parts inventory';
