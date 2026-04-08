# Root Cause Analysis Method: 5 Whys

**Purpose:** Simple, iterative questioning to drill down to root cause
**When to use:** Clear cause-effect chain, involves human error or process failure
**Time:** 15-30 minutes

---

## Method Overview

The 5 Whys technique involves asking "Why?" repeatedly (typically 5 times) to move from symptoms to the underlying root cause. Each "Why" should be answerable with evidence, not speculation.

---

## Example: API Outage

```
1. Why did the API return 503 errors?
   → Health checks failed due to database connection timeouts.

2. Why did database connections time out?
   → All connections in the pool were in use, new requests blocked waiting.

3. Why were all connections in use?
   → Connection pool max_open_conns=25 was insufficient for 6 app replicas.

4. Why was pool size not adjusted when replicas increased from 2 to 6?
   → Configuration was hardcoded; no automatic adjustment when HPA scales.

5. Why was there no automatic adjustment or guardrail?
   → Deployment pipeline lacks validation that pool size matches replica count; no architectural requirement defined.
```

**Root cause:** Deployment pipeline gap allowing configuration drift without validation.

---

## Template for PIR Section

```
### 5 Whys Analysis

1. **Why** did {symptom} occur?
   {Answer with evidence}

2. **Why** did that happen?
   {Drill down one level}

3. **Why** (continued)?
   {Third level}

4. **Why** (continued)?
   {Fourth level}

5. **Why** (continued)?
   {Root cause revealed}

**Root Cause:** {concise statement}

**Evidence supporting root cause:**
- {Evidence 1}
- {Evidence 2}
```

---

## When to Use This Method

✅ **Good for:**
- Process failures (deployment error, configuration mismatch)
- Human error with clear procedural gaps
- Simple technical chains (A → B → C)

❌ **Not good for:**
- Multi-causal failures (use Fishbone instead)
- Complex system interactions (use Timeline RCA)
- Need to assign blame (this method avoids blame)

---

## Facilitation Tips

1. **Start with facts, not opinions** - Base first "Why" on observed metrics/logs
2. **Avoid jumping** - Each "Why" should directly follow from previous answer
3. **Evidence-based** - Every answer should have supporting data (logs, configs)
4. **Stop at process/system level** - Don't go to "human error" as root; instead ask "Why was the human able to make that error?" → System allowed it
5. **Verify** - Can the root cause be fixed to prevent recurrence? If no, keep asking

---

## Common Pitfalls

| Pitfall | Example | Correction |
|---------|---------|------------|
| **Symptom as root cause** | "API went down because error rate spiked" | Keep asking why error rate spiked |
| **Blame assignment** | "Developer forgot to update config" | Ask why system allowed forgotten config |
| **Multiple roots** | Stopping after one, but 3 unrelated issues exist | Use Fishbone for multi-factor |
| **No evidence** | "Probably network issue" | Check metrics/logs to confirm |
| **Jumping 2 levels** | Q: Why error? A: Pool size (skipping connection timeout) | Ensure each answer directly leads to next why |

---

## Real FlyFleet Example Template

```
### 5 Whys: Database Connection Pool Exhaustion

1. Why did API return 503 Service Unavailable?
   → Health check endpoint timed out after 30 seconds

2. Why did health check time out?
   → Health check queries database for connectivity, but got no connection from pool after 30s wait

3. Why was no database connection available?
   → All 100 connections in PostgreSQL pool were in use and held by active queries

4. Why were all connections used?
   → fleet-core was running 6 replicas × ~20 connections each = 120 connections needed, but max_connections=100

5. Why was replica count increased to 6 without increasing max_connections?
   → HPA scaled based on CPU, but database max_connections was manually set and never updated; no automatic coordination between scaling policies

**Root Cause:**
Uncoupled autoscaling: HorizontalPodAutoscaler increased replica count independently of database capacity planning, with no guardrails to prevent pool exhaustion.

**Evidence:**
- Fleet-core deployment changed from 2 to 6 replicas at 13:45 UTC
- PostgreSQL max_connections=100 (static)
- pgpool_connections metric showed 100/100 from 14:10 UTC
- No alerts on connection pool utilization
```

---

**Method Version:** 1.0.0
**Last Updated:** 2026-03-29
