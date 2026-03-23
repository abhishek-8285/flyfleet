-- Migration: 226_trip_cost_attribution
-- Description: Link fuel and expenses to trips and routes

-- Add columns to fuel_logs
ALTER TABLE fuel_logs
ADD COLUMN IF NOT EXISTS trip_id UUID,
ADD COLUMN IF NOT EXISTS route_id UUID;

-- Add columns to fuel_transactions
ALTER TABLE fuel_transactions
ADD COLUMN IF NOT EXISTS trip_id UUID,
ADD COLUMN IF NOT EXISTS route_id UUID;

-- Add columns to expense_reports
ALTER TABLE expense_reports
ADD COLUMN IF NOT EXISTS trip_id UUID,
ADD COLUMN IF NOT EXISTS route_id UUID;

-- Indexes
CREATE INDEX idx_fuel_logs_trip ON fuel_logs(trip_id);
CREATE INDEX idx_expense_reports_trip ON expense_reports(trip_id);

-- Comments
COMMENT ON COLUMN fuel_logs.trip_id IS 'Associated trip ID';
COMMENT ON COLUMN expense_reports.trip_id IS 'Associated trip ID';
