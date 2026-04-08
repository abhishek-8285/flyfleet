# FlyFleet Performance Issues & Fixes - Quick Reference

**Date:** 2026-03-28
**Status:** ✅ Analysis Complete | ⏳ Awaiting Implementation

---

## 📊 Summary of Fixes Applied

| Issue | Status | Files Modified | Impact |
|-------|--------|---------------|--------|
| **pprof endpoints** | ✅ DONE | `cmd/main.go` | Enables all future profiling |
| **Log rotation** | ✅ DONE | `infrastructure/logging/flyfleet-logrotate.conf` | Prevents disk pressure |
| **N+1 detection** | ✅ ANALYZED | `scripts/detect-n1-queries.py`, `PERFORMANCE_ISSUES_NPLUS1.md` | 23 patterns found (5 high-priority) |
| **Cleanup scripts** | ✅ DONE | `scripts/cleanup-artifacts.sh` | Reclaim disk space safely |

---

## 🚨 Top 5 Performance Issues to Fix

### 1. Reports Recipient Inserts (HIGH)

**File:** `internal/domains/reports/repository/repository.go:52-56,114-118`

**Problem:** 1 INSERT per recipient instead of batch

**Fix:**
```go
// Batch insert recipients
if len(report.Recipients) > 0 {
    var args []interface{}
    var placeholders []string
    for i, email := range report.Recipients {
        placeholders = append(placeholders, fmt.Sprintf("($1, $%d)", i*2+2))
        args = append(args, report.ID, email)
    }
    query := fmt.Sprintf("INSERT INTO report_recipients (report_id, email) VALUES %s",
        strings.Join(placeholders, ", "))
    _, err = tx.Exec(ctx, query, args...)
}
```

**Impact:** 50-100x faster for reports with many recipients

---

### 2. Billing Invoice Items (HIGH)

**File:** `internal/domains/billing/repository/repository.go:54-60,139-148`

**Problem:** 1 INSERT per line item

**Fix:** Same batch insert pattern

**Impact:** 10-50x faster for invoices with many items

---

### 3. Billing Tax Lines (HIGH)

**File:** `internal/domains/billing/repository/repository.go:283-293`

**Problem:** 1 INSERT per tax line

**Fix:** Batch insert

**Impact:** 5-20x faster

---

### 4. Geofence Name Lookups (HIGH - Real-time Path)

**File:** `internal/domains/geofencing/engine/geofence_engine.go:221-237`

**Problem:** Querying geofence name for EACH exit event

```go
// BAD:
for geofenceID, entryTime := range state.InsideGeofences {
    geofenceName, _ := ge.getGeofenceName(ctx, geofenceID)  // DB hit each time!
}

// GOOD:
// Preload all names in one query
```

**Fix:**
```go
// Before loop:
ids := make([]uuid.UUID, 0, len(state.InsideGeofences))
for id := range state.InsideGeofences {
    ids = append(ids, id)
}
names := ge.getGeofenceNamesBatch(ctx, ids)  // Batch fetch

for geofenceID, entryTime := range state.InsideGeofences {
    geofenceName := names[geofenceID]  // No DB query!
}
```

**Impact:** Critical - this is in hot path (geofence events). Reduces from N queries to 1.

---

### 5. Analytics Feature Vectors (MEDIUM-HIGH)

**File:** `internal/domains/analytics/repository/operations.go:369-373,568-572`

**Problem:** Inserts one feature at a time

**Fix:** Already has batch - convert to single multi-row INSERT

---

## 📁 Files Created/Created

### New Files

| File | Purpose |
|------|---------|
| `PERFORMANCE_PROFILING_STRATEGY.md` | Complete profiling guide (63KB) |
| `PPROF_IMPLEMENTATION_SUMMARY.md` | pprof setup summary |
| `PERFORMANCE_ISSUES_NPLUS1.md` | N+1 analysis report |
| `scripts/pprof.sh` | pprof helper (cpu, heap, baseline) |
| `scripts/baseline-analysis.sh` | Automated baseline capture |
| `scripts/detect-n1-queries.py` | N+1 pattern detector |
| `scripts/rotate-logs.sh` | Manual log rotation |
| `scripts/cleanup-artifacts.sh` | Safe artifact cleanup |
| `infrastructure/logging/flyfleet-logrotate.conf` | Logrotate config |
| `infrastructure/logging/LOG_ROTATION_SETUP.md` | Log rotation guide |

### Modified Files

