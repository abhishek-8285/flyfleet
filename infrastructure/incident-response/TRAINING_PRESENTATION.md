# FlyFleet Incident Response - Training Deck

**Convert to slides:** Copy each "# Slide X" section into PowerPoint/Google Slides

---

## Slide 1: Title

```
FlyFleet Incident Response
Training Workshop

Platform Engineering Team
March 2026
```

---

## Slide 2: Agenda

```
Today's Session (2 hours)

1. Why Incident Response? (15 min)
   - Current state
   - Risks without framework
   - Business impact

2. Framework Overview (20 min)
   - Severity classification
   - Roles & responsibilities
   - Incident lifecycle

3. Live Demo: API Outage (30 min)
   - Walk through runbook
   - Show commands
   - Fill out templates

4. Hands-On Exercise (30 min)
   - Tabletop scenario
   - Team decision-making
   - Communication practice

5. Next Steps (15 min)
   - Your responsibilities
   - Training schedule
   - Resources & support
```

---

## Slide 3: FlyFleet Production Readiness

```
✅ Infrastructure: 100% Ready
   - Terraform, EKS, monitoring, secrets, backups

✅ Security: A- Grade (0 vulns)
   - Dependency audit, hardening, compliance

✅ Observability: Complete
   - SLOs, Prometheus, Grafana, Alertmanager

✅ Documentation: Comprehensive
   - Deployment guides, runbooks, troubleshooting

⚠️  Incident Response: FRAMEWORK COMPLETE
   - Ready for training and deployment
```

---

## Slide 4: The Problem (Without Framework)

```
What Happens During Incident?

❌ No clear leader → Multiple people making decisions
❌ Slow response → "Who should I page?"
❌ Chaos in comms → 5 different Slack channels
❌ Uninformed stakeholders → Executives scrambling
❌ No learning → Same incident repeats
❌ SLA breaches → Customer churn

Result: MTTR 90+ minutes, PIR rate 20%, repeat incidents 30%
```

---

## Slide 5: The Solution (With Framework)

```
With Incident Management Framework:

✅ Clear IC → One decision-maker
✅ Fast triage → Severity matrix (30 sec)
✅ Unified war room → Single channel, organized
✅ Automated comms → Templates ready to send
✅ Structured PIR → Learn from every incident
✅ Tracked actions → Prevent recurrence

Result: MTTR <30 min, PIR rate >90%, repeat incidents <10%
```

---

## Slide 6: Severity Classification

```
SEV1 - CRITICAL OUTAGE
❌ 100% customers affected OR safety risk
❌ Revenue loss >$10k/hr
❌ Complete service failure
⏱️  Response: 0-5 min (page)
⏱️  Exec notify: 15 min
⏱️  Updates: Every 15 min

SEV2 - MAJOR IMPACT
❌ 25-99% customers affected
❌ Severe degradation
⏱️  Response: 15 min
⏱️  Exec notify: 30 min
⏱️  Updates: Every 30 min

SEV3 - MINOR IMPACT
❌ <25% customers, workaround exists
⏱️  Response: 2 hours (biz hours)

SEV4 - LOW IMPACT
❌ Cosmetic, no user impact
⏱️  Response: 1-2 business days
```

---

## Slide 7: FlyFleet SLOs (Your Guardrails)

```
API Service (fleet-core)
   Availability: 99.95% (30d)
   Latency P95: <300ms
   Error Rate: <0.1%

Telemetry Ingestion
   Throughput: >10,000 events/sec
   Latency P99: <5s

Command Delivery
   Success Rate: 99.9%
   Ack Latency P95: <2s

PostgreSQL
   Availability: 99.99%
   Query P95: <100ms

NATS JetStream
   Storage: <85%
   Balance: ±1% in/out

When SLO violated → Alert → Incident
```

---

## Slide 8: Incident Response Timeline (SEV1)

