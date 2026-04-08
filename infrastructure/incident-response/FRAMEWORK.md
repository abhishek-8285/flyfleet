# FlyFleet Incident Response Framework

**Version:** 1.0.0
**Effective Date:** 2026-03-29
**Owner:** Platform Engineering & SRE Teams
**Based on:** Incident Commander Skill v1.0.0

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [Severity Classification](#severity-classification)
3. [Response Protocols](#response-protocols)
4. [Communication Templates](#communication-templates)
5. [Runbook Index](#runbook-index)
6. [Roles & Responsibilities](#roles--responsibilities)
7. [Incident Lifecycle](#incident-lifecycle)
8. [Post-Incident Review](#post-incident-review)
9. [SLO-Based Detection](#slo-based-detection)
10. [Tools & Integration](#tools--integration)

---

## Overview

This framework establishes standardized procedures for responding to technology incidents at FlyFleet. It is designed to:

- **Minimize downtime** through rapid, coordinated response
- **Ensure clear communication** with stakeholders at all levels
- **Enable blameless analysis** to prevent recurrence
- **Maintain SLO compliance** for customer trust

### Application

This framework applies to all FlyFleet services and infrastructure:

- **Backend Services** (fleet-core, auth, gateway, etc.)
- **Infrastructure** (Kubernetes, database, NATS, Redis)
- **External Dependencies** (AWS, payment processors, etc.)
- **Client Applications** (Web dashboard, mobile apps)

---

## Severity Classification

All incidents MUST be classified according to business impact using the following criteria:

### SEV1 - Critical Outage

**Definition:** Complete service failure affecting all users or critical business functions

**FlyFleet-Specific Indicators:**
- Fleet tracking completely unavailable for all customers
- Emergency SOS functionality failing
- Billing/payment processing down
- Data loss or corruption affecting production data
- Security breach with PII exposure
- Regulatory compliance failure (e.g., India FASTag non-compliance)
- SLA violation with contractual penalties (> 1 hour downtime)

**Financial Impact:** > $10,000/hour or > 25% revenue loss

**Response Requirements:**
- ⚠️ **Immediate** (0-5 min): Page Incident Commander via PagerDuty
- 🚨 **Executive notification** within 15 minutes (CEO/CTO)
- 📢 **Customer status page** update within 15 minutes
- 🏥 **War room** established (dedicated Slack/Video channel)
- 👥 **All-hands** mobilization if needed
- 📊 **Status updates** every 15 minutes until resolution

**Communication Cadence:**
- Executives: Every 15 min
- Customer Support: Real-time
- Customers: Every 15 min via status page
- Engineering: Continuous in war room

---

### SEV2 - Major Impact

**Definition:** Significant degradation affecting subset of users or non-critical functions

**FlyFleet-Specific Indicators:**
- >25% of customers experiencing degraded service
- Telemetry ingestion delayed > 5 minutes
- Command delivery failing (remote vehicle control)
- Reports API slow (> 2s p95) or erroring (> 5%)
- Dashboard not loading for specific regions/tenants
- India compliance features partially unavailable
- Performance issues causing driver frustration

**Financial Impact:** $1,000-$10,000/hour or 5-25% revenue impact

**Response Requirements:**
- ⚠️ **On-call response** within 15 minutes
- 🎯 **Incident Commander** assigned within 30 minutes
- 📢 **Status page** update within 30 minutes (if customer-facing)
- 👥 **Stakeholder notification** within 1 hour
- 📊 **Status updates** every 30 minutes

**Communication Cadence:**
- Engineering Leadership: Every 30 min
- Product Management: Every hour
- Customer Support: Every 30 min
- Customers: Every 30 min via status page

---

### SEV3 - Minor Impact

**Definition:** Limited impact with workarounds available

**FlyFleet-Specific Indicators:**
- Single service/component affected
- <25% of users impacted
- Workaround available (e.g., manual report generation)
- Non-critical feature unavailable (e.g., branding customization)
- Performance degradation (< 2s p95 but > 1s baseline)
- Internal tools issues (admin panel slow)
- Monitoring false positive investigation

**Financial Impact:** < $1,000/hour or minimal revenue impact

**Response Requirements:**
- ⏱️ **Response within 2 hours** (business hours)
- 📝 **Next business day** response acceptable outside hours
- 📧 **Internal team notification** only
- 📊 **Status updates** at key milestones (investigation, mitigation, resolution)
- ⚠️ **Status page** update optional (discretion of IC)

**Communication Cadence:**
- Engineering Team: At milestones
- Product Management: At resolution
- Customer Support: If tickets generated
- Customers: None unless widespread complaints

---

### SEV4 - Low Impact

**Definition:** Minimal impact, cosmetic issues, or planned maintenance

**FlyFleet-Specific Indicators:**
- Cosmetic bugs (UI alignment, typo)
- Documentation errors or gaps
- Logging/monitoring gaps (missing metrics)
- Performance issues with no user impact (> baseline but still fast)
- Dev/test environment issues
- Non-urgent security findings (low severity)
- Feature requests / technical debt

**Response Requirements:**
- 📅 **Response within 1-2 business days**
- 📝 **Standard ticket/issue tracking** (Linear/Jira)
- 👤 **No special escalation** required
- 🔄 **Normal development workflow**

**Communication Cadence:**
- Standard development sprint updates
- No emergency communications

---

## Response Protocols

### Immediate Actions (0-15 minutes)

#### For SEV1/SEV2 Incidents:

1. **🔔 Detection & Alerting**
   - PagerDuty/Opsgenie alert fires automatically from Prometheus/SLO alerts
   - On-call engineer acknowledges within 5 minutes (SEV1) or 15 minutes (SEV2)
   - If no acknowledgment after timeout, escalate to backup on-call, then manager

2. **🏆 Incident Commander Assignment**
   - On-call engineer assumes IC role unless they delegate
   - IC MUST have authority to make decisions and allocate resources
   - IC creates incident tracking ticket immediately (Linear/Jira)
     - Prefix: `INC-YYYY-NNNN`
     - Fields: Severity, Services Affected, Impact Description, Start Time

3. **📞 War Room Establishment**
   - SEV1: Dedicated Slack channel `#incident-{ticket-id}` + Zoom bridge
   - SEV2: Slack channel `#incident-{ticket-id}` (video optional)
   - IC posts initial assessment and invites stakeholders
   - Designate note-taker (separate from IC)

4. **📢 Initial Stakeholder Notification**
   - Use templates in `templates/communication/` (see below)
   - SEV1: Notify executives within 15 min via Slack/phone
   - SEV2: Notify engineering leadership within 30 min via Slack
   - Customer Support: Alert immediately if customer-facing

5. **🔍 Initial Investigation**
   - Check recent deployments: `git log --oneline --since="1 hour ago"`
   - Review monitoring dashboards: Grafana Overview & Service-specific
   - Query logs: Loki or `kubectl logs` if local
   - Check NATS JetStream health: `nats server report`
   - Verify database: `pg_isready`, connection pool stats
   - Check AWS status page for external factors

---

### Active Response (15 min - Resolution)

#### Command & Control

The Incident Commander:

1. **Maintains situational awareness**
   - Active monitoring of metrics in Grafana
   - Regular check-ins with technical teams
   - Track timeline of actions and outcomes

2. **Coordinates resource allocation**
   - Can pull in any team member (SMEs)
   - Escalates to senior leadership as needed
   - Approves emergency spend (e.g., AWS support cases, vendor escalation)

3. **Makes technical trade-off decisions**
   - Rollback vs. fix-forward
   - Risk tolerance under pressure
   - Communication timing and content

4. **Manages external communications**
   - Status page updates
   - Customer support enablement
   - PR/legal coordination for SEV1

#### Technical Response Flow

```
1. Triage & Assessment (0-15 min)
   ├─ Identify affected services
   ├─ Estimate scope of impact
   ├─ Determine likely cause categories
   └─ Check for recent changes

2. Mitigation (15-60 min)
   ├─ Contain damage (disable feature, rollback, throttle)
   ├─ Restore service (restart, failover, scale)
   ├─ Verify mitigation effectiveness
   └─ Establish stabilization period

3. Recovery (60 min - Resolution)
   ├─ Full service restoration
   ├─ Validation (health checks, smoke tests)
   ├─ Gradual traffic ramp-up if needed
   ├─ Monitor for secondary failures
   └─ Declare resolution after 15 min stable

4. Handoff (Post-Resolution)
   ├─ Update status page to "resolved"
   ├─ Notify stakeholders
   ├─ Begin PIR process within 24h
   └─ Ensure monitoring in place for residual issues
```

---

### Communication Standards

#### Update Frequency

| Severity | Stakeholder | Frequency | Channel |
|----------|-------------|-----------|---------|
| SEV1 | Executives | Every 15 min | Slack + SMS |
| SEV1 | Customer Support | Real-time | Slack |
| SEV1 | Customers | Every 15 min | Status page |
| SEV2 | Engineering Leadership | Every 30 min | Slack |
| SEV2 | Product Management | Every hour | Email |
| SEV2 | Customers | Every 30 min | Status page |
| SEV3 | Engineering Team | At milestones | Slack |
| SEV3 | Customer Support | If tickets | Email |

#### Message Format

```
[SEV{level}] {Service} - {Status}

Current Status: {Investigating/Mitigating/Resolved}

Impact:
- {what's affected}
- {scope: X% users, Y regions, Z tenants}

Actions Taken:
- {action 1}
- {action 2}

Next Update: {timestamp}
War Room: {link}
Linear: {ticket link}

--
{IC Name}
```

---

## Communication Templates

See `templates/communication/` directory for full templates:

- **initial-notification-sev1.md** - Executive/Slack alert
- **initial-notification-sev2.md** - Engineering leadership
- **status-update.md** - Reusable update template
- **customer-communication.md** - Status page posts
- **resolution-announcement.md** - All-clear with PIR commitment
- **executive-summary.md** - For C-level post-incident

---

## Runbook Index

Comprehensive runbooks are in `runbooks/`:

| Runbook | Purpose | Trigger |
|---------|---------|---------|
| [API Outage](runbooks/api-outage.md) | API down or high errors | 5xx > 5% or latency > 2s |
| [Telemetry Disruption](runbooks/telemetry-disruption.md) | Ingestion delayed/failing | Telemetry lag > 5min |
| [Command Delivery Failure](runbooks/command-delivery-failure.md) | Device commands not delivered | Command success rate < 95% |
| [Database Performance](runbooks/database-performance.md) | DB slow or connections exhausted | P95 query > 500ms or pool > 90% |
| [NATS JetStream Issues](runbooks/nats-issues.md) | Message backlog or storage full | Consumer lag > 100MB or storage > 90% |
| [Authentication Failure](runbooks/auth-failure.md) | Login issues or token validation errors | Auth error rate > 5% |
| [Data Inconsistency](runbooks/data-inconsistency.md) | Data corruption or sync failures | Discrepancy detection |
| [Security Incident](runbooks/security-incident.md) | Unauthorized access, data breach | Security alert or anomaly |

---

## Roles & Responsibilities

### Incident Commander (IC)

**Primary responsibilities:**
- Own the incident response process
- Make resource allocation decisions
- Coordinate between teams and stakeholders
- Maintain incident timeline
- Ensure proper documentation
- Declare resolution
- Initiate PIR process

**Authority:**
- Can pull in any team member
- Can approve emergency changes
- Can escalate to executives
- Makes final decisions on trade-offs

**When to delegate:** IC can delegate to senior engineer for technical deep-dive, but retains command responsibility

---

### Technical Lead

**Primary responsibilities:**
- Lead technical investigation and diagnosis
- Execute mitigation and recovery actions
- Validate fixes before production deployment
- Coordinate with SMEs (subject matter experts)
- Document technical details for PIR

**Typical role:** Senior engineer from most affected service

---

### Subject Matter Experts (SMEs)

**Areas of expertise:**
- Database administration (PostgreSQL, connection pooling, replication)
- NATS/JetStream (stream management, storage, consumers)
- Kubernetes/EKS (pods, networking, scaling, IAM)
- AWS infrastructure (VPC, security groups, IAM, EBS)
- Security (IAM, KMS, Secrets, audit logs)
- Frontend (dashboard, web app, browser issues)
- Mobile (app-specific issues, push notifications)
- India compliance (FASTag, E-Way Bills, GST)

**Activation:** IC calls SMEs as needed based on incident type

---

### Communications Lead

**Primary responsibilities:**
- Draft and send stakeholder communications
- Update status page
- Coordinate with customer support team
- Manage external inquiries (press, partners)
- Ensure consistent messaging across channels

**Assignment:** IC assigns this role (often Product Manager or Tech Writer)

---

### Note-Taker

**Primary responsibilities:**
- Maintain incident timeline in the war room/Slack
- Capture decisions and rationale
- Document actions taken with timestamps
- Prepare initial PIR draft

**Assignment:** IC assigns this role (often junior engineer or PM)

---

## Incident Lifecycle

### Detection

Incidents enter the system through:

1. **Automated Alerts** (Primary)
   - Prometheus SLO-based alerts (burn rate, latency, error rate)
   - Alertmanager notifications to PagerDuty
   - On-call engineer acknowledgment

2. **Manual Reports** (Secondary)
   - Customer support tickets
   - Sales/account manager escalations
   - Internal monitoring dashboard review
   - Social media/forum monitoring

**IC Action:** Verify alert legitimacy before declaring incident (false positives possible)

---

### Classification

IC classifies severity within 5 minutes of detection using:

- Impact scope (# users, % of service)
- Business criticality (revenue, compliance, safety)
- SLA implications
- Financial loss potential

Use the [Severity Classification](#severity-classification) matrix above.

**Uncertain?** Start higher, can downgrade after investigation. Better to over-respond initially.

---

### Response

Follow protocols in [Response Protocols](#response-protocols) section.

**Key checkpoints:**

| Time | Milestone | Owner |
|------|-----------|-------|
| T+5min | IC assigned, ticket created | On-call |
| T+15min | War room established, exec notified (SEV1) | IC |
| T+30min | Root cause category identified | Technical Lead |
| T+60min | Mitigation in place OR recovery plan clear | IC |
| T+120min | Service restored OR clear path to resolution | IC |

**Escalation triggers:**
- No mitigation within 60 minutes → escalate to VP Engineering
- No recovery path within 120 minutes → executive crisis meeting
- Customer data involved → Legal/Compliance within 30 min

---

### Resolution

**Criteria for resolution:**
- Service fully restored to normal operating conditions
- All affected systems healthy for at least 15 minutes
- No residual issues detected in monitoring
- Customer impact ceased (or workaround stable)

**IC actions:**
1. Verify service health in Grafana (all metrics normal)
2. Confirm with technical lead that issue is resolved
3. Post resolution announcement (see templates)
4. Update status page to "resolved"
5. Close PagerDuty incident
6. Schedule PIR within 24 hours
7. Send follow-up to stakeholders with PIR timeline

---

## Post-Incident Review (PIR)

### PIR Requirements

**Timeline:**
- Draft circulated within 48 hours
- Review meeting within 5 business days
- Final PIR published within 7 days

**Mandatory Components:**

1. **Executive Summary** (50 words max)
2. **Timeline** - Chronological with timestamps (UTC)
3. **Root Cause Analysis** - Using at least one framework:
   - 5 Whys
   - Fishbone diagram
   - timeline-based
4. **Impact Assessment** - Customers, revenue, reputation, compliance
5. **What Went Well** - Positive observations
6. **What Didn't Go Well** - Issues encountered during response
7. **Action Items** - Specific, measurable, with owners and due dates

**Action Item Structure:**

```markdown
- [ ] {Specific task}
  Owner: @team or @person
  Due: YYYY-MM-DD
  Success criterion: {measurable outcome}
```

### PIR Templates

Use the `frameworks/pir-template.md` and `frameworks/rcamethods/` for structured analysis.

**PIR Review Meeting:**
- Attendees: IC, technical lead, SMEs, stakeholders
- Duration: 60 minutes max
- Focus: Learning, not blame
- Outcome: Approved action items with clear owners

---

## SLO-Based Detection

FlyFleet uses the SLI/SLO framework defined in `OBSERVABILITY_DESIGN.md`.

### Key SLOs

| Service | SLI | Target | Alerting |
|---------|-----|--------|----------|
| API (fleet-core) | Availability | 99.95% (30d) | Multi-window burn rate |
| API (fleet-core) | P95 Latency | < 300ms | Direct threshold |
| API (fleet-core) | Error Rate | < 0.1% | Multi-window burn rate |
| Telemetry Ingestion | Throughput | > 10,000 events/sec | Direct threshold |
| Telemetry Ingestion | P99 Latency | < 5s | Direct threshold |
| Command Delivery | Success Rate | 99.9% (30d) | Multi-window burn rate |
| Command Delivery | P95 Ack Latency | < 2s | Direct threshold |
| PostgreSQL | Availability | 99.99% | pg_up metric |
| PostgreSQL | P95 Query Time | < 100ms | Direct threshold |
| NATS JetStream | Storage Usage | < 85% | Direct threshold |
| NATS JetStream | In/Out Balance | ±1% | Direct threshold |

### Alert Configuration

Prometheus recording rules and alerts are in:

```
infrastructure/monitoring/prometheus-values.yaml
```

**Burn Rate Alert Example:**

```yaml
- alert: FlyFleetApiAvailabilityBurnWarning
  expr: |
    (
      1 - (
        sum(rate(http_requests_total{status!~"5..",service="fleet-core"}[1h])) /
        sum(rate(http_requests_total{service="fleet-core"}[1h]))
      )
    ) / (0.05 / (30 * 24)) > 2.0
  for: 5m
  labels:
    severity: warning
    slo: api-availability
  annotations:
    summary: "API availability error budget burning at 2x rate"
```

### Alert Routing

Alertmanager routes alerts by severity:

- **critical** → PagerDuty (immediate page)
- **warning** → Slack #sre-alerts
- **slo-burn** → Slack #sre-alerts with priority
- **info** → Slack #infrastructure-notifications

See `infrastructure/monitoring/alertmanager-config.yaml`

---

## Tools & Integration

### Essential Tools

| Tool | Purpose | Access |
|------|---------|--------|
| **Grafana** | Metrics dashboards | `http://localhost:3000` (local) or cloud URL |
| **Prometheus** | Metrics collection | `http://localhost:9090` |
| **Alertmanager** | Alert routing | `http://localhost:9093` |
| **Loki** (planned) | Log aggregation | `http://localhost:3100` |
| **Tempo** (planned) | Distributed traces | `http://localhost:3200` |
| **PagerDuty** | On-call alerts | web.pagerduty.com |
| **Linear** | Incident tracking | linear.app |
| **Slack** | War room / comms | #incident-* channels |
| **Kubernetes** | Infrastructure | kubectl, Lens |
| **Database** | PostgreSQL client | psql |
| **NATS** | Messaging health | nats server report |

### Local Development Setup

For local incidents (development environment):

```bash
# Check service status
docker-compose ps
# or
ps aux | grep -E "(postgres|redis|nats)"

# View logs
docker-compose logs -f <service>
# or
kubectl logs -f <pod> -n <namespace>

# Restart services
./scripts/restart_services.sh

# Health checks
curl http://localhost:8080/health
curl http://localhost:8080/metrics
```

### Production Commands

```bash
# Kubernetes context
kubectl config use-context eks-prod

# Check pod status
kubectl get pods -n flyfleet -o wide
kubectl get pods -n monitoring

# View logs
kubectl logs -f deployment/fleet-core -n flyfleet --tail=100

# Connect to NATS
nats report --server nats://nats.monitoring.svc.cluster.local:8222

# Database connectivity
kubectl exec -it postgres-0 -n flyfleet -- psql -U fleet_user fleet_db

# Prometheus queries
echo 'query=rate(http_requests_total[5m])' | \
  curl -G "http://prometheus.monitoring.svc.cluster.local:9090/api/v1/query"
```

---

## Training & Runbooks

### Quarterly Drills

Conduct tabletop exercises every quarter to maintain readiness:

1. **SEV1 Drill** - Simulate complete API outage
2. **SEV2 Drill** - Simulate telemetry ingestion lag
3. **SEV3 Drill** - Simulate single-region degradation

### On-Call Handbook

New on-call engineers must review:

- This framework (`FRAMEWORK.md`)
- Runbooks for assigned services (`runbooks/`)
- Grafana dashboards (Overview, API, Infrastructure)
- `docs/DEPLOYMENT_QUICKSTART.md` for deployment knowledge
- `docs/MONITORING_SETUP.md` for monitoring operations

---

## Continuous Improvement

### Metrics to Track

- **MTTD** (Mean Time to Detect): Target < 2 min
- **MTTR** (Mean Time to Resolve): Target < 30 min (SEV1), < 4h (SEV2)
- **Communication Latency** - Time to first stakeholder update
- **PIR Completion Rate** - % of incidents with PIR within 7 days
- **Action Item Completion** - % of PIR actions completed on time

### Review Process

Monthly:
- Review SLO compliance and error budget consumption
- Analyze alert noise/false positives
- Update runbooks based on real incidents

Quarterly:
- Full PIR review meeting (all ICs, engineering leadership)
- Update severity classification if needed
- Conduct drills and tabletops
- Review observability coverage gaps

---

## Appendix

### Emergency Contacts

| Role | Contact | Slack |
|------|---------|-------|
| VP Engineering | [phone] | @vp-engineering |
| Platform Lead | [phone] | @platform-lead |
| Security Officer | [phone] | @security-officer |
| AWS Support | 1-888-AWS-HELP | - |

### External Dependencies

Track critical third-party services:

- **AWS** (EKS, RDS, S3, Secrets Manager) - status.aws.amazon.com
- **NATS** (messaging) - status.nats.io
- **Payment Processors** (Stripe, etc.) - status pages
- **CDN/DNS** (Cloudflare) - status.cloudflare.com

### Compliance Considerations

- **Data Breach** → Involve Legal/Privacy within 30 minutes
- **GDPR** → Document data access, inform DPO
- **India Regulations** (FASTag, GST) → Notify compliance team immediately
- **SOC2** → Log all incident response actions for audit

---

**Framework Version:** 1.0.0
**Last Updated:** 2026-03-29
**Next Review:** 2026-06-29
**Maintained by:** Platform Engineering Team
