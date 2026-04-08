# FlyFleet Incident Severity - Quick Reference Card

**Print this page for on-call desk. Keep visible during incident response.**

---

## Decision Matrix

| Criteria | SEV1 Critical | SEV2 Major | SEV3 Minor | SEV4 Low |
|----------|---------------|------------|------------|----------|
| **User Impact** | 100% or safety | >25% users | <25% users | Single user/feature |
| **Business Impact** | >$10k/hr or safety risk | $1k-$10k/hr | < $1k/hr | None/minimal |
| **Service Status** | Complete outage | Degraded | Partial with workaround | Cosmetic/bug |
| **Response Time** | 0-5 min (page) | 15 min | 2 hours (biz hours) | 1-2 business days |
| **IC Assignment** | 5 min | 30 min | 2 hours | Next day |
| **Exec Notification** | 15 min | 1 hour | EOD | Never |
| **Status Page** | 15 min | 30 min | Optional | None |
| **Update Cadence** | Every 15 min | Every 30 min | Milestones only | Standard tickets |
| **War Room** | Yes (dedicated) | Yes | Optional call bridge | No |
| **All Hands** | Yes if needed | No | Never | Never |

---

## SEV1 - Critical Outage

### Thresholds

| Service | Condition | Threshold |
|---------|-----------|-----------|
| API (fleet-core) | Availability | < 99.9% (30 min) |
| API | Error rate | > 50% (5 min) |
| API | Latency P95 | > 5s (10 min) |
| Telemetry | Ingestion rate | 0 events/sec (2 min) |
| Telemetry | Lag | > 30 min (5 min) |
| Commands | Success rate | < 50% (5 min) |
| Database | Availability | Down (1 min) |
| ALL | Customer safety (SOS) | ANY failure = SEV1 |

### Actions (0-15 min checklist)

- [ ] Acknowledge PagerDuty within 5 min
- [ ] Declare SEV1, create Linear ticket
- [ ] Assemble IC team (self or delegate)
- [ ] Create war room `#incident-INC-YYYY-NNNN`
- [ ] Page executives within 15 min
- [ ] Update status page within 15 min
- [ ] Start 15-min update cadence

---

## SEV2 - Major Impact

### Thresholds

| Service | Condition | Threshold |
|---------|-----------|-----------|
| API | Error rate | > 5% (15 min) |
| API | Latency P95 | > 2s (30 min) |
| Telemetry | Ingestion rate | < 50% of baseline (15 min) |
| Telemetry | Lag | > 10 min (15 min) |
| Commands | Success rate | < 95% (30 min) |
| Database | Queries slow | P95 > 500ms (30min) |
| Dashboard | Loading errors | > 25% users affected |

### Actions (0-30 min checklist)

- [ ] Acknowledge PagerDuty within 15 min
- [ ] Declare SEV2, create Linear ticket
- [ ] Assign IC within 30 min
- [ ] Create war room channel
- [ ] Notify engineering leadership within 30 min
- [ ] Update status page within 30 min (if customer-facing)
- [ ] Start 30-min update cadence

---

## SEV3 - Minor Impact

### Examples

- Single feature broken (e.g., report export fails)
- <25% users affected (tenant-specific)
- Non-critical internal tool down
- Performance degradation (<2s but >1s)
- Monitoring false positive investigation

### Actions

- Acknowledge within 2 hours (business hours)
- Next business day response acceptable off-hours
- Internal team notification
- No war room required (use ticket comments)
- Status page update optional

---

## SEV4 - Low Impact

### Examples

- Cosmetic UI bug (alignment, typo)
- Documentation error
- Logging gap
- Dev environment down
- Feature request

### Actions

- Response within 1-2 business days
- Standard ticket workflow
- No escalation

---

## Severity Estimation Cheat Sheet

### Ask These Questions:

1. **User Impact**: What % of customers affected?
   - 100% → SEV1
   - 25-99% → SEV2
   - <25% → SEV3

2. **Safety/Dollars**: Any safety risk or revenue loss > $10k/hr?
   - Yes → SEV1
   - No → Next question

3. **Service Status**: Complete outage or degraded?
   - Complete outage → SEV1
   - Severe degradation → SEV2
   - Minor degradation → SEV3

4. **Workaround**: Can customers work around issue?
   - No workaround → downgrade to SEV1/2
   - Workaround exists → SEV3

5. **SLA Impact**: Will we violate SLA?
   - Yes (critical SLA) → SEV1
   - Yes (minor SLA) → SEV2
   - No → SEV3 or SEV4

