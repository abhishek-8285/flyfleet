# Admin User Creation Guide

This directory contains scripts to create an admin (super_admin) user for FlyFleet.

## Prerequisites

- PostgreSQL database is running (default: localhost:5433)
- Database initialized with migrations (run `docker-compose up -d postgres` and then migrations)
- Go installed (for `.go` script) OR Docker (for `.sh` script)

## Options

### Option 1: Go Script (Recommended - Proper Argon2id Hashing)

This script creates a user with a properly Argon2id-hashed password that the application can verify.

```bash
# Using default credentials (from .env or defaults)
go run scripts/create_admin.go

# Or build and run:
go build -o create_admin scripts/create_admin.go
./create_admin

# With custom values:
ADMIN_EMAIL=user@example.com ADMIN_PASSWORD=SecurePass123 go run scripts/create_admin.go

# Or using command-line args:
go run scripts/create_admin.go user@example.com SecurePass123 "User Name"
```

The script will:
1. Connect to the database
2. Create "Default Organization" if it doesn't exist
3. Create a `super_admin` user with full permissions (permission_mask = max int64)
4. Hash the password using Argon2id (same as application)

**Note:** If the user already exists, it will upgrade their role to `super_admin`.

### Option 2: Bash + Docker Exec (Quick)

If you have Docker running with the `fleet-postgres` container:

```bash
./scripts/create_admin.sh [email] [password] [name]
```

Example:
```bash
./scripts/create_admin.sh admin@flyfleet.com MySecurePassword123 "Admin User"
```

**Note:** This creates a user with a placeholder password hash. You must reset the password via the application's password reset API afterward.

### Option 3: Via Application API (Most Secure)

Start the application and use the registration endpoint:

```bash
# Start the service
docker-compose up -d fleet-core

# Register via API (creates driver by default, or org_admin if providing org_name)
curl -X POST http://localhost:8084/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@flyfleet.com",
    "password": "YourSecurePassword",
    "name": "Admin User",
    "org_name": "Default Organization"
  }'
```

This creates an `org_admin` user (not super_admin). To promote to super_admin, either:
- Use the admin user management UI (if accessible)
- Or update directly in DB: `UPDATE users SET role = 'super_admin', permission_mask = 9223372036854775807 WHERE email = 'admin@flyfleet.com';`

## Default Values

- **Database:** `localhost:5433`, user `fleet_user`, db `fleet_db`
- **Email:** `admin@flyfleet.com`
- **Password:** `Admin123!` (change immediately!)
- **Name:** `FlyFleet Admin`

## Verification

After creating the user, verify it exists:

```bash
# Using Docker
docker exec fleet-postgres psql -U fleet_user -d fleet_db \
  -c "SELECT id, email, role, org_id FROM users WHERE email = 'admin@flyfleet.com';"

# Or using psql directly
PGPASSWORD=fleet_password psql -h localhost -p 5433 -U fleet_user -d fleet_db \
  -c "SELECT id, email, role, org_id FROM users WHERE email = 'admin@flyfleet.com';"
```

## Next Steps

1. **Start the application:**
   ```bash
   docker-compose up -d
   ```

2. **Access the dashboard:** http://localhost:3000 (or the appropriate URL)

3. **Login** with the credentials you created

4. **If MFA is enforced**, you may need to set up MFA on first login.

## Security Notes

- Change the default password immediately after first login
- Enable MFA for admin accounts
- Keep `JWT_SECRET` strong and unique in production
- The `permission_mask` of `9223372036854775807` gives all permissions (bitmask all-ones)
- Super admin role bypasses most permission checks

## Troubleshooting

**"connection refused"**
- Ensure PostgreSQL is running: `docker ps | grep postgres`
- Check port mapping: `docker port fleet-postgres 5432` should show `0.0.0.0:5433`

**"database does not exist"**
- Run migrations: `docker-compose up -d postgres && go run ./cmd/migrate`

**"user already exists"**
- The script will upgrade existing user to super_admin role
- To reset password, use the password reset API: `POST /api/v1/auth/password/reset/request`

**"password doesn't work"**
- Ensure you used the Go script (Option 1) for proper Argon2id hashing
- If using the bash script (Option 2), reset password via API

## File Descriptions

- `create_admin.go` - Go program that creates user with proper Argon2id password hash
- `create_admin.sh` - Bash wrapper using Docker exec (inserts placeholder hash)
- `README.md` - This file
