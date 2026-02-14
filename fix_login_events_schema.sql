-- Fix login_events table schema to match repository expectations
ALTER TABLE login_events ADD COLUMN IF NOT EXISTS email VARCHAR(255);
ALTER TABLE login_events ADD COLUMN IF NOT EXISTS device VARCHAR(100);
ALTER TABLE login_events ADD COLUMN IF NOT EXISTS location VARCHAR(100);
ALTER TABLE login_events ADD COLUMN IF NOT EXISTS risk_level VARCHAR(20);
ALTER TABLE login_events ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- Support older schema columns if needed, but repository expects created_at
-- If timestamp exists, we can keep it for now.
DO $$ 
BEGIN 
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='login_events' AND column_name='timestamp') THEN
        -- Optionally migrate timestamp to created_at if created_at was just added
        UPDATE login_events SET created_at = "timestamp" WHERE created_at IS NULL;
    END IF;
END $$;
