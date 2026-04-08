# Runbook: Authentication Failure

**Severity:** SEV1 (if 100% auth failing) or SEV2 (partial degradation)
**Service:** Auth Service (JWT, MFA, RBAC)
**SLO Violation:** Auth Error Rate < 0.1%, Latency < 200ms
**Alert:** `AuthServiceHighErrorRate` or `AuthServiceLatencyHigh`

---

## Quick Reference

| What | How |
|------|-----|
| **Check auth service health** | `curl http://auth-service:8080/health` |
| **Check auth error rate** | `curl -s 'http://prometheus:9090/api/v1/query?query=rate(auth_requests_total{status=~"5.."}[5m]) / rate(auth_requests_total[5m])'` |
| **Check JWT secret** | `kubectl get secret auth-secrets -n flyfleet -o jsonpath='{.data.JWT_SECRET}' \| base64 -d` |
| **Restart auth service** | `kubectl rollout restart deployment/auth-service -n flyfleet` |
| **Check Redis sessions** | `kubectl exec -it redis-0 -n flyfleet -- redis-cli ping` |
| **Check RBAC permissions** | `kubectl auth can-i --list --as=system:serviceaccount:flyfleet:auth-service` |

---

## Symptoms

- Users unable to log in (401/403 errors)
- Token validation failing (all requests rejected)
- MFA not working (code rejected)
- Password reset emails not sending
- Session timeouts unexpected
- "Invalid credentials" despite correct password
- "Token expired" immediately after login

---

## Decision Tree

```
Auth Failing?
├─ Login fails (401)?
│   ├─ All users or specific?
│   │   ├─ All → JWT secret invalid?
│   │   │   ├─ Yes → Restore JWT_SECRET from backup
│   │   │   └─ No → Database auth issue?
│   │   │       ├── Yes → Database runbook
│   │   │       └── No → Check auth-service logs
│   │   └─ Specific → User account locked?
│   │       └── Check: kubectl exec... -c "SELECT locked FROM users WHERE email='...'"
│   ├─ MFA failing?
│   │   ├─ TOTP sync issue? → Check time sync, secret key
│   │   └── SMTP issues? → Check email service credentials
│   └─ LDAP/OAuth failing? → Check external provider status
├─ Token validation fails (403)?
│   ├─ Token expired immediately? → Check JWT exp claim, clock skew
│   ├─ Invalid signature? → JWT secret mismatch
│   └─ Missing permissions? → RBAC misconfiguration, check roles
└─ High latency?
    ├─ Database slow? → Database runbook
    ├─ Redis slow? → Check Redis, restart
    └─ CPU high on auth? → Scale auth-service
```

---

## Investigation Steps

### 1. Check Auth Service Health (2 min)

```bash
# Pod status
kubectl get pods -n flyfleet -l app=auth-service

# Health endpoint
kubectl exec -it deployment/auth-service -n flyfleet -- \
  curl -s http://localhost:8080/health

# Readiness/liveness probes
kubectl describe deployment/auth-service -n flyfleet | grep -A5 "Liveness"

# Recent events
kubectl get events -n flyfleet --field-selector involvedObject.kind=Pod,involvedObject.label.app=auth-service --sort-by='.lastTimestamp' | tail -20

# Resource utilization
kubectl top pods -n flyfleet -l app=auth-service
```

---

### 2. Check Auth Metrics (2 min)

```bash
# Error rate
curl -s 'http://prometheus:9090/api/v1/query?query=rate(auth_requests_total{status=~"5..",service="auth"}[5m]) / rate(auth_requests_total{service="auth"}[5m])' | jq

# Latency
curl -s 'http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95, sum(rate(auth_request_duration_seconds_bucket[5m])) by (le))' | jq

# Login success/failure breakdown
curl -s 'http://prometheus:9090/api/v1/query?query=sum(rate(auth_login_attempts_total[5m])) by (status)' | jq

# MFA verification success rate
curl -s 'http://prometheus:9090/api/v1/query?query=rate(mfa_verification_attempts_total{result="success"}[5m]) / rate(mfa_verification_attempts_total[5m])' | jq
```

---

### 3. Check Auth Service Logs (3 min)

