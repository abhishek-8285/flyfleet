# IAM / Authentication — Product Specification

**Document ID:** PSPEC-IAM-001  
**Version:** 1.0  
**Status:** Draft  
**Created:** April 7, 2026  
**Author:** Product Team  
**Priority:** P0 (Q2 Delivery)  
**Related:** [IAM & User Management](./IAM-USER-MANAGEMENT.md) — consolidation plan

---

## 1. Executive Summary

### 1.1 User Story

```
As a fleet operator admin
I want to log in securely and manage my team's access
So that only authorized people can see my fleet data
```

### 1.2 Acceptance Criteria

- [ ] Login with email + password returns JWT + refresh token
- [ ] Refresh token rotation works without session loss
- [ ] Role-based access control (super_admin, org_admin, dispatcher, driver)
- [ ] User CRUD — create, update, deactivate users
- [ ] Password reset via email
- [ ] MFA optional but supported
- [ ] Session management — view active sessions, revoke

### 1.3 Problem

FlyFleet has **two auth implementations** in the monolith:
- `services/auth-service/` — 74 endpoints, 95% complete, battle-tested
- `services/fleet-core/internal/domains/auth/` — duplicate auth v2, redundant

This creates: duplicate code paths, two deployables, inconsistent behavior, and confusion about which is the source of truth.

### 1.4 Solution

Consolidate into a single **IAM domain** (`domains/iam/`) within the fleet-core modular monolith. The auth-service code becomes the canonical IAM domain. The duplicate `domains/auth/` is deprecated. All 7 acceptance criteria are ~95% implemented — this spec documents what exists, identifies remaining gaps, and defines the completion plan.

### 1.5 Business Impact

| Metric | Target |
|--------|--------|
| Deployables | 2 → 1 binary |
| Auth support tickets | -40% |
| Enterprise deals unblocked | 3 deals (~$450K ARR) |
| SOC2 audit controls | Pass |

---

## 2. Current System State

### 2.1 Architecture

FlyFleet is a **modular monolith**. Single deployable: `services/fleet-core/`.

```
services/
├── fleet-core/                    ← THE MONOLITH (port 8084)
│   ├── internal/
│   │   ├── domains/               ← 27 business domains
│   │   │   ├── auth/              ← ⚠️ Duplicate v2 (DEPRECATE)
│   │   │   ├── tracking/          ← ✅ Complete (GPS/telemetry)
│   │   │   ├── geofencing/        ← ✅ Complete
│   │   │   ├── vehicle/           ← ✅ Complete
│   │   │   ├── driver/            ← ✅ Complete
│   │   │   ├── billing/           ← ✅ Complete
│   │   │   └── ...                ← 22 more domains
│   │   ├── middleware/            ← JWKS auth, RBAC, tenant context,
│   │   │                          ← rate limit, CORS, security headers
│   │   ├── platform/              ← Observability, EventBus, metrics
│   │   └── pkg/                   ← Shared utilities
│   └── cmd/main.go                ← Single entry point
│
├── auth-service/                  ← Standalone (port 8080) — TO BE CONSOLIDATED
│   ├── internal/
│   │   ├── handlers/              ← 7: auth, rbac, mfa, profile, session, user, invite
│   │   ├── service/               ← 8: auth, email, invitation, mfa, password, rbac, session, user
│   │   ├── repository/            ← 6: user, role, session, org, audit, password history
│   │   ├── models/                ← User, Role, Organization, Session, MFA, Invitation
│   │   ├── middleware/            ← JWT, CORS, rate limit, RBAC, security headers
│   │   ├── email/                 ← SendGrid, SMTP, NoOp
│   │   ├── redis/                 ← Session store, rate limiter, lockout store
│   │   └── nats/                  ← Event publishing
│   └── migrations/                ← DB schema migrations
│
└── _archive/                      ← Deprecated microservices (DO NOT MODIFY)
```

### 2.2 What's Already Built (Backend ~95%)

#### auth-service — 74 Endpoints, Fully Implemented

