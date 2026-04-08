# Post-Incident Review (PIR) Template

**Purpose:** Structured analysis following any SEV1 or SEV2 incident
**Timeline:** Draft within 48h, review meeting within 5 business days, final within 7 days
**Owner:** Incident Commander (with SMEs)

---

## Document Structure

```markdown
# Post-Incident Review: {Incident Title}

## 1. Executive Summary
- [2-3 sentences max]
- audience: executives, non-technical stakeholders

## 2. Incident Overview
- Timeline
- Impact metrics
- Response team

## 3. Root Cause Analysis
- Timeline of failure
- RCA method(s) used
- Contributing factors

## 4. Impact Assessment
- Customers affected
- Financial impact
- Compliance/SLA impact

## 5. What Went Well
- Positive observations

## 6. What Didn't Go Well
- Issues encountered during response

## 7. Action Items
- Preventive measures
- Corrective measures
- Each with owner and due date

## 8. Appendices
- Full timeline (chronological)
- Metrics graphs
- Relevant logs/excerpts
- Linear ticket link

---

## Detailed Template Sections

Copy and fill in each section below.
```

---

## 1. Executive Summary

```
{One paragraph (3-5 sentences) summarizing the incident for executives.

Include:
- What happened (briefly)
- Impact (who, how long, how bad)
- Root cause (high-level)
- Resolution status
- Next steps (PIR timeline)

Example:

On March 29, 2026, the FlyFleet API experienced a complete outage for 47 minutes (SEV1). All customers were unable to access the dashboard or mobile app, affecting approximately 2,500+ customers and resulting in an estimated $12,000 in lost revenue. The root cause was a database connection pool exhaustion triggered by a deployment that increased replica count without adjusting pool size. The incident was resolved by rolling back to the previous version and increasing the pool size. A post-incident review is underway to prevent recurrence, including automated pool size validation before deployment. Full PIR will be published by April 5, 2026.}
```

---

## 2. Incident Overview

### 2.1 Basic Information

```yaml
Incident ID: INC-YYYY-NNNN
Severity: SEV{1/2/3}
Declared: YYYY-MM-DD HH:MM UTC
Resolved: YYYY-MM-DD HH:MM UTC
Duration: X hours Ym
Linear Ticket: https://linear.app/flyfleet/issue/FF-NNNN
War Room: #incident-INC-YYYY-NNNN
```

### 2.2 Timeline (High-Level)

```
Time (UTC)      | Event
----------------|---------------------------------------------
14:30           | PagerDuty alert triggered (API error rate > 5%)
14:32           | On-call engineer acknowledged, IC declared
14:35           | War room established, executives notified
14:45           | Root cause identified: DB connection pool exhaustion
14:50           | Rollback deployment initiated
14:55           | Service restored, health checks passing
15:00           | Incident resolved, status page updated
15:30           | War room closed, PIR scheduled
```

### 2.3 Affected Services

```markdown
- [x] fleet-core API (primary)
- [ ] telemetry ingestion (not affected)
- [ ] command delivery (minor delay, did not fail)
- [ ] auth-service (no impact)
- [ ] database (root cause location)
- [ ] NATS (no impact)
- [ ] dashboard web (affected by API)
- [ ] mobile app (affected by API)
```

### 2.4 Impact Metrics

```yaml
Customers Affected:
  Total customers: 2,500
  Percentage: 100%
  Regions: Global (all regions)

Business Metrics:
  Revenue loss: ~$12,000 (47 min × ~$15,000/hr)
  SLA violation: Yes (Availability 99.95% target, achieved 99.2%)
  SLA penalty: $5,000 (contractual for enterprise)
  Support tickets: 450 created (normal: 20/hr)
  CSAT score: Dropped from 4.8 to 3.2

Technical Metrics:
  API error rate: Peak 100%, average 40% over incident
  Latency P95: Peak ∞ (timeouts), normal <300ms
  Uptime lost: 47 minutes

Compliance:
  Data breach: No
  GDPR violation: No
  India compliance: No
  SOC2 control failure: Yes (availability control)
```

