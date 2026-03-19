# 🛡️ Disaster Recovery & Data Integrity Analysis

**Author**: Lead System Architect  
**Date**: March 15, 2026  
**Scope**: PITR workflow, NATS JetStream replay, multi-region failover

---

## Executive Summary

This document provides a comprehensive analysis of FlyFleet's disaster recovery and data integrity strategy, covering:

1. **Point-in-Time Recovery (PITR)**: Restoring deleted trips while preserving live GPS pings
2. **NATS JetStream Replay**: Re-hydrating analytics cache without double-counting billing transactions
3. **Regional Failover**: Automated AWS Mumbai → Secondary region failover with active-active database sync

---

## 1️⃣ The Point-in-Time Recovery (PITR): Trip Restoration with Live Data Preservation

### Current Backup Schema

**Location**: `infrastructure/migrations/027_disaster_recovery_schema.sql`

```sql
-- CURRENT: Basic backup tracking (no PITR workflow)
CREATE TABLE backups (
    id UUID PRIMARY KEY,
    type VARCHAR(50) NOT NULL,  -- FULL, INCREMENTAL, SNAPSHOT
    status VARCHAR(50) NOT NULL,  -- PENDING, IN_PROGRESS, COMPLETED, FAILED
    s3_key VARCHAR(255) NOT NULL,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE TABLE restore_jobs (
    id UUID PRIMARY KEY,
    backup_id UUID NOT NULL REFERENCES backups(id),
    status VARCHAR(50) NOT NULL,
    target_environment VARCHAR(50) NOT NULL,
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    performed_by UUID NOT NULL,
    logs TEXT
);
```

### ⚠️ CRITICAL GAP: No PITR Workflow for Partial Data Restoration

**Problem Scenario**: Rogue script deletes 10,000 trips

```
T0: 2024-03-15 10:00:00 - Normal operation
    → 50,000 trips in database
    → GPS pings arriving at 100K/minute

T1: 2024-03-15 10:05:00 - Rogue script executes
    → DELETE FROM trips WHERE org_id = '550e8400-...'
    → 10,000 trips DELETED
    → CASCADE DELETE: trip_waypoints, trip_telemetry also deleted

T2: 2024-03-15 10:06:00 - Alert triggered
    → Monitoring detects 10,000 trip deletion
    → On-call engineer paged

T3: 2024-03-15 10:10:00 - Problem identified
    → Rogue script stopped
    → Need to restore 10,000 trips

PROBLEM:
→ Last backup: 2024-03-15 06:00:00 (4 hours ago)
→ 10,000 NEW GPS pings arrived since deletion (T1 to T3)
→ Standard restore would OVERWRITE live data
→ Vehicles appear to "jump" back in time
→ Live tracking broken for 4 hours
```

