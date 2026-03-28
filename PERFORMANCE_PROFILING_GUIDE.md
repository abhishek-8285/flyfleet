# Performance Profiling Guide - Measuring N+1 Fix Impact

**Date:** 2026-03-28
**Purpose:** Quantify performance improvements from N+1 query fixes

---

## Prerequisites

- Docker & Docker Compose installed and running
- `flyfleet` services started with pprof enabled
- `go` toolchain installed (for pprof analysis)
- Load testing tool (k6) installed

---

## Step 1: Enable pprof in fleet-core

```bash
cd /data/data/com.termux/files/home/flyfleet

# Start fleet-core with pprof enabled
ENABLE_PPROF=true PPROF_ADDR=0.0.0.0:6060 docker-compose up -d fleet-core-service

# Verify pprof is accessible
curl http://localhost:6060/debug/pprof/profile?seconds=5 -o /dev/null && echo "✅ pprof ready"
```

---

## Step 2: Capture Baseline (Before Fixes)

**IMPORTANT:** If you haven't already applied the N+1 fixes, do this FIRST. Otherwise, compare with an old build.

```bash
# Navigate to project root
cd /data/data/com.termux/files/home/flyfleet

# Create baseline directory
mkdir -p ./pprof-baseline

# Capture baseline profiles (30s CPU, heap, goroutines)
./scripts/pprof.sh baseline
# Manually move to baseline dir
mv ./pprof-results/* ./pprof-baseline/
```

Or manually:
```bash
# CPU profile (30 seconds)
curl -s http://localhost:6060/debug/pprof/profile?seconds=30 -o pprof-baseline/cpu-$(date +%Y%m%d-%H%M%S).prof

# Memory heap
curl -s http://localhost:6060/debug/pprof/heap -o pprof-baseline/heap-$(date +%Y%m%d-%H%M%S).prof

# Goroutines
curl -s http://localhost:6060/debug/pprof/goroutine?debug=1 -o pprof-baseline/goroutines-$(date +%Y%m%d-%H%M%S).txt
```

---

## Step 3: Apply N+1 Fixes

If not already applied, apply all changes from:
- `NPLUS1_FIXES_SUMMARY.md`
- Modified files (13 total)

Rebuild and restart:
```bash
cd services/fleet-core
go build -o ../bin/cmd ./cmd/main.go
ENABLE_PPROF=true PPROF_ADDR=0.0.0.0:6060 docker-compose up -d fleet-core-service
```

---

## Step 4: Capture Post-Fix Profiles

```bash
mkdir -p ./pprof-after

./scripts/pprof.sh baseline
mv ./pprof-results/* ./pprof-after/
```

---

## Step 5: Compare Profiles

### CPU Profile Comparison

```bash
# Start pprof web UI
go tool pprof -http=:8081 ./pprof-baseline/cpu-*.prof ./pprof-after/cpu-*.prof
```

Then in browser (`http://localhost:8081/ui/`):
- View top functions by flat/stack time
- Look for reductions in:
  - `Exec` (database query execution)
  - `QueryRow`, `Query` (query preparation)
  - `tx.Exec` (transaction inserts)

### Compare Query Counts

Use pg_stat_statements to see query reduction:

```sql
-- Before and after, run:
SELECT
    query,
    calls,
    rows,
    mean_exec_time,
    total_exec_time
FROM pg_stat_statements
WHERE query LIKE '%INSERT INTO%' OR query LIKE '%SELECT%FROM%'
ORDER BY total_exec_time DESC
LIMIT 20;
```

Expected: **Same queries with higher batch counts** (one batch vs N singles)

### Database Connection Pool Utilization

```sql
-- Check active connections before/after
SELECT
    state,
    count(*) as connections
FROM pg_stat_activity
WHERE datname = 'flyfleet'
GROUP BY state;
```

Expected: Fewer `active` connections under load due to reduced query count.

---

## Step 6: Load Testing with k6

### Geofencing Hot Path Test

```javascript
// tests/k6/geofence-performance.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    stages: [
        { duration: '30s', target: 50 },  // Ramp up to 50 users
        { duration: '2m', target: 50 },   // Hold
        { duration: '30s', target: 0 },   // Ramp down
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'],  // 95% under 500ms
    },
};

const payload = JSON.stringify({
    vehicle_id: '550e8400-e29b-41d4-a716-446655440000',
    latitude: 40.7128,
    longitude: -74.0060,
    timestamp: Date.now(),
    ignition: true
});

const params = {
    headers: { 'Content-Type': 'application/json' },
};

export default function () {
    const res = http.post('http://localhost:8080/api/v1/telemetry/gps', payload, params);
    check(res, { 'status 200': (r) => r.status === 200 });
    sleep(0.1);  // 10 req/sec per VU
}
```

Run test:
```bash
k6 run tests/k6/geofence-performance.js
```

**Expected improvement:** 20-100x faster depending on batch sizes.

### Reports Generation Test

```javascript
// tests/k6/reports-performance.js
export let options = {
    stages: [
        { duration: '1m', target: 20 },
        { duration: '2m', target: 20 },
    ],
};

export default function () {
    // Trigger report generation with multiple recipients
    const res = http.post('http://localhost:8080/api/v1/reports/generate', JSON.stringify({
        name: 'Test Report',
        recipients: ['a@example.com', 'b@example.com', /* ... up to 100 */],
        // ...
    }), { headers: { 'Content-Type': 'application/json' } });
    sleep(1);
}
```

**Expected:** 50-100x faster for reports with many recipients.

### Billing Invoice Creation Test

```javascript
// Test with many invoice items
const invoice = {
    customer_name: 'Test Co',
    items: Array(50).fill().map((_, i) => ({
        description: `Item ${i}`,
        quantity: 1,
        unit_price: 100,
        tax_percent: 10
    }))
};
```

**Expected:** 10-50x faster.

---

## Step 7: Quantify Results

### Metrics to Capture

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| Avg CPU time per request (ms) | | | |
| P95 latency (ms) | | | |
| DB queries per operation | | | |
| Total DB execution time (ms) | | | |
| Connections active (peak) | | | |
| Throughput (req/sec) | | | |

### Sample Results Table

```
Geofence Event Processing (100 exits):
- Before: 2500 DB queries, 1200ms total, 50 connections
- After:  100 DB queries, 60ms total, 5 connections
- Improvement: 20x faster, 96% fewer queries
```

---

## Step 8: Rollback Plan

If issues arise, revert to previous commit:

```bash
git checkout HEAD~1 -- services/fleet-core/
go build -o ../bin/cmd ./cmd/main.go
docker-compose up -d fleet-core-service
```

---

## Notes

- pprof overhead: ~1-3% CPU, negligible for profiling
- Use staging environment for testing, never production
- Capture profiles during sustained load, not idle
- Compare identical load patterns for fair results

---

## Expected Overall Impact

- **Database queries:** 70-85% reduction on affected operations
- **Latency:** 50-80% improvement
- **Connection pool:** 80%+ less pressure
- **Throughput:** 2-100x depending on batch size

**Highest ROI:** Telemetry ingestion, report generation, invoice creation, geofence event processing.
