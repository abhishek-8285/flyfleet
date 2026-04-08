# Root Cause Analysis Method: Fishbone Diagram (Ishikawa)

**Purpose:** Explore multiple contributing factors to complex problems
**When to use:** Multi-causal incidents, system failures with many variables
**Time:** 30-60 minutes with team

---

## Method Overview

Fishbone (Ishikawa) diagrams categorize potential causes into structured buckets (typically 6). For each bucket, identify specific factors that contributed to the incident. The "head" is the problem (incident), "bones" are cause categories, "sub-bones" are specific factors.

---

## Common Categories for IT Incidents

### Standard 6 Categories

1. **People** - Human factors, training, staffing
2. **Processes** - Procedures, workflows, approvals
3. **Technology** - Code, infrastructure, configuration
4. **Policies** - Rules, standards, governance
5. **Tools** - Software, monitoring, automation
6. **External** - Vendors, dependencies, environment

### FlyFleet-Specific Categories

```
Technology
├── Code (bugs, algorithm)
├── Configuration (env vars, params)
├── Infrastructure (K8s, AWS, DB)
└── Architecture (design decisions)

Process
├── Deployment (CI/CD, rollout)
├── Testing (unit, integration, load)
├── Monitoring (alerts, dashboards)
└── Response (runbooks, training)

People
├── Knowledge (training gaps)
├── Staffing (on-call fatigue)
├── Communication (coordination)
└── Decisions (triage, escalation)

Policies
├── Standards (coding, ops)
├── Governance (change approval)
├── Compliance (security, audit)
└── SLAs (targets, penalties)

Tools
├── Observability (metrics, logs)
├── Automation (scripts, CI/CD)
├── Security (scanning, secrets)
└── Testing (load, chaos)

External
├── AWS/Azure/GCP outages
├── Third-party APIs (Stripe, Twilio)
├── Network (ISP, firewall)
└── Regulations (GDPR, India compliance)
```

---

## Example: Fleet-Core API Outage

```
                    API OUTAGE (SEV1)
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
   ┌────▼────┐      ┌─────▼─────┐     ┌─────▼─────┐
   │PEOPLE   │      │PROCESS    │     │TECHNOLOGY │
   ├─────────┤      ├───────────┤     ├───────────┤
   │• DBA not│      │• No pre-  │     │• Pool too │
   │  consulted│      │  flight   │     │  small    │
   │  on scaling│      │  validation│     │• Static   │
   │          │      │• Runbook  │     │  config   │
   └─────────┘      │  not followed│     │           │
                     └───────────┘     └───────────┘
        │                   │                   │
   ┌────▼────┐      ┌─────▼─────┐     ┌─────▼─────┐
   │POLICIES  │      │TOOLS      │     │EXTERNAL   │
   ├─────────┤      ├───────────┤     ├───────────┤
   │• No      │      │• No alert │     │• None     │
   │  capacity│      │  on pool  │     │           │
   │  planning│      │  saturation│     │           │
   │  standard│      │• Insufficient│   │           │
   │          │      │  load test│     │           │
   └─────────┘      └───────────┘     └───────────┘
```

---

## Template for PIR Section

### Fillable Diagram Structure

```markdown
### Fishbone Analysis

```
                    {PROBLEM STATEMENT}
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
   ┌────▼────┐      ┌─────▼─────┐     ┌─────▼─────┐
   │PEOPLE   │      │PROCESS    │     │TECHNOLOGY │
   ├─────────┤      ├───────────┤     ├───────────┤
   │• {factor}│      │• {factor} │     │• {factor} │
   │• {factor}│      │• {factor} │     │• {factor} │
   └─────────┘      └───────────┘     └───────────┘
        │                   │                   │
   ┌────▼────┐      ┌─────▼─────┐     ┌─────▼─────┐
   │POLICIES  │      │TOOLS      │     │EXTERNAL   │
   ├─────────┤      ├───────────┤     ├───────────┤
   │• {factor}│      │• {factor} │     │• {factor} │
   │• {factor}│      │• {factor} │     │• {factor} │
   └─────────┘      └───────────┘     └───────────┘
```

**Problem Statement:** {Concise description of incident}

**Contributing Factors:**

**People:**
- {Factor - e.g., "DBA not consulted on scaling decision"}
- {Factor - e.g., "On-call engineer fatigued"}

**Process:**
- {Factor - e.g., "No architectural review for config changes"}
- {Factor - e.g., "Deployment skipped rollback testing"}

**Technology:**
- {Factor - e.g., "Connection pool not sized for max replicas"}
- {Factor - e.g., "Hardcoded config instead of dynamic"}

**Policies:**
- {Factor - e.g., "No capacity planning policy"}
- {Factor - e.g., "No SLO-based deployment guardrails"}

**Tools:**
- {Factor - e.g., "No alert on pool utilization >80%"}
- {Factor - e.g., "Load testing didn't cover 6-replica scenario"}

**External:**
- {Factor - e.g., "AWS no issues"}
- {Factor - e.g., "None"}
```