### Recommended: PITR with Live Data Preservation

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    POINT-IN-TIME RECOVERY WORKFLOW                      │
│                    (Preserving Live GPS Pings)                          │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 1: IDENTIFY DELETION SCOPE                                        │
│  ─────────────────────────────                                          │
│  Query audit logs to determine exact deletion:                          │
│  ```sql                                                                 │
│  SELECT                                                                 │
│      entity_type,                                                       │
│      COUNT(*) as deleted_count,                                         │
│      MIN(created_at) as deletion_start,                                 │
│      MAX(created_at) as deletion_end                                    │
│  FROM audit_logs                                                        │
│  WHERE action = 'BULK_DELETE'                                           │
│    AND entity_type = 'trip'                                             │
│    AND created_at > NOW() - INTERVAL '1 hour'                           │
│  GROUP BY entity_type;                                                  │
│  ```                                                                    │
│                                                                         │
│  Result:                                                                │
│  ┌─────────────┬───────────────┬─────────────────────┬─────────────────┐│
│  │ entity_type │ deleted_count │ deletion_start      │ deletion_end    ││
│  ├─────────────┼───────────────┼─────────────────────┼─────────────────┤│
│  │ trip        │ 10,000        │ 2024-03-15 10:05:00 │ 2024-03-15 10:05││
│  └─────────────┴───────────────┴─────────────────────┴─────────────────┘│
│                                                                         │
│  Recovery Point Objective (RPO): 2024-03-15 10:04:59 (before deletion)  │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 2: CREATE TEMPORARY RESTORE CLUSTER                               │
│  ─────────────────────────────────────                                  │
│  DO NOT restore to production directly!                                 │
│                                                                         │
│  1. Spin up temporary RDS instance from backup                          │
│     ```bash                                                             │
│     aws rds restore-db-instance-from-db-snapshot \                      │
│       --db-instance-identifier flyfleet-pitr-restore \                  │
│       --db-snapshot-identifier flyfleet-snapshot-20240315-0600 \        │
│       --db-instance-class db.r5.large \                                 │
│       --availability-zone ap-south-1b                                   │
│     ```                                                                 │
│                                                                         │
│  2. Wait for restore to complete (~30 minutes for 500GB)                │
│                                                                         │
│  3. Verify restored data                                                │
│     ```sql                                                              │
│     SELECT COUNT(*) FROM trips WHERE org_id = '550e8400-...';          │
│     -- Should return 50,000 (including deleted 10,000)                 │
│     ```                                                                 │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 3: EXPORT DELETED DATA ONLY                                       │
│  ─────────────────────────────────                                      │
│  Extract only the 10,000 deleted trips from restore cluster:           │
│                                                                         │
│  ```sql                                                                 │
│  -- Export deleted trips to S3                                         │
│  aws s3 cp - <<EOF s3://flyfleet-pitr/deleted-trips-20240315.json     │
│  COPY (                                                                 │
│      SELECT                                                            │
│          t.*,                                                           │
│          tw.waypoint_data,                                              │
│          tt.telemetry_data                                              │
│      FROM trips t                                                      │
│      LEFT JOIN trip_waypoints tw ON tw.trip_id = t.id                  │
│      LEFT JOIN trip_telemetry tt ON tt.trip_id = t.id                  │
│      WHERE t.org_id = '550e8400-...'                                   │
│        AND t.created_at < '2024-03-15 10:05:00'                        │
│      FORMAT JSON                                                        │
│  ) TO STDOUT;                                                          │
│  EOF                                                                    │
│  ```                                                                    │
│                                                                         │
│  4. Verify export integrity                                             │
│     ```bash                                                             │
│     jq 'length' deleted-trips-20240315.json                            │
│     -- Should return 10000                                              │
│     ```                                                                 │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 4: CAPTURE LIVE GPS PINGS (CRITICAL STEP)                         │
│  ─────────────────────────────────────────                              │
│  Preserve GPS pings that arrived AFTER deletion (T1 to present):       │
│                                                                         │
│  ```sql                                                                 │
│  -- Export live GPS pings from production                              │
│  COPY (                                                                 │
│      SELECT * FROM gps_locations                                       │
│      WHERE time > '2024-03-15 10:05:00'  -- After deletion            │
│        AND org_id = '550e8400-...'                                     │
│      FORMAT JSON                                                        │
│  ) TO '/tmp/live-gps-pings-20240315.json';                             │
│  ```                                                                    │
│                                                                         │
│  Also capture current vehicle states:                                   │
│  ```sql                                                                 │
│  COPY (                                                                 │
│      SELECT * FROM vehicles                                            │
│      WHERE org_id = '550e8400-...'                                     │
│      FORMAT JSON                                                        │
│  ) TO '/tmp/vehicle-states-20240315.json';                             │
│  ```                                                                    │
│                                                                         │
│  Store in memory for quick access during restore:                       │
│  ```bash                                                                │
│  redis-cli SET flyfleet:pitr:live-gps-20240315 \                       │
│    "$(cat /tmp/live-gps-pings-20240315.json)" EX 3600                 │
│  ```                                                                    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 5: MERGE AND RESTORE TO PRODUCTION                                │
│  ─────────────────────────────────────────                              │
│  Merge deleted trips with live data, then restore:                     │
│                                                                         │
│  ```python                                                              │
│  # pitr_merge.py                                                       │
│  import json                                                           │
│  import psycopg2                                                       │
│                                                                         │
│  # Load deleted trips from restore cluster                             │
│  with open('deleted-trips-20240315.json') as f:                        │
│      deleted_trips = json.load(f)                                      │
│                                                                         │
│  # Load live GPS pings from Redis                                      │
│  live_gps = redis.get('flyfleet:pitr:live-gps-20240315')              │
│                                                                         │
│  # Connect to production                                               │
│  conn = psycopg2.connect(dsn=PROD_DSN)                                 │
│  cur = conn.cursor()                                                   │
│                                                                         │
│  # Begin transaction                                                   │
│  cur.execute("BEGIN")                                                  │
│                                                                         │
│  # Restore deleted trips (with conflict handling)                      │
│  for trip in deleted_trips:                                            │
│      try:                                                              │
│          cur.execute("""                                               │
│              INSERT INTO trips (id, org_id, vehicle_id, ...)           │
│              VALUES (%s, %s, %s, ...)                                  │
│              ON CONFLICT (id) DO NOTHING                               │
│          """, (trip['id'], trip['org_id'], trip['vehicle_id'], ...))  │
│                                                                         │
│          # Restore waypoints                                           │
│          if trip.get('waypoint_data'):                                 │
│              for wp in trip['waypoint_data']:                          │
│                  cur.execute("""                                       │
│                      INSERT INTO trip_waypoints (trip_id, ...)         │
│                      VALUES (%s, ...)                                  │
│                      ON CONFLICT DO NOTHING                            │
│                  """, (trip['id'], ...))                               │
│      except Exception as e:                                            │
│          logger.error(f"Failed to restore trip {trip['id']}: {e}")    │
│                                                                         │
│  # Re-apply live GPS pings (these take precedence)                     │
│  for ping in live_gps:                                                 │
│      cur.execute("""                                                   │
│          INSERT INTO gps_locations (time, vehicle_id, lat, lng, ...)  │
│          VALUES (%s, %s, %s, %s, ...)                                  │
│          ON CONFLICT (time, vehicle_id) DO UPDATE                     │
│          SET lat = EXCLUDED.lat, lng = EXCLUDED.lng, ...              │
│      """, (ping['time'], ping['vehicle_id'], ping['lat'], ...))       │
│                                                                         │
│  # Commit transaction                                                  │
│  cur.execute("COMMIT")                                                 │
│  ```                                                                    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 6: VERIFY DATA INTEGRITY                                          │
│  ───────────────────────────────                                        │
│  Post-restore validation:                                               │
│                                                                         │
│  ```sql                                                                 │
│  -- Verify trip count                                                   │
│  SELECT COUNT(*) FROM trips WHERE org_id = '550e8400-...';             │
│  -- Should return 50,000 (restored)                                    │
│                                                                         │
│  -- Verify GPS continuity                                               │
│  SELECT                                                                 │
│      vehicle_id,                                                        │
│      COUNT(*) as ping_count,                                            │
│      MIN(time) as first_ping,                                           │
│      MAX(time) as last_ping                                             │
│  FROM gps_locations                                                     │
│  WHERE org_id = '550e8400-...'                                          │
│    AND time > '2024-03-15 09:00:00'                                    │
│  GROUP BY vehicle_id                                                    │
│  HAVING COUNT(*) < 100;  -- Flag vehicles with gaps                   │
│                                                                         │
│  -- Verify no time-travel (vehicle shouldn't jump backward)            │
│  SELECT                                                                 │
│      vehicle_id,                                                        │
│      LAG(time) OVER (PARTITION BY vehicle_id ORDER BY time) as prev,  │
│      time as curr                                                       │
│  FROM gps_locations                                                     │
│  WHERE org_id = '550e8400-...'                                          │
│    AND time > '2024-03-15 10:00:00'                                    │
│  LIMIT 100;                                                             │
│  ```                                                                    │
│                                                                         │
│  Expected result: No gaps, no time-travel anomalies                    │
└─────────────────────────────────────────────────────────────────────────┘
```

### PITR Timeline

| Time | Event | Action |
|------|-------|--------|
| **T0 (10:00)** | Normal operation | 50,000 trips, GPS pings @ 100K/min |
| **T1 (10:05)** | Rogue script deletes 10,000 trips | Alert triggered |
| **T2 (10:06)** | On-call paged | Investigation begins |
| **T3 (10:10)** | Problem identified | Start PITR workflow |
| **T4 (10:15)** | Temporary restore cluster spinning up | Wait ~30 min |
| **T5 (10:45)** | Restore cluster ready | Export deleted trips |
| **T6 (10:50)** | Export complete | Capture live GPS pings |
| **T7 (10:55)** | Live GPS captured | Merge and restore to production |
| **T8 (11:00)** | Restore complete | Verify data integrity |
| **T9 (11:15)** | Verification passed | PITR complete |

**Total Downtime**: ZERO (live GPS preserved)  
**Data Loss**: ZERO (all 10,000 trips restored)  
**RPO**: 0 seconds (point-in-time exact)  
**RTO**: ~75 minutes (from detection to verification)

---

## 2️⃣ NATS JetStream Replay: Analytics Cache Re-hydration

### Current NATS JetStream Configuration

**Location**: `infrastructure/k8s/nats-jetstream.yaml`

```yaml
# CURRENT: Basic JetStream config (no replay strategy)
apiVersion: v1
kind: ConfigMap
metadata:
  name: nats-config
