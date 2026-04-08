# Runbook: Telemetry Ingestion Disruption

**Severity:** SEV2 (major impact) or SEV1 if complete failure
**Service:** Telemetry Ingestion Pipeline (NATS JetStream → fleet-core → Database)
**SLO Violation:** Ingestion Rate > 10,000 events/sec, P99 Latency < 5s
**Alert:** Telemetry lag > 5min or ingestion rate drop > 50%

---

## Quick Reference

| What | How |
|------|-----|
| **Check ingestion rate** | `curl -s 'http://prometheus:9090/api/v1/query?query=rate(nats_jetstream_messages_received_total[1m])'` |
| **Check consumer lag** | `kubectl exec -it nats-0 -n flyfleet -- nats consumer report TELEMETRY FLEET_CORE_CONSUMER` |
| **Check lag metric** | `curl -s 'http://prometheus:9090/api/v1/query?query=telemetry_gaps_total'` |
| **Check NatS storage** | `kubectl exec -it nats-0 -n flyfleet -- nats js info` |
| **Restart consumer** | `kubectl rollout restart deployment/fleet-core -n flyfleet` |
| **Check device connectivity** | `nats server report` and review connection stats |

---

## Symptoms

- Telemetry events not appearing in database
- Dashboard showing stale location data (> 5 min old)
- NATS JetStream storage growing rapidly (unprocessed messages)
- Consumer pending count increasing
- Prometheus alerts: `TelemetryIngestionLagHigh` or `TelemetryThroughputDrop`
- Devices reporting "connection failed" or "publish timeout"

---

## Decision Tree

```
Telemetry Stopped?
├─ Yes → Complete outage?
│   ├─ Yes → NATS server running?
│   │   ├─ No → Start NATS pod
│   │   └─ Yes → Stream/consumer exist?
│   │       ├─ No → Recreate stream/consumer (see NATS runbook)
│   │       └─ Yes → Fleet-core running?
│   │           ├─ No → Start/restart fleet-core
│   │           └─ Yes → Fleet-core connected to NATS?
│   │               └─ Check NATS connection errors in logs
│   └─ No → Partial outage?
│       ├─ Specific device type? → Check device compatibility
│       ├─ Specific region? → Check network/firewall
│       └─ Time-based? → Check device firmware versions
└─ No → Slow ingestion / high latency?
    ├─ Latency > 5s? → Check:
    │   ├─ Database slow → Database runbook
    │   ├─ Fleet-core CPU high → Scale up
    │   └─ NATS ack delays → NATS consumer lag
    └─ Throughput lower than expected?
        ├─ NATS storage full → Expand storage or prune
        ├─ Rate limiting → Check device config
        └─ Network issues → VPC/firewall
```

---

## Investigation Steps

### 1. Verify Ingestion Metrics (2 min)

```bash
# Current ingestion rate
curl -s 'http://prometheus:9090/api/v1/query?query=rate(telemetry_events_received_total[1m])' | jq '.data.result[0].value[1]'

# Historical rate (last 1h)
curl -s 'http://prometheus:9090/api/v1/query?query=rate(telemetry_events_received_total[1h])' | jq

# End-to-end latency P99
curl -s 'http://prometheus:9090/api/v1/query?query=histogram_quantile(0.99, sum(rate(telemetry_ingestion_latency_seconds_bucket[1h])) by (le))' | jq

# Gaps detected (missing sequence numbers)
curl -s 'http://prometheus:9090/api/v1/query?query=telemetry_gaps_total' | jq

# Grafana → Telemetry Ingestion Dashboard
```

**Baseline:** 10,000+ events/sec, P99 < 5s, gaps < 0.001%

---

### 2. Check NATS JetStream Health (3 min)

