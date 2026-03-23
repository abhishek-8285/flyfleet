-- Migration: 072_tracking_enterprise
-- Description: Enterprise GPS Tracking—telemetry_points (PostGIS-partitioned),
--              device_registry, and tracking dedup helper
-- Note: CREATE INDEX CONCURRENTLY requires running outside a transaction block.
--       Run with psql --single-transaction=off or via a migration tool that supports it.
--
-- ============================================================
-- DOWN MIGRATION (rollback plan)
-- ============================================================
-- DROP VIEW IF EXISTS v_live_locations;
-- DROP TABLE IF EXISTS device_commands;
-- DROP TABLE IF EXISTS device_registry;
-- DROP TABLE IF EXISTS telemetry_points_y2026m01;   -- repeat for each partition
-- DROP TABLE IF EXISTS telemetry_points;
-- ============================================================

-- Required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS postgis;

-- ============================================================
-- 1. telemetry_points  (monthly range-partitioned by recorded_at)
-- ============================================================

CREATE TABLE IF NOT EXISTS telemetry_points (
    id            BIGSERIAL,
    event_id      UUID         NOT NULL,    -- from device; unique for dedup
    vehicle_id    UUID         NOT NULL,
    device_id     VARCHAR(100) NOT NULL,
    tenant_id     UUID,
    lat           DOUBLE PRECISION NOT NULL,
    lng           DOUBLE PRECISION NOT NULL,
    speed         REAL         NOT NULL DEFAULT 0,   -- km/h
    heading       SMALLINT     NOT NULL DEFAULT 0,   -- degrees 0-360
    ignition      BOOLEAN      NOT NULL DEFAULT FALSE,
    fuel_level    REAL,                              -- %
    odometer      DOUBLE PRECISION,                  -- km
    event_type    VARCHAR(20)  NOT NULL DEFAULT 'GPS',
    sequence_no   BIGINT       NOT NULL DEFAULT 0,
    recorded_at   TIMESTAMPTZ  NOT NULL,
    received_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    geo           GEOMETRY(Point, 4326),             -- PostGIS point

    PRIMARY KEY (id, recorded_at)
) PARTITION BY RANGE (recorded_at);

-- Unique dedup constraint per partition is enforced at app layer via event_id check
-- (global unique across partitions via partial index pattern below)

-- Seed partitions for 2026
CREATE TABLE IF NOT EXISTS telemetry_points_y2026m01
    PARTITION OF telemetry_points
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

CREATE TABLE IF NOT EXISTS telemetry_points_y2026m02
    PARTITION OF telemetry_points
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');

CREATE TABLE IF NOT EXISTS telemetry_points_y2026m03
    PARTITION OF telemetry_points
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');

CREATE TABLE IF NOT EXISTS telemetry_points_y2026m04
    PARTITION OF telemetry_points
    FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');

CREATE TABLE IF NOT EXISTS telemetry_points_y2026m05
    PARTITION OF telemetry_points
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');

CREATE TABLE IF NOT EXISTS telemetry_points_y2026m06
    PARTITION OF telemetry_points
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');

CREATE TABLE IF NOT EXISTS telemetry_points_y2026m07
    PARTITION OF telemetry_points
    FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');

CREATE TABLE IF NOT EXISTS telemetry_points_y2026m08
    PARTITION OF telemetry_points
    FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');

CREATE TABLE IF NOT EXISTS telemetry_points_y2026m09
    PARTITION OF telemetry_points
    FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');

CREATE TABLE IF NOT EXISTS telemetry_points_y2026m10
    PARTITION OF telemetry_points
    FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');

CREATE TABLE IF NOT EXISTS telemetry_points_y2026m11
    PARTITION OF telemetry_points
    FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');

CREATE TABLE IF NOT EXISTS telemetry_points_y2026m12
    PARTITION OF telemetry_points
    FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');