data:
  nats.conf: |
    jetstream {
      store_dir: "/data/jetstream"
      max_mem_store: 4GB
      max_file_store: 100GB
    }
    
    # Production Streams Configuration
    # Note: Streams are typically created via nats-cli or service init
```

### ⚠️ CRITICAL GAP: No Replay Strategy for Cache Re-hydration

**Problem Scenario**: Total system crash, need to replay events

```
T0: System crash (power outage, region failure)
    → All pods terminated
    → Analytics cache (Redis) lost
    → NATS JetStream messages persisted

T1: System restart
    → Pods come back online
    → Analytics cache EMPTY
    → Need to re-hydrate from JetStream

PROBLEM:
→ billing_v2 domain: FASTag transactions already counted
→ Replay all messages → DOUBLE-COUNT transactions
→ Revenue reports show 2x actual
→ Customer invoices incorrect
```

### Recommended: Idempotent Replay with Domain-Specific Consumers

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    NATS JETSTREAM REPLAY STRATEGY                       │
│                    (Without Double-Counting)                            │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 1: STREAM CONFIGURATION (PRE-CRASH)                               │
│  ──────────────────────────────────────                                 │
│  Configure streams with proper retention and replay policy:            │
│                                                                         │
│  ```bash                                                                │
│  # Create TELEMETRY_STREAM (high volume, short retention)              │
│  nats stream add TELEMETRY_STREAM \                                    │
│    --subjects "telemetry.>" \                                          │
│    --retention limits \            # Delete when all consumers ACK     │
│    --max-msgs 100000000 \          # 100M messages max                 │
│    --max-bytes 50GB \              # 50GB storage limit                │
│    --discard old \                 # Drop oldest when full             │
│    --storage file \                # Persistent storage                │
│    --replicas 3                    # HA across 3 NATS nodes            │
│                                                                         │
│  # Create BILLING_STREAM (lower volume, longer retention)              │
│  nats stream add BILLING_STREAM \                                      │
│    --subjects "billing.>" \                                            │
│    --retention interest \          # Keep while consumers need         │
│    --max-msgs 10000000 \           # 10M messages max                  │
│    --max-bytes 10GB \              # 10GB storage limit                │
│    --discard fail \                # Don't drop, fail instead          │
│    --storage file \                                                      │
│    --replicas 3 \                                                        │
│    --allow-rollup                  # Enable message rollup             │
│                                                                         │
│  # Create ANALYTICS_CACHE_STREAM (for cache rebuild)                   │
│  nats stream add ANALYTICS_CACHE_STREAM \                              │
│    --subjects "analytics.events.>" \                                   │
│    --retention workqueue \         # Each consumer gets own copy       │
│    --max-msgs 50000000 \                                                 │
│    --max-bytes 25GB \                                                    │
│    --storage file \                                                      │
│    --replicas 3                                                          │
│  ```                                                                    │
│                                                                         │
│  Key Configuration Differences:                                         │
│  ┌──────────────────┬────────────┬────────────┬────────────────────┐   │
│  │ Stream           │ Retention  │ Replay     │ Use Case           │   │
│  ├──────────────────┼────────────┼────────────┼────────────────────┤   │
│  │ TELEMETRY_STREAM │ limits     │ Drop old   | GPS pings (lossy)  │   │
│  │ BILLING_STREAM   │ interest   │ Keep all   | Transactions       │   │
│  │ ANALYTICS_CACHE  │ workqueue  │ Per-consumer| Cache rebuild     │   │
│  └──────────────────┴────────────┴────────────┴────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 2: CONSUMER CONFIGURATION WITH ACK FLOOR                          │
│  ─────────────────────────────────────────────────                      │
│  Create durable consumers with ack tracking:                           │
│                                                                         │
│  ```bash                                                                │
│  # Billing consumer (idempotent processing)                            │
│  nats consumer add BILLING_STREAM BILLING_CONSUMER \                   │
│    --deliver all \                 # Start from beginning              │
│    --ack explicit \                # Require explicit ACK              │
│    --max-pending 1000 \            # Max unACKed messages              │
│    --max-deliver 5 \               # Retry 5 times before DLQ          │
│    --backoff linear \              # Linear backoff                    │
│    --filter-subject "billing.fastag.*" \                               │
│    --durable                       # Durable consumer (survives crash) │
│                                                                         │
│  # Analytics consumer (cache builder)                                  │
│  nats consumer add ANALYTICS_CACHE_STREAM ANALYTICS_CONSUMER \         │
│    --deliver all \                                                     │
│    --ack explicit \                                                    │
│    --max-pending 5000 \                                                │
│    --max-deliver 3 \                                                   │
│    --durable                                                           │
│  ```                                                                    │
│                                                                         │
│  Consumer State Tracking:                                               │
│  ```bash                                                                │
│  # Check consumer state (ack floor)                                    │
│  nats consumer info BILLING_STREAM BILLING_CONSUMER                    │
│                                                                         │
│  # Output:                                                              │
│  # Ack Floor: 1234567        ← Last ACKed message sequence             │
│  # Delivered: 1234580        ← Last delivered sequence                 │
│  # Pending: 13               ← UnACKed messages                        │
│  # Redelivered: 5            ← Messages being retried                  │
│  ```                                                                    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 3: IDEMPOTENT MESSAGE PROCESSING                                  │
│  ──────────────────────────────────────                                 │
│  Each domain implements idempotency checks:                            │
│                                                                         │
│  billing_v2 Domain (FASTag transactions):                               │
│  ```go                                                                  │
│  func (s *BillingService) ProcessFASTagTransaction(msg *nats.Msg) {    │
│      var event FASTagEvent                                             │
│      json.Unmarshal(msg.Data, &event)                                  │
│                                                                         │
│      // IDEMPOTENCY CHECK: Has this transaction been processed?        │
│      exists, err := s.repo.TransactionExists(event.TransactionID)      │
│      if err != nil {                                                   │
│          msg.Nak()  // Retry later                                     │
│          return                                                        │
│      }                                                                 │
│      if exists {                                                       │
│          // ALREADY PROCESSED - skip without error                    │
│          s.logger.Info("Transaction already processed, skipping",      │
│              zap.String("transaction_id", event.TransactionID))        │
│          msg.Ack()  // ACK to move ack floor forward                  │
│          return                                                        │
│      }                                                                 │
│                                                                         │
│      // Process transaction (first time)                               │
│      err = s.repo.CreateTransaction(event)                             │
│      if err != nil {                                                   │
│          msg.Nak()  // Retry                                           │
│          return                                                        │
│      }                                                                 │
│                                                                         │
│      // Update ledger (double-entry accounting)                        │
│      err = s.ledger.RecordTransaction(event)                           │
│      if err != nil {                                                   │
│          msg.Nak()  // Retry                                           │
│          return                                                        │
│      }                                                                 │
│                                                                         │
│      // ACK - mark as processed                                        │
│      msg.Ack()                                                         │
│  }                                                                      │
│  ```                                                                    │
│                                                                         │
│  analytics_v2 Domain (Cache events):                                    │
│  ```go                                                                  │
│  func (s *AnalyticsService) ProcessCacheEvent(msg *nats.Msg) {         │
│      var event AnalyticsEvent                                          │
│      json.Unmarshal(msg.Data, &event)                                  │
│                                                                         │
│      // IDEMPOTENCY CHECK: Check event deduplication table             │
│      processed, err := s.repo.IsEventProcessed(event.EventID)          │
│      if processed {                                                    │
│          msg.Ack()  // Already in cache, skip                         │
│          return                                                        │
│      }                                                                 │
│                                                                         │
│      // Update cache                                                    │
│      s.cache.Update(event)                                             │
│                                                                         │
│      // Mark event as processed (for idempotency)                      │
│      s.repo.MarkEventProcessed(event.EventID)                          │
│                                                                         │
│      msg.Ack()                                                         │
│  }                                                                      │
│  ```                                                                    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 4: CRASH RECOVERY AND REPLAY                                      │
│  ───────────────────────────────                                        │
│  After system crash, replay from ack floor:                            │
│                                                                         │
│  ```bash                                                                │
│  # 1. Check consumer state before restart                              │
│  nats consumer info BILLING_STREAM BILLING_CONSUMER \                  │
│    --json > /tmp/billing-consumer-state.json                           │
│                                                                         │
│  # Extract ack_floor from state                                        │
│  ACK_FLOOR=$(jq '.ack_floor' /tmp/billing-consumer-state.json)        │
│  # ACK_FLOOR = 1234567                                                 │
│                                                                         │
│  # 2. Restart consumers (they will resume from ack floor)              │
│  kubectl rollout restart deployment/fleet-core-service                │
│                                                                         │
│  # 3. Consumers automatically resume from ack floor                    │
│  #    - Messages 1 to 1234567: Already ACKed, NOT redelivered          │
│  #    - Messages 1234568+: Delivered for processing                   │
│  #    - Messages marked Redelivered: Being retried                     │
│                                                                         │
│  # 4. Monitor replay progress                                          │
│  watch 'nats consumer info BILLING_STREAM BILLING_CONSUMER'           │
│  ```                                                                    │
│                                                                         │
│  Consumer Restart Behavior:                                             │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ Message Sequence  │ Consumer Action                             │   │
│  ├─────────────────────────────────────────────────────────────────┤   │
│  │ 1 - 1234567       │ Already ACKed → SKIPPED (no double-count)  │   │
│  │ 1234568 - 1234580 │ UnACKed → Redelivered (idempotency check)  │   │
│  │ 1234581+          │ Pending → Delivered for first time         │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  STEP 5: ANALYTICS CACHE REBUILD                                        │
│  ───────────────────────────────                                        │
│  Rebuild analytics cache from replayed events:                         │
│                                                                         │
│  ```go                                                                  │
│  func (s *AnalyticsService) RebuildCacheFromJetStream() error {        │
│      // Connect to NATS                                                │
│      nc, _ := nats.Connect(natsURL)                                    │
│      js, _ := nc.JetStream()                                           │
│                                                                         │
│      // Subscribe to analytics cache stream                            │
│      sub, _ := js.SubscribeSync("analytics.events.>",                  │
│          nats.Durable("ANALYTICS_CONSUMER"),                           │
│          nats.DeliverAll(),                                            │
│      )                                                                 │
│                                                                         │
│      // Process messages until caught up                               │
│      for {                                                             │
│          msg, err := sub.NextMsg(1 * time.Second)                      │
│          if err == nats.ErrTimeout {                                   │
│              // No more pending messages - cache is rebuilt           │
│              break                                                     │
│          }                                                             │
│                                                                         │
│          var event AnalyticsEvent                                      │
│          json.Unmarshal(msg.Data, &event)                              │
│                                                                         │
│          // Idempotency check (same as normal processing)              │
│          if s.repo.IsEventProcessed(event.EventID) {                   │
│              msg.Ack()                                                 │
│              continue                                                  │
│          }                                                             │
│                                                                         │
│          // Update cache                                               │
│          s.cache.Update(event)                                         │
│          s.repo.MarkEventProcessed(event.EventID)                      │
│                                                                         │
│          msg.Ack()                                                     │
│      }                                                                 │
│                                                                         │
│      s.logger.Info("Analytics cache rebuilt from JetStream")          │
│      return nil                                                        │
│  }                                                                      │
│  ```                                                                    │
│                                                                         │
│  Cache Rebuild Progress:                                                │
│  ```bash                                                                │
│  # Monitor cache rebuild                                               │
│  redis-cli HGET analytics:cache:status last_updated                    │
│  redis-cli DBSIZE  # Should show expected key count                   │
│  ```                                                                    │
└─────────────────────────────────────────────────────────────────────────┘
```

