-- Change default value of trips.status to 'CREATED'
ALTER TABLE trips ALTER COLUMN status SET DEFAULT 'CREATED';

-- Standardise any existing lowercase statuses in the trips table to their uppercase equivalents
UPDATE trips SET status = 'CREATED' WHERE status IN ('scheduled', 'created');
UPDATE trips SET status = 'STARTED' WHERE status IN ('in_progress', 'started', 'active', 'ACTIVE');
UPDATE trips SET status = 'COMPLETED' WHERE status = 'completed';
UPDATE trips SET status = 'CANCELLED' WHERE status = 'cancelled';

-- Populate coordinates for seeded active trip to prevent "Unknown Location"
UPDATE trips 
SET pickup_lat = 28.6139, pickup_lng = 77.2090, drop_lat = 26.9124, drop_lng = 75.7873 
WHERE id = '550e8400-e29b-41d4-a716-446655440020';