| Capability | Status | Source File |
|---|---|---|
| Login (email + password → JWT + refresh) | ✅ | `handlers/auth_handler.go` |
| Refresh token rotation | ✅ | `handlers/auth_handler.go`, `service/auth_service.go` |
| Forgot password + email delivery | ✅ | `handlers/auth_handler.go:384`, `service/auth_service.go:418` |
| Reset password + history check | ✅ | `service/auth_service.go:531` (bcrypt 12, history 5) |
| RBAC CRUD + Clone + Matrix + Bulk assign | ✅ | `handlers/rbac_handler.go` (12 endpoints) |
| Effective permissions | ✅ | `handlers/rbac_handler.go:334` |
| MFA (TOTP + SMS + backup codes) | ✅ | `handlers/mfa_handler.go` |
| Session management (list, revoke) | ✅ | `handlers/session_handler.go` |
| User CRUD + invitations | ✅ | `handlers/user_handler.go`, `handlers/invite_handler.go` |
| Email (SendGrid + SMTP + NoOp) | ✅ | `email/` (4 files) |
| Password history enforcement | ✅ | `repository/password_history_repo.go`, migration `003` |
| Account lockout (Redis) | ✅ | `redis/lockout_store.go` |
| Rate limiting (Redis) | ✅ | `redis/rate_limiter.go` |
| NATS event publishing | ✅ | `nats/publisher.go` |
| DB migrations | ✅ | `migrations/` |

#### fleet-core — Shared Infrastructure (Already In Place)

| Capability | Status | Source File |
|---|---|---|
| JWKS auth gate | ✅ | `middleware/jwks_auth.go` |
| BFF auth gate | ✅ | `middleware/bff_auth_gate.go` |
| RBAC middleware | ✅ | `middleware/rbac.go` |
| Require role for mutations | ✅ | `middleware/require_role_for_mutations.go` |
| Tenant context (org scoping) | ✅ | `middleware/tenant_context.go` |
| Rate limiting | ✅ | `middleware/ratelimit.go` |
| Security headers | ✅ | `middleware/security_headers.go` |
| CORS | ✅ | `middleware/cors.go` |
| Trace ID propagation | ✅ | `middleware/trace_id.go` |
| Security hardening | ✅ | Fixed: mass assignment, command injection, template injection, DoS, credential exposure |

#### Duplicate to Deprecate

| Component | Location | Action |
|---|---|---|
| fleet-core auth v2 domain | `internal/domains/auth/` | **Deprecate** — redundant with auth-service |

### 2.3 Frontend State

| Component | File | Status |
|---|---|---|
| Login page | `apps/dashboard-web/` | ✅ Exists |
| Password reset page | `apps/dashboard-web/` | ✅ Exists |
| Admin user list | `apps/dashboard-web/` | ✅ Exists |
| Admin RBAC page | `apps/dashboard-web/` | ✅ Exists |
| Org dropdown in header | `components/common/Layout.tsx` | ❌ Missing (G1) |
| Current org name display | `components/common/Layout.tsx` | ❌ Missing (G2) |
| Role clone button | `pages/admin/rbac/Roles.tsx` | ❌ Missing (G3) |
| Bulk assign UI | `pages/admin/rbac/UserAssignments.tsx` | ❌ Missing (G4) |
| `setOrganization` calls wrong endpoint | `context/AuthProvider.tsx:233` | ❌ Bug (G5) |

---

## 3. Gap Analysis

### 3.1 Backend: 0 Functional Gaps

All 7 acceptance criteria are **already implemented** in auth-service:

| # | Criterion | Status | Evidence |
|---|---|---|---|
| 1 | Login → JWT + refresh token | ✅ | `auth_handler.go` |
| 2 | Refresh token rotation | ✅ | `auth_service.go` — rotation logic |
| 3 | RBAC (4 roles) | ✅ | `rbac_handler.go` — 12 endpoints |
| 4 | User CRUD | ✅ | `user_handler.go` |
| 5 | Password reset via email | ✅ | `auth_service.go:418` + history |
| 6 | MFA optional | ✅ | `mfa_handler.go` — TOTP + SMS |
| 7 | Session management | ✅ | `session_handler.go` |

### 3.2 Architecture: 1 Gap

| Gap | Description | Effort |
|-----|-------------|--------|
| **Consolidation** | Merge auth-service → fleet-core as `domains/iam/`, deprecate `domains/auth/` | 3 days |

### 3.3 Frontend: 5 Gaps

