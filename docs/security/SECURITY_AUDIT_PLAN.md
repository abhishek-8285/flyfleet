# FlyFleet Authentication System - Security Audit Plan

## Executive Summary
This document outlines the comprehensive security audit plan for the FlyFleet Enterprise Authentication System (auth_v2). The audit will validate JWT security, refresh token rotation, brute force protection, MFA implementation, API key management, session security, rate limiting, database integrity, observability, and frontend integration.

## Audit Scope
- **Service**: fleet-core-service-legacy (auth_v2 module)
- **Database**: PostgreSQL (users, refresh_tokens, api_keys, mfa_recovery_codes, auth_audit_logs)
- **Authentication**: JWT (RSA-256), Argon2id password hashing, TOTP MFA
- **Authorization**: RBAC with role-permission mapping

---

## PHASE 1 — JWT Security Tests

### Test Cases
1. **Expired Token Usage**
   - Generate token with 1-second expiry
   - Wait 2 seconds
   - Attempt API call with expired token
   - Expected: 401 Unauthorized

2. **Tampered JWT Signature**
   - Generate valid token
   - Modify payload (change user_id)
   - Re-sign with different key
   - Expected: 401 Unauthorized

3. **Wrong Issuer Claim**
   - Generate token with issuer="WrongIssuer"
   - Expected: 401 Unauthorized

4. **Wrong Audience Claim**
   - Generate token with audience="wrong-audience"
   - Expected: 401 Unauthorized

5. **Replay Attack (Old Token)**
   - Generate and use token
   - Revoke user session
   - Attempt to reuse same token
   - Expected: 401 Unauthorized

### Validation Points
- [ ] No backend panic/crash
- [ ] Audit log entry created for each failure
- [ ] Proper error messages (no information leakage)

---

## PHASE 2 — Refresh Token Security

### Test Cases
1. **Token Rotation**
   - Login → obtain refresh_token_1
   - Use refresh_token_1 → obtain refresh_token_2
   - Attempt to reuse refresh_token_1
   - Expected: 401, all sessions invalidated

2. **Reuse Detection**
   - Simulate concurrent refresh attempts with same token
   - Expected: First succeeds, second triggers security alert

3. **Session Invalidation**
   - Revoke refresh token
   - Attempt to use associated access token
   - Expected: Access token still valid until expiry (15 min)

### Validation Points
- [ ] Old refresh token rejected
- [ ] Security event logged (REFRESH_TOKEN_REUSE_DETECTED)
- [ ] All user sessions revoked on reuse detection

---

## PHASE 3 — Brute Force Protection

### Test Cases
1. **Account Lockout**
   - 50 failed login attempts with wrong password
   - Expected: Account locked after 5 attempts
   - Lock duration: 15 minutes

2. **Lockout State**
   - Verify "account_locked" error message
   - Expected: 403 Forbidden with "account_locked"

3. **Recovery After Lockout**
   - Wait for lockout period to expire
   - Attempt login with correct password
   - Expected: Successful login

### Validation Points
- [ ] Failed login attempts counter incremented
- [ ] Account locked after MaxLoginAttempts (5)
- [ ] Lockout duration enforced (15 minutes)
- [ ] Security event logged

---

## PHASE 4 — MFA Security

### Test Cases
1. **Invalid OTP**
   - Enable MFA for user
   - Submit wrong 6-digit code
   - Expected: 401 Unauthorized

2. **Expired OTP**
   - Generate OTP, wait 60 seconds (2 periods)
   - Submit expired code
   - Expected: 401 Unauthorized

3. **Replayed OTP**
   - Submit same valid OTP twice
   - Expected: Second attempt fails

4. **MFA Bypass Attempt**
   - Attempt to access protected endpoint without MFA verification
   - Expected: 401 Unauthorized

5. **Recovery Code Usage**
   - Use valid recovery code
   - Verify code marked as used
   - Attempt to reuse same recovery code
   - Expected: Second attempt fails

### Validation Points
- [ ] Invalid codes rejected
- [ ] Expired codes rejected
- [ ] Recovery codes single-use
- [ ] Audit log entries for MFA events

---

## PHASE 5 — API Key Security

### Test Cases
1. **Key Revocation**
   - Create API key
   - Use key successfully
   - Revoke key
   - Attempt to use revoked key
   - Expected: 401 Unauthorized

2. **Key Expiration**
   - Create key with 1-second expiry
   - Wait 2 seconds
   - Attempt to use expired key
   - Expected: 401 Unauthorized

3. **Scope Restrictions**
   - Verify API keys have proper scope/permissions
   - Attempt to access unauthorized resource
   - Expected: 403 Forbidden

### Validation Points
- [ ] Revoked keys rejected
- [ ] Expired keys rejected
- [ ] Audit logs for key creation/revocation
- [ ] Last used timestamp updated

