# FlyFleet Incident Response

**Complete incident management framework** for production-ready fleet management platform.

---

## Quick Start

### Immediate (During Incident)

1. **Declare severity** using [SEVERITY_QUICK_REFERENCE.md](SEVERITY_QUICK_REFERENCE.md)
2. **Create war room** and run appropriate **runbook**:
   - API issues → `runbooks/api-outage.md`
   - Telemetry issues → `runbooks/telemetry-disruption.md`
   - Database issues → `runbooks/database-performance.md`
   - Command issues → `runbooks/command-delivery-failure.md`
   - NATS issues → `runbooks/nats-issues.md`
   - Auth issues → `runbooks/auth-failure.md`
   - Data issues → `runbooks/data-inconsistency.md`
   - Security → `runbooks/security-incident.md`
3. Send initial notification:
   - SEV1 → `templates/communication/initial-notification-sev1.md`
   - SEV2 → `templates/communication/initial-notification-sev2.md`
4. Update every 15/30 min using `templates/communication/status-update.md`
5. When resolved, use `templates/communication/resolution-announcement.md`
6. Complete PIR using `frameworks/pir-template.md`

---

## Directory Structure

```
infrastructure/incident-response/
│
├── README.md                          # This file - overview
├── FRAMEWORK.md                       # Complete incident response framework
├── SEVERITY_QUICK_REFERENCE.md        # One-page severity decision guide
│
├── runbooks/                          # 🚨 CRITICAL - Service-specific response procedures
│   ├── api-outage.md                  # Fleet-core API failures
│   ├── telemetry-disruption.md        # Telemetry ingestion issues
│   ├── command-delivery-failure.md    # Device command issues
│   ├── database-performance.md        # PostgreSQL performance
│   ├── nats-issues.md                 # NATS JetStream problems
│   ├── auth-failure.md                # Authentication failures
│   ├── data-inconsistency.md          # Data sync/corruption
│   └── security-incident.md           # Security breaches
│
├── templates/communication/           # 📧 Communication templates
│   ├── initial-notification-sev1.md  # Executive alert (15min)
│   ├── initial-notification-sev2.md  # Engineering alert (30min)
│   ├── status-update.md              # Reusable update template
│   └── customer-communication.md     # Status page, email templates
│
├── frameworks/                        # 📐 Analysis frameworks
│   ├── pir-template.md               # Post-Incident Review template
│   ├── command-center-procedures.md  # War room operations
│   ├── slo-integration-guide.md      # SLO-based incident detection
│   └── rcamethods/                   # Root Cause Analysis methods
│       ├── 5-whys.md                 # 5 Whys technique
│       ├── fishbone.md               # Fishbone/Ishikawa diagram
│       └── timeline-rca.md           # Timeline reconstruction
│
└── playbooks/                        # 🎯 (Coming soon)
    ├── deployment-rollback.md
    ├── database-failover.md
    └── disaster-recovery.md
```

---

## On-Call Quick Reference

### During Incident

**First 5 minutes:**
1. Acknowledge PagerDuty alert ✅
2. Check Grafana dashboard 📊
3. Declare SEV1/SEV2 using severity matrix (SEVERITY_QUICK_REFERENCE.md)
4. Create Linear ticket `INC-YYYY-NNNN`
5. Open war room: `#incident-INC-YYYY-NNNN`
6. Page IC if you're not IC

**First 15 minutes:**
- [ ] Notify executives (SEV1 only) using initial-notification template
- [ ] Update status page (SEV1/2)
- [ ] Start investigation using relevant runbook
- [ ] Appoint note-taker

**First 30 minutes:**
- [ ] Root cause category identified
- [ ] Mitigation path clear OR recovery in progress
- [ ] First status update sent

---

## Severity Cheat Sheet

| Severity | Definition | Response Time | Update Freq |
|----------|------------|---------------|-------------|
| **SEV1** | Complete outage or safety risk | 0-5 min page | Every 15 min |
| **SEV2** | Major degradation (>25% users) | 15 min | Every 30 min |
| **SEV3** | Minor impact, workaround exists | 2 hours | Milestones |
| **SEV4** | Cosmetic, no user impact | 1-2 business days | Standard tickets |

**Full criteria:** See SEVERITY_QUICK_REFERENCE.md

---

## SLO-Based Detection

FlyFleet uses the following SLOs (from OBSERVABILITY_DESIGN.md):

- **API**: 99.95% availability, P95 <300ms, error rate <0.1%
- **Telemetry**: >10k events/sec, P99 <5s
- **Commands**: 99.9% success rate, P95 <2s
- **Database**: 99.99% availability, P95 <100ms
- **NATS**: Storage <85%, in/out balanced