```bash
# Tail recent logs
kubectl logs -f deployment/auth-service -n flyfleet --since=10m

# Filter errors
kubectl logs -f deployment/auth-service -n flyfleet --since=10m 2>&1 | grep -i "error\|failed\|panic"

# Check JWT verification failures
kubectl logs -f deployment/auth-service -n flyfleet --since=10m 2>&1 | grep -i "jwt\|token\|signature"

# Check database connectivity errors
kubectl logs -f deployment/auth-service -n flyfleet --since=10m 2>&1 | grep -i "database\|postgres\|connection"

# Check MFA related errors
kubectl logs -f deployment/auth-service -n flyfleet --since=10m 2>&1 | grep -i "mfa\|totp\|code"
```

**Common patterns:**
- `token expired` - JWT exp/nbf validation, clock skew
- `signature invalid` - JWT_SECRET mismatch
- `user not found` - Database connectivity or user record issue
- `mfa invalid` - TOTP secret mismatch or clock skew
- `context deadline exceeded` - Database or Redis slow

---

### 4. Check Secrets (2 min)

```bash
# Check auth secrets exist
kubectl get secret auth-secrets -n flyfleet

# Verify JWT_SECRET (do not print full secret!)
kubectl get secret auth-secrets -n flyfleet -o jsonpath='{.data.JWT_SECRET}' | wc -c
# Should be non-zero

# If secret missing or rotated recently, check secrets manager
aws secretsmanager get-secret-value --secret-id flyfleet/production/auth 2>&1 | head -5

# Check ExternalSecret sync status (if using ESO)
kubectl get externalsecret -n flyfleet auth-secrets
kubectl describe externalsecret auth-secrets -n flyfleet | grep -A5 "Status:"
```

**If JWT_SECRET changed recently:**
- All existing tokens become invalid
- Users will be logged out
- May cause mass login issues
- Need to communicate outage to users

---

### 5. Check Database (Auth DB) (2 min)

```bash
# Check auth database connectivity
kubectl exec -it deployment/auth-service -n flyfleet -- \
  psql $DATABASE_URL -c "SELECT 1" 2>&1

# Check auth-specific table health
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT count(*) FROM users WHERE is_active = true;"

# Check for locks on auth tables
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT * FROM pg_locks WHERE relation = 'users'::regclass AND NOT granted;"

# Check recent failed login attempts (for lockouts)
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT email, failed_login_attempts, locked_until FROM users WHERE locked_until > NOW() ORDER BY locked_until DESC LIMIT 10;"
```

---

### 6. Check Redis (Session Store) (2 min)

```bash
# Redis health
kubectl exec -it redis-0 -n flyfleet -- redis-cli ping

# Check memory usage
kubectl exec -it redis-0 -n flyfleet -- redis-cli info memory | grep used_memory_human

# Check keyspace (sessions)
kubectl exec -it redis-0 -n flyfleet -- redis-cli info keyspace

# Check if Redis reachable from auth-service
kubectl exec -it deployment/auth-service -n flyfleet -- \
  redis-cli -h $REDIS_HOST ping

# Check auth-service logs for Redis errors
kubectl logs -f deployment/auth-service -n flyfleet --since=10m 2>&1 | grep -i "redis\|session"
```

---

### 7. Check External Identity Providers (2 min)

If using LDAP, OAuth, SAML:

```bash
# Check ExternalIdentity configuration
kubectl get configmap auth-config -n flyfleet -o yaml | grep -A5 "ldap\|oauth\|saml"

# Test LDAP connection
kubectl exec -it deployment/auth-service -n flyfleet -- \
  ldapsearch -x -H ldap://ldap.example.com -b "dc=example,dc=com" -D "cn=admin,dc=example,dc=com" -w "$LDAP_PASSWORD" "(objectClass=person)" 2>&1 | head -10

# Check OAuth provider status (GitHub, Google, etc.)
# Visit provider status page
# - GitHub: https://www.githubstatus.com/
# - Google: https://www.google.com/appsstatus
```

---

## Mitigation Strategies

### Strategy 1: Restart Auth Service

**When:** Service crashed or unresponsive

```bash
kubectl rollout restart deployment/auth-service -n flyfleet
kubectl rollout status deployment/auth-service -n flyfleet --timeout=5m

# Check logs for startup completion
kubectl logs -f deployment/auth-service -n flyfleet | grep -i "listening\|started\|ready"
```

---

### Strategy 2: Fix JWT Secret Mismatch

**When:** Signature invalid errors, token validation failing for all users