---

## PHASE 6 — Session Security

### Test Cases
1. **Multi-Device Login**
   - Login from Device A
   - Login from Device B
   - Verify both sessions active

2. **Admin Session Revocation**
   - Admin revokes Device A session
   - Device A attempts API call
   - Expected: Session invalidated

3. **Session List**
   - Verify session list shows all active sessions
   - Include IP, User-Agent, Last Used

### Validation Points
- [ ] Multiple sessions supported
- [ ] Admin can revoke specific sessions
- [ ] Revoked sessions immediately invalidated
- [ ] Session metadata tracked

---

## PHASE 7 — Rate Limiting

### Test Cases
1. **Login Rate Limit**
   - 100 login attempts per second from same IP
   - Expected: 429 Too Many Requests after 5 attempts/minute

2. **Registration Rate Limit**
   - 100 registration attempts per second
   - Expected: 429 after 3 attempts/minute

3. **Service Stability**
   - Verify service remains responsive under load
   - No database connection exhaustion
   - Redis rate limiter functional

### Validation Points
- [ ] Rate limits enforced (5 login/min, 3 register/min)
- [ ] 429 status returned
- [ ] Service remains stable
- [ ] Redis rate limiting functional

---

## PHASE 8 — Database Integrity

### Test Cases
1. **Foreign Key Constraints**
   - Attempt to create refresh_token with non-existent user_id
   - Expected: Foreign key violation

2. **Token Expiration Cleanup**
   - Verify cleanup job removes expired tokens
   - Check refresh_tokens table for old records

3. **Audit Log Integrity**
   - Verify audit logs are immutable
   - Check all security events logged

### Validation Points
- [ ] Foreign keys enforced
- [ ] Expired tokens cleaned up
- [ ] Audit logs complete and immutable
- [ ] No orphaned records

---

## PHASE 9 — Observability

### Test Cases
1. **Stats Endpoint**
   - GET /auth/stats
   - Verify metrics populated (users, sessions, etc.)

2. **Security Telemetry**
   - GET /auth/security-telemetry
   - Verify security metrics (failed logins, MFA usage)

3. **Audit Logs**
   - GET /auth/audit-logs
   - Verify logs contain security events

### Validation Points
- [ ] Metrics endpoints return data
- [ ] No errors in response
- [ ] Data accuracy verified

---

## PHASE 10 — Frontend Integration

### Test Cases
1. **Session Expiration Modal**
   - Verify frontend shows modal when session expires
   - Redirect to login

2. **Unauthorized Redirect**
   - Access protected route without auth
   - Expected: Redirect to /login

3. **MFA Flow Rendering**
   - Verify MFA setup/verify pages render correctly
   - QR code generation works

4. **Telemetry Charts**
   - Verify security telemetry charts render
   - No console errors

### Validation Points
- [ ] UI handles auth states correctly
- [ ] No console errors
- [ ] Proper redirects
- [ ] MFA flow functional

---

## Vulnerability Classification

### Critical
- Authentication bypass
- Token forgery
- SQL injection in auth queries
- Password hash exposure

### High
- Brute force without protection
- Session fixation
- Insecure token storage
- Missing rate limiting

### Medium
- Information leakage in errors
- Weak password policy
- Missing security headers
- Insufficient logging

### Low
- Verbose error messages
- Missing HTTPS enforcement
- Outdated dependencies
- Documentation gaps

---

## Test Execution Plan

### Prerequisites
1. Service running on localhost:8084
2. Database populated with test data
3. Redis available for rate limiting
4. Test user credentials available

### Test Tools
- curl for API testing
- Custom Go test scripts for JWT manipulation
- Database queries for integrity checks
- Browser for frontend testing

### Execution Order
1. Database integrity (Phase 8) - Foundation
2. JWT security (Phase 1) - Core auth
3. Refresh tokens (Phase 2) - Session management
4. Brute force (Phase 3) - Attack prevention
5. MFA (Phase 4) - Additional security
6. API keys (Phase 5) - M2M auth
7. Sessions (Phase 6) - Multi-device
8. Rate limiting (Phase 7) - DoS prevention
9. Observability (Phase 9) - Monitoring
10. Frontend (Phase 10) - User experience

---

## Success Criteria

The FlyFleet IAM system will be declared **PRODUCTION READY** if:
1. All Critical tests pass
2. All High tests pass
3. At least 90% of Medium tests pass
4. No authentication bypass vulnerabilities
5. No data leakage vulnerabilities
6. Service remains stable under load
7. Audit logging functional

---

## Report Format

Final report will include:
1. Executive Summary
2. Test Results by Phase
3. Vulnerabilities Found (with classification)
4. Recommendations
5. Production Readiness Verdict