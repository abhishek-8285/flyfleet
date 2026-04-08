# Customer Communication Templates

**Purpose:** Status page posts and customer-facing messages for incidents
**Tone:** Transparent, empathetic, professional. Avoid technical jargon.
**Frequency:** SEV1 every 15 min, SEV2 every 30 min, SEV3 optional

---

## Status Page Post Template

**Used for:** Public status page updates (status.flyfleet.com)
**Format:** Markdown, includes sections, updates chronological (newest top)

---

### Incident Active

```
## {Service} - {Status} - Started {YYYY-MM-DD HH:MM UTC}

**Status:** {Active / Investigating / Mitigating / Monitoring}
**Severity:** SEV{1/2/3}
**Last Updated:** {YYYY-MM-DD HH:MM UTC}

### What's Happening

{2-3 sentence plain language description of issue. Avoid technical terms.}

Example:
"We are experiencing issues with our fleet tracking service. Some customers may see delayed location updates or see errors when accessing the dashboard. Our engineering team is investigating."

### Who Is Affected

{Describe impact clearly}
- {Impact 1} - {description}
- {Impact 2} - {description}

Example:
- Dashboard users: May experience slow loading or errors
- Mobile app users: Location tracking may show stale data (up to 5 min old)
- API consumers: Increased error rates (5xx responses)

### What We're Doing

{List actions in progress - reassure customers}
- Our engineering team was alerted at {time} and is actively investigating
- We have identified the likely cause and are implementing a fix
- We are monitoring the situation closely and will continue to update

### Workaround (if available)

{Clear step-by-step instructions for customers to work around issue}

**If no workaround:**
> "There is no workaround at this time. We apologize for the inconvenience and appreciate your patience as we work to resolve this."

**If workaround exists:**
> "In the meantime, you can:
> 1. Refresh the dashboard to retry loading
> 2. Use the mobile app as an alternative (issues not affecting app)
> 3. Access historical reports (unchanged)"

### Timeline

- **Detected:** {YYYY-MM-DD HH:MM UTC}
- **Investigating:** {Earliest action}
- **Mitigation:** {If known: deploying fix at HH:MM}
- **Expected Resolution:** {time window or "assessing"}

### Updates

Updates will be provided every {15/30/60} minutes until resolved.

**Next update by:** {YYYY-MM-DD HH:MM UTC}

---
*For immediate support, contact support@flyfleet.com or call {phone}*
*Subscribe to updates: [RSS feed] / [Email notifications]*

---

### Example: API Service Degradation

```
## API Service - Mitigating - Started 2026-03-29 10:15 UTC

**Status:** Mitigating
**Severity:** SEV2
**Last Updated:** 2026-03-29 10:45 UTC

### What's Happening

We are experiencing degraded API performance, causing some customers to see slower responses or errors when accessing the FlyFleet dashboard and reports.

### Who Is Affected

- Dashboard users: Page loads are slow (5-10 seconds) and some buttons may fail
- API consumers: Increased error rate (about 3% of requests failing)
- Mobile app: Minor impact (slightly slower refresh)

Approximately 30% of our customers are affected, mostly in the US East region.

### What We're Doing

Our engineering team identified the issue at 10:15 UTC and is actively working on a fix. We have found that one of our database replicas is experiencing lag, which is causing slow responses. We are promoting a healthy replica to restore performance.

We restored full service at 10:50 UTC and are monitoring to confirm the fix.

### Workaround

If you're experiencing issues:
1. Refresh the page to retry (may succeed after a few tries)
2. Use the mobile app as an alternative (less impacted)
3. Access historical reports (not affected by API latency)

### Timeline

- **Detected:** 2026-03-29 10:15 UTC
- **Investigating:** Completed by 10:30 UTC
- **Mitigation:** Deployed at 10:45 UTC
- **Expected Resolution:** 10:55 UTC (monitoring post-fix)

### Updates

Updates will be provided every 30 minutes until resolved.

**Next update by:** 2026-03-29 11:15 UTC

---
*For support, email support@flyfleet.com or call 1-800-FLYFLEET*
*Subscribe to updates: https://status.flyfleet.com/rss*
```

---

## Resolution Announcement

```
## {Service} - Resolved - {YYYY-MM-DD HH:MM UTC}

**Status:** ✅ Resolved
**Severity:** SEV{1/2}
**Resolved At:** {YYYY-MM-DD HH:MM UTC} ({duration} after detection)

### Resolution Summary

The issue with {service} has been resolved and all systems are operating normally.

### What Happened

{Brief 2-sentence summary of issue and cause. Example:
"We experienced a database connection pool exhaustion that caused the fleet API to become unavailable. This was triggered by a deployment that increased replica count without adjusting pool size."
}

### Impact

- **Duration:** {X} minutes (from {start} to {end})
- **Customers Affected:** {Approximate number or percentage}
- **Services:** All {services} restored to normal operation

### What We Fixed

{List actions taken to resolve}
- {Action 1}
- {Action 2}
- {Action 3}

Example:
- Rolled back fleet-core from v1.23.5 to v1.23.4
- Increased database connection pool size from 25 to 50
- Restarted all affected pods to pick up new configuration

### Prevention

{One sentence about what we're doing to prevent recurrence.
"This has triggered a review of our deployment process to ensure connection pool sizes are automatically adjusted when scaling."}

### Follow-up

- A detailed post-incident review (PIR) will be completed within 7 days
- Action items from the PIR will be tracked in our public roadmap
- We apologize for the inconvenience and thank you for your patience

