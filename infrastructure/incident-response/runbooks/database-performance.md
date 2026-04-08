# Runbook: Database Performance (PostgreSQL)

**Severity:** SEV1 (if down) or SEV2 (if slow/inconsistent)
**Service:** PostgreSQL (fleet_db)
**SLO Violation:** Availability < 99.99%, P95 Query Time > 100ms
**Alert:** `PostgreSQLDown` or `PostgreSQLQueryLatencyHigh`

---

## Quick Reference

| What | How |
|------|-----|
| **Check DB uptime** | `kubectl exec -it postgres-0 -n flyfleet -- pg_isready` |
| **Check connections** | `kubectl exec -it postgres-0 -n flyfleet -- psql -c "SELECT count(*) FROM pg_stat_activity;"` |
| **Check slow queries** | `kubectl exec -it postgres-0 -n flyfleet -- psql -c "SELECT * FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"` |
| **Check table locks** | `kubectl exec -it postgres-0 -n flyfleet -- psql -c "SELECT * FROM pg_locks WHERE NOT granted;"` |
| **Check disk space** | `kubectl exec -it postgres-0 -n flyfleet -- df -h /var/lib/postgresql/data` |
| **Restart DB** | `kubectl rollout restart statefulset/postgres -n flyfleet` ⚠️ Use with caution |

---

## Symptoms

- API requests timing out (> 30s)
- 5xx errors with "context deadline exceeded" or "too many connections"
- High pod CPU/memory on postgres-0
- Queries taking > 1 second (normally < 100ms)
- Connection pool exhaustion (all connections used)
- Table locks causing contention
- Disk full errors ("no space left on device")
- Replication lag (if using read replicas)

---

## Decision Tree

```
DB Slow or Down?
├─ DB Down (no connection)?
│   ├─ Pod CrashLoopBackOff? → Check logs, increased resources
│   ├─ Pod Pending? → Insufficient resources, check node
│   ├─ Disk Full? → Cleanup logs/backups, expand PVC
│   └─ Network Issue? → Check service, security groups
├─ DB Slow?
│   ├─ High CPU? → Check long-running queries, kill if needed
│   ├─ High connections? → Check pool leaks, increase max_connections
│   ├─ Disk I/O high? → Check pg_stat_statements, add indexes
│   ├─ Locks? → Identify blocking query, terminate if needed
│   └─ Low memory? → Increase shared_buffers or pod memory
└─ Specific queries slow?
    ├─ Missing index? → EXPLAIN ANALYZE, add index
    ├─ Table bloat? → VACUUM FULL or pg_repack
    └─ Parameter sniffing? → Use prepared statements, plan caching
```

---

## Investigation Steps

### 1. Check Database Health (2 min)

```bash
# Pod status
kubectl get pods -n flyfleet -l app=postgres

# Database readiness
kubectl exec -it postgres-0 -n flyfleet -- pg_isready

# Current connections count
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT count(*) FROM pg_stat_activity;"

# Compare to max_connections
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SHOW max_connections;"

# Status of autovacuum
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT relname, last_vacuum, last_autovacuum FROM pg_stat_all_tables WHERE schemaname = 'public' ORDER BY last_autovacuum IS NULL DESC LIMIT 10;"
```

---

### 2. Check Resource Utilization (3 min)

```bash
# Pod resource usage
kubectl top pod postgres-0 -n flyfleet

# Check if hitting limits
kubectl describe pod postgres-0 -n flyfleet | grep -A5 "Limits"

# Disk space
kubectl exec -it postgres-0 -n flyfleet -- df -h /var/lib/postgresql/data

# Database size
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT pg_size_pretty(pg_database_size('fleet_db'));"

# Table sizes (top 10)
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size FROM pg_tables ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC LIMIT 10;"
```

**Red flags:**
- CPU > 90% sustained
- Memory hitting limit (OOM risk)
- Disk > 85% full
- Database size growing too fast (consider partitioning)

---

### 3. Check Slow Queries (3 min)

