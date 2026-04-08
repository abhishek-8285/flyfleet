# Runbook: API Outage (fleet-core)

**Severity:** SEV1 or SEV2 depending on impact
**Service:** fleet-core (main backend API)
**SLO Violation:** Availability >99.95%, Error Rate <0.1%, Latency P95 <300ms
**Alert:** `FlyFleetApiAvailabilityBurnCritical` or `FlyFleetApiErrorRateCritical`

---

## Quick Reference

| What | How |
|------|-----|
| **Check API health** | `curl http://localhost:8080/health` or `curl http://<prod-url>/health` |
| **View pod status** | `kubectl get pods -n flyfleet -l app=fleet-core` |
| **Check logs** | `kubectl logs -f deployment/fleet-core -n flyfleet --tail=100` |
| **Check metrics** | Grafana → "FlyFleet API Dashboard" |
| **Rollback deployment** | `kubectl rollout undo deployment/fleet-core -n flyfleet` |
| **Restart pods** | `kubectl rollout restart deployment/fleet-core -n flyfleet` |
| **Scale up** | `kubectl scale deployment/fleet-core --replicas=5 -n flyfleet` |

---

## Symptoms

- API returning 5xx errors (50x, 502, 503, 504)
- Health endpoint failing (`/health` returns non-200)
- High error rate (> 5% of requests)
- Latency spike (P95 > 2s)
- Complete unavailability (connection refused/timeout)

---

## Decision Tree

```
API Down?
├─ Yes → Health check failing?
│   ├─ Yes → Pods running?
│   │   ├─ No → Deploy/rollout issue → Run: kubectl rollout status
│   │   └─ Yes → Pods healthy?
│   │       ├─ No → Check logs for panic/crash → May need restart
│   │       └─ Yes → LoadBalancer/Ingress issue → Check ingress controller
│   └─ No → Application responding?
│       ├─ No → Database connection? NATS? → Check DB/NATS connectivity
│       └─ Yes → Nginx/Ingress misconfigured → Check ingress logs
└─ No → Error rate high or latency spike?
    ├─ Error rate high → Check:
    │   ├─ Recent deployment → Rollback
    │   ├─ Database errors → Database runbook
    │   ├─ NATS timeout → NATS runbook
    │   └─ Authentication failures → Auth runbook
    └─ Latency high → Check:
        ├─ Database slow queries → Database runbook
        ├─ High CPU/Memory → HPA or increase resources
        ├─ External API slow → Circuit breaker?
        └─ Network issues → VPC/security groups
```

---

## Investigation Steps (In Order)

### 1. Verify Impact & Scope (2 min)

```bash
# Check if API is down globally
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health

# Check recent error rate from Prometheus
curl -s 'http://prometheus:9090/api/v1/query?query=rate(http_requests_total{status=~"5..",service="fleet-core"}[5m])' | jq

# Check which endpoints are failing
curl -s 'http://prometheus:9090/api/v1/query?query=rate(http_requests_total{status=~"5..",service="fleet-core"}[5m]) by (handler)' | jq
```

**Questions:**
- Is this affecting all users or specific tenants?
- Which endpoints are failing? (GET/POST, specific routes)
- When did it start? Correlate with deployments

### 2. Check Kubernetes Status (2 min)

```bash
# Pod status
kubectl get pods -n flyfleet -l app=fleet-core -o wide

# Recent events
kubectl get events -n flyfleet --field-selector involvedObject.kind=Pod,involvedObject.name=fleet-core --sort-by='.lastTimestamp'

# Deployment status
kubectl rollout status deployment/fleet-core -n flyfleet

# Resource utilization
kubectl top pods -n flyfleet -l app=fleet-core
```

**Check for:**
- CrashLoopBackOff
- ImagePullBackOff
- Pending (insufficient resources)
- High CPU/Memory (> 90%)

### 3. Application Logs (5 min)

```bash
# Recent logs (last 5 min)
kubectl logs -f deployment/fleet-core -n flyfleet --since=5m

# Panic/crash stack traces
kubectl logs -f deployment/fleet-core -n flyfleet --previous 2>&1 | grep -A20 "panic:"

#Database connection errors
kubectl logs -f deployment/fleet-core -n flyfleet --since=5m 2>&1 | grep -i "database\|postgres\|connection"

# NATS connection errors
kubectl logs -f deployment/fleet-core -n flyfleet --since=5m 2>&1 | grep -i "nats\|jetstream"
```

**Common patterns:**
- `context deadline exceeded` → Timeout issues (DB/NATS)
- `pq: password authentication failed` → DB credentials problem
- `nats: no servers available for service` → NATS connectivity
- `panic:` → Application crash (need restart/rollback)