```
T+0min   Pager fires
T+2min   Acknowledged
T+5min   SEV1 declared, IC assigned, war room created
T+15min  Executives notified, status page updated
T+30min  Root cause category identified
T+60min  Mitigation in place OR recovery started
T+90min  Service restored (target)
T+120min Resolution declared

Key: Detection → Response → Recovery → Learning
```

---

## Slide 9: Roles & Responsibilities

```
Incident Commander (IC)
👑 Decision authority
👑 Coordinates resources
👑 Declares severity
👑 Communicates with execs
👑 Initiates PIR

Technical Lead
🔧 Leads investigation
🔧 Executes recovery
🔧 Validates fixes
🔧 Documents technical details

Subject Matter Expert (SME)
🎯 Provides expertise (DB, NATS, K8s, etc.)
🎯 Executes specialized commands
🎑 Consults on architectural decisions

Communications Lead
📢 Stakeholder updates
📢 Status page management
📢 Customer notifications
📢 Media/PR coordination

Note-Taker
📝 Maintains timeline
📝 Captures decisions
📝 Documents actions
📝 Prepares PIR draft
```

---

## Slide 10: Command Center (War Room)

```
SETUP (0-5 min)
┌─────────────────────────────────────┐
│  Slack: #incident-INC-2026-0123    │
│  Bridge: zoom.us/j/123456          │
│  IC: @sarah-chen                   │
│  Linear: INC-2026-0124             │
│  Grafana: [link]                   │
│  Runbook: api-outage.md            │
└─────────────────────────────────────┘

CHANNEL DISCIPLINE
✅ Use threads for side discussions
✅ @incident for urgent attention
✅ Post updates at top (not in thread)
✅ Share screenshots/links
❌ No @channel (spams observers)
❌ No speculation without evidence
❌ No blame/criticism

OBSERVERS
👀 Executives, other teams = read-only
👀 Use #incident-observers for questions
👀 IC will update you, don't ask "status?"
```

---

## Slide 11: Communication Standards

```
CADENCE BY SEVERITY

SEV1
├─ Executives: Every 15 min (Slack + SMS)
├─ Customer Support: Real-time (Slack)
├─ Customers: Every 15 min (status page)
└─ Engineering: Continuous (war room)

SEV2
├─ Eng Leadership: Every 30 min (Slack)
├─ Product: Every hour (email)
├─ Customers: Every 30 min (status page)
└─ Engineering: Every 30 min (war room)

SEV3
├─ Team: At milestones (Slack)
└─ Customers: Optional

UPDATE FORMAT
[SEV1] API Outage - Update #3 - 14:50 UTC
Status: Mitigating | Progress: 70%
✅ Error rate: 15% ↓ from 40%
🔧 Action: Rollback complete, 4/4 pods healthy
⏱️  Next: 15:05 UTC
Linear: INC-2026-0123
```

---

## Slide 12: Runbooks - Your Playbook

```
runbooks/api-outage.md
   Symptoms: 5xx errors, health failing
   Quick cmds: kubectl get pods, curl health, check DB
   Investigation: 6 steps (ordered)
   Mitigations: Restart, rollback, scale, increase pool
   Recovery: Validation checklist

runbooks/telemetry-disruption.md
   Symptoms: Ingestion lag, stale dashboard
   Quick cmds: nats consumer report, check fleet-core logs
   Investigation: 5 steps
   Mitigations: Restart consumer, scale, expand storage, purge

runbooks/database-performance.md
   Symptoms: Slow queries, timeout errors
   Quick cmds: pg_isready, check connections, slow queries
   Investigation: 6 steps
   Mitigations: Kill queries, add index, increase pool, VACUUM

... 5 more runbooks covering all FlyFleet services
```

---

## Slide 13: Live Demo - API Outage Walkthrough

