# Root Cause Analysis Method: Timeline Reconstruction

**Purpose:** Chronologically sequence events to understand incident progression
**When to use:** Complex multi-service failures, cascading outages, unclear cause-effect
**Time:** 60-120 minutes (depending on incident duration)

---

## Method Overview

Timeline reconstruction creates a minute-by-minute (or second-by-second) account of the incident, from first anomalous event through detection, response, and resolution. This reveals:

- Early warning signs that were missed
- Lag between cause and effect
- Decision points and outcomes
- Handoffs and communication gaps
- Cascading failures

A detailed timeline is **required** for all SEV1 incidents.

---

## Timeline Structure

Use this table format:

```
Time (UTC)      | Event Type | Component | Description | Impact | Action
----------------|------------|-----------|-------------|--------|--------
14:23:45        | Metric     | fleet-core| API error rate rises to 5% | Warning detected | (none yet)
14:24:12        | Alert      | Prometheus| Alert "HighErrorRate" fires | Alert sent | PagerDuty
14:25:03        | Human      | On-call   | Pager acknowledged | - | Acknowledged page
14:25:30        | Decision   | IC        | SEV1 declared, war room opened | Investigation begins | Notified execs
...
```

**Event Types:**
- **Metric** - Monitoring threshold crossed
- **Alert** - Alerting system fired
- **Log** - Log entry (error, warning)
- **Human** - Human action (decision, command, communication)
- **Deployment** - Code/config change
- **External** - AWS outage, third-party failure
- **Customer** - Customer report/submission

---

## Example: API Outage Timeline (Partial)

```
Time (UTC)      | Event Type | Component        | Description                                      | Impact           | Action
----------------|------------|------------------|--------------------------------------------------|------------------|---------
13:45:00        | Deployment | CI/CD            | Deployed fleet-core v1.23.5 to 6 replicas        | Latency begins   | Automated
13:45:30        | Metric     | fleet-core       | Connection pool active: 60/100                  | Warning          | (none)
13:46:00        | Metric     | fleet-core       | Connection pool active: 75/100                  | Warning          | (none)
13:47:00        | Metric     | fleet-core       | Connection pool active: 90/100                  | Critical         | (none)
13:48:00        | Log        | fleet-core       | "context deadline exceeded" (10 instances)     | Latency ↑        | (none)
13:48:30        | Log        | PostgreSQL       | "remaining connection slots are reserved"      | DB rejecting      | (none)
13:49:00        | Metric     | fleet-core       | P95 latency: 5000ms (target 300ms)             | SLA breach       | (none)
13:50:00        | Metric     | fleet-core       | Error rate: 5%                                  | Customer impact  | (none)
13:50:45        | Alert      | Prometheus       | "ApiHighErrorRate" fired                       | PagerDuty notified| Alert
13:51:12        | Human      | On-call (James)  | Pager acknowledged                             | Response begins  | Clicked "ack"
13:51:30        | Human      | On-call (James)  | Checked Grafana, saw latency spike             | -                | Grafana
13:52:00        | Human      | On-call (James)  | Declared SEV1, created INC-2026-0124           | IC assumed       | Linear ticket
13:52:30        | Human      | On-call (James)  | Created war room #incident-...                 | Team joined      | Slack
13:52:45        | Human      | On-call (James)  | Notified VP Engineering via SMS               | Exec awareness   | SMS to +1...
...
```

---

## How to Build the Timeline

### Step 1: Data Collection (Gather Evidence)

**Sources to query:**