### Idempotency Key Design

```go
// Idempotency key format for different event types
type IdempotencyKey struct {
    Domain      string    // billing, analytics, telemetry
    EventType   string    // fastag_transaction, gps_ping, alert_triggered
    EventID     string    // Unique event identifier
    Timestamp   time.Time // Event timestamp
    Checksum    string    // SHA256 of event payload
}

// Generate idempotency key
func GenerateIdempotencyKey(domain, eventType, eventID string, payload []byte) string {
    checksum := sha256.Sum256(payload)
    return fmt.Sprintf("%s:%s:%s:%x", domain, eventType, eventID, checksum[:8])
}

// Check if event already processed
func (s *BillingService) IsEventProcessed(key string) (bool, error) {
    // Check Redis first (fast path)
    exists, err := s.redis.Exists(context.Background(), "idem:"+key)
    if err != nil {
        return false, err
    }
    if exists {
        return true, nil
    }
    
    // Check database (source of truth)
    var count int
    err = s.db.QueryRow("SELECT COUNT(*) FROM idempotency_log WHERE key = $1", key).Scan(&count)
    return count > 0, err
}

// Mark event as processed
func (s *BillingService) MarkEventProcessed(key string, payload []byte) error {
    // Write to database (durable)
    _, err := s.db.Exec(`
        INSERT INTO idempotency_log (key, payload_hash, processed_at)
        VALUES ($1, $2, NOW())
        ON CONFLICT (key) DO NOTHING
    `, key, sha256.Sum256(payload))
    
    // Update Redis (fast path for future checks)
    s.redis.SetEx(context.Background(), "idem:"+key, "1", 7*24*time.Hour)
    
    return err
}
```