```
SCENARIO
14:30 Alert: API error rate 50%
14:31 On-call acknowledges
14:32 Declares SEV1

STEP 1 - CHECK GRAFANA
Open: https://grafana.flyfleet.com/d/fleet-api
See: Error rate spiking, latency ∞

STEP 2 - FOLLOW RUNBOOK
Open: runbooks/api-outage.md
→ Investigation Step #1: Verify impact
curl localhost:8080/health → fails

→ Step #2: Check K8s
kubectl get pods -l app=fleet-core
Pods: Running, no crashes

→ Step #3: Check logs
kubectl logs fleet-core --since=10m | grep error
"context deadline exceeded" "pq: sorry, too many clients"

STEP 3 - IDENTIFY ROOT CAUSE
Check DB connections: 100/100 (exhausted)
fleet-core replicas: 6
DB pool size: 25
Need: 6 × 5 = 30 connections, only 25 available!

STEP 4 - MITIGATE
Rollback to previous version (pool=50)
kubectl rollout undo deployment/fleet-core
Wait 3 min, health checks 200 ✓

STEP 5 - RESOLVE
Error rate drops to <1% in 5 min
Declare resolved at 14:55 (25 min total)
```

---

## Slide 14: Communication Templates

```
INITIAL NOTIFICATION (SEV1)
🚨 [SEV1] Fleet Core API - Complete Outage
📅 Detected: 2026-03-29 14:30 UTC
🎯 Impact: 100% customers, $15k/hr loss
👥 IC: Sarah Chen (@sarah-chen)
📊 Metrics: Error 100%, Latency ∞
🚨 Actions: Rollback in progress
📢 Next update: 14:45 UTC
Linear: INC-2026-0123
War Room: #incident-INC-2026-0123

→ Send to: Executives, IC, war room
→ Time: Within 15 min of declaration

STATUS UPDATE
[SEV1] API Outage - Update #3 - 14:50 UTC
Status: Mitigating | Progress: 70%
✅ Rollback complete, 4/4 pods healthy
📊 Error: 15% ↓ from 40%
⏱️  Next: 15:05 UTC

→ Post in: War room channel (top-level, not thread)
→ Frequency: Every 15 min (SEV1) or 30 min (SEV2)
```

---

## Slide 15: Root Cause Analysis (PIR)

```
POST-INCIDENT REVIEW (PIR) TIMELINE

T+0h   Incident resolved
T+24h  PIR meeting scheduled
T+48h  Draft PIR circulated
T+5d   PIR review meeting (60 min)
T+7d   Final PIR published

PIR STRUCTURE (from template)
1. Executive Summary (3 sentences)
2. Incident Overview (timeline, impact, team)
3. Root Cause Analysis
   - 5 Whys OR Fishbone OR Timeline
4. Impact Assessment (customers, $, SLA)
5. What Went Well
6. What Didn't Go Well
7. Action Items (with owners & due dates)
8. Appendices (full timeline, metrics, logs)

ACTION ITEMS
ID  Category  Description              Owner   Due       Status
FF-PIR-2026-0124-001
  Preventive  Add pool size validation  @marcus  2026-04-05  Not Started

Track in Linear project until all done!
```

---

## Slide 16: Tabletop Exercise - Scenario

```
SCENARIO: Telemetry Ingestion Lag

14:00 Alert: Telemetry lag > 10 minutes
14:05 On-call acknowledges
14:07 Declares SEV2

YOUR TEAM'S TASK:
1. Determine severity (why SEV2 not SEV1?)
2. Open correct runbook
3. Follow first 3 investigation steps
4. Identify likely root cause
5. Choose mitigation strategy
6. Send first status update (use template)

HINTS:
- Check NATS consumer lag
- Check fleet-core processing rate
- Check database insert performance
- Likely: Consumer can't keep up with ingestion rate

DISCUSSION QUESTIONS:
- What SLO violated?
- Which metrics to check first?
- When to scale vs. optimize?
- Who to page for help?
```

---

## Slide 17: Your Responsibilities

