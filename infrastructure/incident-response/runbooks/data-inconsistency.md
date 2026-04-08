# Runbook: Data Inconsistency

**Severity:** SEV1 (data corruption) or SEV2 (stale/missing data)
**Services:** Database (PostgreSQL), Caches (Redis), Telemetry Ingestion
**SLO Violation:** Data integrity, accuracy requirements
**Alert:** Manual detection (no automated alert typically)

---

## Quick Reference

| What | How |
|------|-----|
| **Check data sync** | Compare counts: `kubectl exec -it postgres-0 -- psql -c "SELECT COUNT(*) FROM events;"` vs API count |
| **Check Redis consistency** | Compare cached vs DB: `redis-cli GET fleet:vehicle:123` vs DB record |
| **Check sequence gaps** | `kubectl exec -it postgres-0 -- psql -c "SELECT MAX(id) - COUNT(*) FROM events;"` |
| **Force cache refresh** | `redis-cli DEL fleet:vehicle:123` |
| **Rebuild cache** | Run cache warming script: `./scripts/warm-cache.sh` |

---

## Symptoms

- Same data showing differently in dashboard vs API
- Missing telemetry points despite device reporting sent
- Duplicate records in database
- Stale data in UI (not updating)
- Cache and DB out of sync (Redis shows old values)
- Sequence number gaps in event timeline
- Checksum/hash mismatches in data exports
- Financial calculations incorrect (billing discrepancies)

---

## Decision Tree

```
Data Inconsistent?
├─ Specific record or widespread?
│   ├─ Single record? → Cache stale?
│   │   ├─ Yes → Clear Redis key for that record
│   │   └─ No → Database record corrupted? Check DB logs
│   └─ Widespread?
│       ├─ Cache/DB mismatch? → Cache invalidation issue
│       │   ├─ Write-through not working? → Check event publishing
│       │   └─ Cache TTL too long? → Reduce TTL
│       ├─ Duplicate records? → Check for retry loops, missing idempotency
│       └─ Missing records? → Ingestion failure, check NATS backlog
├─ Stale data in dashboard?
│   ├─ API returns correct data? → Cache problem, warm/refresh cache
│   └─ API also stale? → Database problem, check replication
└─ Sequence gaps?
    ├─ NATS messages dropped? → NATS runbook
    └─ Database insert failure? → Check DB constraints, deadlocks
```

---

## Investigation Steps

### 1. Define the Inconsistency (2 min)

**What exactly is inconsistent?**
- [ ] Cache vs Database mismatch
- [ ] API vs Database mismatch
- [ ] Different microservices showing different data
- [ ] Duplicate records in same table
- [ ] Missing records expected to exist
- [ ] Sequence gaps in time-series data
- [ ] Calculated fields wrong (totals, sums)
- [ ] Cross-tenant data leakage

**Example scenario:** Dashboard shows vehicle location from 5 minutes ago, but API returns current location.

---

### 2. Identify Affected Data Domain (2 min)

Which tables/collections are affected?

| Domain | Tables | Typical Issues |
|--------|--------|----------------|
| **Vehicles** | vehicles, vehicle_telemetry | Stale locations, duplicate positions |
| **Drivers** | drivers, driver_locations | Old location, missing status |
| **Billing** | invoices, payments, trips | Amount mismatch, missing trips |
| **Compliance** | india_fastag, india_eway_bills | Missing records, wrong status |
| **Telemetry** | telemetry_events | Sequence gaps, duplicates |
| **Auth** | users, orgs, roles | Permission mismatches |

**Action:** Query affected tables directly to get ground truth.

---

### 3. Compare Data Between Systems (3 min)

```bash
# Case: Dashboard stale but API fresh (cache issue)
# Get data from API
curl -s http://localhost:8080/api/v1/fleet/vehicles/123 | jq '.location.updated_at' > api_time.txt

# Get from database
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT updated_at FROM vehicles WHERE id = '123';" > db_time.txt

# Get from Redis
kubectl exec -it redis-0 -n flyfleet -- redis-cli GET fleet:vehicle:123 > redis_val.txt
kubectl exec -it redis-0 -n flyfleet -- redis-cli TTL fleet:vehicle:123 > redis_ttl.txt

# Compare timestamps
echo "API: $(cat api_time.txt)"
echo "DB:  $(cat db_time.txt)"
echo "Redis TTL: $(cat redis_ttl.txt)"
```

**Expected:** API and DB match (within seconds). Redis may be stale if cache not invalidated.

---

