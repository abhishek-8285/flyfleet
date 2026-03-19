-- 062_telemetry_partitioning.sql

-- 1. Rename existing table to retain data
DO $$ 
BEGIN
  IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'telemetry_events') THEN
    ALTER TABLE telemetry_events RENAME TO telemetry_events_legacy;
  END IF;
END $$;

-- 2. Create the new native Declarative Time-Partitioned table
CREATE TABLE telemetry_events (
    id UUID NOT NULL,
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    speed REAL,
    heading REAL,
    fuel_level REAL,
    engine_rpm INT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
) PARTITION BY RANGE (created_at);

-- 3. Create indices on the partitioned table (indices are propagated to partitions)
CREATE INDEX idx_telemetry_org_id_created_at ON telemetry_events (org_id, created_at DESC);
CREATE INDEX idx_telemetry_vehicle_id_created_at ON telemetry_events (vehicle_id, created_at DESC);

-- 4. Enable RLS on the new table
ALTER TABLE telemetry_events ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON telemetry_events
    FOR ALL
    USING (org_id = get_current_org_id());

-- 5. Auto-create the partitions for the next 3 months (and current month)
DO $$
DECLARE
    start_date DATE := date_trunc('month', CURRENT_DATE);
    end_date DATE;
    partition_name TEXT;
    i INT;
BEGIN
    FOR i IN 0..3 LOOP
        start_date := date_trunc('month', CURRENT_DATE) + (i || ' month')::interval;
        end_date := start_date + '1 month'::interval;
        partition_name := 'telemetry_events_p' || to_char(start_date, 'YYYY_MM');

        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF telemetry_events FOR VALUES FROM (%L) TO (%L);',
            partition_name, start_date, end_date
        );
    END LOOP;
END $$;

-- Note: We do *not* drop telemetry_events_legacy here to avoid data loss. 
-- A background migration script should chunk-insert legacy data into the new partitioned table.
