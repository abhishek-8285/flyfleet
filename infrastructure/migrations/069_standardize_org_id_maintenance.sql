-- Migration: 069_standardize_org_id_maintenance.sql
-- Description: Rename organization_id to org_id across all maintenance tables
-- Created: 2026-03-28

SET search_path TO public;

-- ========================================
-- Rename organization_id → org_id in maintenance tables
-- ========================================

-- maintenance_vendors
ALTER TABLE IF EXISTS maintenance_vendors RENAME COLUMN organization_id TO org_id;

-- inspection_templates
ALTER TABLE IF EXISTS inspection_templates RENAME COLUMN organization_id TO org_id;

-- vehicle_inspections
ALTER TABLE IF EXISTS vehicle_inspections RENAME COLUMN organization_id TO org_id;

-- maintenance_settings
ALTER TABLE IF EXISTS maintenance_settings RENAME COLUMN organization_id TO org_id;

-- inventory_parts
ALTER TABLE IF EXISTS inventory_parts RENAME COLUMN organization_id TO org_id;

-- purchase_orders
ALTER TABLE IF EXISTS purchase_orders RENAME COLUMN organization_id TO org_id;

-- maintenance_technicians
ALTER TABLE IF EXISTS maintenance_technicians RENAME COLUMN organization_id TO org_id;

-- component_warranties
ALTER TABLE IF EXISTS component_warranties RENAME COLUMN organization_id TO org_id;

-- vehicle_downtime
ALTER TABLE IF EXISTS vehicle_downtime RENAME COLUMN organization_id TO org_id;

-- maintenance_schedules
ALTER TABLE IF EXISTS maintenance_schedules RENAME COLUMN organization_id TO org_id;

-- maintenance_logs
ALTER TABLE IF EXISTS maintenance_logs RENAME COLUMN organization_id TO org_id;

-- ========================================
-- Rename indexes that include organization_id (optional but nice)
-- ========================================

-- Indexes on organization_id are automatically renamed by PostgreSQL,
-- but we may want to rename them for consistency.
-- Note: index names remain the same effectively, but the column they reference changed.

-- Example: idx_maintenance_vendors_org -> idx_maintenance_vendors_org_id? Actually the index name does not contain column name typically.
-- The existing indexes: idx_maintenance_vendors_org (from migration 034) references org_id after rename automatically.

-- ========================================
-- Verify: Check that columns renamed
-- ========================================
DO $$
DECLARE
    tables_renamed INT := 0;
BEGIN
    SELECT COUNT(*) INTO tables_renamed
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name IN (
          'maintenance_vendors', 'inspection_templates', 'vehicle_inspections',
          'maintenance_settings', 'inventory_parts', 'purchase_orders',
          'maintenance_technicians', 'component_warranties', 'vehicle_downtime',
          'maintenance_schedules', 'maintenance_logs'
      )
      AND column_name = 'org_id';

    RAISE NOTICE 'Maintenance tables with org_id column: % (expected 11)', tables_renamed;
END $$;
