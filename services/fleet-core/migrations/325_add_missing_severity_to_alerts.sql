-- Migration: 325_add_missing_severity_to_alerts
-- Description: Ensures the 'severity' column exists in 'alerts' and 'sos_incidents' tables.

DO $$ 
BEGIN 
    -- 1. Patch 'alerts' table
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'alerts') THEN
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'alerts' AND column_name = 'severity') THEN
            -- Check if 'priority' exists instead (some legacy versions used priority)
            IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'alerts' AND column_name = 'priority') THEN
                ALTER TABLE alerts RENAME COLUMN priority TO severity;
            ELSE
                ALTER TABLE alerts ADD COLUMN severity VARCHAR(20) DEFAULT 'warning';
            END IF;
        END IF;
    END IF;

    -- 2. Patch 'sos_incidents' table
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'sos_incidents') THEN
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'sos_incidents' AND column_name = 'severity') THEN
             IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'sos_incidents' AND column_name = 'priority') THEN
                ALTER TABLE sos_incidents RENAME COLUMN priority TO severity;
            ELSE
                ALTER TABLE sos_incidents ADD COLUMN severity VARCHAR(20) DEFAULT 'CRITICAL';
            END IF;
        END IF;
    END IF;
END $$;
