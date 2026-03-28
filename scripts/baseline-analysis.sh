#!/bin/bash
# FlyFleet Baseline Performance Analysis
# This script captures baseline performance metrics:
# - Slow queries from pg_stat_statements
# - Database connection pool statistics
# - Service health and metrics
# - Optional: k6 load test run

set -e

echo "🔍 FlyFleet Baseline Performance Analysis"
echo "========================================"
echo ""

# Check if Docker containers are running
if ! docker ps | grep -q fleet-postgres; then
    echo "❌ PostgreSQL container not running. Start services first:"
    echo "   cd flyfleet && docker-compose up -d postgres redis nats fleet-core-service"
    exit 1
fi

# Wait for service to be ready
echo "⏳ Waiting for fleet-core-service to be healthy..."
for i in {1..30}; do
    if curl -s http://localhost:8081/healthz > /dev/null 2>&1; then
        echo "✅ Service is healthy"
        break
    fi
    sleep 2
    echo -n "."
done
echo ""

OUTPUT_DIR="./performance-baseline/$(date +%Y-%m-%d_%H%M%S)"
mkdir -p "$OUTPUT_DIR"

echo "📁 Output directory: $OUTPUT_DIR"
echo ""

# 1. Slow Query Analysis
echo "📊 1/5 Analyzing slow queries from pg_stat_statements..."
docker exec fleet-postgres psql -U fleet_user -d fleet_db -c "
COPY (
  SELECT
    mean_exec_time as mean_ms,
    calls,
    stddev_exec_time as stddev_ms,
    left(query, 120) as query
  FROM pg_stat_statements
  WHERE calls > 100
  ORDER BY mean_exec_time DESC
  LIMIT 30
) TO STDOUT WITH CSV HEADER
" > "$OUTPUT_DIR/slow-queries-raw.csv" 2>/dev/null || echo "⚠️  Could not query pg_stat_statements (may need to enable extension)"

# Additional query stats
docker exec fleet-postgres psql -U fleet_user -d fleet_db -c "
SELECT
  count(*) as total_tracked_queries,
  round(sum(calls)::numeric, 0) as total_calls,
  round(sum(total_exec_time)::numeric, 2) as total_exec_time_sec,
  round(avg(mean_exec_time)::numeric, 2) as avg_mean_ms
FROM pg_stat_statements;
" > "$OUTPUT_DIR/pg-stat-statements-summary.txt" 2>/dev/null || true

echo "   ✓ Slow queries saved to: $OUTPUT_DIR/slow-queries-raw.csv"

# 2. Database Connection Pool Stats
echo "🔌 2/5 Checking database connection pool..."
docker exec fleet-postgres psql -U fleet_user -d fleet_db -c "
SELECT
  application_name,
  state,
  count(*) as connections,
  round(avg(backend_start)::numeric, 1) as avg_age_sec
FROM pg_stat_activity
WHERE application_name LIKE 'flyfleet-%'
GROUP BY application_name, state
ORDER BY application_name, state;
" > "$OUTPUT_DIR/db-connections.txt" 2>/dev/null || true
echo "   ✓ Connection pool stats saved"

# 3. Database Load Metrics
echo "📈 3/5 Querying database load metrics..."
docker exec fleet-postgres psql -U fleet_user -d fleet_db -c "
SELECT
  datname,
  numbackends,
  xact_commit,
  xact_rollback,
  blks_hit,
  blks_read,
  round(100.0 * blks_hit / nullif(blks_hit + blks_read, 0), 2) as cache_hit_ratio
FROM pg_stat_database
WHERE datname = 'fleet_db';
" > "$OUTPUT_DIR/database-stats.txt" 2>/dev/null || true

# 4. Service Health Check
echo "🏥 4/5 Checking service health..."
curl -s http://localhost:8081/ready | jq '.' > "$OUTPUT_DIR/service-health.json" 2>/dev/null || echo "{}" > "$OUTPUT_DIR/service-health.json"
echo "   ✓ Health check saved"

# 5. Prometheus Metrics Snapshot (if available)
echo "📊 5/5 Attempting to capture Prometheus metrics..."
if curl -s http://localhost:9090/api/v1/query?query=up > /dev/null 2>&1; then
    echo "   Prometheus is running, saving key metrics..."
    cat > "$OUTPUT_DIR/prometheus-metrics.txt" << EOF
# Run these queries in Prometheus UI to track baselines:

# API Request Rate
# rate(http_requests_total[5m])

# API Error Rate
# rate(http_requests_total{status=~"5.."}[5m])

# P95 Latency
# histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))

# Database Slow Queries
# pg_stat_statements_mean_exec_time > 0.1

# Connection Pool Utilization
# pgpool_active_connections / pgpool_max_connections

EOF
    echo "   ✓ Prometheus query templates saved"
else
    echo "   ⚠️  Prometheus not running on :9090, skipping"
fi

# Summary
echo ""
echo "✅ Baseline analysis complete!"
echo ""
echo "📊 Summary:"
echo "   Output: $OUTPUT_DIR"
echo ""
echo "Next steps:"
echo "  1. Review slow-queries-raw.csv for queries >100ms"
echo "  2. Check database-connections.txt for pool saturation"
echo "  3. Run load test: cd tests/k6 && k6 run emergency_load_test.js"
echo "  4. Capture pprof profiles: ENABLE_PPROF=true docker-compose restart fleet-core-service && ./scripts/pprof.sh baseline"
echo ""
echo "📄 Full strategy: PERFORMANCE_PROFILING_STRATEGY.md"
echo ""
