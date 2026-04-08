# 🚀 Deployment & Schema Evolution Analysis

**Author**: Lead System Architect  
**Date**: March 15, 2026  
**Scope**: Expand & contract pattern, migration guardrails, domain kill switch

---

## Executive Summary

This document provides a comprehensive analysis of FlyFleet's deployment and schema evolution strategy, covering:

1. **Expand & Contract Pattern**: Handling breaking changes during rolling Kubernetes updates
2. **Migration Guardrails**: Preventing long-running locks with 130+ migrations
3. **Domain Kill Switch**: Hot-disabling NATS consumers without pod restart

---

## 1️⃣ The Expand & Contract Pattern: Breaking Change Management

### Current Deployment Architecture

**Location**: `infrastructure/k8s/deployment.yaml`

```yaml
# CURRENT: Standard rolling update (no expand/contract support)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: flyfleet-api-gateway
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1       # Allow 1 extra pod during update
      maxUnavailable: 0 # Zero downtime required
  template:
    spec:
      containers:
      - name: api-gateway
        image: flyfleet/api-gateway:latest
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
```

### ⚠️ CRITICAL GAP: No Expand & Contract Support for Schema Changes

**Problem Scenario**: Breaking change to `vehicles` table

```sql
-- OLD SCHEMA (v1.0)
CREATE TABLE vehicles (
    id UUID PRIMARY KEY,
    license_plate VARCHAR(20),
    model VARCHAR(50),
    make VARCHAR(50),
    year INTEGER,
    status VARCHAR(20)  -- 'active', 'inactive', 'maintenance'
);

-- NEW SCHEMA (v2.0) - BREAKING CHANGE
ALTER TABLE vehicles 
ADD COLUMN vehicle_class VARCHAR(20) NOT NULL DEFAULT 'class_a',
ADD COLUMN emission_standard VARCHAR(20),
ALTER COLUMN status TYPE VARCHAR(30);  -- Expanding enum
```

**Rolling Update Problem**:
```
T0: Deployment starts (3 replicas)
    → Pod-1: v1.0 code, expects OLD schema
    → Pod-2: v1.0 code, expects OLD schema
    → Pod-3: v1.0 code, expects OLD schema

T1: Migration applied (NEW schema)
    → vehicles.vehicle_class column added
    → vehicles.status type changed

T2: Rolling update begins
    → Pod-4 (v2.0) starts, connects to NEW schema ✅
    → Pod-1 (v1.0) terminates
    → Pod-5 (v2.0) starts, connects to NEW schema ✅
    → Pod-2 (v1.0) terminates
    → Pod-6 (v2.0) starts, connects to NEW schema ✅
    → Pod-3 (v1.0) terminates

PROBLEM: During T1-T2 overlap:
    → Pod-1/2/3 (v1.0) query vehicles.vehicle_class → COLUMN NOT FOUND
    → Pod-1/2/3 (v1.0) insert with old status enum → TYPE MISMATCH
    → 500 errors during deployment window
```

### Recommended: 3-Phase Expand & Contract Pattern

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    EXPAND & CONTRACT DEPLOYMENT                         │
│                    (Zero-Downtime Schema Evolution)                     │
└─────────────────────────────────────────────────────────────────────────┘

SCENARIO: Add vehicles.vehicle_class column (NOT NULL with default)

