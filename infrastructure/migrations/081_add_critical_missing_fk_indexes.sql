-- Migration: 081_add_critical_missing_fk_indexes.sql
-- Description: Add remaining critical FK indexes for query performance
-- Created: 2026-03-29
-- Priority: High - these are on frequently queried columns (org_id, device_id, author_id)

SET search_path TO public;

\echo 'Adding critical missing foreign key indexes...';

-- ========================================
-- Note: telemetry_events is partitioned and already has indexes on child partitions
--       (telemetry_events_p*_org_id_created_at_idx). Skip it.
-- ========================================

-- alert_rules.org_id
-- Critical for: "get all alert rules for org X"
-- Has duplicate FK constraints (both fk_alert_rules_organization and alert_rules_org_id_fkey)
-- Index supports both
-- Note: alert_rules does not have deleted_at column, so no partial index
CREATE INDEX IF NOT EXISTS idx_alert_rules_org_id ON alert_rules(org_id);

-- driver_audit_logs.org_id
-- Critical for: "get audit log for org X"
CREATE INDEX IF NOT EXISTS idx_driver_audit_logs_org_id ON driver_audit_logs(org_id);

-- incident_comments.author_id
-- Used for: "get all comments by user Y"
CREATE INDEX IF NOT EXISTS idx_incident_comments_author_id ON incident_comments(author_id) WHERE author_id IS NOT NULL;

-- vehicle_assignments.device_id
-- Used for: "find which vehicle a device is assigned to"
CREATE INDEX IF NOT EXISTS idx_vehicle_assignments_device_id ON vehicle_assignments(device_id) WHERE device_id IS NOT NULL;

-- vehicle_inspections.org_id
-- Critical for: "get all inspections for org X"
CREATE INDEX IF NOT EXISTS idx_vehicle_inspections_org_id ON vehicle_inspections(org_id);

\echo '====================================================='
\echo 'Critical FK indexes added!'
\echo '====================================================='
\echo 'Remaining unindexed FKs should be low-priority audit columns or edge cases.'
\echo '====================================================='
