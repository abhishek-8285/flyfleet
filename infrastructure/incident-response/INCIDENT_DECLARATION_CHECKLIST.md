# Incident Declaration Checklist

**Use this when PagerDuty alerts.** Fill out as you go during first 15 minutes.

---

## Pre-Checklist (0-5 min)

- [ ] **Acknowledge PagerDuty** - Click acknowledge button immediately
- [ ] **Check Grafana** - Open Overview dashboard, note metrics
- [ ] **Determine severity** - Use SEVERITY_QUICK_REFERENCE.md
  - Circle one: SEV1 / SEV2 / SEV3 / SEV4
- [ ] **Decide to declare** - If SEV1/2, declare immediately

---

## SEV1 Declaration (< 5 min after acknowledgment)

### Mandatory Actions (Ticking Clock)

| Task | Details | Done? |
|------|---------|-------|
| **Create Linear ticket** | Prefix: `INC-2026-` + auto number | [ ] |
| **Assign IC** | Self or delegate someone | [ ] |
| **Create war room** | `#incident-INC-2026-NNNN` | [ ] |
| **Page IC backup** | If not self-assigning | [ ] |
| **Notify executives** | Within 15 min (SMS/Slack) | [ ] |
| **Update status page** | Within 15 min | [ ] |
| **Invite SMEs** | @dba-team, @infra-sre, etc. | [ ] |
| **Appoint note-taker** | Designate someone | [ ] |

---

## SEV2 Declaration (< 15 min after acknowledgment)

| Task | Details | Done? |
|------|---------|-------|
| **Create Linear ticket** | `INC-2026-NNNN` | [ ] |
| **Assign IC** | Within 30 min of detection | [ ] |
| **Create war room** | `#incident-INC-2026-NNNN` | [ ] |
| **Notify Eng Leadership** | Engineering managers/directors | [ ] |
| **Update status page** | Within 30 min (if customer-facing) | [ ] |
| **Invite SMEs** | Relevant to symptom | [ ] |

---

## Initial Assessment (First 15 min)

Fill this out for your first update:

```
 INCIDENT DETAILS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Service(s):
  - fleet-core API
  - telemetry-ingestion
  - command-delivery
  - database (PostgreSQL)
  - NATS JetStream
  - Other: __________

Symptoms:
  - [ ] 5xx errors (rate: ___%)
  - [ ] High latency (P95: ___ms)
  - [ ] Complete outage (health failing)
  - [ ] Data missing/stale
  - [ ] Commands failing
  - [ ] Other: __________

Customer Impact:
  - % of customers affected: ___
  - Regions: __________
  - Workaround available: Yes / No

SLO Violations:
  - [ ] API availability < 99.95%
  - [ ] Latency > 300ms
  - [ ] Error rate > 0.1%
  - [ ] Telemetry lag > 5min
  - [ ] Command success < 99.9%

Root Cause Category (check one):
  - [ ] Infrastructure (K8s, AWS)
  - [ ] Database (PostgreSQL)
  - [ ] Code (recent deploy)
  - [ ] Configuration (env vars, params)
  - [ ] External (AWS, NATS, Stripe)
  - [ ] Unknown / investigating

Current Action:
  - [ ] Investigating (no fix yet)
  - [ ] Rolling back deployment
  - [ ] Scaling up service
  - [ ] Restarting service
  - [ ] Consulting SME
  - [ ] Other: __________

ETA to Resolution:
  - [ ] < 30 min
  - [ ] 1-2 hours
  - [ ] Unknown
```

---

## First Status Update Template

Copy-paste into war room channel:

```
[{SEV}] {Service} - Initial Status - {HH:MM UTC}

---
📊 METRICS
• Error Rate: {X}%
• Latency P95: {X}ms
• Throughput: {X} rps

---
🎯 IMPACT
• Customers affected: {X}% ({approx count})
• Symptoms: {list}

---
🔍 CURRENT STATUS
• Assessment: {Investigating / Mitigating}
• Root cause: {Unknown / Likely: category}
• Action: {Current action being taken}

---
👉 NEXT UPDATE
• {HH:MM UTC} (in 15/30 min)

---
Linear: INC-2026-NNNN
Grafana: [link to dashboard]
Runbook: [link]
```

---

## War Room Channel Setup

**Create:**
```bash
/slack create incident-INC-2026-0123
```

**Set topic:**
```
IC: {Name} | Linear: INC-2026-0123 | Bridge: {zoom link} | Updates every 15min
```

**Invite:**
```
/invite @sre-oncall @dba-team @infra-sre @security-team @comms-lead @{specific-sme}
```

**Pin important:**
- Linear ticket link
- Grafana dashboard
- Runbook link
- Timeline doc (Google Doc)

---

## Notable Thresholds (Memorize)

### SLO-Based

