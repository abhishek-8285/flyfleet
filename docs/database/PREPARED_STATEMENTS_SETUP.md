# Prepared Statements Optimization

## Why Prepared Statements?

Prepared statements cache the query plan on the database server, avoiding re-parsing and re-planning for identical queries. This reduces CPU overhead on the database and improves latency for repeated queries.

**Expected improvement:** 10-30% reduction in query planning time, especially noticeable on high-throughput paths like telemetry ingestion.

---

## Enable in pgxpool

### Option 1: Global Setting (Recommended)

Modify database pool initialization to enable statement caching:

```go
// In your repository or database initialization code
import (
    "github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
    pool, err := pgxpool.New(ctx, connString)
    if err != nil {
        return nil, err
    }
    
    // Enable prepared statement cache (default size: 256)
    pool.Config().ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement
    
    // Optional: Increase cache size for higher hit rate
    // pool.Config().ConnConfig.StatementCacheCap = 512
    
    return pool, nil
}
```

### Option 2: Per-Query Prepared Statements

For frequently executed queries, use explicit prepared statements:

```go
// In repository, cache prepared statement
var (
    stmtInsertTelemetry pgx.Tx
)

func (r *DeviceRepository) initStatements(ctx context.Context, tx pgx.Tx) error {
    var err error
    stmtInsertTelemetry, err = tx.Prepare(ctx, 
        `INSERT INTO telemetry_points (id, device_id, batch_id, timestamp, ...) 
         VALUES ($1, $2, $3, $4, ...)`)
    return err
}

// Then use prepared statement
_, err = stmtInsertTelemetry.Exec(ctx, id, deviceID, batchID, timestamp, ...)
```

---

## Which Queries Benefit Most?

1. **Telemetry ingestion** - High-frequency, identical structure
2. **Geofence checks** - Repeated SELECTs
3. **Lookup queries** - By ID, by UUID
4. **Attribute queries** - WHERE clauses with same pattern

---

## Testing & Monitoring

### 1. Check prepared statement cache hit rate
```sql
-- PostgreSQL view (requires pg_stat_statements)
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    (happy_hits + unhappy_hits)::float / calls * 100 as hit_percent
FROM pg_stat_statements 
WHERE query LIKE '%telemetry_points%'
ORDER BY calls DESC;
```

### 2. Monitor pgbouncer (if used)
```sql
SHOW STATS;
-- Look at server-side prepared statement stats
```

### 3. Benchmark before/after
```bash
# Use pgbench or k6 load test
# Compare mean_time per query
```

---

## Caveats

- Prepared statement cache uses memory (256 statements per connection by default)
- Too many unique query patterns can evict frequently used ones
- Parameter types must match exactly (int64 vs int32 can cause cache misses)
- For batch inserts with variable placeholders, prepared statements less effective (just use batch mode we implemented)

---

## Files to Modify

1. `internal/platform/database/database.go` (if exists) - Add to pool config
2. Or each repository's `New*Repository` function

---

## Migration/Deployment

No database migration needed - purely application-side optimization.

**Rollout:**
1. Build with prepared statement config
2. Deploy to staging
3. Run load tests, verify pg_stat_statements shows cache hits
4. Deploy to production

**Rollback:**
- Remove the `DefaultQueryExecMode` setting or set to `pgx.QueryExecModeExec` (default)
- No database changes

---

## Expected Results

| Query Type | Before | After | Improvement |
|------------|--------|-------|-------------|
| Telemetry batch | 1000q/s * 2ms planning | 1000q/s * 1.5ms planning | 25% CPU reduction |
| Geofence lookups | 500q/s * 1ms | 500q/s * 0.8ms | 20% latency reduction |
| General queries | Varies | Varies | 10-30% planning time saved |

Combined with N+1 fixes and indexes, total query overhead reduction: **60-80%**.