```bash
# Top 10 slowest queries (by average time)
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT query, mean_exec_time, calls
    FROM pg_stat_statements
    ORDER BY mean_exec_time DESC
    LIMIT 10;
  "

# Top 10 most frequent queries (to see if indexing needed)
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT query, calls, mean_exec_time
    FROM pg_stat_statements
    ORDER BY calls DESC
    LIMIT 10;
  "

# Queries currently running (> 5s)
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT pid, now() - pg_stat_activity.query_start AS duration, query, state
    FROM pg_stat_activity
    WHERE state = 'active' AND (now() - pg_stat_activity.query_start) > interval '5 seconds'
    ORDER BY duration DESC;
  "

# Check for sequential scans (missing indexes)
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT schemaname, tablename, seq_scan, idx_scan, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
    FROM pg_stat_user_tables
    WHERE seq_scan > 0 AND idx_scan = 0
    ORDER BY seq_scan DESC
    LIMIT 10;
  "
```

**Actionable findings:**
- High `mean_exec_time` on frequent queries → Index needed
- `seq_scan > 0` but `idx_scan = 0` → Missing index
- Long-running active queries → Consider killing if blocking others

---

### 4. Check Locks & Blocking (3 min)

```bash
# Current locks
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT
      locktype, database, relation::regclass, mode, granted, pid,
      age(clock_timestamp(), query_start) as age, query
    FROM pg_locks pl
    JOIN pg_stat_activity psa ON pl.pid = psa.pid
    WHERE NOT granted OR age(clock_timestamp(), query_start) > interval '1 second'
    ORDER BY age DESC;
  "

# Blocking queries
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT blocked_locks.pid AS blocked_pid,
           blocked_activity.usename AS blocked_user,
           blocking_locks.pid AS blocking_pid,
           blocking_activity.usename AS blocking_user,
           blocked_activity.query AS blocked_statement,
           blocking_activity.query AS current_statement_in_blocking_process
    FROM pg_catalog.pg_locks blocked_locks
    JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_locks.pid = blocked_activity.pid
    JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
      AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
      AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
      AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
      AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
      AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
      AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
      AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
      AND blocking_locks.pid != blocked_locks.pid
    JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_locks.pid = blocking_activity.pid
    WHERE NOT blocked_locks.granted;
  "
```

**If blocking lock found:**
- Identify blocking query
- Evaluate if safe to terminate (using `pg_terminate_backend(pid)`)
- Consider implementing `statement_timeout` to prevent long locks

---

### 5. Check Connection Pool (2 min)

```bash
# Check pgbouncer if used (FlyFleet may use pgxpool directly)
# Check application connection pool settings
kubectl exec -it deployment/fleet-core -n flyfleet -- \
  ps aux | grep fleet-core | grep -oP 'DATABASE_MAX_OPEN_CONNS=\K\d+'

# Or check env in deployment
kubectl get deployment fleet-core -n flyfleet -o jsonpath='{.spec.template.spec.containers[0].env}'

# Count connections per user
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT usename, count(*) FROM pg_stat_activity GROUP BY usename ORDER BY count(*) DESC;"

# Check idle connections consuming resources
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT count(*) FROM pg_stat_activity WHERE state = 'idle' AND state_change < NOW() - INTERVAL '5 minutes';"
```

**Connection pool exhaustion fixes:**
- Increase `max_connections` in PostgreSQL config (requires restart)
- Increase application pool size settings
- Check for connection leaks (connections not closed in code)
- Implement connection pool health checks

---

### 6. Check Replication (if applicable) (2 min)

```bash
# Replication status
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT application_name, state, sync_state, pg_size_pretty(pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn)) as replication_lag FROM pg_stat_replication;"

# Check if read replica is behind
kubectl exec -it postgres-read-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT pg_last_wal_replay_lsn(), pg_current_wal_lsn();"

# Replication lag in seconds
# Compare primary's current WAL LSN with replica's replay LSN
```

**If lag > 5 seconds:**
- Check network between primary and replica
- Check replica CPU/IOWait
- Consider adding replicas or optimizing queries

---

## Mitigation Strategies

### Strategy 1: Restart PostgreSQL (Last Resort)

**When:** Database completely unresponsive, no other options

⚠️ **HIGH RISK:** Can cause data loss if transactions in progress. Use only as last resort.

