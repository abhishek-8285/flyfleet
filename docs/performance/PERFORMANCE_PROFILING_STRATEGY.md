# FlyFleet Performance Profiling Strategy

**Date:** 2026-03-28
**Engineer:** Claude Code Performance Profiler
**Project:** FlyFleet Fleet Management System
**Tech Stack:** Go 1.26, Chi, PostgreSQL (TimescaleDB), NATS JetStream, Redis

---

## Executive Summary

FlyFleet has **excellent observability foundations** (OpenTelemetry, Prometheus, Grafana, pg_stat_statements) but **lacks CPU/memory profiling endpoints** (pprof). The static analysis revealed large binary artifacts (60MB) and substantial log files, but these are primarily build artifacts and can be managed.

### Key Findings

| Component | Status | Risk Level |
|-----------|--------|------------|
| **Database Pooling** | ✅ Excellent | Low |
| **Observability Stack** | ✅ Comprehensive | Low |
| **pprof Endpoints** | ❌ Missing | Medium |
| **Load Testing** | ✅ Present (k6) | Low |
| **pg_stat_statements** | ✅ Enabled | Low |
| **Large Files** | ⚠️ Binaries/Logs | Low (artifacts) |
| **God Classes** | ⚠️ Partially Refactored | Medium |

### Immediate Priorities

1. **Add pprof endpoints** for CPU/memory profiling (30 min)
2. **Establish baseline metrics** with existing Prometheus SLOs (1 day)
3. **Analyze slow queries** using pg_stat_statements (1 day)
4. **Run load tests** to validate performance targets (1 day)
5. **Goroutine leak detection** (1 hour)

---

## 1. Current State Analysis

### 1.1 Existing Performance Infrastructure ✅

FlyFleet already has strong foundations:

**Metrics & Monitoring**
- Prometheus client_golang integrated (`github.com/prometheus/client_golang v1.4.0`)
- Custom metrics in `internal/platform/metrics/`
- SLO recording rules: `infrastructure/monitoring/slo-recording-rules.yml`
- Grafana dashboards: `grafana-dashboard-flyfleet.json`
- HTTP metrics middleware: `metrics.HTTPMetricsMiddleware(logger)`

**Database Performance**
- `pg_stat_statements` preloaded in TimescaleDB (`shared_preload_libraries=pg_stat_statements,timescaledb`)
- Sophisticated tiered connection pooling with RLS enforcement (`internal/database/pools.go`):
  - Critical: 500 connections (auth, emergency, tracking)
  - High: 300 connections (billing, vehicle, driver)
  - Medium: 150 connections (analytics, reports)
  - Low: 50 connections (diagnostics, integrations)
- Automatic connection health checks (`HealthCheckPeriod: 1 minute`)
- RLS overhead monitoring (alerts if >5ms) ✅

**Tracing**
- OpenTelemetry SDK v1.40.0 (`go.opentelemetry.io/otel v1.40.0`)
- Jaeger/Tempo exporter configured
- Trace context propagation across services

**Load Testing**
- k6 scripts in `tests/k6/`:
  - `emergency_load_test.js` (SOS/incident response)
  - `finance_load_test.js` (billing operations)
  - `platform_load_test.js` (general platform load)

**Alerting**
- SLO-based alerts with burn rate detection
- Alertmanager configuration: `alertmanager-config-enhanced.yaml`

### 1.2 Gaps & Missing Pieces

