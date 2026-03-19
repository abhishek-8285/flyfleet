ALTER TABLE users ADD COLUMN IF NOT EXISTS created_by uuid;
ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_by uuid;
ALTER TABLE users OWNER TO fleet_user;
