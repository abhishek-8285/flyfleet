# Slack Setup for FlyFleet Incident Response

**Configure Slack channels, bots, and integrations**

---

## Channel Structure

### 1. Alert Channel (Read-Only for Responders)

**Name:** `#sre-alerts`
**Purpose:** PagerDuty alerts, SLO alerts, system notifications
**Members:** All engineers, leadership
**Settings:**
- Public channel
- Archive: Never auto-archive
- Who can post: Only apps (PagerDuty, Alertmanager bot)
- Topic: "FlyFleet SRE alerts - critical notifications only"

**Integrations:**
- PagerDuty bot (post incidents)
- Alertmanager bot (post alerts)
- Grafana bot (post SLO dashboard snapshots)

---

### 2. Incident War Rooms (Temporary)

**Naming:** `#incident-INC-YYYY-NNNN`
**Purpose:** Active incident coordination
**Members:** IC, Technical Lead, SMEs, Comms Lead, Note-Taker
**Settings:**
- Private channel (invite-only)
- Archive after 30 days (auto)
- Who can post: All members
- Topic: "IC: @name | Linear: INC-YYYY-NNNN | Bridge: [link]"

**Creation:**
```bash
# Manual (when needed)
/slack create incident-INC-2026-0123

# Or auto-create via PagerDuty webhook
# Configure PagerDuty → Slack integration to auto-create
```

---

### 3. SRE Team Channel (Ongoing)

**Name:** `#sre-team`
**Purpose:** Daily SRE discussion, questions, knowledge sharing
**Members:** All SREs, Platform Engineers
**Settings:**
- Public channel
- Archive: Never
- Who can post: All members
- Topic: "FlyFleet SRE - Platform Engineering"

---

### 4. Incident Observers (Optional)

**Name:** `#incident-observers`
**Purpose:** Executives and other teams who want updates but shouldn't interfere
**Members:** Leadership, Product, Customer Support leads
**Settings:**
- Public channel
- Read-only for observers (moderated)
- Who can post: IC and Comms Lead only
- Topic: "Observations channel for non-responders"

---

## Required Slack Apps/Bots

### 1. PagerDuty App

**Install:** From Slack App Directory
**Configuration:**
- Connect to PagerDuty workspace
- Subscribe to `#sre-alerts`
- Set notification format:
  ```
  🚨 [SEV1] {title}
  Assigned to: @{assignee}
  Incident: {link}
  ```

**Test:**
```bash
# Trigger test incident from PagerDuty
# Should appear in #sre-alerts within 30 seconds
```

---

### 2. Alertmanager Bot (Custom)

**Purpose:** Post Prometheus alerts to Slack
**Implementation:** Use webhook integration

**Setup:**
1. In Slack: Create incoming webhook for `#sre-alerts`
   - Settings → Apps → Incoming Webhooks → Add to `#sre-alerts`
   - Copy webhook URL

2. In Alertmanager (K8s):
```yaml
receivers:
  - name: 'slack-sre-alerts'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK/URL'
        channel: '#sre-alerts'
        title: '{{ .GroupLabels.alertname }}'
        text: '{{ .CommonAnnotations.summary }}'
        color: '{{ if eq .Severity "critical" }}danger{{ else }}warning{{ end }}'
```

**Test:** Trigger low-priority alert, verify it posts.

---

### 3. Grafana Bot (Optional)

**Purpose:** Share dashboard snapshots in war room
**Install:** Grafana Slack app
**Configure:** Connect to Grafana, authorize in `#incident-*` channels

**Usage in war room:**
```
/grafana snapshot https://grafana.flyfleet.com/d/fleet-api?viewPanel=error-rate
→ Posts image of panel in channel
```

---

## User Permissions & Roles

### Incident Commander (IC) Permissions

In Slack:
- ✅ Can rename war room channel (for clarity)
- ✅ Can invite SMEs as needed
- ✅ Can pin important messages
- ✅ Can set channel purpose/topic
- ✅ Can use @incident mention (create custom user group)

