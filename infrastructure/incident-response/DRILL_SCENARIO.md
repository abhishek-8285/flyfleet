# Incident Response Drill: Tabletop Exercise

**Scenario:** "Telemetry Ingestion Lag Crisis"

**Duration:** 60 minutes
**Participants:** 6-8 people (IC, Technical Lead, SMEs, Comms Lead, Note-Taker, Observers)
**Materials:** This document, whiteboard/flipchart, markers, laptop (optional)

---

## Drill Setup

### Roles to Assign

| Role | Person | Responsibilities During Drill |
|------|--------|------------------------------|
| **Facilitator** | @sre-lead or trainer | Runs drill, injects scenarios, times sections |
| **Incident Commander (IC)** | Senior engineer | Makes decisions, coordinates team |
| **Technical Lead** | Mid-level engineer | Leads investigation, executes commands |
| **DBA SME** | Database expert | Database diagnostics |
| **NATS SME** | Messaging expert | NATS/telemetry expertise |
| **Communications Lead** | Product/PM | Drafts stakeholder updates |
| **Note-Taker** | Junior engineer | Maintains timeline |
| **Observers** | Other engineers, management | Watch, ask questions during debrief |

**Total:** 8 people max. If fewer, combine roles (DBA+NATS SME = Platform SME).

---

### Pre-Drill Setup (5 min)

1. **Facilitator** reads scenario setup (next page)
2. **IC** opens war room channel: `#drill-incident-2026-test`
3. **Note-Taker** creates timeline doc (Google Doc shared)
4. **All** review SEVERITY_QUICK_REFERENCE.md (5 min skim)

---

## Scenario: Telemetry Ingestion Lag Crisis

### Background (Read Aloud by Facilitator)

```
Date: Monday, March 30, 2026
Time: 14:00 UTC
You are the on-call SRE team for FlyFleet.

RECENT EVENTS:
- 13:15 UTC: Deployment of fleet-core v1.24.0 (optimize telemetry batch size)
- 13:30 UTC: Deployment of NATSJetStream config (increase retention to 30 days)
- 13:45 UTC: Alert: NATS storage at 75% (warning acknowledged, no action taken)
- 14:00 UTC: Alert fires: "TelemetryIngestionLagHigh"
```

### Current State (14:05 UTC)

**Metrics in Grafana:**
```
Telemetry Ingestion Lag: 12 minutes (threshold: 5 min) ❌
Telemetry Ingestion Rate: 4,000 events/sec (normal: 10,000) ❌
NATS Storage: 88% and rising 1%/min ⚠️
NATS Consumer Pending: 500,000 messages (normal: <10,000) ❌
fleet-core CPU: 45% (normal: 20%) ⚠️
fleet-core Memory: 60% (normal: 30%) ⚠️
Database Insert Latency: 500ms (normal: 50ms) ❌
Database Connections: 45/100 (normal: 20/100) ⚠️
```

**Customer Reports Starting:**
- Support ticket: "Vehicle locations 15 minutes old"
- Dashboard showing stale data
- Some customers reporting "no new trips logged in 20 minutes"

**Alertmanager:**
- PagerDuty page sent to on-call (@engineer-oncall)
- Slack #sre-alerts: 10 messages about lag

---

## Drill Timeline (60 min total)

### Phase 1: Initial Response (0-15 min)

**Facilitator says:** "It's 14:05 UTC. You just acknowledged the PagerDuty alert. Go."

**Team must:**
1. [ ] Check Grafana dashboard (Technical Lead)
2. [ ] Declare severity (IC) - justify why SEV1 or SEV2
3. [ ] Create Linear ticket: `INC-2026-TEST-001`
4. [ ] Create war room: `#incident-INC-2026-TEST`
5. [ ] Assign roles (if not already): IC, Tech Lead, SMEs, Comms, Note-Taker
6. [ ] Send initial notification (Comms Lead uses template)
7. [ ] Begin investigation (follow runbook)

**Facilitator injects at 5 min:**
```
💥 INJECTION #1
Customer Support Slack:
"@sre-team We're getting flood of tickets about stale locations.
Can you give ETA? CEO is asking.

 Ticket volume: 50/min (normal: 2/min)"
```

**Team should:** Acknowledge, include Comms Lead to handle

**Facilitator injects at 10 min:**
```
💥 INJECTION #2
NATS SME (@nats-expert) DM to IC:
"I'm looking at NATS storage. We're at 88% and filling 1%/min.
At this rate we'll hit 100% in 12 minutes. If storage fills,
NATS will reject new messages. That's catastrophic.

What do we do?"
```

**IC must decide:** Purge stream? Expand storage? Scale NATS?

**End Phase 1 Check:**
- [ ] Severity declared (SEV1 or SEV2, with rationale)
- [ ] War room created
- [ ] Initial investigation started (correct systems identified)
- [ ] Communications handling

---

### Phase 2: Investigation & Mitigation (15-35 min)