### 4. Check Cache Invalidation Path (3 min)

If cache is stale, check event flow:

```bash
# Check if vehicle update events are being published
kubectl exec -it nats-0 -n flyfleet -- \
  nats stream report TELEMETRY | grep "Subject: vehicle.update"

# Check fleet-core logs for cache invalidation
kubectl logs -f deployment/fleet-core -n flyfleet --since=10m 2>&1 | grep -i "redis\|cache\|publish"

# Check Redis for recent updates
kubectl exec -it redis-0 -n flyfleet -- redis-cli --scan --pattern "fleet:vehicle:*" | head -10 | xargs -I {} kubectl exec -it redis-0 -n flyfleet -- redis-cli TTL {}

# Verify cache TTL reasonable (e.g., 300s = 5min). Too long = stale data.
```

**Root cause:**
- Event not published → fleet-core bug
- Event published but consumer not processing → consumer lag
- Consumer processing but cache set with wrong TTL → config bug
- Cache invalidation skipped → idempotency check too aggressive

---

### 5. Check for Duplicate Records (3 min)

```bash
# Find duplicate vehicle records by VIN or device_id
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT device_id, COUNT(*) as count
    FROM vehicles
    GROUP BY device_id
    HAVING COUNT(*) > 1
    ORDER BY count DESC
    LIMIT 10;
  "

# Find duplicate telemetry (same device, same timestamp ±1s)
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT device_id, received_at, COUNT(*) as count
    FROM telemetry_events
    WHERE received_at > NOW() - INTERVAL '1 hour'
    GROUP BY device_id, received_at
    HAVING COUNT(*) > 1
    ORDER BY count DESC
    LIMIT 10;
  "

# Check for duplicate billing entries (same trip, same amount)
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT trip_id, COUNT(*) as count
    FROM trip_billing_entries
    GROUP BY trip_id
    HAVING COUNT(*) > 1
    ORDER BY count DESC
    LIMIT 10;
  "
```

**If duplicates found:**
- Identify source: ingestion retry without deduplication
- Check for missing unique constraints on device_id + timestamp

---

### 6. Check Sequence Gaps (2 min)

For time-series data (telemetry):

```bash
# Check if sequence numbers have gaps
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT device_id,
           MAX(sequence_number) - COUNT(*) as gaps,
           MIN(sequence_number) as min_seq,
           MAX(sequence_number) as max_seq,
           COUNT(*) as count
    FROM telemetry_events
    WHERE received_at > NOW() - INTERVAL '1 hour'
    GROUP BY device_id
    HAVING MAX(sequence_number) - COUNT(*) > 0
    ORDER BY gaps DESC
    LIMIT 10;
  "

# Check NATS sequence numbers vs DB
kubectl exec -it nats-0 -n flyfleet -- nats stream report TELEMETRY | grep "FirstSeq\|LastSeq"
# Compare total messages in stream vs count in DB
```

**If gaps:** Messages dropped in NATS or ingestion failed before DB insert.

---

### 7. Check for Cross-Tenant Data Leakage (2 min)

**Severity:** Security incident if data visible across orgs

```bash
# Check if any records have wrong org_id
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT COUNT(*) FROM vehicles WHERE org_id IS NULL;
  "

# Check for vehicles from one org appearing in another's queries
ORG1="org_123"
ORG2="org_456"
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT COUNT(*) FROM vehicles
    WHERE org_id = '$ORG1'
      AND id IN (SELECT vehicle_id FROM vehicle_assignments WHERE org_id = '$ORG2');
  "

# Review recent queries for missing org_id filter
kubectl logs -f deployment/fleet-core -n flyfleet --since=1h 2>&1 | grep -E "SELECT.*FROM.*WHERE.*org_id" | wc -l
```

**If leakage found:** This is a SEV1 security incident. Notify Security Officer immediately.

---

## Mitigation Strategies

### Strategy 1: Fix Cache Invalidation (Cache Stale)

```bash
# Manual cache refresh for affected keys
kubectl exec -it redis-0 -n flyfleet -- \
  redis-cli --scan --pattern "fleet:vehicle:*" | xargs -L 1000 redis-cli DEL

# OR for specific records
kubectl exec -it redis-0 -n flyfleet -- redis-cli SET fleet:vehicle:123 "$(get_fresh_from_db 123)"

# Restart fleet-core to clear in-memory cache (if any)
kubectl rollout restart deployment/fleet-core -n flyfleet

# Verify cache repopulated
kubectl exec -it redis-0 -n flyfleet -- redis-cli EXISTS fleet:vehicle:123
```

