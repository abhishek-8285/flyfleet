# FlyFleet Incident Response - Quick Reference Card

**Print this page (8.5" × 11") and keep at your desk or in on-call bag**

---

## 🚨 Severity Decision (30 seconds)

```
What % of customers impacted?
├─ 100% or safety risk → SEV1
├─ 25-99%           → SEV2
├─ <25%             → SEV3
└─ Cosmetic/bug     → SEV4

Revenue loss >$10k/hr? → SEV1
SLA violation critical? → SEV1
Emergency SOS failing? → SEV1
Complete outage? → SEV1
```

---

## ⏱️ Response Time SLAs

| Severity | Acknowledge | Declare | War Room | Exec Notify | Status Page | Updates |
|----------|-------------|---------|----------|-------------|-------------|---------|
| **SEV1** | 5 min | 5 min | 5 min | 15 min | 15 min | Every 15 min |
| **SEV2** | 15 min | 30 min | 30 min | 1 hour | 30 min | Every 30 min |
| **SEV3** | 2 hours | 2 hours | Optional | - | Optional | Milestones |
| **SEV4** | 1-2 days | Next day | - | - | - | Standard tickets |

---

## 📋 0-15 Minute Checklist (SEV1)

**Ticking clock - do in order:**

```
Minute 0-2:
☐ Acknowledge PagerDuty
☐ Open Grafana dashboard
☐ Determine severity (use matrix above)

Minute 2-5:
☐ Create Linear ticket: INC-YYYY-NNNN
☐ Assign yourself as IC (or delegate)
☐ Create war room: #incident-INC-YYYY-NNNN
☐ Page IC backup if you're not IC

Minute 5-10:
☐ Invite SMEs: @dba-team @infra-sre @security-team
☐ Appoint note-taker (someone else)
☐ Send SEV1 notification (use template)
☐ Update status page

Minute 10-15:
☐ Open relevant runbook
☐ Begin investigation (follow steps)
☐ First status update drafted
```

---

## 📞 Emergency Contacts

```
ON-CALL (auto): PagerDuty
INCIDENT COMMANDER: @ic-rotation (Slack) or see schedule
DATABASE: @dba-team or +1-XXX-XXX-XXXX
NATS/MESSAGING: @nats-team
SECURITY: @security-oncall or +1-XXX-XXX-XXXX
VP ENGINEERING: +1-XXX-XXX-XXXX (after 60 min no progress)
AWS SUPPORT: 1-888-825-4375 (24/7)
CUSTOMER SUPPORT: support@flyfleet.com
LEGAL: +1-XXX-XXX-XXXX (data breach, compliance)
```

---

## 🔧 Quick Commands (Memorize)

### Kubernetes
```bash
# Pod status
kubectl get pods -n flyfleet -o wide

# Logs (last 10 min)
kubectl logs -f deployment/fleet-core -n flyfleet --since=10m

# Events
kubectl get events -n flyfleet --sort-by='.lastTimestamp' | tail -20

# Rollback
kubectl rollout undo deployment/fleet-core -n flyfleet
kubectl rollout status deployment/fleet-core -n flyfleet

# Restart
kubectl rollout restart deployment/fleet-core -n flyfleet

# Scale
kubectl scale deployment fleet-core --replicas=5 -n flyfleet
```

### Database (PostgreSQL)
```bash
# Is DB ready?
kubectl exec -it postgres-0 -n flyfleet -- pg_isready

# Connection count
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT count(*) FROM pg_stat_activity;"

# Long-running queries (>5s)
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c \
  "SELECT pid, now() - query_start AS duration, query FROM pg_stat_activity WHERE state='active' AND now() - query_start > interval '5 seconds';"

# Kill query
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT pg_terminate_backend(PID);"
```

### NATS JetStream
```bash
# Server health
kubectl exec -it nats-0 -n flyfleet -- nats server report

# List streams
kubectl exec -it nats-0 -n flyfleet -- nats stream ls

# Consumer status
kubectl exec -it nats-0 -n flyfleet -- \
  nats consumer report TELEMETRY FLEET_CORE_CONSUMER

# Stream info
kubectl exec -it nats-0 -n flyfleet -- nats stream info TELEMETRY
```