### 4. Database Health (3 min)

```bash
# Check PostgreSQL pod/status
kubectl get pods -n flyfleet -l app=postgres

# Database connectivity from fleet-core pod
kubectl exec -it deployment/fleet-core -n flyfleet -- \
  psql $DATABASE_URL -c "SELECT 1" 2>&1

# Connection pool stats (if metrics available)
curl -s 'http://prometheus:9090/api/v1/query?query=pgpool_active_connections' | jq

# Long-running queries
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT pid, now() - pg_stat_activity.query_start AS duration, query FROM pg_stat_activity WHERE state = 'active' AND (now() - pg_stat_activity.query_start) > interval '5 seconds' ORDER BY duration DESC LIMIT 10;"
```

**Check:**
- Connection pool exhaustion (all connections used)
- Long-running queries blocking
- Replication lag (if using read replicas)
- Disk space: `kubectl exec postgres-0 -- df -h /var/lib/postgresql/data`

### 5. NATS JetStream Health (2 min)

```bash
# Check NATS server status
kubectl exec -it nats-0 -n flyfleet -- nats server report

# Check JetStream streams
kubectl exec -it nats-0 -n flyfleet -- nats stream report

# Check consumer lag
kubectl exec -it nats-0 -n flyfleet -- nats consumer report TELEMETRY --server nats://nats.flyfleet.svc.cluster.local:4222

# Storage usage
kubectl exec -it nats-0 -n flyfleet -- nats js info
```

**Check:**
- Stream storage > 85%? → Need to expand or prune
- Consumer pending messages growing? → Consumer lag
- Server not running?

### 6. Network & Ingress (2 min)

```bash
# Check ingress controller
kubectl get pods -n ingress-nginx
kubectl logs -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx --tail=50

# Check service endpoints
kubectl get endpoints -n flyfleet fleet-core

# Check security groups (AWS)
aws ec2 describe-security-groups --group-ids $(aws ec2 describe-instances --filters Name=tag:Name,Values=fleet-core-* --query 'Reservations[].Instances[].SecurityGroups[].GroupId' --output text)

# Check load balancer health
kubectl describe service fleet-core -n flyfleet | grep -A5 "LoadBalancer"
```

---

## Mitigation Strategies

### Strategy 1: Restart Application (Quick Reset)

**When to use:** Application in CrashLoopBackOff or stuck state

```bash
kubectl rollout restart deployment/fleet-core -n flyfleet
kubectl rollout status deployment/fleet-core -n flyfleet --timeout=5m
```

**Monitor:**
- Pods coming up: `kubectl get pods -n flyfleet -w`
- Health endpoint returning 200
- Logs for errors: `kubectl logs -f deployment/fleet-core -n flyfleet`

**If still failing:** Move to rollback strategy

---

### Strategy 2: Rollback to Previous Version

**When to use:** Recent deployment caused outage, pods running but unhealthy

```bash
# Check revision history
kubectl rollout history deployment/fleet-core -n flyfleet

# Rollback to previous revision
kubectl rollout undo deployment/fleet-core -n flyfleet

# Verify rollback
kubectl rollout status deployment/fleet-core -n flyfleet --timeout=5m

# Confirm new version
kubectl get pods -n flyfleet -l app=fleet-core -o jsonpath='{.items[0].spec.containers[0].image}'
```

**If rollback fails** (images deleted): Need to rebuild and push previous version

---

### Strategy 3: Scale Out

**When to use:** High load causing timeouts, but application logic correct

```bash
# Check current replicas
kubectl get deployment fleet-core -n flyfleet

# Scale up (from default 2 to 5)
kubectl scale deployment fleet-core --replicas=5 -n flyfleet

# Monitor scaling
kubectl get pods -n flyfleet -l app=fleet-core -w
kubectl top pods -n flyfleet -l app=fleet-core

# After stabilization, scale back down gracefully
# (Leave at 5 for 30 min, then reduce to 3)
```

**Note:** Ensure HorizontalPodAutoscaler is configured: `kubectl get hpa -n flyfleet`

---

### Strategy 4: Database Connection Pool Increase

**When to use:** Database connection errors, "too many connections"