### 2.5 Response Team

```markdown
**Incident Commander:** Sarah Chen (@sarah-chen)
- Assumed command: 14:32 UTC
- Total duration: 1h 23m

**Technical Lead:** Marcus Johnson (@marcus-j)
- Led investigation and recovery
- Database SME

**Subject Matter Experts:**
- @dba-team (PostgreSQL) - Database diagnostics and tuning
- @infra-sre - Kubernetes deployment and rollback
- @security-team - Verified no security implications

**Supporting Roles:**
- Communications Lead: Priya Sharma (@priya-sharma) - Stakeholder updates
- Note Taker: Wei Zhang (@wei-z) - Timeline documentation
- Customer Support: Daniel Kim (@daniel-support) - Ticket triage

**Escalations:**
- VP Engineering notified at 14:35 UTC
- Legal notified at 15:00 UTC (SLA penalty review)
- No executive crisis meeting required
```

---

## 3. Root Cause Analysis

### 3.1 Timeline of Failure (Detailed)

**Chronological reconstruction with UTC timestamps:**

```
YYYY-MM-DD HH:MM:SS | Action / Event | Who | Impact
-------------------|----------------|-----|-------
13:45:00 | Deployment: fleet-core v1.23.5 to 6 replicas | CI/CD | Latency begins increasing
14:00:00 | Connection pool metrics: active 95/100, waiting 5 | Auto | Warning threshold crossed
14:10:00 | Database connections at 100/100 (exhausted) | Auto | New connections blocked
14:15:00 | API error rate begins rising (from 0.1% to 5%) | Auto | First customer impact
14:20:00 | Health checks start failing on 2 pods | Auto | Kubernetes restarts 2 pods
14:25:00 | All pods failing health checks | Auto | Full outage begins
14:30:12 | PagerDuty alert: API error rate 50% | Alert | On-call paged
14:32:00 | On-call acknowledges, declares SEV1, assigns IC | James | Incident begins
...
```

**Use table format. Include:** Monitoring alerts, actions taken, decisions made, state changes.

---

### 3.2 Root Cause Statement

```
One-sentence root cause:

"The database connection pool was undersized for the number of application replicas, causing connection exhaustion and API failures."

Expanded:

After deployment of fleet-core v1.23.5 that increased replicas from 2 to 6, the database connection pool (max_open_conns=25) was insufficient to serve all replicas (6 replicas × ~5 connections each = 30 needed). When all 25 connections were used, new requests blocked waiting for connection, causing timeouts and health check failures, leading to complete outage.

The configuration change (replica count increase) was not accompanied by a connection pool adjustment, and there was no automatic guardrail to validate pool size against replica count.
```

---

### 3.3 RCA Method: 5 Whys

```
1. Why did the API go down?
   → Health checks failed due to database connection timeouts

2. Why did database connections time out?
   → All connections in pool were in use, new requests blocked

3. Why were all connections in use?
   → Connection pool max_open_conns=25 was insufficient for 6 replicas × 5 connections each = 30 concurrent connections needed

4. Why was pool size not adjusted when replicas increased?
   → Deployment increased replica count via HorizontalPodAutoscaler, but pool size was hardcoded in application config, not dynamically sized

5. Why was there no detection/prevention of this mismatch?
   → No automated validation of pool size vs. replica count in deployment pipeline, no alert on connection pool saturation before full exhaustion
```

---

### 3.4 RCA Method: Fishbone Diagram (Ishikawa)

**Categories and findings:**