```
ON-CALL ENGINEER (rotating)
✅ Acknowledge alerts within SLA (5/15 min)
✅ Declare severity using matrix
✅ Follow runbook steps
✅ Send timely updates
✅ Preserve logs for PIR
✅ Complete PIR action items

INCIDENT COMMANDER (senior engineer)
✅ Assume command within 5 min (SEV1) / 30 min (SEV2)
✅ Create war room, coordinate team
✅ Make resource allocation decisions
✅ Communicate with executives
✅ Escalate when stuck
✅ Initiate PIR within 24h

NOTE-TAKER (rotating)
✅ Document timeline in real-time
✅ Capture decisions with rationale
✅ Save metrics snapshots
✅ Prepare PIR draft
✅ Maintain evidence package

ALL ENGINEERS
✅ Read framework docs (before on-call)
✅ Participate in quarterly drills
✅ Complete assigned PIR actions
✅ Update runbooks when you learn something
```

---

## Slide 18: Training Schedule

```
REQUIRED (before first on-call shift)
├─ 1-hour session (today!)
├─ Read FRAMEWORK.md (30 min)
├─ Memorize SEVERITY_QUICK_REFERENCE (10 min)
├─ Study runbooks for your services (30 min each)
└─ Shadow 1 incident as observer (1 incident)

QUARTERLY (mandatory)
├─ Tabletop drill (1 hour)
├─ PIR review meeting (30 min)
└─ Framework update briefing (15 min)

AS-NEEDED
├─ New runbook training (when service added)
├─ IC training (for senior engineers)
├─ Communications training (for ICs/comms leads)
└─ Advanced RCA methods (for PIR leads)
```

---

## Slide 19: Resources & Support

```
DOCUMENTATION
📚 Main: infrastructure/incident-response/README.md
📚 Framework: FRAMEWORK.md (complete)
📚 Quick Ref: SEVERITY_QUICK_REFERENCE.md (print this!)
📚 Runbooks: runbooks/*.md (8 services)
📚 Templates: templates/communication/*.md

TOOLS
🛠️  Grafana: https://grafana.flyfleet.com
🛠️  Prometheus: https://prometheus.flyfleet.com
🛠️  Alertmanager: https://alertmanager.flyfleet.com
🛠️  Linear: https://linear.app/flyfleet
🛠️  Slack: #sre-alerts, #incident-*

EMERGENCY CONTACTS
👥 On-call: PagerDuty (auto)
👥 DBA: @dba-team (Slack) / +1-XXX-XXX-XXXX
👥 Security: @security-oncall
👥 VP Eng: @vp-engineering (SME escalation)
👥 AWS Support: 1-888-825-4375

QUESTIONS?
💬 Slack: #sre-team
📧 Email: sre@flyfleet.com
📖 Docs: See README.md for detailed index
```

---

## Slide 20: Questions & Practice

```
NEXT STEPS

1. Complete today's training
   - Fill feedback form
   - Schedule follow-up session if needed

2. Review assigned runbooks
   - API? → Study api-outage.md
   - DB? → Study database-performance.md
   - etc.

3. Print quick reference card
   - Keep at desk or on-call bag

4. Join first tabletop drill
   - Calendar invite coming
   - 1 hour, interactive, low-stakes

5. Shadow current on-call engineer
   - Sit in war room (read-only)
   - See real incident response

REMEMBER:
- Framework exists to help YOU
- Don't innovate during incident - follow runbooks
- Ask for help early (escalation is good)
- Every incident is a learning opportunity
- We are blameless - focus on systems, not people

Thank you!
Platform Engineering Team
```

---

**How to use:**
1. Copy each "Slide X" section into separate slide
2. Add FlyFleet branding/logo
3. Include screenshots of Grafana dashboards
4. Add speaker notes for presenters
5. Practice timing (aim for 2 hours total)

**Delivery tips:**
- Be interactive (ask questions, poll audience)
- Use real examples from recent incidents (if any)
- Live demo commands (type them, don't just show)
- Encourage "stupid questions" - better to ask now than during incident
- Have quick reference cards printed to hand out

**Success criteria:** After training, every participant can:
- Declare SEV1 within 30 seconds of alert
- Open correct runbook without hesitation
- Send first status update within 15 minutes
- Explain their role in war room

---

**Training Deck Version:** 1.0.0
**Last Updated:** 2026-03-29
