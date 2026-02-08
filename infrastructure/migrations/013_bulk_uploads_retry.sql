-- Add retry_count to bulk_jobs table
ALTER TABLE bulk_jobs ADD COLUMN retry_count INT DEFAULT 0;