┌─────────────────────────────────────────────────────────────────────────┐
│  PHASE 1: EXPAND (Backward-Compatible Addition)                         │
│  ──────────────────────────────────────                                 │
│  Duration: 24-48 hours (allow all pods to deploy)                       │
│                                                                         │
│  Migration (ADDITIVE ONLY):                                             │
│  ```sql                                                                 │
│  -- Step 1: Add column with DEFAULT (nullable first)                   │
│  ALTER TABLE vehicles                                                  │
│  ADD COLUMN vehicle_class VARCHAR(20) DEFAULT 'class_a';               │
│                                                                         │
│  -- Step 2: Backfill existing rows (in batches to avoid lock)          │
│  UPDATE vehicles                                                       │
│  SET vehicle_class = 'class_a'                                         │
│  WHERE vehicle_class IS NULL                                           │
│  AND id BETWEEN ? AND ?;  -- Batch processing                          │
│                                                                         │
│  -- Step 3: Add NOT NULL constraint (after backfill complete)          │
│  ALTER TABLE vehicles                                                  │
│  ALTER COLUMN vehicle_class SET NOT NULL;                              │
│  ```                                                                    │
│                                                                         │
│  Code Changes (v1.1 - COMPATIBLE):                                      │
│  ```go                                                                  │
│  // OLD CODE (still works)                                              │
│  type Vehicle struct {                                                  │
│      ID           uuid.UUID                                            │
│      LicensePlate string                                               │
│      Status       string                                               │
│      // vehicle_class ignored in v1.1                                  │
│  }                                                                      │
│                                                                         │
│  // NEW CODE (v2.0 - not deployed yet)                                  │
│  type Vehicle struct {                                                  │
│      ID           uuid.UUID                                            │
│      LicensePlate string                                               │
│      Status       string                                               │
│      VehicleClass string  // ← New field, optional in v1.1             │
│  }                                                                      │
│  ```                                                                    │
│                                                                         │
│  Deployment:                                                            │
│  → Apply migration (ADD COLUMN)                                        │
│  → Deploy v1.1 code (ignores new column)                               │
│  → All pods (v1.0 and v1.1) work with expanded schema                  │
│  → Wait 24-48 hours for full rollout                                   │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ All pods running v1.1+
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  PHASE 2: TRANSITION (Dual-Write Support)                               │
│  ───────────────────────────────                                        │
│  Duration: 7 days (allow data migration)                                │
│                                                                         │
│  Code Changes (v2.0 - DUAL WRITE):                                      │
│  ```go                                                                  │
│  type VehicleService struct {                                           │
│      repo *VehicleRepository                                            │
│  }                                                                      │
│                                                                         │
│  // CreateVehicle writes to BOTH old and new format                    │
│  func (s *VehicleService) CreateVehicle(ctx context.Context, v *Vehicle) error {
│      // Write new format (primary)                                     │
│      err := s.repo.CreateV2(ctx, v)                                     │
│      if err != nil {                                                    │
│          return err                                                     │
│      }                                                                  │
│                                                                         │
│      // ALSO write legacy format (compatibility during transition)     │
│      legacyVehicle := v.ToLegacy()  // Drop vehicle_class               │
│      _ = s.repo.CreateLegacy(ctx, legacyVehicle)                        │
│                                                                         │
│      return nil                                                         │
│  }                                                                      │
│                                                                         │
│  // GetVehicle reads from new format, fallback to legacy               │
│  func (s *VehicleService) GetVehicle(ctx context.Context, id uuid.UUID) (*Vehicle, error) {
│      // Try new format first                                           │
│      v, err := s.repo.GetV2(ctx, id)                                    │
│      if err == nil {                                                    │
│          return v, nil                                                  │
│      }                                                                  │
│                                                                         │
│      // Fallback to legacy format                                      │
│      legacy, err := s.repo.GetLegacy(ctx, id)                           │
│      if err != nil {                                                    │
│          return nil, err                                                │
│      }                                                                  │
│                                                                         │
│      // Upgrade legacy to new format                                   │
│      return legacy.Upgrade(), nil                                       │
│  }                                                                      │
│  ```                                                                    │
│                                                                         │
│  Deployment:                                                            │
│  → Deploy v2.0 code (dual-write support)                               │
│  → All pods read/write both formats                                    │
│  → Zero errors during transition                                       │
│  → Monitor for 7 days                                                  │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ Dual-write stable, no errors
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  PHASE 3: CONTRACT (Remove Legacy Support)                              │
│  ─────────────────────────────────                                      │
│  Duration: 24 hours (final cleanup)                                     │
│                                                                         │
│  Migration (REMOVE LEGACY):                                             │
│  ```sql                                                                 │
│  -- Step 1: Verify no legacy-only data                                 │
│  SELECT COUNT(*) FROM vehicles                                         │
│  WHERE vehicle_class IS NULL;  -- Should be 0                          │
│                                                                         │
│  -- Step 2: Drop legacy columns (if any)                               │
│  -- (In this example, no columns to drop)                              │
│                                                                         │
│  -- Step 3: Add indexes for new query patterns                         │
│  CREATE INDEX idx_vehicles_class ON vehicles(vehicle_class);           │
│  ```                                                                    │
│                                                                         │
│  Code Changes (v2.1 - CLEAN):                                           │
│  ```go                                                                  │
│  // REMOVE dual-write logic                                            │
│  func (s *VehicleService) CreateVehicle(ctx context.Context, v *Vehicle) error {
│      // Only write new format                                          │
│      return s.repo.CreateV2(ctx, v)                                     │
│  }                                                                      │
│                                                                         │
│  func (s *VehicleService) GetVehicle(ctx context.Context, id uuid.UUID) (*Vehicle, error) {
│      // Only read new format                                           │
│      return s.repo.GetV2(ctx, id)                                       │
│  }                                                                      │
│  ```                                                                    │
│                                                                         │
│  Deployment:                                                            │
│  → Deploy v2.1 code (clean, no dual-write)                             │
│  → Apply cleanup migration                                             │
│  → Rollout complete                                                    │
│                                                                         │
│  TOTAL DOWNTIME: ZERO                                                   │
└─────────────────────────────────────────────────────────────────────────┘
```

### Expand & Contract Timeline

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    EXPAND & CONTRACT TIMELINE                           │
└─────────────────────────────────────────────────────────────────────────┘

Day 0:  PHASE 1 EXPAND begins
        → Migration: ADD COLUMN vehicle_class DEFAULT 'class_a'
        → Deploy v1.1 code (ignores new column)
        → Duration: 24-48 hours
        
Day 2:  PHASE 1 EXPAND complete
        → All pods running v1.1+
        → Schema expanded, all pods compatible
        
Day 2-9: PHASE 2 TRANSITION
        → Deploy v2.0 code (dual-write support)
        → Read/write both old and new format
        → Duration: 7 days (monitoring period)
        
Day 9:  PHASE 2 TRANSITION complete
        → No errors during transition
        → Data fully migrated
        
Day 10: PHASE 3 CONTRACT begins
        → Deploy v2.1 code (clean, no dual-write)
        → Migration: DROP legacy columns, ADD indexes
        → Duration: 24 hours
        
Day 11: DEPLOYMENT COMPLETE
        → All pods running v2.1
        → Schema contracted to new format
        → Zero downtime achieved
```

