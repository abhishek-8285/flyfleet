# Runbook: NATS JetStream Issues

**Severity:** SEV1 (if messaging completely down) or SEV2 (performance degradation)
**Service:** NATS JetStream (message broker for telemetry and commands)
**SLO Violation:** Storage < 85%, In/Out balance ±1%, Consumer lag < 100MB
**Alert:** `NATSStorageHigh` or `NATSMessageBacklogHigh`

---

## Quick Reference

| What | How |
|------|-----|
| **Check NATS server health** | `kubectl exec -it nats-0 -n flyfleet -- nats server report` |
| **Check all streams** | `kubectl exec -it nats-0 -n flyfleet -- nats stream ls` |
| **Check consumer lag** | `kubectl exec -it nats-0 -n flyfleet -- nats consumer report <stream> <consumer>` |
| **Check connections** | `kubectl exec -it nats-0 -n flyfleet -- nats server report --json \| jq '.connections'` |
| **Restart NATS** | `kubectl rollout restart deployment/nats -n flyfleet` |
| **Purge stale messages** | `kubectl exec -it nats-0 -n flyfleet -- nats stream purge <stream> --filter "<subject>"` |

---

## Symptoms

- Telemetry ingestion lagging or stopped
- Commands not reaching devices
- NATS pod high CPU/memory
- Connection errors: `nats: no servers available`
- Storage > 85% (risk of full disk)
- Consumer pending messages growing rapidly
- `Dropped` count increasing in stream stats
- Prometheus alerts on NATS metrics

---

## Decision Tree

```
NATS Issue?
├─ Server not running?
│   ├─ Pod CrashLoopBackOff? → Check logs, disk space
│   ├─ Pod Pending? → Insufficient resources
│   └─ Network? → Service DNS, security groups
├─ Storage full (>90%)?
│   ├─ Yes → Purge old streams or expand storage (EMERGENCY)
│   └─ No → Consumer lag high?
│       ├─ Yes → Consumer slow? → Check fleet-core, scale up
│       │        or Stream full? → Check storage, purge
│       └─ No → Connection issues?
│           ├─ Clients can't connect? → Check security groups, firewall
│           └─ High connection count? → Check for leaks, max_connections
└─ Messages Dropped?
    ├─ Stream storage full? → Expand or purge
    ├─ Consumer slow? → Scale consumers
    └─ Disk I/O bottleneck? → Check node disk, consider faster storage
```

---

## Investigation Steps

### 1. NATS Server Health (2 min)

```bash
# Pod status
kubectl get pods -n flyfleet -l app=nats

# NATS server report (comprehensive)
kubectl exec -it nats-0 -n flyfleet -- nats server report

# Quick summary
kubectl exec -it nats-0 -n flyfleet -- nats server report --json | jq '.info'

# Connection count
kubectl exec -it nats-0 -n flyfleet -- nats server report --json | jq '.connections | length'

# Check for recent disconnects in logs
kubectl logs -f deployment/nats -n flyfleet --since=10m 2>&1 | grep -i "disconnect\|error"
```

**Check:**
- `server_id` present? Server is running
- `connections` count reasonable? (< 1000 typical)
- `uptime` reasonable? (days/weeks normal)
- Any error lines in logs?

---

### 2. JetStream Streams Health (3 min)

```bash
# List all streams
kubectl exec -it nats-0 -n flyfleet -- nats stream ls

# Detailed report for each critical stream
kubectl exec -it nats-0 -n flyfleet -- nats stream report TELEMETRY
kubectl exec -it nats-0 -n flyfleet -- nats stream report COMMANDS
kubectl exec -it nats-0 -n flyfleet -- nats stream report SOS  # if exists

# OR get all stream info as JSON
kubectl exec -it nats-0 -n flyfleet -- nats stream report --json | jq '.[] | {stream: .config.name, storage: .config.storage, messages: .state.messages, bytes: .state.bytes}'
```

**Critical metrics per stream:**
- `Messages`: Total count
- `Bytes`: Storage used
- `FirstSeq` / `LastSeq`: Sequence range (gap = lost messages)
- `State` of stream: `OK` vs `ERROR`

---

### 3. Consumer Lag & Pending (3 min)