-- Future year
CREATE TABLE IF NOT EXISTS telemetry_points_y2027m01
    PARTITION OF telemetry_points
    FOR VALUES FROM ('2027-01-01') TO ('2027-02-01');

-- ── Indexes ─────────────────────────────────────────────────────────────────
-- Hot-path query: latest points for a vehicle
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tp_vehicle_time
    ON telemetry_points (vehicle_id, recorded_at DESC);

-- BRIN for time-range scans (very cheap on monotonically increasing data)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tp_brin_recorded
    ON telemetry_points USING BRIN (recorded_at);

-- PostGIS geo index for nearby queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tp_geo
    ON telemetry_points USING GIST (geo);

-- Dedup index on event_id (non-unique; uniqueness enforced at app layer per perf constraints)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tp_event_id
    ON telemetry_points (event_id);

-- Event-type filter index (anomaly queries)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tp_event_type
    ON telemetry_points (event_type) WHERE event_type != 'GPS';

-- ── Trigger: auto-populate geo column ────────────────────────────────────────
CREATE OR REPLACE FUNCTION tp_set_geo()
RETURNS TRIGGER AS $$
BEGIN
    NEW.geo := ST_SetSRID(ST_MakePoint(NEW.lng, NEW.lat), 4326);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_tp_set_geo
BEFORE INSERT OR UPDATE ON telemetry_points
FOR EACH ROW EXECUTE FUNCTION tp_set_geo();

-- ============================================================
-- 2. device_registry  (control plane)
-- ============================================================

CREATE TABLE IF NOT EXISTS device_registry (
    device_id         VARCHAR(100) PRIMARY KEY,
    vehicle_id        UUID,
    firmware_version  VARCHAR(50),
    sim_iccid         VARCHAR(30),
    carrier           VARCHAR(50),
    last_heartbeat    TIMESTAMPTZ,
    status            VARCHAR(20) NOT NULL DEFAULT 'active'
                          CHECK (status IN ('active','offline','decommissioned')),
    certificate_hash  VARCHAR(64),      -- SHA-256 of device certificate
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_dreg_vehicle
    ON device_registry (vehicle_id) WHERE vehicle_id IS NOT NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_dreg_heartbeat
    ON device_registry (last_heartbeat DESC) WHERE status = 'active';

-- ============================================================
-- 3. device_commands  (if not already present from migration 044)
-- ============================================================

CREATE TABLE IF NOT EXISTS device_commands (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id   VARCHAR(100) NOT NULL REFERENCES device_registry(device_id) ON DELETE CASCADE,
    type        VARCHAR(30) NOT NULL
                    CHECK (type IN ('ping','reboot','immobilize','config_update')),
    params      JSONB       NOT NULL DEFAULT '{}',
    status      VARCHAR(20) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','executed','expired','failed')),
    issued_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL,
    executed_at TIMESTAMPTZ,
    result      TEXT,
    issued_by   UUID        -- operator UUID
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_dcmd_device_pending
    ON device_commands (device_id, issued_at ASC)
    WHERE status = 'pending';

-- ============================================================
-- 4. Live-location convenience view
-- ============================================================

CREATE OR REPLACE VIEW v_live_locations AS
SELECT DISTINCT ON (vehicle_id)
    vehicle_id,
    device_id,
    lat,
    lng,
    speed,
    heading,
    ignition,
    recorded_at AS last_seen
FROM telemetry_points
WHERE received_at > NOW() - INTERVAL '10 minutes'
ORDER BY vehicle_id, recorded_at DESC;

-- ============================================================
-- Comments
-- ============================================================
COMMENT ON TABLE telemetry_points IS 'Monthly-partitioned AIS-140 telemetry — hot store (30-90 day retention)';
COMMENT ON TABLE device_registry  IS 'IoT device control plane: firmware, SIM, certificate, health';
COMMENT ON TABLE device_commands  IS 'Outbound command queue for device control channel';
COMMENT ON VIEW  v_live_locations IS 'Latest GPS point per vehicle within last 10 minutes';