| Metric | Normal | Warning | Critical |
|--------|--------|---------|----------|
| API Error Rate | <0.1% | 1-5% (SEV2) | >50% (SEV1) |
| API Latency P95 | <300ms | 1-2s (SEV2) | >5s (SEV1) |
| API Availability | >99.95% | 99-99.95% | <99% |
| DB Connections | <80% max | 80-90% | >90% |
| NATS Storage | <70% | 70-85% | >85% (SEV2), >95% (SEV1) |

### Operational

| Metric | Threshold |
|--------|-----------|
| Alert acknowledgment | <5 min (SEV1), <15 min (SEV2) |
| IC assignment | <5 min (SEV1), <30 min (SEV2) |
| Exec notification (SEV1) | <15 min |
| Status page update | <15 min (SEV1), <30 min (SEV2) |
| First update to war room | <30 min after declaration |
| Mitigation action | <60 min (SEV1), <2h (SEV2) |

---

## Common Mistakes Checklist

**Avoid these:**

- [ ] **Wait to declare** - If unsure, declare SEV1, can downgrade
- [ ] **No war room** - Always create, even for SEV2
- [ ] **IC does technical work** - IC coordinates, doesn't type
- [ ] **No note-taker** - PIR timeline impossible without
- [ ] **Late status update** - Stick to cadence religiously
- [ ] **Restart before logs** - Preserve evidence first
- [ ] **Siloed investigation** - Keep everyone in loop via war room
- [ ] **Forget status page** - Customers check there first
- [ ] **Ignore runbooks** - They exist to save time
- [ ] **No Linear ticket** - Everything must be tracked

---

## Escalation Triggers

**Escalate to VP Engineering if:**

- No mitigation path after 60 minutes (SEV1) / 120 minutes (SEV2)
- Revenue impact > $50k/hour threshold
- Customer safety at risk (SOS failure affecting emergency)
- Security breach suspected (PII exposure)
- IC requests escalation

**Escalate to Security Officer if:**

- Unauthorized access detected
- Data exfiltration suspected
- Malware/ransomware found
- Fraudulent activity
- Compliance violation possible

---

## Evidence Preservation (SEV1 Only)

Before rebooting/restarting:

```bash
# Save logs (24h before to 1h after)
kubectl logs -f deployment/fleet-core -n flyfleet --since=24h > /evidence/fleet-core.log
kubectl logs -f deployment/auth-service -n flyfleet --since=24h > /evidence/auth.log
kubectl logs -f postgres-0 -n flyfleet --since=24h > /evidence/postgres.log

# Save deployment history
kubectl rollout history deployment/fleet-core -n flyfleet --revision=all > /evidence/deployments.txt

# Save configs
kubectl get deployment,configmap,secret -n flyfleet -o yaml > /evidence/k8s-configs.yaml

# Screenshot Grafana panels (use browser screenshot)
# Save to /evidence/grafana-{timestamp}.png

# Export Linear ticket (via API or print to PDF)
```

Don't clean up until PIR approved!

---

## First Hour Timeline (SEV1)

```
Minute 0-5:   Acknowledge, declare, create war room
Minute 5-15:  Notify execs, update status page, appoint note-taker
Minute 15-30: Root cause category identified, mitigation in progress
Minute 30-60: Service restored OR clear recovery path established
Minute 60+:   Resolution and stabilization
```

If missed any milestone → escalate.

---

## Severity Change Decision

**Downgrade from SEV1 to SEV2 if:**
- Mitigation in place, customers recovering
- Impact reduced to <25% users
- Root cause confirmed but fix not yet deployed

**Downgrade from SEV2 to SEV3 if:**
- Workaround found
- Impact localized to single tenant
- No SLA risk

**Never upgrade without cause:**
- If impact increases or burn accelerates
- If treatment not working (no improvement in 30 min)

Document severity changes in war room and Linear ticket.

---

## Quick Phone Tree

**If phone paging instead of Slack:**

```
SEV1 Call Tree:
1. On-call acknowledges (5min)
2. On-call calls IC (if different person)
3. IC calls VP Engineering
4. IC calls Legal (if data breach)
5. IC calls PR/Communications (if external)

SEV2 Call Tree:
1. On-call → IC
2. IC calls Engineering Manager
3. IC calls Product Manager
```

---

## After Resolution

**Checklist:**

- [ ] Verify 15 min stable (metrics normal)
- [ ] Update status page to "resolved"
- [ ] Send resolution announcement (Linear + war room)
- [ ] Close PagerDuty incident
- [ ] Archive war room channel (`/archive`)
- [ ] Schedule PIR (within 24h, send calendar invite)
- [ ] Forward PIR action items to Linear project
- [ ] Thank team publicly (#engineering)
- [ ] IC takes minimum 1 day off (if SEV1 > 2 hours)

---

## This Checklist Itself

**Use this PDF/print this page.**

Keep at desk or in on-call documentation. During incident, don't innovate - follow the checklist.

**Questions?** See FRAMEWORK.md or ask @sre-lead.

---

**Checklist Version:** 1.0.0
**Last Updated:** 2026-03-29
