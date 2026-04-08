# Command Center Procedures (War Room Operations)

**Purpose:** Standardized operation of incident war room for SEV1/SEV2 incidents
**Audience:** Incident Commanders, Technical Leads, Note-Takers, SMEs

---

## War Room Setup

### Slack Channel Creation

**Automated (preferred):** Use PagerDuty/Linear integration to auto-create `#incident-INC-YYYY-NNNN`

**Manual (if needed):**
```bash
# Create channel
/slack create incident-INC-2026-0123

# Set purpose
/slack set-purpose "SEV1 API Outage - IC: @sarah-chen | Linear: INC-2026-0123 | Bridge: zoom.us/j/123"

# Invite core team
/invite @sarah-chen @marcus-j @dba-team @infra-sre @priya-sharma

# Set topic
/topic IC: Sarah Chen | Technical Lead: Marcus Johnson | Updates every 15 min
```

**Channel naming:** `#incident-{ticket-number}` (consistent)

---

### Video Bridge

**Setup:** Create Zoom/Teams bridge quickly:

```
Zoom:
https://zoom.us/j/[meeting-id]?pwd=[password]
- Set waiting room OFF (for rapid join)
- Mute participants on entry
- Record for evidence (if legal approves)

Teams:
https://teams.microsoft.com/l/meetup-join/[meeting-id]
- Presenter mode for screen sharing
- Chat enabled for parallel discussion
```

**Bridge link** must be in channel topic and initial notification.

---

### Roles & Permissions

| Role | Permissions | Responsibilities |
|------|-------------|------------------|
| **Incident Commander** | Can post to channel, make decisions, assign tasks | Overall command, timeline, escalation |
| **Technical Lead** | Can post, direct technical actions | Investigation, recovery, validation |
| **SMEs** | Can post replies, ask questions | Provide expertise, execute commands |
| **Communications Lead** | Can post updates to #announcements (with IC approval) | Stakeholder/customer comms |
| **Note-Taker** | Read channel, edit pinned timeline doc | Document timeline, decisions |
| **Observers** | Read-only (mute channel) | Executives, other teams (not responders) |

**Slack channel settings:**
- Allow all members to post (IC manages)
- Grant `@incident` mention power (urgent attention)
- Pin timeline document and key links
- Enable thread mode for organized discussion

---

## Communication Protocols

### Channel Discipline

**DO:**
- Use threads for side discussions (keep main channel clean)
- Use `@incident` for urgent (attention needed immediately)
- Use `@here` sparingly (only all-responder needed)
- Tag people when asking them to do something: `@dba-team check DB connections`
- Post status update questions in dedicated update thread

**DON'T:**
- Don't create multiple incident channels (keep all in one)
- Don't use `@channel` (spams everyone incl. observers)
- Don't ask "are we good yet?" in main channel (check pinned timeline)
- Don't speculate without evidence (tag IC if uncertain)

---

### Status Update Cadence (SEV1)

```
IC posts update at top of channel (not in thread):

[SEV1] API Outage - Update #4 - 14:50 UTC

Status: Mitigating | Progress: 70% | ETA: 14:55 UTC

📊 Metrics:
• Error rate: 15% ↓ from 40%
• Latency P95: 800ms ↓ from 2500ms

🔄 Actions taken since last update:
• 14:45: Rollback complete, 4/4 pods healthy
• 14:48: Health checks 200 on all pods
• 14:49: Traffic shifting to 50% healthy pods

🎯 Next steps:
1. Complete traffic ramp to 100% (ETA 14:55)
2. Monitor error rate to <1%

Next update: 15:05 UTC

Linear: INC-2026-0123
Grafana: https://grafana...
```

**Why top-level (not thread)?**
- Pinned for quick reference
- Easy to scan for newcomers
- Export to exec summaries

---

### Thread Usage

**Thread main update** for questions/clarifications:

```
[SEV1] API Outage - Update #4 - 14:50 UTC
│
├─ @jane-exec: "What's customer impact estimate?"
│  → @IC-sarah: ~100% customers, partial degradation for most
│
├─ @dba-team: "DB connections good?"
│  → @marcus-j: Yes, 12/100 active now
│
└─ Observers: Read-only, no need to reply
```

---

### Escalation via Slack

**Paging backup:**
```
@backup-ic-sre Please join war room, primary IC needing handoff at 18:00 UTC

Reason: IC shift change
Linear: INC-2026-0124
Status: Investigating, no resolution yet
```

**Escalation to VP:**
```
@vp-engineering Escalating for resource decision:

We need to expand NATS storage immediately (critical). Current storage 95% full, ingestion lag increasing. Proposal: scale PVC from 50Gi to 100Gi.

Can you approve $500 emergency storage扩容 (one-time cost)?

War room: #incident-INC-2026-0124
```

