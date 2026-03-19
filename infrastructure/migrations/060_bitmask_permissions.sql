-- 060_bitmask_permissions.sql

-- Step 1: Add the BIGINT permission_mask to the users table.
-- Postgres BIGINT is 8 bytes, which holds a 64-bit integer, perfect for checking bitmasks.
ALTER TABLE users ADD COLUMN IF NOT EXISTS permission_mask BIGINT DEFAULT 0;

-- Optional Step 2: Seed existing users depending on their old string-based roles
-- 1 = PermReadVehicles, 2 = PermWriteVehicles, 4 = PermReadCosts, 8 = PermDeleteTrips, etc.
-- For standard 'org_admin', let's say they get a bunch of bits up to 255.
-- Since this is a new feature, we default to 0 and let the app handle JIT seeding or
-- just leave it 0 until permission UI is fully built.

-- For safety, update super_admins to have full bits (-1 or ~0 in bitwise, which is max bigint but safe to keep as high)
-- Or explicitly assign some bits. Let's just set it to a very high integer for super_admin:
-- Max BIGINT is 9223372036854775807 (which means all 63 positive bits are 1).
UPDATE users SET permission_mask = 9223372036854775807 WHERE role = 'super_admin';