| # | Gap | File | Priority | Effort |
|---|-----|------|----------|--------|
| G1 | No org dropdown in header | `Layout.tsx` | P0 | 2h |
| G2 | Current org name not shown | `Layout.tsx` | P0 | 1h |
| G3 | Role clone button missing | `Roles.tsx` | P1 | 1h |
| G4 | Bulk assign UI missing | `UserAssignments.tsx` | P2 | 3h |
| G5 | `setOrganization` calls `/auth/refresh` instead of `/auth/orgs/{id}/switch` | `AuthProvider.tsx:233` | P1 | 0.5h |

### 3.4 Integration: 2 Gaps

| Gap | Description | Priority |
|-----|-------------|----------|
| IAM ↔ fleet-core tenant isolation | Verify org_id filtering across all domains | P0 |
| Web-BFF auth passthrough | BFF (port 8081) forwards auth to consolidated IAM | P0 |

---

## 4. User Stories & Detailed Acceptance Criteria

### 4.1 US-001: Login & Token Management (P0)

```
As any user, I want to log in with email and password
So that I can access the fleet dashboard
```

**Status:** ✅ Implemented

| # | Criterion | Status |
|---|---|---|
| 1 | Email + password validation | ✅ |
| 2 | JWT access token returned on success | ✅ |
| 3 | Refresh token in httpOnly cookie | ✅ |
| 4 | Generic error on failure (no enumeration) | ✅ |
| 5 | Account lockout after failed attempts (Redis) | ✅ |
| 6 | **E2E: login → access fleet-core endpoint** | ❌ Needs test |

### 4.2 US-002: Token Refresh (P0)

```
As a logged-in user, I want my session to stay active
So that I don't lose my work
```

**Status:** ✅ Implemented

| # | Criterion | Status |
|---|---|---|
| 1 | Old token invalidated, new token issued | ✅ |
| 2 | Grace period for overlapping requests | ✅ |
| 3 | Session state preserved | ✅ |
| 4 | Expired token → 401 | ✅ |
| 5 | **E2E: 24h session without logout** | ❌ Needs test |

### 4.3 US-003: RBAC (P0)

```
As an org admin, I want to assign roles to users
So that they can only access what they need
```

**Status:** ✅ Implemented

| # | Criterion | Status |
|---|---|---|
| 1 | 4 default roles defined | ✅ |
| 2 | Custom role creation | ✅ |
| 3 | Role clone | ✅ |
| 4 | Permission matrix (resource × action) | ✅ |
| 5 | Bulk assign | ✅ |
| 6 | Effective permissions endpoint | ✅ |
| 7 | **E2E: dispatcher blocked from admin** | ❌ Needs test |

### 4.4 US-004: User CRUD (P0)

```
As an org admin, I want to create, update, deactivate users
So that I can manage my team's access
```

**Status:** ✅ Implemented

| # | Criterion | Status |
|---|---|---|
| 1 | Create user via email invite | ✅ |
| 2 | Update user details | ✅ |
| 3 | Deactivate (soft delete, revokes sessions) | ✅ |
| 4 | List with pagination + filtering | ✅ |
| 5 | Resend invitation | ✅ |
| 6 | **E2E: create → invite → accept → login** | ❌ Needs test |

### 4.5 US-005: Password Reset (P0)

```
As a user who forgot their password, I want to reset via email
So that I can regain access
```

**Status:** ✅ Implemented

| # | Criterion | Status |
|---|---|---|
| 1 | Forgot password → email (always 200) | ✅ |
| 2 | Reset token: 1h expiry, single-use | ✅ |
| 3 | Password history enforcement (last 5) | ✅ |
| 4 | Bcrypt cost 12 | ✅ |
| 5 | All sessions invalidated on reset | ✅ |
| 6 | **E2E: forgot → email → reset → login** | ❌ Needs test |

### 4.6 US-006: MFA (P1)

```
As a security-conscious admin, I want to enable MFA
So that my account is more secure
```

**Status:** ✅ Implemented

| # | Criterion | Status |
|---|---|---|
| 1 | TOTP setup (QR + manual) | ✅ |
| 2 | TOTP verification during login | ✅ |
| 3 | SMS fallback | ✅ |
| 4 | Backup codes (10, single-use) | ✅ |
| 5 | MFA disable flow | ✅ |
| 6 | **E2E: enable TOTP → login with code** | ❌ Needs test |

### 4.7 US-007: Session Management (P1)

```
As a logged-in user, I want to see and revoke my sessions
So that I can secure my account
```

