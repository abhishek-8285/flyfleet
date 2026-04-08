# FlyFleet Emergency Contacts - Incident Response

**Keep this document accessible at all times (printed copy in server room, saved in phone)**

---

## 🚨 Immediate Response (0-15 minutes)

### All Incidents
```
On-Call SRE: PagerDuty (auto)
Backup On-Call: @sre-backup (Slack) or +1-XXX-XXX-XXXX
Incident Commander: @ic-rotation (Slack)
```

### SEV1 Critical Escalation
```
VP Engineering: +1-XXX-XXX-XXXX (after 60 min no progress)
Security Officer: +1-XXX-XXX-XXXX (if PII involved)
Legal/Compliance: +1-XXX-XXX-XXXX (data breach, GDPR)
AWS Support: 1-888-825-4375 (infrastructure issues)
```

### SEV2 Major Escalation
```
Engineering Manager: @eng-manager (Slack)
VP Engineering: +1-XXX-XXX-XXXX (after 2h no resolution)
```

---

## 👥 On-Call Rotations

### Primary SRE On-Call
**Schedule:** Weekly rotation, Monday 00:00 UTC
**Current:** @engineer1 (Week of Mar 30)
**Backup:** @engineer2
**Shadow:** @engineer3

**Contact:** PagerDuty auto-page | Slack: @sre-oncall

```
Schedule URL: https://sre.pagerduty.com/schedules/...
Public Calendar: https://calendar.google.com/.../sre-rotation
```

---

### Incident Commander Rotation
**Schedule:** Weekly, Monday 09:00 UTC
**Current:** @senior-sre1 (Sarah Chen)
**Backup:** @senior-sre2 (Marcus Johnson)
**Shadow:** @senior-sre3

**Contact:** Slack: @ic-rotation

```
IC Calendar: https://calendar.google.com/.../ic-rotation
Overrides: Submit to @sre-lead 1 week ahead
```

---

### Security On-Call
**Schedule:** Weekly, Monday 00:00 UTC
**Current:** @security-engineer1
**Backup:** @security-engineer2

**Contact:** PagerDuty security escalation | Slack: @security-oncall

**When to page:**
- Unauthorized access detected
- PII exposure suspected
- Malware/ransomware
- DDoS attack
- Compliance violation

---

## 📞 Key Teams & SMEs

### Database Team
```
@dba-team (Slack group)
Lead: @dba-lead
Engineer: @dba-engineer1
Engineer: @dba-engineer2

Phone (urgent): +1-XXX-XXX-XXXX
After-hours pager: @dba-oncall (via PagerDuty)
Runbook: database-performance.md
```

**When to call:** Database down, slow queries, connection pool exhausted, replication lag

---

### NATS/Messaging Team
```
@nats-team
Lead: @nats-lead
Expert: @nats-expert1
Expert: @nats-expert2

Phone: +1-XXX-XXX-XXXX
Slack: @nats-team
Runbook: nats-issues.md
```

**When to call:** JetStream storage full, consumer lag, messages dropped, streams unavailable

---

### Infrastructure/SRE Platform
```
@infra-sre (Slack group)
Lead: @infra-sre-lead
Engineers: @infra-sre1, @infra-sre2

Phone: +1-XXX-XXX-XXXX
Slack: @infra-sre
Runbooks: All infrastructure runbooks
```

**When to call:** Kubernetes issues, networking, AWS, deployment problems, scaling issues

---

### Security Team
```
@security-team (Slack)
CISO: @ciso
Security Engineer: @security-engineer1
Security Engineer: @security-engineer2

Phone (urgent): +1-XXX-XXX-XXXX
Pager: PagerDuty security escalation
Email: security@flyfleet.com
Runbook: security-incident.md
```

**When to call:** Any security incident, data breach, unauthorized access, compliance questions

---

### Customer Support
```
Support Team: @support-team
Support Lead: @support-lead (Daniel Kim)
Phone: 1-800-FLYFLEET (1-800-359-3533)
Email: support@flyfleet.com
Slack: #support-team
Hours: 24/7 (rotating)
```

**When to involve:**
- Customer impact > 25%
- Status page needs update
- Need to draft customer communication
- Ticket volume spike (>10x normal)
- Customer complaints about incident

