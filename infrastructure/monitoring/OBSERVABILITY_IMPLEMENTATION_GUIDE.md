# FlyFleet Observability Implementation Guide

**Version:** 2.0
**Date:** 2026-03-28
**Status:** PRODUCTION READY

---

## Overview

This guide documents the complete observability strategy for FlyFleet, covering:
- SLI/SLO framework for 27 domains
- Comprehensive metrics instrumentation
- Alert optimization
- Dashboards (business, technical, cost)
- Log aggregation and distributed tracing
- Cost optimization observability

---

## Table of Contents

1. [Architecture](#architecture)
2. [SLI/SLO Framework](#slislo-framework)
3. [Metrics Instrumentation](#metrics-instrumentation)
4. [Alerting Strategy](#alerting-strategy)
5. [Dashboards](#dashboards)
6. [Logs & Traces](#logs--traces)
7. [Cost Optimization](#cost-optimization)
8. [Implementation Plan](#implementation-plan)
9. [Runbooks](#runbooks)

---

## Architecture

### Current Stack

```
┌──────────────────────────────────────────────────────────┐
│                     FlyFleet Stack                        │
├──────────────────────────────────────────────────────────┤
│  Go Services (27 domains)                                │
│  · fleet-core (modular monolith)                         │
│  · 27 domain packages                                    │
├──────────────────────────────────────────────────────────┤
│  Messaging & Data                                        │
│  · NATS JetStream (events)                              │
│  · PostgreSQL 15 (TimescaleDB)                          │
│  · Redis 7 (caching)                                    │
│  · MinIO (object storage)                               │
├──────────────────────────────────────────────────────────┤
│  Observability Stack                                    │
│  · OpenTelemetry v1.40 (tracing)                        │
│  · Prometheus (metrics & alerts)                        │
│  · Loki (logs)                                          │
│  · Tempo (traces)                                       │
│  · Grafana (dashboards)                                 │
│  · Alertmanager (notifications)                         │
├──────────────────────────────────────────────────────────┤
│  Infrastructure                                         │
│  · Kubernetes (EKS)                                     │
│  · Helm charts                                         │
│  · External Secrets Operator                           │
└──────────────────────────────────────────────────────────┘
```

### Instrumentation Standards

All domains MUST follow this pattern:

```go
// 1. Define metrics struct
type Metrics struct {
    RequestsTotal  *prometheus.CounterVec
    Latency        *prometheus.HistogramVec
    ErrorsTotal    *prometheus.CounterVec
}

// 2. Create init function
var GlobalMetrics = NewMetrics()

func NewMetrics() *Metrics {
    return &Metrics{
        RequestsTotal: promauto.NewCounterVec(...),
        Latency: promauto.NewHistogramVec(...),
        ErrorsTotal: promauto.NewCounterVec(...),
    }
}
```

---

## SLI/SLO Framework

### Global SLOs (All Services)

| SLO | Target | Window | Burn Rate Warning | Burn Rate Critical |
|-----|--------|--------|-------------------|--------------------|
| Availability | 99.95% | 30d | 2.0x | 5.0x |
| Latency p95 | 300ms | 7d | N/A | N/A |
| Error Rate | 0.1% | 30d | N/A | N/A |

### Domain-Specific SLOs

#### Critical Domains

**Emergency Response:**
- SOS delivery: 99.99% within 60s (p95), 99.9% within 180s (p99)
- Notification success: 99.9%
- Emergency response time: p95 < 60s, p99 < 180s
- Error budget consumption warning: 1.5x, critical: 3.0x

**Telemetry Ingestion:**
- Throughput: 10,000 events/sec minimum
- Ingestion latency: p99 < 5s
- Data loss rate: < 0.01%

**Command & Control:**
- Command success rate: 99.9%
- Command ACK latency: p95 < 2s
- Queue depth: < 10,000 pending

#### Financial Domains

**Billing:**
- Invoice generation: 99.8% success rate
- Payment processing: 99.5% success rate
- Invoice accuracy: 99.99%

**Finance/India Compliance:**
- GST return filing: 100% on-time (mandatory)
- E-Way Bill generation: 99.5% success
- FASTag recharge: 99.8% success

### SLO Recording Rules (Prometheus)

See: `slo-recording-rules.yml`

All SLOs use standardized naming:
```
sli:<domain>:<indicator>:<aggregation>
Examples:
- sli:api:availability:percentage
- sli:telemetry:latency:p99
- sli:commands:success_rate:percentage
```

---

## Metrics Instrumentation

### Completed Domains (27/27)

✅ **With Metrics:**
1. analytics
2. billing
3. compliance
4. customer
5. diagnostics
6. driver
7. emergency
8. finance
9. fuel
10. geofencing
11. governance
12. india
13. integrations
14. maintenance
15. monitoring
16. platform_compat
17. reports
18. route
19. settings
20. tracking
21. trip
22. vehicle
23. video
24. alert
25. alerts
26. auth
27. **admin_support** (new)

### Metric Naming Convention

```
flyfleet_<domain>_<metric>_<unit>
Examples:
- flyfleet_auth_login_attempts_total
- flyfleet_vehicle_telemetry_messages_total
- flyfleet_trip_duration_seconds
```

### Required Metrics per Domain

**Every domain MUST have:**
- ✅ Request/operation count (Counter)
- ✅ Error/failure count (Counter)
- ✅ Operation duration histogram (Histogram)
- ✅ Success rate gauge (Gauge) (optional, computed)

**Example Implementation:**

```go
package tracking

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
    TelemetryReceived   *prometheus.CounterVec
    TelemetryDropped    *prometheus.CounterVec
    ProcessingDuration  *prometheus.HistogramVec
    ProcessingErrors    *prometheus.CounterVec
}

func NewMetrics() *Metrics {
    return &Metrics{
        TelemetryReceived: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "flyfleet_tracking_telemetry_received_total",
                Help: "Telemetry messages received",
            },
            []string{"org_id", "device_id", "message_type"},
        ),
        // ... more metrics
    }
}
```

---

## Alerting Strategy

### Alert Categories

| Category | Count | Severity | Notes |
|----------|-------|----------|-------|
| SLO Burn Rate | 8 | Warning/Critical | Multi-window |
| Infrastructure | 10 | Warning/Critical | K8s, DB, NATS |
| Application | 12 | All levels | API, telemetry, devices |
| TLS/Compliance | 5 | Warning/Critical | Cert expiry |

### Alert Optimization Checklist

- [ ] Remove duplicates (e.g., DeviceOfflineSpike exists in 2 files)
- [ ] Tune thresholds (NATS backlog: 10GB → 5GB)
- [ ] Add runbook URLs to all alerts
- [ ] Include org_id in labels for multi-tenancy
- [ ] Set appropriate `for` duration (avoid flapping)
- [ ] Group by alertname, service, severity, org_id
- [ ] Implement escalation policies in Alertmanager

### Alert Quality Targets

| Metric | Target | Current |
|--------|--------|---------|
| Alert acknowledgment rate | >90% | TBD |
| Mean time to acknowledge | <15m | TBD |
| False positive rate | <10% | TBD |
| MTTR by alert type | Track | TBD |

### Missing Alerts to Add (Priority)

1. **High:** Telemetry message loss rate > 0.1%
2. **High:** Command queue depth > 10,000
3. **High:** Billing reconciliation mismatch > 0.5%
4. **Medium:** Geofence validation errors > 2%
5. **Medium:** Emergency SMS delivery failure > 5%
6. **Medium:** Route planning 5xx rate > 1%
7. **Low:** Map tile service errors
8. **Low:** DNS failure rate > 1%

---

## Dashboards

### Available Dashboards

#### 1. Overview (Existing)
- Basic golden signals
- Request rate, error rate, latency
- Database connections, pod status
- CPU, memory, NATS messages

#### 2. SLO Dashboard (Existing)
- Error budget remaining
- Burn rate tracking (multi-window)
- API availability, latency
- Command success rate, telemetry metrics

#### 3. Business Metrics (NEW)
**File:** `dashboards/business-metrics-dashboard.json`
- Active organizations, trips, drivers
- Revenue trends, profit margins
- Customer NPS, CSAT
- Fleet utilization
- Support performance

#### 4. Log Analytics (NEW)
**File:** `dashboards/log-aggregation-dashboard.json`
- Log volume by level/domain
- Error rate trends
- Top error messages
- Real-time log streams (emergency, auth)
- Log volume heatmap

#### 5. Distributed Tracing (NEW)
**File:** `dashboards/distributed-tracing-dashboard.json`
- Trace ingestion rate
- Average trace duration (p50, p95)
- Service dependency map
- Error rate heatmap
- Trace explorer

#### 6. Cost Optimization (NEW)
**File:** `dashboards/cost-optimization-dashboard.json`
- Total daily cost by service
- Budget vs actual
- Resource utilization efficiency
- Idle resource waste
- Cost anomalies
- Rightsizing opportunities
- Observability stack cost breakdown

### Dashboard Organization in Grafana

```
FlyFleet/
├── Overview
├── SLOs
├── Business Metrics
├── Domain Dashboards/
│   ├── API
│   ├── Tracking
│   ├── Emergency
│   ├── Billing
│   └── ...
├── Logs
├── Traces
└── Cost Optimization
```

---

## Logs & Traces

### Log Aggregation (Loki)

**Structured Logging Requirements:**

All logs MUST include these fields:

```json
{
  "timestamp": "2026-03-28T12:34:56Z",
  "level": "INFO",
  "org_id": "uuid",
  "domain": "tracking",
  "service": "fleet-core",
  "trace_id": "trace-id",
  "span_id": "span-id",
  "user_id": "uuid (if applicable)",
  "request_id": "uuid",
  "message": "Human-readable description",
  "error": "Error message (if error)",
  "stack_trace": "... (if panic)",
  "duration_ms": 123 (if applicable)
}
```

**Log Levels:**
- DEBUG: Detailed debugging info (dev only)
- INFO: Normal operations, trip started, etc.
- WARN: Expected errors, retries, deprecations
- ERROR: Operation failed, but service continues
- FATAL: Service crash, panic

**Log Retention:**
- Production: 30 days
- Staging: 7 days
- Development: 1 day

### Distributed Tracing (Tempo)

**Instrumentation Requirements:**

Every HTTP request MUST create a span:

```go
import "go.opentelemetry.io/otel"

func (h *Handler) DoSomething(w http.ResponseWriter, r *http.Request) {
    ctx, span := otel.Tracer("fleet-core").Start(r.Context(), "DoSomething")
    defer span.End()

    // Add attributes
    span.SetAttributes(
        attribute.String("org_id", orgID.String()),
        attribute.String("user_id", userID.String()),
    )

    // Your business logic here
}
```

**Trace Sampling Strategy:**
- Production: 1% head-based sampling (adjustable)
- Error traces: 100% sampling (tail sampling)
- High-latency traces (>1s): 100% sampling

**Trace Attributes:**
```go
span.SetAttributes(
    attribute.String("service.name", "fleet-core"),
    attribute.String("service.version", "v2.1.85"),
    attribute.String("deployment.environment", "production"),
    attribute.String("http.method", r.Method),
    attribute.String("http.route", routePattern),
    attribute.Int("http.status_code", statusCode),
    attribute.Float64("duration_ms", duration),
)
```

**Service Dependencies to Trace:**
1. HTTP handlers (all endpoints)
2. Database queries (use pgx instrumentation)
3. NATS publish/subscribe
4. Redis operations
5. External API calls (payment gateways, SMS, etc.)
6. gRPC calls (if any)

---

## Cost Optimization

### Cost Metrics to Track

**Infrastructure Costs:**
- Total daily cloud spend (USD)
- By service (compute, storage, network, managed)
- By environment (prod, staging, dev)

**Cost Efficiency:**
- Resource utilization (CPU, memory averages)
- Idle resources (unattached volumes, stopped VMs)
- Overprovisioning percentage
- Unallocated/untagged costs

**Savings Opportunities:**
- Spot instance coverage (target: 50%+ for non-critical)
- Reserved instance coverage (target: 70%+ for steady-state)
- Storage tier optimization (move old data to IA)
- Rightsizing opportunities

**Budget Tracking:**
- Monthly budget per org
- Budget consumption percentage
- Forecasted spend (30-day)
- Cost anomalies detected

### Alert on Cost Anomalies

```yaml
- alert: CostAnomalyDetected
  expr: flyfleet_cost_anomaly_score > 0.8
  for: 1h
  labels:
    severity: warning
  annotations:
    summary: "Cost anomaly detected: ${{ $value }} above baseline"
    description: "Service {{ $labels.service }} spending is {{ $labels.anomaly_type }} by ${{ $value }}/day"
```

### Cost Allocation Tags (Mandatory)

All resources MUST have these tags:

| Tag | Values | Required |
|-----|--------|----------|
| org_id | UUID | ✅ |
| environment | prod/staging/dev | ✅ |
| service | fleet-core, nats, postgres | ✅ |
| team | platform, fleet, billing | ✅ |
| project | flyfleet-main | ✅ |
| cost_center | cc-1234 | ⚠️ (if billing external) |

---

## Implementation Plan

### Phase 1: Foundation (Week 1-2)

**Goal:** Complete metrics instrumentation

- [x] Create metrics.go for all 27 domains ✅
- [ ] Add metrics registration to domain initialization
- [ ] Ensure all handlers increment metrics
- [ ] Deploy to staging and validate

**Validation Checklist:**
```bash
# Check all metrics are registered
curl http://localhost:8080/metrics | grep flyfleet_

# Verify metric cardinality
# Should be < 10,000 unique time series total
```

### Phase 2: Alert Optimization (Week 3)

**Goal:** Implement optimized alert rules

- [ ] Remove duplicate alerts
- [ ] Tune thresholds (per report)
- [ ] Add runbook URLs
- [ ] Configure Alertmanager routing/escalation
- [ ] Test alerts in staging

**Testing:**
```bash
# Load test rules
promtool check rules alert-rules.yml

# Test alert firing with synthetic metrics
# Use prometheus-synthetic-metrics-generator
```

### Phase 3: Dashboards (Week 4)

**Goal:** Deploy all dashboards to Grafana

- [ ] Import business metrics dashboard
- [ ] Import log analytics dashboard (Loki datasource)
- [ ] Import tracing dashboard (Tempo datasource)
- [ ] Import cost optimization dashboard
- [ ] Organize dashboard folders
- [ ] Set up dashboard sharing/export

**Users:**
- Executives: Business Metrics (reporting weekly)
- SREs: Overview + SLOs + Cost (daily)
- Developers: Domain-specific + Logs + Traces (as needed)
- Finance: Cost Optimization (weekly)

### Phase 4: Advanced Observability (Month 2)

**Goal:** Mature monitoring capabilities

- [ ] Implement anomaly detection on metrics
- [ ] Set up cost anomaly alerts
- [ ] Create runbook automation
- [ ] Implement automated capacity forecasting
- [ ] Set up alert fatigue tracking
- [ ] Create incident post-mortem templates

---

## Runbooks

### Alert Runbook Template

```markdown
# Alert: [Alert Name]

## Summary
Brief description of what this alert means.

## Impact
- User-facing: Yes/No
- Severity: Critical/Warning/Info
- SLA Impact: Yes/No

## Investigation Steps
1. Check Grafana dashboard: [link]
2. Check related metrics: `metric_name`
3. Review logs: `loki query`
4. Examine traces: `tempo query`

## Common Causes
- Cause 1
- Cause 2
- Cause 3

## Resolution Actions
1. Immediate fix: ...
2. Escalation: if unresolved after X minutes
3. Follow-up: create JIRA ticket

## Prevention
- Long-term fixes needed
- Process improvements

## References
- Architecture diagram
- Service documentation
- Related alerts
```

### Stored Locations

```
.github/runbooks/
├── slo-burn-alert.md
├── api-latency-investigation.md
├── api-errors-critical.md
├── telemetry-throughput.md
├── telemetry-latency.md
├── command-reliability.md
├── devices-offline.md
├── db-connection-pool.md
├── slow-queries.md
├── nats-storage.md
├── nats-backlog.md
└── high-memory.md
```

---

## Maintenance

### Quarterly Review

**Items to review:**
1. Alert performance (noise, false positives)
2. SLO targets (are they still appropriate?)
3. Metric cardinality (prevent explosion)
4. Dashboard usage (are they being used?)
5. Cost vs benefit (are we over-instrumenting?)

### Alert Tuning Process

1. **Collect feedback** from on-call engineers
2. **Analyze** alert history (true vs false positives)
3. **Adjust** thresholds, durations, labels
4. **Test** in staging
5. **Deploy** with monitoring

### Metric Deprecation

Before removing a metric:
1. Check Grafana dashboards using it (query usage)
2. Check alert rules using it
3. Announce deprecation in team channel
4. Keep metric for 30 days with `deprecated=true` label
5. Remove after confirmation

---

## Contacts

**Observability Team:**
- Slack: #observability
- Email: observability@flyfleet.com
- On-call: SRE rotation

**Runbook Repository:**
https://github.com/flyfleet/runbooks

**Grafana:**
https://grafana.flyfleet.com

**Documentation:**
https://docs.flyfleet.com/observability

---

## Appendix

### A. Metric Naming Reference

See: `slo-framework-complete.json` for complete SLI/SLO definitions.

### B. Alert Rule Reference

| Alert | Severity | Runbook | SLO |
|-------|----------|---------|-----|
| FlyFleetApiAvailabilityBurnWarning | warning | slo-burn-alert.md | ✅ |
| FlyFleetApiAvailabilityBurnCritical | critical | slo-burn-critical.md | ✅ |
| ... | ... | ... | ... |

### C. Dashboard Queries

Common PromQL patterns:

```promql
# Error rate (last 5m)
rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m])

# 95th percentile latency
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[1h])) by (le))

# 30-day error budget remaining
1 - (sum(increase(http_requests_total{status=~"5.."}[30d])) / sum(increase(http_requests_total[30d])))

# Burn rate (1h vs 30d target)
(1 - (success_rate_1h)) / ((1 - target) / (30 * 24))
```

---

**Next Steps:**
1. Review this guide with your team
2. Schedule implementation kickoff meeting
3. Assign Phase 1 tasks to developers
4. Set up staging environment for testing
5. Plan production rollout (canary deployment)

**Questions?** Contact observability team or open an issue in the documentation repo.
