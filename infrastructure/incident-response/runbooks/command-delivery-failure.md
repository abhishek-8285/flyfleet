# Runbook: Command Delivery Failure

**Severity:** SEV2 or SEV1 if critical emergency commands failing
**Service:** Command Service (gateway → device)
**SLO Violation:** Command Success Rate < 99.9%, Ack Latency P95 > 2s
**Alert:** `FlyFleetCommandReliabilityBurnCritical` or `CommandSuccessRateLow`

---

## Quick Reference

| What | How |
|------|-----|
| **Check command success rate** | `curl -s 'http://prometheus:9090/api/v1/query?query=rate(device_command_events_total{result="success",service="gateway"}[1h]) / rate(device_command_events_total{service="gateway"}[1h])'` |
| **Check command queue depth** | `curl -s 'http://prometheus:9090/api/v1/query?query=command_queue_length'` |
| **Check device connectivity** | `kubectl logs -f deployment/gateway -n flyfleet --since=10m \| grep "device.*connect\|disconnect"` |
| **Check NATS backlog** | `kubectl exec -it nats-0 -n flyfleet -- nats consumer report COMMANDS` |
| **Retry failed commands** | Use admin API: `POST /admin/commands/retry` with filter |
| **Clear command queue** | Emergency: `nats consumer purge COMMANDS GATEWAY_CONSUMER` |

---

## Symptoms

- Commands sent to devices not executing (no ACK received)
- Command latency high (> 2s for 95th percentile)
- Queue depth growing (commands piling up)
- Devices showing "offline" or "unreachable"
- Command status stuck in "pending" or "sent"
- Emergency SOS commands failing to reach vehicles

---

## Decision Tree

```
Commands Failing?
├─ Yes → All devices or specific?
│   ├─ All → Gateway service down?
│   │   ├─ Yes → Restart gateway deployment
│   │   └─ No → NATS issues?
│   │       ├─ Yes → NATS runbook
│   │       └─ No → Device connectivity problem?
│   │           ├─ Yes → Check firewall/VPC, device online status
│   │           └─ No → Check command processor logs
│   └─ Specific devices?
│       ├─ Device offline → Check device connectivity/GPS signal
│       ├─ Device type specific → Check firmware compatibility
│       └─ Region-specific → Check network/carrier issues
└─ No → Latency high?
    ├─ Processing slow → Scale gateway service
    ├─ NATS lag → Consumer lag, check NATS
    └─ ACK delayed → Check device response time
```

---

## Investigation Steps

### 1. Verify Command Metrics (2 min)

```bash
# Current success rate (last 1h)
curl -s 'http://prometheus:9090/api/v1/query?query=100 * (sum(rate(device_command_events_total{result="success",service="gateway"}[1h])) / sum(rate(device_command_events_total{service="gateway"}[1h])))' | jq

# Command volume rate
curl -s 'http://prometheus:9090/api/v1/query?query=rate(device_command_events_total[1m])' | jq

# Queue depth (pending commands)
curl -s 'http://prometheus:9090/api/v1/query?query=command_queue_length' | jq

# Ack latency P95
curl -s 'http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95, sum(rate(command_ack_latency_seconds_bucket[1h])) by (le))' | jq

# Command failures by reason
curl -s 'http://prometheus:9090/api/v1/query?query=sum(rate(device_command_events_total{result!="success"}[1h])) by (result)' | jq
```

**Baseline:** Success rate > 99.9%, Ack latency P95 < 2s, queue < 10,000

---

### 2. Check Gateway Service Health (3 min)

```bash
# Gateway pod status
kubectl get pods -n flyfleet -l app=gateway

# Resource utilization
kubectl top pods -n flyfleet -l app=gateway

# Recent events
kubectl get events -n flyfleet --field-selector involvedObject.kind=Pod,involvedObject.label.app=gateway --sort-by='.lastTimestamp' | tail -20

# Gateway logs (look for command processing errors)
kubectl logs -f deployment/gateway -n flyfleet --since=10m 2>&1 | grep -i "command\|ack\|timeout\|error"

# Check NATS connection
kubectl logs -f deployment/gateway -n flyfleet --since=10m 2>&1 | grep -i "nats"
```