```bash
# Check current secret
kubectl get secret auth-secrets -n flyfleet -o yaml

# Retrieve correct secret from AWS Secrets Manager
aws secretsmanager get-secret-value --secret-id flyfleet/production/auth --query SecretString --output json | jq -r '.JWT_SECRET' > jwt-secret.txt

# Create/update Kubernetes secret
kubectl create secret generic auth-secrets \
  --from-literal=JWT_SECRET="$(cat jwt-secret.txt)" \
  --dry-run=client -o yaml | kubectl apply -f -

# OR if using ExternalSecret, ensure sync is successful
kubectl get externalsecret auth-secrets -n flyfleet
# If status shows error, check ESO logs

# Restart auth service to pick up new secret
kubectl rollout restart deployment/auth-service -n flyfleet
```

---

### Strategy 3: Reset Account Lockouts

**When:** Specific users locked out due to failed attempts

```bash
# Unlock specific user
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    UPDATE users
    SET locked_until = NULL,
        failed_login_attempts = 0
    WHERE email = 'user@example.com';
  "

# Unlock all users temporarily (emergency)
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "
    UPDATE users
    SET locked_until = NULL,
        failed_login_attempts = 0
    WHERE is_active = true;
  "
```

**After unlock:** Monitor for suspicious activity (could be attack)

---

### Strategy 4: Clear Redis Sessions

**When:** Session issues, tokens not being invalidated properly

```bash
# Check Redis keys
kubectl exec -it redis-0 -n flyfleet -- redis-cli keys "*"

# Flush all sessions (logs out all users) - EMERGENCY ONLY
kubectl exec -it redis-0 -n flyfleet -- redis-cli FLUSHDB
# OR flush specific pattern
kubectl exec -it redis-0 -n flyfleet -- redis-cli --scan --pattern "session:*" | xargs -L 1000 redis-cli del

# Restart auth service to clear in-memory cache
kubectl rollout restart deployment/auth-service -n flyfleet
```

**Impact:** All users logged out. Communicate before executing.

---

### Strategy 5: Scale Auth Service

**When:** High latency, auth-service CPU/memory saturated

```bash
# Check current replicas
kubectl get deployment auth-service -n flyfleet

# Scale to 3 replicas
kubectl scale deployment auth-service --replicas=3 -n flyfleet

# Check HPA if configured
kubectl get hpa -n flyfleet | grep auth-service

# Ensure service can handle multiple replicas (stateless)
# Verify session store is in Redis, not in-memory
```

---

### Strategy 6: Database Connection Pool Increase

**When:** Database connection errors, "too many connections" in logs

```bash
# Increase max_connections in postgres (see database-performance.md)
# Increase auth-service DB pool
kubectl edit deployment auth-service -n flyfleet

# Add/modify env:
# - name: DATABASE_MAX_OPEN_CONNS
#   value: "50"  # default 25
# - name: DATABASE_MAX_IDLE_CONNS
#   value: "25"

kubectl rollout restart deployment/auth-service -n flyfleet
```

---

## MFA-Specific Issues

### TOTP Time Sync

**Symptom:** MFA codes rejected, "invalid code" error

**Fix:**
1. Check server time is synced (NTP)
```bash
kubectl exec -it deployment/auth-service -n flyfleet -- date
# Should be within 30 seconds of real time
```

2. Check MFA secret key stored correctly:
```bash
kubectl exec -it postgres-0 -n flyfleet -- \
  psql -U fleet_user fleet_db -c "SELECT email, mfa_secret IS NOT NULL AS has_mfa FROM users WHERE email = 'user@example.com';"
```

3. Allow time skew (if intentional):
   - Increase TOTP window from default ±30s to ±60s in config

---

### SMS/Email MFA Delivery

**Symptom:** MFA codes not arriving via SMS or email

**Check:**
```bash
# Check SMTP configuration (email)
kubectl get configmap auth-config -n flyfleet -o yaml | grep -A10 "smtp_"

# Check Twilio/other SMS provider credentials
kubectl get secret auth-secrets -n flyfleet -o jsonpath='{.data.TWILIO_SID}'

# Check auth-service logs for delivery errors
kubectl logs -f deployment/auth-service -n flyfleet --since=10m 2>&1 | grep -i "smtp\|twilio\|sms\|delivery"
```

**Fix:**
- Verify SMTP/SMS provider is reachable
- Check credentials not expired
- Verify message quota not exceeded
- Retry: Most providers have rate limits

---

## Recovery & Validation

### Post-Mitigation Checks

1. **Health Check**
```bash
kubectl exec -it deployment/auth-service -n flyfleet -- \
  curl -s http://localhost:8080/health
# Should return 200
```

2. **Login Test**
```bash
# Test login with known credentials
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"correct_password"}' \
  -w "\nHTTP Status: %{http_code}\n"

# Should return 200 with JWT token
```

