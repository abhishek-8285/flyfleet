# SLO Integration with Incident Response

**Purpose:** How SLO-based alerts map to incident response workflows
**Based on:** OBSERVABILITY_DESIGN.md
**Target Audience:** SREs, Incident Commanders, Alert Managers

---

## Overview

FlyFleet uses **SLO-driven alerting** with **multi-window burn rates** for faster, more accurate incident detection. This guide explains how SLO alerts integrate with the incident response framework.

---

## SLO to Incident Mapping

### SLO Metrics & Alerts

| SLO | Indicator | Target | Alert Rule | Expected Severity |
|-----|-----------|--------|------------|-------------------|
| API Availability | % successful requests | 99.95% (30d) | `ApiAvailabilityBurnCritical` | SEV1 |
| API Latency | P95 duration | <300ms | `ApiLatencyViolation` | SEV2 |
| API Error Rate | % 5xx responses | <0.1% | `ApiErrorRateBurnCritical` | SEV1 |
| Telemetry Throughput | events/sec | >10,000 | `TelemetryThroughputDrop` | SEV2 |
| Telemetry Latency | P99 ingestion | <5s | `TelemetryLatencyViolation` | SEV2 |
| Command Success Rate | % successful | 99.9% (30d) | `CommandSuccessBurnWarning` | SEV2 |
| Command Latency | P95 ack time | <2s | `CommandLatencyHigh` | SEV2 |
| PostgreSQL Avail | pg_up | 100% | `PostgreSQLDown` | SEV1 |
| PostgreSQL Latency | P95 query | <100ms | `PostgreSQLQueryLatencyHigh` | SEV2 |
| NATS Storage | % used | <85% | `NATSStorageHigh` (warning) / `NATSStorageCritical` | SEV2/SEV1 |

---

## Alert Rules Reference

### Multi-Window Burn Rate Alerts

```yaml
# Example: API Availability
- alert: FlyFleetApiAvailabilityBurnCritical
  expr: |
    (
      1 - (
        sum(rate(http_requests_total{status!~"5..",service="fleet-core"}[30m])) /
        sum(rate(http_requests_total{service="fleet-core"}[30m]))
      )
    ) / (0.05 / (30 * 24)) > 5.0  # 5x fast burn
  for: 2m
  labels:
    severity: critical
    slo: api-availability
  annotations:
    summary: "API availability error budget exhausted rapidly"
```

**Burn rate calculation:**
- Error budget consumed at 5x expected rate over 30min → SEV1
- Error budget consumed at 2x expected rate over 1h → SEV2 (warning)

---

### Threshold Alerts (Direct)

```yaml
- alert: FlyFleetApiLatencyViolationWarning
  expr: |
    histogram_quantile(0.95,
      sum(rate(http_request_duration_seconds_bucket{service="fleet-core"}[1h])) by (le)
    ) > 0.300
  for: 10m
  labels:
    severity: warning
    slo: api-latency
  annotations:
    summary: "API P95 latency above 300ms target"
```

---

## Alert Routing

Alertmanager routes based on `severity` label:

```
critical  → PagerDuty (immediate page) + Slack #sre-alerts
warning   → Slack #sre-alerts (no page)
info      → Slack #infrastructure-notifications
```

**SLO-specific routing:**
```yaml
routes:
  - match:
      severity: critical
    receiver: critical-pagerduty
    group_wait: 10s
    repeat_interval: 1h

  - match:
      slo: api-availability
    receiver: api-sre-team
    continue: true

  - match:
      slo: command-reliability
    receiver: device-team
```

---

## On-Call Response Flowchart

```
┌──────────────────────────────────────────────┐
│  Alert fires from Prometheus/Alertmanager   │
└─────────────────┬────────────────────────────┘
                  │
                  ▼
        ┌─────────────────┐
        │  PagerDuty      │
        │  (phone/sms)    │
        └────────┬────────┘
                 │
                 ▼
        ┌─────────────────┐
        │ Acknowledge within │
        │ SEV1: 5 min       │
        │ SEV2: 15 min       │
        └────────┬────────┘
                 │
                 ▼
        ┌─────────────────┐
        │ Check Grafana   │
        │ Dashboard       │
        └────────┬────────┘
                 │
                 ▼
        ┌─────────────────────────────────┐
        │ Is SLO violated?                │
        │ (error budget burn / threshold) │
        └────────┬────────────────────────┘
                 │
          YES ───┴─── NO
          │            │
          ▼            ▼
    ┌─────────┐  ┌──────────┐
    │ Declare │  │ False    │
    │ SEV1/2  │  │ positive │
    └────┬────┘  └──────────┘
         │
         ▼
    ┌─────────────────────┐
    │ Create Linear Ticket│
    │ INC-YYYY-NNNN       │
    └─────────┬───────────┘
              │
              ▼
    ┌─────────────────────┐
    │ Build timeline from │
    │ monitoring data     │
    └─────────┬───────────┘
              │
              ▼
    ┌─────────────────────┐
    │ Run relevant        │
    │ runbook             │
    └─────────────────────┘
```