```bash
# Stream summary
kubectl exec -it nats-0 -n flyfleet -- nats stream report TELEMETRY

# Consumer details
kubectl exec -it nats-0 -n flyfleet -- nats consumer report TELEMETRY FLEET_CORE_CONSUMER

# Pending messages (lag)
kubectl exec -it nats-0 -n flyfleet -- nats consumer report TELEMETRY FLEET_CORE_CONSUMER | grep "Pending"

# Storage usage
kubectl exec -it nats-0 -n flyfleet -- nats js info

# Server connections (devices connected)
kubectl exec -it nats-0 -n flyfleet -- nats server report --json | jq '.connections'
```

**Critical checks:**
- `Pending Msgs` rapidly increasing? Consumer falling behind
- `Storage` > 85%? Risk of full disk → Storage exhaustion crisis
- `State` of consumer: `Active` (good) vs `Stalled` (problem)
- `Dropped` count? Messages being dropped due to backlog

---

### 3. Check Fleet-Core Consumption (3 min)

```bash
# Fleet-core pod status
kubectl get pods -n flyfleet -l app=fleet-core

# Check for NATS connection errors in logs
kubectl logs -f deployment/fleet-core -n flyfleet --since=10m 2>&1 | grep -i "nats\|jetstream\|connection"

# Check telemetry processing metrics
curl -s 'http://prometheus:9090/api/v1/query?query=rate(telemetry_processing_events_total[1m])' | jq

# Consumer lag metric (if instrumented)
curl -s 'http://prometheus:9090/api/v1/query?query=nats_jetstream_consumer_pending_bytes{stream="TELEMETRY",consumer="FLEET_CORE_CONSUMER"}' | jq

# Check fleet-core resource utilization
kubectl top pods -n flyfleet -l app=fleet-core
```

**Common issues:**
- `nats: no servers available` → NATS connectivity problem
- `jetstream: pull consumer stalled` → Consumer not processing
- High CPU? → Processing bottleneck, need optimization or scale

---

### 4. Check Database Ingestion (3 min)

```bash
# Check ingestion rate into database (new data per minute)
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT COUNT(*) FROM telemetry_events WHERE created_at > NOW() - INTERVAL '1 minute';"

# Check for slow inserts (if pg_stat_statements enabled)
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT query, mean_exec_time, calls FROM pg_stat_statements WHERE query LIKE '%telemetry_events%' ORDER BY mean_exec_time DESC LIMIT 5;"

# Check table size growth rate
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT pg_size_pretty(pg_total_relation_size('telemetry_events'));"

# Check for table locks
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT pid, relation, mode, granted FROM pg_locks WHERE relation = 'telemetry_events'::regclass;"
```

**Issues:**
- Table size growing too fast? → Partitioning or TTL needed
- Insert taking > 100ms? → Index issues or I/O bottleneck
- Lock contention? → Too many writers, need optimization

---

### 5. Check Device Connectivity (2 min)

```bash
# NATS connection count from devices
kubectl exec -it nats-0 -n flyfleet -- nats server report --json | jq '.connections[] | {cid, account, ip, port, uptime}'

# Check for connection errors in NATS logs
kubectl logs -f deployment/nats -n flyfleet --since=10m 2>&1 | grep -i "error\|connection\|auth"

# If using MQTT bridge, check MQTT connections
kubectl logs -f deployment/mqtt-bridge -n flyfleet --since=10m

# Check if specific device types failing
# Query database for last telemetry by device type
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT device_type, MAX(created_at) FROM telemetry_events GROUP BY device_type ORDER BY MAX(created_at) DESC;"
```

---

## Mitigation Strategies

### Strategy 1: Restart Telemetry Consumer (Fleet-Core)

**When:** Consumer lag growing, but NATS healthy and messages available

```bash
# Restart fleet-core to reinitialize consumer
kubectl rollout restart deployment/fleet-core -n flyfleet
kubectl rollout status deployment/fleet-core -n flyfleet --timeout=5m

# Monitor consumer lag after restart
watch -n 5 'kubectl exec -it nats-0 -n flyfleet -- nats consumer report TELEMETRY FLEET_CORE_CONSUMER | grep Pending'
```

**Expected:** Lag should start decreasing within 1-2 minutes

---

### Strategy 2: Increase Consumer Acknowledgment Batch Size