```bash
# Graceful restart (preferred)
kubectl rollout restart statefulset/postgres -n flyfleet

# Monitor restart
kubectl get pods -n flyfleet -l app=postgres -w
kubectl logs -f postgres-0 -n flyfleet

# Wait for primary to be ready (if using replication)
kubectl exec -it postgres-0 -n flyfleet -- pg_isready

# Check replication promotion (if failover)
kubectl get pods -n flyfleet -l app=postgres
```

**After restart:**
- Check for replication re-sync (large lag possible)
- Verify application connectivity
- Run health checks
- Investigate root cause before production stable

---

### Strategy 2: Kill Long-Running Queries

**When:** Specific queries blocking others, high CPU

```bash
# Identify blocking query (from Investigation Step 4)
# Get PID of blocking query

# Terminate backend
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT pg_terminate_backend(<PID>);"

# OR terminate all queries running > 5 minutes
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT pg_terminate_backend(pid)
    FROM pg_stat_activity
    WHERE state = 'active'
      AND (now() - pg_stat_activity.query_start) > interval '5 minutes'
      AND pid <> pg_backend_pid();
  "
```

**Caution:** Application may throw errors. Communicate before killing critical queries.

---

### Strategy 3: Add Missing Index

**When:** Query slow due to seq scan, high load on table

Based on `pg_stat_statements` query, create index:

```bash
# Example: If query filters on device_id and org_id
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    CREATE INDEX CONCURRENTLY idx_telemetry_device_org
    ON telemetry_events (device_id, org_id)
    WHERE device_id IS NOT NULL;
  "

# Monitor query improvement
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT query, mean_exec_time, calls
    FROM pg_stat_statements
    WHERE query LIKE '%telemetry_events%'
    ORDER BY mean_exec_time DESC
    LIMIT 1;
  "
```

**Best practice:** Use `CREATE INDEX CONCURRENTLY` to avoid table lock. Run during off-peak if possible.

---

### Strategy 4: Increase max_connections

**When:** "too many connections" errors, connection pool exhausted

```bash
# Edit PostgreSQL config (via ConfigMap or Helm values)
kubectl edit configmap postgres-config -n flyfleet

# In postgresql.conf:
# max_connections = 200  # default 100, increase based on memory (2MB per connection)

# Also increase application pool size to match
kubectl edit deployment fleet-core -n flyfleet
# Add env: DATABASE_MAX_OPEN_CONNS=150

# Restart PostgreSQL (required for config change)
kubectl rollout restart statefulset/postgres -n flyfleet
kubectl rollout restart deployment/fleet-core -n flyfleet

# Verify new connection limit
kubectl exec -it postgres-0 -n flyfleet -- psql -c "SHOW max_connections;"
```

**Memory calculation:** `max_connections * work_mem * 3` approx. Ensure pod memory sufficient.

---

### Strategy 5: Table Maintenance (VACUUM/ANALYZE)

**When:** Table bloat, slow queries on large tables, high dead tuples

```bash
# Check dead tuples
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT schemaname, tablename, n_dead_tup, n_live_tup
    FROM pg_stat_user_tables
    WHERE n_dead_tup > 1000
    ORDER BY n_dead_tup DESC
    LIMIT 10;
  "

# VACUUM ANALYZE specific table (concurrent, online)
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "VACUUM ANALYZE telemetry_events;"

# For heavily bloated tables, use pg_repack (install extension first)
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT pg_repack.repack_table('telemetry_events');"
```

**Schedule:** Regular autovacuum should handle this. If not, tune autovacuum thresholds.

---

### Strategy 6: Expand Storage (PVC)

**When:** Disk > 85% full, need more space

```bash
# Edit PVC to increase size (if storage class supports expansion)
kubectl edit pvc data-postgres-0 -n flyfleet

# Modify:
# spec:
#   resources:
#     requests:
#       storage: 100Gi  # Increase from current size

# Check if PVC allows expansion
kubectl get pvc data-postgres-0 -n flyfleet -o jsonpath='{.spec.storageClassName}'
# Ensure storage class allows expansion (allowVolumeExpansion: true)

# Verify expansion
kubectl get pvc data-postgres-0 -n flyfleet

# Inside pod, verify new size
kubectl exec -it postgres-0 -n flyfleet -- df -h /var/lib/postgresql/data
```