**Status:** ✅ Implemented

| # | Criterion | Status |
|---|---|---|
| 1 | List active sessions | ✅ |
| 2 | Revoke individual | ✅ |
| 3 | Revoke all others | ✅ |
| 4 | Concurrent session limit | ✅ |
| 5 | **E2E: 2 devices → revoke one → other works** | ❌ Needs test |

---

## 5. Architecture: IAM Domain in the Modular Monolith

### 5.1 Target Structure

auth-service code moves into fleet-core as the `iam` domain, following the established `internal/domains/` pattern used by all 27 domains:

```
services/fleet-core/internal/
├── domains/
│   ├── iam/                          # ← NEW (from auth-service)
│   │   ├── handlers/                 # 7 handler files
│   │   │   ├── auth_handler.go
│   │   │   ├── rbac_handler.go
│   │   │   ├── mfa_handler.go
│   │   │   ├── profile_handler.go
│   │   │   ├── session_handler.go
│   │   │   ├── user_handler.go
│   │   │   └── invite_handler.go
│   │   ├── service/                  # 8 service files
│   │   │   ├── auth_service.go
│   │   │   ├── rbac_service.go
│   │   │   ├── mfa_service.go
│   │   │   ├── password_service.go
│   │   │   ├── session_service.go
│   │   │   ├── user_service.go
│   │   │   ├── email_service.go
│   │   │   └── invitation_service.go
│   │   ├── repository/               # 6 repository files
│   │   │   ├── user_repo.go
│   │   │   ├── role_repo.go
│   │   │   ├── session_repo.go
│   │   │   ├── org_repo.go
│   │   │   ├── audit_repo.go
│   │   │   └── password_history_repo.go
│   │   ├── models/                   # 6 model files
│   │   │   ├── user.go
│   │   │   ├── role.go
│   │   │   ├── organization.go
│   │   │   ├── session.go
│   │   │   ├── mfa.go
│   │   │   └── invitation.go
│   │   └── email/                    # 3 provider files
│   │       ├── sendgrid.go
│   │       ├── smtp.go
│   │       └── noop.go
│   │
│   ├── auth/                         # ⚠️ DEPRECATE — duplicate v2
│   └── ...                           # 26 other domains
│
├── middleware/                       # Already exists — no changes
│   ├── jwks_auth.go
│   ├── rbac.go
│   ├── tenant_context.go
│   ├── ratelimit.go
│   └── security_headers.go
│
├── platform/                         # Shared: observability, EventBus
│
└── cmd/main.go                       # Register IAM routes
```

### 5.2 Route Registration

IAM routes register on the existing fleet-core router in `cmd/main.go`:

```go
// cmd/main.go — existing pattern
r.Route("/api/v1", func(r chi.Router) {
    // IAM domain (from auth-service)
    iam := iamModule.New(dbPool, redisPool, eventBus, cfg)
    iam.RegisterRoutes(r)    // 74 endpoints

    // Other domains (existing)
    tracking.RegisterRoutes(r)
    vehicle.RegisterRoutes(r)
    // ...
})
```

### 5.3 Module Interfaces (Cross-Domain Calls)

Other domains call IAM via Go interfaces, **not HTTP**:

```go
// What other domains can ask IAM
type UserProvider interface {
    GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
    GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]OrgInfo, error)
}

type PermissionChecker interface {
    HasPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)
    GetEffectivePermissions(ctx context.Context, userID uuid.UUID) (map[string][]string, error)
}

type OrgResolver interface {
    ResolveOrg(ctx context.Context, userID, orgID uuid.UUID) (*models.Organization, error)
}
```

### 5.4 Shared Infrastructure (No Duplication)

| Concern | auth-service (current) | fleet-core (target) | Action |
|---------|----------------------|---------------------|--------|
| JWT validation | Own middleware | `middleware/jwks_auth.go` | **Use fleet-core's** |
| RBAC | Own middleware | `middleware/rbac.go` | **Use fleet-core's** |
| Rate limiting | Redis-based | `middleware/ratelimit.go` | **Use fleet-core's** |
| CORS | Own middleware | `middleware/cors.go` | **Use fleet-core's** |
| Security headers | Own middleware | `middleware/security_headers.go` | **Use fleet-core's** |
| Redis | Own client | Shared pool | **Share pool** |
| NATS | Own publisher | EventBus | **Use EventBus** |
| DB | Own pool | Shared pgxpool | **Share pool** |
| Logging | zap | zap | **Same** |