**When:** Consumer keeping up but needs more throughput

```bash
# Edit fleet-core deployment to increase batch size
kubectl edit deployment/fleet-core -n flyfleet

# Add/modify env variable:
# - name: NATS_MAX_BATCH_SIZE
#   value: "100"  # default 50, increase to 100 or 200
# - name: NATS_ACK_WAIT
#   value: "30s"  # ack wait time

kubectl rollout restart deployment/fleet-core -n flyfleet
```

**Note:** Balance throughput vs. latency. Higher batches = better throughput, slightly higher latency.

---

### Strategy 3: Scale Fleet-Core Horizontally

**When:** Processing bottleneck (CPU high, single consumer too slow)

```bash
# Check current replicas
kubectl get deployment fleet-core -n flyfleet

# Scale to 3 replicas (careful: each gets separate consumer)
kubectl scale deployment fleet-core --replicas=3 -n flyfleet

# Monitor each consumer's lag
for i in $(seq 0 2); do
  echo "Replica $i:"
  # NATS is shared, so lag distributed across consumers
done

# Check total processing rate
curl -s 'http://prometheus:9090/api/v1/query?query=sum(rate(telemetry_processing_events_total[1m]))' | jq
```

**Warning:** Multiple consumers create competing subscribers. Ensure NATS stream has `Storage: File` and `Retention: Limits` with sufficient `MaxMsgs`.

---

### Strategy 4: Expand NATS JetStream Storage

**When:** Storage > 85% and growing (risk of disk full)

```bash
# Check current storage limits
kubectl exec -it nats-0 -n flyfleet -- nats stream info TELEMETRY

# Edit Helm values to increase storage
cd infrastructure/helm/flyfleet
# Edit values.yaml or pass via --set

# If using PVC, expand it (requires manual intervention)
kubectl edit pvc nats-jetstream-0 -n flyfleet
# Increase storage: spec.resources.requests.storage: "100Gi" → "200Gi"

# If using hostPath, expand disk or add more storage

# After expansion, NATS will auto-use new space
kubectl exec -it nats-0 -n flyfleet -- nats stream info TELEMETRY
```

**Alternative:** Add second JetStream storage tier with different retention policy

---

### Strategy 5: Purge Stale/Corrupt Messages

**When:** Data corruption suspected, specific messages causing processing failures

⚠️ **WARNING:** This deletes messages from the stream. Use only if data is corrupt and re-ingestion possible.

```bash
# Purge messages older than X days (e.g., 7 days)
kubectl exec -it nats-0 -n flyfleet -- \
  nats stream purge TELEMETRY --older 168h --yes

# OR purge specific subject if only subset corrupt
kubectl exec -it nats-0 -n flyfleet -- \
  nats stream purge TELEMETRY --filter "telemetry.corrupt.>" --yes

# After purge, verify storage freed
kubectl exec -it nats-0 -n flyfleet -- nats stream info TELEMETRY
```

---

### Strategy 6: Bypass NATS Temporarily (Emergency Ingest)

**When:** NATS completely down and telemetry critical, use direct DB insertion temporarily

⚠️ **USE ONLY FOR EMERGENCY** - Bypasses reliability guarantees, data loss possible

```bash
# Deploy emergency ingestion service
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: emergency-ingest
  namespace: flyfleet
spec:
  replicas: 2
  selector:
    matchLabels:
      app: emergency-ingest
  template:
    metadata:
      labels:
        app: emergency-ingest
    spec:
      containers:
      - name: ingest
        image: flyfleet/emergency-ingest:latest
        command: ["node", "direct-db-ingest.js"]
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: flyfleet-secrets
              key: database-url
EOF

# Monitor
kubectl logs -f deployment/emergency-ingest -n flyfleet
```

**After NATS recovery:**
- Stop emergency-ingest
- Compare data counts to ensure no duplicates or gaps
- Replay missed events from devices if possible

---

## Recovery & Validation

### Validation Steps