### Monitoring
```bash
# Grafana (browser)
https://grafana.flyfleet.com/d/fleet-overview

# Prometheus query (API)
curl -s 'http://prometheus:9090/api/v1/query?query=up' | jq

# Alertmanager
https://alertmanager.flyfleet.com/#/alerts
```

---

## 📚 Runbooks (Which One?)

| Symptom | Runbook |
|---------|---------|
| API 5xx errors, health failing | `api-outage.md` |
| Telemetry lag, stale dashboard | `telemetry-disruption.md` |
| Commands failing, SOS not working | `command-delivery-failure.md` |
| DB slow, queries timing out | `database-performance.md` |
| NATS storage full, consumer lag | `nats-issues.md` |
| Login failing, MFA broken | `auth-failure.md` |
| Data missing, duplicates, stale cache | `data-inconsistency.md` |
| Security breach, malware, DDoS | `security-incident.md` |

**Location:** `infrastructure/incident-response/runbooks/`

---

## 📝 Communication Templates

### SEV1 Initial (to executives)
```
🚨 [SEV1] {Service} - {Brief}

📅 Detected: {time UTC}
🎯 Impact: {X}% customers, ${loss}/hr
👥 IC: @{name} | War Room: #incident-...
📊 Metrics: {key metrics}
🚨 Actions: {current action}
📢 Next: {HH:MM UTC}

Linear: INC-YYYY-NNNN
```

**Send within 15 min to:** Executives, IC, war room

---

### Status Update (Every 15/30 min)
```
[SEV1] {Service} - Update #{n} - {HH:MM UTC}

Status: {Investigating/Mitigating/Monitoring/Resolved}
Progress: {X}% or stage

📊 Metrics:
• Error: {X}% → {Y}% (↑/↓)
• Latency: {X}ms → {Y}ms

🔧 Actions taken:
• {HH:MM} Action → Result
• {HH:MM} Action → Result

🎯 Next:
1. Action (ETA {time})
2. Action (ETA {time})

Next update: {HH:MM UTC}

Linear: INC-YYYY-NNNN
```

**Post in:** War room channel (top-level, not thread)

---

## 🎯 SLO Reference

| Service | SLI | Target | Alert |
|---------|-----|--------|-------|
| API (fleet-core) | Availability | 99.95% (30d) | Burn rate >5x |
| API | P95 Latency | <300ms | >300ms for 10min |
| API | Error Rate | <0.1% | >50% (SEV1), >5% (SEV2) |
| Telemetry | Ingestion Rate | >10,000/sec | <50% for 2min |
| Telemetry | P99 Latency | <5s | >10s for 5min |
| Commands | Success Rate | 99.9% (30d) | Burn rate >2x |
| PostgreSQL | Availability | 99.99% | Down for 1min |
| PostgreSQL | P95 Query | <100ms | >500ms for 10min |
| NATS | Storage | <85% | >85% for 10min |

**Full SLO details:** See OBSERVABILITY_DESIGN.md

---

## 🔄 Incident Timeline (SEV1 Target)

```
T+0   Alert fires
T+2   Acknowledged
T+5   SEV1 declared, war room created, IC assigned
T+15  Executives notified, status page updated
T+30  Root cause category identified
T+60  Mitigation in place OR recovery started
T+90  Service restored (target)
T+120 Resolution declared (after 15min stable)

Total: ~2 hours max (target 90 min)
```

---

## 📊 Success Metrics (What We Track)

| Metric | Target | Why |
|--------|--------|-----|
| MTTD (detection) | <2 min | Alerting works |
| MTTR (SEV1) | <30 min | Fast recovery |
| First update latency | <15 min | Comms discipline |
| PIR completion | 7 days | Learning culture |
| Action items on-time | >80% | Prevention works |
| Repeat incidents | <10% | Fixes prevent recurrence |
| Confidence score | >4.0/5.0 | Team ready |

---

## ✅ Before You Finish Shift