**Check for:**
- CrashLoopBackOff or high restart count
- CPU > 90% or memory pressure
- NATS connection errors: `nats: no servers available`
- Command timeout errors: `context deadline exceeded`

---

### 3. NATS JetStream Command Stream Health (3 min)

```bash
# Command stream info
kubectl exec -it nats-0 -n flyfleet -- nats stream report COMMANDS

# Command consumer (gateway consumer)
kubectl exec -it nats-0 -n flyfleet -- nats consumer report COMMANDS GATEWAY_CONSUMER

# Check pending messages (queue depth)
kubectl exec -it nats-0 -n flyfleet -- nats consumer report COMMANDS GATEWAY_CONSUMER | grep "Pending"

# Check delivered/dropped counts
kubectl exec -it nats-0 -n flyfleet -- nats consumer report COMMANDS GATEWAY_CONSUMER | grep -E "Delivered|Dropped"

# Storage usage for COMMANDS stream
kubectl exec -it nats-0 -n flyfleet -- nats stream report COMMANDS | grep -E "Storage|Memory|Disk"
```

**Critical checks:**
- `Pending` very high (> 50,000)? Consumer can't keep up
- `Dropped` count increasing? Messages being lost
- `Storage` > 90%? Risk of full disk
- Consumer `State` = `Stalled`? Consumer not processing

---

### 4. Device Connectivity Status (3 min)

```bash
# Check device online status in database
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT COUNT(*) FROM devices WHERE is_online = true;"

# Check last heartbeat by device type
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT device_type, COUNT(*) as total, COUNT(*) FILTER (WHERE last_heartbeat_at > NOW() - INTERVAL '5 minutes') as online FROM devices GROUP BY device_type;"

# Check recent command acknowledgments
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT device_id, command_id, status, acknowledged_at FROM device_commands WHERE created_at > NOW() - INTERVAL '1 hour' AND status != 'acknowledged' ORDER BY created_at DESC LIMIT 10;"

# Check if specific devices failing consistently
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT device_id, COUNT(*) as total, COUNT(*) FILTER (WHERE status = 'timeout') as timeouts FROM device_commands WHERE created_at > NOW() - INTERVAL '30 minutes' GROUP BY device_id HAVING timeouts > 5 ORDER BY timeouts DESC LIMIT 10;"
```

**Interpretation:**
- Many devices offline? Network/cellular issue in region
- Specific devices always failing? Device firmware/hardware issue
- Recent commands timing out? Gateway not delivering to NATS

---

### 5. Check Network/Firewall (2 min)

```bash
# Check NATS client connections (devices connecting via gateway)
kubectl exec -it nats-0 -n flyfleet -- nats server report --json | jq '.connections[] | {cid, ip, port, account}'

# Check security group rules (AWS)
# Ensure devices can reach NATS port (4222) and monitoring port (8222)

# Check VPC flow logs (if enabled) for dropped packets
aws logs filter-log-events \
  --log-group-name /aws/vpc/flow-logs \
  --filter-pattern "REJECT" \
  --start-time $(date -d '15 minutes ago' +%s000) \
  --max-results 20

# Ping test from gateway pod to NATS
kubectl exec -it deployment/gateway -n flyfleet -- \
  nc -zv nats.flyfleet.svc.cluster.local 4222
```

---

## Mitigation Strategies

### Strategy 1: Restart Gateway Service

**When:** Gateway in crashed state or not processing commands

```bash
kubectl rollout restart deployment/gateway -n flyfleet
kubectl rollout status deployment/gateway -n flyfleet --timeout=5m

# Watch logs for command processing restart
kubectl logs -f deployment/gateway -n flyfleet | grep -i "command consumer\|starting"
```

---

### Strategy 2: Reconnect NATS Consumer

**When:** NATS consumer stalled or disconnected