3. **Token Validation**
```bash
# Extract token from login response
TOKEN="eyJ..."

# Test protected endpoint
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/profile -w "\nHTTP Status: %{http_code}\n"

# Should return 200 with user profile
```

4. **Error Rate Back to Normal**
```bash
curl -s 'http://prometheus:9090/api/v1/query?query=rate(auth_requests_total{status=~"5..",service="auth"}[5m]) / rate(auth_requests_total{service="auth"}[5m])' | jq
# Should be < 0.001 (0.1%)
```

5. **Latency Normal**
```bash
curl -s 'http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95, sum(rate(auth_request_duration_seconds_bucket[5m])) by (le))' | jq
# Should be < 0.2s (200ms)
```

---

## Security Incident Response

If auth incident involves security breach:

### Signs of Compromise
- Unusual login patterns (geographically impossible travel)
- Brute force attacks (many failed logins from single IP)
- Credential stuffing detected
- MFA bypass attempts
- Privilege escalation attempts

### Response
1. **IMMEDIATE:** Enable IP-based rate limiting
```bash
kubectl exec -it deployment/auth-service -n flyfleet -- \
  curl -X POST http://localhost:8080/admin/security/block-ip -d '{"ip":"attacker.ip"}'
```

2. **Force password reset for affected users**

3. **Invalidate all active sessions**
```bash
kubectl exec -it redis-0 -n flyfleet -- redis-cli FLUSHDB
kubectl rollout restart deployment/auth-service -n flyfleet
```

4. **Enable emergency MFA requirement**
```bash
# Update config to require MFA for all users
kubectl edit configmap auth-config -n flyfleet
# Set: mfa_required_for_all: "true"
kubectl rollout restart deployment/auth-service -n flyfleet
```

5. **Escalate to Security Officer** immediately (SEV1)

---

## Post-Incident Analysis

### PIR Focus Areas

1. **Root Cause Categories:**
   - Secret rotation without coordinated rollout
   - Database connectivity interruption
   - Redis cache failure or eviction
   - External provider outage (LDAP, OAuth)
   - Clock skew causing JWT exp/nbf failures
   - Code bug (password hash mismatch, token parsing error)
   - Configuration error (Auth0/Okta settings changed)

2. **Preventive Actions:**
   - Implement secret versioning with automated rollout (gradual)
   - Add health checks for external providers
   - Implement graceful degradation (allow cached tokens if DB down)
   - Add alert on auth error rate > 0.1%
   - Set up monitoring for JWT signature verification failures
   - Implement canary deployments for auth-service (small % first)
   - Add audit logging for all auth events (success/failure)
   - Implement account lockout monitoring (detect attacks)

3. **Security Enhancements:**
   - Rate limiting per IP/account
   - Account lockout after 5 failed attempts
   - Alert on impossible travel (logins from distant locations in short time)
   - IP whitelisting for admin accounts
   - Session timeout configuration audit

---

## Prevention & Monitoring

### Essential Alerts

```yaml
- alert: AuthServiceDown
  expr: up{service="auth"} == 0
  for: 1m

- alert: AuthErrorRateHigh
  expr: rate(auth_requests_total{status=~"5..",service="auth"}[5m]) / rate(auth_requests_total{service="auth"}[5m]) > 0.001
  for: 5m

- alert: AuthLatencyHigh
  expr: histogram_quantile(0.95, sum(rate(auth_request_duration_seconds_bucket[5m])) by (le)) > 0.2
  for: 5m

- alert: AuthJWTSecretMissing
  expr: length(auth_jwt_secret_bytes) == 0
  for: 1m

- alert: MFAFailureRateHigh
  expr: rate(mfa_verification_attempts_total{result="failure"}[5m]) / rate(mfa_verification_attempts_total[5m]) > 0.05
  for: 5m

- alert: FailedLogin Spike
  expr: rate(auth_login_attempts_total{status="invalid_credentials"}[5m]) > 100
  for: 2m
```

### Dashboard Panels

1. Auth error rate (gauge)
2. Login success vs failure (stacked area)
3. Active sessions in Redis (gauge)
4. MFA verification rate (gauge)
5. Auth service latency P95 (timeseries)
6. JWT verification failures (counter)

---

## Related Runbooks

- [Security Incident](security-incident.md) - For suspected breaches
- [API Outage](api-outage.md) - Auth often causes API failures when failing
- [Database Performance](database-performance.md) - Auth database issues

---

**Last Updated:** 2026-03-29
**Review Date:** 2026-06-29
**Owner:** Security & Identity Team