### Breaking Change Classification

| Change Type | Pattern | Risk | Example |
|-------------|---------|------|---------|
| **Add Column (nullable)** | Expand only | Low | `ADD COLUMN notes TEXT` |
| **Add Column (NOT NULL + DEFAULT)** | Expand + Contract | Low | `ADD COLUMN status VARCHAR(20) DEFAULT 'active'` |
| **Remove Column** | Expand + Contract | High | `DROP COLUMN legacy_field` |
| **Rename Column** | Expand + Contract | High | `RENAME COLUMN old_name TO new_name` |
| **Change Type (widen)** | Expand + Contract | Medium | `VARCHAR(50) → VARCHAR(100)` |
| **Change Type (narrow)** | Expand + Contract | Critical | `VARCHAR(100) → VARCHAR(50)` |
| **Add Index** | Expand only | Low | `CREATE INDEX idx_status` |
| **Remove Index** | Expand + Contract | Low | `DROP INDEX idx_old` |

---

## 2️⃣ Migration Guardrails: Preventing Long-Running Locks

### Current Migration System

**Location**: `cmd/migrate/main.go`

```go
// CURRENT: Sequential migration execution (NO guardrails)
func runMigrations(db *sql.DB) {
    files, _ := filepath.Glob("migrations/*.sql")
    sort.Strings(files)
    
    for _, file := range files {
        content, _ := os.ReadFile(file)
        
        tx, _ := db.Begin()
        _, err := tx.Exec(string(content))  // ← ENTIRE migration in one transaction
        if err != nil {
            tx.Rollback()
            log.Fatalf("failed to apply %s: %v", file, err)
        }
        
        tx.Commit()  // ← Lock held until commit
    }
}
```

### ⚠️ CRITICAL GAP: No Lock Timeout or Batch Processing

**Problem**: Long-running migration blocks database

```sql
-- migrations/072_tracking_enterprise.sql (REAL EXAMPLE)
-- This migration could take 30+ minutes on large tables

ALTER TABLE gps_locations ADD COLUMN org_id UUID;
ALTER TABLE obd_readings ADD COLUMN org_id UUID;
ALTER TABLE fuel_readings ADD COLUMN org_id UUID;

-- Add 100M rows × 3 tables = 300M row updates
-- Each row update acquires lock
-- Total lock time: 30+ minutes

-- During migration:
-- → All INSERT/UPDATE/SELECT queries BLOCKED
-- → Database throttling triggers (90% pool usage)
-- → Auth service can't login users
-- → Emergency service can't process SOS
-- → CRITICAL SERVICES STARVED
```

