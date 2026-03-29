-- Migration: 080_add_remaining_audit_fk_indexes.sql
-- Description: Add indexes to remaining audit foreign key columns (created_by, updated_by, verified_by)
-- Created: 2026-03-29
-- Purpose: Complete FK index coverage for better JOIN performance on audit queries

SET search_path TO public;

\echo 'Adding remaining audit foreign key indexes...';

-- ========================================
-- These indexes were missed in migration 070
-- They are important for queries that filter by audit columns
-- ========================================

-- access_reviews.created_by
CREATE INDEX IF NOT EXISTS idx_access_reviews_created_by ON access_reviews(created_by) WHERE created_by IS NOT NULL;

-- alert_rules.created_by
CREATE INDEX IF NOT EXISTS idx_alert_rules_created_by ON alert_rules(created_by) WHERE created_by IS NOT NULL;

-- driver_documents.created_by
CREATE INDEX IF NOT EXISTS idx_driver_documents_created_by ON driver_documents(created_by) WHERE created_by IS NOT NULL;

-- driver_documents.updated_by
CREATE INDEX IF NOT EXISTS idx_driver_documents_updated_by ON driver_documents(updated_by) WHERE updated_by IS NOT NULL;

-- driver_documents.verified_by
CREATE INDEX IF NOT EXISTS idx_driver_documents_verified_by ON driver_documents(verified_by) WHERE verified_by IS NOT NULL;

-- drivers.updated_by
CREATE INDEX IF NOT EXISTS idx_drivers_updated_by ON drivers(updated_by) WHERE updated_by IS NOT NULL;

-- fuel_logs.updated_by
CREATE INDEX IF NOT EXISTS idx_fuel_logs_updated_by ON fuel_logs(updated_by) WHERE updated_by IS NOT NULL;

\echo '====================================================='
\echo 'Audit FK indexes added!'
\echo '====================================================='
\echo 'Verification:'
\echo '  SELECT tablename, indexname FROM pg_indexes WHERE indexname IN ('
\echo '    ''idx_access_reviews_created_by'','
\echo '    ''idx_alert_rules_created_by'','
\echo '    ''idx_driver_documents_created_by'','
\echo '    ''idx_driver_documents_updated_by'','
\echo '    ''idx_driver_documents_verified_by'','
\echo '    ''idx_drivers_updated_by'','
\echo '    ''idx_fuel_logs_updated_by'''
\echo '  );'
\echo '====================================================='