---

## Detailed FlyFleet Example

### Incident: Telemetry Ingestion Lag (5 min)

```
                    TELEMETRY LAG > 5 MIN (SEV2)
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
   ┌────▼────┐      ┌─────▼─────┐     ┌─────▼─────┐
   │PEOPLE   │      │PROCESS    │     │TECHNOLOGY │
   ├─────────┤      ├───────────┤     ├───────────┤
   │• Ingestion│      │• No       │     │• NATS     │
   │  team      │      │  capacity │     │  storage  │
   │  unaware   │      │  planning │     │  too small│
   │  of backlog│      │  review   │     │• Consumer  │
   │            │      │  backlog   │     │  batch size│
   │            │      │  thresholds│     │  too small │
   └─────────┘      └───────────┘     └───────────┘
        │                   │                   │
   ┌────▼────┐      ┌─────▼─────┐     ┌─────▼─────┐
   │POLICIES  │      │TOOLS      │     │EXTERNAL   │
   ├─────────┤      ├───────────┤     ├───────────┤
   │• No       │      │• No alert │     │• None     │
   │  backlog  │      │  on       │     │           │
   │  monitoring│      │  pending msgs│   │           │
   │  requirement│     │  threshold │     │           │
   │            │      │• Insufficient│   │           │
   │            │      │  visibility│     │           │
   └─────────┘      └───────────┘     └───────────┘
```

**Analysis:**

From this diagram, we see incident had 14 contributing factors across all 6 categories. This was a systemic failure (not a single point of failure). Major themes:

1. **Technology limitation:** NATS storage not sized for burst load
2. **Process gap:** No capacity planning or backlog review process
3. **Tool gap:** No monitoring on pending messages (only on throughput)
4. **Policy absence:** No SLO requiring backlog < threshold

**Root Cause:** Cumulative effect of multiple small gaps rather than one big failure.

**Prevention:** Address gaps in multiple categories (not just technology).

---

## How to Facilitate

**Preparation:**
1. Gather cross-functional team (SREs, devs, DBA, product, PM)
2. Prepare virtual whiteboard (Miro, FigJam, or draw.io)
3. Write problem statement at head of fishbone

**Session (45 min):**
1. **Brainstorm categories** (5 min) - Fill in category labels relevant to incident
2. **Silent brainstorming** (10 min) - Each participant adds sticky notes to categories they own
3. **Group discussion** (20 min) - Review each factor, cluster similar, debate importance
4. **Root cause identification** (10 min) - Which factors, if fixed, would prevent recurrence?

**Follow-up:**
- Document all factors (even minor ones)
- Prioritize action items addressing top 3 root causes
- Assign owners from team that identified factors

---

## When to Use vs. Other Methods

| Scenario | Method |
|----------|--------|
| Single cause chain, clear progression | 5 Whys |
| Multiple independent factors | Fishbone |
| Complex timeline with dependencies | Timeline RCA |
| Need to understand sequence | Timeline RCA |
| Quick analysis with 2-3 people | 5 Whys |
| Formal review with 6-8 stakeholders | Fishbone |

**For FlyFleet incidents:**
- API/Database failures → 5 Whys (usually clean chain)
- Multi-service cascading failures → Fishbone (multiple contributors)
- Time-sensitive major incident → Timeline (first) then Fishbone (deep dive)

---

## Evidence Collection per Category

```
People:
- Interview transcripts
- On-call shift schedules
- Training completion records

Process:
- Deployment pipeline configs
- CI/CD logs
- Runbook checklists

Technology:
- System metrics (Grafana)
- Logs (kubectl logs)
- Config files (git history)

Policies:
- Compliance documents
- Change approval records
- SLA definitions

Tools:
- Monitoring rule definitions
- Alertmanager configs
- Test coverage reports

External:
- AWS status page screenshots
- Third-party incident reports
- Network logs from ISP
```

---

## Pitfalls to Avoid

| Pitfall | Solution |
|---------|----------|
| **Blame-focused** ("People messed up") | Shift to process: "Why did process allow person to make error?" |
| **Too many factors** (>20) | Cluster similar, focus on top 5-7 per incident |
| **Superficial** ("Technology failed") | Dig deeper: Which tech? What specific failure mode? |
| **No action items** | Ensure each category has at least one improvement action |
| **Done in isolation** | Include diverse perspectives (dev, ops, product, security) |

---

**Method Version:** 1.0.0
**Last Updated:** 2026-03-29
