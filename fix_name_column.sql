-- Rename full_name to name in users table to match repository implementation
DO $$ 
BEGIN 
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='full_name') THEN
        ALTER TABLE users RENAME COLUMN full_name TO name;
    END IF;
END $$;