**Custom User Group:** `@incident`
- Members: All potential ICs (senior SREs)
- Usage: `/incident @incident` mentions all ICs (when needed)

---

### War Room Discipline

**Rules (post in channel description):**
```
🚨 INCIDENT WAR ROOM - RESPONDERS ONLY

✅ DO:
- Use threads for side discussions
- @mention specific people for action items
- Share screenshots/links to metrics
- Keep conversation focused on resolution

❌ DON'T:
- Use @channel (spams observers)
- Ask "what's the status?" (check pinned update)
- Debate in main thread (use thread or huddle)
- Invite observers without IC approval

OBSERVERS: Use #incident-observers for questions

Updates: IC posts at top-level every 15/30 min
```

---

## Automation & Integrations

### Auto-Create War Room (Optional)

Use PagerDuty webhook + Slack API:

**PagerDuty → Slack integration:**
1. In PagerDuty: Services → {your service} → Integrations
2. Add "Slack" integration
3. Select `#sre-alerts` as destination
4. Enable "Create incident channel" option

**Or use custom webhook:**
```python
# Incident channel creator bot
import slack_sdk

def create_incident_channel(incident_id, severity, title):
    client = slack_sdk.WebClient(token=SLACK_BOT_TOKEN)
    channel_name = f"incident-{incident_id}"

    # Create private channel
    response = client.conversations_create(
        name=channel_name,
        is_private=True
    )

    channel_id = response["channel"]["id"]

    # Set topic
    client.conversations_setTopic(
        channel=channel_id,
        topic=f"IC: TBD | Linear: {incident_id} | Bridge: {zoom_link}"
    )

    # Invite on-call and IC
    client.conversations_invite(
        channel=channel_id,
        users=['sre-oncall', 'ic-rotation']
    )

    # Post initial template
    client.chat_postMessage(
        channel=channel_id,
        text=f"[{severity}] {title}\n\nWar room created. IC to assign self.\nLinear: {incident_id}"
    )

    return channel_id
```

---

### War Room Reminder Bot

Schedule reminders for updates:

```python
# In war room, IC or bot posts update reminders
# Using Slack Workflow Builder or cron:

Every 15 minutes during SEV1:
  Send message to #incident-INC-123:
  "⚠️ Time for status update. IC: Please post update using template.
  Next update due in 15 minutes."

Every 30 minutes during SEV2:
  Same but 30min interval
```

---

## Channel Templates (Create Once)

### Template: Incident War Room (save as message)

```
🚨 [SEV{1/2}] {Service} - {Title}

Incident Command Post Established

📅 Detected: {YYYY-MM-DD HH:MM UTC}
🎯 Severity: SEV{1/2} - {Description}
👥 IC: @{ic-slack} (decision authority)
📊 Linear: INC-YYYY-NNNN
🌉 Bridge: {Zoom/Teams link}

📋 Quick Links:
• Grafana: {dashboard link}
• Runbook: {runbook link}
• Timeline: {Google Doc link}

📢 Update Cadence: Every {15/30} minutes

---
Responders: Use this channel for coordination.
Observers: Use #incident-observers or wait for updates.
Next update: {HH:MM UTC}
```

**Usage:** Copy-paste when creating channel, fill brackets.

---

## Member Management

### On-Call Rotation Groups

Create Slack user groups:

```
@sre-oncall
  - @engineer1 (current)
  - @engineer2 (backup)
  - @engineer3 (shadow)

@ic-rotation
  - @senior-sre1 (current)
  - @senior-sre2 (backup)
  - @senior-sre3 (shadow)

@dba-team
  - @dba-lead
  - @dba-engineer1
  - @dba-engineer2

@nats-team
  - @nats-expert1
  - @nats-expert2

@security-oncall
  - @sec-engineer1 (current)
  - @sec-engineer2 (backup)
```

**Why:** Mention `@sre-oncall` pages current on-call automatically based on schedule (if Slack PagerDuty integration active).

