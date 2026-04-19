-- Repair Invoicing Tables Schema
-- Migration: 310_repair_invoicing_full.sql

-- 1. Repair 'invoices' table
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS customer_phone TEXT;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS created_by UUID;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS updated_by UUID;

-- Backfill created_by/updated_by with a system-level user if NULL
UPDATE invoices SET created_by = '63a20bb5-f3f4-48df-8a43-fc2d0a50c84c' WHERE created_by IS NULL;
UPDATE invoices SET updated_by = '63a20bb5-f3f4-48df-8a43-fc2d0a50c84c' WHERE updated_by IS NULL;

-- 2. Repair 'invoice_items' table
ALTER TABLE invoice_items ADD COLUMN IF NOT EXISTS created_by UUID;
ALTER TABLE invoice_items ADD COLUMN IF NOT EXISTS updated_by UUID;

-- Backfill invoice_items creator/updater
UPDATE invoice_items SET created_by = '63a20bb5-f3f4-48df-8a43-fc2d0a50c84c' WHERE created_by IS NULL;
UPDATE invoice_items SET updated_by = '63a20bb5-f3f4-48df-8a43-fc2d0a50c84c' WHERE updated_by IS NULL;

-- 3. Ensure billing_address is TEXT (already exists but for safety)
-- Duplicate columns or legacy ones might exist, but we stick to whatrepository expects.