1. **Ingestion Rate Recovery**
```bash
# Should be > 10,000 events/sec
curl -s 'http://prometheus:9090/api/v1/query?query=rate(telemetry_events_received_total[1m])' | jq

# Check last 5 min trend
curl -s 'http://prometheus:9090/api/v1/query?query=increase(telemetry_events_received_total[5m])' | jq
```

2. **Latency Normalization**
```bash
# P99 latency < 5s
curl -s 'http://prometheus:9090/api/v1/query?query=histogram_quantile(0.99, sum(rate(telemetry_ingestion_latency_seconds_bucket[5m])) by (le))' | jq
```

3. **Consumer Lag Decreasing**
```bash
kubectl exec -it nats-0 -n flyfleet -- nats consumer report TELEMETRY FLEET_CORE_CONSUMER | grep Pending
# Should be stable or decreasing
```

4. **Database Freshness**
```bash
# Check recent data
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT COUNT(*) FROM telemetry_events WHERE created_at > NOW() - INTERVAL '1 minute';"
# Should be > 10,000 (or your expected rate)
```

5. **Dashboard Updates**
- Verify latest vehicle positions appear in real-time
- Check no gaps in tracking timeline

### Stabilization (30 min)

- Continue monitoring NATS storage (should not grow unchecked)
- Monitor DB insert rate (should be stable)
- Watch fleet-core CPU/memory (should be normal range)
- Confirm no error spikes in Prometheus alerts

---

## Post-Incident Analysis

### PIR Focus Areas

1. **Root Cause Categories:**
   - NATS configuration (storage size, retention policy)
   - Consumer processing speed (code performance, database bottlenecks)
   - Network issues (VPC, security groups, NAT Gateway)
   - Device-side issues (firmware, connectivity)
   - Database degradation (indexes, vacuum, locks)

2. **Data Loss Assessment:**
   - Compare expected vs. actual event count during incident
   - Check for sequence number gaps
   - Identify if NATS messages were dropped (Dropped count)
   - Assess re-ingestion feasibility from devices

3. **Preventive Actions:**
   - Increase NATS storage or add tiered retention
   - Optimize telemetry processor (batch inserts, connection pooling)
   - Add alert on consumer lag derivative (increasing rate)
   - Implement circuit breaker for DB connection failures
   - Add device-side buffering for offline scenarios

---

## Prevention & Monitoring

### Essential Alerts (Already in OBSERVABILITY_DESIGN)

```yaml
- alert: TelemetryIngestionLagHigh
  expr: telemetry_gaps_total > 1000
  for: 5m

- alert: TelemetryThroughputDrop
  expr: rate(telemetry_events_received_total[5m]) < 10000
  for: 2m

- alert: TelemetryLatencyP99High
  expr: histogram_quantile(0.99, sum(rate(telemetry_ingestion_latency_seconds_bucket[5m])) by (le)) > 5
  for: 5m

- alert: NATSStorageHigh
  expr: nats_jetstream_storage_usage_bytes / nats_jetstream_storage_limit_bytes > 0.85
  for: 10m
```

### Dashboard Panels (Grafana)

Create panel queries:

1. **Ingestion Rate** (timeseries): `rate(telemetry_events_received_total[1m])`
2. **Consumer Lag** (gauge): `nats_jetstream_consumer_pending_bytes{consumer="FLEET_CORE_CONSUMER"}`
3. **End-to-End Latency P99** (gauge): `histogram_quantile(0.99, ...)`
4. **NATS Storage %** (gauge): `nats_jetstream_storage_usage_bytes / nats_jetstream_storage_limit_bytes * 100`
5. **Processing Errors** (counter): `rate(telemetry_processing_errors_total[1m])`

---

## Related Runbooks

- [NATS JetStream Issues](nats-issues.md) - For NATS-specific problems
- [Database Performance](database-performance.md) - If DB is ingestion bottleneck
- [API Outage](api-outage.md) - If telemetry affects API availability

---

**Last Updated:** 2026-03-29
**Review Date:** 2026-06-29
**Owner:** Platform Engineering Team