---

### Legal & Compliance
```
Legal Team: @legal
General Counsel: @gc (Sarah Johnson)
Compliance Officer: @compliance-officer
Phone (urgent): +1-XXX-XXX-XXXX
Email: legal@flyfleet.com
Slack: #legal-team (private)
```

**When to involve IMMEDIATELY:**
- Data breach (PII exposure)
- Regulatory notification required (GDPR 72h, India 24h)
- SLA penalty disputes
- Contractual liability questions
- Media inquiries
- Regulatory audit during incident

---

### Product & Engineering Leadership

```
VP Engineering: @vp-engineering (David Park)
  Phone: +1-XXX-XXX-XXXX
  Escalate: No progress after 60min (SEV1) or 2h (SEV2)
  Slack: @david-park

Engineering Director: @eng-director
  Slack: @eng-director
  Escalate: Resource constraints, staffing issues

Product VP: @product-vp
  Slack: @product-vp
  Escalate: Major feature impact, customer commitments

CTO: @cto
  Slack: @cto
  Escalate: Strategic decisions, major outages affecting revenue >$50k/hr
```

---

## 🌐 External Support

### Cloud Providers

**AWS (Amazon Web Services)**
```
Status Page: https://status.aws.amazon.com/
Support: 1-888-825-4375 (24/7)
Enterprise Support: https://console.aws.amazon.com/support/home
Account Team: [your account manager]
```

**When to call AWS:**
- EKS cluster unresponsive
- RDS down or unreachable
- VPC/networking issues you can't resolve
- EC2 instance problems
- S3 access issues affecting backups

**Before calling:** Check AWS status page, run diagnostics, gather CloudWatch logs, have account ID ready

---

### Third-Party Services

**NATS (messaging)**
```
Status: https://status.nats.io/
Support: support@nats.io
Enterprise: [contact if you have enterprise support]
```

**Stripe (payments)**
```
Status: https://stripe.com/status
Support: https://support.stripe.com/contact
Emergency: +1-XXX-XXX-XXXX (if you have enterprise)
```

**Twilio (SMS/MFA)**
```
Status: https://status.twilio.com/
Support: +1-XXX-XXX-XXXX
```

**SendGrid (email)**
```
Status: https://status.sendgrid.com/
Support: support@sendgrid.com
```

---

## 🆘 International Emergency Numbers

If working from different regions:

```
USA/Canada: 911
UK: 999
EU: 112
India: 100, 101 (ambulance), 102 (fire)
Singapore: 999
Australia: 000

For international travel: Pre-program local emergency numbers
```

*Note: These are for physical emergencies (fire, medical). Not for technical incidents.*

---

## 📋 Incident Command Post Contacts

### War Room Setup
```
Zoom Bridge: https://zoom.us/j/FlyFleetWarRoom?pwd=...
  PIN: 123456
  Dial-in: +1-XXX-XXX-XXXX (US) or +44... (UK)

Slack War Room: #incident-INC-YYYY-NNNN (create per incident)

Video: Enable waiting room OFF for rapid access
Recording: Only with Legal approval
```

### Documentation Collaboration
```
Google Drive: /FlyFleet/Incidents/INC-YYYY-NNNN
  - Timeline doc
  - Metrics screenshots
  - Evidence folder

Linear: https://linear.app/flyfleet/issue/FF-NNNN

Confluence: [if used] Incident page template
```

---

## 🏥 Medical & Safety (for on-site incidents)

**If incident involves physical safety (vehicle accident, injury):**

```
CEO: +1-XXX-XXX-XXXX
HR: +1-XXX-XXX-XXXX
Legal: +1-XXX-XXX-XXXX

Emergency Services: 911 (or local equivalent)
Report: Physical location, nature of emergency, injuries

OSHA (if work-related injury in US): 1-800-321-6742
Report within 8 hours for severe injuries
```

---

## 🔐 Security Hotline

**Anonymous security reporting:**
```
Security Hotline: +1-XXX-XXX-XXXX
Email: security-hotline@flyfleet.com (anonymous option)
Web form: https://flyfleet.com/security/report

Available 24/7, anonymous reporting allowed
Escalates to Security Officer within 15 minutes
```

