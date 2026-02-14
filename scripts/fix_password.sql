-- Update Delta Admin Password to match Super Admin (test123)
-- This allows us to login as admin@delta.io

UPDATE users 
SET password_hash = (SELECT password_hash FROM users WHERE email = 'admin@test.com')
WHERE email = 'admin@delta.io';
