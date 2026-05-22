-- Repair any existing NULL values in the is_deleted column to FALSE
UPDATE trips SET is_deleted = FALSE WHERE is_deleted IS NULL;