**At 15 min, Facilitator says:** "It's 14:20 UTC. 15 minutes in. What's your status?"

**Team must:**
1. [ ] Root cause category identified (Database bottleneck)
2. [ ] Following runbook: `telemetry-disruption.md` step 3 or 4
3. [ ] Executing diagnostics (check DB slow queries, connection pool)
4. [ ] Decide on mitigation (likely: increase pool size, add index, scale fleet-core)

**Facilitator provides data when asked:**

```
Q: Team asks "What are slow queries?"
A: (Facilitator provides)
SELECT query, mean_exec_time, calls FROM pg_stat_statements
WHERE query LIKE '%telemetry_events%'
ORDER BY mean_exec_time DESC LIMIT 3;

Results:
- INSERT INTO telemetry_events (device_id, ...) VALUES (...) ON CONFLICT ...
  mean_exec_time: 450ms, calls: 50,000  ← SLOW
- SELECT COUNT(*) FROM telemetry_events WHERE device_id = ...
  mean_exec_time: 25ms, calls: 120,000  ← OK
```

**Q: "What's connection pool status?"**
```
kubectl exec -it postgres-0 -- psql -c "SELECT count(*) FROM pg_stat_activity;"
Result: 85/100 (85% utilized)
```

**At 25 min, Facilitator injects:**
```
💥 INJECTION #3
Security Officer (@security-team) joins war room:
"Hey, I'm watching this. Any data loss concerns?
Are we seeing telemetry gaps or duplicates?
Need to assess if this is security incident."
```

**Technical Lead must:** Address security concerns (likely no breach, but check for data gaps)

**End Phase 2 Check:**
- [ ] Root cause confirmed: Database insert bottleneck
- [ ] Mitigation plan decided (e.g., add index, increase pool)
- [ ] Commands typed and ready to execute
- [ ] Impact assessment updated (how much data loss?)

---

### Phase 3: Recovery & Resolution (35-50 min)

**At 35 min, Facilitator says:** "Execute your mitigation plan."

**Team must:**
1. [ ] Type actual commands (Facilitator can simulate success)
2. [ ] Describe expected outcome
3. [ ] Define success criteria (error rate <1%, lag <5min)
4. [ ] Monitor metrics post-mitigation

**Example good commands:**
```bash
# Add index (if chosen)
kubectl exec -it postgres-0 -- psql -U fleet_user -c \
  "CREATE INDEX CONCURRENTLY idx_telemetry_device_received ON telemetry_events(device_id, received_at);"

# Increase connection pool
kubectl set env deployment/fleet-core DATABASE_MAX_OPEN_CONNS=50
kubectl rollout restart deployment/fleet-core

# Scale fleet-core horizontally
kubectl scale deployment fleet-core --replicas=4
```

**Facilitator responds with simulated metrics improvement:**
```
After index created (15 min later):
Telemetry lag: 10 min → 8 min → 5 min → 2 min ✓
Ingestion rate: 4k → 7k → 10k events/sec ✓
DB insert latency: 500ms → 450ms → 80ms ✓
```

**At 45 min, Facilitator says:** "Metrics are improving. When do you declare resolved?"

**Criteria for resolution (from runbook):**
- ✅ Lag < 5 min for 15 consecutive minutes
- ✅ Ingestion rate > 9,000 events/sec
- ✅ No errors in logs for 10 min
- ✅ Database healthy (connections < 80%)

**Team declares:** "Resolved at 14:52 UTC"

**Comms Lead sends:** Resolution announcement to status page

**End Phase 3 Check:**
- [ ] Mitigation executed successfully
- [ ] Metrics improved and stable
- [ ] Resolution declared with timestamp
- [ ] Status page updated to "resolved"

---

### Phase 4: PIR & Handoff (50-60 min)

**At 50 min, Facilitator says:** "Incident resolved. Start PIR preparation."

**Team must complete (in remaining 10 min):**

1. [ ] **Timeline** - Note-Taker shares draft timeline (key events only)
2. [ ] **Root Cause** - Apply 5 Whys or Fishbone
3. [ ] **Impact** - Quantify: customers affected, data loss, duration
4. [ ] **What Went Well** - 3 positive observations
5. [ ] **What Didn't Go Well** - 3 issues to improve
6. [ ] **Action Items** - 3-5 specific tasks with owners and due dates

**Example 5 Whys:**
```
1. Why telemetry lag? → Ingestion 6 min behind
2. Why ingestion slow? → DB insert 450ms (should be <50ms)
3. Why insert slow? → Missing index on telemetry_events(device_id, received_at)
4. Why missing index? → Schema change added column but didn't index
5. Why not caught? → Load testing didn't cover high-volume insert scenario
```

**Facilitator collects** PIR draft (one-page summary from team)

---

## Debrief (15 min after drill)

**Facilitator asks:**

1. **What went well?**
   - "We declared SEV2 quickly"
   - "Comms lead sent first update in 12 min"
   - "DB index fix was right"