---

## SLO Burn Rate to Severity Decision Tree

```
Error Budget Consumption Rate (burn rate):
- Fast: >5x expected consumption over short window
- Average: ~1x (expected)
- Slow: <1x (remaining budget slowly)

Alert thresholds:
- Fast burn (5x) for 2min → SEV1 critical (page)
- Fast burn (2x) for 5min → SEV2 warning (Slack)
- Slow burn (>1x) for 24h → SEV3 (ticket, monitor)

Example Calculation:
Target availability: 99.95% over 30 days
Allowed downtime: 30d × 0.05% = 216 minutes

If 30min track:
- Expected burn: (30min / 216min) = 0.139 (13.9%)
- Alert at 5×: 0.139 × 5 = 0.694 (69.4% consumed in 30min)
- If current availability = 30% error rate = 70% consumed → Alert fires
```

---

## Runbook Selection by SLO Violation

When SLO alert fires, consult appropriate runbook:

| SLO Violation | Alert Rule | Runbook |
|---------------|------------|---------|
| API availability ↓ | `ApiAvailabilityBurn*` | `runbooks/api-outage.md` |
| API latency ↑ | `ApiLatencyViolation` | `runbooks/api-outage.md` |
| API error rate ↑ | `ApiErrorRate*` | `runbooks/api-outage.md` |
| Telemetry throughput ↓ | `TelemetryThroughputDrop` | `runbooks/telemetry-disruption.md` |
| Telemetry latency ↑ | `TelemetryLatencyViolation` | `runbooks/telemetry-disruption.md` |
| Command success ↓ | `CommandSuccessBurn*` | `runbooks/command-delivery-failure.md` |
| Command latency ↑ | `CommandLatencyHigh` | `runbooks/command-delivery-failure.md` |
| PostgreSQL down | `PostgreSQLDown` | `runbooks/database-performance.md` |
| PostgreSQL latency ↑ | `PostgreSQLQueryLatencyHigh` | `runbooks/database-performance.md` |
| NATS storage ↑ | `NATSStorageHigh/Critical` | `runbooks/nats-issues.md` |
| Auth failures | `AuthErrorRateHigh` | `runbooks/auth-failure.md` |

---

## Dashboard Quick Links

Store these in war room topic for instant access:

```
Grafana Dashboards:
- Overview: https://grafana.flyfleet.com/d/fleet-overview
- API: https://grafana.flyfleet.com/d/fleet-api
- Telemetry: https://grafana.flyfleet.com/d/fleet-telemetry
- Commands: https://grafana.flyfleet.com/d/fleet-commands
- Infrastructure: https://grafana.flyfleet.com/d/fleet-infra
- SLOs: https://grafana.flyfleet.com/d/fleet-slos

Prometheus:
- Query: https://prometheus.flyfleet.com/graph

Alertmanager:
- Silences: https://alertmanager.flyfleet.com/#/silences
- Alerts: https://alertmanager.flyfleet.com/#/alerts
```

---

## Example SLO Incident Walkthrough

### Scenario: API Error Rate Burn Alert

```
14:30 UTC: Alertmanager fires: FlyFleetApiErrorRateBurnWarning
- expr: error rate 5% (threshold <0.1% for 2x burn)
- severity: warning, slo: api-errors

On-call (@james) acknowledges within 3 min.

14:33: IC checks Grafana:
- Error rate: rising from 1% → 5% over 10 min
- Latency: stable (200ms)
- Traffic: normal (1000 rps)
- Recent deployment: fleet-core v1.23.5 at 14:00 UTC

14:35: Declares SEV2 (error rate >5%, not yet critical 50%)

Checks runbook api-outage.md → Follows "Investigation Steps, #3: Application Logs"

14:37: Logs show "database connection timeout" errors

14:40: Checks DB - connection pool at 100/100

14:45: Identifies root cause: pool too small for 6 replicas

14:50: Rollback to v1.23.4 (pool=50)

15:00: Error rate back to 0.2%, latency normal

15:05: Declares resolution (15 min stable)

Total incident duration: 35 min (SEV2)

PIR action items:
- Add pre-deployment pool size validation
- Alert on pool >80% utilization
```

---

## SLO Error Budget Monitoring

### Error Budget Remaining

```promql
# API availability error budget remaining (percentage)
100 - (
  sum(rate(http_requests_total{status=~"5..",service="fleet-core"}[30d])) /
  sum(rate(http_requests_total{service="fleet-core"}[30d])) * 100
)
```

**Context:**
- Starting budget: 100% - target (99.95% = 0.05% allowed errors)
- Track consumption over rolling 30-day window
- When remaining budget < 20% → Review with VP Engineering

---

### Burn Rate Alert Tuning

**False positive?** Too sensitive. Adjust:
```yaml
# Increase threshold from 2x to 3x
... / (0.05 / (30 * 24)) > 3.0  # instead of 2.0
```