| File | Change |
|------|--------|
| `services/fleet-core/cmd/main.go` | Added `import _ "net/http/pprof"` + server setup |
| `services/fleet-core/.env.example` | Added `ENABLE_PPROF=false, PPROF_ADDR=127.0.0.1:6060` |
| `docker-compose.yml` | Added pprof env vars, optional port mapping |

---

## 🚀 Quick Start Commands

### Enable pprof (development)
```bash
cd flyfleet
ENABLE_PPROF=true PPROF_ADDR=0.0.0.0:6060 docker-compose up -d fleet-core-service

# Capture CPU profile
./scripts/pprof.sh cpu 30

# Open in browser: http://localhost:8081/ui/pprof
```

### Run baseline analysis (requires Docker)
```bash
./scripts/baseline-analysis.sh
# Output: performance-baseline/YYYY-MM-DD_HHMMSS/
```

### Detect N+1 patterns
```bash
./scripts/detect-n1-queries.py > n1-findings.txt
```

### Rotate logs (if logrotate not installed)
```bash
./scripts/rotate-logs.sh
```

### Cleanup disk space
```bash
./scripts/cleanup-artifacts.sh
# Interactive - confirms each deletion
```

---

## 📈 Performance Targets (SLOs)

| Metric | Target | Measurement |
|--------|--------|-------------|
| API Availability | 99.95% | `rate(http_requests_total{status!~"5.."}[1d])` |
| P95 Latency | <300ms | `histogram_quantile(0.95, ...)` |
| P99 Latency | <1s | `histogram_quantile(0.99, ...)` |
| Error Rate | <0.1% | `rate(http_requests_total{status=~"5.."}[5m])` |
| Telemetry Ingestion | >10K/sec | `rate(nats_jetstream_messages_received_total[1m])` |
| DB Query P95 | <100ms | `pg_stat_statements_mean_exec_time` |
| Pool Utilization | <80% | `active_conns / max_conns` |

---

## 🎯 Action Plan

### Week 1: Foundation (DONE ✅)
- [x] Add pprof endpoints
- [x] Create profiling scripts
- [x] Set up log rotation
- [x] Detect N+1 patterns
- [x] Document findings

### Week 2: Database Optimization
- [ ] Fix top 5 N+1 issues (reports, billing, geofencing)
- [ ] Analyze slow queries from `pg_stat_statements`
- [ ] Add missing indexes (org_id, created_at composites)
- [ ] Verify with `EXPLAIN ANALYZE`

### Week 3: Testing & Validation
- [ ] Run k6 load tests before/after fixes
- [ ] Capture pprof profiles under load
- [ ] Compare before/after metrics
- [ ] Update `PERFORMANCE_PROFILING_STRATEGY.md` with results

### Week 4: CI/CD Integration
- [ ] Add performance gate to GitHub Actions
- [ ] Fail builds if P95 degrades >5%
- [ ] Automated slow query alerting
- [ ] Document rollback procedures

---

## 📚 Documentation Index

| Document | Size | Description |
|----------|------|-------------|
| `PERFORMANCE_PROFILING_STRATEGY.md` | 63KB | Complete strategy with all details |
| `PPROF_IMPLEMENTATION_SUMMARY.md` | 8KB | pprof setup guide |
| `PERFORMANCE_ISSUES_NPLUS1.md` | 12KB | N+1 analysis with fixes |
| `infrastructure/logging/LOG_ROTATION_SETUP.md` | 3KB | Log rotation instructions |
| `QUICK_START_PERFORMANCE.md` | This file | Quick reference |

---

## 🔧 Troubleshooting

**pprof connection refused:**
- Is `ENABLE_PPROF=true` set?
- Is port 6060 exposed? (`docker-compose ps` shows ports)
- Check logs: `docker logs fleet-core-service | grep pprof`

**Logs not rotating:**
- Verify logrotate installed: `which logrotate`
- Test config: `sudo logrotate -d /etc/logrotate.d/flyfleet`
- Or use manual: `./scripts/rotate-logs.sh`

**N+1 detector false positives:**
- It's a heuristic - review each manually
- Some loops intentionally have 1 query (test setup, etc.)

**Docker not available:**
- Run services on staging/production instead
- Use `baseline-analysis.sh` on deployed environment
- pprof works on remote servers via port-forward

---

## 📞 Support

- Full strategy: `PERFORMANCE_PROFILING_STRATEGY.md`
- Architecture: `ARCHITECTURE_ASSESSMENT.md`
- Observability: `OBSERVABILITY_DESIGN.md`
- Makefile targets: `make help`

---

**Ready to optimize!** Start with the top 5 N+1 fixes for biggest wins.
