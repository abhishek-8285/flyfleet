-- 063_retention_cron.sql

-- Creates a stored procedure to drop partitions older than 90 days.
-- This can be executed via pg_cron extension or a Go background worker.

CREATE OR REPLACE FUNCTION drop_old_telemetry_partitions(retention_days INT DEFAULT 90)
RETURNS VOID AS $$
DECLARE
    partition_record RECORD;
    partition_date DATE;
    cutoff_date DATE;
BEGIN
    cutoff_date := CURRENT_DATE - retention_days;

    FOR partition_record IN
        SELECT
            c.relname AS partition_name
        FROM
            pg_class c
        JOIN
            pg_inherits i ON c.oid = i.inhrelid
        JOIN
            pg_class p ON i.inhparent = p.oid
        WHERE
            p.relname = 'telemetry_events'
    LOOP
        -- Extract the date part from the partition name (assuming format: telemetry_events_pYYYY_MM)
        -- Example: telemetry_events_p2023_01
        
        BEGIN
            -- Using regex to extract YYYY_MM
            partition_date := to_date(substring(partition_record.partition_name from 'p(\d{4}_\d{2})'), 'YYYY_MM');
            
            -- Because the partition name represents the START of the month, 
            -- we want to drop it if the ENTIRE month is safely past the cutoff date.
            -- Adding 1 month to the start date gives us the upper bound of the partition.
            IF (partition_date + interval '1 month') < cutoff_date THEN
                EXECUTE format('DROP TABLE IF EXISTS %I;', partition_record.partition_name);
                RAISE NOTICE 'Dropped old partition: %', partition_record.partition_name;
            END IF;
        EXCEPTION WHEN OTHERS THEN
            -- Ignore parsing errors for malformed names, just log it if necessary
            RAISE NOTICE 'Skipping unrecognized partition name: %', partition_record.partition_name;
        END;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Optional: If pg_cron is installed, we can schedule it automatically:
-- SELECT cron.schedule('drop_telemetry_partitions', '0 3 * * *', 'SELECT drop_old_telemetry_partitions(90);');