### Replay Safety Matrix

| Scenario | Without Idempotency | With Idempotency |
|----------|---------------------|------------------|
| **Normal replay** | Double-count all transactions | ✅ Skip already ACKed |
| **Partial ACK** | Some transactions counted twice | ✅ Idempotency check prevents |
| **Consumer restart** | Re-process all pending | ✅ Resume from ack floor |
| **Message redelivery** | Process same message N times | ✅ Process once, skip N-1 |

---

## 3️⃣ The 'Bihar' Failover: Multi-Region Disaster Recovery

### Current Terraform Configuration

**Location**: `infrastructure/terraform/main.tf`

```hcl
# CURRENT: Single-region deployment (no failover)
provider "aws" {
  region = var.aws_region  # Hardcoded to single region
}

# VPC Configuration (single region)
resource "aws_vpc" "flyfleet_vpc" {
  cidr_block = "10.0.0.0/16"
  # No cross-region replication
}
```

### ⚠️ CRITICAL GAP: No Multi-Region Failover

**Problem Scenario**: AWS Mumbai (ap-south-1) goes offline

```
T0: 2024-03-15 14:00:00 - Normal operation
    → Primary: ap-south-1 (Mumbai)
    → 50,000 active vehicles tracking
    → 100K GPS pings/minute

T1: 2024-03-15 14:05:00 - Mumbai region failure
    → AWS ap-south-1 outage
    → All pods unreachable
    → Database unavailable
    → NATS cluster offline

T2: 2024-03-15 14:06:00 - Monitoring alerts
    → Health checks failing
    → On-call paged

PROBLEM:
→ No secondary region configured
→ No automated failover
→ Manual DNS switchover required
→ Database has no cross-region replica
→ RTO: 4+ hours (manual rebuild)
→ RPO: Last backup (4 hours of data loss)
```