```bash
# Check current pool settings in deployment
kubectl describe deployment fleet-core -n flyfleet | grep -A5 "env"

# Increase pool size temporarily (edit deployment)
kubectl edit deployment fleet-core -n flyfleet
# Add/modify env:
# - name: DATABASE_MAX_OPEN_CONNS
#   value: "50"  # default 25
# - name: DATABASE_MAX_IDLE_CONNS
#   value: "25"  # default 25

# Or if using ConfigMap/Helm values:
helm upgrade flyfleet infrastructure/helm/flyfleet \
  --namespace flyfleet \
  --set env.DATABASE_MAX_OPEN_CONNS=50

# Restart pods to pick up new config
kubectl rollout restart deployment/fleet-core -n flyfleet
```

---

### Strategy 5: Clear NATS JetStream Stream (Data Corruption)

**When to use:** Telemetry data corrupted, causing processing failures

⚠️ **WARNING:** This deletes all messages in the stream. Only use if data is corrupt and re-ingestion possible.

```bash
# Check stream info first
kubectl exec -it nats-0 -n flyfleet -- nats stream report TELEMETRY

# Delete and recreate stream
kubectl exec -it nats-0 -n flyfleet -- nats stream delete TELEMETRY
kubectl exec -it nats-0 -n flyfleet -- nats stream add TELEMETRY --subjects "telemetry.>" --storage file --retention limits --max-msgs=10000000 --max-bytes=10GB

# Recreate consumers
kubectl exec -it nats-0 -n flyfleet -- nats consumer add TELEMETRY FLEET_CORE_CONSUMER --deliver all --ack exact --max-delivery=3
```

**After:** Coordinate with telemetry team to resume ingestion

---

## Recovery & Validation

### Post-Mitigation Checks

1. **API Health Check**
```bash
# Test health endpoint (should return 200)
curl -v http://localhost:8080/health 2>&1 | grep HTTP

# Test a few critical endpoints
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/v1/fleet/vehicles
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/v1/tracking/telemetry

# Automated smoke test (if available)
./tests/run-smoke-tests.sh
```

2. **Monitor Metrics**
   - Grafana → API Dashboard
   - Check: Error rate < 0.1%, Latency P95 < 300ms, Availability 100%
   - Wait 5-10 minutes for metrics to stabilize

3. **Check Logs**
```bash
kubectl logs -f deployment/fleet-core -n flyfleet --since=5m | grep -i error
# Should show no errors OR only expected non-critical warnings
```

4. **Verify Database**
   - Check connection pool not exhausted
   - No long-running queries
   - Replication healthy (if applicable)

5. **Customer Validation**
   - If status page was updated, verify customer reports
   - Customer support: Are tickets decreasing?
   - Sample external API calls from different regions

### Stabilization Period

- Maintain elevated monitoring for 30 minutes post-recovery
- IC monitors Grafana continuously for first 15 minutes
- Update status page to "resolved" after 15 minutes stable
- Keep war room open for 30 min post-resolution for regression

---

## Post-Incident Actions

### Immediate (Within 24h)

1. ✅ Schedule PIR meeting (within 5 business days)
2. ✅ Preserve all logs (don't rotate/delete)
3. ✅ Save Prometheus query snapshots (Grafana → Share → Link)
4. ✅ Export incident timeline from Slack war room
5. ✅ Document decisions in Linear ticket

### PIR Checklist

- [ ] Complete timeline with UTC timestamps
- [ ] Root cause analysis (5 Whys or Fishbone)
- [ ] Impact assessment (users, revenue, duration)
- [ ] What went well?
- [ ] What didn't go well?
- [ ] Action items with owners and due dates
  - [ ] Example: "Increase DATABASE_MAX_OPEN_CONNS from 25 to 50"
  - [ ] Example: "Add circuit breaker for payment API"
  - [ ] Example: "Add automated rollback on failed health checks"

### Follow-up

- Review PIR in engineering all-hands
- Track action items in Linear project
- Update runbooks based on learnings
- Create new alerts if monitoring gaps identified

---

## Related Runbooks

- [Database Performance](database-performance.md) - If database is root cause
- [NATS JetStream Issues](nats-issues.md) - If messaging system implicated
- [Authentication Failure](auth-failure.md) - If auth errors predominant
- [Data Inconsistency](data-inconsistency.md) - If data corruption discovered

---

## Contacts & Escalation

| Role | Contact | When to Escalate |
|------|---------|------------------|
| On-call SRE | PagerDuty | IC for SEV1/SEV2 |
| Database Admin | @dba-slack | DB connection issues, corruption |
| NATS Expert | @nats-sme | JetStream problems, consumer lag |
| Security Officer | @security-slack | Suspicious activity, breach |
| VP Engineering | [phone] | No recovery path after 2h, revenue impact > $50k/hr |

---

**Last Updated:** 2026-03-29
**Review Date:** 2026-06-29
**Owner:** Platform Engineering Team