### 5.5 Middleware Chain

The existing fleet-core middleware chain already handles everything IAM needs:

```
Request → CORS → Rate Limit → Security Headers → JWKS Auth → RBAC → Tenant Context → Handler
```

No new middleware required.

---

## 6. API Contracts

### 6.1 Complete Endpoint Inventory

**74 endpoints** across 4 groups. All functional in auth-service.

#### Public (No Auth) — 11 endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/identify` | Identify user by email |
| POST | `/api/v1/auth/login` | Login → JWT + refresh token |
| POST | `/api/v1/auth/refresh` | Rotate refresh token |
| POST | `/api/v1/auth/register` | Self-register |
| POST | `/api/v1/auth/forgot-password` | Send reset email |
| POST | `/api/v1/auth/reset-password` | Complete password reset |
| POST | `/api/v1/auth/verify-email` | Verify email token |
| POST | `/api/v1/auth/mfa/verify` | Verify MFA code |
| GET | `/api/v1/auth/invite/{token}` | View invite details |
| POST | `/api/v1/auth/invite/accept` | Accept invite |
| POST | `/api/v1/auth/resolve-orgs` | Frontend alias → identify |

#### Authenticated (JWT Required) — 14 endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/logout` | Logout + invalidate session |
| POST | `/api/v1/auth/change-password` | Change password |
| GET | `/api/v1/auth/password-policy` | Get password policy |
| POST | `/api/v1/auth/orgs/{orgId}/switch` | Switch org → new JWT |
| GET | `/api/v1/auth/orgs` | List user's orgs |
| GET | `/api/v1/auth/profile` | Get current profile |
| PUT | `/api/v1/auth/profile` | Update profile |
| GET | `/api/v1/auth/sessions` | List active sessions |
| DELETE | `/api/v1/auth/sessions/{id}` | Revoke session |
| DELETE | `/api/v1/auth/sessions/all` | Revoke all sessions |
| POST | `/api/v1/auth/mfa/setup/totp` | Initiate TOTP setup |
| POST | `/api/v1/auth/mfa/setup/totp/verify` | Verify TOTP → enable |
| POST | `/api/v1/auth/mfa/disable` | Disable MFA |
| GET | `/api/v1/auth/mfa/backup-codes` | Get backup codes |

#### Admin — User Management — 8 endpoints

| Method | Path | Permission |
|--------|------|-----------|
| GET | `/api/v1/admin/users` | user:read |
| GET | `/api/v1/admin/users/{id}` | user:read |
| GET | `/api/v1/admin/users/{id}/audit-log` | user:read |
| PUT | `/api/v1/admin/users/{id}/status` | user:write |
| DELETE | `/api/v1/admin/users/{id}` | user:delete |
| POST | `/api/v1/admin/users/invite` | user:write |
| POST | `/api/v1/admin/users/invite/{id}/resend` | user:write |
| DELETE | `/api/v1/admin/users/invite/{id}` | user:write |

#### Admin — RBAC — 13 endpoints

| Method | Path | Permission |
|--------|------|-----------|
| GET | `/api/v1/admin/rbac/roles` | role:read |
| GET | `/api/v1/admin/rbac/matrix` | role:read |
| GET | `/api/v1/admin/rbac/permissions` | role:read |
| GET | `/api/v1/admin/rbac/users/{id}/roles` | role:read |
| GET | `/api/v1/admin/rbac/users/{id}/permissions` | role:read |
| POST | `/api/v1/admin/rbac/roles` | role:write |
| PUT | `/api/v1/admin/rbac/roles/{id}` | role:write |
| POST | `/api/v1/admin/rbac/roles/{id}/clone` | role:write |
| PUT | `/api/v1/admin/rbac/roles/{id}/permissions` | role:write |
| POST | `/api/v1/admin/rbac/users/assign` | role:write |
| POST | `/api/v1/admin/rbac/users/assign/bulk` | role:write |
| DELETE | `/api/v1/admin/rbac/users/unassign` | role:write |
| DELETE | `/api/v1/admin/rbac/roles/{id}` | role:delete |

### 6.2 Key Request/Response Examples

#### POST /api/v1/auth/login