| Source | What to Extract | Tool/Command |
|--------|-----------------|--------------|
| **Prometheus** | All metrics at 1m resolution for incident window | Grafana Explore or `query_range` API |
| **Alertmanager** | Alerts that fired, timestamps, receivers | Check `alerts.json` or `/api/v2/alerts` |
| **Grafana** | Dashboard panel snapshots (take screenshots) | Screenshot tool |
| **Kubernetes** | Pod events, restarts, deployment rollout history | `kubectl get events --sort-by='.lastTimestamp'` |
| **Application Logs** | All logs filtered by time range (include ERROR, WARN) | `kubectl logs --since=...` |
| **Database Logs** | Slow queries, connection errors | `kubectl logs postgres-0 --since=...` |
| **Deployment History** | Deployments in previous 24h | `kubectl rollout history deployment/fleet-core` |
| **AWS CloudTrail** | API calls (if infrastructure changes) | CloudTrail console |
| **Slack War Room** | All messages in incident channel | Export from Slack |
| **Ticket/Linear** | Comments, status changes | Linear API or UI |

**Duration to cover:**
- Incident start: at least 2 hours BEFORE detection (early signs)
- Incident duration: detection to resolution
- Post-resolution: 1 hour after resolution to verify stability

---

### Step 2: Chronological Ordering

1. **Extract timestamp** from each data source
2. **Merge into single list** with consistent UTC timestamps
3. **Sort by timestamp** ascending
4. **Remove duplicates** (same event from multiple sources)
5. **Add milliseconds** if available (round to nearest second if not)

**Tool:** Use spreadsheet or script:

```bash
# Example: merge multiple log sources
# Extract: timestamp, source, message, severity

cat prometheus-metrics.csv | \
awk -F, '{print $1 "|METRIC|" $2 "|" $3}' > timeline.raw

cat app-logs.txt | \
grep -E "ERROR|WARN" | \
awk '{print $1 "|LOG|" $2 "|" $0}' >> timeline.raw

sort -t'|' -k1,1 timeline.raw > timeline.sorted

# Format as table
column -t -s '|' timeline.sorted > timeline.md
```

---

### Step 3: Event Categorization

**Label each event with type:**

| Type | Definition | Examples |
|------|------------|----------|
| **Metric** | Numeric value crossing threshold | error_rate > 5%, CPU > 90% |
| **Log** | Application/system log entry | "connection timeout", "panic" |
| **Alert** | Alertmanager/PagerDuty fired | "HighErrorRate fired" |
| **Human** | Person took action | "Rolled back deployment" |
| **Deployment** | Code/config change deployed | `kubectl set image...` |
| **External** | Third-party/system event | "AWS status: degraded" |
| **Detection** | Someone noticed issue manually | "Customer reported" |

---

### Step 4: Impact Annotation

For each event, describe **what changed**:

```
Good events (stable):
  - Metrics within SLO
  - All replicas healthy
  - Error rate < 0.1%

Degradation events:
  - Error rate rising
  - Latency increasing
  - Connections increasing

Customer impact events:
  - First customer report received
  - Support tickets created
  - Status page update posted

Recovery events:
  - Rollback started
  - Health checks passing
  - Traffic shifting back

Resolution events:
  - Error rate returned to normal
  - All systems healthy
  - Incident declared resolved
```

---

### Step 5: Decision Point Identification

Mark key inflection points:

```
[🔴 CRITICAL DECISION]
14:35 UTC - IC decides to rollback vs. fix-forward
- Options: Rollback (safe, known) vs. hotfix (risky, unknown)
- Chosen: Rollback (due to deployment as likely cause)
- Rationale: Recent deploy, quickest path to recovery

[🟡 ESCALATION]
14:40 UTC - VP Engineering page due to no progress after 10 min
- Reason: Over 10 min with no mitigation path
- Outcome: Additional resources allocated

[🟢 MILESTONE]
14:55 UTC - Service restored (health checks 200 for 5 min)
- Criteria: 95% traffic to healthy pods, error rate < 1%
- Verified by: @marcus-j confirmed in war room
```

---

## Timeline Analysis

After building timeline, analyze for patterns:

### Pattern 1: Early Warning Signs

