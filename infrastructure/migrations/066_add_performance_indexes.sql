-- Migration: 065_add_performance_indexes.sql
-- Description: Add high-impact composite indexes for common query patterns
-- Created: 2026-03-28

SET search_path TO public;

-- ========================================
-- Support Tickets Indexes
-- ========================================
-- Common query patterns:
-- - "Show open high-priority tickets for my org"
-- - "Get ticket history with comments for org"
-- - "Support dashboard: tickets by status/priority"

-- Composite index for ticket filtering (most common support queries)
CREATE INDEX IF NOT EXISTS idx_support_tickets_org_status_priority_created
ON support_tickets(org_id, status, priority, created_at DESC);

-- Composite index for assigned tickets performance
CREATE INDEX IF NOT EXISTS idx_support_tickets_assigned_status_created
ON support_tickets(assigned_to, status, created_at DESC)
WHERE assigned_to IS NOT NULL;

-- Incident comments with ticket lookup
CREATE INDEX IF NOT EXISTS idx_incident_comments_ticket_created
ON incident_comments(ticket_id, created_at DESC);

-- ========================================
-- Trips Indexes
-- Common queries: vehicle trip history, driver assignments, org-level reporting
-- ========================================

-- Composite for vehicle trip history (common: "show all trips for vehicle X")
-- Uses scheduled_start (the trip's planned start time)
CREATE INDEX IF NOT EXISTS idx_trips_org_vehicle_time
ON trips(org_id, vehicle_id, scheduled_start DESC);

-- Composite for driver trip schedule (common: "upcoming trips for driver Y")
-- Includes status for filtering (scheduled, in_progress, etc.)
CREATE INDEX IF NOT EXISTS idx_trips_org_driver_status_time
ON trips(org_id, driver_id, status, scheduled_start DESC);

-- ========================================
-- Billing Ledger Indexes
-- Common: financial reports per org, time-based queries, invoice lookups
-- ========================================

-- Time-based queries for billing reports
CREATE INDEX IF NOT EXISTS idx_ledger_org_created
ON billing_ledger_entries(org_id, created_at DESC);

-- Invoice-focused queries
CREATE INDEX IF NOT EXISTS idx_ledger_invoice_org
ON billing_ledger_entries(invoice_id, org_id)
WHERE invoice_id IS NOT NULL;

-- Credit transaction history for balance reconciliation
CREATE INDEX IF NOT EXISTS idx_credit_txn_balance_created
ON billing_credit_transactions(balance_id, created_at DESC);

-- ========================================
-- Driver Behavior Indexes
-- Common: recent behavior scores, driver trend analysis
-- ========================================

-- Optimization for: "get latest behavior score for driver"
CREATE INDEX IF NOT EXISTS idx_driver_behavior_driver_recent
ON driver_behavior(driver_id, period_start DESC);

-- Org-level behavior queries for fleet analytics
CREATE INDEX IF NOT EXISTS idx_driver_behavior_org_period
ON driver_behavior(org_id, period_start DESC);

-- ========================================
-- Telemetry Indexes (TimescaleDB hypertables already have time index)
-- Add vehicle + org composite for common filter patterns
-- ========================================

-- GPS locations filtered by vehicle and org
CREATE INDEX IF NOT EXISTS idx_gps_vehicle_org_time
ON gps_locations(vehicle_id, org_id, time DESC);

-- OBD readings with vehicle + org filtering
CREATE INDEX IF NOT EXISTS idx_obd_vehicle_org_time
ON obd_readings(vehicle_id, org_id, time DESC);

-- Fuel readings composite index
CREATE INDEX IF NOT EXISTS idx_fuel_vehicle_org_time
ON fuel_readings(vehicle_id, org_id, time DESC);

-- Vehicle telemetry composite
CREATE INDEX IF NOT EXISTS idx_vehicle_telemetry_vehicle_org_time
ON vehicle_telemetry(vehicle_id, org_id, time DESC);

-- ========================================
-- Driver Incidents Indexes
-- ========================================

-- Driver incident history queries
CREATE INDEX IF NOT EXISTS idx_incidents_driver_time
ON driver_incidents(driver_id, timestamp DESC);

-- Org-level incident reporting
CREATE INDEX IF NOT EXISTS idx_incidents_org_type_time
ON driver_incidents(org_id, type, timestamp DESC);

-- ========================================
-- Validation: ensure all indexes created successfully
-- ========================================
DO $$
DECLARE
    missing_indexes INT := 0;
BEGIN
    -- Check critical indexes exist
    SELECT COUNT(*) INTO missing_indexes
    FROM pg_indexes
    WHERE schemaname = 'public'
      AND indexname IN (
          'idx_support_tickets_org_status_priority_created',
          'idx_trips_org_vehicle_time',
          'idx_trips_org_driver_status_time',
          'idx_ledger_org_created',
          'idx_driver_behavior_driver_recent',
          'idx_gps_vehicle_org_time'
      );

    IF missing_indexes != 6 THEN
        RAISE WARNING 'Some indexes may not have been created. Count: %', missing_indexes;
    END IF;
END $$;

COMMENT ON INDEX idx_support_tickets_org_status_priority_created IS 'Optimizes support dashboard queries by org, status, priority';
COMMENT ON INDEX idx_trips_org_vehicle_time IS 'Speeds up vehicle trip history queries';
COMMENT ON INDEX idx_trips_org_driver_status_time IS 'Optimizes driver schedule and upcoming trip lookups';
COMMENT ON INDEX idx_ledger_org_created IS 'Enables fast time-based billing reports per org';
COMMENT ON INDEX idx_driver_behavior_driver_recent IS 'Quick retrieval of latest driver behavior scores';
COMMENT ON INDEX idx_gps_vehicle_org_time IS 'Optimizes telemetry queries filtered by vehicle and org';
