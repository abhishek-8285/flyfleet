UPDATE users SET created_by = '00000000-0000-0000-0000-000000000000'::uuid WHERE created_by IS NULL;
UPDATE users SET updated_by = '00000000-0000-0000-0000-000000000000'::uuid WHERE updated_by IS NULL;
UPDATE users SET updated_at = now() WHERE updated_at IS NULL;
UPDATE users SET mfa_methods = '{}' WHERE mfa_methods IS NULL;
UPDATE users SET permissions = '{}' WHERE permissions IS NULL;
UPDATE users SET last_password_change = now() WHERE last_password_change IS NULL;
UPDATE users SET created_at = now() WHERE created_at IS NULL;

ALTER TABLE users ALTER COLUMN created_by SET DEFAULT '00000000-0000-0000-0000-000000000000'::uuid;
ALTER TABLE users ALTER COLUMN updated_by SET DEFAULT '00000000-0000-0000-0000-000000000000'::uuid;
ALTER TABLE users ALTER COLUMN updated_at SET DEFAULT now();
ALTER TABLE users ALTER COLUMN mfa_methods SET DEFAULT '{}';
ALTER TABLE users ALTER COLUMN permissions SET DEFAULT '{}';
ALTER TABLE users ALTER COLUMN last_password_change SET DEFAULT now();