2. **What was hard?**
   - "We debated SEV1 vs SEV2 for 10 min"
   - "Forgot to check NATS consumer lag until 20 min in"
   - "Comms template formatting took time"

3. **What would you do differently next time?**
   - "Declare SEV1 immediately (storage filling = high risk)"
   - "Parallel investigation: DB + NATS at same time"
   - "Use runbook checklist to avoid missing steps"

4. **Framework gaps?**
   - "Runbook should emphasize storage crisis first"
   - "Need template for Security Officer notification earlier"
   - "Add command to check NATS lag metric quickly"

5. **Knowledge gaps?**
   - "I don't know nats stream report commands" → Training needed
   - "Unclear on when to scale vs. optimize" → Add to runbook decision tree

---

## Evaluation Criteria

**Pass/Fail Thresholds:**

| Criteria | Passing | Failing |
|----------|---------|---------|
| Severity declaration | <5 min after alert | >10 min |
| War room setup | <10 min | >15 min |
| Root cause identified | <20 min | >30 min |
| Mitigation executed | <35 min | >45 min |
| Status updates | Sent on schedule | Late or missing |
| PIR draft complete | All sections drafted | Major gaps |
| Team communication | Clear, coordinated | Siloed, confused |

**Score = % of criteria met. Target: >80% (4/5)**

---

## Drill Scorecard (Facilitator Fills)

| Evaluation Item | Score (1-5) | Notes |
|-----------------|-------------|-------|
| Severity declaration speed | | |
| War room setup efficiency | | |
| Investigation approach (correct systems) | | |
| Command execution (accuracy) | | |
| Communication quality (clarity, timeliness) | | |
| Decision-making under pressure | | |
| Team coordination | | |
| PIR components completeness | | |
| adhered to runbook? | | |
| Overall performance | | |

**Overall rating:** ⭐⭐⭐⭐⭐ (circle)

**Top 3 strengths:**
1.
2.
3.

**Top 3 improvements:**
1.
2.
3.

**Action items from drill:**
- [ ] Update runbook: Add storage crisis decision tree
- [ ] Conduct NATS commands training session
- [ ] Pre-write status page templates for common scenarios
- [ ] Review storage alert thresholds (currently 75% warning, 88% critical)
- [ ] Run second drill in 2 weeks to reinforce

---

## Post-Drill Actions

### Within 24 Hours

1. **Facilitator** sends scorecard + notes to all participants
2. **Platform team** creates Linear tickets for identified gaps
3. **Update runbooks** based on lessons learned
4. **Schedule follow-up training** (if needed)

### Within 1 Week

1. **Participants** complete reading assignments (any runbooks they struggled with)
2. **IC candidates** shadow real incident (if any)
3. **Review drill recording** (if recorded) - self-improvement
4. **Update individual** incident response knowledge (self-assessment)

### Before Next Drill

- [ ] All action items from this drill complete
- [ ] Runbooks updated with new decision trees
- [ ] Training deck refined based on feedback
- [ ] New scenarios developed (different services)
- [ ] Quarterly drill scheduled

---

## Scenario Variations (For Future Drills)

### Scenario 2: API Complete Outage (SEV1)
- Fleet-core pods CrashLoopBackOff
- Database connection errors
- Rollback needed under pressure

### Scenario 3: Security Incident (SEV1)
- Compromised credentials detected
- Unauthorized access to PII
- Legal/Compliance escalation

### Scenario 4: Database Performance (SEV2)
- Slow queries, high load
- Connection pool exhaustion
- Potential data corruption

### Scenario 5: Command Delivery Failure (SEV1)
- Emergency SOS commands failing
- Queue depth growing
- Device offline issues

---

## Facilitator Guide

### Before Drill

- [ ] Print scorecards (one per participant)
- [ ] Set up war room test channel
- [ ] Prepare injects (print or separate doc)
- [ ] Test Grafana dashboard access (share screen)
- [ ] Have runbooks accessible (URLs)
- [ ] Assign roles ahead or let team choose

### During Drill

- **Keep time:** Strict 60 min + 15 min debrief
- **Inject scenarios:** At predetermined times or when team stalls
- **Be realistic:** Don't give answers, let team figure it out
- **Observe only:** Facilitator doesn't participate in response
- **Document:** Take notes on scorecard

### After Drill

- **Debrief immediately** while fresh
- **Focus on process, not individuals** (blameless culture)
- **Celebrate wins** (what went right)
- **Action items concrete** (who, what, when)
- **Follow up** on action items within 1 week

---

## Success Metrics for Drill Program

| Metric | Target |
|--------|--------|
| Drill participation rate | 100% of on-call engineers |
| PIR completion rate (post-drill) | 100% |
| Action items from drills completed | >90% |
| Real incident MTTR improvement | -20% per quarter |
| Confidence survey score (engineers) | >4.0/5.0 |
| Framework adherence in real incidents | >95% follow runbooks |

---

**Drill Scenario Version:** 1.0.0
**Last Updated:** 2026-03-29
**Next Review:** After first drill completion