### Recommended: Migration Guardrails System

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    MIGRATION GUARDRAILS SYSTEM                          │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│  GUARDRAIL 1: LOCK TIMEOUT ENFORCEMENT                                  │
│  ─────────────────────────────────────                                  │
│                                                                         │
│  Problem: ALTER TABLE locks entire table                               │
│  Solution: Set statement_timeout per migration                         │
│                                                                         │
│  Implementation:                                                        │
│  ```sql                                                                 │
│  -- migrations/072_tracking_enterprise.sql (MODIFIED)                  │
│                                                                         │
│  -- Set timeout for this migration (5 minutes)                         │
│  SET statement_timeout = '5min';                                       │
│  SET lock_timeout = '1min';                                            │
│                                                                         │
│  -- Use CONCURRENTLY for index creation (no lock)                      │
│  CREATE INDEX CONCURRENTLY IF NOT EXISTS                               │
│  idx_gps_org_id ON gps_locations(org_id);                              │
│                                                                         │
│  -- Use online schema change for ALTER (PostgreSQL 12+)                │
│  -- Or use pg_repack for zero-downtime ALTER                           │
│  ```                                                                    │
│                                                                         │
│  Go Implementation:                                                     │
│  ```go                                                                  │
│  func runMigrationsWithGuardrails(db *sql.DB) error {                   │
│      files, _ := filepath.Glob("migrations/*.sql")                     │
│                                                                         │
│      for _, file := range files {                                      │
│          // Check if migration is "long-running"                       │
│          metadata, _ := parseMigrationMetadata(file)                   │
│                                                                         │
│          if metadata.RequiresLock {                                    │
│              // Acquire migration lock                                 │
│              lockAcquired, err := tryAcquireMigrationLock(db, file)    │
│              if err != nil || !lockAcquired {                          │
│                  log.Printf("Migration %s skipped - lock unavailable",  │
│                      file)                                              │
│                  continue  // Skip, retry next deployment              │
│              }                                                          │
│              defer releaseMigrationLock(db, file)                      │
│          }                                                              │
│                                                                         │
│          // Set statement timeout                                      │
│          timeout := metadata.TimeoutMinutes                            │
│          if timeout == 0 {                                             │
│              timeout = 5  // Default 5 minutes                         │
│          }                                                              │
│                                                                         │
│          _, err = db.Exec(fmt.Sprintf(                                 │
│              "SET statement_timeout = '%dmin'", timeout))              │
│                                                                         │
│          // Execute migration                                          │
│          content, _ := os.ReadFile(file)                               │
│          _, err = db.Exec(string(content))                             │
│          if err != nil {                                               │
│              if strings.Contains(err.Error(), "canceling statement") { │
│                  log.Printf("Migration %s TIMEOUT - rolling back",     │
│                      file)                                              │
│                  continue  // Don't fail deployment, skip migration   │
│              }                                                          │
│              return err                                                 │
│          }                                                              │
│      }                                                                  │
│      return nil                                                         │
│  }                                                                      │
│  ```                                                                    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  GUARDRAIL 2: BATCH PROCESSING FOR LARGE UPDATES                        │
│  ─────────────────────────────────────────────                          │
│                                                                         │
│  Problem: UPDATE 100M rows holds lock for 30+ minutes                  │
│  Solution: Batch updates in small chunks                               │
│                                                                         │
│  Implementation:                                                        │
│  ```sql                                                                 │
│  -- migrations/NEW_batch_update_template.sql                           │
│                                                                         │
│  -- Create batch processing function                                   │
│  CREATE OR REPLACE FUNCTION batch_update_org_id(batch_size INT)        │
│  RETURNS VOID AS $$                                                    │
│  DECLARE                                                               │
│      updated_count INT;                                                │
│      total_updated INT := 0;                                           │
│  BEGIN                                                                 │
│      LOOP                                                              │
│          -- Update one batch at a time                                 │
│          UPDATE gps_locations                                          │
│          SET org_id = (SELECT org_id FROM vehicles                     │
│                        WHERE vehicles.id = gps_locations.vehicle_id)   │
│          WHERE id IN (                                                 │
│              SELECT id FROM gps_locations                              │
│              WHERE org_id IS NULL                                      │
│              LIMIT batch_size                                          │
│          );                                                            │
│                                                                         │
│          GET DIAGNOSTICS updated_count = ROW_COUNT;                    │
│          total_updated := total_updated + updated_count;               │
│                                                                         │
│          -- Exit if no more rows                                       │
│          EXIT WHEN updated_count = 0;                                  │
│                                                                         │
│          -- Small delay to allow other queries                         │
│          PERFORM pg_sleep(0.1);                                        │
│                                                                         │
│          -- Log progress                                               │
│          RAISE NOTICE 'Batch complete: % rows updated (total: %)',     │
│              updated_count, total_updated;                             │
│      END LOOP;                                                         │
│                                                                         │
│      RAISE NOTICE 'Migration complete: % total rows updated',          │
│          total_updated;                                                │
│  END;                                                                  │
│  $$ LANGUAGE plpgsql;                                                  │
│                                                                         │
│  -- Execute with batch size of 10,000                                  │
│  SELECT batch_update_org_id(10000);                                    │
│  ```                                                                    │
│                                                                         │
│  Benefits:                                                              │
│  • Each batch holds lock for ~100ms (not 30 minutes)                   │
│  • Other queries can interleave between batches                        │
│  • Progress visible in real-time                                       │
│  • Can be paused/resumed if needed                                     │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  GUARDRAIL 3: DEPLOYMENT-TIME MIGRATION CHECK                           │
│  ─────────────────────────────────────────────                          │
│                                                                         │
│  Problem: Migration runs during deployment, blocks traffic             │
│  Solution: Pre-check migration impact, defer if too risky              │
│                                                                         │
│  Implementation:                                                        │
│  ```go                                                                  │
│  // internal/database/migration_guard.go (NEW)                         │
│  package database                                                       │
│                                                                         │
│  type MigrationMetadata struct {                                        │
│      File            string                                            │
│      RequiresLock    bool                                              │
│      TimeoutMinutes  int                                               │
│      EstimatedRows   int64  // From ANALYZE statistics                 │
│      EstimatedDuration time.Duration                                   │
│      RiskLevel       string // low, medium, high, critical             │
│  }                                                                      │
│                                                                         │
│  // parseMigrationMetadata extracts metadata from SQL comments         │
│  func parseMigrationMetadata(file string) (*MigrationMetadata, error) { │
│      content, _ := os.ReadFile(file)                                   │
│                                                                         │
│      // Parse metadata from header comments                            │
│      // -- migration: 072_tracking_enterprise                          │
│      // -- lock_required: true                                         │
│      // -- timeout_minutes: 30                                         │
│      // -- estimated_rows: 100000000                                   │
│      // -- risk_level: high                                            │
│                                                                         │
│      metadata := &MigrationMetadata{                                   │
│          File: file,                                                   │
│      }                                                                 │
│                                                                         │
│      // Extract metadata using regex                                   │
│      re := regexp.MustCompile(`--\s*(\w+):\s*(.+)`)                    │
│      for _, match := range re.FindAllStringSubmatch(string(content), -1) {
│          switch match[1] {                                             │
│          case "lock_required":                                         │
│              metadata.RequiresLock = match[2] == "true"                │
│          case "timeout_minutes":                                       │
│              metadata.TimeoutMinutes, _ = strconv.Atoi(match[2])       │
│          case "risk_level":                                            │
│              metadata.RiskLevel = match[2]                             │
│          }                                                              │
│      }                                                                 │
│                                                                         │
│      return metadata, nil                                              │
│  }                                                                      │
│                                                                         │
│  // ShouldDeferMigration checks if migration should be deferred        │
│  func ShouldDeferMigration(db *sql.DB, metadata *MigrationMetadata) bool {
│      // Defer high-risk migrations during peak hours                   │
│      if metadata.RiskLevel == "critical" {                             │
│          hour := time.Now().Hour()                                     │
│          if hour >= 9 && hour <= 18 {  // Peak hours 9 AM - 6 PM      │
│              return true  // Defer to off-peak                         │
│          }                                                              │
│      }                                                                 │
│                                                                         │
│      // Defer if database is under load                                │
│      var activeConnections int                                         │
│      db.QueryRow("SELECT COUNT(*) FROM pg_stat_activity                │
│                   WHERE state = 'active'").Scan(&activeConnections)    │
│                                                                         │
│      if activeConnections > 50 {  // High load threshold               │
│          return true  // Defer until load decreases                    │
│      }                                                                 │
│                                                                         │
│      return false  // Safe to run                                      │
│  }                                                                      │
│  ```                                                                    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  GUARDRAIL 4: DATABASE THROTTLING BYPASS FOR MIGRATIONS                 │
│  ─────────────────────────────────────────────────────────              │
│                                                                         │
│  Problem: Migration triggers database throttling (90% pool usage)      │
│  Solution: Dedicated migration pool, separate from application         │
│                                                                         │
│  Implementation:                                                        │
│  ```go                                                                  │
│  // internal/database/pools.go (ENHANCED)                              │
│  type DatabasePools struct {                                           │
│      Critical   *pgxpool.Pool  // auth, emergency                     │
│      High       *pgxpool.Pool  // billing, vehicle, tracking          │
│      Medium     *pgxpool.Pool  // analytics, reports                  │
│      Low        *pgxpool.Pool  // diagnostics, integrations           │
│      Migration  *pgxpool.Pool  // ← NEW: Dedicated for migrations     │
│  }                                                                      │
│                                                                         │
│  func NewDatabasePools(...) (*DatabasePools, error) {                  │
│      // ... existing pools ...                                         │
│                                                                         │
│      // Create dedicated migration pool (small, isolated)              │
│      migrationConfig, _ := pgxpool.ParseConfig(databaseURL)            │
│      migrationConfig.MaxConns = 2  // Only 2 connections for migrations
│      migrationConfig.MinConns = 0                                       │
│                                                                         │
│      // Tag migration connections for monitoring                       │
│      migrationConfig.ConnConfig.RuntimeParams["application_name"] =    │
│          "flyfleet-migration"                                          │
│                                                                         │
│      pools.Migration, _ = pgxpool.NewWithConfig(ctx, migrationConfig)  │
│                                                                         │
│      return pools, nil                                                 │
│  }                                                                      │
│                                                                         │
│  // Migration runner uses dedicated pool                              │
│  func runMigrations(db *pgxpool.Pool) error {                          │
│      // Use migration pool, not application pools                      │
│      migrationPool := db.GetMigrationPool()                            │
│                                                                         │
│      // This won't compete with auth/vehicle/tracking queries          │
│      _, err := migrationPool.Exec(ctx, migrationSQL)                   │
│      return err                                                         │
│  }                                                                      │
│  ```                                                                    │
│                                                                         │
│  PostgreSQL Configuration:                                              │
│  ```sql                                                                 │
│  -- postgresql.conf                                                    │
│                                                                         │
│  # Connection limits per application_name                              │
│  ALTER ROLE "flyfleet_migration" SET statement_timeout = '30min';      │
│  ALTER ROLE "flyfleet_migration" SET lock_timeout = '5min';            │
│                                                                         │
│  # Monitor migration connections separately                            │
│  SELECT application_name, COUNT(*)                                    │
│  FROM pg_stat_activity                                                │
│  WHERE application_name LIKE 'flyfleet%'                              │
│  GROUP BY application_name;                                            │
│  ```                                                                    │
└─────────────────────────────────────────────────────────────────────────┘
```

### Migration Metadata Standard

```sql
-- All migrations must include header metadata
-- migrations/072_tracking_enterprise.sql