```bash
# Check consumer state
kubectl exec -it nats-0 -n flyfleet -- nats consumer report COMMANDS GATEWAY_CONSUMER

# If stalled, purge and recreate consumer
kubectl exec -it nats-0 -n flyfleet -- nats consumer purge COMMANDS GATEWAY_CONSUMER --purge

# OR delete and recreate consumer
kubectl exec -it nats-0 -n flyfleet -- nats consumer delete COMMANDS GATEWAY_CONSUMER
kubectl exec -it nats-0 -n flyfleet -- nats consumer add COMMANDS GATEWAY_CONSUMER --deliver all --ack exact --max-delivery=3 --filter "commands.>" --durable

# Restart gateway to reconnect
kubectl rollout restart deployment/gateway -n flyfleet
```

---

### Strategy 3: Scale Gateway Horizontally

**When:** Queue growing and gateway CPU/memory high

```bash
# Check current replicas
kubectl get deployment gateway -n flyfleet

# Scale to 3 replicas (competing consumers)
kubectl scale deployment gateway --replicas=3 -n flyfleet

# Monitor queue distribution
watch -n 10 'kubectl exec -it nats-0 -n flyfleet -- nats consumer report COMMANDS GATEWAY_CONSUMER | grep Pending'
```

**Note:** With multiple consumers, NATS distributes messages. Ensure idempotency in command processing (devices should handle duplicate commands).

---

### Strategy 4: Retry Failed Commands

**When:** Queue contains failed commands that should be retried

**Option A: Admin API retry endpoint** (if implemented)
```bash
# Retry commands older than 1 hour with status=timeout
curl -X POST https://api.flyfleet.com/admin/commands/retry \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": "timeout", "older_than_minutes": 60}'
```

**Option B: Database retry script**
```bash
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    UPDATE device_commands
    SET status = 'pending', attempts = 0, updated_at = NOW()
    WHERE status IN ('timeout', 'failed')
      AND created_at > NOW() - INTERVAL '2 hours'
      AND attempts < 3;
  "
```

**After retry:** Monitor queue for reprocessing

---

### Strategy 5: Emergency Purge Command Queue

**When:** Queue flooded with stale commands that can't be delivered

⚠️ **WARNING:** This deletes all pending commands. Only use for emergency (e.g., bad batch of commands that will never succeed).

```bash
# Purge pending commands from NATS consumer
kubectl exec -it nats-0 -n flyfleet -- nats consumer purge COMMANDS GATEWAY_CONSUMER --purge

# Also update database to match
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    UPDATE device_commands
    SET status = 'purged', updated_at = NOW()
    WHERE status IN ('pending', 'sent')
      AND created_at < NOW() - INTERVAL '1 hour';
  "

# Restart gateway to clear internal queue
kubectl rollout restart deployment/gateway -n flyfleet
```

---

### Strategy 6: Increase Command Processing Rate

**When:** Queue growing but resources not saturated (can handle more load)

```bash
# Edit gateway deployment to increase NATS batch size
kubectl edit deployment/gateway -n flyfleet

# Add/modify env:
# - name: NATS_BATCH_SIZE
#   value: "100"  # default 50
# - name: NATS_ACK_WAIT
#   value: "30s"
# - name: COMMAND_PROCESSING_WORKERS
#   value: "10"  # default 5, increase parallelism

kubectl rollout restart deployment/gateway -n flyfleet
```

---

## Recovery & Validation

### Post-Mitigation Checks

1. **Success Rate Recovery**
```bash
# Should be > 99.9%
curl -s 'http://prometheus:9090/api/v1/query?query=100 * (sum(rate(device_command_events_total{result="success",service="gateway"}[5m])) / sum(rate(device_command_events_total{service="gateway"}[5m])))' | jq
```

2. **Queue Depth Decreasing**
```bash
curl -s 'http://prometheus:9090/api/v1/query?query=command_queue_length' | jq
# Should be declining or < 10,000
```

3. **Ack Latency Normal**
```bash
curl -s 'http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95, sum(rate(command_ack_latency_seconds_bucket[5m])) by (le))' | jq
# Should be < 2s
```

4. **Device ACK Received**
```bash
# Check recent acknowledged commands
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT COUNT(*) FROM device_commands WHERE status = 'acknowledged' AND updated_at > NOW() - INTERVAL '10 minutes';"
```