**Long-term fix:**
- Publish vehicle update events on every write
- Ensure cache consumer idempotent
- Set appropriate TTL (300s) as fallback

---

### Strategy 2: Fix Duplicate Records

```bash
# Identify duplicates with window function
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    WITH duplicates AS (
      SELECT id,
             ROW_NUMBER() OVER (PARTITION BY device_id, received_at ORDER BY created_at) as rn
      FROM telemetry_events
      WHERE received_at > NOW() - INTERVAL '1 hour'
    )
    SELECT COUNT(*) FROM duplicates WHERE rn > 1;
  "

# Delete duplicates, keep earliest
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    WITH duplicates AS (
      SELECT id,
             ROW_NUMBER() OVER (PARTITION BY device_id, received_at ORDER BY created_at) as rn
      FROM telemetry_events
      WHERE received_at > NOW() - INTERVAL '1 hour'
    )
    DELETE FROM telemetry_events
    WHERE id IN (SELECT id FROM duplicates WHERE rn > 1);
  "

# Add unique constraint to prevent future duplicates
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    ALTER TABLE telemetry_events
    ADD CONSTRAINT unique_telemetry_per_device_time
    UNIQUE (device_id, received_at);
  "
```

**Caution:** Deleting data. Ensure duplicates are truly redundant first.

---

### Strategy 3: Replay Missing Events (Sequence Gaps)

If NATS messages were dropped but source data available:

```bash
# Get first and last sequence numbers expected
kubectl exec -it nats-0 -n flyfleet -- nats stream report TELEMETRY

# Check gap size
# If small (< 1000), may be acceptable network re-delivery

# For large gaps, need to re-ingest from device history (if stored)
# Or request devices to re-send telemetry (push notification)

# Publish replay command to devices
kubectl exec -it nats-0 -n flyfleet -- \
  nats pub commands "{\"command\":\"resync_telemetry\",\"since\":\"2024-01-01T00:00:00Z\"}" --subject "device.command.resync"
```

---

### Strategy 4: Fix Cross-Tenant Leakage (SEV1)

If data from one org visible to another:

```bash
# 1. IMMEDIATE: Identify all leaked records
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT v.id, v.device_id, v.org_id as actual_org, va.org_id as query_org
    FROM vehicles v
    JOIN vehicle_assignments va ON v.id = va.vehicle_id
    WHERE v.org_id != va.org_id;
  " > leaks.csv

# 2. ISOLATE: Block all queries until fix deployed
# Implement temporary hard enforcement in repository layer
# Add org_id check to EVERY query if missing

# 3. ROOT CAUSE: Find code path with missing org_id filter
kubectl logs -f deployment/fleet-core -n flyfleet --since=24h 2>&1 | grep -v "org_id" | grep "SELECT.*FROM" > suspect_queries.txt

# 4. FIX: Add org_id to all queries (code fix required)
# See: internal/domains/*/repository/*.go

# 5. VALIDATE after fix: Run the leak detection query again, expect 0 rows
```

---

### Strategy 5: Database Consistency Check

Run comprehensive consistency check:

```bash
# Check for orphaned records (violating FK constraints)
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT COUNT(*) FROM vehicles v
    LEFT JOIN organizations o ON v.org_id = o.id
    WHERE o.id IS NULL;
  "

# Check for missing driver assignments
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT COUNT(*) FROM driver_assignments da
    LEFT JOIN drivers d ON da.driver_id = d.id
    WHERE d.id IS NULL;
  "

# Check data freshness (last updated)
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT
      COUNT(*) FILTER (WHERE updated_at < NOW() - INTERVAL '1 hour') as stale_1h,
      COUNT(*) FILTER (WHERE updated_at < NOW() - INTERVAL '6 hours') as stale_6h,
      COUNT(*) as total
    FROM vehicles;
  "
```

---

## Recovery & Validation

### Validation Steps

1. **Cache Consistency**
```bash
# Sample check across multiple records
for id in 1 2 3 4 5; do
  api_time=$(curl -s http://localhost:8080/api/v1/fleet/vehicles/$id | jq -r '.updated_at')
  db_time=$(kubectl exec -it postgres-0 -n flyfleet -- psql -U fleet_user fleet_db -c "SELECT updated_at FROM vehicles WHERE id = '$id';" -t -A)
  [ "$api_time" = "$db_time" ] && echo "Vehicle $id: OK" || echo "Vehicle $id: MISMATCH (API:$api_time DB:$db_time)"
done
```