-- migration: 072_tracking_enterprise
-- description: Add org_id to telemetry tables for RLS
-- lock_required: true
-- timeout_minutes: 30
-- estimated_rows: 100000000
-- estimated_duration: 25min
-- risk_level: high
-- off_peak_only: true
-- batch_processing: true
-- batch_size: 10000
-- rollback_safe: true

-- Migration SQL follows...
ALTER TABLE gps_locations ADD COLUMN org_id UUID;
```

### Migration Risk Matrix

| Risk Level | Criteria | Deployment Window | Approval Required |
|------------|----------|-------------------|-------------------|
| **Low** | < 1M rows, no locks | Any time | Auto-approved |
| **Medium** | 1M-10M rows, short locks | Off-peak (6 PM - 6 AM) | Tech lead |
| **High** | 10M-100M rows, long locks | Off-peak, weekend | Engineering manager |
| **Critical** | >100M rows, table rewrite | Maintenance window | CTO + DBA |

---

## 3️⃣ The Kill Switch: Hot-Disabling NATS Consumers

### Current NATS Subscription Pattern

**Location**: `internal/domains/maintenance/service/telemetry_worker.go`

```go
// CURRENT: Subscriptions created at startup, no hot-disable
func (w *TelemetryWorker) Start() error {
    sub, err := w.nc.Subscribe("fleet.telemetry.gps", func(msg *nats.Msg) {
        // Process telemetry...
    })
    if err != nil {
        return fmt.Errorf("failed to subscribe: %w", err)
    }
    
    w.logger.Info("Subscribed to fleet.telemetry.gps")
    
    // Subscription lives until worker stops
    // NO mechanism to hot-disable without restart
    <-w.stopCh
    sub.Unsubscribe()
}
```

### ⚠️ CRITICAL GAP: No Domain-Level Kill Switch

**Problem**: Need to stop india_v2 processing without restarting monolith

```
Scenario: india_v2 integration causing issues
→ VAHAN API rate limit exceeded
→ Every GPS ping triggers VAHAN lookup
→ NATS queue backs up (10K pending messages)
→ Other domains affected (alert_v2, billing_v2)
→ System-wide degradation

