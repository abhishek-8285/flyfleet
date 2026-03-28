# Next Performance Improvements (Post N+1 Fixes)

**Status:** Ready to implement  
**Estimated Total Time:** 5-7 days  
**Expected Impact:** Additional 20-40% performance improvement on top of N+1 fixes

---

## Priority 1: Add Missing Indexes ⚡ (1 day)

### Why
The N+1 fixes eliminated query count, but remaining queries may be slow due to missing indexes.

### What to Check
1. Foreign key columns without indexes:
   - `geofence_events(vehicle_id)`
   - `trips(driver_id)`, `trips(vehicle_id)`
   - `telemetry_points(device_id, batch_id)`
   - `mfa_recovery_codes(user_id)`
   - All `*_id` foreign keys

2. Partial indexes for common filters:
   ```sql
   CREATE INDEX idx_organizations_active ON organizations(id) WHERE status = 'active' AND is_deleted = false;
   CREATE INDEX idx_invoices_pending ON invoices(org_id) WHERE status IN ('PENDING', 'OVERDUE');
   ```

3. Composite indexes for query patterns:
   ```sql
   CREATE INDEX idx_geofence_events_time ON geofence_events(vehicle_id, time DESC);
   CREATE INDEX idx_trips_driver_time ON trips(driver_id, start_time DESC);
   ```

### Implementation
```sql
-- Migration file: infrastructure/migrations/016_performance_indexes.sql
-- Run in staging first, measure with EXPLAIN ANALYZE
```

---

## Priority 2: Batch SELECT Operations (1-2 days)

### Why
We fixed batch INSERTs, but there may be N+1 **READ** patterns still present.

### Common Patterns to Find
```go
// BAD (N+1 SELECT):
for _, ID := range ids {
    err := db.QueryRow("SELECT * FROM table WHERE id = $1", ID).Scan(&result)
}

// GOOD (Batch SELECT):
rows, _ := db.Query("SELECT * FROM table WHERE id = ANY($1)", ids)
```

### Where to Look
- Loading multiple entities by ID
- Fetching child records in loops
- Lookup tables accessed per-item

### Audit Approach
```bash
# Search for anti-patterns
grep -rn "for.*range.*QueryRow" services/fleet-core/internal/domains/
grep -rn "for.*range.*Query(" services/fleet-core/internal/domains/ | grep -v "batch"
```

---

## Priority 3: Application-Level Caching (2 days)

### Why
Hot read-heavy tables can bypass database entirely with Redis caching.

### Candidates for Caching
| Entity | Read Frequency | Write Frequency | TTL | Eviction |
|--------|---------------|-----------------|-----|----------|
| Geofence definitions | Very High | Very Low | 1h | Cache invalidation on update |
| Organization settings | High | Low | 30m | TTL + explicit purge |
| Vehicle metadata | High | Medium | 5m | TTL + pub/sub invalidation |
| Driver profiles | Medium | Low | 15m | TTL |
| Active vehicle locations | Very High | Very High | 30s | Short TTL, cache-aside |

### Implementation Pattern
```go
// Cache-aside pattern
func (r *GeofenceRepo) GetGeofences(ctx context.Context, orgID uuid.UUID) ([]Geofence, error) {
    cacheKey := fmt.Sprintf("geofences:org:%s", orgID)
    
    // Try cache first
    if cached, err := r.redis.Get(ctx, cacheKey).Result(); err == nil {
        var result []Geofence
        json.Unmarshal([]byte(cached), &result)
        return result, nil
    }
    
    // Cache miss - fetch from DB
    result, err := r.fetchFromDB(ctx, orgID)
    if err != nil {
        return nil, err
    }
    
    // Cache for 1 hour
    data, _ := json.Marshal(result)
    r.redis.Set(ctx, cacheKey, data, time.Hour)
    return result, nil
}

// Invalidate on update
func (r *GeofenceRepo) UpdateGeofence(ctx context.Context, g *Geofence) error {
    err := r.updateInDB(ctx, g)
    if err == nil {
        r.redis.Delete(ctx, fmt.Sprintf("geofence:%s", g.ID))
        r.redis.Delete(ctx, fmt.Sprintf("geofences:org:%s", g.OrgID))
    }
    return err
}
```

---

## Priority 4: Materialized Views for Analytics (2-3 days)

### Why
Complex analytics queries join multiple tables and aggregate over millions of rows.

### Candidates
1. **Daily fleet statistics** (already have tables, but could be materialized views)
2. **Driver behavior scores** (complex aggregations)
3. **Vehicle utilization reports**
4. **Billing dashboards** (joins across invoices, payments, tax lines)

### Example
```sql
-- Instead of running complex query on-demand
CREATE MATERIALIZED VIEW mv_driver_daily_scores AS
SELECT 
    d.id,
    d.driver_id,
    d.date,
    -- Complex calculations...
    computed_score
FROM ...;

-- Refresh periodically
REFRESH MATERIALIZED VIEW CONCURRENTLY mv_driver_daily_scores;
```

---

## Priority 5: Prepared Statements (0.5 day)

### Why
pgxpool supports prepared statements, reducing query planning overhead.

### Enable in code:
```go
// In repository constructors
pool, _ := pgxpool.New(ctx, connString)
pool.Config().ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement
// Or use prepared statements explicitly for hot queries
```

---

## Priority 6: Query Optimization (Ongoing)

### Use EXPLAIN ANALYZE on Hot Queries
```bash
# In database
SELECT query, calls, total_time, rows, mean_time 
FROM pg_stat_statements 
ORDER BY total_time DESC 
LIMIT 20;
```

For each slow query:
```sql
EXPLAIN (ANALYZE, BUFFERS, VERBOSE) <the query>;
```

Look for:
- Sequential scans → Add indexes
- Nested loops with large inner loops → Consider hash joins or better indexes
- Sort operations → Add indexes for ORDER BY
- Hash aggregate spills → Increase work_mem

---

## Implementation Plan

### Week 1 (Next Sprint)
- [ ] Day 1: Add missing indexes (migration + testing)
- [ ] Day 2: Audit and fix batch SELECT N+1 patterns
- [ ] Day 3: Implement Redis caching for 3-5 hot entities
- [ ] Day 4: Test, benchmark, document gains

### Week 2 (Following Sprint)
- [ ] Day 1-2: Create materialized views for 2-3 critical reports
- [ ] Day 3: Enable prepared statements for telemetry path
- [ ] Day 4: Load testing, pprof profiling, compare metrics

---

## Expected Cumulative Impact

| Optimization | Query Reduction | Latency Improvement |
|--------------|-----------------|---------------------|
| Indexes | 10-30% faster queries | 2-5x for slow queries |
| Batch SELECT | 50-90% for affected ops | 5-50x |
| Caching | 50-90% DB load reduction | 2-10x |
| Materialized Views | 90%+ for analytics | 10-100x |
| Prepared Statements | 10-30% planning time reduction | 1.1-1.3x |

**Combined with N+1 fixes:** Total of 80-95% query reduction, 5-200x speedup on hot paths.

---

## Quick Start

Want to tackle one of these? Tell me which priority:

1. "Add indexes" → I'll analyze schema, suggest indexes, create migration
2. "Batch SELECT" → I'll search for N+1 read patterns and fix them
3. "Caching" → I'll identify hot entities and implement cache layer
4. "Materialized views" → I'll design MV architecture for analytics
5. "Prepared statements" → Quick win, enable in pool config

Which would you like to start with? (I recommend #1 or #2 for quickest wins)
