# Runbook: Security Incident

**Severity:** SEV1 (always)
**Service:** Security-sensitive systems (Auth, API, Database, Network)
**Escalation:** Security Officer + VP Engineering + Legal

---

## Quick Reference

| Incident Type | Immediate Action | Contact |
|---------------|------------------|---------|
| **Data Breach** | Isolate affected systems | Security Officer |
| **Unauthorized Access** | Revoke credentials, block IP | Security Officer |
| **Malware/Ransomware** | Disconnect from network | Security Officer + Infra |
| **DDoS Attack** | Activate mitigations | Infra + AWS Support |
| **Insider Threat** | Preserve logs, disable account | Security Officer + HR |
| **Compliance Violation** | Notify Legal immediately | Legal + Security |

---

## Incident Types

### 1. Data Exposure / Breach
- PII (driver personal info) accessed without authorization
- Financial data (payments, invoices) leaked
- Location data of drivers exposed
- Database dump found on public internet

### 2. Unauthorized Access
- Compromised credentials used (stolen JWT, password)
- Privilege escalation detected
- Role/permission misuse
- Access from unusual geographic locations

### 3. Malware / Ransomware
- Malicious code detected in repository or running in pods
- Ransom note found, files encrypted
- Cryptocurrency miner detected (high CPU)

### 4. Denial of Service
- Application layer DDoS (HTTP flood)
- Network layer DDoS (SYN flood, UDP flood)
- Resource exhaustion attack

### 5. Compliance Violation
- India FASTag/E-Way Bill data mishandled
- GDPR data subject rights not honored
- SOC2 control failure
- Unauthorized data transfer across borders

---

## Decision Tree

```
Security Incident Detected?
├─ Data Breach?
│   ├─ Yes → What data exposed?
│   │   ├─ PII (PII includes driver info, location) → SEV1, notify Legal immediately (GDPR 72h)
│   │   ├─ Financial → SEV1, notify Legal + Finance
│   │   └─ Secrets (API keys, JWT secret) → SEV1, rotate ALL secrets
│   └─ Scope unknown → Isolate systems, preserve evidence
├─ Unauthorized Access?
│   ├─ Compromised credentials? → Force password reset, revoke all tokens
│   ├─ Privilege escalation? → Disable account, audit role assignments
│   └─ Suspicious IP? → Block IP, enable MFA requirement
├─ Malware Detected?
│   ├─ Container compromised? → Quarantine pod, rebuild from clean image
│   ├─ Repository compromised? → Git revert, scan all commits
│   └─ Cryptominer? → Kill process, patch vulnerability
├─ DDoS Attack?
│   ├─ Application layer → Enable WAF rate limiting
│   └─ Network layer → Contact AWS Shield, enable VPC flow logs
└─ Compliance?
    ├─ GDPR violation → Notify DPO within 72h
    ├─ India regulations → Notify compliance team immediately
    └─ SOC2 control failure → Document incident for audit
```

---

## Investigation Steps

### 1. Initial Containment (0-15 min)

**SEV1 Response - ISOLATE IMMEDIATELY**

1. **Assess scope without disturbing evidence**
   - Which systems are affected? Database? API? Kubernetes cluster?
   - What type of data? PII, financial, operational?
   - Is attacker still active? Check for ongoing suspicious activity

2. **Isolate affected systems (carefully)**

   **Option A: Network isolation (most common)**
   ```bash
   # Isolate specific pod
   kubectl label pod compromised-pod-names isolation=quarantine
   # NetworkPolicy blocks all ingress/egress
   kubectl apply -f - <<EOF
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: isolate-compromised
     namespace: flyfleet
   spec:
     podSelector:
       matchLabels:
         isolation: quarantine
     policyTypes:
     - Ingress
     - Egress
   EOF
   ```

   **Option B: Scale down compromised service (preserve logs)**
   ```bash
   kubectl scale deployment compromised-service --replicas=0 -n flyfleet
   # Pod remains, but no new requests
   ```

   **Option C: Disable user accounts**
   ```bash
   kubectl exec -it postgres-0 -n flyfleet -- \
     psql -U fleet_user fleet_db -c \
     "UPDATE users SET is_active = false WHERE email = 'compromised@example.com';"
   ```