Current Solution:
→ Restart entire monolith pod
→ Downtime: 2-3 minutes
→ All domains affected during restart
→ NOT acceptable for production

Required Solution:
→ Hot-disable india_v2 NATS consumers
→ Other domains continue processing
→ Zero downtime
```

### Recommended: Domain Kill Switch Implementation

```go
// internal/platform/domain_registry.go (NEW)
package platform

import (
    "context"
    "sync"
    
    "github.com/nats-io/nats.go"
    "go.uber.org/zap"
)

// DomainState represents the operational state of a domain
type DomainState string

const (
    DomainStateActive   DomainState = "ACTIVE"
    DomainStateDraining DomainState = "DRAINING"  // Stop accepting new, finish existing
    DomainStateDisabled DomainState = "DISABLED"  // Fully stopped
)

// DomainRegistry manages the lifecycle of all domain consumers
type DomainRegistry struct {
    mu       sync.RWMutex
    domains  map[string]*DomainController
    natsConn *nats.Conn
    logger   *zap.Logger
}

// DomainController manages a single domain's subscriptions
type DomainController struct {
    Name        string
    State       DomainState
    Subscriptions []*nats.Subscription
    StopCh      chan struct{}
    Config      DomainConfig
}

// DomainConfig defines domain-specific settings
type DomainConfig struct {
    Name            string
    Subjects        []string  // NATS subjects to subscribe
    Handler         func(*nats.Msg)
    EnabledByDefault bool
    Priority        string  // critical, high, medium, low
}

// NewDomainRegistry creates the registry
func NewDomainRegistry(nc *nats.Conn, logger *zap.Logger) *DomainRegistry {
    return &DomainRegistry{
        domains:  make(map[string]*DomainController),
        natsConn: nc,
        logger:   logger,
    }
}

// RegisterDomain registers a domain controller
func (r *DomainRegistry) RegisterDomain(config DomainConfig) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if _, exists := r.domains[config.Name]; exists {
        return fmt.Errorf("domain %s already registered", config.Name)
    }
    
    controller := &DomainController{
        Name:   config.Name,
        State:  DomainStateActive,
        StopCh: make(chan struct{}),
        Config: config,
    }
    
    // Subscribe if enabled by default
    if config.EnabledByDefault {
        if err := r.subscribeDomain(controller); err != nil {
            return err
        }
    }
    
    r.domains[config.Name] = controller
    
    r.logger.Info("Domain registered",
        zap.String("domain", config.Name),
        zap.Bool("enabled", config.EnabledByDefault))
    
    return nil
}

// subscribeDomain creates NATS subscriptions for a domain
func (r *DomainRegistry) subscribeDomain(controller *DomainController) error {
    for _, subject := range controller.Config.Subjects {
        sub, err := r.natsConn.Subscribe(subject, func(msg *nats.Msg) {
            // Check if domain is still active before processing
            controller.mu.RLock()
            state := controller.State
            controller.mu.RUnlock()
            
            if state != DomainStateActive {
                r.logger.Debug("Domain disabled, dropping message",
                    zap.String("domain", controller.Name),
                    zap.String("subject", subject))
                return
            }
            
            // Process message
            controller.Config.Handler(msg)
        })
        
        if err != nil {
            return fmt.Errorf("failed to subscribe to %s: %w", subject, err)
        }
        
        controller.Subscriptions = append(controller.Subscriptions, sub)
    }
    
    return nil
}

