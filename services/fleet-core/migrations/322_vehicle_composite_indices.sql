-- +goose Up
-- +goose StatementBegin
-- Migration: 322_vehicle_composite_indices.sql
-- Purpose: Optimize multi-tenant queries and enforce strict data isolation.

-- 1. Composite Indices for High-Frequency Filtering
CREATE INDEX IF NOT EXISTS idx_vehicles_tenant_status ON vehicles (tenant_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vehicles_tenant_branch ON vehicles (tenant_id, branch_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vehicles_tenant_type ON vehicles (tenant_id, vehicle_type) WHERE deleted_at IS NULL;

-- 2. Telemetry Optimization
CREATE INDEX IF NOT EXISTS idx_vehicles_gps_active ON vehicles (tenant_id, gps_enabled) WHERE deleted_at IS NULL AND gps_enabled = TRUE;

-- 3. Registration Uniqueness Hardening
-- Enforce global uniqueness for active registration numbers at the database level.
-- (Already handled by idx_unique_registration_active in 321, but adding a descriptive comment/documentation here)

-- +goose StatementEnd