| Gap | Impact | Effort |
|-----|--------|--------|
| **No pprof endpoints** | Cannot profile CPU/heap/goroutines at runtime | 30 min |
| **Large binaries in tmp/** | Wasted disk space, backup bloat | 1 hour (cleanup) |
| **Unbounded log growth** | Disk pressure, I/O contention | 2 hours (logrotate) |
| **No query analysis** | Unknown slow queries in production | 1 day |
| **No goroutine monitoring** | Potential leaks undetected | 2 hours |

### 1.3 Static Analysis Results (Revisited)

From the `performance_profiler.py` scan:

**Large Files** (artifacts, not source code issues):
```
services/fleet-core/main: 60.5MB         ← Release binary (debug symbols)
tmp/fleet-backend: 58.4MB               ← Build artifact (should be .gitignore'd)
services/_archive/*: 35-60MB each       ← Archived binaries (legacy)
tmp/logs/*.log: 4-45MB                  ← Log rotation needed
pgdata/pg_wal/*: 16MB each              ← Normal WAL segments
```

**Assessment**: These are **build/diagnostic artifacts**, not production code. Actions:
- Add `tmp/` and `services/_archive/` to `.gitignore` (likely already done)
- Implement log rotation for `tmp/logs/`
- Clean up old binaries periodically

**Bundle Analysis**: 416 JS-like files, 2.9MB weighted. Low risk for Go backend.

---

## 2. Performance Profiling Plan

### Phase 1: Enable Runtime Profiling (pprof) ⚡ QUICK WIN

**Objective**: Add pprof endpoints to enable on-demand CPU, memory, and goroutine profiling.

**Implementation** (add to `cmd/main.go`):

```go
import (
    _ "net/http/pprof"
    // ... existing imports
)

func main() {
    // ... existing initialization

    // Initialize router (r := chi.NewRouter())

    // Add pprof endpoints on separate port (security: not exposed publicly)
    if os.Getenv("ENABLE_PPROF") == "true" {
        pprofServer := &http.Server{
            Addr:    ":6060",
            Handler: http.DefaultServeMux, // pprof registers on DefaultServeMux
        }
        go func() {
            logger.Info("pprof server starting on :6060")
            if err := pprofServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
                logger.Fatal("pprof server failed", zap.Error(err))
            }
        }()
        defer pprofServer.Shutdown(ctx)
    }

    // Start main server
    // ... existing server startup
}
```

**Security**:
- Bind pprof to localhost only or internal network (use `ENABLE_PPROF=true PPROF_ADDR=127.0.0.1:6060`)
- Never expose pprof in production without auth/k8s network policies
- Use as a diagnostic tool in staging/development

**Usage** (once deployed):
```bash
# CPU profile (30s)
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/profile?seconds=30

# Memory heap profile
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/heap

# Goroutine leak detection
curl http://localhost:6060/debug/pprof/goroutine?debug=1 > goroutines.txt

# Flame graph (goroutines)
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/goroutine

# Block profile (for mutex contention)
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/block

# Alloc objects profile (allocation rate)
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/allocs
```

**Integration**: Add `ENABLE_PPROF` to `.env.example` and docker-compose.yml for staging.

---

### Phase 2: Baseline Establishment 📊 BEFORE OPTIMIZATION

**Objective**: Record current performance characteristics before any changes.

**2.1 Prometheus Baseline Export**

Extract existing SLO metrics for 7 days:

```bash
# Prometheus recording rules already compute SLIs
# Export data for:
# - sli:api:availability:percentage (target: 99.95%)
# - sli:api:latency:p95 (target: <300ms)
# - sli:api:error_rate:percentage (target: <0.1%)
# - sli:telemetry:ingestion_rate (target: >10,000/sec)
# - sli:telemetry:latency:p99 (target: <5s)
# - pg_stat_statements_mean_exec_time (target: <100ms)
# - connection_pool_utilization (target: <80%)
```

Create a `baseline-performance-YYYY-MM-DD.md` report in `flyfleet/PERFORMANCE_REPORTS/`.

**2.2 Database Query Baseline**

Run slow query analysis:

```sql
-- Enable pg_stat_statements (already loaded, ensure extension exists)
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Get top 20 slowest queries by mean execution time (min 100 calls)
SELECT
  round(mean_exec_time::numeric, 2) AS mean_ms,
  calls,
  round(total_exec_time::numeric, 2) AS total_ms,
  round(stddev_exec_time::numeric, 2) AS stddev_ms,
  left(query, 100) AS query
FROM pg_stat_statements
WHERE calls > 100
ORDER BY mean_exec_time DESC
LIMIT 20;

-- Reset stats before load test (to capture fresh data)
SELECT pg_stat_statements_reset();
```

Export results to `flyfleet/queries/slow-queries-baseline-$(date +%Y%m%d).sql`.

**2.3 Load Test Baseline**

Run existing k6 tests against staging (or local with realistic data volume):

```bash
# Emergency service load test
k6 run tests/k6/emergency_load_test.js \
  --env BASE_URL=http://localhost:8081 \
  --out json=results/emergency-$(date +%Y%m%d-%H%M%S).json

# Finance load test
k6 run tests/k6/finance_load_test.js \
  --env BASE_URL=http://localhost:8081 \
  --out json=results/finance-$(date +%Y%m%d-%H%M%S).json

# Platform load test
k6 run tests/load/platform_load_test.js \
  --env BASE_URL=http://localhost:8081 \
  --out json=results/platform-$(date +%Y%m%d-%H%M%S).json
```

**Record**:
- P50, P95, P99 latencies per endpoint
- Request per second (RPS)
- Error rate (5xx, 4xx)
- Check if thresholds are met:
  - From OBSERVABILITY_DESIGN.md: P95 API latency <300ms, error rate <0.1%
  - From architecture: P99 telemetry <5s

**2.4 Resource Utilization Baseline**

With services running under load (use `-vus 50 -duration 5m`), record:

```bash
# Container metrics
docker stats fleet-postgres fleet-nats fleet-redis fleet-core-service --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}"

# Process metrics (on host)
pid=$(pgrep -f "fleet-core-service")
ps -p $pid -o %cpu,%mem,rss,vsz,etime

# Go runtime metrics (via pprof once added)
curl http://localhost:6060/debug/pprof/heap > heap-baseline.prof
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu-baseline.prof
curl http://localhost:6060/debug/pprof/goroutine?debug=1 > goroutines-baseline.txt
```

**2.5 Connection Pool Baseline**

Check pool sizes during load:

```sql
-- Active connections per pool
SELECT
  application_name,
  state,
  count(*) as connections
FROM pg_stat_activity
WHERE application_name LIKE 'flyfleet-%'
GROUP BY application_name, state
ORDER BY application_name, state;

-- Pool utilization (from app metrics)
# Should show: active/max connections <80% for all pools
```

---

### Phase 3: Database Query Optimization

**3.1 Identify Slow Queries**

Use results from Phase 2.2. Focus on queries with:
- `mean_exec_time > 100ms`
- High `stddev_exec_time` (inconsistent performance)
- `calls > 1000` (high frequency, even if moderately fast)

**3.2 Analyze Query Plans**

For each slow query:

```sql
-- Capture the exact query from pg_stat_statements
-- Then run EXPLAIN ANALYZE to see the execution plan
EXPLAIN (ANALYZE, BUFFERS, VERBOSE, FORMAT TEXT) <paste query here>;

-- Look for:
-- ❌ Seq Scan on large table → missing index
-- ❌ Nested Loop with >1000 rows → N+1 query pattern
-- ❌ Sort on disk → temp file, add index or materialized view
-- ❌ Hash Join with huge hash table → memory pressure
-- ✅ Index Scan, Hash Join (small), Merge Join (sorted data)
```

**3.3 N+1 Query Detection**

Search repository code for loops with DB queries:

```bash
# Find patterns like: for { x := range items { db.Query(..., x.ID) }}
grep -r "for.*range" services/fleet-core/internal/domains/ -A3 --include="*.go" | grep -B2 "Query\|Exec"

# Better: Write a script to detect N+1 (or use existing test coverage)
# Look for queries inside loops that could be batched or JOINed
```

**Before/After Template**:

```markdown
## Query Optimization: [Endpoint/Feature]

**Date:** 2026-03-28
**Query:** (paste slow query)

### Baseline
- Mean execution: 450ms
- Calls: 2,400/min
- Total DB time: 18s/min

### Root Cause
- N+1 pattern: fetching N items then querying related data in loop
- Missing index on `organization_id`

### Fix Applied
- Replaced N+1 with single JOIN query
- Added composite index: `CREATE INDEX CONCURRENTLY idx_table_org_id ON table(org_id)`

### After
- Mean execution: 12ms
- Calls: 1 (instead of N+1)
- Total DB time: 0.2s/min

### Improvement
- **37x faster**, 98% reduction in DB time
```

**3.4 Index Optimization**

Create missing indexes identified from `EXPLAIN`:

```sql
-- Common patterns in FlyFleet (based on domains):
-- 1. org_id filters (multi-tenancy)
CREATE INDEX CONCURRENTLY idx_table_org_id ON table_name(org_id);

-- 2. Composite indexes for common WHERE + ORDER BY
CREATE INDEX CONCURRENTLY idx_table_org_created ON table_name(org_id, created_at DESC);

-- 3. Indexes for foreign keys (if not already)
CREATE INDEX CONCURRENTLY idx_related_table_id ON related_table(id);

-- 4. Partial indexes for active records only
CREATE INDEX CONCURRENTLY idx_active_org ON table_name(org_id) WHERE deleted_at IS NULL;

-- 5. TimescaleDB: automatic compression, retention policies
SELECT add_retention_policy('telemetry', INTERVAL '90 days');
SELECT add_compression_policy('telemetry', INTERVAL '7 days');
```

**3.5 Connection Pool Tuning**

Current pool config (`pools.go`):
```go
config.MaxConns = 500/300/150/50  // Already tiered by domain priority
config.MinConns = maxConns / 10
config.MaxConnLifetime = time.Hour
config.MaxConnIdleTime = 30 * time.Minute
config.HealthCheckPeriod = 1 * time.Minute
```

**Monitor**:
- If `active_connections > max_connections * 0.8` → increase pool size
- If `idle_connections > min_connections` for long periods → reduce pool size (save memory)
- Adjust tier allocations based on real usage metrics

---

### Phase 4: Load Testing & Performance Validation

**4.1 Create Additional k6 Scenarios**

Existing tests: emergency, finance, platform. Add:

- **Tracking/Telemetry Ingestion** (high-throughput: 10K+ events/sec)
- **Bulk Operations** (driver/vehicle imports)
- **Concurrent User Simulation** (multiple orgs)
- **Mixed Workload** (read-heavy: 80% GET, 20% POST/PUT)

**Template** (`tests/k6/telemetry_ingestion.js`):

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

const errorRate = new Rate('errors');
const ingestionLatency = new Trend('telemetry_ingestion_duration');
const eventsPerSecond = new Counter('telemetry_events_per_second');

export const options = {
  stages: [
    { duration: '1m', target:.target: 500 },  // Ramp to 500 VUs
    { duration: '5m', target: 500 },  // Sustain
    { duration: '1m', target: 1000 }, // Spike
    { duration: '2m', target: 0 },    // Ramp down
  ],
  thresholds: {
    telemetry_ingestion_duration: ['p(95)<1000'],  // 95% < 1s end-to-end
    errors: ['rate<0.01'],  // <1% errors
  },
};

export default function() {
  // Simulate device sending telemetry
  const payload = JSON.stringify({
    device_id: `device-${__VU}`,
    latitude: 37.7749 + Math.random() * 0.1,
    longitude: -122.4194 + Math.random() * 0.1,
    speed_kmh: Math.floor(Math.random() * 120),
    timestamp: new Date().toISOString(),
    org_id: `org-${__VU % 10}`,  // 10 orgs
  });

  const start = Date.now();
  const res = http.post('http://localhost:8081/api/v1/telemetry', payload, {
    headers: { 'Content-Type': 'application/json' },
  });
  ingestionLatency.add(Date.now() - start);
  eventsPerSecond.add(1);

  check(res, { 'status 200': (r) => r.status === 200 }) || errorRate.add(1);
  sleep(0.1);  // ~10 events/sec per VU
}
```

**4.2 Run Load Tests with Profiling**

```bash
# Start service with pprof enabled
ENABLE_PPROF=true docker-compose up fleet-core-service

# Run load test while capturing profiles
k6 run --out influxdb=http://localhost:8086 tests/k6/telemetry_ingestion.js

# Meanwhile, capture profiles at peak load:
# - CPU profile (60s capture)
curl "http://localhost:6060/debug/pprof/profile?seconds=60" -o cpu-peak.prof
# - Memory heap
curl "http://localhost:6060/debug/pprof/heap" -o heap-peak.prof
# - Goroutines
curl "http://localhost:6060/debug/pprof/goroutine?debug=1" -o goroutines-peak.txt
```

**4.3 Compare Before/After**

Use the **Before/After Measurement Template** from SKILL.md.

Create PR description template:
```markdown
## Performance Optimization: [What Was Fixed]

### Problem
[1-2 sentences]

### Baseline (Before)
| Metric | Value |
|--------|-------|
| P95 latency | 450ms |
| DB queries/req | 23 |
| P99 telemetry | 8.2s |

### Fix Applied
- Refactored N+1 query in analytics repository
- Added composite index on `telemetry.device_id, created_at`
- Connected pool size tuned: Critical 500→600

### After
| Metric | Before | After | Delta |
|--------|--------|-------|-------|
| P95 latency | 450ms | 89ms | **-80%** |
| DB queries/req | 23 | 3 | **-87%** |
| P99 telemetry | 8.2s | 1.2s | **-85%** |

### Verification
- Load test: `k6 run tests/k6/analytics.json` (P95:89ms, error:0.0%)
- Query plan: [link to EXPLAIN ANALYZE output]
- Profile diff: `go tool pprof -http=:8081 cpu-baseline.prof cpu-after.prof`
```

---

### Phase 5: Memory & Goroutine Leak Detection

**5.1 Heap Profile Comparison**

```bash
# Capture baseline (idle)
curl http://localhost:6060/debug/pprof/heap > heap-idle.prof

# Run load test for 5 minutes
k6 run --duration 5m tests/k6/platform_load_test.js

# Capture after load
curl http://localhost:6060/debug/pprof/heap > heap-after-load.prof

# Compare
go tool pprof -http=:8081 heap-idle.prof heap-after-load.prof

# Look for:
# - Objects that didn't get GC'd (inuse_objects)
# - Serving vs total GC cycles ratio (if serving << total, many short-lived objs)
# - Memory growth trend: baseline vs after
```

**5.2 Goroutine Leak Detection**

```bash
# Baseline goroutines
curl http://localhost:6060/debug/pprof/goroutine?debug=1 > goroutines-idle.txt
grep -c "^goroutine " goroutines-idle.txt  # e.g., 150

# After load test
curl http://localhost:6060/debug/pprof/goroutine?debug=1 > goroutines-after.txt
grep -c "^goroutine " goroutines-after.txt  # e.g., 180

# If significantly higher (e.g., 500+), investigate:
# - Unclosed channels
# - select without default (blocked forever)
# - Workers not shutting down
# Use `go tool pprof -http=:8081 http://localhost:6060/debug/pprof/goroutine` to see stack traces
```

**5.3 Block Profiling** (for mutex/channel contention)

```bash
# Enable block profile (already compiled in, just fetch)
curl "http://localhost:6060/debug/pprof/block?debug=1" > block.txt
# Look for many goroutines blocked on sync.Mutex or channel ops
go tool pprof http://localhost:6060/debug/pprof/block
```

---

### Phase 6: Continuous Performance Monitoring

**6.1 CI/CD Integration**

Add performance gates to GitHub Actions:

```yaml
# .github/workflows/performance-gate.yml
name: Performance Gate
on:
  push:
    branches: [main, release-*]
  pull_request:

jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.26'
      - name: Start services
        run: docker-compose up -d postgres redis nats
      - name: Build and run service
        run: |
          go build -o fleet-core ./services/fleet-core/cmd/main.go
          ENABLE_PPROF=true ./fleet-core &
      - name: Wait for health
        run: |
          for i in {1..30}; do
            if curl -s http://localhost:8081/healthz | grep -q pass; then break; fi
            sleep 2
          done
      - name: Run k6 load test
        run: |
          docker run --network=host loadimpact/k6 run \
            --out json=results.json \
            tests/k6/platform_load_test.js
      - name: Check thresholds
        run: |
          # Parse results.json, check P95 < 300ms, error rate < 1%
          # Fail if violated
```

**6.2 Automated Slow Query Alerting**

Prometheus rule (already exists, verify):

```yaml
# infrastructure/monitoring/slo-burn-rate-alerts.yml
- alert: SlowQueryDetected
  expr: pg_stat_statements_mean_exec_time > 0.5  # 500ms
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Slow query detected in {{ $labels.query }}"
    description: "Mean exec time: {{ $value }}s"
```

**6.3 Performance Regression Dashboard**

Enhance Grafana dashboard with:
- P50/P95/P99 trends over time (7d, 30d)
- Slow query count (pg_stat_statements > 100ms)
- Connection pool utilization per tier
- Goroutine count
- Heap memory growth rate
- GC pause durations

---

## 3. Before/After Measurement Template

Use this for every performance optimization:

```markdown
## Performance Optimization: [Brief Description]

**Date:** YYYY-MM-DD
**Engineer:** @username
**Ticket:** PROJ-123

### Problem
Describe what was slow and how you observed it:
- API endpoint `/api/v1/analytics/reports` had P95 latency of 2.1s
- Database CPU at 95% during peak hours
- Memory grew 500MB/hour (suspected leak)

### Root Cause
What the profiler/analyzer revealed:
- N+1 query: fetching 200 tasks then querying assignees in loop
- Missing index on `tasks(org_id, created_at)`
- Goroutine leak: workers not closing channel consumers

### Baseline (Before)
| Metric | Value |
|--------|-------|
| P95 latency | 2100ms |
| DB queries/request | 201 |
| Mean exec time (slow query) | 350ms |
| Goroutines (idle) | 150 |
| Goroutines (under load) | 5,000 (leaked 4,850) |
| Heap used | 450MB |

Profiler evidence:
- Flamegraph: `cpu-baseline.prof` (query hot spot)
- Slow query: `SELECT * FROM users WHERE id = $1` called 200x in loop
- Goroutine stack: blocked on channel receive

### Fix Applied
Specific changes made:
- Replaced N+1 with JOIN: `SELECT t.*, u.name FROM tasks t LEFT JOIN users u ON u.id = t.assignee_id`
- Added index: `CREATE INDEX CONCURRENTLY idx_tasks_org_created ON tasks(org_id, created_at DESC);`
- Fixed leak: call `consumer.Stop()` in `defer` in worker loop

Code diff:
```go
// Before: N+1 loop
for _, task := range tasks {
    assignee, _ := db.GetUser(task.AssigneeID)  // 200 queries
    task.Assignee = assignee
}

// After: Single JOIN query
rows, _ := db.Query(`
    SELECT t.*, u.name FROM tasks t
    LEFT JOIN users u ON u.id = t.assignee_id
    WHERE t.org_id = $1 ORDER BY t.created_at DESC LIMIT 200
`, orgID)
```

### After
| Metric | Before | After | Delta |
|--------|--------|-------|-------|
| P95 latency | 2100ms | 120ms | **-94%** |
| DB queries/request | 201 | 1 | **-99.5%** |
| Mean exec time | 350ms | 8ms | **-97.7%** |
| Goroutines (under load) | 5,000 | 180 | **-96.4%** |
| Heap used | 450MB | 120MB | **-73%** |
| Error rate | 2.3% | 0% | **-100%** |

### Verification
- Load test: `k6 run tests/k6/analytics_reports.js` (P95:120ms, error:0%)
- Query plan: `EXPLAIN ANALYZE` shows Index Scan (0.8ms) vs Seq Scan (350ms)
- Profile: `go tool pprof -http=:8081 cpu-baseline.prof cpu-after.prof`
- Memory leak fixed: heap stable after 30m load (verified GC cycles)

### Rollout
- Deployed to staging 2026-03-28 14:30 UTC
- Monitoring: Grafana dashboard shows P95 dropped from 2.1s → 120ms
- No alerts triggered during canary (5% → 25% → 100%)
- PR: https://github.com/flyfleet/fleet-core-service/pull/1234
```

---

## 4. Optimization Checklist for FlyFleet

### Database Optimizations

- [ ] **Enable pg_stat_statements** (already done ✅)
- [ ] **Analyze top 20 slow queries** (weekly)
- [ ] **Add missing indexes** on:
  - `[org_id]` columns (all tables)
  - `[org_id, created_at]` (common listing queries)
  - `[device_id, created_at]` (telemetry queries)
  - Foreign keys not indexed
- [ ] **Fix N+1 queries** (search for queries in loops)
- [ ] **Add LIMIT clauses** to all unbounded list queries
- [ ] **Use SELECT specific columns** instead of `SELECT *`
- [ ] **TimescaleDB compression** on telemetry >7 days old
- [ ] **Connection pool tuning** based on utilization metrics

### Go Runtime Optimizations

- [ ] **Add pprof endpoints** (`import _ "net/http/pprof"`)
- [ ] **Enable block profiling** if mutex contention suspected
- [ ] **Monitor goroutine count** (should not grow unbounded)
- [ ] **Check for memory leaks** (heap growth >10% after GC)
- [ ] **Reduce large struct allocations** in hot paths
- [ ] **Use sync.Pool** for frequently allocated objects (e.g., buffers)
- [ ] **Profile GC pause times** (`go tool pprof http://localhost:6060/debug/pprof/gc`)

### API & HTTP Optimizations

- [ ] **Add response caching** (Cache-Control headers for non-PII data)
- [ ] **Compress responses** (gzip/brotli) - check if middleware already exists
- [ ] **Parallelize independent DB calls** (use `errgroup.Group` or `sync.WaitGroup`)
- [ ] **Implement pagination** (limit/offset or cursor-based)
- [ ] **Rate limiting per user/org** (already in place? check `redis_rate_limiter.go`)
- [ ] **Request size limits** (already: `chiMiddleware.Timeout(60s)` ✅)

---

## 5. Common Pitfalls (FlyFleet-Specific)

1. **Ignoring P99** - P50 may be fine while tail latency is catastrophic. Always check P95/P99.

2. **Testing on dev data** - 100 rows vs 100M rows is different. Load test with production-scale data volume (use `pg_dump --data-only --rows` or `pgbench`).

3. **Connection pool misconfiguration** - Too few connections → queuing; too many → DB overload. Monitor `pgpool_active_connections`.

4. **Not measuring after fix** - Always run the same load test before and after. Use identical seed data.

5. **Load testing production** - Use staging with production-like data size. Never load test prod.

6. **Missing context timeouts** - All DB/NATS/Redis calls must have timeouts (currently: `chiMiddleware.Timeout(60s)` covers HTTP, but ensure DB calls use `context.WithTimeout`).

7. **RLS overhead** - Currently monitored (>5ms alert). If RLS setup becomes bottleneck, consider connection pool size increase.

8. **God classes prevent optimization** - Large files like `analytics_service.go` (2135→146 lines ✅) are hard to profile. Keep files <1000 lines.

---

## 6. Recommended Actions (FlyFleet)

### Immediate (This Week)

- [ ] **Add pprof endpoints** on `:6060` (internal only, gated by `ENABLE_PPROF`)
- [ ] **Capture baseline metrics** (7 days of Prometheus data export)
- [ ] **Run baseline load tests** (emergency, finance, platform)
- [ ] **Extract top 10 slow queries** from `pg_stat_statements`
- [ ] **Clean up large binaries** from `tmp/` and `services/_archive/` (already .gitignore'd)

### Short Term (Next Sprint)

- [ ] **Optimize top 3 slow queries** (N+1 fixes, index additions)
- [ ] **Tune connection pool sizes** based on utilization data
- [ ] **Add goroutine count metric** to Prometheus
- [ ] **Implement log rotation** for `tmp/logs/` (logrotate: daily, keep 7)
- [ ] **Add telemetry load test** (10K events/sec target)

### Medium Term (Next Quarter)

- [ ] **Refactor remaining god classes** (if any >1000 lines)
- [ ] **Add memory leak detection** to CI (heap profile diff)
- [ ] **Implement continuous profiling** (e.g., `pyroscope` or `parca` for always-on profiling)
- [ ] **Set up automated index recommendations** (using `pg_stat_statements` + `hypopg`)
- [ ] **Performance budget in CI** (fail if P95 increases >10%)

---

## 7. References & Resources

### Go Profiling

- `go tool pprof` docs: https://pkg.go.dev/net/http/pprof
- Runtime/pprof: https://pkg.go.dev/runtime/pprof
- Practical profiling: https://go.dev/blog/pprof

### PostgreSQL Performance

- pg_stat_statements: https://www.postgresql.org/docs/current/pgstatstatements.html
- EXPLAIN ANALYZE: https://www.postgresql.org/docs/current/using-explain.html
- Indexing strategies: https://use-the-index-luke.com/

### Load Testing

- k6 documentation: https://k6.io/docs/
- Distributed load testing: k6 Cloud or Grafana Cloud k6

### FlyFleet-Specific

- Architecture: `ARCHITECTURE_ASSESSMENT.md`
- Observability: `OBSERVABILITY_DESIGN.md`
- Database pools: `services/fleet-core/internal/database/pools.go`
- Load tests: `tests/k6/`, `tests/load/`
- Monitoring configs: `infrastructure/monitoring/`

---

## Appendix: Sample pprof Integration Code

```go
// Add to cmd/main.go

import (
    _ "net/http/pprof"
    "net/http"
)

func setupPprof() {
    if os.Getenv("ENABLE_PPROF") != "true" {
        return
    }

    addr := os.Getenv("PPROF_ADDR")
    if addr == "" {
        addr = "127.0.0.1:6060"
    }

    go func() {
        logger.Info("Starting pprof server", zap.String("addr", addr))
        if err := http.ListenAndServe(addr, nil); err != nil && err != http.ErrServerClosed {
            logger.Fatal("pprof server failed", zap.Error(err))
        }
    }()
}

// In main():
func main() {
    ctx := context.Background()
    cfg := loadConfig()
    logger := setupLogger(cfg)

    // ... existing init

    setupPprof()  // Add this line

    // Start main server
    startServer(ctx, cfg, logger)
}
```

Add to `.env.example`:
```
# Enable pprof profiling endpoints (development/staging only)
ENABLE_PPROF=false
PPROF_ADDR=127.0.0.1:6060
```

---

**Next Steps**: Execute Phase 1 (pprof setup) and Phase 2 (baseline) in parallel. Report findings in `flyfleet/PERFORMANCE_PROFILING_REPORT_2026-03-28.md`.