---

## 📱 On-Call Phone Numbers (Save to Contacts)

Create contact group "FlyFleet On-Call":

```
On-Call SRE          +1-XXX-XXX-XXXX   (PagerDuty auto)
Backup SRE           +1-XXX-XXX-XXXX   (Slack: @sre-backup)
Incident Commander   +1-XXX-XXX-XXXX   (Slack: @ic-rotation)
Database Team        +1-XXX-XXX-XXXX   (Slack: @dba-team)
Security On-Call     +1-XXX-XXX-XXXX   (PagerDuty security)
VP Engineering       +1-XXX-XXX-XXXX   (Slack: @vp-engineering)
Legal                +1-XXX-XXX-XXXX   (Slack: @legal)
AWS Support          1-888-825-4375
Customer Support     1-800-FLYFLEET (1-800-359-3533)
```

**Tip:** Add these contacts to your phone's "ICE" (In Case of Emergency) list

---

## 🗃️ Document Locations

**Quick access URLs:**

| Document | Path/URL |
|----------|----------|
| Incident Framework | `infrastructure/incident-response/README.md` |
| Severity Quick Ref | `infrastructure/incident-response/SEVERITY_QUICK_REFERENCE.md` |
| Runbooks | `infrastructure/incident-response/runbooks/` |
| PIR Template | `infrastructure/incident-response/frameworks/pir-template.md` |
| Training Deck | `infrastructure/incident-response/TRAINING_PRESENTATION.md` |
| Drill Scenario | `infrastructure/incident-response/DRILL_SCENARIO.md` |
| Emergency Contacts | THIS FILE! |
| PagerDuty Config | `infrastructure/incident-response/OPERATIONAL_ASSETS/PagerDuty_Configuration.md` |
| Slack Setup | `infrastructure/incident-response/OPERATIONAL_ASSETS/Slack_Setup.md` |

**Local paths:**
```
/data/data/com.termux/files/home/flyfleet/infrastructure/incident-response/
```

---

## ⏰ Handoff Times

**Weekly handoffs (Monday 00:00 UTC):**
- SRE On-Call rotation changes
- Security On-Call rotation changes
- IC rotation changes (if Monday)

**Handoff call:** 00:30 UTC, Zoom bridge (same as war room)
**Duration:** 15 minutes
**Agenda:**
- Upcoming deployments/maintenance
- Known issues/hotspots
- Pending PIR actions
- Team availability (vacations, meetings)

---

## 🆘 When All Else Fails

**If you cannot reach anyone:**
1. Try alternate channel (Slack vs. phone)
2. Escalate to next level (backup, then manager's manager)
3. Use phone tree (call down list until someone answers)
4. Page AWS Support if infrastructure completely down
5. If no response after 30 min, drive to office (if local) - physical presence works

**Last resort:**
```
CEO: +1-XXX-XXX-XXXX (only for true SEV1, >$50k/hr loss)
```

---

## 🔄 Update This Document

This contact list must be updated:

- **Immediately:** When phone numbers change
- **Monthly:** Rotations, vacant positions
- **Quarterly:** Full audit, distribute updated copy

**How to update:**
1. Edit this file in Git
2. Pull request with changes
3. Notify @sre-lead to merge and redistribute
4. Print new copies (keep in multiple locations)

**Distribution:**
- Server room (printed copy)
- On-call desk (printed copy)
- Slack pin in #sre-team
- GitHub repo (this file)
- Each engineer's phone contacts

---

## 📝 Acknowledgment

**By signing below, you acknowledge you have:**
- [ ] Read this emergency contacts document
- [ ] Saved key phone numbers to your phone
- [ ] Reviewed incident response framework
- [ ] Understood escalation procedures
- [ ] Know when to call each contact

**Signature:** ____________________  **Date:** ___________

**Print name:** ____________________

**Email:** ____________________

**Role:** □ SRE □ IC □ DBA □ Security □ Other: _______

---

**Document Version:** 1.0.0
**Last Updated:** 2026-03-29
**Next Review:** 2026-06-29 (quarterly)
**Owner:** Platform Engineering Team

**EMERGENCY - DO NOT REMOVE FROM SERVER ROOM**