---

## Note-Taking & Timeline

### Real-Time Timeline Document

**Create immediately** (Google Doc or markdown file):

```markdown
# Incident Timeline: INC-2026-0124

**IC:** Sarah Chen
**Note-Taker:** Wei Zhang
**War Room:** #incident-INC-2026-0124

---

## Chronological Log

All times UTC

### 2026-03-29

**14:30:12** [Alert] Prometheus "ApiHighErrorRate" fired
- Error rate 52% threshold 50% for 2m
- Sent to PagerDuty, on-call (@jameswong-sre)

**14:31:00** [Human] On-call acknowledged PagerDuty
- Acknowledged by: James Wong
- Comment: "Investigating"

**14:31:45** [Decision] IC declared SEV1
- Declared by: James Wong (on-call acting as IC)
- Linear ticket created: FF-1234 / INC-2026-0124

**14:32:00** [Action] War room created
- Slack: #incident-INC-2026-0124
- Bridge: https://zoom.us/j/123456
- Invited: @marcus-j, @dba-team, @infra-sre, @priya-sharma

[... continue chronologic]

---

## Summary of Key Events

| Time | Event | Owner |
|------|-------|-------|
| 14:32 | IC declared | James |
| 14:45 | Root cause identified | Marcus |
| 14:50 | Rollback complete | Infra |
| 15:00 | Resolved | Sarah |

## Decision Log

| Time | Decision | Rationale | Outcome |
|------|----------|-----------|---------|
| 14:40 | Rollback approved | Deploy suspected, fast recovery path | Successful |
| 14:50 | Extend IC shift | IC fatigue, needed continuity | Handoff at 18:00 |

## Evidence Links

- Grafana dashboard: https://grafana...
- Prometheus query: https://prometheus...
- Linear ticket: https://linear...
- Slack export: (saved separately)
```

---

## Handoff Protocol

### IC Shift Handoff (for long incidents > 4h)

**When:** IC shift change or IC needs to hand off

**Process:**
1. IC briefs incoming IC (15 min sync)
2. Review timeline, current status, next steps
3. Transfer decision authority formally
4. Announce in war room:
```
@incident Handing off IC duties from @sarah-chen to @patricia-sre at 18:00 UTC
@patricia-sre is now Incident Commander.
All future decisions and escalation through Patricia.
Thanks Sarah for 4h of leadership!
```

5. Outgoing IC available as advisor for 1 hour, then off

**Documentation:**
- Update Linear ticket with IC change
- Note handoff timestamp and reason in timeline

---

### End of Incident

**IC declares resolution:**
```
[SEV1] RESOLVED - Fleet-core API restored at 15:12 UTC

All systems healthy, error rate <0.1% for 15 min.

War room closing now. PIR scheduled for Monday 14:00 UTC.

Thank you all for outstanding response.
```

**Close war room:**
- Archive channel (`/archive #incident-...`)
- Post summary in `#engineering` and `#all-company`
- Send resolution email to execs

---

## Evidence Preservation

### What to Preserve

- ✅ Slack war room channel (export)
- ✅ Timeline document (Google Doc history)
- ✅ Linear ticket (comments, status changes)
- ✅ Grafana dashboard snapshots (export JSON)
- ✅ Prometheus queries (save query + result)
- ✅ Application logs (from affected pods, 24h before to 1h after)
- ✅ Database logs (same window)
- ✅ Deployment history (kubectl rollout history)
- ✅ Configuration at time of incident (git commit SHAs, Helm values)
- ✅ AWS CloudTrail events (if infrastructure change involved)

**How to preserve:**
```
# Create evidence package
mkdir -p /evidence/INC-2026-0124/{slack,logs,metrics,config}
# Copy files...
# Create manifest with hashes
find . -type f -exec sha256sum {} \; > SHA256SUMS
```

**Secure storage:**
- Encrypt if contains PII: `gpg -c evidence.tar.gz`
- Upload to S3 with restricted IAM: `aws s3 cp --sse aws:kms`
- Store in compliance vault if legal hold required

---

## Observer Guidelines

**For executives and non-responders:**

- `#incident-*` channels are for responders only. Do NOT:
  - Ask "what's the status?" (read pinned updates)
  - Suggest random fixes (unless asked by IC)
  - @mention responders ("paging someone...")
- If you need information, IC or Comms Lead will provide it
- Trust the IC to make decisions; support their choices
- Defer all customer/media questions to Comms Lead

**Observer channels:** Use `#incident-observers` for questions (read/write)

---

## Crisis Communication (Inside War Room)