3. **Preserve evidence**
   ```bash
   # Snapshot pod logs (before deleting)
   kubectl logs compromised-pod -n flyfleet --since=24h > /evidence/compromised-pod-$(date +%s).log

   # Copy pod filesystem (forensics)
   kubectl cp compromised-pod:/tmp /evidence/pod-filesystem-$(date +%s) -n flyfleet

   # Export database state (read-only query, don't stop DB)
   kubectl exec -it postgres-0 -n flyfleet -- \
     pg_dump -U fleet_user -Fc fleet_db > /evidence/db-dump-$(date +%s).dump

   # Capture container metadata
   kubectl get pod compromised-pod -n flyfleet -o yaml > /evidence/pod-metadata-$(date +%s).yaml
   ```

---

### 2. Evidence Collection (15-60 min)

**Collect comprehensive evidence:**

```bash
# Create evidence directory structure
mkdir -p /evidence/{logs,metrics,config,network,audit}
```

#### Kubernetes Logs
```bash
# All pods in affected namespace
for pod in $(kubectl get pods -n flyfleet -o jsonpath='{.items[*].metadata.name}'); do
  kubectl logs $pod -n flyfleet --since=72h > /evidence/logs/$pod.log
done

# Previous container logs (if restarted)
for pod in $(kubectl get pods -n flyfleet -o jsonpath='{.items[*].metadata.name}'); do
  kubectl logs $pod -n flyfleet --previous --since=72h > /evidence/logs/$pod-previous.log 2>/dev/null || true
done
```

#### Audit Logs
```bash
# Kubernetes audit logs (if enabled)
kubectl logs -n kube-system -l k8s-app=aws-alb-ingress-controller --since=72h > /evidence/audit/k8s-audit.log 2>/dev/null || true

# Auth service audit (SQL queries for sensitive actions)
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT * FROM audit_logs
    WHERE created_at > NOW() - INTERVAL '72 hours'
      AND action IN ('login', 'password_change', 'role_change', 'data_export', 'admin_action')
    ORDER BY created_at DESC;
  " > /evidence/audit/auth-audit.csv
```

#### Network Traffic
```bash
# Enable VPC Flow Logs for affected subnets (AWS)
aws ec2 create-flow-logs \
  --resource-type Subnet \
  --resource-ids subnet-xxx subnet-yyy \
  --traffic-type ALL \
  --log-destination-type cloud-watch-logs \
  --log-group-name flyfleet-security-audit

# Capture NATS messages (subject to privacy laws)
# DANGER: May contain PII. Only if authorized and necessary.
# kubectl exec -it nats-0 -n flyfleet -- nats stream report --json > /evidence/nats-streams.json
```

#### Configuration
```bash
# Deployment configs
kubectl get deployment,configmap,secret,service -n flyfleet -o yaml > /evidence/config/k8s-resources.yaml

# Helm values
cat infrastructure/helm/flyfleet/values.yaml > /evidence/config/helm-values.yaml
```

---

### 3. Root Cause Analysis (60-120 min)

**Common Security Incident Investigation:**

#### Incident Type: Compromised Credentials

**Check:**
- Where did login originate? (IP, user agent, time)
- Was MFA bypassed? (Check MFA logs)
- Was this credential stuffing? (Multiple rapid failures then success)
- Password reuse? (Check if password in breach databases)

```bash
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT email, last_login_at, last_login_ip, mfa_enabled
    FROM users
    WHERE email = 'compromised@example.com';
  "

# Check login history from same IP
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    SELECT email, created_at FROM audit_logs
    WHERE action = 'login' AND ip_address = 'attacker.ip'
      AND created_at > NOW() - INTERVAL '24 hours'
    ORDER BY created_at DESC;
  "
```

**Actions:**
- Revoke all active tokens for user
- Force password reset (token sent to verified email/phone)
- Block attacker IP at load balancer/WAF
- Check if attacker accessed other accounts (same IP pattern)

---

#### Incident Type: API Key / Secret Leak

**Check:**
- Where was secret exposed? (GitHub? Logs? Client-side?)
- When was it first misused?
- How many requests made with compromised secret?