---

## Red Flags → Always SEV1

- ✅ Emergency SOS not working
- ✅ Payment processing failure
- ✅ Data loss or corruption detected
- ✅ Security breach (PII exposure)
- ✅ Database completely unavailable
- ✅ All replicas down
- ✅ Compliance violation (FASTag/E-Way Bill)
- ✅ SLA penalty > $10k at risk
- ✅ Customer safety at risk

---

## Quick Commands (Memorize)

```bash
# Health check
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health

# Pod status
kubectl get pods -n flyfleet -l app=fleet-core

# Rollback deployment
kubectl rollout undo deployment/fleet-core -n flyfleet

# Restart service
kubectl rollout restart deployment/fleet-core -n flyfleet

# Scale up (emergency)
kubectl scale deployment fleet-core --replicas=5 -n flyfleet

# Check DB
kubectl exec -it postgres-0 -n flyfleet -- pg_isready

# Check NATS
kubectl exec -it nats-0 -n flyfleet -- nats server report

# Check logs (last 5 min)
kubectl logs -f deployment/fleet-core -n flyfleet --since=5m

# Stream logs to file
kubectl logs -f deployment/fleet-core -n flyfleet --since=5m > /tmp/fleet-core.log
```

---

## Who to Page

| Role | PagerDuty | Slack | When |
|------|-----------|-------|------|
| On-call SRE | Auto | @sre-oncall | All incidents |
| Database Admin | @dba-oncall | @dba-team | DB issues, high connections |
| NATS Expert | @nats-sme | @nats-team | Messaging issues |
| Security Officer | @security-oncall | @security-team | PII breach, auth issues |
| VP Engineering | Manual | @vp-engineer | SEV1 or escalation |
| Infra SRE | @infra-sre | @infra-team | K8s, network issues |

---

## Communication Cadence

| Severity | Stakeholder | Timing | Method |
|----------|-------------|--------|--------|
| SEV1 | Executives | 15 min | Slack + SMS |
| SEV1 | Customer Support | Real-time | Slack |
| SEV1 | Customers | 15 min | Status page |
| SEV2 | Engineering Leads | 30 min | Slack |
| SEV2 | Product | 1 hour | Email |

---

## Common Initial Mistakes

❌ **Don't:**
- Investigate for >5 min before declaring SEV1
- Restart services before investigating (may lose evidence)
- Use `kubectl delete pod` (use `rollout restart`)
- Assume it's NATS/DB without checking metrics first
- Forget to create war room

✅ **Do:**
- Declare SEV1 if unsure (can downgrade)
- Preserve logs first (`kubectl logs --since=1h > file`)
- Use runbooks (they exist for a reason)
- Page backup IC if shift > 4 hours
- Update status page promptly

---

## SLO-Based Severity

Use SLO burn rate to auto-classify:

```
Burn rate > 5x (fast) → SEV1
Burn rate 2-5x → SEV2
SLO violation but low burn → SEV3
```

See `OBSERVABILITY_DESIGN.md` for SLO definitions and alert rules.

---

## Severity Change Protocol

**Downgrading:**
- IC decision, document reason in timeline
- Update war room channel topic
- Send notification to stakeholders: "SEV1 → SEV2 at 15:30 UTC due to ..."
- Adjust update cadence accordingly

**Upgrading:**
- Any responder can suggest, IC decides
- If impact increases or no progress after timeout
- Notify additional stakeholders immediately
- Increase communication frequency

---

## Three Golden Questions

Every 15/30 minutes, IC must answer:

1. **What's the customer impact right now?**
   - % affected, symptoms, regions

2. **What are we doing to fix it?**
   - Current action, ETA, confidence level

3. **What could go wrong?**
   - Risks, blockers, worst-case scenario

---

## Emergency Contact Summary

| Who | Contact | When |
|-----|---------|------|
| On-call SRE | PagerDuty | All incidents |
| IC backup | @sre-backup | IC unavailable |
| Database | @dba-team | DB issues |
| Security | @security-oncall | Data/access incidents |
| VP Engineering | +1-xxx-xxx-xxxx | SEV1, no IC |
| Legal | +1-xxx-xxx-xxxx | Data breach, compliance |
| AWS Support | 1-888-825-4375 | Infrastructure |
| Customer Support | support@flyfleet.com | Customer inquiries |

---

**Quick Ref Version:** 1.0.0
**Last Updated:** 2026-03-29