```
┌─────────────────────────────────────────────────────┐
│                    API OUTAGE                       │
├─────────────┬─────────────┬─────────────┬───────────┤
│  People     │  Processes  │  Technology  │  External │
├─────────────┼─────────────┼─────────────┼───────────┤
│ • DBA not   │ • No        │ • Connection │ • None    │
│   consulted │   architectural│ pool sizing  │           │
│   on scaling│   review for│   formula    │           │
│             │   pool sizing│ • No dynamic │           │
│             │   changes   │   sizing     │           │
├─────────────┼─────────────┼─────────────┼───────────┤
│  Policies   │  Procedures │  Tools       │  Data     │
├─────────────┼─────────────┼─────────────┼───────────┤
│ • Deployment│ • No pre-   │ • No alert  │ • No      │
│   process   │   flight    │   on pool   │   pool    │
│   doesn't   │   check for │   utilization│   utilization│
│   validate  │   pool sizing│   at 80%    │   metrics │
│   config    │   with replica│ • No monitoring│         │
│             │   count      │   of replica│           │
│             │             │   count vs pool│         │
└─────────────┴─────────────┴─────────────┴───────────┘
```

---

### 3.5 Contributing Factors

```markdown
**Direct Cause:**
- Connection pool exhaustion (max_open_conns=25 too small for 6 replicas)

**Contributing Factors:**
1. Configuration drift - pool size not updated when replica count changed
2. Lack of guardrails - no validation in deployment pipeline
3. Monitoring gap - no alert on pool utilization > 80%
4. Architecture - static pool sizing vs. dynamic replica count
5. Testing gap - load testing did not simulate multi-replica scenario

**Not Contributing:**
- Database performance was normal
- Network was stable
- NATS/JetStream not involved
- No code bugs in business logic
```

---

### 3.6 Evidence

**Metrics:**
- [Link to Grafana panel showing connection pool exhaustion](URL)
- [Link to Prometheus query of `pgpool_connections`](URL)
- [Prometheus alert that fired](screenshot or query)

**Logs:**
- fleet-core logs showing "context deadline exceeded" (excerpt)
- PostgreSQL logs showing "remaining connection slots are reserved" (excerpt)

**Configuration:**
- fleet-core deployment showing replica count: 6
- fleet-core config showing DATABASE_MAX_OPEN_CONNS=25

**Code:**
- Repository connection pool initialization code (relevant lines)

---

## 4. Impact Assessment

### 4.1 Customer Impact

```markdown
**Scope:**
- Total customers: 2,500
- Fully affected: 100% (all customers)
- Partially affected: 0% (complete outage)
- Duration: 47 minutes (14:25 - 15:12 UTC)

**Impact Types:**
- Dashboard/web app: Completely unavailable
- Mobile app: Unable to load data
- API consumers: 100% error rate
- Emergency SOS: Not functional (critical safety issue)

**Customer Feedback:**
- Support tickets: 450 (vs. normal 20/hr)
- Top complaints: "Can't track vehicles", "Dashboard won't load"
- No customer churn reported (yet)
- CSAT: 3.2 (down from 4.8)

**Geographic Impact:**
- Global (all regions)
- No regional concentration
```

### 4.2 Financial Impact

```yaml
Direct Revenue Loss:
  Calculation: $15,000/hr × 0.78 hr = $11,700
  Source: Average revenue per hour from billing system

SLA Penalties:
  Enterprise contracts: $5,000 penalty for >30 min downtime
  Number of contracts: 2
  Total penalty: $10,000

Support Cost:
  Additional agent hours: 10 hrs × $50/hr = $500
  Overtime cost: $200

Total Direct Cost: $22,400

Indirect Impact (TBD):
  Customer churn estimate: 1-2% in next 90 days?
  Brand damage: TBD from CSAT recovery
  Sales cycle delays: Deals paused during incident
```

### 4.3 SLA & Compliance Impact

```yaml
SLA Availability:
  Target: 99.95% (26.28 min downtime/month acceptable)
  Achieved this month: 99.89% (43 min downtime)
  Violation: Yes (43 > 26.28)
  Penalty triggered: Yes

SOC2:
  Trust criterion: Availability (A1.2)
  Status: Control failure
 客户影响: Notified auditors, will be included in annual report
  Remediation: Action items tracked

GDPR:
  Data access: No unauthorized access
  Data availability: Temporary denial of access (not a breach)
  Notification required: No

India Compliance:
  FASTag/E-Way Bills: No impact (not used during outage)
  Data localization: No impact
```

