-- =====================================================
-- Performance Indexes Migration
-- Created: 2026-03-28
-- Purpose: Add missing indexes to optimize query performance
-- Based on: N+1 query analysis and common access patterns
-- =====================================================

-- Safety: Run in staging first, verify with EXPLAIN ANALYZE
-- This migration is REVERSE-SAFE (indexes can be dropped if needed)

\echo 'Creating performance indexes...'

-- =====================================================
-- 1. Foreign Key Indexes (Critical for JOINs)
-- =====================================================

-- Geofencing
CREATE INDEX IF NOT EXISTS idx_geofence_events_vehicle_id 
ON geofence_events(vehicle_id) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_geofence_events_time 
ON geofence_events(time DESC) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_geofence_assignments_vehicle 
ON geofence_assignments(vehicle_id, is_active) 
WHERE is_active = true;

-- Trips
CREATE INDEX IF NOT EXISTS idx_trips_driver_id 
ON trips(driver_id) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_trips_vehicle_id 
ON trips(vehicle_id) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_trips_time_range 
ON trips(tenant_id, start_time DESC, end_time DESC) 
WHERE deleted_at IS NULL;

-- Telemetry
CREATE INDEX IF NOT EXISTS idx_telemetry_points_device_batch 
ON telemetry_points(device_id, batch_id, recorded_at DESC) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_telemetry_points_time 
ON telemetry_points(recorded_at DESC) 
WHERE deleted_at IS NULL;

-- Billing
CREATE INDEX IF NOT EXISTS idx_invoices_org_status 
ON invoices(org_id, status, due_date) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_invoice_items_invoice_id 
ON invoice_items(invoice_id) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_payment_records_invoice 
ON payment_records(invoice_id, payment_date DESC) 
WHERE deleted_at IS NULL;

-- Reports
CREATE INDEX IF NOT EXISTS idx_scheduled_reports_org 
ON scheduled_reports(org_id, status, next_run AT TIME ZONE 'UTC') 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_report_executions_report 
ON report_executions(report_id, started_at DESC) 
WHERE deleted_at IS NULL;

-- Auth
CREATE INDEX IF NOT EXISTS idx_mfa_recovery_codes_user 
ON mfa_recovery_codes(user_id, used) 
WHERE used = false;

CREATE INDEX IF NOT EXISTS idx_login_history_user_time 
ON login_history(user_id, login_time DESC) 
WHERE deleted_at IS NULL;

-- Organizations
CREATE INDEX IF NOT EXISTS idx_organizations_status 
ON organizations(status, is_deleted) 
WHERE is_deleted = false;

-- Users
CREATE INDEX IF NOT EXISTS idx_users_org_email 
ON users(org_id, email) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_users_driver_id 
ON users(driver_id) 
WHERE driver_id IS NOT NULL AND deleted_at IS NULL;

-- Vehicles
CREATE INDEX IF NOT EXISTS idx_vehicles_org_status 
ON vehicles(org_id, status, is_active) 
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_vehicles_device_id 
ON vehicles(device_id) 
WHERE device_id IS NOT NULL AND deleted_at IS NULL;

-- Drivers
CREATE INDEX IF NOT EXISTS idx_drivers_org_status 
ON drivers(org_id, status, is_active) 
WHERE deleted_at IS NULL;

-- =====================================================
-- 2. Partial Indexes for Common Filter Patterns
-- =====================================================

-- Active Records (excluding deleted)
CREATE INDEX IF NOT EXISTS idx_vehicle_verifications_org_pending 
ON vehicle_verifications(org_id, verification_date DESC) 
WHERE status = 'pending' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_driver_verifications_org_pending 
ON driver_verifications(org_id, verification_date DESC) 
WHERE status = 'pending' AND deleted_at IS NULL;

-- Recent Records (time-based queries)
CREATE INDEX IF NOT EXISTS idx_geofence_events_recent 
ON geofence_events(vehicle_id, time DESC) 
WHERE time >= NOW() - INTERVAL '7 days' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_telemetry_points_recent 
ON telemetry_points(device_id, recorded_at DESC) 
WHERE recorded_at >= NOW() - INTERVAL '24 hours' AND deleted_at IS NULL;

-- =====================================================
-- 3. Composite Indexes for Common Query Patterns
-- =====================================================

-- Organization queries with status filters
CREATE INDEX IF NOT EXISTS idx_orgs_active_lookup 
ON organizations(is_deleted, status, created_at DESC) 
WHERE is_deleted = false AND status = 'active';

-- Invoice queries by org + status + date
CREATE INDEX IF NOT EXISTS idx_invoices_org_date 
ON invoices(org_id, invoice_date DESC, status) 
WHERE deleted_at IS NULL;

-- Trip queries by driver + time range
CREATE INDEX IF NOT EXISTS idx_trips_driver_daterange 
ON trips(driver_id, start_time, end_time) 
WHERE deleted_at IS NULL;

-- =====================================================
-- 4. GIN/GIST Indexes for JSON/Array Columns (if used)
-- =====================================================

-- Example: If using JSONB for filters/metadata
-- CREATE INDEX IF NOT EXISTS idx_scheduled_reports_filters_gin 
-- ON scheduled_reports USING GIN((filters->>'report_type'));

-- =====================================================
-- 5. Unique Constraints (if missing)
-- =====================================================

-- Ensure unique constraints where needed
-- ALTER TABLE mfa_recovery_codes ADD CONSTRAINT unique_user_active_code UNIQUE (user_id, code_hash) WHERE used = false;

\echo 'Performance indexes created successfully!'
\echo 'Next steps:'
\echo '  1. Analyze with EXPLAIN ANALYZE: EXPLAIN (ANALYZE, BUFFERS) your slow queries'
\echo '  2. Check pg_stat_statements for query improvements'
\echo '  3. Monitor index usage: SELECT * FROM pg_stat_user_indexes WHERE idx_scan > 0;'
\echo '  4. Consider CONCURRENTLY if adding to production: CREATE INDEX CONCURRENTLY ...'