All alerts fire from these SLO violations via Prometheus multi-window burn rates.

---

## Post-Incident Process

1. **Within 24h**: Schedule PIR meeting
2. **Within 48h**: Draft PIR using `frameworks/pir-template.md`
   - Complete timeline
   - Root cause analysis (5 Whys, Fishbone, or Timeline)
   - Impact assessment
   - What went well / didn't go well
   - Action items with owners and due dates
3. **Within 5 days**: Hold PIR review meeting
   - IC, Technical Lead, SMEs, stakeholders
   - Review action items
   - Approve final PIR
4. **Within 7 days**: Final PIR published
5. **Track action items** in Linear project until complete

---

## Essential Commands

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

### Database
```bash
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT 1"

# Connection count
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT count(*) FROM pg_stat_activity;"

# Long queries
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT pid, query FROM pg_stat_activity WHERE state='active' AND now() - query_start > interval '5 seconds';"
```

### NATS
```bash
kubectl exec -it nats-0 -n flyfleet -- nats server report
kubectl exec -it nats-0 -n flyfleet -- nats stream ls
kubectl exec -it nats-0 -n flyfleet -- nats consumer report TELEMETRY FLEET_CORE_CONSUMER
```

### Monitoring
```bash
# Grafana direct (use web UI)
https://grafana.flyfleet.com/d/fleet-overview

# Prometheus query
curl -s 'http://prometheus:9090/api/v1/query?query=up' | jq

# Alertmanager
https://alertmanager.flyfleet.com/#/alerts
```

---

## Emergency Contacts

| Role | Contact | Slack |
|------|---------|-------|
| On-call SRE | PagerDuty | @sre-oncall |
| Database Admin | +1-XXX-XXX-XXXX | @dba-team |
| Security Officer | +1-XXX-XXX-XXXX | @security-oncall |
| VP Engineering | +1-XXX-XXX-XXXX | @vp-engineering |
| Legal | +1-XXX-XXX-XXXX | @legal |
| AWS Support | 1-888-825-4375 | - |
| Customer Support | support@flyfleet.com | @support-team |

---

## Training & Drills

### Required Reading (before on-call)

- [ ] FRAMEWORK.md (complete)
- [ ] SEVERITY_QUICK_REFERENCE.md
- [ ] All assigned service runbooks
- [ ] command-center-procedures.md (IC role only)
- [ ] pir-template.md (IC role only)

### Quarterly Drills

- SEV1 simulation: API complete outage
- SEV2 simulation: Telemetry lag
- Security incident: Data breach response

---

## Templates Index

| Template | Location | Use Case |
|----------|----------|----------|
| Initial SEV1 notification | `templates/communication/initial-notification-sev1.md` | Exec alert within 15min |
| Initial SEV2 notification | `templates/communication/initial-notification-sev2.md` | Eng leadership within 30min |
| Status update | `templates/communication/status-update.md` | Every 15/30 min |
| Customer communication | `templates/communication/customer-communication.md` | Status page, email |
| Resolution announcement | `templates/communication/resolution-announcement.md` | All-clear |
| PIR template | `frameworks/pir-template.md` | Post-incident review |

---

## Filing Guidelines

### File Naming

- Runbooks: lowercase with hyphens, descriptive: `api-outage.md`
- Templates: include purpose and severity: `initial-notification-sev1.md`
- Frameworks: descriptive name: `command-center-procedures.md`

### Headings

Use ATX style (`#`) for headings. Don't skip levels.

### Code Blocks

```markdown
```bash
command here
```
```

### Tables

Use GitHub-flavored markdown tables:

```markdown
| Header 1 | Header 2 |
|---------|----------|
| cell 1   | cell 2   |
```

---

## Contributing

To update incident response docs:

1. Edit relevant file
2. Test commands in local FlyFleet environment if applicable
3. Ensure templates are fillable without modification
4. Submit PR with `[incident-response]` prefix in title
5. PR requires review from:
   - Platform Engineering lead
   - SRE team member
   - (Optional) Security team for security docs

---

## Support

**Questions about incident response:**
- Slack: `#sre-team` or `#platform-engineering`
- Docs owner: Platform Engineering Team
- Last framework update: 2026-03-29

**Found an issue in docs?** Open issue in GitHub with label `documentation` and component `incident-response`.

---

**Framework Version:** 1.0.0
**Last Updated:** 2026-03-29
**Maintained by:** Platform Engineering & SRE Teams