2. **No Duplicates**
```bash
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT COUNT(*) FROM (
      SELECT device_id, received_at, COUNT(*)
      FROM telemetry_events
      GROUP BY device_id, received_at
      HAVING COUNT(*) > 1
    ) dupes;
  "
# Should return 0
```

3. **No Sequence Gaps**
```bash
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT SUM(gaps) FROM (
      SELECT device_id, MAX(sequence_number) - COUNT(*) as gaps
      FROM telemetry_events
      WHERE received_at > NOW() - INTERVAL '1 hour'
      GROUP BY device_id
    ) per_device;
  "
# Should return 0 or very low number (tolerate small gaps due to network)
```

4. **No Cross-Tenant Leakage**
```bash
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT COUNT(*) FROM vehicles v
    JOIN vehicle_assignments va ON v.id = va.vehicle_id
    WHERE v.org_id != va.org_id;
  "
# Should return 0
```

5. **Data Freshness**
```bash
# Most vehicles updated within last 5 minutes
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT COUNT(*) FROM vehicles
    WHERE updated_at > NOW() - INTERVAL '5 minutes';
  "
# Should match total vehicles count (all recent)
```

---

## Post-Incident Analysis

### PIR Focus Areas

1. **Root Cause Categories:**
   - Cache invalidation bug (events not published or consumed)
   - Missing `org_id` enforcement (SQL queries)
   - Duplicate ingestion (no deduplication, retry loops)
   - NATS message loss or consumer rebalancing issues
   - Database transaction isolation level too low (dirty reads)
   - Clock skew between services causing timestamp mismatches
   - Idempotency key missing on writes (allows duplicates)

2. **Data Loss Assessment:**
   - Count affected records: how many records were inconsistent?
   - Identify records that were lost vs. duplicated vs. stale
   - Can data be reconstructed? From device history, backup, or external source?
   - Was there financial impact (billing errors)?

3. **Preventive Actions:**
   - Add unique constraints on critical tables (device_id + timestamp)
   - Implement idempotency keys for all writes (replay-safe)
   - Improve cache invalidation: publish events on every write, consumers must handle
   - Add data consistency monitoring (daily audit job that checks cache vs DB, duplicate detection)
   - Enforce `org_id` in all repository queries (code review check)
   - Implement optimistic locking with version numbers (detect concurrent updates)
   - Add sequence numbers to events, detect gaps in ingestion
   - Implement data reconciliation jobs (periodic cache warming, duplicate detection)

4. **Immediate Remediation:**
   - Rebuild cache from database ground truth
   - Delete duplicate records, keep latest (or earliest based on business rule)
   - Backfill missing data if sources available
   - Notify affected customers if their data was incorrect

---

## Prevention & Monitoring

### Data Consistency Audits

Create daily cron job to detect inconsistencies:

```bash
#!/bin/bash
# infrastructure/incident-response/scripts/consistency-audit.sh

# Check cache vs DB mismatch
# Check duplicate records
# Check sequence gaps
# Report to Slack if issues found
```

Add to cron: `0 2 * * * /path/to/consistency-audit.sh`

### Essential Alerts

```yaml
- alert: DataDuplicationDetected
  expr: count_over_time(duplicate_records_total[1h]) > 0
  for: 5m

- alert: CacheDBMismatch
  expr: cache_db_mismatch_count > 0
  for: 10m

- alert: SequenceGapDetected
  expr: telemetry_sequence_gaps_total > 1000
  for: 5m

- alert: CrossTenantDataLeakage
  expr: cross_tenant_leakage_count > 0
  for: 1m  # SEV1 - security incident
```

---

## When to Declare SEV1

Treat as SEV1 if:
- Financial data corruption (invoices, payments)
- Cross-tenant data leakage (security)
- Mass data loss (> 1 hour of telemetry)
- Regulatory reporting affected (India compliance)

Immediate actions:
1. Stop writes (maintenance mode)
2. Preserve all logs for forensic analysis
3. Notify Security Officer and Legal
4. Engage incident commander
5. Begin data recovery from backup

---

## Related Runbooks

- [Database Performance](database-performance.md) - If DB is source of inconsistency
- [Telemetry Disruption](telemetry-disruption.md) - Telemetry data gaps
- [Security Incident](security-incident.md) - Cross-tenant leakage

---

**Last Updated:** 2026-03-29
**Review Date:** 2026-06-29
**Owner:** Platform Engineering Team