```
13:45:00  Warning: Pool active 60/100
13:46:00  Warning: Pool active 75/100
13:47:00  Warning: Pool active 90/100
13:48:00  Error: "context deadline exceeded"
13:50:00  Customer impact: 5% error rate

OBSERVATION: Warnings existed 5 min before errors, 15 min before customer impact.
IMPROVEMENT: Alert on warning thresholds (pool > 80%) not just errors.
```

### Pattern 2: Detection Lag

```
13:50:00  Error rate reaches 5%
14:25:00  PagerDuty fires (35 min later)

OBSERVATION: 35-minute delay between customer impact and detection.
IMPROVEMENT: Check alert thresholds; did metric not cross threshold earlier?
EXPLANATION: Alert threshold set at 10%, should be 5% or lower for faster detection.
```

### Pattern 3: Response Delays

```
14:25:00  Alert fires
14:27:00  On-call acknowledges (2 min) ✓ Good
14:32:00  IC declared (7 min after alert) - Slow
14:45:00  Rollback initiated (20 min after IC) - Too slow

OBSERVATION: 20-minute gap between IC assign and mitigation action.
ROOT CAUSES:
  - Initial investigation focused on wrong systems (checked NATS first)
  - No runbook guidance for "high error rate" scenario
  - IC inexperienced with this service
IMPROVEMENTS:
  - Create runbook decision trees
  - IC training on common failure patterns
  - "Mitigation first" principle (don't over-investigate before containing)
```

### Pattern 4: Communication Gaps

```
14:35:00  IC notified executives ✓
14:45:00  Customer Support NOTIFIED (10 min late)
15:00:00  Status page updated (25 min late)

OBSERVATION: Customer-facing comms lagging.
ROOT CAUSE: Communications Lead not added to war room until 14:40.
IMPROVEMENT: Auto-include comms lead in initial war room creation.
```

### Pattern 5: Cascading Failures

```
14:10:00  DB pool 100/100
14:12:00  Fleet-core pod 1 health check fails
14:12:30  Kubernetes restarts pod 1
14:13:00  Other pods get additional load (5→6 active)
14:13:30  All pods now overloaded, full outage

OBSERVATION: Local failure (single pod) cascaded to full outage.
PREVENTION: Circuit breaker or bulkhead pattern to isolate failing pods.
```

---

## PIR Section: Timeline Analysis

### Include in PIR:

1. **High-level timeline** (as in Incident Overview) - 10-15 key events
2. **Detailed chronological log** (Appendix) - All events with timestamps
3. **Analysis of delays** - Detection time, response time, resolution time breakdown
4. **Decision points** - Key decisions with rationale and outcome
5. **Early warnings** - What we saw before impact that we missed

---

## Template for PIR Appendix

