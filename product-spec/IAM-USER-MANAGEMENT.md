# Product Specification: IAM & User Management

> **FlyFleet — Enterprise Fleet Management Platform**
> Scope: Login · Roles · Organization Switching · Password Reset
> Architecture: **Modular Monolith**
> Version 3.0 (Modular Monolith) · 2026-04-04 · Draft

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Current System State](#2-current-system-state)
3. [Target Architecture: Modular Monolith](#3-target-architecture-modular-monolith)
4. [Gap Analysis](#4-gap-analysis)
5. [IAM Module Design](#5-iam-module-design)
6. [User Stories](#6-user-stories)
7. [Technical Specifications](#7-technical-specifications)
8. [Prioritization & Timeline](#8-prioritization--timeline)
9. [Risks](#9-risks)
10. [Team & Budget](#10-team--budget)
11. [Appendix](#11-appendix)

---

## 1. Executive Summary

**Problem**: The current system has auth logic split across two services (auth-service on 8080, fleet-core on 8084) with duplicate auth implementations. The frontend has 4 UI gaps. Operations are complex with two deployables.

**Solution**: Consolidate into a **modular monolith** — single binary with clear internal module boundaries. The IAM module absorbs auth-service code directly. No inter-service HTTP calls. Direct function calls between modules.

**What's Already Done (Backend)**: 95% of IAM backend code exists in `services/auth-service/` — 7 handlers, 8 services, email delivery, RBAC with clone/matrix/bulk, MFA, sessions, invitations, password history enforcement.

**What Needs Building**: 5 frontend UI gaps + modular monolith consolidation.

**Business Impact**:
- Single deploy, single binary, simpler ops
- Direct function calls between modules (no HTTP overhead)
- Reduce auth support tickets by 40%
- Unblock enterprise deals blocked by IAM completeness

---

## 2. Current System State

### 2.1 What Exists Today

| Component | Location | Status |
|-----------|----------|--------|
| **auth-service** | `services/auth-service/` | ✅ Active, port 8080, 7 handlers, 8 services |
| **fleet-core** | `services/fleet-core/` | ✅ Active, port 8084, 29 domains |
| **Frontend** | `apps/dashboard-web/` | ✅ Login, reset, admin pages exist |
| **DB migrations** | `infrastructure/migrations/` | ✅ 100+ migrations |
| **Shared packages** | `packages/common/go/pkg/security/` | ✅ JWT, RBAC, tenant |

### 2.2 Backend: 95% Complete

**auth-service** has everything:
- 7 handlers: auth, RBAC, MFA, profile, session, user, invite
- 8 services: auth, email, invitation, MFA, password, RBAC, session, user
- Email delivery: SendGrid + SMTP + NoOp providers
- 38 API endpoints, all wired and functional
- RBAC includes: CloneRole, GetPermissionMatrix, BulkAssignRole, GetUserPermissions
- Password reset: ForgotPassword → email → ResetPasswordWithToken → history check → bcrypt(12) → session revoke
- Org switching: `POST /api/v1/auth/orgs/{orgId}/switch` returns new JWT tokens

### 2.3 Frontend: 4 Gaps Confirmed

| # | Gap | File | Priority |
|---|-----|------|----------|
| G1 | No org dropdown in header | `components/common/Layout.tsx` | P0 |
| G2 | Current org name not shown | `components/common/Layout.tsx` | P0 |
| G3 | Role clone button missing | `pages/admin/rbac/Roles.tsx` | P1 |
| G4 | Bulk assign UI missing | `pages/admin/rbac/UserAssignments.tsx` | P2 |
| G5 | `setOrganization` calls `/auth/refresh` not `/auth/orgs/{id}/switch` | `context/AuthProvider.tsx:233` | P1 |

---

## 3. Target Architecture: Modular Monolith

### 3.1 Why Modular Monolith

| Factor | Microservices (Current) | Modular Monolith (Target) |
|--------|------------------------|--------------------------|
| Deployables | 2+ (auth-service + fleet-core) | 1 binary |
| Inter-module communication | HTTP over network | Direct function calls |
| Database | Shared PostgreSQL | Shared PostgreSQL, clear table ownership |
| Local dev | Run 2+ services | `go run cmd/main.go` |
| Scaling | Scale independently | Scale the whole thing |
| Team size | Needs 5+ engineers | Works with 2-3 engineers |
| Complexity | High (service discovery, retries, circuit breakers) | Low (in-process calls) |
| Future extraction | Possible when module boundaries are clear | Same — clean boundaries enable extraction |

### 3.2 Target Structure

```
services/fleet-core/
├── cmd/
│   └── main.go                    # Single entry point, router setup
├── internal/
│   ├── config/                    # Shared config (viper)
│   ├── database/                  # Shared DB connection pool
│   ├── middleware/                # Shared middleware (auth, RBAC, CORS, rate limit)
│   │
│   ├── modules/
│   │   ├── iam/                   # ← IAM MODULE (this spec)
│   │   │   ├── handler/           # HTTP handlers (from auth-service)
│   │   │   │   ├── auth_handler.go
│   │   │   │   ├── rbac_handler.go
│   │   │   │   ├── mfa_handler.go
│   │   │   │   ├── profile_handler.go
│   │   │   │   ├── session_handler.go
│   │   │   │   ├── user_handler.go
│   │   │   │   └── invite_handler.go
│   │   │   ├── service/           # Business logic (from auth-service)
│   │   │   │   ├── auth_service.go
│   │   │   │   ├── email_service.go
│   │   │   │   ├── rbac_service.go
│   │   │   │   ├── password_service.go
│   │   │   │   ├── mfa_service.go
│   │   │   │   ├── session_service.go
│   │   │   │   ├── user_service.go
│   │   │   │   └── invitation_service.go
│   │   │   ├── repository/        # Data access
│   │   │   │   ├── user_repo.go
│   │   │   │   ├── role_repo.go
│   │   │   │   ├── org_repo.go
│   │   │   │   ├── session_repo.go
│   │   │   │   └── password_history_repo.go
│   │   │   ├── models/            # Domain models
│   │   │   │   ├── user.go
│   │   │   │   ├── role.go
│   │   │   │   ├── organization.go
│   │   │   │   └── session.go
│   │   │   └── email/             # Email providers
│   │   │       ├── sendgrid.go
│   │   │       ├── smtp.go
│   │   │       └── noop.go
│   │   │
│   │   ├── fleet/                 # Fleet module (from fleet-core domains)
│   │   ├── vehicle/               # Vehicle module
│   │   ├── driver/                # Driver module
│   │   ├── trip/                  # Trip module
│   │   ├── maintenance/           # Maintenance module
│   │   ├── billing/               # Billing module
│   │   ├── analytics/             # Analytics module
│   │   └── governance/            # Governance module
│   │
│   └── shared/                    # Cross-cutting concerns
│       ├── security/              # JWT, RBAC policies, tenant isolation
│       └── observability/         # Logging, metrics, tracing
│
├── migrations/                    # All DB migrations (combined)
├── config.toml                    # Single config file
├── go.mod
└── Dockerfile                     # Single container image
```

### 3.3 Module Communication

```
┌─────────────────────────────────────────────────────────┐
│                    MONOLITH (port 8080)                  │
│                                                          │
│  HTTP Request → Router → Middleware Chain                │
│                     ↓                                    │
│              Auth Middleware (JWT validate)              │
│                     ↓                                    │
│              RBAC Middleware (permission check)          │
│                     ↓                                    │
│              Org Scope Middleware (tenant isolation)     │
│                     ↓                                    │
│  ┌─────────────────────────────────────────────────┐    │
│  │  IAM Module                                     │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────────┐  │    │
│  │  │ Handler  │→ │ Service  │→ │ Repository   │  │    │
│  │  │          │  │          │  │              │  │    │
│  │  │ auth     │  │ auth     │  │ user_repo    │  │    │
│  │  │ rbac     │  │ rbac     │  │ role_repo    │  │    │
│  │  │ mfa      │  │ email    │  │ org_repo     │  │    │
│  │  │ session  │  │ password │  │ session_repo │  │    │
│  │  └──────────┘  └──────────┘  └──────────────┘  │    │
│  └─────────────────────────────────────────────────┘    │
│                     ↓ (direct function call)             │
│  ┌─────────────────────────────────────────────────┐    │
│  │  Fleet Module                                   │    │
│  │  handler → service → repository                 │    │
│  └─────────────────────────────────────────────────┘    │
│                     ↓ (direct function call)             │
│  ┌─────────────────────────────────────────────────┐    │
│  │  Vehicle Module                                 │    │
│  │  handler → service → repository                 │    │
│  └─────────────────────────────────────────────────┘    │
│                                                          │
│  Shared: PostgreSQL pool · Redis pool · NATS client     │
└─────────────────────────────────────────────────────────┘
```

**Key principle**: Modules call each other via Go interfaces, not HTTP. The IAM module exposes interfaces like `UserProvider`, `OrgResolver`, `PermissionChecker` that other modules consume directly.

### 3.4 Migration Path

```
Phase 1 (Week 1): Scaffold monolith
├── Copy auth-service code → monolith/internal/modules/iam/
├── Copy fleet-core domains → monolith/internal/modules/fleet/ etc.
├── Single router in cmd/main.go
├── Shared middleware chain
└── Single config.toml

Phase 2 (Week 2): Wire IAM module
├── Register all IAM routes on monolith router
├── Verify all 38 endpoints work
├── Integration test: login → org switch → fleet data
└── Deprecate standalone auth-service

Phase 3 (Week 3): Frontend gaps
├── G1+G2: Header org dropdown + current org display
├── G3: Role clone button
├── G4: Bulk assign UI
├── G5: Fix setOrganization endpoint
└── E2E tests

Phase 4 (Week 4): Polish & deploy
├── Single Dockerfile
├── CI/CD pipeline update
├── Load testing
└── Production deploy
```

---

## 4. Gap Analysis

### 4.1 Backend: 0 Gaps

All backend IAM functionality exists and works:

| Capability | Status | Location |
|---|---|---|
| Login, Register, Identify | ✅ | `auth_handler.go` |
| ForgotPassword + email | ✅ | `auth_handler.go:384`, `auth_service.go:418` |
| ResetPassword + history | ✅ | `auth_service.go:531` |
| Org switch API | ✅ | `auth_handler.go:215` |
| RBAC CRUD + Clone + Matrix + Bulk | ✅ | `rbac_handler.go` (12 endpoints) |
| Effective permissions | ✅ | `rbac_handler.go:334` |
| MFA (TOTP + SMS) | ✅ | `mfa_handler.go` |
| Sessions | ✅ | `session_handler.go` |
| Invitations | ✅ | `invite_handler.go` |
| Email (SendGrid + SMTP) | ✅ | `email/` package |
| Password history table | ✅ | Migration `003_password_history.sql` |
| Org switch events table | ✅ | Migration `004_org_theme.sql` |

### 4.2 Frontend: 5 Gaps

| # | Gap | File | Effort |
|---|-----|------|--------|
| G1 | No org dropdown in header | `Layout.tsx` | 2h |
| G2 | Current org name not shown | `Layout.tsx` | 1h |
| G3 | Role clone button | `Roles.tsx` | 1h |
| G4 | Bulk assign UI | `UserAssignments.tsx` | 3h |
| G5 | `setOrganization` wrong endpoint | `AuthProvider.tsx` | 0.5h |

### 4.3 Architecture: 1 Gap

| Gap | Details | Effort |
|-----|---------|--------|
| Modular monolith consolidation | Merge auth-service + fleet-core into single binary | 3 days |

---

## 5. IAM Module Design

### 5.1 Module Boundary

The IAM module owns these database tables:
- `users`, `organizations`, `user_organizations`
- `roles`, `permissions`, `user_roles`, `role_permissions`
- `auth_tokens`, `mfa_codes`, `mfa_policies`, `mfa_events`
- `invitations`, `password_history`, `org_switch_events`
- `security_events`, `audit_logs`

The IAM module exposes these interfaces to other modules:

```go
// What other modules can ask IAM
type UserProvider interface {
    GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
    GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]models.OrgInfo, error)
}

type PermissionChecker interface {
    HasPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)
    GetEffectivePermissions(ctx context.Context, userID uuid.UUID) (map[string][]string, error)
}

type OrgResolver interface {
    ResolveOrg(ctx context.Context, userID, orgID uuid.UUID) (*models.Organization, error)
}
```

### 5.2 Module Internal Structure

```
modules/iam/
├── handler/          # HTTP layer — thin, delegates to service
├── service/          # Business logic — validates, orchestrates
├── repository/       # Data access — SQL queries
├── models/           # Domain types — User, Role, Organization, Session
└── email/            # Email providers — SendGrid, SMTP, NoOp
```

**Handler → Service → Repository** pattern. No handler talks to DB directly. No service knows about HTTP.

### 5.3 Routes (All 38 Endpoints)

All routes registered under `/api/v1/` on the monolith router:

**Public (no auth)**:
```
POST   /auth/identify
POST   /auth/login
POST   /auth/forgot-password
POST   /auth/register
POST   /auth/mfa/verify
POST   /auth/refresh
POST   /auth/reset-password
POST   /auth/verify-email
GET    /auth/invite/{token}
POST   /auth/invite/accept
POST   /auth/resolve-orgs          # Frontend alias → /auth/identify
```

**Authenticated (JWT required)**:
```
POST   /auth/logout
POST   /auth/change-password
GET    /auth/password-policy
POST   /auth/orgs/{orgId}/switch
GET    /auth/orgs
GET    /auth/profile
PUT    /auth/profile
GET    /auth/sessions
DELETE /auth/sessions/{id}
DELETE /auth/sessions/all
POST   /auth/mfa/setup/totp
POST   /auth/mfa/setup/totp/verify
POST   /auth/mfa/disable
GET    /auth/mfa/backup-codes
```

**Admin (JWT + RBAC)**:
```
GET    /admin/users                    # user:read
GET    /admin/users/{id}               # user:read
GET    /admin/users/{id}/audit-log     # user:read
PUT    /admin/users/{id}/status        # user:write
DELETE /admin/users/{id}               # user:delete
POST   /admin/users/invite             # user:write
POST   /admin/users/invite/{id}/resend # user:write
DELETE /admin/users/invite/{id}        # user:write
GET    /admin/rbac/roles               # role:read
GET    /admin/rbac/matrix              # role:read
GET    /admin/rbac/permissions         # role:read
GET    /admin/rbac/users/{id}/roles    # role:read
GET    /admin/rbac/users/{id}/permissions  # role:read
POST   /admin/rbac/roles               # role:write
PUT    /admin/rbac/roles/{id}          # role:write
POST   /admin/rbac/roles/{id}/clone    # role:write
PUT    /admin/rbac/roles/{id}/permissions  # role:write
POST   /admin/rbac/users/assign        # role:write
POST   /admin/rbac/users/assign/bulk   # role:write
DELETE /admin/rbac/users/unassign      # role:write
DELETE /admin/rbac/roles/{id}          # role:delete
```

---

## 6. User Stories

### US-001: Header Org Dropdown (P0, 2h)
> As a multi-org user, I want to see and switch organizations from the header.

**Acceptance Criteria**:
- [ ] Header shows current org name with dropdown indicator
- [ ] Dropdown lists all user's orgs with role and vehicle count
- [ ] Selecting org calls `POST /api/v1/auth/orgs/{orgId}/switch`
- [ ] New JWT tokens stored in Redux
- [ ] Dashboard data refreshes
- [ ] Toast: "Switched to {org_name}"

### US-002: Current Org Display (P0, 1h)
> As a user, I want to always see which organization I'm in.

**Acceptance Criteria**:
- [ ] Current org name displayed in header next to org icon
- [ ] If user has only 1 org, show name without dropdown
- [ ] If user has 2+ orgs, show name with dropdown indicator

### US-003: Fix setOrganization (P1, 0.5h)
> As a developer, I want setOrganization to use the correct API endpoint.

**Acceptance Criteria**:
- [ ] `AuthProvider.tsx` calls `POST /api/v1/auth/orgs/{orgId}/switch` instead of `/auth/refresh`
- [ ] Response tokens stored correctly
- [ ] No page reload on switch

### US-004: Role Clone Button (P1, 1h)
> As an admin, I want to clone a role from the roles list.

**Acceptance Criteria**:
- [ ] "Clone" icon/button on each role row in `Roles.tsx`
- [ ] Clicking calls `POST /api/v1/admin/rbac/roles/{id}/clone`
- [ ] Cloned role appears in list with "(Copy)" suffix
- [ ] Toast confirmation

### US-005: Bulk Role Assignment (P2, 3h)
> As an admin, I want to assign roles to multiple users at once.

**Acceptance Criteria**:
- [ ] Checkboxes on user rows in `UserAssignments.tsx`
- [ ] "Select All" checkbox
- [ ] Action bar appears when users selected: "Assign Role" dropdown + button
- [ ] Calls `POST /api/v1/admin/rbac/users/assign/bulk`
- [ ] Toast with count: "Role assigned to 5 users"

---

## 7. Technical Specifications

### 7.1 Monolith Entry Point

```go
// cmd/main.go
func main() {
    cfg := config.Load()
    db := database.Connect(cfg.Database)
    redis := cache.Connect(cfg.Redis)
    nats := messaging.Connect(cfg.NATS)

    r := chi.NewRouter()

    // Shared middleware
    r.Use(cors.Handler)
    r.Use(securityHeaders)
    r.Use(requestLogger)

    r.Route("/api/v1", func(r chi.Router) {
        // IAM Module
        iam := iam.NewModule(db, redis, nats, cfg.IAM)
        iam.RegisterRoutes(r)

        // Fleet Module
        fleet := fleet.NewModule(db, redis, nats, cfg.Fleet)
        fleet.RegisterRoutes(r)

        // ... other modules
    })

    http.ListenAndServe(":"+cfg.Port, r)
}
```

### 7.2 IAM Module Registration

```go
// internal/modules/iam/module.go
type Module struct {
    authHandler    *handler.AuthHandler
    rbacHandler    *handler.RBACHandler
    mfaHandler     *handler.MFAHandler
    profileHandler *handler.ProfileHandler
    sessionHandler *handler.SessionHandler
    userHandler    *handler.UserHandler
    inviteHandler  *handler.InviteHandler
}

func (m *Module) RegisterRoutes(r chi.Router) {
    // Public routes
    r.Post("/auth/identify", m.authHandler.Identify)
    r.Post("/auth/login", m.authHandler.Login)
    r.Post("/auth/forgot-password", m.authHandler.ForgotPassword)
    // ... all 38 endpoints

    // Authenticated routes
    r.Group(func(r chi.Router) {
        r.Use(authMiddleware)
        r.Post("/auth/logout", m.authHandler.Logout)
        r.Post("/auth/orgs/{orgId}/switch", m.authHandler.SwitchOrg)
        // ...
    })

    // Admin routes
    r.Group(func(r chi.Router) {
        r.Use(authMiddleware)
        r.Use(rbacMiddleware)
        r.Get("/admin/rbac/matrix", m.rbacHandler.GetPermissionMatrix)
        // ...
    })
}
```

### 7.3 Middleware Chain (Single Binary)

```
Request → CORS → Rate Limit → Security Headers → Auth (JWT) → RBAC → Org Scope → Handler
```

Same chain as before, but all in-process. No network hops.

### 7.4 Frontend API Client

No changes needed. `apiClient.ts` already:
- Uses `baseURL: '/api/v1'`
- Sends `Authorization: Bearer {token}` header
- Sends `X-Org-ID: {org_id}` from localStorage
- Handles 401/403 by redirecting to login

Only change: `AuthProvider.tsx` line 233 — swap `/auth/refresh` for `/auth/orgs/${orgId}/switch`.

---

## 8. Prioritization & Timeline

### 8.1 Effort Estimate

| Task | Effort | Dependencies |
|------|--------|-------------|
| Scaffold monolith structure | 1 day | None |
| Move auth-service → IAM module | 1 day | Scaffold |
| Move fleet-core domains → modules | 1 day | Scaffold |
| Single router + middleware | 0.5 day | Both moves |
| Integration testing | 0.5 day | Router |
| G1+G2: Header org dropdown + display | 0.5 day | None |
| G3: Role clone button | 0.25 day | None |
| G4: Bulk assign UI | 0.5 day | None |
| G5: Fix setOrganization | 0.25 day | None |
| E2E tests + polish | 0.5 day | All above |
| **Total** | **~6 days** | |

### 8.2 Timeline

```
Day 1-2   Monolith scaffold + module migration
Day 3     Single router + integration testing
Day 4     G1+G2: Header org dropdown + current org display
Day 5     G3: Role clone + G4: Bulk assign + G5: Fix setOrganization
Day 6     E2E tests + Docker + deploy
```

---

## 9. Risks

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|-----------|
| Module coupling during migration | Medium | High | Keep module interfaces narrow — only expose what other modules need |
| Import cycle between modules | Medium | Medium | Use interfaces in `shared/` package, not direct type references |
| Frontend org switch breaks data | Low | Critical | RLS policies validated with tests; org_id in JWT + query param |
| Duplicate auth implementations | High | Medium | Deprecate fleet-core auth v2 immediately; single source of truth |

---

## 10. Team & Budget

| Role | FTE | Duration |
|------|-----|----------|
| Backend Engineer (Go) | 1 | 3 days (monolith consolidation) |
| Frontend Engineer (React) | 1 | 2 days (5 UI gaps) |
| **Total** | | **~6 days, 2 engineers** |

---

## 11. Appendix

### A. Module Dependency Graph

```
monolith
├── iam module → (no dependencies on other modules)
├── fleet module → depends on iam (UserProvider, OrgResolver)
├── vehicle module → depends on iam (OrgResolver), fleet
├── driver module → depends on iam (OrgResolver), fleet
├── billing module → depends on iam (UserProvider)
└── analytics module → depends on all (read-only)
```

### B. What Changes vs Current State

| Aspect | Current | Target |
|--------|---------|--------|
| Deployables | 2 services | 1 binary |
| Ports | 8080 + 8084 | 8080 |
| Auth implementations | 2 (auth-service + fleet-core auth v2) | 1 (IAM module) |
| Inter-module calls | HTTP | Direct function calls |
| Docker containers | 2+ | 1 |
| Frontend gaps | 5 unaddressed | 5 addressed |

### C. What Stays the Same

- All 38 API endpoints (same paths, same request/response formats)
- All database tables and migrations
- All email providers (SendGrid, SMTP, NoOp)
- All middleware logic (JWT, RBAC, org scope, rate limiting)
- Frontend API client (axios, baseURL, headers)
- Password policy (bcrypt 12, history 5, min 12 chars)

---

*Version 3.0 (Modular Monolith) · 2026-04-04 · Next Review: 2026-04-11*
