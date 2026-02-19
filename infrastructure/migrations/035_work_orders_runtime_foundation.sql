-- Runtime foundation for maintenance work orders.
-- Keeps local integration environments functional even when earlier migration chains drift.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS work_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    vehicle_id UUID,
    title TEXT NOT NULL,
    description TEXT,
    maintenance_type TEXT NOT NULL DEFAULT 'PREVENTIVE',
    priority TEXT NOT NULL DEFAULT 'MEDIUM',
    status TEXT NOT NULL DEFAULT 'OPEN',
    assigned_to UUID,
    scheduled_date TIMESTAMPTZ,
    completed_date TIMESTAMPTZ,
    estimated_cost DOUBLE PRECISION NOT NULL DEFAULT 0,
    actual_cost DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    vendor_id UUID,
    approval_status TEXT NOT NULL DEFAULT 'PENDING',
    cost_breakdown JSONB NOT NULL DEFAULT '{}'::jsonb,
    inspection_id UUID
);

ALTER TABLE work_orders ADD COLUMN IF NOT EXISTS organization_id UUID;
ALTER TABLE work_orders ADD COLUMN IF NOT EXISTS vehicle_id UUID;
ALTER TABLE work_orders ADD COLUMN IF NOT EXISTS title TEXT;
ALTER TABLE work_orders ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE work_orders ADD COLUMN IF NOT EXISTS maintenance_type TEXT NOT NULL DEFAULT 'PREVENTIVE';
ALTER TABLE work_orders ADD COLUMN IF NOT EXISTS priority TEXT NOT NULL DEFAULT 'MEDIUM';
ALTER TABLE work_orders ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'OPEN';
ALTER TABLE work_orders ADD COLUMN IF NOT EXISTS assigned_to UUID;
ALTER TABLE work_orders ADD COLUMN IF NOT EXISTS scheduled_date TIMESTAMPTZ;
ALTER TABLE work_orders ADD COLUMN IF NOT EXISTS completed_date TIMESTAMPTZ;
ALTER TABLE work_orders ADD COLUMN IF NOT EXISTS estimated_cost DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE work_orders ADD COLUMN IF NOT EXISTS actual_cost DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE work_orders ADD COLUMN IF NOT EXISTS created_by UUID;
ALTER TABLE work_orders ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE work_orders ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE work_orders ADD COLUMN IF NOT EXISTS vendor_id UUID;
ALTER TABLE work_orders ADD COLUMN IF NOT EXISTS approval_status TEXT NOT NULL DEFAULT 'PENDING';
ALTER TABLE work_orders ADD COLUMN IF NOT EXISTS cost_breakdown JSONB NOT NULL DEFAULT '{}'::jsonb;
ALTER TABLE work_orders ADD COLUMN IF NOT EXISTS inspection_id UUID;

CREATE INDEX IF NOT EXISTS idx_work_orders_org ON work_orders (organization_id);
CREATE INDEX IF NOT EXISTS idx_work_orders_org_status ON work_orders (organization_id, status);
CREATE INDEX IF NOT EXISTS idx_work_orders_org_priority ON work_orders (organization_id, priority);
CREATE INDEX IF NOT EXISTS idx_work_orders_created_at ON work_orders (created_at DESC);
