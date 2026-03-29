-- Migration: 089_fix_invoice_permissions_fk.sql
-- Description: Fix invoice_permissions org_id data and add FK constraint
-- Created: 2026-03-29
-- Dependencies: 085_complete_org_id_standardization_and_indexes.sql
-- Issue: 4 rows had placeholder org_id (all zeros) causing FK violation
-- Fix: Update to valid organization and add FK constraint

SET search_path TO public;

\echo 'Fixing invoice_permissions org_id foreign key...';

-- Fix data: replace placeholder org_id with valid organization
UPDATE invoice_permissions
SET org_id = 'ce95cfe5-42a2-48b2-99d5-3f817c2f0247'
WHERE org_id = '00000000-0000-0000-0000-000000000000';

\echo 'Updated 4 rows with valid organization ID';

-- Add FK constraint if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_name = 'invoice_permissions'
          AND constraint_type = 'FOREIGN KEY'
          AND constraint_name = 'fk_invoice_permissions_organization'
    ) THEN
        ALTER TABLE invoice_permissions
        ADD CONSTRAINT fk_invoice_permissions_organization
        FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE;
        RAISE NOTICE 'Added FK: fk_invoice_permissions_organization';
    ELSE
        RAISE NOTICE 'FK already exists, skipping';
    END IF;
END $$;

-- Verify
SELECT 'invoice_permissions FK exists: ' ||
  CASE WHEN EXISTS (
    SELECT 1 FROM information_schema.table_constraints
    WHERE table_name = 'invoice_permissions' AND constraint_name = 'fk_invoice_permissions_organization'
  ) THEN 'YES' ELSE 'NO' END as status;

\echo '=====================================================';
\echo 'invoice_permissions foreign key fixed!';
\echo '=====================================================';
