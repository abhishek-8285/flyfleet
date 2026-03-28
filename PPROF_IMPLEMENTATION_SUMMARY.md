# pprof Performance Profiling - Implementation Summary

**Date:** 2026-03-28
**Status:** ✅ Complete

---

## What Was Done

### 1. Added pprof Endpoints to FlyFleet Backend

**File Modified:** `services/fleet-core/cmd/main.go`

**Changes:**
- Added import: `_ "net/http/pprof"`
- Added pprof server startup (when `ENABLE_PPROF=true`)
- pprof listens on `127.0.0.1:6060` by default (configurable via `PPROF_ADDR`)
- Includes graceful shutdown

**Code Added (around line 326):**
```go
// 6.2 Start pprof server for performance profiling (if enabled)
if os.Getenv("ENABLE_PPROF") == "true" {
    pprofAddr := os.Getenv("PPROF_ADDR")
    if pprofAddr == "" {
        pprofAddr = "127.0.0.1:6060"
    }
    pprofSrv := &http.Server{
        Addr:    pprofAddr,
        Handler: http.DefaultServeMux,
    }
    go func() {
        logger.Info("Starting pprof server", zap.String("addr", pprofAddr))
        if err := pprofSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatal("pprof server failed", zap.Error(err))
        }
    }()
    go func() {
        <-ctx.Done()
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        pprofSrv.Shutdown(shutdownCtx)
    }()
}
```

---

### 2. Updated Configuration Files

**`.env.example`** (services/fleet-core/.env.example)
Added:
```bash
# Performance profiling (development/staging only - disable in production)
ENABLE_PPROF=false
PPROF_ADDR=127.0.0.1:6060
```

**`docker-compose.yml`** (root)
Added environment variables:
```yaml
- ENABLE_PPROF=${ENABLE_PPROF:-false}
- PPROF_ADDR=${PPROF_ADDR:-127.0.0.1:6060}
```

Added optional port mapping (commented):
```yaml
ports:
  - "8081:8084"
  # Uncomment the following line to expose pprof port (disable in production!)
  # - "6060:6060"
```

---

### 3. Created Helper Scripts

**`scripts/pprof.sh`** (executable)
- Convenient wrapper for pprof commands
- Supports: cpu, heap, goroutine, alloc, block, mutex, baseline
- Auto-saves profiles with timestamps
- Built-in comparison mode

**Usage:**
```bash
cd flyfleet
./scripts/pprof.sh cpu 30          # 30s CPU profile
./scripts/pprof.sh heap            # Heap snapshot
./scripts/pprof.sh baseline        # All baselines
./scripts/pprof.sh list            # Show saved profiles
```

**`scripts/baseline-analysis.sh`** (executable)
- Captures database slow queries (pg_stat_statements)
- Reports connection pool statistics
- Checks service health
- Generates Prometheus query templates

**Usage:**
```bash
cd flyfleet
./scripts/baseline-analysis.sh
# Output: performance-baseline/YYYY-MM-DD_HHMMSS/
```

---

### 4. Created Comprehensive Strategy Document

**`PERFORMANCE_PROFILING_STRATEGY.md`** (63KB)

Contains:
- ✅ Complete pprof integration guide
- ✅ Database query optimization (EXPLAIN ANALYZE, N+1 detection)
- ✅ Load testing with k6 (including 3 new test scenarios)
- ✅ Before/After measurement template (for PRs)
- ✅ CI/CD integration (GitHub Actions performance gate)
- ✅ Optimization checklist (DB, Go runtime, API)
- ✅ FlyFleet-specific pitfalls & best practices
- ✅ SLO targets (from OBSERVABILITY_DESIGN.md)

---

## How to Enable pprof

### 1. Local Development (manual binary)

```bash
cd services/fleet-core
ENABLE_PPROF=true go run ./cmd/main.go
# pprof available at http://localhost:6060/debug/pprof/
```

### 2. Docker Compose (staging)

```bash
# Edit docker-compose.yml, uncomment port mapping:
#   - "6060:6060"

# Or use environment override:
ENABLE_PPROF=true PPROF_ADDR=0.0.0.0:6060 docker-compose up fleet-core-service

# Access from host:
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof -http=:8081 cpu.prof
```

### 3. Kubernetes (production - for debugging only)

Add to deployment:
```yaml
env:
  - name: ENABLE_PPROF
    value: "true"
  - name: PPROF_ADDR
    value: "127.0.0.1:6060"
# Port not exposed - use kubectl port-forward for secure access:
kubectl port-forward deploy/fleet-core-service 6060:6060
```

---

## Quick Start Commands

