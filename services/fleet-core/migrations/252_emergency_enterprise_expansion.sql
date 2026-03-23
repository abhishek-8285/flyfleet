-- Migration: Phase 3 - Offline SOS Event Queue
CREATE TABLE IF NOT EXISTS device_event_queue (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id   UUID NOT NULL,
    event_type   TEXT NOT NULL DEFAULT 'SOS',
    payload      JSONB NOT NULL DEFAULT '{}',
    retry_count  INT NOT NULL DEFAULT 0,
    status       TEXT NOT NULL DEFAULT 'PENDING',  -- PENDING, DELIVERED, FAILED
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    next_retry_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_device_event_queue_vehicle ON device_event_queue(vehicle_id);
CREATE INDEX idx_device_event_queue_status  ON device_event_queue(status, next_retry_at);

-- Migration: Phase 5 - Geo-Based Authority Routing
CREATE TABLE IF NOT EXISTS emergency_authority_zones (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    authority_type TEXT NOT NULL,         -- POLICE, AMBULANCE, FIRE
    zone_name      TEXT NOT NULL,
    contact_number TEXT NOT NULL,
    api_endpoint   TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed with representative authority zones (bounding boxes instead of full PostGIS polygons
-- to remain compatible without the PostGIS extension for now)
INSERT INTO emergency_authority_zones (authority_type, zone_name, contact_number, api_endpoint)
VALUES
  ('POLICE',    'Delhi Police - South District',     '100', 'https://erp.delhi.police.gov.in/api'),
  ('AMBULANCE', 'CATS Ambulance - Delhi',             '102', 'https://cats.nic.in/dispatch/api'),
  ('FIRE',      'Delhi Fire Services',                '101', 'https://fire.delhi.gov.in/api'),
  ('POLICE',    'Mumbai Police - Dharavi',            '100', 'https://erp.mumbaipolice.gov.in/api'),
  ('AMBULANCE', '108 Ambulance - Maharashtra',        '108', 'https://108ambulance.mhvss.in/api'),
  ('FIRE',      'Mumbai Fire Brigade',                '101', 'https://fire.mcgm.gov.in/api')
ON CONFLICT DO NOTHING;

-- Migration: Phase 6 - Incident Dashboard Metrics View
CREATE OR REPLACE VIEW emergency_dashboard_metrics AS
SELECT
    DATE(created_at)                                                    AS day,
    COUNT(*)                                                            AS total_incidents,
    COUNT(*) FILTER (WHERE status = 'RESOLVED')                        AS resolved_incidents,
    COUNT(*) FILTER (WHERE status = 'TRIGGERED')                       AS open_incidents,
    COUNT(*) FILTER (WHERE severity = 'CRITICAL')                      AS critical_count,
    COUNT(*) FILTER (WHERE simulation_mode = true)                     AS drill_count,
    ROUND(AVG(
        EXTRACT(EPOCH FROM (resolved_at - created_at))
    ) FILTER (WHERE resolved_at IS NOT NULL)::NUMERIC, 2)              AS avg_response_time_seconds
FROM sos_incidents
GROUP BY DATE(created_at)
ORDER BY day DESC;