**If you're IC or on-call ending shift:**
- [ ] Resolution verified (15 min stable)
- [ ] Status page updated to "resolved"
- [ ] PagerDuty incident acknowledged/resolved
- [ ] War room archived (`/archive #incident-...`)
- [ ] PIR meeting scheduled (within 24h)
- [ ] Linear ticket closed (linked to PIR)
- [ ] Thank you message to team
- [ ] Handoff to next IC (if continuing)

---

## ❌ Common Mistakes (Avoid!)

| Mistake | Impact |
|---------|--------|
| Wait to declare SEV1 | Delays response, violates SLA |
| No war room created | Fragmented communication |
| IC does technical work | Loses command perspective |
| No note-taker | PIR impossible |
| Late status updates | Execs uninformed, lose trust |
| Restart before logs | Lose evidence for PIR |
| Forget status page | Customers in dark |
| Don't follow runbook | Waste time, miss steps |
| No PIR after SEV1/2 | Repeat same incident |
| Blame individuals | Destroy psychological safety |

---

## 🛠️ Tools Quick Links

**Save these bookmarks:**

| Tool | URL | Purpose |
|------|-----|---------|
| Grafana Overview | `https://grafana.flyfleet.com/d/fleet-overview` | All metrics |
| Grafana API | `https://grafana.flyfleet.com/d/fleet-api` | API health |
| Grafana SLOs | `https://grafana.flyfleet.com/d/fleet-slos` | Error budget |
| Prometheus | `https://prometheus.flyfleet.com` | Query metrics |
| Alertmanager | `https://alertmanager.flyfleet.com` | Manage alerts |
| Linear | `https://linear.app/flyfleet` | Incident tickets |
| Slack #sre-alerts | `slack://channel?team=...&id=sre-alerts` | Alerts |
| Docs | `flyfleet/infrastructure/incident-response/` | Runbooks |

---

## 📚 Framework Files (Know These)

```
infrastructure/incident-response/
├── README.md                          ← Start here
├── FRAMEWORK.md                       ← Full procedures
├── SEVERITY_QUICK_REFERENCE.md       ← Print this!
├── INCIDENT_DECLARATION_CHECKLIST.md ← Use during alert
│
├── runbooks/                          ← CRITICAL - Open during incident
│   ├── api-outage.md
│   ├── telemetry-disruption.md
│   ├── database-performance.md
│   └── ... (8 total)
│
├── templates/communication/
│   ├── initial-notification-sev1.md
│   ├── status-update.md
│   └── customer-communication.md
│
└── frameworks/
    ├── pir-template.md               ← After resolution
    ├── command-center-procedures.md
    └── slo-integration-guide.md
```

---

## 🆘 When Stuck

**During incident, if unsure:**
1. Ask in war room: "What should I check next?"
2. Open runbook, read decision tree
3. Call SME: @dba-team, @nats-team, etc.
4. Escalate to backup IC or VP Eng (if >60 min no progress)

**Remember:**
- It's ok to ask for help
- Better to ask than guess
- Framework is guide, not straitjacket
- Use best judgment, document rationale

---

## 📞 Quick Phone Tree (If Slack Down)

```
1. On-call acknowledges via PagerDuty (still works?)
2. IC calls VP Eng if SEV1
3. IC calls Technical Lead
4. IC calls key SMEs
5. Use conference bridge: +1-XXX-XXX-XXXX
```

---

## 🎯 Golden Questions (Every Update)

IC must answer these every 15/30 min:

1. **Customer impact right now?**
   (% affected, symptoms, regions)

2. **What are we doing to fix it?**
   (Current action, ETA, confidence level)

3. **What could go wrong?**
   (Risks, blockers, worst-case)

---

## Takeaways

✅ **Declare SEV1 fast** (can downgrade later)
✅ **Create war room** (single channel)
✅ **Follow runbook** (saves time)
✅ **Update on schedule** (every 15/30 min)
✅ **Preserve logs** (for PIR)
✅ **Complete PIR** (learn and improve)

---

**Quick Ref Card v1.0** | FlyFleet Incident Response
Last Updated: 2026-03-29 | Print and keep at desk!