---
*If you are still experiencing issues, please contact support@flyfleet.com*

*All systems operational as of {time}*
```

---

## Email Templates

### SEV1 Customer Email (for widespread impact)

**Subject:** Important: FlyFleet Service Disruption - {Brief Issue}

**Send:** To all affected customers (from billing data or usage logs)

```
Dear FlyFleet Customer,

We are writing to inform you of a service disruption that may have affected your use of FlyFleet.

Incident Details:
- Issue: {Brief description, e.g., "API outage affecting dashboard and mobile app"}
- Started: {time UTC}
- Status: {Active / Mitigating / Resolved}
- Impact: {What you may have experienced}

Current Status:
{Current status paragraph, 2-3 sentences}

What We're Doing:
Our engineering team is actively working to resolve this issue. We have {taken action / identified root cause and deploying fix / monitoring recovery}.

Expected Resolution:
We expect to restore full service by {time or "within the next hour"}.

Workaround:
{If available, describe. If not: "There is no workaround at this time."}

We understand this disruption impacts your operations and sincerely apologize. We are committed to preventing incidents of this nature in the future.

Updates:
- Status page: https://status.flyfleet.com (updated every 15 minutes)
- Support: support@flyfleet.com or 1-800-FLYFLEET

Sincerely,
The FlyFleet Team
```

---

### Post-Incident Follow-up Email (after PIR)

**Subject:** Incident Update: {Service Issue} - Post-Incident Report

```
Dear FlyFleet Customer,

We are following up on the service incident that occurred on {date}.

Incident Summary:
- Duration: {X hours/minutes}
- Impact: {Description}
- Root Cause: {Brief explanation}

Actions Taken:
We have implemented the following fixes:
1. {Fix 1}
2. {Fix 2}
3. {Fix 3}

Prevention:
To prevent recurrence, we are:
- {Action 1} (completed)
- {Action 2} (ETA {date})
- {Action 3} (ongoing)

Post-Incident Report:
The full post-incident review is available at: {link to PIR on docs site}

We apologize again for the disruption and appreciate your patience as we continuously improve our platform.

Best regards,
FlyFleet Engineering Team
```

---

## Internal Customer Support Templates

### Quick Response to Customer Inquiry (during incident)

```
Hi {Customer Name},

Thank you for reaching out. We are aware of an issue affecting {service} that may be causing the problem you described.

Our engineering team is actively working on a fix. You can track the status on our status page: https://status.flyfleet.com

Expected resolution: {time}

In the meantime, {workaround if available}.

We apologize for the inconvenience and will update you as soon as the issue is resolved.

Best,
{Support Agent Name}
FlyFleet Support
```

---

### Compensation Offer (for SEV1 incidents)

```
We sincerely apologize for the service disruption you experienced on {date}.

As a token of our regret, we are issuing a credit of {amount or percentage} to your account. This will appear on your next invoice.

No action is required on your part.

Thank you for your patience and understanding.

Sincerely,
FlyFleet Customer Success
```

**Standard credits:**
- SEV1 impacting >50% customers: 10% credit (up to 1 month)
- SEV2 impacting 25-50%: 5% credit
- SEV3 impacting <25%: $50 account credit or 2.5%

---

## Crisis Communications (Media/PR)

If media outlets or large influencers inquire:

```
For media inquiries, please contact:
{PR Person Name}
{Email}
{Phone}

We are aware of the technical issue and our team is working to resolve it. We will provide additional information as soon as possible.

[Only if safe to share:]
- No data breach occurred
- No customer data was compromised
- We expect resolution within {time}

DO NOT speculate on root cause. Refer all questions to PR.
```

---

## Slack/Discord Community Updates

**For community channels where customers gather:**

```
:warning: **Service Update** - {HH:MM UTC}

We're experiencing {issue description}. Our team is investigating.

**Status:** {Active/Mitigating}
**Impact:** {Brief}
**Next update:** {HH:MM UTC}

Status page for details: https://status.flyfleet.com

Our support team is aware and will respond to tickets as updates are available.
```

---

## Best Practices

1. **Be first, be right, be credible** - Update quickly even if "still investigating"
2. **No technical jargon** - "Database issue" not "connection pool exhaustion"
3. **Empathize** - Acknowledge customer frustration
4. **Clear timeline** - When detected, when expect fix
5. **Provide workarounds** if even partial
6. **Promise and deliver updates** - Meet stated cadence
7. **Own the mistake** - "We are experiencing" not "A third-party provider is having issues"
8. **Declare victory when done** - Clear "resolved" status, not "monitoring"

---

## Accessibility

For status page:
- Use proper HTML headings (## for sections)
- Alt text for any images
- Color not sole indicator of status (use text "Active", "Resolved")
- Screen reader friendly

---

## Templates at a Glance

| Template | Use | Update Freq |
|----------|-----|-------------|
| `status-page-active.md` | Status page during incident | 15/30 min |
| `status-page-resolved.md` | Status page when fixed | Once |
| `customer-email-sev1.md` | Email to all customers | Once per incident |
| `customer-email-pir.md` | PIR follow-up email | After PIR complete |
| `support-quick-response.md` | Support agent reply | As needed |
| `compensation-offer.md` | Credit notification | With invoice |
| `slack-community.md` | Community channels | Every update |
| `media-response.md` | PR/External only | On request |

---

**Template Version:** 1.0.0
**Last Updated:** 2026-03-29
