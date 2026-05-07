-- +goose Up
-- +goose StatementBegin
-- Migration: 320_vehicle_document_system.sql
-- Purpose: Implement a robust, unified document model and harden operational constraints.

-- 1. Hardened Operational Constraints on vehicles table
ALTER TABLE vehicles DROP CONSTRAINT IF EXISTS chk_ais140_gps;
ALTER TABLE vehicles ADD CONSTRAINT chk_ais140_gps 
    CHECK (NOT ais140_compliant OR gps_enabled);

ALTER TABLE vehicles DROP CONSTRAINT IF EXISTS chk_blacklist_status;
ALTER TABLE vehicles ADD CONSTRAINT chk_blacklist_status 
    CHECK (NOT is_blacklisted OR status = 'SUSPENDED');

-- Ensure Year is within reasonable bounds
ALTER TABLE vehicles DROP CONSTRAINT IF EXISTS chk_vehicle_year;
ALTER TABLE vehicles ADD CONSTRAINT chk_vehicle_year 
    CHECK (year >= 1900 AND year <= EXTRACT(YEAR FROM CURRENT_DATE) + 1);

-- 2. Refined Document Model
CREATE TABLE IF NOT EXISTS vehicle_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    document_type VARCHAR(50) NOT NULL, -- INSURANCE, PUC, PERMIT, FITNESS
    document_number VARCHAR(100) NOT NULL,
    expiry_date DATE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING', -- PENDING, APPROVED, REJECTED
    verified_at TIMESTAMP WITH TIME ZONE,
    verified_by UUID REFERENCES users(id),
    document_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_vehicle_docs_vehicle ON vehicle_documents(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_docs_tenant ON vehicle_documents(tenant_id);

-- 3. Backfill from legacy vehicle_verifications
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'vehicle_verifications') THEN
        -- Insurance Backfill
        INSERT INTO vehicle_documents (tenant_id, vehicle_id, document_type, document_number, expiry_date, status, created_at, updated_at)
        SELECT tenant_id, vehicle_id, 'INSURANCE', 'LEGACY-' || id::text, insurance_expiry, 
               CASE WHEN status = 'verified' THEN 'APPROVED' ELSE 'PENDING' END,
               created_at, updated_at
        FROM vehicle_verifications
        WHERE insurance_expiry IS NOT NULL;

        -- PUC Backfill
        INSERT INTO vehicle_documents (tenant_id, vehicle_id, document_type, document_number, expiry_date, status, created_at, updated_at)
        SELECT tenant_id, vehicle_id, 'PUC', 'LEGACY-PUC-' || id::text, puc_expiry, 
               CASE WHEN status = 'verified' THEN 'APPROVED' ELSE 'PENDING' END,
               created_at, updated_at
        FROM vehicle_verifications
        WHERE puc_expiry IS NOT NULL;
    END IF;
END $$;

-- 4. Normalize Status & Types
ALTER TABLE vehicles DROP CONSTRAINT IF EXISTS chk_vehicles_status;
ALTER TABLE vehicles ADD CONSTRAINT chk_vehicles_status 
    CHECK (status IN ('ACTIVE', 'MAINTENANCE', 'SUSPENDED', 'DECOMMISSIONED', 'IDLE', 'PROVISIONING', 'RETIRED'));

-- +goose StatementEnd