```bash
# Check consumers for each stream
kubectl exec -it nats-0 -n flyfleet -- nats consumer report TELEMETRY FLEET_CORE_CONSUMER
kubectl exec -it nats-0 -n flyfleet -- nats consumer report COMMANDS GATEWAY_CONSUMER

# Get all consumers across all streams
kubectl exec -it nats-0 -n flyfleet -- nats consumer ls --all

# Focus on key metrics:
# - Pending: messages ready but not yet delivered
# - Delivered: messages delivered but not yet acked
# - AckFloor: last acknowledged message sequence
```

**Critical thresholds:**
- `Pending` > 100,000? Consumer falling behind significantly
- `Delivered` high? Consumer slow to ack (processing bottleneck)
- `AckFloor` not advancing? Consumer not processing at all

---

### 4. Prometheus Metrics (2 min)

```bash
# Stream storage usage
curl -s 'http://prometheus:9090/api/v1/query?query=nats_jetstream_storage_usage_bytes' | jq

# Messages in/out balance (should be roughly equal)
curl -s 'http://prometheus:9090/api/v1/query?query=rate(nats_jetstream_messages_received_total[1m])' | jq
curl -s 'http://prometheus:9090/api/v1/query?query=rate(nats_jetstream_messages_sent_total[1m])' | jq

# Consumer pending by consumer
curl -s 'http://prometheus:9090/api/v1/query?query=nats_jetstream_consumer_pending_bytes' | jq

# Connection count
curl -s 'http://prometheus:9090/api/v1/query?query=nats_connections' | jq

# Dropped messages
curl -s 'http://prometheus:9090/api/v1/query?query=nats_jetstream_messages_dropped_total' | jq
```

---

### 5. Check Disk & Node Resources (3 min)

```bash
# NATS pod node
kubectl get pod nats-0 -n flyfleet -o jsonpath='{.spec.nodeName}'

# Node disk usage
kubectl describe node <node-name> | grep -A10 "Allocated resources"

# NATS PVC (if using persistent storage)
kubectl get pvc -n flyfleet
kubectl describe pvc <nats-pvc-name> -n flyfleet

# Disk space on node (from NATS pod)
kubectl exec -it nats-0 -n flyfleet -- df -h /data/nats/jetstream

# Check for disk pressure on node
kubectl describe node <node-name> | grep -i "disk"
```

**If disk full:**
- Immediate emergency (messages will be dropped)
- Need to purge or expand within minutes

---

### 6. Check Client Connectivity (2 min)

```bash
# Who is connected?
kubectl exec -it nats-0 -n flyfleet -- nats server report --json | jq '.connections[] | {cid, account, ip, port, proto, uptime}'

# Connection stats
kubectl exec -it nats-0 -n flyfleet -- nats server report --json | jq '.connections | length'

# Check for zombie connections (in warnings)
kubectl logs -f deployment/nats -n flyfleet --since=10m 2>&1 | grep -i "zombie\|tls\|auth"
```

---

## Mitigation Strategies

### Strategy 1: Restart NATS Server

**When:** Server not responding, high errors

```bash
kubectl rollout restart deployment/nats -n flyfleet
kubectl rollout status deployment/nats -n flyfleet --timeout=5m

# Check recovery
kubectl exec -it nats-0 -n flyfleet -- nats server report

# Monitor connections re-establishing
kubectl logs -f deployment/nats -n flyfleet
```

**Impact:** Brief interruption (seconds) while server restarts. Clients auto-reconnect.

---

### Strategy 2: Purge Stale Messages (Emergency Storage Full)

**When:** Storage > 90%, risk of dropping new messages

⚠️ **WARNING:** This permanently deletes messages. Use only if storage crisis or corrupt data.

```bash
# Check storage first
kubectl exec -it nats-0 -n flyfleet -- nats stream report TELEMETRY

# Purge entire stream (loses ALL messages in that stream)
kubectl exec -it nats-0 -n flyfleet -- nats stream purge TELEMETRY --purge

# OR purge by age (e.g., messages older than 24h)
kubectl exec -it nats-0 -n flyfleet -- nats stream purge TELEMETRY --older 24h --yes

# OR purge by subject (specific problematic topic)
kubectl exec -it nats-0 -n flyfleet -- nats stream purge TELEMETRY --filter "telemetry.test.>" --yes

# After purge, verify storage freed
kubectl exec -it nats-0 -n flyfleet -- nats stream report TELEMETRY
```

