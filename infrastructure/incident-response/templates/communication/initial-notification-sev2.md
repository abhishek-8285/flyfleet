# SEV2 Initial Notification Template

**Use:** Alert engineering leadership within 30 minutes of SEV2 declaration
**Channel:** Slack (#engineering-alerts or #sre-alerts)
**Audience:** VP Engineering, Engineering Directors, Principal Engineers, Product VP, Customer Support Lead

---

```
⚠️ [SEV2] {Service/System Name} - {Brief Description}

 Incident Response Activated

---
📅 DETAILS
• Declared: {YYYY-MM-DD HH:MM UTC}
• Severity: SEV2 - Major Impact
• Incident Commander: {Name} (@{slack})
• Warlroom: #incident-INC-YYYY-NNNN
• Linear Ticket: INC-YYYY-NNNN → [Link]

---
🎯 IMPACT
• Affected Services: {services}
• Customer Impact: {e.g., "25% of customers experiencing slow dashboard loads", "Mobile app reporting degraded telemetry"}
• Estimated Scope: {X% of users, specific regions, specific tenant groups}
• SLA Impact: {None yet / Potential if unresolved >2h}

---
🔍 INITIAL ASSESSMENT
• Symptoms: {e.g., "API P95 latency at 2.5s (target 300ms), error rate 3%"}
• Root Cause Category: {Infrastructure / Database / Network / Code / External}
• ETA to Mitigation: {tentative estimate or "unknown"}

---
👥 RESPONSE TEAM
• Incident Commander: {Name} (@{slack})
• Technical Lead: {Name} (@{slack})
• SMEs Engaged: {list}
• Communications Lead: {Name} (@{slack}) - Coordinating customer updates

---
📊 KEY METRICS
• Error Rate: {X}% (normal: <0.1%)
• Latency P95: {X}ms (normal: <300ms)
• Throughput: {X} rps (normal: ~baseline)
• Active Users Impacted: {X}/{total} or {X}%

---
🚀 CURRENT ACTIONS
• [ ] {Step 1} - {Owner}
• [ ] {Step 2} - {Owner}
• [ ] {Step 3} - {Owner}

---
📢 UPDATE CADENCE
• Engineering Updates: Every 30 minutes in this channel
• Customer Updates: Every 60 minutes via status page (if applicable)
• Next Engineering Update: {YYYY-MM-DD HH:MM UTC}

---
📞 ESCALATION PATH
• IC unavailable → Backup IC: @{backup-ic-slack}
• No progress after 2h → VP Engineering: @{vp-engineering}
• Customer pain → Customer Support Lead: @{support-lead}
• Legal concerns → Legal: @{legal}

---
💡 BACKGROUND
{Optional: 1-2 sentences on context, e.g., "Issue started following deployment of v1.2.3 at 14:15 UTC. Rollback in progress."}

---
---
SEV2 Protocol Active
- War room for responders only
- Observers: Read-only in this channel, updates via war room thread
- All hands: No, unless escalated to SEV1
- Customer communication: Customer Support handling via status page

---
{IC Name}
Incident Commander
FlyFleet Platform Engineering
```

---

## Usage Instructions

1. **Fill bracketed fields** with actual data
2. **Send within 30 minutes** of SEV2 declaration
3. **Post in:** `#engineering-alerts` or `#sre-alerts` Slack channel
4. **Create Linear ticket** first and link it
5. **Follow 30-minute update cadence**

---

## Example (Filled)

```
⚠️ [SEV2] Dashboard API - Degraded Performance - Latency >2s

 Incident Response Activated

---
📅 DETAILS
• Declared: 2026-03-29 10:15 UTC
• Severity: SEV2 - Major Impact
• Incident Commander: James Wong (@jameswong-sre)
• Warlroom: #incident-INC-2026-0124
• Linear Ticket: INC-2026-0124 → https://linear.app/flyfleet/issue/FF-1234

---
🎯 IMPACT
• Affected Services: dashboard-api (frontend), fleet-core read replicas
• Customer Impact: Dashboard loading slowly (>10s), some reports timing out
• Estimated Scope: ~30% of customers (mostly US East region)
• SLA Impact: None yet, monitoring for breach if >2h

---
🔍 INITIAL ASSESSMENT
• Symptoms: Dashboard API P95 latency 2.5s (target 300ms), error rate 3% (target <0.1%)
• Root Cause Category: Database - read replica lag ~15 seconds
• ETA to Mitigation: ~30 minutes (promoting new replica)

---
👥 RESPONSE TEAM
• Incident Commander: James Wong (@jameswong-sre)
• Technical Lead: Alex Kim (@alexk-fleet)
• SMEs Engaged: @dba-team (PostgreSQL), @infra-sre (Kubernetes)
• Communications Lead: Maria Santos (@maria-comm)

---
📊 KEY METRICS
• Error Rate: 3.2% (normal: <0.1%)
• Latency P95: 2500ms (normal: <300ms)
• Throughput: 850 rps (normal: 1200 rps)
• Active Users Impacted: ~750/2500 (30%)

---
🚀 CURRENT ACTIONS
• [ ] Check replication lag on read replicas - @dba-team
• [ ] Promote healthy replica if lag persists - @dba-team
• [ ] Update connection pool to prefer primary - @alexk-fleet
• [ ] Draft customer status page update - @maria-comm

---
📢 UPDATE CADENCE
• Engineering Updates: Every 30 minutes in this channel
• Customer Updates: Status page update in 30 min, then hourly
• Next Engineering Update: 2026-03-29 10:45 UTC

---
📞 ESCALATION PATH
• IC unavailable → Backup IC: @patricia-sre
• No progress after 2h → VP Engineering: @david-vpeng
• Customer pain spike → Support Lead: @daniel-support
• Legal concerns → Legal: @legal-team

---
💡 BACKGROUND
Issue started following today's 09:00 UTC deployment of database connection pool changes (fleet-core v1.23.5). Initial hypothesis: read replica connection routing misconfiguration. Rollback candidate ready.

---
---
SEV2 Protocol Active
- War room: #incident-INC-2026-0124 (responders only)
- Observers: Read-only here, live updates in war room
- All hands: No (escalate if degrades to SEV1)
- Customer status: Updates via status page every hour

---
James Wong
Incident Commander
FlyFleet Platform Engineering
```

---

## When to Escalate to SEV1

Escalate immediately (and send SEV1 template) if:

- ✅ Impact > 50% of customers (up from 30%)
- ✅ Error rate > 10% or latency > 5s P95
- ✅ Database primary failover needed
- ✅ Data corruption discovered
- ✅ No mitigation path within 60 minutes
- ✅ Revenue loss > $10k/hr threshold crossed

---

**Template Version:** 1.0.0
**Last Updated:** 2026-03-29