```bash
# Search Git history for leaked secret (if suspect repo leak)
git log -p --all -S 'YOUR_SECRET_KEY' | head -100

# Check deployment configs
kubectl get secret -n flyfleet -o yaml | grep -A5 "API_KEY"

# Rotate secret in AWS Secrets Manager
aws secretsmanager rotate-secret --secret-id flyfleet/production/api-keys

# Generate new secret
aws secretsmanager get-secret-value --secret-id flyfleet/production/api-keys-new
```

**Actions:**
1. Rotate secret immediately (create new, deploy updated)
2. Invalidate all old tokens (add to blocklist)
3. Review access logs for misuse
4. Add alert on secret usage from unknown IPs

---

#### Incident Type: Malware / Cryptominer

**Symptom:** High CPU on unrelated pods, suspicious processes

```bash
# Check running processes in all pods
for pod in $(kubectl get pods -n flyfleet -o jsonpath='{.items[*].metadata.name}'); do
  echo "=== $pod ==="
  kubectl exec -it $pod -n flyfleet -- ps aux 2>/dev/null | grep -E "(xmrig|cryptonight|miner)" || echo "No miner"
done

# Check for suspicious network connections
for pod in $(kubectl get pods -n flyfleet -o jsonpath='{.items[*].metadata.name}'); do
  echo "=== $pod ==="
  kubectl exec -it $pod -n flyfleet -- netstat -tulpn 2>/dev/null | grep -E "(mining|pool|stratum)" || echo "No suspicious ports"
done

# Check container images for known malware hash
crane manifest flyfleet/fleet-core:latest 2>/dev/null | grep -i "xmrig\|malware" || echo "Clean"
```

**Actions:**
1. **Immediate:** Delete compromised pod, force new deployment from clean image
   ```bash
   kubectl delete pod <compromised-pod> -n flyfleet
   # Deployment will recreate with same image spec
   ```
2. **If image itself compromised:** Rebuild from clean base, scan all images
   ```bash
   trivy image flyfleet/fleet-core:latest
   # If detections, rebuild with updated base image
   ```
3. **Check for persistence:** CronJobs, DaemonSets with malicious code
   ```bash
   kubectl get cronjob -n flyfleet
   kubectl get daemonset -n flyfleet
   ```
4. **Rotate all secrets** that pod had access to (secrets mounted)

---

#### Incident Type: DDoS Attack

**Symptoms:** High traffic from many IPs, API latency spike, 503 errors

```bash
# Check ingress/load balancer metrics
kubectl top pods -n ingress-nginx
kubectl logs -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx --tail=100 | grep -i "limiting\|rate"

# Check source IP distribution
kubectl logs -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx --since=5m | \
  awk '{print $1}' | sort | uniq -c | sort -nr | head -20

# AWS WAF if configured
aws wafv2 get-web-acl --name flyfleet-waf --scope CLOUDFRONT
```

**Mitigations:**

1. **Enable AWS Shield Advanced** (if subscription)
   ```bash
   aws shield enable-protection --name flyfleet-production --resource-arn arn:aws:elasticloadbalancing:...
   ```

2. **Rate limiting at ingress**
   ```yaml
   # Add to ingress-nginx ConfigMap
   rate-limiting:
     enabled: true
     requests-per-second: 100
     burst: 200
   ```

3. **WAF Rules** (if using AWS WAF)
   - Rate-based rule: block IPs with > 1000 requests/5min
   - IP block list for known bad actors
   - Geo-block if attack from specific country

4. **Scale horizontally**
   ```bash
   kubectl scale deployment fleet-core --replicas=10 -n flyfleet
   kubectl autoscale deployment fleet-core --cpu-percent=80 --min=5 --max=20 -n flyfleet
   ```

5. **Enable caching** (reduce origin load)
   ```bash
   kubectl annotate ingress flyfleet-ingress nginx.ingress.kubernetes.io/configuration-snippet='proxy_cache_path /tmp/nginx levels=1:2 keys_zone=my_cache:10m inactive=60m;'
   ```

---

## Communication & Escalation

### Immediate Notifications (SEV1 - Security)

**Must notify within 15 minutes:**

| Role | Contact | When |
|------|---------|------|
| Security Officer | [phone] / @security-slack | Immediately |
| VP Engineering | [phone] | Immediately |
| Legal/Compliance | [phone] | Immediately (PII breach) |
| CEO | [phone] | If customer data or revenue impact |
| PR/Communications | [email] | If public disclosure needed |
| Customer Support Lead | [slack] | To prepare for customer inquiries |