**After purge:**
- Consumers may need reset to `deliver all` to reprocess
- Notify downstream teams (telemetry team)
- Monitor for impact

---

### Strategy 3: Scale NATS Cluster

**When:** Connection count near limit or high CPU/memory

```bash
# Check current cluster size
kubectl get deployment -n flyfleet | grep nats

# Scale to 3 nodes (requires proper cluster configuration)
kubectl scale deployment nats --replicas=3 -n flyfleet

# Monitor cluster formation
kubectl logs -f deployment/nats -n flyfleet | grep -i "cluster\|peers"

# Check cluster info from any node
kubectl exec -it nats-0 -n flyfleet -- nats server report --json | jq '.cluster'
```

**Prerequisites:**
- NATS configured for clustering (requires StatefulSet with headless service)
- JetStream storage in shared storage (NFS, EFS) or use RAFT
- Evaluate if clustering necessary vs. vertical scaling

---

### Strategy 4: Increase Storage

**When:** Storage consistently > 85%, need more capacity

```bash
# Check if using PVC
kubectl get pvc -n flyfleet | grep nats
kubectl describe pvc <nats-pvc> -n flyfleet

# Expand PVC (if storage class allows expansion)
kubectl edit pvc <nats-pvc> -n flyfleet
# Modify: spec.resources.requests.storage: "100Gi" → "200Gi"

# If using hostPath, expand node disk or add new volume
# Or migrate to larger storage class

# Verify expansion inside pod
kubectl exec -it nats-0 -n flyfleet -- df -h /data/nats/jetstream
```

---

### Strategy 5: Reset Consumer (Re-sync Lag)

**When:** Consumer pending very high, consumer stuck or behind

```bash
# Check consumer state
kubectl exec -it nats-0 -n flyfleet -- nats consumer report TELEMETRY FLEET_CORE_CONSUMER

# Reset consumer to start from beginning (replay all messages)
kubectl exec -it nats-0 -n flyfleet -- nats consumer reset TELEMETRY FLEET_CORE_CONSUMER

# OR set to specific sequence (skip old messages)
kubectl exec -it nats-0 -n flyfleet -- nats consumer move --count 10000 TELEMETRY FLEET_CORE_CONSUMER

# Restart consumer service to reconnect with fresh state
kubectl rollout restart deployment/fleet-core -n flyfleet
```

**Warning:** Resetting consumer causes message replay. Ensure consumer is idempotent (can handle duplicates).

---

### Strategy 6: Create Dead Letter Queue (DLQ)

**When:** Messages repeatedly failing to process, blocking stream

```bash
# Create DLQ stream
kubectl exec -it nats-0 -n flyfleet -- nats stream add DLQ --subjects "dlq.>" --retention limits --max-msgs=100000 --max-bytes=1GB

# Configure consumer with DLQ (requires consumer re-creation)
kubectl exec -it nats-0 -n flyfleet -- nats consumer delete TELEMETRY FLEET_CORE_CONSUMER
kubectl exec -it nats-0 -n flyfleet -- nats consumer add TELEMETRY FLEET_CORE_CONSUMER --deliver all --ack exact --max-delivery=3 --dlq DLQ

# Restart consumer service
kubectl rollout restart deployment/fleet-core -n flyfleet
```

**After DLQ:**
- Monitor DLQ stream growth
- Set up alert on DLQ size
- Build admin tool to inspect and replay DLQ messages

---

## Recovery & Validation

### Validation Steps

1. **NATS Server Health**
```bash
kubectl exec -it nats-0 -n flyfleet -- nats server report --json | jq '.info.state'
# Should be "SERVER_ACTIVE"
```

2. **Stream Health**
```bash
kubectl exec -it nats-0 -n flyfleet -- nats stream ls
# All streams should be listed with no errors
kubectl exec -it nats-0 -n flyfleet -- nats stream report TELEMETRY | grep "State"
# Should be "OK"
```

3. **Storage Normal**
```bash
kubectl exec -it nats-0 -n flyfleet -- nats stream report TELEMETRY | grep "Storage"
# Should be < 85%
```

