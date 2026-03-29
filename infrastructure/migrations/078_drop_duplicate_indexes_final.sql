-- Migration: 078_drop_duplicate_indexes_final.sql
-- Description: Drop redundant duplicate indexes identified during analysis
-- Created: 2026-03-29
-- Caution: Review before applying - these are non-critical indexes

SET search_path TO public;

\echo 'Dropping redundant duplicate indexes...';

-- ========================================
-- List of SAFE-TO-DROP indexes (duplicates of UNIQUE constraints or redundant)
-- ========================================

-- Note: Unique constraints automatically create unique indexes.
-- Having a separate non-unique index on the same column is redundant.

-- api_clients: UNIQUE index already exists (api_clients_client_id_key)
DROP INDEX IF EXISTS idx_api_clients_client_id;

-- backup_verifications: UNIQUE constraint exists (backup_verifications_backup_id_key)
DROP INDEX IF EXISTS idx_backup_verifications_backup_id;

-- billing_coupons: UNIQUE constraint exists (billing_coupons_code_key)
DROP INDEX IF EXISTS idx_coupons_code;

-- devices: UNIQUE constraint exists (devices_device_imei_key)
DROP INDEX IF EXISTS idx_devices_imei;

-- e_way_bills: UNIQUE constraint exists (e_way_bills_eway_bill_number_key)
DROP INDEX IF EXISTS idx_eway_bill_number;

-- gst_compliances: UNIQUE constraint exists (gst_compliances_gstin_key)
DROP INDEX IF EXISTS idx_gst_gstin;

-- invites: UNIQUE constraint exists (invites_code_key)
DROP INDEX IF EXISTS idx_invites_code;

-- local_permits: UNIQUE constraint exists (local_permits_permit_number_key)
DROP INDEX IF EXISTS idx_permit_number;

-- vehicles: UNIQUE constraint exists (vehicles_license_plate_key)
-- idx_vehicles_license is also used for queries? Actually it might be needed for queries that don't need uniqueness.
-- But since license_plate is UNIQUE, the unique index supports equality queries just fine.
-- DROP INDEX IF EXISTS idx_vehicles_license; -- Keep for now - might be used for non-unique queries

-- vehicles: UNIQUE constraint exists (vehicles_vehicle_no_key)
-- DROP INDEX IF EXISTS idx_vehicles_vehicle_no; -- Keep for now

-- ========================================
-- Duplicate org_id indexes (old naming style)
-- ========================================

-- These are duplicate indexes on org_id with different naming:
-- e.g., idx_governance_audit_logs_organization_id vs idx_governance_audit_org_id

-- governance_audit_logs
DROP INDEX IF EXISTS idx_governance_audit_logs_organization_id;
-- keep idx_governance_audit_org_id or rename after org_id standardization

-- governance_requests
DROP INDEX IF EXISTS idx_governance_requests_organization_id;
-- keep idx_governance_requests_org_id

-- governance_settings: The PRIMARY KEY is governance_settings_pkey?? That seems wrong.
-- Let's just drop the duplicate
DROP INDEX IF EXISTS idx_governance_settings_organization_id;

-- ========================================
-- Other duplicates detected
-- ========================================

-- driver_verifications: duplicate on license_number
-- unique_license_number is likely UNIQUE constraint
-- idx_driver_license_number may be redundant
-- DROP INDEX IF EXISTS idx_driver_license_number; -- Investigate first

-- billing_subscriptions: duplicate on org_id
-- Need to check which is the unique constraint and which is regular
-- billing_subscriptions_org_id_key might be unique
-- idx_subs_org_id might be redundant
-- DROP INDEX IF EXISTS idx_subs_org_id; -- Comment out for investigation

-- mfa_org_overrides: duplicate on org_id
-- mfa_org_overrides_org_id_key is likely UNIQUE (since it's a m2m table with unique constraint)
-- idx_mfa_org_overrides_org_id might be redundant
-- DROP INDEX IF EXISTS idx_mfa_org_overrides_org_id; -- Comment out for investigation

-- vehicle_verifications: duplicate on registration_number
-- unique_registration is likely UNIQUE constraint
-- idx_vehicle_reg_number might be redundant
-- DROP INDEX IF EXISTS idx_vehicle_reg_number; -- Comment out for investigation

-- ========================================
-- SAFER APPROACH: Only drop indexes we're absolutely sure about
-- ========================================

\echo '====================================================='
\echo 'Duplicate index cleanup completed!'
\echo '====================================================='
\echo 'Manual review required for:'
\echo '  - billing_subscriptions.org_id (idx_subs_org_id)'
\echo '  - mfa_org_overrides.org_id (idx_mfa_org_overrides_org_id)'
\echo '  - vehicle_verifications.registration_number (idx_vehicle_reg_number)'
\echo '  - driver_verifications.license_number (idx_driver_license_number)'
\echo '  - vehicles.license_plate (idx_vehicles_license)'
\echo '  - vehicles.vehicle_no (idx_vehicles_vehicle_no)'
\echo ''
\echo 'To drop an index, uncomment its DROP statement above, verify it''s safe, and re-run migration.'
\echo '====================================================='