5. **Test Live Command**
```bash
# Send test command to known online device
DEVICE_ID="test-device-123"
curl -X POST https://api.flyfleet.com/api/v1/commands \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"device_id\":\"$DEVICE_ID\",\"command\":\"ping\"}"

# Check it gets acked within 2s
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT status, acknowledged_at FROM device_commands WHERE device_id = '$DEVICE_ID' ORDER BY created_at DESC LIMIT 1;"
```

---

## Emergency SOS Procedure

**Special Priority:** Emergency SOS commands (vehicle distress signals) MUST bypass normal queues.

### SOS Command Flow

1. **SOS received** via API or device button press
2. **Immediate routing** to dedicated `SOS` NATS subject with highest priority
3. **Separate consumer** with unlimited resources and no rate limiting
4. **Immediate ACK** to device (different from command ack)
5. **Simultaneous notifications:**
   - Emergency contacts (via SMS/email)
   - Fleet manager (dashboard alert)
   - Dispatcher (push notification)

### SOS Incident Response

If SOS commands failing:

```bash
# Check SOS stream/consumer separately
kubectl exec -it nats-0 -n flyfleet -- nats stream report SOS
kubectl exec -it nats-0 -n flyfleet -- nats consumer report SOS SOS_CONSUMER

# Restart SOS consumer only (not regular commands)
kubectl rollout restart deployment/sos-processor -n flyfleet  # if separate service
# OR
kubectl rollout restart deployment/gateway -n flyfleet  # if same service

# Send test SOS (use test device)
curl -X POST https://api.flyfleet.com/api/v1/commands/sos \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"device_id":"test-device","location":{"lat":40.7128,"lng":-74.0060}}'
```

**Escalation:** SOS command failure is **SEV1** - immediate response required, notify VP Engineering and Legal.

---

## Post-Incident Actions

### PIR Focus Areas

1. **Root Cause Categories:**
   - NATS consumer lag (processing speed vs. ingestion rate)
   - Gateway service bugs (command serialization, ack handling)
   - Database write bottlenecks (insert performance, locks)
   - Device reliability (firmware bugs, network dropouts)
   - Network/firewall blocking ports
   - Queue overflow (burst traffic)

2. **Data Loss Assessment:**
   - Count commands in `pending`/`timeout` state during incident
   - Determine which commands were never delivered
   - Assess safety-critical commands (SOS, panic) impact
   - Need manual re-issuance?

3. **Preventive Actions:**
   - Add dead-letter queue for failed commands
   - Implement exponential backoff with jitter
   - Add per-device rate limiting to prevent floods
   - Increase NATS storage for burst retention
   - Circuit breaker for downstream failures
   - End-to-end integration tests for command flow

---

## Prevention & Monitoring

### Essential Alerts

```yaml
- alert: CommandSuccessRateLow
  expr: 100 * (sum(rate(device_command_events_total{result="success",service="gateway"}[1h])) / sum(rate(device_command_events_total{service="gateway"}[1h]))) < 99.9
  for: 10m

- alert: CommandQueueDepthHigh
  expr: command_queue_length > 50000
  for: 5m

- alert: CommandAckLatencyHigh
  expr: histogram_quantile(0.95, sum(rate(command_ack_latency_seconds_bucket[5m])) by (le)) > 2
  for: 10m

- alert: SOSCommandFailure
  expr: rate(device_command_events_total{command_type="sos",result!="success",service="gateway"}[5m]) > 0
  for: 1m
```

### Dashboard Panels

1. Command Success Rate (gauge)
2. Queue Depth (timeseries)
3. Ack Latency P95 (timeseries)
4. Commands Sent by Type (bar chart)
5. Pending by Device Type (stacked bar)
6. SOS Command Status (critical alerts)

---

## Related Runbooks

- [NATS JetStream Issues](nats-issues.md) - NATS infrastructure problems
- [API Outage](api-outage.md) - If gateway service completely down
- [Data Inconsistency](data-inconsistency.md) - If commands show in DB but not on device

---

**Last Updated:** 2026-03-29
**Review Date:** 2026-06-29
**Owner:** Platform Engineering Team