```
## Appendix A: Detailed Timeline

**Data Sources:** Prometheus (query_range), Kubernetes events, Linear tickets, Slack incident channel, PostgreSQL logs

**Note:** All timestamps UTC

| Time | Source | Event | Details |
|------|--------|-------|---------|
| 13:45:00 | Deployment | fleet-core v1.23.5 deployed | Replicas: 2→6, Image: abc123 |
| 13:45:30 | Metrics | db_connections_active: 60/100 | from fleet-core:db_connections |
| 13:46:00 | Metrics | db_connections_active: 75/100 | |
| 13:47:00 | Metrics | db_connections_active: 90/100 | ⚠️ Warning threshold crossed (80%) |
| 13:48:00 | Log | fleet-core | "context deadline exceeded" (10×) |
| 13:49:00 | Metrics | api_error_rate: 5% | First customer impact |
| 13:50:45 | Alert | Prometheus | Alert "ApiHighErrorRate" fired (error_rate > 5% for 2m) |
| 13:51:12 | Human | On-call (James) | Pager acknowledged |
| ... | ... | ... | ... |

**Total events:** 187 rows (condensed to 47 key events above)

---

## Analysis of Timeline

### Detection Lag

- **First Customer Impact:** 13:49:00 (error rate > 5%)
- **Alert Fired:** 13:50:45 (65 seconds later) ✓ Good
- **On-call Acknowledged:** 13:51:12 (27 seconds after alert) ✓ Good
- **IC Declared:** 13:52:00 (48 seconds after ack) ✓ Good
- **Total detection-to-response:** 3 minutes ✓ Acceptable

Note: Detection was timely. No improvement needed on alerting.

### Response Execution

- **First mitigation action (rollback) started:** 14:15:00 (23 min after IC)
- **Service restored:** 14:42:00 (50 min total duration)

**Gap analysis (13:52-14:15 = 23 min):**
- 13:52-14:00 (8 min): Investigation focused on wrong systems (NATS, Redis)
- 14:00-14:10 (10 min): Correctly identified DB pool issue but debated fix options
- 14:10-14:15 (5 min): Executive approval for rollback (delayed by waiting for VP)

**Root cause of delay:**
1. Initial hypothesis wrong (blamed NATS instead of DB) → Wasted 8 min
2. No clear "stop investigation, mitigate first" rule → Debate continued during outage
3. Rollback approval required VP (should be IC authority)

**Improvements:**
- Add decision tree to runbook: "When error rate > 50%, rollback immediately"
- Empower IC to approve rollback without escalation
- IC training: Mitigate first, investigate later (if rollback safe)

---

### Communication Timeline

| Time | Action | Status |
|------|--------|--------|
| 13:52:00 | Executives notified (Slack) | ✓ On time |
| 14:00:00 | Customer Support notified | ❌ 8 min late (should be 30 min) |
| 14:25:00 | Status page updated | ❌ 33 min late (should be 30 min) |
| 14:30:00 | Customer email blast | ❌ 38 min late (SEV1: 15 min) |

**Issue:** Comms lead not added to war room until 13:55, lacked context to act quickly.

**Improvement:** Auto-add @comms-lead to #incident-* channels via PagerDuty/Linear webhook.

---

## Key Insights from Timeline

1. **Detection was NOT the problem** - Alerting worked, on-call responded fast.
2. **Investigation was the bottleneck** - Wrong initial hypothesis cost 8 min.
3. **Mitigation delayed by decision-making** - 5 min spent waiting for VP approval (unnecessary).
4. **Communication lagged due to access** - Comms lead excluded from initial war room.
5. **Total time to mitigation:** 23 min (target < 15 min for SEV1)

---

## Timeline Reconstruction Template (For Note-Taker)

Use this during active incident (in war room):

```
# Incident Timeline Log

T=0: Incident declared (record exact time)

[Half-hourly intervals, add events as they occur]

=== 13:50-14:20 ===
13:52:00 [IC] Sarah declared SEV1, created INC-2026-0124
13:52:30 [WAR] Created #incident-INC-2026-0124, invited team
13:53:00 [METRIC] db_connections: 98/100 (from @dba-team)
13:54:00 [DECISION] Decide to investigate DB first (not NATS)
13:56:00 [LOG] fleet-core: "connection timeout" (increased frequency)
14:00:00 [ACTION] Checked pg_stat_activity - all connections active
14:02:00 [HYPOTHESIS] Connection pool exhaustion due to 6 replicas
14:05:00 [VALIDATION] Verified: 6 replicas × 5 connections each = 30 needed, pool=25
14:10:00 [DECISION] Rollback approved (VP paged, consent given)
14:15:00 [ACTION] Rollback initiated: kubectl rollout undo fleet-core
14:20:00 [STATUS] 2 pods rolled back, health checks passing

=== 14:20-14:50 ===
[...]
```

**Who maintains:** Designated note-taker (not IC, separate role)
**Where:** In war room Slack thread pinned message OR shared Google Doc
**Format:** Simple chronological list, timestamps always UTC

---

---

**Method Version:** 1.0.0
**Last Updated:** 2026-03-29
