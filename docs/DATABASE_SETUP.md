# FlyFleet - Database Setup Guide

This guide covers backing up, restoring, and setting up the database on a new device.

## Quick Start (New Device)

### 1. Prerequisites

```bash
# Install PostgreSQL client (Ubuntu/Debian)
sudo apt install postgresql-client

# Or on Termux
pkg install postgresql

# Install Docker (recommended for full stack)
curl -fsSL https://get.docker.com | sh
```

### 2. Start Infrastructure

```bash
# Copy environment config
cp .env.example .env
# Edit .env with your passwords

# Start PostgreSQL, Redis, NATS
docker-compose up -d postgres redis nats

# Wait for PostgreSQL to be ready
sleep 10
```

### 3. Run Migrations

```bash
# Apply all migrations (creates schema)
./migrate.sh
```

### 4. Seed Default Data

```bash
# Option A: Seed essential bootstrap data (admin user, org, roles)
PGPASSWORD=fleet_password psql -h localhost -p 5433 -U fleet_user -d fleet_db \
    -f infrastructure/migrations/106_seed_essential.sql

# Option B: Restore from backup (if you have one)
./scripts/db-restore.sh backups/flyfleet_YYYYMMDD_HHMMSS.sql.gz
```

### 5. Start Services

```bash
# Start all services
docker-compose up -d

# Or run fleet-core locally
cd services/fleet-core && go run cmd/main.go
```

### 6. Verify

```bash
# Check database tables
PGPASSWORD=fleet_password psql -h localhost -p 5433 -U fleet_user -d fleet_db \
    -c "\dt"

# Check default admin user exists
PGPASSWORD=fleet_password psql -h localhost -p 5433 -U fleet_user -d fleet_db \
    -c "SELECT id, email, name, status FROM users WHERE email = 'admin@flyfleet.local';"

# Login with default admin:
# Email: admin@flyfleet.local
# Password: change_me_123
# ⚠ CHANGE THIS IMMEDIATELY!
```

---

## Database Backup

### Create Backup

```bash
# Automated backup script (excludes telemetry/audit logs for size)
./scripts/db-backup.sh

# Or specify output file
./scripts/db-backup.sh backups/my-backup.sql.gz

# Manual pg_dump command
pg_dump -h localhost -p 5433 -U fleet_user -d fleet_db \
    --clean --if-exists --no-owner --no-privileges \
    --exclude-table-data="*_audit_log" \
    --exclude-table-data="*_telemetry*" | gzip > backups/backup.sql.gz
```

### Docker Backup

```bash
# Backup from running PostgreSQL container
docker exec fleet-postgres pg_dump -U fleet_user -d fleet_db \
    --clean --if-exists --no-owner --no-privileges | gzip > backups/backup.sql.gz
```

### List Backups

```bash
ls -lh backups/
```

---

## Database Restore

### From Backup File

```bash
# Interactive restore (asks for confirmation)
./scripts/db-restore.sh backups/flyfleet_20260411_150000.sql.gz

# Or manual restore
gunzip -c backups/backup.sql.gz | psql -h localhost -p 5433 -U fleet_user -d fleet_db
```

### From Scratch (Fresh Setup)

```bash
# 1. Drop and recreate database
PGPASSWORD=fleet_password psql -h localhost -p 5433 -U fleet_user -d postgres \
    -c "DROP DATABASE IF EXISTS fleet_db;"
PGPASSWORD=fleet_password psql -h localhost -p 5433 -U fleet_user -d postgres \
    -c "CREATE DATABASE fleet_db;"

# 2. Run migrations
./migrate.sh

# 3. Seed data
PGPASSWORD=fleet_password psql -h localhost -p 5433 -U fleet_user -d fleet_db \
    -f infrastructure/migrations/106_seed_essential.sql
```

---

## Transfer to Another Device

### Option 1: Backup + Restore (Recommended)

```bash
# ON SOURCE DEVICE:
./scripts/db-backup.sh backups/latest.sql.gz

# Transfer the backup file (scp, rsync, USB, etc.)
scp backups/latest.sql.gz user@new-device:/path/to/flyfleet/backups/

# ON TARGET DEVICE:
./scripts/db-restore.sh backups/latest.sql.gz
./migrate.sh  # Apply any new migrations
```

### Option 2: Full Project Sync

```bash
# ON SOURCE DEVICE:
# 1. Create backup
./scripts/db-backup.sh

# 2. Sync project files (excluding node_modules, .git, etc.)
rsync -avz --exclude='node_modules' --exclude='.git' --exclude='backups' \
    /path/to/flyfleet/ user@new-device:/path/to/flyfleet/

# ON TARGET DEVICE:
# 1. Install dependencies
cd /path/to/flyfleet
npm install
cd services/fleet-core && go mod download && cd ../..

# 2. Restore database
./scripts/db-restore.sh backups/flyfleet_*.sql.gz

# 3. Start services
docker-compose up -d
```

---

## Troubleshooting

### Connection Refused

```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Check port mapping
docker port fleet-postgres

# Test connection
PGPASSWORD=fleet_password psql -h localhost -p 5433 -U fleet_user -d fleet_db -c "SELECT 1;"
```

### Permission Denied

```bash
# Ensure user has correct permissions
PGPASSWORD=fleet_password psql -h localhost -p 5433 -U fleet_user -d fleet_db \
    -c "\du fleet_user"
```

### Migration Failures

```bash
# Check which migrations have been applied
PGPASSWORD=fleet_password psql -h localhost -p 5433 -U fleet_user -d fleet_db \
    -c "SELECT * FROM schema_migrations ORDER BY version;"

# Re-run migrations (idempotent - safe to re-run)
./migrate.sh
```

### Large Backup/Restore

```bash
# For very large databases, exclude more tables
pg_dump -h localhost -p 5433 -U fleet_user -d fleet_db \
    --clean --if-exists --no-owner --no-privileges \
    --exclude-table-data="*_audit*" \
    --exclude-table-data="*_telemetry*" \
    --exclude-table-data="*_logs" | gzip > backup.sql.gz
```

---

## Database Schema Overview

```
Core Tables:
├── users                    # User accounts
├── organizations            # Fleet organizations
├── user_organizations       # User-org mappings
├── roles                    # RBAC roles
├── user_roles               # User-role assignments
├── vehicles                 # Vehicle registry
├── drivers                  # Driver registry
├── trips                    # Trip management
├── telemetry_*              # GPS/time-series data (TimescaleDB)
├── geofences                # Spatial zones
├── maintenance_records      # Vehicle maintenance
├── invoices                 # Billing/invoicing
├── audit_log                # Audit trail
├── system_config            # System configuration
├── vehicle_types            # Vehicle type catalog
├── fuel_types               # Fuel type catalog
└── notification_templates   # Alert/notification templates
```

---

## Default Credentials

| Role | Email | Password |
|------|-------|----------|
| Super Admin | admin@flyfleet.local | change_me_123 |

⚠ **Change default passwords immediately after setup!**
