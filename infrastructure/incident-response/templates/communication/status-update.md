# Status Update Template (Generic)

**Use:** Reusable template for any incident update (SEV1/SEV2/SEV3)
**Frequency:** Every 15 min (SEV1), 30 min (SEV2), at milestones (SEV3)

---

```
[{SEV}] {Service} - Status Update #{UpdateNumber}

---
📅 Update Time: {YYYY-MM-DD HH:MM UTC}

---
📊 CURRENT STATUS
• Status: {Investigating / Mitigating / Monitoring / Resolved}
• Progress: {percentage or stage, e.g., "60% - rollback complete, testing in progress"}
• ETA to Resolution: {Specific time or "Assessing"}

---
📉 IMPACT UPDATE
• Scope: {Updated affected users/customers, e.g., "Reduced to 10% of users"}
• Severity Change: {Yes/No} {if changed: "Downgraded to SEV2 at 10:30 UTC"}
• New Symptoms: {Any new issues discovered}
• Resolved Symptoms: {Any issues that have been fixed}

---
🔧 ACTIONS TAKEN SINCE LAST UPDATE
{List actions with timestamps}
• {HH:MM} {Action taken} → {Outcome}
  - {Details: commands run, configs changed, resources scaled}
  - Result: {Success / Partial / Failed}
  - Next: {Follow-up action if any}

Example:
• 10:45 UTC → Rolled back fleet-core to v1.23.4
  - `kubectl rollout undo deployment/fleet-core`
  - Result: Success, pods coming up
  - Next: Wait 5 min for health checks

---
📈 METRICS (Compared to Last Update)
Before: {Previous metrics}
Now: {Current metrics}

• Error Rate: {X}% → {Y}% ({direction})
• Latency P95: {X}ms → {Y}ms
• Throughput: {X} rps → {Y} rps
• Active Users: {X} → {Y}

---
🔍 INVESTIGATION STATUS
• Root Cause: {Confirmed / Likely / Hypothesis / Unknown}
• If confirmed: {brief one-sentence root cause}

Evidence gathered:
- ✅ {item 1 checked, e.g., "Database connectivity verified"}
- ❓ {item 2 pending, e.g., "Analyzing slow query logs"}
- ⏳ {item 3 pending, e.g., "Waiting for AWS support response"}

Key findings since last update:
{bullet points of new information}

---
👥 TEAM STATUS
• IC: {Name} (active since {time})
• Technical Lead: {Name}
• SMEs Current: {Who is actively working}
• SMEs On Deck: {Who is on standby}

Any team member fatigue?
- {Name}: {status, e.g., "12h shift, considering handoff"}
- No fatigue reported

---
📞 ESCALATION DECISION
Escalate? {Yes/No/Pending}
Reason: {if Yes: "No progress after 2h", "Impact increased", etc.}

If escalating:
- Target Severity: {SEV1}
- Justification: {brief}
- Notify: {executives, legal, etc.}

---
💬 COMMUNICATIONS
• Status Page: {Updated / Pending} ({link})
• Customer Support: {Notified at {time}, ticket volume {low/med/high}}
• Executive Notification: {Sent at {time} / Not yet / Not required}
• External Comms: {None yet / PR engaged / Regulators notified}

---
⚠️ RISKS & BLOCKERS
Risks:
- {Risk 1} - {Mitigation planned}
- {Risk 2} - {Escalation trigger if ...}

Blockers:
- {Blocker 1} - {Owner working to resolve}
- {Blocker 2} - {ETA resolution}

Example:
- Risk: Rollback might not fix issue if root cause infrastructure → Have fix-forward plan ready
- Blocker: Waiting on AWS support case response (opened 11:00 UTC) - ETA 1h

---
🎯 NEXT ACTIONS (Next 30 min)
1. {Action 1} - {Owner} (by {time})
2. {Action 2} - {Owner} (by {time})
3. {Action 3} - {Owner} (by {time})

---
📅 NEXT UPDATE
• Scheduled: {YYYY-MM-DD HH:MM UTC}
• Unless: {event that would trigger out-of-cycle update}
• Channel: {This thread / #incident-...}

---
🔗 LINKS
• Linear Ticket: {Link}
• Grafana Dashboard: {Link}
• Runbook: {Link to relevant runbook}
• War Room: {Slack channel link}
• AWS Support Case: {if opened}
• Google Doc Timeline: {if using shared doc}

---
---
Update Template v1.0 | IC: {Name} | Incident: INC-YYYY-NNNN
```

---

## Usage Instructions by Severity

### SEV1 (Every 15 min)
- Keep updates concise, focus on actions taken and new info
- Always include ETA (update if changes)
- Highlight risks/blockers prominently
- Use "Progress:" field to show % or stage

### SEV2 (Every 30 min)
- Similar format, slightly less detail acceptable
- Clear escalation path if not improving
- Customer impact focus

