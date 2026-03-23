-- Migration: 071_vehicle_enterprise
-- Description: Enterprise-grade Vehicle Asset & Compliance Management Schema
-- Spec: Vehicle Module — fleet-core-service-legacy
-- Note: CREATE INDEX CONCURRENTLY requires running outside a transaction block.
--       Run with psql --single-transaction=off or via a migration tool that supports it.
--
-- ============================================================
-- DOWN MIGRATION (rollback plan)
-- ============================================================
-- DROP VIEW IF EXISTS vehicles_due_for_compliance;
-- DROP TRIGGER IF EXISTS trg_vehicles_gps_uniqueness ON vehicles;
-- DROP FUNCTION IF EXISTS check_gps_device_uniqueness();
-- DROP TABLE IF EXISTS vehicle_events;
-- DROP TABLE IF EXISTS vehicle_documents;
-- DROP TABLE IF EXISTS maintenance_schedule;
-- DROP TABLE IF EXISTS vehicle_financials;
-- DROP TABLE IF EXISTS vehicle_compliance;
-- DROP TABLE IF EXISTS vehicles;
-- ALTER TABLE vehicles_legacy RENAME TO vehicles;   -- restore legacy table
-- ============================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- 1. DROP & RECREATE vehicles (enterprise schema)
-- ============================================================

-- Rename old table to backup, migrate critical data back in
ALTER TABLE vehicles RENAME TO vehicles_legacy;

CREATE TABLE vehicles (
    -- Identity
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id           UUID NOT NULL,                           -- = org_id
    registration_number VARCHAR(50)  NOT NULL,
    chassis_number      TEXT,                                    -- stored AES-256 encrypted at app layer
    engine_number       TEXT,                                    -- stored AES-256 encrypted at app layer
    vin_number          VARCHAR(50),

    -- Classification
    vehicle_type        VARCHAR(30)  NOT NULL DEFAULT 'TRUCK'
                            CHECK (vehicle_type IN ('TRUCK','BUS','VAN','CAR','TRAILER','TANKER','TIPPER','CRANE','OTHER')),
    capacity_tonnage    DECIMAL(10,2),
    fuel_type           VARCHAR(20)  NOT NULL DEFAULT 'DIESEL'
                            CHECK (fuel_type IN ('PETROL','DIESEL','CNG','ELECTRIC','HYBRID','LPG')),
    fuel_tank_capacity  DECIMAL(10,2),
    ownership_type      VARCHAR(20)  NOT NULL DEFAULT 'OWNED'
                            CHECK (ownership_type IN ('OWNED','LEASED','FINANCED','VENDOR')),

    -- Fleet hierarchy
    parent_vehicle_id   UUID REFERENCES vehicles(id) ON DELETE SET NULL,
    branch_id           UUID,
    cost_center_id      UUID,

    -- Status lifecycle
    status              VARCHAR(20)  NOT NULL DEFAULT 'ACTIVE'
                            CHECK (status IN ('ACTIVE','MAINTENANCE','SUSPENDED','DECOMMISSIONED')),

    -- Device binding
    gps_device_id       VARCHAR(100),
    immobilizer_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ais140_compliant    BOOLEAN NOT NULL DEFAULT FALSE,

    -- Metadata migrated from legacy
    make                VARCHAR(100),
    model               VARCHAR(100),
    year                INTEGER,
    color               VARCHAR(50),

    -- Soft-delete & audit
    created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at          TIMESTAMPTZ,
    version             INTEGER NOT NULL DEFAULT 1,              -- optimistic locking

    CONSTRAINT uq_vehicles_tenant_reg UNIQUE (tenant_id, registration_number)
);

-- Migrate data from legacy
INSERT INTO vehicles (
    id, tenant_id, registration_number, vin_number,
    vehicle_type, fuel_type, status,
    make, model, year, color,
    created_at, updated_at, deleted_at
)
SELECT
    id,
    org_id,
    license_plate,
    vin,
    vehicle_type,
    UPPER(COALESCE(fuel_type, 'DIESEL')),
    CASE WHEN status = 'active' THEN 'ACTIVE'
         WHEN status = 'maintenance' THEN 'MAINTENANCE'
         WHEN status = 'retired' THEN 'DECOMMISSIONED'
         ELSE 'SUSPENDED' END,
    make, model, year, color,
    created_at, updated_at, deleted_at
FROM vehicles_legacy
ON CONFLICT (id) DO NOTHING;

