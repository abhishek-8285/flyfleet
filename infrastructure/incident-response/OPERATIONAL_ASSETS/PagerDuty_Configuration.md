# PagerDuty Configuration for FlyFleet Incident Response

**Import these escalation policies into your PagerDuty account**

---

## Policy 1: SEV1 Critical Response

**Name:** `FlyFleet SEV1 Critical`
**Type:** Escalation Policy
**Description:** For critical incidents (SEV1)

### Rule: SEV1 Immediate Response

```
Trigger: Any SEV1 alert (severity: critical)

Step 1 - Page On-Call (0 min)
  User: @sre-oncall (on-call rotation)
  via: phone + sms + push
  timeout: 5 min

Step 2 - Escalate to backup (5 min)
  User: @sre-backup
  via: phone + sms + push
  timeout: 5 min

Step 3 - Page IC (10 min)
  User: @ic-rotation (senior on-call)
  via: phone + sms + push
  timeout: 5 min

Step 4 - VP Engineering (15 min)
  User: @vp-engineering
  via: phone + sms + push
  timeout: 30 min

Step 5 - CEO (30 min)
  User: @ceo
  via: phone + sms
  timeout: 60 min
```

**Targets:**
- On-call schedule: `FlyFleet SRE Primary`
- IC schedule: `FlyFleet Incident Commanders`

---

## Policy 2: SEV2 Major Response

**Name:** `FlyFleet SEV2 Major`
**Type:** Escalation Policy

### Rule: SEV2 Response

```
Trigger: Any SEV2 alert (severity: warning with slo tag)

Step 1 - Page On-Call (0 min)
  User: @sre-oncall
  via: push + email
  timeout: 15 min

Step 2 - Escalate to backup (15 min)
  User: @sre-backup
  via: push + email
  timeout: 15 min

Step 3 - Engineering Manager (30 min)
  User: @eng-manager
  via: push + email
  timeout: 60 min

Step 4 - VP Engineering (60 min)
  User: @vp-engineering
  via: push + email
  timeout: 120 min
```

---

## Policy 3: Security Incidents

**Name:** `FlyFleet Security Incidents`
**Type:** Escalation Policy

### Rule: Security Alerts

```
Trigger: Any security-related alert
  - auth failures > threshold
  - impossible travel
  - data exfiltration

Step 1 - Page Security On-Call (0 min)
  User: @security-oncall
  via: phone + sms + push
  timeout: 5 min

Step 2 - Security Officer (10 min)
  User: @security-officer
  via: phone + sms
  timeout: 15 min

Step 3 - Legal/Compliance (20 min)
  User: @legal
  via: email + push
  timeout: 60 min

Step 4 - VP Engineering (30 min)
  User: @vp-engineering
  via: phone + push
  timeout: 60 min
```

---

## On-Call Schedules

### Schedule 1: Primary SRE On-Call

**Name:** `FlyFleet SRE Primary`
**Type:** Schedule
**Time zone:** UTC
**Rotation:** Weekly (Monday 00:00 UTC)

**Rules:**
```
Week 1: @engineer1
Week 2: @engineer2
Week 3: @engineer3
Week 4: @engineer1 (repeat)

Override: If engineer unavailable, @sre-backup covers
```

**Handoff:** Monday 00:00 UTC. Handoff note in #sre-team Slack.

---

### Schedule 2: Incident Commander Rotation

**Name:** `FlyFleet Incident Commanders`
**Type:** Schedule
**Time zone:** UTC
**Rotation:** Weekly (Monday 09:00 UTC)

**Rules:**
```
Primary IC: Senior SRE (rotation every week)
Backup IC: Next in seniority (automatically escalates)

Week 1: @senior-sre1
Week 2: @senior-sre2
Week 3: @senior-sre3
Week 4: @senior-sre1

IC always on-call (expected to respond within 5 min for SEV1)
```

---

### Schedule 3: Security On-Call

**Name:** `FlyFleet Security On-Call`
**Type:** Schedule
**Rotation:** Weekly (Monday 00:00 UTC)

```
Week 1: @security-engineer1
Week 2: @security-engineer2
...
```

---

## Alert Severity Mapping

Configure your alert rules (in Prometheus Alertmanager) to set:

| Alert Rule | Severity Label | PagerDuty Routing Key |
|------------|----------------|------------------------|
| `*BurnCritical` (SLO fast burn) | critical | FlyFleet SEV1 Critical |
| `*BurnWarning` (SLO slow burn) | warning | FlyFleet SEV2 Major |
| `*High` (infrastructure) | warning | FlyFleet SEV2 Major |
| `*Down` (service down) | critical | FlyFleet SEV1 Critical |
| `Security*` (security alerts) | critical | FlyFleet Security Incidents |

**Example Alertmanager config:**

```yaml
route:
  receiver: default
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h

  routes:
    - match:
        severity: critical
      receiver: critical-pagerduty
      routes:
        - match:
            alertname: 'Security*'
          receiver: security-incidents
        - match:
            slo: 'api-availability'
          receiver: sev1-critical

    - match:
        severity: warning
      receiver: sev2-major

receivers:
  - name: 'sev1-critical'
    pagerduty_configs:
      - routing_key: 'YOUR_PAGERDUTY_SEV1_KEY'
        severity: 'critical'

  - name: 'sev2-major'
    pagerduty_configs:
      - routing_key: 'YOUR_PAGERDUTY_SEV2_KEY'
        severity: 'warning'

  - name: 'security-incidents'
    pagerduty_configs:
      - routing_key: 'YOUR_PAGERDUTY_SECURITY_KEY'
        severity: 'critical'
```