### SEV3 (At milestones only)
- Update at: Investigation complete, mitigation deployed, resolution verified
- Can use abbreviated version:
  ```
  [SEV3] Issue resolved
  - Problem: {brief}
  - Fix: {applied}
  - Impact: {minimal, no customer reports}
  - Next: Post-incident review scheduled
  ```

---

## Example SEV1 Update (Context: API outage, rollback in progress)

```
[SEV1] Fleet Core API - Status Update #3

---
📅 Update Time: 2026-03-29 14:50 UTC

---
📊 CURRENT STATUS
• Status: Mitigating
• Progress: 70% - Rollback complete, health checks in progress
• ETA to Resolution: 14:55 UTC (5 min)

---
📉 IMPACT UPDATE
• Scope: Still 100% customers affected
• Severity Change: No (remains SEV1)
• New Symptoms: None
• Resolved Symptoms: Pods restarting without errors

---
🔧 ACTIONS TAKEN SINCE LAST UPDATE
• 14:45 UTC → Completed fleet-core rollback to v1.23.4
  - `kubectl rollout status` confirmed all 4 replicas ready
  - Result: Success
• 14:48 UTC → Health checks passing on all pods
  - `/health` endpoint returns 200 on all pods
  - Database connectivity verified
• 14:49 UTC → Started traffic routing to new pods
  - Ingress updated, 50% traffic to rolled-back pods

---
📈 METRICS (Compared to Last Update)
Before: Error rate 100%, Latency N/A
Now: Error rate 15%, Latency P95 800ms

Progress: Error rate dropping as traffic increases to healthy pods

---
🔍 INVESTIGATION STATUS
• Root Cause: Confirmed
• Root Cause: Database connection pool exhaustion due to missing `max_connections` increase after recent scaling. Pool size 25 insufficient for 6 replicas × 5 connections each = 30 needed.

Evidence:
- ✅ DB connections hit 100/100 at incident start
- ✅ fleet-core v1.23.4 used default pool size 25
- ✅ Rollback to v1.23.3 (pool size 50) resolving issue
- ✅ Monitoring confirmed pool exhaustion

---
👥 TEAM STATUS
• IC: Sarah Chen (active since 14:32 UTC)
• Technical Lead: Marcus Johnson
• SMEs: @dba-team completed, @infra-sre monitoring rollout
• Team fatigue: No fatigue reported, shifts organized

---
📞 ESCALATION STATUS
• Escalate? No
• Reason: Clear progress, ETA 5 min, metrics improving
• Executives: Already notified (14:35 UTC), next update at 15:00

---
💬 COMMUNICATIONS
• Status Page: Updated at 14:40 UTC - "Degraded Performance, rolling back fix"
• Customer Support: Ticketing volume 150/min (normal 20/min), template responses active
• Executive Notification: Sent 14:35 UTC, next at 15:00 UTC
• External Comms: PR aware, no external statements yet

---
⚠️ RISKS & BLOCKERS
Risks:
- Rollback may not fully resolve if underlying configuration issue persists → Monitoring connection pool metrics closely
- Second rollback wave might cause disruption → Staggering traffic increments (50% → 75% → 100%)

Blockers:
- None currently

---
🎯 NEXT ACTIONS (Next 15 min)
1. Complete traffic ramp to 100% - @infra-sre (by 14:55 UTC)
2. Monitor error rate to zero - @marcus-j (ongoing)
3. Prepare resolution announcement draft - @priya-sharma (by 14:55 UTC)

---
📅 NEXT UPDATE
• Scheduled: 2026-03-29 15:00 UTC
• Unless: Resolution achieved earlier

---
🔗 LINKS
• Linear: INC-2026-0123
• Grafana: https://grafana.flyfleet.com/d/fleet-api?viewPanel=error-rate
• Runbook: runbooks/api-outage.md
• War Room: #incident-INC-2026-0123

---
---
Update #3 | IC: Sarah Chen | INC-2026-0123
```

---

## Update Best Practices

1. **Be factual, not speculative** - "We observed X" not "We think Y caused it"
2. **Timestamps on actions** - Show progression
3. **Metrics tell story** - Include before/after
4. **Clear ownership** - Who is doing what
5. **Honest about unknowns** - "ETA unknown until we complete step X"
6. **Acknowledge challenges** - "We hit a blocker with X, working on Y"
7. **Forward-looking** - "Next 30 minutes we will..."

---

## Update Frequency Discipline

| Severity | Promise | Reality (exceed) |
|----------|---------|------------------|
| SEV1 | Every 15 min | Every 10-12 min (show progress) |
| SEV2 | Every 30 min | Every 20-25 min (keep stakeholders ahead) |
| SEV3 | At milestones | Milestone + one follow-up |

**Never miss an update cadence** - Even if no new info, send: "No significant change. Still investigating, ETA still 14:00 UTC."

---

**Template Version:** 1.0.0
**Last Updated:** 2026-03-29
