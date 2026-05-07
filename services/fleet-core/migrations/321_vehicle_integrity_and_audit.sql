-- +goose Up
-- +goose StatementBegin
-- Migration: 321_vehicle_integrity_and_audit.sql
-- Purpose: Finalize vehicle data integrity, implement audit logging, and expand diagnostics.

-- 1. Uniqueness Constraint with Soft-Delete Awareness
-- Using a partial unique index to allow reuse of registration number after a vehicle is soft-deleted.
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_registration_active 
ON vehicles (registration_number) 
WHERE deleted_at IS NULL;

-- 2. Audit Logging System
CREATE TABLE IF NOT EXISTS vehicle_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    action VARCHAR(100) NOT NULL, -- e.g., STATUS_CHANGE, DRIVER_ASSIGNMENT, FIELD_UPDATE
    field_name VARCHAR(100),
    old_value TEXT,
    new_value TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_vehicle_audit_vehicle ON vehicle_audit_logs(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_audit_tenant ON vehicle_audit_logs(tenant_id);

-- 3. Document Versioning Support
ALTER TABLE vehicle_documents ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;

-- 4. Expanded Diagnostics in vehicles table
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS last_ping_ip VARCHAR(45); -- Support for IPv6
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS network_type VARCHAR(50); -- CELLULAR, WIFI, SATELLITE

-- 5. Soft-Delete Filter Enforcement Index
CREATE INDEX IF NOT EXISTS idx_vehicles_deleted_at_null ON vehicles(id) WHERE deleted_at IS NULL;

-- +goose StatementEnd