**Template for security incident notification:**

```
[SEV1-SECURITY] Security Incident: {Brief Description}

Time Detected: {timestamp UTC}
Type: {Data Breach / Unauthorized Access / Malware / DDoS / Compliance}
Scope: {affected systems, estimated data volume}
Data Types: {PII / Financial / Location / None}
Custodian: {Incident Commander Name}

Actions Taken:
- [ ] Systems isolated
- [ ] Evidence preserved
- [ ] Investigation in progress

Next Update: {timestamp}

DO NOT DISCUSS EXTERNALLY. Refer all questions to Security Officer.
```

---

### Legal & Compliance Requirements

#### GDPR (EU) - Data Breach

Notify Data Protection Officer **within 72 hours** if:
- PII of EU residents affected
- Unauthorized access, loss, or alteration
- Likely to result in risk to rights and freedoms

**Notification must include:**
- Nature of breach
- Approximate number of individuals affected
- Categories of personal data compromised
- Likely consequences
- Measures taken to contain and mitigate

#### India Compliance (FlyFleet India)

If FASTag/E-Way Bill data affected:
- **Immediate** notification to GST Authority (within 24h)
- Notify affected transporters/drivers
- Review data localization compliance (data stored in India?)

#### SOC2 (Trust Services Criteria)

Document:
- Security incident timeline
- Root cause
- Impact assessment
- Remediation actions
- Customer notification (if material)

---

### Customer Notification

**When required:**
- PII exposed
- Financial data compromised
- Service disruption with security implications
- Legal requirement (GDPR, CCPA, etc.)

**Process:**
1. Draft notification with Legal/PR approval
2. Send to affected customers via email (and SMS if urgent)
3. Publish on status page (if broad impact)
4. Offer support line for questions

**Template:**
```
Subject: Security Incident Notification - FlyFleet

Dear {Customer},

We are writing to inform you of a security incident that may have affected your data.

Incident Summary:
{1-2 sentence description}

Data Potentially Affected:
- {List data types: name, email, location, financial, etc.}

What We're Doing:
- {Containment actions}
- {Investigation steps}
- {Notification authorities}

What You Should Do:
- Change your password
- Enable MFA (if not already)
- Monitor accounts for suspicious activity
- Contact support if you notice anything unusual

Timeline:
Detected: {date}
Containment: {date}
Ongoing: Investigation and remediation

Contact:
Security Incident Hotline: {phone}
Email: security@flyfleet.com

We apologize for this incident and are taking all necessary steps to prevent recurrence.

Sincerely,
FlyFleet Security Team
```

---

## Post-IncidentActions

### PIR for Security Incidents

Additional PIR sections beyond standard:

1. **Threat Actor Assessment**
   - External vs. internal
   - Sophistication level (script kiddie vs. APT)
   - Motivation (financial, espionage, hacktivism, insider)

2. **Attack Vector**
   - Phishing / credential theft
   - Vulnerability exploitation (CVE)
   - Misconfiguration
   - Supply chain compromise
   - Insider threat

3. **Data Assets Affected**
   - Types of data (PII, financial, operational)
   - Number of records (exact or estimate)
   - Data sensitivity classification
   - Could data be used for fraud/identity theft?

4. **Impact Assessment**
   - Could data be reconstructed?
   - Was data exfiltrated? (How do we know?)
   - Was data encrypted (at rest, in transit)?
   - Financial loss (fraud, fines, lawsuits)

5. **Remediation Steps**
   - Access revocations (tokens, API keys, certificates)
   - System rebuilds (from clean images)
   - Configuration fixes (principle of least privilege)
   - Secret rotations (all credentials)
   - Monitoring enhancements (detection of similar attacks)

6. **Legal/Compliance Outcomes**
   - Authorities notified (with dates)
   - Customer notifications sent (count)
   - Fines imposed or expected
   - Insurance claim filed

---

### Security Hardening Actions

**Common improvements after security incident:**

1. **Access Control**
   - Implement MFA for all admin access
   - Enforce least privilege (RBAC audit)
   - Add IP allowlisting for sensitive systems
   - Implement Just-In-Time access (temporary elevated privileges)