```json
// Request
{
  "email": "admin@fleetops.io",
  "password": "correct-horse-battery-staple",
  "mfa_code": "123456"
}

// Response (200)
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "usr_abc123",
    "email": "admin@fleetops.io",
    "name": "Rajesh Kumar",
    "roles": ["org_admin"],
    "organizations": [
      { "id": "org_xyz", "name": "FleetOps India", "role": "org_admin" }
    ]
  }
}

// Set-Cookie: refresh_token=eyJ...; Path=/; HttpOnly; Secure; SameSite=Strict; Max-Age=604800

// Error (401)
{ "error": "invalid_credentials", "message": "Invalid email or password" }
```

#### POST /api/v1/auth/refresh

```json
// Response (200) — new access + new refresh token
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 900
}
// Set-Cookie: new refresh_token (old one invalidated)
```

#### POST /api/v1/auth/orgs/{orgId}/switch

```json
// Response (200) — new JWT with new org context
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 900,
  "organization": {
    "id": "org_xyz",
    "name": "FleetOps India",
    "role": "org_admin"
  }
}
```

---

## 7. Data Models

### 7.1 Tables Owned by IAM Domain

All tables already exist via auth-service migrations. **No new tables needed.**

| Table | Purpose | Key Fields |
|-------|---------|-----------|
| `users` | User accounts | id, email, password_hash, name, phone, status, mfa_enabled, last_login_at |
| `organizations` | Tenant entities | id, name, slug, settings, mfa_required, session_limit |
| `user_organizations` | User↔Org mapping | user_id, org_id, role_id, is_primary |
| `roles` | Role definitions | id, org_id, name, is_system, permissions (JSONB) |
| `sessions` | Active sessions | id, user_id, org_id, device_info, ip_address, is_active, expires_at |
| `auth_tokens` | Refresh tokens | id, user_id, token_hash, is_active, expires_at, replaced_by |
| `mfa_codes` | MFA secrets | user_id, method, totp_secret, backup_codes_hash, is_enabled |
| `invitations` | User invites | id, org_id, email, role_id, token, status, expires_at |
| `password_history` | Password history | user_id, password_hash, created_at |
| `audit_logs` | Security audit | org_id, user_id, action, resource_type, details (JSONB) |

### 7.2 Key Constraints

```sql
users.email UNIQUE
users.status IN ('active', 'invited', 'deactivated', 'locked')
UNIQUE(org_id, name) ON roles
roles.permissions = JSONB array: ['vehicle:read', 'vehicle:write', ...]
sessions.expires_at < now() → auto-cleanup (hourly job)
ALL queries filter by org_id via user_organizations JOIN
```

### 7.3 Entity Relationships

```
organizations ──┐
                ├── user_organizations ── users ── auth_tokens
roles ──────────┘          │               ├── sessions
                           │               ├── mfa_codes
                           │               └── password_history
                           │
                           └── invitations
```

---

## 8. Security Requirements

### 8.1 Already Implemented

| Control | Status | Location |
|---|---|---|
| Password hashing (bcrypt 12) | ✅ | `service/auth_service.go` |
| Password history (last 5) | ✅ | `service/password_service.go` |
| Account lockout (Redis) | ✅ | `redis/lockout_store.go` |
| Rate limiting | ✅ | `redis/rate_limiter.go` + `middleware/ratelimit.go` |
| JWT signing | ✅ | `service/auth_service.go` |
| httpOnly cookies | ✅ | `handlers/auth_handler.go` |
| Input validation (DTO pattern) | ✅ | `models/requests.go` |
| Parameterized queries | ✅ | All repositories |
| Mass assignment prevention | ✅ | Fixed |
| OS command injection prevention | ✅ | Fixed |
| Template injection prevention | ✅ | Fixed |
| Credential exposure prevention | ✅ | Fixed (`json:"-"` on sensitive fields) |
| Error info leakage prevention | ✅ | Fixed (generic errors) |
| Multi-tenant isolation | ✅ | org_id in all queries |

### 8.2 Security Posture

The auth-service has **zero known vulnerabilities**. Fleet-core has had 5 security fixes applied. The consolidated IAM domain inherits all fleet-core security hardening.

---

## 9. Frontend Specifications

### 9.1 Gap Inventory & Fixes