---

## Integration with Slack

### PagerDuty → Slack Bot

Install PagerDuty Slack app and configure:

```
Channel: #sre-alerts
Bot: @pagerduty

Settings:
- Post on all incidents (not just page)
- Include: incident title, severity, assigned to, link
- Do NOT include: phone numbers (privacy)
```

**Expected format:**
```
🚨 [SEV1] Fleet Core API - Complete Outage
Assigned to: @sre-oncall
Incident URL: https://pagerduty.com/incidents/123
```

---

## Testing Your Configuration

### Test SEV1 Escalation

```bash
# Trigger test alert via PagerDuty API (test only)
curl -X POST https://events.pagerduty.com/v2/enqueue \
  -H "Content-Type: application/json" \
  -H "Authorization: Token token=YOUR_INTEGRATION_KEY" \
  -d '{
    "routing_key": "YOUR_SEV1_INTEGRATION_KEY",
    "event_action": "trigger",
    "payload": {
      "summary": "TEST: SEV1 incident response",
      "severity": "critical",
      "source": "pagerduty-test",
      "component": "test",
      "custom_details": {
        "service": "fleet-core",
        "slo": "api-availability"
      }
    }
  }'
```

Expected:
- T+0: On-call (@sre-oncall) paged
- T+5min: Backup paged if no ack
- T+10min: IC paged
- T+15min: VP Eng paged

---

### Test SEV2 Escalation

Same as above but use `"severity": "warning"` and SEV2 integration key.

Expected:
- On-call gets push notification
- After 15 min: backup notified
- After 30 min: Eng Manager notified

---

## Maintenance

### Weekly (On-Call Handoff)

During handoff call (Monday 00:30 UTC):
- [ ] Verify PagerDuty schedule updated correctly
- [ ] Confirm no upcoming vacations (adjust overrides)
- [ ] Check phone numbers in profiles (ensure working)
- [ ] Test push notification (each person acknowledges test)

### Monthly

- [ ] Review escalation policies (are they still appropriate?)
- [ ] Update contacts (promotions, departures)
- [ ] Test all integration keys (expire annually)
- [ ] Review incident metrics: time-to-acknowledge, escalations used
- [ ] Adjust timeouts if needed (based on team feedback)

### Quarterly

- [ ] Audit on-call rotations (fairness, workload)
- [ ] Review policy effectiveness (are we reaching right people?)
- [ ] Update after organizational changes (new roles, teams)

---

## Troubleshooting

### Problem: On-call not receiving pages

**Check:**
1. User profile has correct phone number & push device
2. User is in the schedule (check PagerDuty web UI)
3. Escalation policy includes correct schedule/user
4. Alert has correct `routing_key` or `service` mapping
5. PagerDuty integration key active (not expired)

---

### Problem: Wrong person paged

**Check:**
1. Schedule override active (someone on vacation)?
2. Escalation policy points to correct schedule/user
3. Alert `incident_key` or `service` mapping correct
4. User still on-call (handoff happened?)

---

### Problem: Escalation too fast/slow

**Adjust:**
- Escalation `timeout` in policy (currently 5 min, 15 min, etc.)
- Consider timezone differences (use UTC everywhere)
- Add reminder before escalation (PagerDuty feature)

---

## Emergency Override

If PagerDuty unavailable or someone unreachable:

1. **Manual escalation:**
   ```
   #pagerduty channel in Slack
   /pd escalations create --incident INC-123 --user @vp-engineering
   ```

2. **Phone tree fallback:**
   See EMERGENCY_CONTACTS.md (coming soon) for phone numbers

3. **Slack as backup:**
   Post in `#sre-alerts` tagging people:
   ```
   @sre-backup Please respond to INC-123 (severe telemetry lag)
   ```

---

## Required Setup Steps

1. **Create schedules** in PagerDuty:
   - `FlyFleet SRE Primary`
   - `FlyFleet Incident Commanders`
   - `FlyFleet Security On-Call`

2. **Enter team members** with:
   - Name, email, phone number
   - Slack handle
   - Timezone (UTC preferred)

3. **Create escalation policies** (import above JSON/YAML)

4. **Configure alert integration**:
   - Generate integration keys for each policy
   - Update Alertmanager config with keys
   - Test end-to-end

5. **Install Slack app** and connect channels

6. **Document**:
   - Add PagerDuty admin as contact in EMERGENCY_CONTACTS.md
   - Update FRAMEWORK.md with actual phone numbers
   - Create runbook: "PagerDuty failure fallback"

---

## Sample User Setup (PagerDuty)

```json
{
  "user": {
    "name": "Sarah Chen",
    "email": "sarah@flyfleet.com",
    "time_zone": "UTC",
    "slack_user_channel": "@sarah-chen",
    "phone_numbers": [
      {
        "type": "mobile",
        "country_code": "1",
        "phone_number": "555-555-1234"
      }
    ],
    "notification_rules": [
      {
        "type": "push",
        "start_delay_in_minutes": 0
      },
      {
        "type": "phone_call",
        "start_delay_in_minutes": 5
      },
      {
        "type": "sms",
        "start_delay_in_minutes": 5
      }
    ]
  }
}
```

---

**Configuration Version:** 1.0.0
**Last Updated:** 2026-03-29
**To implement:** Copy policies into PagerDuty web UI or use API