2. **Monitoring & Detection**
   - Add audit logging for all sensitive operations
   - Alert on unusual patterns (impossible travel, bulk exports)
   - Deploy SIEM (Splunk, ELK, Datadog) for correlation
   - Enable VPC Flow Logs, CloudTrail, S3 access logs

3. **Secret Management**
   - Rotate all secrets (even uninvolved ones)
   - Move to AWS Secrets Manager (no env vars)
   - Add secret scanning in CI/CD (GitGuardian, truffleHog)
   - Implement automatic rotation (90 days)

4. **Software Supply Chain**
   - Scan all container images (Trivy, Clair)
   - Sign images with cosign, enforce verification
   - Implement SBOM (Software Bill of Materials)
   - Use only official base images, minimal

5. **Network Security**
   - Network policies (default deny, explicit allow)
   - TLS everywhere (mTLS for service-to-service)
   - WAF rules (OWASP Top 10 protection)
   - DDoS protection (AWS Shield Advanced)

6. **Data Protection**
   - Encrypt all data at rest (RDS, S3, EBS)
   - Column-level encryption for PII
   - Data classification and tagging
   - Data loss prevention (DLP) scanning

---

## Prevention & Monitoring

### Security Monitoring Rules

```yaml
# Prometheus alerts for security events

- alert: FailedLogin Spike
  expr: rate(auth_login_attempts_total{status="invalid_credentials"}[5m]) > 100
  for: 2m
  labels:
    severity: warning
    category: security
  annotations:
    summary: "High volume of failed logins detected"
    description: "{{ $value }} failed logins in last 5 minutes"

- alert: Impossible Travel
  expr: avg_over_time(user_login_location_change_km[1h]) > 1000
  for: 5m
  labels:
    severity: critical
    category: security
  annotations:
    summary: "User login from geographically distant locations"

- alert: Privilege Escalation
  expr: changes(audit_role_changes_total[5m]) > 0
  for: 1m
  labels:
    severity: critical
    category: security
  annotations:
    summary: "Role change detected for user"

- alert: Data Exfiltration
  expr: sum(rate(data_export_bytes_total[5m])) by (user_id) > 100000000  # 100MB
  for: 5m
  labels:
    severity: critical
    category: security
  annotations:
    summary: "Large data export detected"
```

---

### Audit Logging Requirements

All services MUST log:

1. **Authentication Events**
   - Login success/failure
   - Logout
   - Password change
   - MFA enrollment/challenge

2. **Authorization Events**
   - Role/permission changes
   - Access to sensitive data
   - Admin actions

3. **Data Access**
   - Export of PII/financial data
   - Deletion of records
   - Configuration changes

4. **Infrastructure**
   - kubectl commands (audit logs)
   - Secret access (AWS CloudTrail)
   - Database queries (if PII accessed)

**Log format:**
```json
{
  "timestamp": "2026-03-29T10:30:00Z",
  "event_type": "data_export",
  "user_id": "user-123",
  "user_email": "user@example.com",
  "ip_address": "192.168.1.1",
  "user_agent": "curl/7.68.0",
  "resource_type": "vehicles",
  "record_count": 150,
  "action": "export",
  "org_id": "org-123"
}
```

---

## Evidence Handling & Chain of Custody

For legal/compliance:

1. **Document** every action taken during incident
2. **Preserve** logs in read-only storage (WORM)
3. **Hash** evidence files for integrity verification
   ```bash
   sha256sum evidence.tar.gz > evidence.tar.gz.sha256
   ```
4. **Record** who accessed evidence and when
5. **Store** in secure location (encrypted, access controlled)
6. **Retain** for at least 7 years (SOC2) or as legal requires

---

## Emergency Contacts

| Role | Name | Phone | Slack |
|------|------|-------|-------|
| Security Officer | TBD | [phone] | @security |
| VP Engineering | TBD | [phone] | @vp-engineering |
| Legal/Compliance | TBD | [phone] | @legal |
| AWS Support | 24/7 | 1-888-AWS-HELP | - |
| External IR Firm | [vendor] | [phone] | - |

---

## Related Runbooks

- [Authentication Failure](auth-failure.md) - If compromise via auth system
- [Data Inconsistency](data-inconsistency.md) - For data corruption (non-security)
- [API Outage](api-outage.md) - If DDoS causes outage

---

**Last Updated:** 2026-03-29
**Review Date:** 2026-06-29
**Owner:** Security Team