---

## 5. What Went Well

```markdown
**Positive observations to acknowledge:**

- **Detection:** Alert fired within 5 minutes of error rate crossing threshold (14:30 from 14:25 start)
- **IC Assignment:** On-call acknowledged within 2 minutes and declared SEV1 promptly (14:32)
- **War Room:** Slack channel created and stakeholders notified within 3 minutes
- **Rollback:** Rollback command executed correctly and completed in 5 minutes
- **Communication:** Status updates sent on schedule (15-min cadence)
- **Customer Support:** Team used prepared templates, handled 450+ tickets efficiently
- **No Data Loss:** All data remained intact, no corruption
- **No Security Incident:** No unauthorized access during outage
- **Team Coordination:** DBA and SRE collaborated effectively
- **Executive Awareness:** Leadership informed within 5 minutes of IC assignment

**Specific praises:**
- @marcus-j correctly hypothesized connection pool issue within 10 minutes
- @dba-team provided immediate assistance despite off-hours
- @priya-sharma drafted clear customer communications quickly
```

---

## 6. What Didn't Go Well

```markdown
**Areas for improvement:**

**During Response:**
- **Initial Diagnosis:** Spent 15 minutes checking NATS and Redis unnecessarily before focusing on DB (wasted time)
- **Communications:** First update to executives lacked clear ETA (improved in subsequent updates)
- **War Room Organization:** Multiple parallel Slack threads caused confusion (centralized in #incident- channel after 30 min)
- **Customer Support Inbound:** Ticket surge overwhelmed initial queue, some delays in first hour
- **Status Page Updates:** One 30-min update was 12 minutes late (technical difficulty with publishing)

**Systemic Issues:**
- **Configuration Management:** Pool size hardcoded, not tied to replica count
- **Deployment Guardrails:** No pre-flight check for pool size vs. replicas
- **Monitoring:** No alert on connection pool utilization > 80% (only on errors)
- **Testing:** Load testing did not cover 6-replica scenario
- **Documentation:** Runbook did not explicitly call out connection pool sizing for scale events
- **Runbook:** Not followed initially (checked wrong systems first)

**Team:**
- On-call engineer reported fatigue (12h shift) - consider shift length review
- IC had to remind team to use status thread for updates (discipline needed)
```

---

## 7. Action Items

### 7.1 Preventive (Fix the System)

| # | Action | Owner | Due Date | Success Criteria |
|---|--------|-------|----------|------------------|
| P1 | Add pre-deployment validation: pool_size >= (replicas × 5) | @marcus-j | 2026-04-05 | Deployment fails if pool too small |
| P2 | Implement dynamic connection pool sizing based on replica count | @alexk-fleet | 2026-04-12 | Pool auto-adjusts when HPA scales |
| P3 | Add Prometheus alert: pool_utilization > 80% for 5m | @infra-sre | 2026-04-01 | Alert fires in test environment |
| P4 | Update load testing to cover 1, 2, 6 replica scenarios | @qa-lead | 2026-04-10 | Test suite includes multi-replica |
| P5 | Document connection pool sizing guidelines in runbook | @wei-z | 2026-04-03 | Runbook section complete |
| P6 | Implement deployment approval guardrail in CI/CD | @ci-cd-team | 2026-04-15 | Pipeline blocks invalid deployments |
| P7 | Conduct tabletop exercise on database failures | @sarah-chen | 2026-04-20 | All engineers participate |

### 7.2 Corrective (Fix the Immediate)

| # | Action | Owner | Due Date | Status |
|---|--------|-------|----------|--------|
| C1 | Increase fleet-core DATABASE_MAX_OPEN_CONNS to 50 (done 14:55) | @infra-sre | 2026-03-29 | ✅ Done |
| C2 | Review and adjust pool size for all services | @dba-team | 2026-04-02 | ⏳ In progress |
| C3 | Process SLA penalty payment to affected customers | @billing-team | 2026-04-15 | ⏳ Pending |
| C4 | Draft customer compensation email (10% credit) | @priya-sharma | 2026-03-30 | ⏳ In progress |
| C5 | Update runbooks with lessons learned | @wei-z | 2026-04-03 | ⏳ Pending |