| # | Gap | File | Fix |
|---|-----|------|-----|
| G1 | No org dropdown in header | `Layout.tsx` | Add dropdown → `GET /api/v1/auth/orgs` |
| G2 | Current org name not shown | `Layout.tsx` | Display from Redux store |
| G3 | Role clone button missing | `Roles.tsx` | Button → `POST /admin/rbac/roles/{id}/clone` |
| G4 | Bulk assign UI missing | `UserAssignments.tsx` | Checkboxes + action bar → bulk assign API |
| G5 | Wrong org switch endpoint | `AuthProvider.tsx:233` | `/auth/refresh` → `/auth/orgs/${orgId}/switch` |

### 9.2 Fix G5: setOrganization (Priority: P1, 0.5h)

```typescript
// context/AuthProvider.tsx — line 233
// BEFORE:
const response = await api.post('/auth/refresh', { org_id: orgId });
// AFTER:
const response = await api.post(`/auth/orgs/${orgId}/switch`);
const { access_token, organization } = response.data;
dispatch(setTokens({ access_token, organization }));
```

### 9.3 Fix G1+G2: Header Org Dropdown (Priority: P0, 3h)

```tsx
// components/common/Layout.tsx — addition
const OrganizationSwitch = () => {
  const { organizations, currentOrg, setOrganization } = useAuth();

  if (organizations.length <= 1) {
    return <span>{currentOrg?.name}</span>;  // Single org, no dropdown
  }

  return (
    <select value={currentOrg?.id} onChange={(e) => setOrganization(e.target.value)}>
      {organizations.map((org) => (
        <option key={org.id} value={org.id}>
          {org.name} — {org.role}
        </option>
      ))}
    </select>
  );
};
```

### 9.4 Fix G3: Role Clone Button (Priority: P1, 1h)

```tsx
// pages/admin/rbac/Roles.tsx
const handleClone = async (roleId: string) => {
  await api.post(`/admin/rbac/roles/${roleId}/clone`, {
    name: `${role.name} (Copy)`,
  });
  toast.success('Role cloned');
  refetchRoles();
};
```

### 9.5 Fix G4: Bulk Assign UI (Priority: P2, 3h)

```tsx
// pages/admin/rbac/UserAssignments.tsx
// Checkboxes on user rows → action bar → POST /admin/rbac/users/assign/bulk
```

---

## 10. Migration & Rollout Plan

### 10.1 Phase 1: IAM Domain Consolidation (3 days)

```
Day 1: Move auth-service code into fleet-core
├── Copy handlers/ → fleet-core/internal/domains/iam/handlers/
├── Copy service/ → fleet-core/internal/domains/iam/service/
├── Copy repository/ → fleet-core/internal/domains/iam/repository/
├── Copy models/ → fleet-core/internal/domains/iam/models/
├── Copy email/ → fleet-core/internal/domains/iam/email/
└── Update import paths (auth-service → fleet-core)

Day 2: Wire into fleet-core monolith
├── Register all 74 IAM routes in cmd/main.go
├── Replace auth-service Redis/NATS with fleet-core shared pools
├── Remove duplicate middleware (use fleet-core's existing)
├── Deprecate domains/auth/ (rename to auth_deprecated/)
├── Verify all endpoints respond correctly
└── Integration test: login → org switch → fleet data

Day 3: Clean up
├── Merge migrations (auth-service → fleet-core/migrations/)
├── Single Dockerfile (remove auth-service Dockerfile)
├── Update docker-compose (remove auth-service container)
├── Update .env (remove AUTH_SERVICE_PORT)
└── Smoke test: full auth → fleet workflow
```

### 10.2 Phase 2: Frontend Gaps (1.5 days)

```
Day 4: G1+G2 (Header org dropdown + display)
└── Layout.tsx + AuthProvider.tsx

Day 5: G3+G4+G5 (RBAC UI + org switch fix)
├── Roles.tsx, UserAssignments.tsx, AuthProvider.tsx:233
└── E2E tests for all 5 gaps
```

### 10.3 Phase 3: Testing & Polish (2 days)

```
Day 6: Integration tests
├── Login → access fleet-core endpoints
├── Token rotation (100 refresh cycles)
├── RBAC enforcement (dispatcher blocked from admin)
├── Org isolation (org A cannot see org B data)
└── Password reset end-to-end

Day 7: Security audit + deployment
├── govulncheck on consolidated codebase
├── Rate limit testing
├── Single Docker build
└── Production deploy
```

