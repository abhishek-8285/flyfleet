-- Migration: 079_drop_redundant_unique_duplicate_indexes.sql
-- Description: Drop non-unique indexes that duplicate unique constraint indexes
-- Created: 2026-03-29
-- Analysis: All target columns have UNIQUE constraints, making separate indexes redundant

SET search_path TO public;

\echo 'Dropping redundant duplicate indexes on uniquely constrained columns...';

-- ========================================
-- These indexes are safe to drop because:
-- 1. The column has a UNIQUE constraint
-- 2. The UNIQUE constraint automatically creates a unique index
-- 3. Both indexes serve the same query optimization purpose
-- 4. The unique index is strictly better (supports equality + enforces uniqueness)
-- ========================================

-- billing_subscriptions.org_id
-- Unique constraint: billing_subscriptions_org_id_key
-- Redundant index: idx_subs_org_id
DROP INDEX IF EXISTS idx_subs_org_id;

-- mfa_org_overrides.org_id
-- Unique constraint: mfa_org_overrides_org_id_key
-- Redundant index: idx_mfa_org_overrides_org_id
DROP INDEX IF EXISTS idx_mfa_org_overrides_org_id;

-- vehicle_verifications.registration_number
-- Unique constraint: unique_registration
-- Redundant index: idx_vehicle_reg_number
DROP INDEX IF EXISTS idx_vehicle_reg_number;

-- driver_verifications.license_number
-- Unique constraint: unique_license_number
-- Redundant index: idx_driver_license_number
DROP INDEX IF EXISTS idx_driver_license_number;

-- vehicles.license_plate
-- Unique constraint: vehicles_license_plate_key
-- Redundant index: idx_vehicles_license
DROP INDEX IF EXISTS idx_vehicles_license;

-- vehicles.vehicle_no
-- Unique constraint: vehicles_vehicle_no_key
-- Redundant index: idx_vehicles_vehicle_no
DROP INDEX IF EXISTS idx_vehicles_vehicle_no;

\echo '====================================================='
\echo 'Redundant duplicate indexes dropped!'
\echo '====================================================='
\echo 'Verification:'
\echo '  SELECT indexname FROM pg_indexes WHERE indexname IN ('
\echo '    ''idx_subs_org_id'', ''idx_mfa_org_overrides_org_id'','
\echo '    ''idx_vehicle_reg_number'', ''idx_driver_license_number'','
\echo '    ''idx_vehicles_license'', ''idx_vehicles_vehicle_no'''
\echo '  ); -- Should return 0 rows'
\echo '====================================================='
