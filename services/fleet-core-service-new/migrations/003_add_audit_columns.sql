-- Migration: Add Audit Columns (CreatedBy/UpdatedBy)
-- Note: UUID references users(id) in the main shared schema

ALTER TABLE vehicles 
ADD COLUMN IF NOT EXISTS created_by UUID,
ADD COLUMN IF NOT EXISTS updated_by UUID;

ALTER TABLE drivers
ADD COLUMN IF NOT EXISTS created_by UUID,
ADD COLUMN IF NOT EXISTS updated_by UUID;

ALTER TABLE work_orders
ADD COLUMN IF NOT EXISTS created_by UUID,
ADD COLUMN IF NOT EXISTS updated_by UUID;

ALTER TABLE vehicle_inspections
ADD COLUMN IF NOT EXISTS created_by UUID,
ADD COLUMN IF NOT EXISTS updated_by UUID;

ALTER TABLE fuel_logs
ADD COLUMN IF NOT EXISTS created_by UUID,
ADD COLUMN IF NOT EXISTS updated_by UUID;