---

## Recovery & Validation

### Validation Steps

1. **Database Health**
```bash
kubectl exec -it postgres-0 -n flyfleet -- pg_isready
# Should return "accepting connections"
```

2. **Connection Success**
```bash
kubectl exec -it deployment/fleet-core -n flyfleet -- \
  psql $DATABASE_URL -c "SELECT 1" 2>&1
# Should return "?column?" = 1
```

3. **Query Performance**
```bash
# P95 query time < 100ms
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT histogram_quantile(0.95, jsonb_array_elements_text((pg_stat_statements_reset()).json)::numeric) as p95_ms
    FROM pg_stat_statements LIMIT 1;
  "
# (Simpler) Check top slow queries < 500ms
```

4. **Connection Pool Normal**
```bash
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT count(*) FROM pg_stat_activity;"
# Should be well below max_connections
```

5. **API Health**
```bash
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health
# Should return 200

# Check error rate via Prometheus
curl -s 'http://prometheus:9090/api/v1/query?query=rate(http_requests_total{status=~"5..",service="fleet-core"}[5m])' | jq
# Should be near zero
```

---

## Post-Incident Analysis

### PIR Focus Areas

1. **Root Cause Categories:**
   - Unoptimized query (missing index, bad query plan)
   - Connection leak in application code (connections not closed)
   - Insufficient resources (CPU, memory, disk)
   - Table bloat (high dead tuples from deletes/updates)
   - Lock contention (long-running transactions)
   - Autovacuum misconfiguration (not cleaning up)
   - Replication lag (if using replicas, primary overloaded)

2. **Preventive Actions:**
   - Add missing indexes (document schema changes)
   - Implement query plan monitoring (auto-alert on seq scan increase)
   - Enforce connection pool best practices (context cancellation, proper closing)
   - Increase `max_connections` or implement connection pooling (pgbouncer)
   - Tune autovacuum (lower thresholds, more aggressive)
   - Add read replicas to distribute read load
   - Implement database query caching layer (Redis)
   - Set `statement_timeout` to prevent runaway queries

3. **Schema Improvements:**
   - Partition large tables by date (telemetry_events by month)
   - Implement TTL for old data (archive to S3)
   - Use TimescaleDB for time-series data (telemetry)
   - Consider sharding by org_id for multi-tenant scale

---

## Prevention & Monitoring

### Essential Alerts

```yaml
- alert: PostgreSQLDown
  expr: up{service="postgres"} == 0
  for: 1m

- alert: PostgreSQLHighConnections
  expr: pg_stat_activity_count > 0.8 * postgres_max_connections
  for: 5m

- alert: PostgreSQLQueryLatencyHigh
  expr: histogram_quantile(0.95, sum(rate(pg_stat_statements_mean_exec_time_seconds_bucket[5m])) by (le)) > 0.1
  for: 10m

- alert: PostgreSQLDiskSpaceLow
  expr: (node_filesystem_size_bytes{mountpoint="/var/lib/postgresql/data"} - node_filesystem_free_bytes{mountpoint="/var/lib/postgresql/data"}) / node_filesystem_size_bytes{mountpoint="/var/lib/postgresql/data"} > 0.85
  for: 30m

- alert: PostgreSQLReplicationLagHigh
  expr: pg_replication_lag_seconds > 5
  for: 2m
```

### Dashboard Panels

1. Connection count vs max (gauge)
2. Top 5 slowest queries (table)
3. Query throughput (queries/sec)
4. Cache hit ratio (should be > 99%)
5. Table sizes (top 10 by rows)
6. Dead tuples (bloat indicator)
7. Replication lag (if applicable)

---

## Related Runbooks

- [API Outage](api-outage.md) - If database is root cause of API failures
- [Data Inconsistency](data-inconsistency.md) - If corruption suspected
- [Telemetry Disruption](telemetry-disruption.md) - Telemetry DB performance

---

**Last Updated:** 2026-03-29
**Review Date:** 2026-06-29
**Owner:** Database Team