### Recommended: Active-Active Multi-Region Failover

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    MULTI-REGION FAILOVER ARCHITECTURE                   │
│                    (Mumbai → Hyderabad Active-Active)                   │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│  REGION TOPOLOGY                                                        │
│  ─────────────────                                                      │
│                                                                         │
│  ┌─────────────────────────┐         ┌─────────────────────────┐       │
│  │  PRIMARY: ap-south-1    │         │  SECONDARY: ap-south-2  │       │
│  │  (Mumbai)               │         │  (Hyderabad)            │       │
│  │                         │         │                         │       │
│  │  ┌───────────────────┐  │         │  ┌───────────────────┐  │       │
│  │  │  EKS Cluster      │  │         │  │  EKS Cluster      │  │       │
│  │  │  - fleet-core ×3  │  │         │  │  - fleet-core ×3  │  │       │
│  │  │  - web-bff ×2     │  │         │  │  - web-bff ×2     │  │       │
│  │  │  - NATS ×3        │◄─┼─────────┼─►│  - NATS ×3        │  │       │
│  │  └───────────────────┘  │         │  └───────────────────┘  │       │
│  │                         │         │                         │       │
│  │  ┌───────────────────┐  │         │  ┌───────────────────┐  │       │
│  │  │  RDS PostgreSQL   │  │         │  │  RDS PostgreSQL   │  │       │
│  │  │  - Primary        │──┼─────────┼─►│  - Read Replica   │  │       │
│  │  │  (ap-south-1a)    │  │         │  │  (ap-south-2a)    │  │       │
│  │  └───────────────────┘  │         │  └───────────────────┘  │       │
│  │                         │         │                         │       │
│  │  ┌───────────────────┐  │         │  ┌───────────────────┐  │       │
│  │  │  ElastiCache      │  │         │  │  ElastiCache      │  │       │
│  │  │  Redis Cluster    │◄─┼─────────┼─►│  Redis Cluster    │  │       │
│  │  └───────────────────┘  │         │  └───────────────────┘  │       │
│  └─────────────────────────┘         └─────────────────────────┘       │
│              │                                     │                    │
│              └──────────────┬──────────────────────┘                    │
│                             │                                           │
│                    ┌────────▼────────┐                                  │
│                    │  Route 53       │                                  │
│                    │  Health Checks  │                                  │
│                    │  + Failover     │                                  │
│                    └────────┬────────┘                                  │
│                             │                                           │
│                    ┌────────▼────────┐                                  │
│                    │  Global         │                                  │
│                    │  Accelerator    │                                  │
│                    └─────────────────┘                                  │
└─────────────────────────────────────────────────────────────────────────┘
```

### Terraform Multi-Region Configuration

```hcl
# infrastructure/terraform/main.tf (ENHANCED)

# Multi-region provider configuration
provider "aws" {
  alias  = "primary"
  region = var.primary_region  # ap-south-1 (Mumbai)
}

provider "aws" {
  alias  = "secondary"
  region = var.secondary_region  # ap-south-2 (Hyderabad)
}

# Primary region VPC
resource "aws_vpc" "primary_vpc" {
  provider = aws.primary
  cidr_block = "10.0.0.0/16"
  
  tags = {
    Name = "flyfleet-vpc-primary"
    Environment = var.environment
    Region = "primary"
  }
}

# Secondary region VPC
resource "aws_vpc" "secondary_vpc" {
  provider = aws.secondary
  cidr_block = "10.1.0.0/16"  # Different CIDR to avoid overlap
  
  tags = {
    Name = "flyfleet-vpc-secondary"
    Environment = var.environment
    Region = "secondary"
  }
}

# VPC Peering for cross-region communication
resource "aws_vpc_peering_connection" "cross_region" {
  provider  = aws.primary
  vpc_id    = aws_vpc.primary_vpc.id
  peer_vpc_id = aws_vpc.secondary_vpc.id
  peer_region = var.secondary_region
  
  tags = {
    Name = "flyfleet-vpc-peering"
  }
}

# Primary RDS with cross-region read replica
resource "aws_db_instance" "primary_db" {
  provider = aws.primary
  
  identifier = "flyfleet-db-primary"
  engine     = "postgres"
  engine_version = "15.4"
  instance_class = "db.r5.2xlarge"
  
  # Multi-AZ for high availability within region
  multi_az = true
  
  # Enable automated backups
  backup_retention_period = 7
  backup_window = "03:00-04:00"
  
  # Enable cross-region read replica
  publicly_accessible = false
  
  tags = {
    Name = "flyfleet-db-primary"
    Region = "primary"
  }
}