```bash
# 1. Build and start with pprof
cd flyfleet
ENABLE_PPROF=true PPROF_ADDR=0.0.0.0:6060 docker-compose up -d fleet-core-service

# 2. Wait for health
curl http://localhost:8081/healthz

# 3. Capture CPU profile (30s)
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/profile?seconds=30

# 4. Or use helper script
./scripts/pprof.sh cpu 30

# 5. View in browser: http://localhost:8081/ui/pprof
```

---

## Performance Targets (from OBSERVABILITY_DESIGN.md)

| Metric | Target | Measurement |
|--------|--------|-------------|
| API Availability | 99.95% | `rate(http_requests_total{status!~"5.."}[1d])` |
| P95 Latency | <300ms | `histogram_quantile(0.95, ...)` |
| P99 Latency | <1000ms | `histogram_quantile(0.99, ...)` |
| Error Rate | <0.1% | `rate(http_requests_total{status=~"5.."}[5m])` |
| Telemetry Ingestion | >10,000/sec | `rate(nats_jetstream_messages_received_total[1m])` |
| DB Query P95 | <100ms | `pg_stat_statements_mean_exec_time` |
| Pool Utilization | <80% | `active_conns / max_conns` |

---

## Next Steps

1. **Run baseline analysis** to capture current state:
   ```bash
   ./scripts/baseline-analysis.sh
   ```

2. **Analyze slow queries** from `performance-baseline//slow-queries-raw.csv`
   - Look for `mean_ms > 100`
   - Use `EXPLAIN ANALYZE` to understand plans
   - Add missing indexes

3. **Run load tests** to validate against SLOs:
   ```bash
   cd tests/k6
   k6 run emergency_load_test.js --out json=../results/emergency.json
   k6 run finance_load_test.js
   k6 run platform_load_test.js
   ```

4. **Profile under load**:
   ```bash
   # Start service with pprof
   ENABLE_PPROF=true docker-compose restart fleet-core-service

   # Run k6 while capturing profiles
   k6 run tests/k6/platform_load_test.js &
   ./scripts/pprof.sh cpu 60
   ```

5. **Fix N+1 queries** - biggest win typically 90%+ latency improvement
   - Search: `grep -r "for.*range" services/fleet-core/internal/domains/ -A3 | grep -B2 "Query"`
   - Convert loops with queries to single JOIN or batch queries

---

## Files Modified/Created

| File | Change |
|------|--------|
| `services/fleet-core/cmd/main.go` | Added pprof import and server setup |
| `services/fleet-core/.env.example` | Added ENABLE_PPROF, PPROF_ADDR |
| `docker-compose.yml` | Added pprof env vars, commented port mapping |
| `scripts/pprof.sh` | **NEW** - pprof helper script (executable) |
| `scripts/baseline-analysis.sh` | **NEW** - baseline capture script (executable) |
| `PERFORMANCE_PROFILING_STRATEGY.md` | **NEW** - Complete 63KB strategy doc |
| `PPROF_IMPLEMENTATION_SUMMARY.md` | This file |

---

## Verification

### Build Test
```bash
cd services/fleet-core
go build ./cmd/main.go
# ✅ Success (no errors)
```

### Configuration Check
```bash
# pprof is disabled by default - safe for production
grep ENABLE_PPROF docker-compose.yml
# Output: ENABLE_PPROF=${ENABLE_PPROF:-false}  ✅ Default: false
```

### Helper Scripts
```bash
./scripts/pprof.sh list
# Output: 📂 Available profiles in ./pprof-results/
#         (no profiles found - that's OK, first run will create)
./scripts/baseline-analysis.sh --help 2>/dev/null || echo "Run without args to execute"
```

---

## Security Notes

- ✅ pprof **disabled by default** (`ENABLE_PPROF=false`)
- ✅ Default bind: `127.0.0.1:6060` (localhost only)
- ✅ Not exposed in production without explicit action
- ✅ Separate from main API port (no interference)
- ✅ Contains sensitive information (memory, goroutines, stack traces)
- ⚠️ If exposing externally, use network policies or additional auth layer

**Never enable pprof in production without:**
1. Network-level access control (VPN, internal network, mTLS)
2. Temporary enablement (for debugging only)
3. Proper audit logging of who accessed pprof endpoints

---

## Support & Documentation

- Full strategy: `PERFORMANCE_PROFILING_STRATEGY.md`
- pprof docs: https://pkg.go.dev/net/http/pprof
- FlyFleet Observability: `OBSERVABILITY_DESIGN.md`
- Architecture: `ARCHITECTURE_ASSESSMENT.md`

---

**Ready to profile!** 🚀

Start with: `./scripts/baseline-analysis.sh`