4. **Consumer Lag Decreasing**
```bash
kubectl exec -it nats-0 -n flyfleet -- nats consumer report TELEMETRY FLEET_CORE_CONSUMER | grep Pending
# Should be stable or decreasing over time
```

5. **Message Flow Resumed**
```bash
# Check in/out balance from Prometheus
curl -s 'http://prometheus:9090/api/v1/query?query=increase(nats_jetstream_messages_received_total[5m])' | jq
curl -s 'http://prometheus:9090/api/v1/query?query=increase(nats_jetstream_messages_sent_total[5m])' | jq
# Should be roughly equal (allowing for lag)
```

---

## Post-Incident Analysis

### PIR Focus Areas

1. **Root Cause Categories:**
   - Storage too small for ingestion rate
   - Consumer processing slower than ingestion (need to scale consumers)
   - Consumer crash/disconnect not detected (need better monitoring)
   - Message backlog due to downstream bottleneck (DB slow)
   - NATS server resource exhaustion (CPU/memory/disk I/O)
   - Network issues causing disconnects
   - Configuration error (misconfigured retention, storage)

2. **Preventive Actions:**
   - Increase storage capacity with headroom (target < 70% utilization)
   - Add consumer autoscaling based on lag metric
   - Implement circuit breaker for downstream failures (prevent uncontrolled backlog)
   - Add alerts on consumer lag derivative (rate of increase)
   - Set up NATS monitoring dashboards (storage, pending, connections)
   - Implement DLQ for poison messages
   - Add alert on stream `FirstSeq` gap > 1 (detect dropped messages)
   - Consider adding monitoring for NATS process CPU and memory

3. **Operational Improvements:**
   - Document NATS admin operations (stream create/delete, consumer reset)
   - Build admin scripts for common operations
   - Set up regular storage usage reviews
   - Test disaster recovery (stream backup/restore)

---

## Prevention & Monitoring

### Essential Alerts

```yaml
- alert: NATSStorageHigh
  expr: nats_jetstream_storage_usage_bytes / nats_jetstream_storage_limit_bytes > 0.85
  for: 10m

- alert: NATSStorageCritical
  expr: nats_jetstream_storage_usage_bytes / nats_jetstream_storage_limit_bytes > 0.95
  for: 2m

- alert: NATSMessageBacklogHigh
  expr: nats_jetstream_consumer_pending_bytes > 100000000  # 100MB
  for: 5m

- alert: NATSDroppedMessages
  expr: rate(nats_jetstream_messages_dropped_total[5m]) > 0
  for: 2m

- alert: NATSConnectionsHigh
  expr: nats_connections > 800
  for: 5m

- alert: NATSServerDown
  expr: up{service="nats"} == 0
  for: 1m
```

### Dashboard Panels

1. Storage usage % by stream (stacked bar)
2. Consumer pending (top 5 consumers)
3. Messages in/out balance (timeseries)
4. Connection count (gauge + time series)
5. Dropped messages counter (delta)
6. Stream message count by subject

---

## Common Troubleshooting Commands

```bash
# Test NATS connectivity from another pod
kubectl exec -it deployment/fleet-core -n flyfleet -- \
  nats --server nats://nats.flyfleet.svc.cluster.local:4222 ping

# Publish test message
kubectl exec -it deployment/fleet-core -n flyfleet -- \
  nats --server nats://nats.flyfleet.svc.cluster.local:4222 pub test.subject "hello"

# Subscribe to test (run in another terminal)
kubectl exec -it deployment/fleet-core -n flyfleet -- \
  nats --server nats://nats.flyfleet.svc.cluster.local:4222 sub test.subject

# Get stream info in JSON
kubectl exec -it nats-0 -n flyfleet -- nats stream report TELEMETRY --json | jq '.'
```

---

## Related Runbooks

- [Telemetry Disruption](telemetry-disruption.md) - Telemetry pipeline issues
- [Command Delivery Failure](command-delivery-failure.md) - Command pipeline issues
- [API Outage](api-outage.md) - If NATS issues cause API failures

---

**Last Updated:** 2026-03-29
**Review Date:** 2026-06-29
**Owner:** Platform Engineering Team