---

## Channel Archiving & Retention

### Auto-Archive War Rooms

After incident resolved:

```
/channel archive #incident-INC-2026-0123
```

**Keep for:** 30 days (for PIR reference), then archive to compliance storage if needed.

**Export before archiving:**
```bash
# Use Slack export tool or API
slack-export --channel incident-INC-2026-0123 --output ~/incident-archives/
```

---

## Notification Settings

### Critical: Ensure People Get Paged

**Settings to check per user:**
1. `Settings → Notifications → Direct messages`
   - ✅ "Send notifications for direct messages"
   - ✅ "Send notifications for mentions"

2. `Settings → Notifications → Keywords`
   - Add: "urgent", "critical", "@incident"

3. `Settings → Notifications → Mobile`
   - ✅ "Send mobile notifications"
   - ✅ "Always send notifications"

**Test:** Send test @mention to each on-call person.

---

## Emergency Communication Plan

### If Slack Down

Alternative channels:
1. **Phone tree:** Call tree from EMERGENCY_CONTACTS.md
2. **Email:** Send to sre-team@flyfleet.com (all engineers)
3. **SMS:** Group SMS via ChatOps tool or manual
4. **Zoom/Teams:** War room bridge still works (phone dial-in)

**Procedure:**
```
1. IC declares: "Slack down, switching to phone tree"
2. Call VP Eng, then Technical Lead, then SMEs
3. Use conference bridge number for war room
4. Document decisions, share via email afterward
5. When Slack back, post summary to #sre-alerts
```

---

## Monitoring & Compliance

### Slack Audit Logs (Enterprise Grid)

Enable audit logs to track:
- Who created/destroyed channels
- Who invited/removed members
- Messages deleted
- File downloads

**Review quarterly:**
- Ensure no unauthorized war room access
- Verify all incidents have corresponding channel
- Check that observers followed rules

---

## Integration Checklist

- [ ] PagerDuty app installed and configured
- [ ] Alertmanager webhook set up and tested
- [ ] Channels created: `#sre-alerts`, `#sre-team`
- [ ] User groups defined: `@sre-oncall`, `@ic-rotation`, `@dba-team`
- [ ] On-call schedules in PagerDuty match Slack groups
- [ ] War room creation automation tested (optional)
- [ ] All engineers trained on channel discipline
- [ ] Emergency contact procedure documented
- [ ] Audit logs enabled (if Enterprise Grid)
- [ ] Notification settings verified for on-call

---

## Sample Incident in Slack

```
#sre-alerts
🚨 [SEV1] Fleet Core API - Complete Outage
Assigned to: @sre-oncall (Sarah)
Incident: https://pagerduty.com/incidents/ABC123
 triggered: 14:30 UTC

---
@sarah-chen: Acknowledged. Declaring SEV1. Creating war room.

#incident-INC-2026-0123 created
Invited: @marcus-j @dba-team @infra-sre @priya-sharma

[sarah-chen pinned a message]

[sarah-chen changed channel topic]

---
#incident-INC-2026-0123

[sarah-chen] [Thread]
IC: Sarah Chen
Linear: INC-2026-0123
Bridge: https://zoom.us/j/123456?

@sarah-chen: Initial assessment: API error rate 100%. Checking Grafana now.
[marcus-j] I'm seeing DB connection pool exhausted.
[sarah-chen] @dba-team, please check PostgreSQL connections.
[dba-team] 100/100 active, all from fleet-core.
[sarah-chen] OK root cause likely pool size. Running rollback.

--- 15 minutes later

[sarah-chen] [Top-level message]
[SEV1] API Outage - Update #3 - 14:50 UTC

Status: Mitigating | Progress: 70%
✅ Error rate: 15% (was 40%)
🔧 Rollback complete, 4/4 pods healthy
⏱️  Next update: 15:05 UTC

Linear: INC-2026-0123
```

---

**Slack Setup Version:** 1.0.0
**Last Updated:** 2026-03-29
