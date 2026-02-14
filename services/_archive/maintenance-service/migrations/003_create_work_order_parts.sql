-- Migration: 003_create_work_order_parts.sql
-- Creates the junction table for work order to parts relationship

CREATE TABLE IF NOT EXISTS work_order_parts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    work_order_id UUID NOT NULL REFERENCES work_orders(id) ON DELETE CASCADE,
    part_id UUID NOT NULL REFERENCES inventory_parts(id) ON DELETE RESTRICT,
    quantity INT NOT NULL CHECK (quantity > 0),
    unit_cost DECIMAL(10, 2) NOT NULL,
    total_cost DECIMAL(10, 2) GENERATED ALWAYS AS (quantity * unit_cost) STORED,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT unique_part_per_order UNIQUE (work_order_id, part_id)
);

-- Indexes
CREATE INDEX idx_work_order_parts_order ON work_order_parts(work_order_id);
CREATE INDEX idx_work_order_parts_part ON work_order_parts(part_id);

COMMENT ON TABLE work_order_parts IS 'Parts assigned to work orders';