# Secondary region read replica (async replication)
resource "aws_db_instance" "secondary_replica" {
  provider = aws.secondary
  
  identifier = "flyfleet-db-replica"
  replicate_source_db = aws_db_instance.primary_db.arn
  
  # Read replica in secondary region
  instance_class = "db.r5.xlarge"  # Smaller for read-only
  
  # Enable promotion for failover
  skip_final_snapshot = true
  
  tags = {
    Name = "flyfleet-db-replica"
    Region = "secondary"
  }
}

# Route 53 health check for primary region
resource "aws_route53_health_check" "primary_health" {
  fqdn              = aws_lb.primary_lb.dns_name
  port              = 443
  type              = "HTTPS"
  resource_path     = "/health"
  failure_threshold = 3
  request_interval  = 10
  
  tags = {
    Name = "flyfleet-primary-health"
  }
}

# Route 53 failover record
resource "aws_route53_record" "api_failover" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "api.flyfleet.com"
  type    = "A"
  
  failover_routing_policy {
    type = "PRIMARY"
    health_check_id = aws_route53_health_check.primary_health.id
  }
  
  alias {
    name                   = aws_lb.primary_lb.dns_name
    zone_id                = aws_lb.primary_lb.zone_id
    evaluate_target_health = true
  }
}

# Secondary region failover record
resource "aws_route53_record" "api_failover_secondary" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "api.flyfleet.com"
  type    = "A"
  
  failover_routing_policy {
    type = "SECONDARY"
  }
  
  alias {
    name                   = aws_lb.secondary_lb.dns_name
    zone_id                = aws_lb.secondary_lb.zone_id
    evaluate_target_health = true
  }
}
```

### Automated Failover Sequence

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    AUTOMATED FAILOVER SEQUENCE                          │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│  T0: NORMAL OPERATION                                                   │
│  ───────────────────                                                    │
│  Primary Region (ap-south-1):                                           │
│  → Route 53: api.flyfleet.com → Primary ALB                           │
│  → Database: Primary RDS (read-write)                                  │
│  → Cache: Primary Redis Cluster                                        │
│  → NATS: Primary JetStream Cluster                                     │
│                                                                         │
│  Secondary Region (ap-south-2):                                         │
│  → Route 53: STANDBY (not receiving traffic)                           │
│  → Database: Read Replica (async replication lag < 1s)                │
│  → Cache: Warm standby (replicated from primary)                       │
│  → NATS: Standby cluster (sync via NATS leaf nodes)                    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ Primary region failure detected
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  T+0s: FAILURE DETECTION                                                │
│  ─────────────────────                                                  │
│  Route 53 Health Check:                                                 │
│  → GET https://api.flyfleet.com/health                                │
│  → Response: TIMEOUT (no response from primary)                       │
│  → Retry 1 (10s later): TIMEOUT                                       │
│  → Retry 2 (10s later): TIMEOUT                                       │
│  → Retry 3 (10s later): TIMEOUT                                       │
│                                                                         │
│  Health Check Status: UNHEALTHY (3 consecutive failures)               │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ Route 53 automatic failover
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  T+30s: DNS FAILOVER                                                    │
│  ─────────────────                                                      │
│  Route 53 Failover Routing:                                             │
│  → Detect primary health check UNHEALTHY                              │
│  → Switch DNS record to SECONDARY                                     │
│  → TTL: 60 seconds (clients refresh within 60s)                       │
│                                                                         │
│  Global Accelerator (optional):                                         │
│  → Detect endpoint unhealthy                                          │
│  → Reroute traffic to secondary endpoint                              │
│  → Faster than DNS (no TTL delay)                                     │
│                                                                         │
│  New Traffic Flow:                                                      │
│  → api.flyfleet.com → Secondary ALB (ap-south-2)                      │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ Database promotion
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  T+60s: DATABASE PROMOTION                                              │
│  ─────────────────────                                                  │
│  Promote read replica to primary:                                       │
│  ```bash                                                                │
│  aws rds promote-read-replica \                                        │
│    --db-instance-identifier flyfleet-db-replica \                      │
│    --region ap-south-2                                                 │
│  ```                                                                    │
│                                                                         │
│  Promotion Process:                                                     │
│  → Stop replication from primary                                       │
│  → Promote to read-write                                               │
│  → Update security groups (allow app connections)                      │
│  → Update Parameter Group (enable write operations)                    │
│  → Duration: ~2-3 minutes                                              │
│                                                                         │
│  Connection String Update:                                              │
│  → Old: flyfleet-db-primary.xxx.ap-south-1.rds.amazonaws.com          │
│  → New: flyfleet-db-replica.yyy.ap-south-2.rds.amazonaws.com          │
│                                                                         │
│  Update Secrets Manager / Parameter Store:                              │
│  ```bash                                                                │
│  aws secretsmanager update-secret \                                    │
│    --secret-id flyfleet/db/connection \                                │
│    --secret-string '{"host":"flyfleet-db-replica.yyy.ap-south-2..."}' │
│  ```                                                                    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ Application reconfiguration
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  T+180s: APPLICATION FAILOVER                                           │
│  ──────────────────────────                                             │
│  Kubernetes Deployment Update:                                          │
│  ```bash                                                                │
│  # Update deployment to use secondary database                         │
│  kubectl set env deployment/fleet-core-service \                       │
│    DATABASE_URL=$(aws secretsmanager get-secret-value \                │
│      --secret-id flyfleet/db/connection \                              │
│      --query SecretString --output text) \                             │
│    --namespace flyfleet                                                │
│                                                                         │
│  # Rollout restart to pick up new config                               │
│  kubectl rollout restart deployment/fleet-core-service \               │
│    --namespace flyfleet                                                │
│  ```                                                                    │
│                                                                         │
│  NATS Cluster Reconfiguration:                                          │
│  ```bash                                                                │
│  # Update NATS config to use secondary cluster                         │
│  kubectl patch configmap nats-config \                                 │
│    --patch '{"data":{"nats.conf":"jetstream { store_dir: /data }"}}'  │
│                                                                         │
│  # Restart NATS pods                                                   │
│  kubectl rollout restart statefulset/nats                              │
│  ```                                                                    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ Verification
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  T+300s: FAILOVER COMPLETE                                              │
│  ───────────────────────                                                │
│  Verification Checks:                                                   │
│  ```bash                                                                │
│  # 1. Verify DNS resolution                                             │
│  dig api.flyfleet.com +short                                          │
│  # Should resolve to secondary ALB IP                                  │
│                                                                         │
│  # 2. Verify database is writable                                       │
│  psql -h flyfleet-db-replica.yyy.ap-south-2.rds.amazonaws.com \       │
│    -c "INSERT INTO failover_test (timestamp) VALUES (NOW())"          │
│  # Should succeed (write confirmed)                                    │
│                                                                         │
│  # 3. Verify application health                                         │
│  curl https://api.flyfleet.com/health                                 │
│  # Should return 200 OK                                                │
│                                                                         │
│  # 4. Verify GPS tracking                                               │
│  curl -H "Authorization: Bearer $TOKEN" \                             │
│    https://api.flyfleet.com/api/v1/tracking/vehicles/active           │
│  # Should return active vehicles                                       │
│  ```                                                                    │
│                                                                         │
│  Failover Complete:                                                     │
│  → All traffic routing to ap-south-2 (Hyderabad)                       │
│  → Database accepting writes in secondary                              │
│  → Application pods healthy                                            │
│  → GPS tracking resumed                                                │
│                                                                         │
│  RTO: 5 minutes (from detection to verification)                       │
│  RPO: < 1 second (async replication lag)                               │
└─────────────────────────────────────────────────────────────────────────┘
```