// DisableDomain hot-disables a domain's consumers
func (r *DomainRegistry) DisableDomain(ctx context.Context, domainName string) error {
    r.mu.Lock()
    controller, exists := r.domains[domainName]
    r.mu.Unlock()
    
    if !exists {
        return fmt.Errorf("domain %s not found", domainName)
    }
    
    controller.mu.Lock()
    defer controller.mu.Unlock()
    
    if controller.State == DomainStateDisabled {
        r.logger.Info("Domain already disabled", zap.String("domain", domainName))
        return nil
    }
    
    r.logger.Info("Disabling domain",
        zap.String("domain", domainName),
        zap.String("previous_state", string(controller.State)))
    
    // Transition to draining first (finish existing messages)
    controller.State = DomainStateDraining
    
    // Wait for in-flight messages to complete (max 30 seconds)
    drainTimeout := time.After(30 * time.Second)
    select {
    case <-drainTimeout:
        r.logger.Warn("Domain drain timeout, forcing disable",
            zap.String("domain", domainName))
    case <-time.After(5 * time.Second):
        // Allow 5 seconds for in-flight processing
    }
    
    // Unsubscribe from all subjects
    for _, sub := range controller.Subscriptions {
        if err := sub.Unsubscribe(); err != nil {
            r.logger.Error("Failed to unsubscribe",
                zap.String("domain", domainName),
                zap.Error(err))
        }
    }
    
    controller.State = DomainStateDisabled
    controller.Subscriptions = nil
    
    r.logger.Info("Domain disabled",
        zap.String("domain", domainName),
        zap.Int("subscriptions_removed", len(controller.Subscriptions)))
    
    return nil
}

// EnableDomain hot-enables a domain's consumers
func (r *DomainRegistry) EnableDomain(ctx context.Context, domainName string) error {
    r.mu.Lock()
    controller, exists := r.domains[domainName]
    r.mu.Unlock()
    
    if !exists {
        return fmt.Errorf("domain %s not found", domainName)
    }
    
    controller.mu.Lock()
    defer controller.mu.Unlock()
    
    if controller.State == DomainStateActive {
        r.logger.Info("Domain already active", zap.String("domain", domainName))
        return nil
    }
    
    r.logger.Info("Enabling domain",
        zap.String("domain", domainName),
        zap.String("previous_state", string(controller.State)))
    
    // Subscribe to all subjects
    if err := r.subscribeDomain(controller); err != nil {
        return err
    }
    
    controller.State = DomainStateActive
    
    r.logger.Info("Domain enabled",
        zap.String("domain", domainName),
        zap.Int("subscriptions_created", len(controller.Subscriptions)))
    
    return nil
}

// GetDomainState returns the current state of a domain
func (r *DomainRegistry) GetDomainState(domainName string) DomainState {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    controller, exists := r.domains[domainName]
    if !exists {
        return ""
    }
    
    controller.mu.RLock()
    defer controller.mu.RUnlock()
    
    return controller.State
}

// ListDomains returns all registered domains and their states
func (r *DomainRegistry) ListDomains() map[string]DomainState {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    states := make(map[string]DomainState)
    for name, controller := range r.domains {
        controller.mu.RLock()
        states[name] = controller.State
        controller.mu.RUnlock()
    }
    
    return states
}
```

### Domain Registration in Main

```go
// cmd/main.go (ENHANCED)
func main() {
    // ... existing setup ...
    
    // Create domain registry
    domainRegistry := platform.NewDomainRegistry(nc, logger)
    
    // Register all domains with kill switch support
    
    // India V2 Domain (can be hot-disabled)
    domainRegistry.RegisterDomain(platform.DomainConfig{
        Name: "india_v2",
        Subjects: []string{
            "india.fastag.transaction.created",
            "india.ewaybill.created",
            "india.vahan.verification.requested",
            "india.sarathi.verification.requested",
        },
        Handler:          indiaHandlers.HandleIndiaEvent,
        EnabledByDefault: true,
        Priority:         "medium",  // Can be disabled if causing issues
    })
    
    // Alert V2 Domain (CRITICAL - never disable)
    domainRegistry.RegisterDomain(platform.DomainConfig{
        Name: "alert_v2",
        Subjects: []string{
            "alert.rule.triggered",
            "alert.notification.send",
        },
        Handler:          alertHandlers.HandleAlertEvent,
        EnabledByDefault: true,
        Priority:         "critical",  // Never disable
    })
    
    // Billing V2 Domain (HIGH - disable only in emergency)
    domainRegistry.RegisterDomain(platform.DomainConfig{
        Name: "billing_v2",
        Subjects: []string{
            "billing.invoice.generated",
            "billing.payment.received",
        },
        Handler:          billingHandlers.HandleBillingEvent,
        EnabledByDefault: true,
        Priority:         "high",
    })
    
    // ... register all other domains ...
    
    // Expose kill switch via HTTP API
    r := chi.NewRouter()
    
    // Domain management endpoints (protected)
    r.Route("/api/control/domains", func(r chi.Router) {
        r.Use(authMiddleware)  // Require admin auth
        r.Use(rbacMiddleware("admin:domains:manage"))
        
        r.Get("/", func(w http.ResponseWriter, r *http.Request) {
            // List all domains and their states
            states := domainRegistry.ListDomains()
            jsonOK(w, map[string]interface{}{
                "domains": states,
            })
        })
        
        r.Post("/{domain}/disable", func(w http.ResponseWriter, r *http.Request) {
            domain := chi.URLParam(r, "domain")
            
            // Check if domain can be disabled (not critical)
            if domain == "alert_v2" || domain == "auth_v2" {
                http.Error(w, "Cannot disable critical domain", http.StatusBadRequest)
                return
            }
            
            err := domainRegistry.DisableDomain(r.Context(), domain)
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            
            jsonOK(w, map[string]string{
                "status": "disabled",
                "domain": domain,
            })
        })
        
        r.Post("/{domain}/enable", func(w http.ResponseWriter, r *http.Request) {
            domain := chi.URLParam(r, "domain")
            
            err := domainRegistry.EnableDomain(r.Context(), domain)
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            
            jsonOK(w, map[string]string{
                "status": "enabled",
                "domain": domain,
            })
        })
    })
    
    // ... rest of server setup ...
}
```

### Kill Switch API Usage

```bash
# List all domains and their states
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8084/api/control/domains