---

## 11. Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Login success rate | >99.5% | Auth logs |
| Token refresh without session loss | 100% | Integration test |
| RBAC permission bypasses | 0 | Security audit |
| Frontend gaps closed | 5/5 | E2E tests |
| Deployables | 2 → 1 | docker-compose ps |
| Auth support tickets | -40% | Monthly report |

---

## 12. Risks & Mitigations

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|-----------|
| Import path conflicts during merge | Medium | Medium | `go mod tidy` + systematic rename |
| Duplicate auth causes confusion | High | Medium | Rename `domains/auth/` → `auth_deprecated/` immediately |
| Frontend org switch breaks fleet data | Low | Critical | RLS policies validated with tests |
| Migration downtime | Low | High | Dry run in staging, maintenance window |
| Redis/NATS connection conflicts | Low | Medium | Use fleet-core shared pools, test early |

---

## 13. Timeline & Capacity

| Phase | Effort | Team | Duration |
|-------|--------|------|----------|
| IAM domain consolidation | 3 days | 1 Go engineer | Days 1-3 |
| Frontend gaps (G1-G5) | 1.5 days | 1 React engineer | Days 4-5 |
| Integration testing + audit | 2 days | Both engineers | Days 6-7 |
| **Total** | **~6.5 days** | **2 engineers** | **1 sprint** |

---

## 14. Appendix

### A. File Mapping (auth-service → fleet-core/domains/iam)

| auth-service Source | fleet-core Target |
|---|---|
| `internal/handlers/auth_handler.go` | `internal/domains/iam/handlers/auth_handler.go` |
| `internal/handlers/rbac_handler.go` | `internal/domains/iam/handlers/rbac_handler.go` |
| `internal/handlers/mfa_handler.go` | `internal/domains/iam/handlers/mfa_handler.go` |
| `internal/handlers/profile_handler.go` | `internal/domains/iam/handlers/profile_handler.go` |
| `internal/handlers/session_handler.go` | `internal/domains/iam/handlers/session_handler.go` |
| `internal/handlers/user_handler.go` | `internal/domains/iam/handlers/user_handler.go` |
| `internal/handlers/invite_handler.go` | `internal/domains/iam/handlers/invite_handler.go` |
| `internal/service/auth_service.go` | `internal/domains/iam/service/auth_service.go` |
| `internal/service/rbac_service.go` | `internal/domains/iam/service/rbac_service.go` |
| `internal/service/mfa_service.go` | `internal/domains/iam/service/mfa_service.go` |
| `internal/service/password_service.go` | `internal/domains/iam/service/password_service.go` |
| `internal/service/session_service.go` | `internal/domains/iam/service/session_service.go` |
| `internal/service/user_service.go` | `internal/domains/iam/service/user_service.go` |
| `internal/service/email_service.go` | `internal/domains/iam/service/email_service.go` |
| `internal/service/invitation_service.go` | `internal/domains/iam/service/invitation_service.go` |
| `internal/repository/*.go` | `internal/domains/iam/repository/` |
| `internal/models/*.go` | `internal/domains/iam/models/` |
| `internal/email/*.go` | `internal/domains/iam/email/` |
| `migrations/*.sql` | `migrations/` (merged) |

### B. Permission Matrix

| Resource | super_admin | org_admin | dispatcher | driver |
|----------|:-----------:|:---------:|:----------:|:------:|
| vehicle:read/write | ✅ | ✅ | ✅ | ❌ |
| driver:read/write | ✅ | ✅ | ✅ | ❌ |
| trip:read/write | ✅ | ✅ | ✅ | read(own) |
| user:read/write | ✅ | ✅ | ❌ | ❌ |
| role:read/write | ✅ | ✅ | ❌ | ❌ |
| billing:read/write | ✅ | ✅ | ❌ | ❌ |
| settings:read/write | ✅ | ✅ | ❌ | ❌ |

### C. Related Documents

- [IAM & User Management](./IAM-USER-MANAGEMENT.md) — Modular monolith consolidation plan (v3.0)
- [IAM Test Report](./IAM-TEST-REPORT.md) — Existing test results
- CLAUDE.md — Architecture guidelines
- QWEN.md — Project overview (modular monolith)

---

*Version 1.0 · 2026-04-07 · Next Review: 2026-04-14*
