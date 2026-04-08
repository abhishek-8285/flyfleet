# SEV1 Initial Notification Template

**Use:** Alert executives and stakeholders within 15 minutes of SEV1 declaration
**Channel:** Slack + SMS (if automated), Email
**Audience:** CEO, CTO, VP Engineering, VP Product, VP Sales, Legal, Security Officer

---

```
🚨 [SEV1] {Service/System Name} - {Brief Description of Issue}

 Incident Command Post (ICP) Established

---
📅 TIMELINE
• Detected: {YYYY-MM-DD HH:MM UTC}
• IC Assumed Command: {YYYY-MM-DD HH:MM UTC}
• Warlroom: {Slack Channel Link} | Bridge: {Zoom/Teams Link}

---
🎯 IMPACT ASSESSMENT
• Customer Impact: {Description, e.g., "All customers unable to track vehicles", "Payment processing failed"}
• Estimated Affected Users: {Number or percentage}
• Business Impact: {Revenue loss estimate, compliance risk, safety risk}
• SLA Impact: {SLA(s) violated, penalty risk if applicable}

---
🛠️ TECHNICAL SUMMARY
• Affected Services: {list, e.g., fleet-core API, telemetry ingestion}
• Symptoms: {5xx errors, latency spike, service unavailable}
• Initial Root Cause Category: {Infrastructure / Code Deploy / Database / External Provider / Unknown}
• Current Status: {Investigating / Mitigating / Restoring}

---
👥 RESPONSE TEAM
• Incident Commander (IC): {Name} (@{slack})
  - Primary decision authority
• Technical Lead: {Name} (@{slack})
  - Leading investigation and recovery
• Subject Matter Experts:
  - Database: @{dba-slack}
  - Infrastructure: @{infra-slack}
  - Security: @{security-slack} (if applicable)
• Communications Lead: {Name} (@{slack})
  - Stakeholder and customer communications

---
📊 METRICS (Latest)
• API Error Rate: {X}% (threshold: 0.1%)
• API Latency P95: {X}ms (threshold: <300ms)
• Telemetry Ingestion Rate: {X} events/sec (threshold: >10k)
• Command Success Rate: {X}% (threshold: 99.9%)
• Database Connections: {X}/{max} (threshold: <80%)

---
🚨 ACTIONS IN PROGRESS
• [ ] {Action 1} - {Owner}
• [ ] {Action 2} - {Owner}
• [ ] {Action 3} - {Owner}

---
📢 NEXT UPDATE
• Time: {YYYY-MM-DD HH:MM UTC} (in 15 minutes)
• Channel: This thread
• If urgent before then: Page IC directly

---
⚠️ IMPORTANT
• DO NOT forward this alert externally
• All customer inquiries → Customer Support (#support-team)
• Media inquiries → PR team (@{pr-contact})
• This incident is logged as: INC-YYYY-NNNN
• War room is for responders only - observers please use read-only

---
🔒 CONFIDENTIALITY NOTICE
This communication contains sensitive internal information about a technology incident.
Do not share outside FlyFleet without approval from Legal/Communications.

---
{IC Name}
Incident Commander
FlyFleet Platform Engineering
```

---

## Usage Instructions

1. **Fill in bracketed fields** {like.this} with actual values
2. **Execute immediately** - Send within 15 minutes of SEV1 declaration
3. **Post in two places:**
   - CEO/executive Slack channel (or email if no Slack)
   - War room channel (`#incident-INC-YYYY-NNNN`)
4. **Follow up** with SMS/pager if no acknowledgment within 5 minutes

---

## Example (Filled Out)

```
🚨 [SEV1] Fleet Core API - Complete Outage - 5xx Errors

 Incident Command Post (ICP) Established

---
📅 TIMELINE
• Detected: 2026-03-29 14:30 UTC
• IC Assumed Command: 2026-03-29 14:32 UTC
• Warlroom: #incident-INC-2026-0123 | Bridge: https://zoom.us/j/123456789

---
🎯 IMPACT ASSESSMENT
• Customer Impact: All customers unable to access dashboard or track vehicles
• Estimated Affected Users: 100% (2,500+ customers)
• Business Impact: ~$15,000/hour revenue loss, safety risk for emergency SOS
• SLA Impact: Availability SLA 99.95% violated, potential penalty for enterprise contracts

---
🛠️ TECHNICAL SUMMARY
• Affected Services: fleet-core API (all endpoints)
• Symptoms: 503 Service Unavailable on all API endpoints, health check failing
• Initial Root Cause Category: Database (PostgreSQL) - possible connection pool exhaustion
• Current Status: Investigating - checking DB connectivity and pool metrics

---
👥 RESPONSE TEAM
• Incident Commander (IC): Sarah Chen (@sarah-chen)
  - Primary decision authority
• Technical Lead: Marcus Johnson (@marcus-j)
  - Leading investigation and recovery
• Subject Matter Experts:
  - Database: @dba-team
  - Infrastructure: @infra-sre
  - Security: @security-team
• Communications Lead: Priya Sharma (@priya-sharma)
  - Stakeholder and customer communications

---
📊 METRICS (Latest)
• API Error Rate: 100% (threshold: 0.1%)
• API Latency P95: N/A (all failing)
• Telemetry Ingestion Rate: 0 events/sec (threshold: >10k)
• Command Success Rate: 0% (threshold: 99.9%)
• Database Connections: 100/100 (threshold: <80%)

---
🚨 ACTIONS IN PROGRESS
• [ ] Verify PostgreSQL pod health - @dba-team
• [ ] Check connection pool metrics - @marcus-j
• [ ] Restart fleet-core pods if DB healthy - @infra-sre
• [ ] Prepare customer communication draft - @priya-sharma

---
📢 NEXT UPDATE
• Time: 2026-03-29 14:45 UTC (in 15 minutes)
• Channel: This thread
• If urgent before then: Page IC directly via PagerDuty

---
⚠️ IMPORTANT
• DO NOT forward this alert externally
• All customer inquiries → Customer Support (#support-team)
• Media inquiries → PR team (@lauren-pr)
• This incident is logged as: INC-2026-0123
• War room is for responders only - observers please use read-only

---
🔒 CONFIDENTIALITY NOTICE
This communication contains sensitive internal information about a technology incident.
Do not share outside FlyFleet without approval from Legal/Communications.

---
Sarah Chen
Incident Commander
FlyFleet Platform Engineering
```

---

## After Sending

1. **Monitor acknowledgments** - Ensure all recipients saw it
2. **Update ticket** - Link Linear ticket in follow-up
3. **Follow 15-minute cadence** - Send updates strictly on schedule
4. **Escalate non-response** - If key stakeholder doesn't acknowledge, escalate to their manager

---

## Variations

### If Data Breach Suspected

Add after Impact Assessment:
```
🔐 SECURITY INCIDENT
• Type: {PII Exposure / Data Exfiltration / Unauthorized Access}
• Data Types: {Customer PII, Financial Data, Location Data}
• Estimated Records: {Number}
• Security Officer Engaged: @security-officer
• Legal Notification Status: Pending / In progress
• External Forensic: {Yes/No} - {Vendor if yes}
```

---

### If External Provider Outage

Add after Technical Summary:
```
🌐 EXTERNAL DEPENDENCY
• Provider: {AWS / NATS / Stripe / etc.}
• Provider Status: {Link to provider status page}
• Provider Incident ID: {if available}
• Expected Resolution: {time if available}
• Workaround Available: {Yes/No} - {describe}
```

---

**Template Version:** 1.0.0
**Last Updated:** 2026-03-29