### What to Communicate

| Channel | Frequency | Audience | Owner |
|---------|-----------|----------|-------|
| War room updates | Every 15 min (SEV1) | Responders | IC |
| Executive Slack | Every 15 min (SEV1) | Executives | IC |
| Customer Support | Every 30 min | Support team | Comms Lead |
| Status page | Every 30 min | Customers | Comms Lead |
| All-company | At milestones | All employees | Comms Lead |

---

### Information Sharing

**What to share inside war room:**
- Current metrics (Grafana screenshot)
- Hypothesis updates
- Actions taken and outcomes
- Questions for SMEs
- Escalation decisions
- Customer impact reports

**What NOT to share:**
- Speculation without evidence
- Blame or criticism of team members
- External communications (PR draft) until approved
- PII/PCI data in logs (mask before sharing)

---

## Tools Quick Reference

### Commands Run Often

```bash
# Check pod status
kubectl get pods -n flyfleet -o wide --sort-by='.status.phase'

# Tail logs
kubectl logs -f deployment/fleet-core -n flyfleet --tail=100

# Check events
kubectl get events -n flyfleet --sort-by='.lastTimestamp' | tail -20

# Resource usage
kubectl top pods -n flyfleet

# Rollback deployment
kubectl rollout undo deployment/fleet-core -n flyfleet
kubectl rollout status deployment/fleet-core -n flyfleet

# Execute DB query
kubectl exec -it postgres-0 -n flyfleet -- psql -U fleet_user fleet_db -c "SELECT 1"

# Grafana direct link
https://grafana.flyfleet.com/d/fleet-api?viewPanel=error-rate&time=now-30m

# Prometheus direct query
'http://prometheus:9090/api/v1/query?query=rate(http_requests_total[5m])'

# Check external status
https://status.aws.amazon.com/  # AWS
https://status.nats.io/         # NATS
```

---

## Common Mistakes to Avoid

| Mistake | Consequence | How to Avoid |
|---------|-------------|--------------|
| **IC does technical work** | Loses command perspective | IC focuses on coordination, not typing |
| **Multiple war rooms** | Information fragmented | One channel only |
| **No note-taker** | Timeline lost, PIR impossible | Assign at declaration |
| **Late comms sync** | Executives/customers uninformed | Comms lead added immediately |
| **Debating publicly** | Wastes time, unnerves observers | Use threads or move to side channel |
| **No handoffs** | IC burnout, no knowledge transfer | Schedule shifts for >4h incidents |
| **Destroying logs** | No evidence for PIR | Preserve everything until PIR approved |
| **Speculation as fact** | Misleads response | Label hypothesis clearly: "HYPOTHESIS: ..." |

---

## Emergency Power-Off Procedures

If incident involves security breach requiring isolation:

```
🔒 LOCKDOWN PROCEDURE (SEV1-Security)

1. IC: @security-officer (mandatory) assumes lead
2. Isolate compromised systems via NetworkPolicy (see security-incident.md)
3. Preserve all logs (do NOT purge/kill until forensics)
4. Legal notified (preserve attorney-client privilege)
5. AWS support engaged (if need to preserve S3/EC2)
6. No system destruction until approved by Legal + Security
```

---

## Post-Incident War Room Closure

**End-of-incident checklist:**

- [ ] Resolution verified (stable for 15+ min)
- [ ] Status page updated to "resolved"
- [ ] Customer support final communication sent
- [ ] All responders thanked in war room
- [ ] War room archived (`/archive #incident-...`)
- [ ] PIR scheduled (within 24h) and invite sent
- [ ] Evidence package collected and stored securely
- [ ] Linear ticket status set to "resolved", linked to PIR doc
- [ ] Follow-up actions logged in Linear project
- [ ] Executive summary drafted by IC (for leadership meeting)

---

## War Room Templates Summary

| Template | Location | Use |
|----------|----------|-----|
| Initial notification (SEV1) | `templates/communication/initial-notification-sev1.md` | Notify execs |
| Status update | `templates/communication/status-update.md` | Every 15/30 min |
| Timeline doc | `frameworks/command-center-procedures.md` (this) | Real-time log |
| Handoff announcement | `templates/communication/handoff.md` | IC shift change |
| Resolution announcement | `templates/communication/resolution-announcement.md` | All-clear |

---

## Support Channels

| Need | Contact |
|------|---------|
| War room tech support (Slack issues) | @slack-admin |
| Zoom/Teams support | @it-support |
| Escalate IC decisions | VP Engineering |
| Legal hold questions | @legal |
| Compliance concerns | @compliance-officer |

---

**Procedures Version:** 1.0.0
**Last Updated:** 2026-03-29