### Failover Runbook

| Step | Action | Duration | Owner |
|------|--------|----------|-------|
| **T+0s** | Route 53 health check fails | 30s | Automated |
| **T+30s** | DNS failover to secondary | 60s (TTL) | Automated |
| **T+90s** | RDS replica promotion initiated | 2-3 min | Automated |
| **T+180s** | Database promotion complete | - | Automated |
| **T+240s** | Application config update | 1 min | Automated |
| **T+300s** | Verification complete | - | On-call |

### Active-Active Database Sync

```sql
-- PostgreSQL logical replication for active-active sync
-- Primary (ap-south-1)
CREATE PUBLICATION flyfleet_publication FOR ALL TABLES;

-- Secondary (ap-south-2)
CREATE SUBSCRIPTION flyfleet_subscription
  CONNECTION 'host=flyfleet-db-primary.ap-south-1.rds.amazonaws.com dbname=fleet_db'
  PUBLICATION flyfleet_publication
  WITH (copy_data = true, synchronous_commit = off);

-- For true active-active (bidirectional sync)
-- Secondary also publishes, Primary also subscribes
-- Conflict resolution: Last-write-wins based on updated_at timestamp

-- Conflict resolution function
CREATE OR REPLACE FUNCTION resolve_conflict()
RETURNS TRIGGER AS $$
BEGIN
  -- Last-write-wins based on updated_at
  IF NEW.updated_at > OLD.updated_at THEN
    RETURN NEW;
  ELSE
    RETURN OLD;
  END IF;
END;
$$ LANGUAGE plpgsql;

-- Apply to all tables
CREATE TRIGGER trg_resolve_conflict
  BEFORE UPDATE ON trips
  FOR EACH ROW
  EXECUTE FUNCTION resolve_conflict();
```

---

## Summary: Production-Ready Configuration

### 1. PITR Workflow

```bash
# PITR settings
PITR_ENABLED=true
PITR_TEMP_CLUSTER_PREFIX=flyfleet-pitr-restore
PITR_LIVE_DATA_CAPTURE=true
PITR_MERGE_CONFLICT_RESOLUTION=live_wins
```

### 2. NATS JetStream Replay

```bash
# JetStream settings
JETSTREAM_ENABLED=true
JETSTREAM_STREAM_RETENTION=interest  # billing, analytics
JETSTREAM_CONSUMER_DURABLE=true
JETSTREAM_IDEMPOTENCY_CHECK=true
JETSTREAM_ACK_TRACKING=true
```

### 3. Multi-Region Failover

```bash
# Failover settings
FAILOVER_ENABLED=true
FAILOVER_PRIMARY_REGION=ap-south-1
FAILOVER_SECONDARY_REGION=ap-south-2
FAILOVER_HEALTH_CHECK_INTERVAL=10s
FAILOVER_DNS_TTL=60s
FAILOVER_AUTO_PROMOTE_REPLICA=true
FAILOVER_RTO_TARGET=5m
FAILOVER_RPO_TARGET=1s
```

---

## Conclusion

The current implementation has **three critical gaps**:

1. **No PITR workflow** → Cannot restore deleted data without overwriting live GPS
2. **No idempotent replay** → Risk of double-counting billing transactions after crash
3. **No multi-region failover** → 4+ hour RTO if Mumbai region fails

**Recommended immediate actions**:

1. Implement PITR workflow with live data preservation
2. Configure JetStream with idempotent consumers and ack tracking
3. Set up active-active multi-region deployment with automated failover

These changes ensure **zero data loss recovery**, **idempotent replay**, and **5-minute RTO** for regional failures.