-- Indexes
CREATE INDEX CONCURRENTLY idx_vehicles_tenant         ON vehicles(tenant_id);
CREATE INDEX CONCURRENTLY idx_vehicles_tenant_reg     ON vehicles(tenant_id, registration_number);
CREATE INDEX CONCURRENTLY idx_vehicles_status         ON vehicles(status);
CREATE INDEX CONCURRENTLY idx_vehicles_branch         ON vehicles(branch_id);
CREATE INDEX CONCURRENTLY idx_vehicles_parent         ON vehicles(parent_vehicle_id);
CREATE INDEX CONCURRENTLY idx_vehicles_gps_device     ON vehicles(gps_device_id) WHERE gps_device_id IS NOT NULL;
CREATE INDEX CONCURRENTLY idx_vehicles_deleted        ON vehicles(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX CONCURRENTLY idx_vehicles_created        ON vehicles(created_at DESC);
CREATE INDEX CONCURRENTLY idx_vehicles_ais140         ON vehicles(ais140_compliant) WHERE ais140_compliant = TRUE;

-- ============================================================
-- 2. vehicle_compliance
-- ============================================================
CREATE TABLE IF NOT EXISTS vehicle_compliance (
    vehicle_id              UUID PRIMARY KEY REFERENCES vehicles(id) ON DELETE CASCADE,

    -- Expiry dates (India-specific compliance)
    rc_expiry_date          DATE,
    insurance_expiry_date   DATE,
    puc_expiry_date         DATE,
    fitness_expiry_date     DATE,
    permit_expiry_date      DATE,
    permit_type             VARCHAR(50),    -- National, State, etc.
    insurance_policy_no     VARCHAR(100),

    -- Flags
    is_blacklisted          BOOLEAN NOT NULL DEFAULT FALSE,
    legal_hold              BOOLEAN NOT NULL DEFAULT FALSE,

    -- Verification
    last_verified_at        TIMESTAMPTZ,
    verified_by             UUID,

    created_at              TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Compound index for expiry alerting
CREATE INDEX CONCURRENTLY idx_compliance_rc_expiry          ON vehicle_compliance(rc_expiry_date)          WHERE rc_expiry_date IS NOT NULL;
CREATE INDEX CONCURRENTLY idx_compliance_insurance_expiry   ON vehicle_compliance(insurance_expiry_date)   WHERE insurance_expiry_date IS NOT NULL;
CREATE INDEX CONCURRENTLY idx_compliance_puc_expiry         ON vehicle_compliance(puc_expiry_date)         WHERE puc_expiry_date IS NOT NULL;
CREATE INDEX CONCURRENTLY idx_compliance_fitness_expiry     ON vehicle_compliance(fitness_expiry_date)     WHERE fitness_expiry_date IS NOT NULL;
CREATE INDEX CONCURRENTLY idx_compliance_permit_expiry      ON vehicle_compliance(permit_expiry_date)      WHERE permit_expiry_date IS NOT NULL;
CREATE INDEX CONCURRENTLY idx_compliance_blacklisted        ON vehicle_compliance(is_blacklisted)          WHERE is_blacklisted = TRUE;
CREATE INDEX CONCURRENTLY idx_compliance_legal_hold         ON vehicle_compliance(legal_hold)              WHERE legal_hold = TRUE;

-- ============================================================
-- 3. vehicle_financials
-- ============================================================
CREATE TABLE IF NOT EXISTS vehicle_financials (
    vehicle_id              UUID PRIMARY KEY REFERENCES vehicles(id) ON DELETE CASCADE,

    acquisition_cost        DECIMAL(15,2),
    depreciation_method     VARCHAR(20)  CHECK (depreciation_method IN ('STRAIGHT_LINE','DECLINING_BALANCE','UNITS_OF_PRODUCTION')),
    depreciation_rate       DECIMAL(5,2),
    residual_value          DECIMAL(15,2),
    emi_amount              DECIMAL(15,2),
    lease_end_date          DATE,
    insurance_premium       DECIMAL(15,2),

    -- Running cost aggregates (updated by downstream services)
    total_maintenance_cost  DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_fuel_cost         DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_toll_cost         DECIMAL(15,2) NOT NULL DEFAULT 0,

    -- Total Cost of Ownership (cached)
    tco_cached              DECIMAL(15,2),
    tco_calculated_at       TIMESTAMPTZ,

    created_at              TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- 4. maintenance_schedule
-- ============================================================
CREATE TABLE IF NOT EXISTS maintenance_schedule (
    vehicle_id              UUID PRIMARY KEY REFERENCES vehicles(id) ON DELETE CASCADE,

    next_service_date       DATE,
    next_service_odometer   DECIMAL(10,2),
    engine_hours            DECIMAL(10,2),
    predictive_score        DECIMAL(5,2),     -- 0-100, higher = more urgent
    last_service_date       DATE,

    created_at              TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX CONCURRENTLY idx_maint_next_date    ON maintenance_schedule(next_service_date)     WHERE next_service_date IS NOT NULL;
CREATE INDEX CONCURRENTLY idx_maint_pred_score   ON maintenance_schedule(predictive_score DESC) WHERE predictive_score IS NOT NULL;

-- ============================================================
-- 5. vehicle_documents (immutable)
-- ============================================================
CREATE TABLE IF NOT EXISTS vehicle_documents (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id      UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    document_type   VARCHAR(30) NOT NULL
                        CHECK (document_type IN ('RC','INSURANCE','PERMIT','FITNESS','PUC','OTHER')),
    file_url        TEXT NOT NULL,
    version         INTEGER NOT NULL DEFAULT 1,
    hash            VARCHAR(64) NOT NULL,   -- SHA-256 hex
    uploaded_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    uploaded_by     UUID
    -- NOTE: No UPDATE allowed on this table (enforced at app layer)
);

CREATE INDEX CONCURRENTLY idx_vdoc_vehicle          ON vehicle_documents(vehicle_id);
CREATE INDEX CONCURRENTLY idx_vdoc_type             ON vehicle_documents(document_type);
CREATE INDEX CONCURRENTLY idx_vdoc_uploaded         ON vehicle_documents(uploaded_at DESC);

-- ============================================================
-- 6. vehicle_events (append-only, hash-chained audit)
-- ============================================================
CREATE TABLE IF NOT EXISTS vehicle_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id      UUID NOT NULL REFERENCES vehicles(id) ON DELETE RESTRICT,
    event_type      VARCHAR(50) NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}',
    hash            VARCHAR(64) NOT NULL,   -- SHA-256 of (vehicle_id || event_type || payload || prev_hash)
    prev_hash       VARCHAR(64),            -- NULL for first event per vehicle
    actor_id        UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
    -- NOTE: No UPDATE or DELETE allowed on this table (append-only)
);

CREATE INDEX CONCURRENTLY idx_vevt_vehicle   ON vehicle_events(vehicle_id);
CREATE INDEX CONCURRENTLY idx_vevt_type      ON vehicle_events(event_type);
CREATE INDEX CONCURRENTLY idx_vevt_created   ON vehicle_events(created_at DESC);

-- Row-level security helper: prevent updates/deletes
CREATE OR REPLACE RULE vehicle_events_no_update AS ON UPDATE TO vehicle_events DO INSTEAD NOTHING;
CREATE OR REPLACE RULE vehicle_events_no_delete AS ON DELETE TO vehicle_events DO INSTEAD NOTHING;

-- ============================================================
-- 7. Compliance expiry view
-- ============================================================
CREATE OR REPLACE VIEW vehicles_due_for_compliance AS
SELECT
    v.id            AS vehicle_id,
    v.tenant_id,
    v.registration_number,
    v.status,
    c.rc_expiry_date,
    c.insurance_expiry_date,
    c.puc_expiry_date,
    c.fitness_expiry_date,
    c.permit_expiry_date,
    LEAST(
        c.rc_expiry_date,
        c.insurance_expiry_date,
        c.puc_expiry_date,
        c.fitness_expiry_date,
        c.permit_expiry_date
    ) AS earliest_expiry,
    (LEAST(
        c.rc_expiry_date,
        c.insurance_expiry_date,
        c.puc_expiry_date,
        c.fitness_expiry_date,
        c.permit_expiry_date
    ) - CURRENT_DATE) AS days_to_expiry
FROM vehicles v
JOIN vehicle_compliance c ON c.vehicle_id = v.id
WHERE v.deleted_at IS NULL
  AND v.status != 'DECOMMISSIONED'
  AND LEAST(
        c.rc_expiry_date,
        c.insurance_expiry_date,
        c.puc_expiry_date,
        c.fitness_expiry_date,
        c.permit_expiry_date
  ) IS NOT NULL
ORDER BY earliest_expiry ASC;

-- ============================================================
-- 8. GPS device uniqueness function
-- ============================================================
CREATE OR REPLACE FUNCTION check_gps_device_uniqueness()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.gps_device_id IS NOT NULL THEN
        IF EXISTS (
            SELECT 1 FROM vehicles
            WHERE gps_device_id = NEW.gps_device_id
              AND id != NEW.id
              AND deleted_at IS NULL
        ) THEN
            RAISE EXCEPTION 'GPS device % is already bound to another vehicle', NEW.gps_device_id;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_vehicles_gps_uniqueness
BEFORE INSERT OR UPDATE ON vehicles
FOR EACH ROW EXECUTE FUNCTION check_gps_device_uniqueness();

-- ============================================================
-- COMMENTS
-- ============================================================
COMMENT ON TABLE vehicles                IS 'Vehicle aggregate root — enterprise master';
COMMENT ON TABLE vehicle_compliance      IS 'India-regulatory compliance dates and flags (RC, Insurance, PUC, Fitness, Permit)';
COMMENT ON TABLE vehicle_financials      IS 'Acquisition cost, depreciation, EMI, lease, and TCO aggregates';
COMMENT ON TABLE maintenance_schedule    IS 'Predictive maintenance schedule per vehicle';
COMMENT ON TABLE vehicle_documents       IS 'Immutable document store (RC, Insurance, Permit, Fitness scans)';
COMMENT ON TABLE vehicle_events          IS 'Append-only, hash-chained audit event log';
COMMENT ON COLUMN vehicles.version       IS 'Optimistic concurrency control version counter';
COMMENT ON COLUMN vehicles.chassis_number IS 'AES-256 encrypted at application layer';
COMMENT ON COLUMN vehicles.engine_number  IS 'AES-256 encrypted at application layer';