**Too slow?** Detect faster:
```yaml
# Use shorter window (30m instead of 1h)
... [30m] ... > 5.0
for: 1m  # instead of 5m
```

---

## SLO Dashboard Must-Have Panels

Create dedicated SLO dashboard with:

1. **Error Budget Remaining** (gauge) per SLO
   - api-availability: 85% remaining
   - command-success: 92% remaining

2. **Burn Rate** (timeseries) - current vs. expected
   - Fast burn? Alert

3. **SLO Compliance Status** (table)
   - Service | SLO | Target | Current | Status
   - fleet-core-api | Availability | 99.95% | 99.97% | ✅ OK

4. **Time to Burn** (estimate)
   - At current rate, budget exhausted in: X days

5. **Alert Summary** (count by service)
   - Active alerts: 3 (2 warning, 1 critical)

---

## During Incident: SLO-First Mindset

**Instead of:** "The API is slow"

**Think:** "This is an SLO violation: latency SLO breached 10min ago, error budget consumption rate is 3x average."

**Actions driven by SLO:**
- Which SLO violated? → Determines runbook and severity
- Burn rate? → Determines urgency (fast burn = act now)
- Error budget remaining? → Determines flexibility (can we burn rest during maintenance?)

---

## Post-Incident: SLO Review

After SEV1/SEV2, review:

1. **Did SLO alert fire in time?**
   - Detection lag < 2min? ✅ Good
   - Detection lag > 10min? Fix monitoring

2. **Was severity correct?**
   - SEV1 for fast burn? Correct
   - SEV2 but burned budget fast? Should be SEV1

3. **Did SLO capture incident?**
   - Incident affected customer experience but no SLO fired? Missing SLO → Create new one

4. **Error budget burn calculation accurate?**
   - Recalculate post-incident, verify

---

## SLO Alert Simulation

Test your SLO alerts before production:

```bash
# Simulate high error rate (inject failures in test)
curl -X POST http://prometheus:9090/api/v1/query -d 'query=increase(http_requests_total{status="500"}[1h])'

# Check alert evaluation delay
- Alert for: `for: 5m` means 5min after threshold crossed
- Expected: Alert within threshold + 5min

# Burn rate calculation sanity check:
# If 30d budget = 216min allowed downtime
# 30m fast burn threshold (5x) = 69.4% consumed in 30min
# Therefore: availability must be < 30.6% for 30min to trigger

# Test: Generate 50% error rate for 3min, see if alert fires
kubectl exec -it fleet-core -- curl -X POST http://localhost:8080/admin/metrics/inject?error_rate=50&duration=180
```

---

## Common SLO Alert Issues

| Symptom | Root Cause | Fix |
|---------|------------|-----|
| Alert fires but no customer impact | SLO too sensitive (threshold too low) | Increase target or adjust burn rate |
| No alert despite customer impact | SLO missing or wrong metric | Add new SLO, fix instrumentation |
| Alert fires 30min late | `for:` duration too long or scrape interval misaligned | Reduce `for:` to 2min, ensure Prometheus scrapes every 15s |
| Flapping alerts | Threshold too close to normal variance | Add hysteresis, increase threshold or for duration |
| Alert fatigue from same issue | No alert grouping or deduplication | Configure Alertmanager `group_by`, `repeat_interval` |

---

## SLO vs. Traditional Threshold Alerts

| Aspect | Traditional Threshold | SLO Burn Rate |
|--------|---------------------|---------------|
| **Time window** | Fixed (5min, 1h) | Relative to SLO period (30d) |
| **Focus** | Instant state | Consumption rate |
| **False positives** | High (normal variance) | Low (accounts for expected noise) |
| **Actionability** | "Something bad now" | "Will breach if continues" |
| **Alert noise** | Many | Few (only when budget burning) |
| **User experience correlation** | Weak | Strong (directly tied to user impact) |

**Recommendation:** Use SLO alerts for SEV classification. Keep traditional alerts for infra issues (node down, disk full).

---

## Integration Checklist

- [ ] All critical services have defined SLOs (API, DB, Telemetry, Commands)
- [ ] Prometheus recording rules for SLIs deployed
- [ ] Burn rate alerts configured with correct severity labels
- [ ] Alertmanager routing by `slo` label works
- [ ] SLO dashboard populated with 20+ panels
- [ ] Error budget tracked in dashboard
- [ ] On-call trained on SLO-to-severity mapping
- [ ] Runbooks reference SLO indicators
- [ ] Post-incident includes SLO burn analysis
- [ ] Monthly SLO review scheduled with VP Engineering

---

## Further Reading

- **Google SRE Workbook:** Chapter "Defining SLIs and SLOs"
- **Prometheus Best Practices:** Alerting on SLOs
- **Observability Design (FlyFleet):** `OBSERVABILITY_DESIGN.md`

---

**Integration Guide Version:** 1.0.0
**Last Updated:** 2026-03-29