# Response:
{
  "domains": {
    "india_v2": "ACTIVE",
    "alert_v2": "ACTIVE",
    "billing_v2": "ACTIVE",
    "auth_v2": "ACTIVE",
    "tracking_v2": "ACTIVE"
  }
}

# Hot-disable india_v2 (stops processing immediately)
curl -X POST -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8084/api/control/domains/india_v2/disable

# Response:
{
  "status": "disabled",
  "domain": "india_v2"
}

# Verify domain is disabled
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8084/api/control/domains

# Response:
{
  "domains": {
    "india_v2": "DISABLED",  # ← Now disabled
    "alert_v2": "ACTIVE",
    "billing_v2": "ACTIVE",
    "auth_v2": "ACTIVE",
    "tracking_v2": "ACTIVE"
  }
}

# Re-enable india_v2 after fixing the issue
curl -X POST -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8084/api/control/domains/india_v2/enable
```

### Kill Switch State Machine

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    DOMAIN KILL SWITCH STATE MACHINE                     │
└─────────────────────────────────────────────────────────────────────────┘

                    ┌─────────────┐
                    │   ACTIVE    │ ← Normal operation
                    │ (processing)│
                    └──────┬──────┘
                           │ POST /domains/{name}/disable
                           ▼
                    ┌─────────────┐
                    │  DRAINING   │ ← Stop accepting new messages
                    │(finishing   │    Finish in-flight processing
                    │ in-flight)  │    Max 30 seconds
                    └──────┬──────┘
                           │ After drain complete (or timeout)
                           ▼
                    ┌─────────────┐
                    │  DISABLED   │ ← Fully stopped
                    │  (stopped)  │    NATS subscriptions removed
                    └──────┬──────┘
                           │ POST /domains/{name}/enable
                           ▼
                    ┌─────────────┐
                    │   ACTIVE    │ ← Subscriptions recreated
                    │ (processing)│    Resume normal operation
                    └─────────────┘
```

### Domain Priority Classification

| Priority | Domains | Can Disable? | Approval Required |
|----------|---------|---------------|-------------------|
| **Critical** | auth_v2, alert_v2, emergency | ❌ Never | N/A |
| **High** | billing_v2, tracking_v2, vehicle | ⚠️ Emergency only | Engineering manager |
| **Medium** | india_v2, integrations_v2, reports_v2 | ✅ Yes | Tech lead |
| **Low** | analytics, governance, settings | ✅ Yes | Auto-approved |

---

## Summary: Production-Ready Configuration

### 1. Expand & Contract Pattern

```bash
# Deployment settings
EXPAND_CONTRACT_ENABLED=true
PHASE1_EXPAND_DURATION=48h      # Hours for expand phase
PHASE2_TRANSITION_DURATION=7d   # Days for dual-write phase
PHASE3_CONTRACT_DURATION=24h    # Hours for contract phase
```

### 2. Migration Guardrails

```bash
# Migration settings
MIGRATION_LOCK_TIMEOUT=5m           # Max lock duration
MIGRATION_STATEMENT_TIMEOUT=30m     # Max statement duration
MIGRATION_BATCH_SIZE=10000          # Rows per batch
MIGRATION_DEDICATED_POOL=true       # Separate connection pool
MIGRATION_OFF_PEAK_ONLY=true        # Defer high-risk migrations
```

### 3. Domain Kill Switch

```bash
# Kill switch settings
KILL_SWITCH_ENABLED=true
KILL_SWITCH_DRAIN_TIMEOUT=30s       # Max drain time
KILL_SWITCH_CRITICAL_DOMAINS=auth_v2,alert_v2,emergency
KILL_SWITCH_ADMIN_RBAC=admin:domains:manage
```

---

## Conclusion

The current implementation has **three critical gaps**:

1. **No expand & contract pattern** → Breaking changes cause downtime during rolling updates
2. **No migration guardrails** → Long-running migrations trigger database throttling
3. **No domain kill switch** → Must restart entire monolith to disable a single domain

**Recommended immediate actions**:

1. Implement expand & contract deployment workflow
2. Add migration metadata and guardrails (lock timeout, batch processing)
3. Build domain registry with hot-disable capability

These changes ensure **zero-downtime deployments**, **safe schema evolution**, and **operational flexibility**.