### 7.3 Process Improvements

| # | Action | Owner | Due Date |
|---|--------|-------|----------|
| I1 | Review all deployment runbooks for missing guardrails | @platform-lead | 2026-04-10 |
| I2 | Implement mandatory IC training for all senior engineers | @eng-ops | 2026-05-01 |
| I3 | Review on-call shift length policy (12h → 8h) | @hr-eng | 2026-04-15 |
| I4 | Create incident review checklist for ICs | @sarah-chen | 2026-04-08 |

---

## 8. Appendices

### Appendix A: Full Chronological Timeline

*(Attach separate document or expand this section with minute-by-minute details from note-taker's log)*

**Source:** War room Slack export, monitoring graphs, ticket comments

**Format:**
```
2026-03-29 14:25:12 - Metric alert fired in Prometheus (api_error_rate 45%)
2026-03-29 14:25:45 - PagerDuty page sent to on-call (@jameswong-sre)
2026-03-29 14:27:10 - James acknowledged page
2026-03-29 14:28:00 - James declared SEV1, created Linear ticket FF-1234
2026-03-29 14:28:30 - Incident Commander assigned: James Wong (self)
2026-03-29 14:29:00 - War room channel #incident-INC-2026-0123 created
...
```

### Appendix B: Metrics & Graphs

Include screenshots or links to:

- Grafana dashboard showing error rate spike
- Prometheus query results (connection pool saturation)
- Kubernetes pod status timeline
- Customer support ticket volume

### Appendix C: Relevant Logs

Key excerpts (not full logs):

```
fleet-core pod logs:
2026-03-29T14:25:00Z error Get(" contexts deadline exceeded"...
2026-03-29T14:26:00Z error driver: connection wrapper: failed to ping DB: pq: sorry, too many clients already

PostgreSQL logs:
2026-03-29T14:25:15Z LOG:  connection to database by user "fleet_user" failed: FATAL: sorry, too many clients already
```

### Appendix D: Configuration at Time of Incident

```yaml
fleet-core deployment:
  replicas: 6
  env:
    DATABASE_MAX_OPEN_CONNS: 25
    DATABASE_MAX_IDLE_CONNS: 25

PostgreSQL:
  max_connections: 100
  shared_buffers: 256MB
```

### Appendix E: Linear Ticket & References

- Linear: INC-2026-0123 / FF-1234
- Grafana dashboard: https://grafana.flyfleet.com/d/abc123
- Runbook used: runbooks/api-outage.md
- Deployment: fleet-core v1.23.5 (commit abcdef)

---

## 9. Sign-Off

**Review meeting attendees:**

```
Name              | Role                | Attended | Approved
------------------|---------------------|----------|---------
Sarah Chen        | Incident Commander  | Yes      | Yes
Marcus Johnson    | Technical Lead      | Yes      | Yes
David Park        | VP Engineering      | Yes      | Yes
Priya Sharma      | Communications      | Yes      | Yes
Wei Zhang         | Note Taker          | Yes      | N/A
```

**PIR Approved By:** David Park (@david-park) on YYYY-MM-DD

**PIR Published:** YYYY-MM-DD

**Action Items Tracked In:** Linear project "PIR Actions - INC-2026-0123"

---

## Template Usage Checklist

- [ ] Executive summary written (readable by non-technical exec)
- [ ] Timeline includes all key events with UTC timestamps
- [ ] Root cause clearly stated (not technical debt dump)
- [ ] RCA method applied (5 Whys or Fishbone)
- [ ] Impact metrics quantified (customers, dollars, minutes)
- [ ] What Went Well / What Didn't balanced (blameless)
- [ ] Action items have clear owners and due dates
- [ ] All appendices completed
- [ ] Review meeting held with stakeholders
- [ ] Document approved by IC and VP Eng

---

**Template Version:** 1.0.0
**Last Updated:** 2026-03-29
