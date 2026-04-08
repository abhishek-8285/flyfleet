# Technical Specification: IAM & User Management

> **FlyFleet — Enterprise Fleet Management Platform**
> Tech Spec Version 1.0 · 2026-04-04 · Based on Product Spec v1.0
> Status: **Draft** — Ready for Review

---

## Table of Contents

1. [Architecture Decision](#1-architecture-decision)
2. [Service Design](#2-service-design)
3. [Database Schema & Migrations](#3-database-schema--migrations)
4. [API Contract](#4-api-contract)
5. [Authentication Flow](#5-authentication-flow)
6. [RBAC & Permission System](#6-rbac--permission-system)
7. [Session Management](#7-session-management)
8. [Email Delivery](#8-email-delivery)
9. [Middleware Chain](#9-middleware-chain)
10. [Frontend Integration](#10-frontend-integration)
11. [Security Implementation](#11-security-implementation)
12. [Testing Strategy](#12-testing-strategy)
13. [Observability](#13-observability)
14. [Deployment & Rollout](#14-deployment--rollout)
15. [Migration Plan](#15-migration-plan)
16. [Open Decisions](#16-open-decisions)

---

## 1. Architecture Decision

### 1.1 Decision: Build as New Auth-Service Microservice

**ADR-001: Auth Service Placement**

| Option | Pros | Cons | Verdict |
|--------|------|------|---------|
| **New auth-service** (Port 8080) | Clean separation, independent scaling, follows microservice pattern | Network hop, distributed tracing needed | ✅ **Selected** |
| Revive monolith | Simpler deployment | Contradicts architecture, hard to scale auth independently | ❌ Rejected |
| Add to web-bff | No extra service | Violates separation of concerns, BFF becomes fat | ❌ Rejected |

**Rationale**: The product spec requires ~45 API endpoints, complex business logic (MFA, RBAC, sessions, email), and independent scaling (10,000 concurrent logins/sec). A dedicated microservice is the only architecture that supports these requirements.

### 1.2 System Architecture

```
┌─────────────┐     ┌─────────────┐     ┌──────────────────┐
│ Dashboard   │     │ Mobile App  │     │ API Clients      │
│ (React)     │     │ (React)     │     │ (REST/gRPC)      │
└──────┬──────┘     └──────┬──────┘     └────────┬─────────┘
       │                   │                     │
       └───────────────────┼─────────────────────┘
                           │
                    ┌──────▼──────┐
                    │   web-bff   │  Port 8081
                    │             │  ─ API aggregation
                    │             │  ─ Request routing
                    │             │  ─ Rate limiting (edge)
                    └──────┬──────┘
                           │
              ┌────────────┼────────────────┐
              │            │                │
       ┌──────▼──────┐ ┌───▼───────┐ ┌─────▼────────┐
       │ auth-service│ │settings-  │ │fleet-core-   │
       │    :8080    │ │svc :8091  │ │svc   :8084   │
       │             │ │           │ │              │
       │ • Login     │ │ • Profile │ │ • Org data   │
       │ • MFA       │ │ • Config  │ │ • Vehicles   │
       │ • RBAC      │ │           │ │ • Trips      │
       │ • Sessions  │ │           │ │              │
       │ • Password  │ │           │ │              │
       │ • Org switch│ │           │ │              │
       │ • Invites   │ │           │ │              │
       └──────┬──────┘ └───────────┘ └──────────────┘
              │
    ┌─────────┼──────────────┐
    │         │              │
 ┌──▼──┐  ┌───▼───┐  ┌──────▼──────┐
 │Post-│  │ Redis │  │ NATS        │
 │greSQL│  │       │  │ JetStream   │
 │     │  │ •Sessions  │ • Audit events
 │ •Users     │ • Rate limits│ • Lockout events
 │ •Roles     │ • Lockouts   │ • Email queue
 │ •Tokens    │ • Temp tokens│
 │ •MFA       │              │
 └─────┘  └─────────┘  └─────────────┘
```

### 1.3 Service Dependencies

| Dependency | Purpose | Fallback |
|------------|---------|----------|
| PostgreSQL 15+ | Primary data store | Read-only mode (cached tokens) |
| Redis 7+ | Session cache, rate limits, MFA codes | Direct DB queries (slower) |
| NATS JetStream | Async events (audit, email, lockout) | Synchronous fallback |
| SMTP/SendGrid | Email delivery | Queue + retry (3 attempts) |

### 1.4 Shared Package Reuse

| Package | Path | Usage |
|---------|------|-------|
| `security/jwt.go` | `packages/common/go/pkg/security/` | JWT generation/verification |
| `security/rbac_policies.go` | `packages/common/go/pkg/security/` | Policy manifest, role constants |
| `security/tenant.go` | `packages/common/go/pkg/security/` | Org ID context extraction |
| `security/security.go` | `packages/common/go/pkg/security/` | PII masking, audit logger, gRPC interceptor |

---

## 2. Service Design

### 2.1 Project Structure

```
services/auth-service/
├── cmd/
│   └── main.go                    # Entry point, router, DI wire
├── internal/
│   ├── config/
│   │   └── config.go              # Viper-based config (port, DB, JWT, SMTP)
│   ├── database/
│   │   ├── connection.go          # pgx pool initialization
│   │   └── migrations.go          # Migration runner (embed FS)
│   ├── redis/
│   │   ├── client.go              # Redis client initialization
│   │   ├── session_store.go       # Session caching
│   │   ├── rate_limiter.go        # Token bucket rate limiter
│   │   └── lockout_store.go       # Account lockout tracking
│   ├── models/
│   │   ├── user.go                # User domain model
│   │   ├── organization.go        # Organization model
│   │   ├── role.go                # Role & permission models
│   │   ├── session.go             # Session model
│   │   ├── mfa.go                 # MFA code/model
│   │   ├── invitation.go          # Invitation model
│   │   ├── audit.go               # Audit event model
│   │   └── requests.go            # Request/Response DTOs
│   ├── repository/
│   │   ├── user_repo.go           # User CRUD, search, status
│   │   ├── org_repo.go            # Org queries, user-org mapping
│   │   ├── role_repo.go           # Role CRUD, permission matrix
│   │   ├── session_repo.go        # Token persistence
│   │   ├── mfa_repo.go            # MFA code storage
│   │   ├── invitation_repo.go     # Invite token management
│   │   ├── audit_repo.go          # Audit log writes
│   │   └── password_history_repo.go  # Password history checks
│   ├── service/
│   │   ├── auth_service.go        # Login, identify, MFA orchestration
│   │   ├── password_service.go    # Reset, change, policy, history
│   │   ├── session_service.go     # Session lifecycle, revocation
│   │   ├── org_switch_service.go  # Org context switching
│   │   ├── mfa_service.go         # TOTP generation/verification, SMS
│   │   ├── rbac_service.go        # Role CRUD, permission matrix, cloning
│   │   ├── user_service.go        # User CRUD, search, status management
│   │   ├── invitation_service.go  # Invite creation, acceptance
│   │   ├── email_service.go       # Email templating, delivery
│   │   └── lockout_service.go     # Account lockout, recovery
│   ├── handlers/
│   │   ├── auth_handler.go        # /auth/* endpoints
│   │   ├── user_handler.go        # /admin/users/* endpoints
│   │   ├── rbac_handler.go        # /admin/rbac/* endpoints
│   │   ├── session_handler.go     # /auth/sessions/* endpoints
│   │   ├── profile_handler.go     # /auth/profile endpoints
│   │   └── invite_handler.go      # /auth/invite/* endpoints
│   ├── middleware/
│   │   ├── auth.go                # JWT validation, claim injection
│   │   ├── rbac.go                # Permission checking
│   │   ├── org_scope.go           # Organization context enforcement
│   │   ├── rate_limit.go          # IP/email/endpoint rate limiting
│   │   ├── cors.go                # CORS configuration
│   │   ├── csrf.go                # CSRF protection
│   │   ├── security_headers.go    # CSP, HSTS, X-Frame-Options
│   │   └── audit.go               # Request/response audit logging
│   ├── email/
│   │   ├── templates/
│   │   │   ├── password_reset.html
│   │   │   ├── password_reset.txt
│   │   │   ├── invitation.html
│   │   │   ├── invitation.txt
│   │   │   ├── account_locked.html
│   │   │   └── account_locked.txt
│   │   ├── provider.go            # Email provider interface
│   │   ├── sendgrid.go            # SendGrid implementation
│   │   └── smtp.go                # SMTP fallback
│   └── nats/
│       ├── publisher.go           # NATS JetStream publisher
│       └── events.go              # Event type definitions
├── migrations/                    # Embedded SQL migrations
│   ├── 001_iam_foundation.sql
│   ├── 002_invitations.sql
│   ├── 003_password_history.sql
│   └── 004_org_theme.sql
├── Dockerfile
├── go.mod
├── go.sum
├── config.toml
└── README.md
```

### 2.2 Configuration

**config.toml**:
```toml
[server]
port = 8080
host = "0.0.0.0"
log_level = "info"
read_timeout = "10s"
write_timeout = "30s"
idle_timeout = "60s"

[database]
url = "${DATABASE_URL}"
max_open_conns = 25
max_idle_conns = 5
conn_max_lifetime = "30m"

[redis]
url = "${REDIS_URL}"
pool_size = 10

[nats]
url = "${NATS_URL}"
stream = "AUTH_EVENTS"

[security]
jwt_secret = "${JWT_SECRET}"
jwt_access_ttl = "15m"
jwt_refresh_ttl = "7d"
jwt_action_ttl = "30m"
bcrypt_cost = 12
max_login_attempts = 5
lockout_duration = "15m"
password_history_length = 5

[email]
provider = "sendgrid"  # sendgrid | smtp
from_address = "noreply@flyfleet.io"
from_name = "FlyFleet"
sendgrid_api_key = "${SENDGRID_API_KEY}"
smtp_host = "${SMTP_HOST}"
smtp_port = 587
smtp_user = "${SMTP_USER}"
smtp_password = "${SMTP_PASSWORD}"

[rate_limit]
login_per_ip = "5/15m"
reset_per_email = "3/15m"
reset_per_ip = "5/15m"
identify_per_ip = "10/15m"

[cors]
allowed_origins = ["http://localhost:5173", "https://app.flyfleet.io"]
allowed_methods = ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
allowed_headers = ["Authorization", "Content-Type", "X-Org-ID"]
allow_credentials = true
max_age = "3600s"
```

### 2.3 go.mod Dependencies

```go
module github.com/flyfleet/auth-service

go 1.25.7

require (
    github.com/golang-jwt/jwt/v5 v5.2.0
    github.com/google/uuid v1.6.0
    github.com/jackc/pgx/v5 v5.7.0
    github.com/redis/go-redis/v9 v9.7.0
    github.com/nats-io/nats.go v1.39.1
    github.com/spf13/viper v1.19.0
    go.uber.org/zap v1.27.0
    golang.org/x/crypto v0.31.0          // bcrypt, TOTP
    github.com/pquerna/otp v1.4.0        // TOTP generation
    github.com/sendgrid/sendgrid-go v3.14.0
    github.com/go-chi/chi/v5 v5.2.0      // HTTP router
    github.com/go-chi/cors v1.2.1        // CORS middleware
    github.com/go-chi/rateLimiter v1.0.0 // Rate limiting
    github.com/rs/cors v1.11.0           // Alternative CORS
)
```

---

## 3. Database Schema & Migrations

### 3.1 Migration Strategy

**Approach**: Build on existing 90+ migrations. Create 4 new migrations for IAM-specific gaps.

**Existing migrations that cover IAM** (no changes needed):
- `001_enterprise_auth_schema.sql` — users, organizations, user_organizations, auth_tokens, mfa_codes
- `007_mfa_security_suite.sql` — MFA enhancements
- `008_mfa_policy_hardening.sql` — MFA policies per org
- `010_audit_module_expansion.sql` — audit_logs
- `017_security_operations.sql` — security_events
- `040_rls_foundation.sql` — Row-level security
- `060_bitmask_permissions.sql` — permission_mask on users

### 3.2 New Migration: 001_iam_foundation.sql

```sql
-- IAM User Management: Foundation Tables
-- Adds missing tables referenced in product spec but not yet created

-- ───────────────────────────────────────────────────────────────────
-- 1. ROLES TABLE (if not exists)
-- ───────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT false,  -- system roles cannot be deleted
    permission_mask BIGINT DEFAULT 0, -- 64-bit permission bitmask
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(org_id, name)
);

-- ───────────────────────────────────────────────────────────────────
-- 2. PERMISSIONS TABLE (fine-grained permission catalog)
-- ───────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource VARCHAR(100) NOT NULL,       -- vehicles, drivers, users, reports
    action VARCHAR(50) NOT NULL,          -- create, read, update, delete, export
    description TEXT,
    bit_position INT UNIQUE NOT NULL,     -- 0-63 for BIGINT bitmask
    UNIQUE(resource, action)
);

-- ───────────────────────────────────────────────────────────────────
-- 3. ROLE_PERMISSIONS (many-to-many junction)
-- ───────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    UNIQUE(role_id, permission_id)
);

-- ───────────────────────────────────────────────────────────────────
-- 4. USER_ROLES (many-to-many: users can have multiple roles)
-- ───────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    assigned_by UUID REFERENCES users(id),
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, role_id, org_id)
);

-- ───────────────────────────────────────────────────────────────────
-- 5. MFA_POLICIES (per-org MFA configuration)
-- ───────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS mfa_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    require_mfa BOOLEAN DEFAULT false,
    allowed_methods TEXT[] DEFAULT '{totp,sms}',  -- totp, sms, email
    remember_device_days INT DEFAULT 30,
    enforced_for_roles TEXT[] DEFAULT '{org_admin,fleet_manager}',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(org_id)
);

-- ───────────────────────────────────────────────────────────────────
-- 6. MFA_EVENTS (audit trail for MFA events)
-- ───────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS mfa_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,    -- setup_enabled, setup_disabled, code_verified, failed_attempt
    method VARCHAR(20),                  -- totp, sms
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ───────────────────────────────────────────────────────────────────
-- INDEXES
-- ───────────────────────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_roles_org_id ON roles(org_id);
CREATE INDEX IF NOT EXISTS idx_roles_org_deleted ON roles(org_id, deleted_at);
CREATE INDEX IF NOT EXISTS idx_permissions_resource_action ON permissions(resource, action);
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions(permission_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_org_id ON user_roles(org_id);
CREATE INDEX IF NOT EXISTS idx_mfa_policies_org_id ON mfa_policies(org_id);
CREATE INDEX IF NOT EXISTS idx_mfa_events_user_id ON mfa_events(user_id);
CREATE INDEX IF NOT EXISTS idx_mfa_events_type ON mfa_events(event_type);
```

### 3.3 New Migration: 002_invitations.sql

```sql
-- Invitation Onboarding

CREATE TABLE IF NOT EXISTS invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    invited_by UUID NOT NULL REFERENCES users(id),
    token VARCHAR(512) NOT NULL UNIQUE,
    role_ids UUID[] NOT NULL DEFAULT '{}',
    message TEXT,
    status VARCHAR(20) DEFAULT 'pending',  -- pending, accepted, expired, revoked
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    accepted_at TIMESTAMP WITH TIME ZONE,
    accepted_user_id UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_invitations_org_id ON invitations(org_id);
CREATE INDEX IF NOT EXISTS idx_invitations_email ON invitations(email);
CREATE INDEX IF NOT EXISTS idx_invitations_token ON invitations(token);
CREATE INDEX IF NOT EXISTS idx_invitations_status ON invitations(status);
CREATE INDEX IF NOT EXISTS idx_invitations_expires_at ON invitations(expires_at);
```

### 3.4 New Migration: 003_password_history.sql

```sql
-- Password History Enforcement

CREATE TABLE IF NOT EXISTS password_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_password_history_user_id ON password_history(user_id);
CREATE INDEX IF NOT EXISTS idx_password_history_created ON password_history(user_id, created_at DESC);

-- Add lockout columns to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS failed_logins INT DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_until TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_ip INET;
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(512);
```

### 3.5 New Migration: 004_org_theme.sql

```sql
-- Organization theming for org switch UI

ALTER TABLE organizations ADD COLUMN IF NOT EXISTS theme_color VARCHAR(7);      -- hex color
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS theme_logo_url VARCHAR(512); -- logo URL
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS default_role_id UUID REFERENCES roles(id);

-- Add theme to user_organizations for per-user role in org
ALTER TABLE user_organizations ADD COLUMN IF NOT EXISTS role_id UUID REFERENCES roles(id);

-- Org switch audit trail
CREATE TABLE IF NOT EXISTS org_switch_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    from_org_id UUID REFERENCES organizations(id),
    to_org_id UUID NOT NULL REFERENCES organizations(id),
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_org_switch_user_id ON org_switch_events(user_id);
CREATE INDEX IF NOT EXISTS idx_org_switch_created ON org_switch_events(created_at DESC);
```

### 3.6 Enhanced Users Table

```sql
-- Final users table structure after all migrations:
--
-- id                    UUID         PK
-- email                 VARCHAR(255) UNIQUE, NOT NULL
-- password_hash         TEXT         NOT NULL
-- name                  VARCHAR(255)
-- phone                 VARCHAR(50)
-- avatar_url            VARCHAR(512)
-- mfa_enabled           BOOLEAN      DEFAULT false
-- mfa_secret            TEXT         -- TOTP secret (encrypted at rest)
-- mfa_backup_codes      TEXT[]       -- encrypted backup codes
-- status                VARCHAR(50)  DEFAULT 'active' -- active, disabled, locked
-- email_verified        BOOLEAN      DEFAULT false
-- failed_logins         INT          DEFAULT 0
-- locked_until          TIMESTAMP    -- account lockout expiry
-- last_login_at         TIMESTAMP
-- last_login_ip         INET
-- permission_mask       BIGINT       DEFAULT 0  -- from 060_bitmask_permissions
-- created_at            TIMESTAMP
-- updated_at            TIMESTAMP
```

### 3.7 Seed Data: Default Permissions

```sql
-- Seed the permission catalog
INSERT INTO permissions (resource, action, description, bit_position) VALUES
('vehicles', 'create', 'Create new vehicles', 0),
('vehicles', 'read', 'View vehicle details and list', 1),
('vehicles', 'update', 'Edit vehicle information', 2),
('vehicles', 'delete', 'Delete vehicles', 3),
('vehicles', 'export', 'Export vehicle data', 4),
('drivers', 'create', 'Add new drivers', 5),
('drivers', 'read', 'View driver profiles', 6),
('drivers', 'update', 'Edit driver information', 7),
('drivers', 'delete', 'Remove drivers', 8),
('drivers', 'export', 'Export driver data', 9),
('trips', 'create', 'Create new trips', 10),
('trips', 'read', 'View trip details', 11),
('trips', 'update', 'Edit trip information', 12),
('trips', 'delete', 'Delete trips', 13),
('trips', 'export', 'Export trip data', 14),
('users', 'create', 'Invite/create users', 15),
('users', 'read', 'View user profiles', 16),
('users', 'update', 'Edit user settings', 17),
('users', 'delete', 'Disable/delete users', 18),
('roles', 'create', 'Create custom roles', 19),
('roles', 'read', 'View role definitions', 20),
('roles', 'update', 'Edit role permissions', 21),
('roles', 'delete', 'Delete custom roles', 22),
('reports', 'read', 'View reports', 23),
('reports', 'export', 'Export reports', 24),
('audit_logs', 'read', 'View audit logs', 25),
('audit_logs', 'export', 'Export audit logs', 26),
('settings', 'read', 'View org settings', 27),
('settings', 'update', 'Edit org settings', 28),
('alerts', 'read', 'View alerts', 29),
('alerts', 'resolve', 'Resolve alerts', 30),
('maintenance', 'create', 'Create maintenance tasks', 31),
('maintenance', 'read', 'View maintenance data', 32),
('maintenance', 'update', 'Update maintenance status', 33),
('fuel', 'read', 'View fuel consumption', 34),
('fuel', 'create', 'Log fuel entries', 35),
('billing', 'read', 'View billing info', 36),
('billing', 'export', 'Export billing data', 37),
('analytics', 'read', 'View analytics dashboard', 38),
('analytics', 'export', 'Export analytics', 39),
('integrations', 'read', 'View integrations', 40),
('integrations', 'manage', 'Configure integrations', 41),
('governance', 'read', 'View governance dashboard', 42),
('governance', 'manage', 'Manage governance policies', 43);
```

### 3.8 Seed Data: Default System Roles

```sql
-- System roles (is_system = true, cannot be edited/deleted)
-- These are org-scoped — each org gets its own set
-- Created programmatically during org creation via service layer
```

**Role Permission Matrix (seed)**:

| Permission | Super Admin | Org Admin | Fleet Manager | Driver | Viewer |
|-----------|:-----------:|:---------:|:-------------:|:------:|:------:|
| vehicles:create | ✅ | ✅ | ✅ | ❌ | ❌ |
| vehicles:read | ✅ | ✅ | ✅ | ✅ | ✅ |
| vehicles:update | ✅ | ✅ | ✅ | ❌ | ❌ |
| vehicles:delete | ✅ | ✅ | ❌ | ❌ | ❌ |
| vehicles:export | ✅ | ✅ | ✅ | ❌ | ✅ |
| drivers:create | ✅ | ✅ | ✅ | ❌ | ❌ |
| drivers:read | ✅ | ✅ | ✅ | ✅ | ✅ |
| drivers:update | ✅ | ✅ | ✅ | ❌ | ❌ |
| drivers:delete | ✅ | ✅ | ❌ | ❌ | ❌ |
| trips:create | ✅ | ✅ | ✅ | ✅ | ❌ |
| trips:read | ✅ | ✅ | ✅ | ✅ | ✅ |
| trips:update | ✅ | ✅ | ✅ | ✅ | ❌ |
| trips:delete | ✅ | ✅ | ❌ | ❌ | ❌ |
| users:create | ✅ | ✅ | ❌ | ❌ | ❌ |
| users:read | ✅ | ✅ | ✅ | ❌ | ✅ |
| users:update | ✅ | ✅ | ❌ | ❌ | ❌ |
| users:delete | ✅ | ✅ | ❌ | ❌ | ❌ |
| roles:* | ✅ | ✅ | ❌ | ❌ | ❌ |
| reports:read | ✅ | ✅ | ✅ | ✅ | ✅ |
| reports:export | ✅ | ✅ | ✅ | ❌ | ✅ |
| audit_logs:read | ✅ | ✅ | ✅ | ❌ | ✅ |
| audit_logs:export | ✅ | ✅ | ❌ | ❌ | ❌ |
| settings:read | ✅ | ✅ | ✅ | ❌ | ✅ |
| settings:update | ✅ | ✅ | ❌ | ❌ | ❌ |
| alerts:read | ✅ | ✅ | ✅ | ✅ | ✅ |
| alerts:resolve | ✅ | ✅ | ✅ | ❌ | ❌ |
| maintenance:* | ✅ | ✅ | ✅ | ❌ | ❌ |
| billing:read | ✅ | ✅ | ❌ | ❌ | ✅ |
| analytics:read | ✅ | ✅ | ✅ | ❌ | ✅ |
| integrations:manage | ✅ | ✅ | ❌ | ❌ | ❌ |
| governance:* | ✅ | ✅ | ❌ | ❌ | ❌ |

---

## 4. API Contract

### 4.1 Full Endpoint Registry

#### Authentication (Public)

| # | Method | Path | Description | Rate Limit |
|---|--------|------|-------------|------------|
| 1 | `POST` | `/api/v1/auth/identify` | Identify user, return orgs + MFA status | 10/IP/15m |
| 2 | `POST` | `/api/v1/auth/login` | Authenticate with password | 5/IP/15m |
| 3 | `POST` | `/api/v1/auth/mfa/send` | Trigger MFA challenge | 5/IP/15m |
| 4 | `POST` | `/api/v1/auth/mfa/verify` | Verify MFA code | 10/IP/15m |
| 5 | `POST` | `/api/v1/auth/refresh` | Refresh access token | 10/IP/15m |
| 6 | `POST` | `/api/v1/auth/logout` | Revoke current session | — |
| 7 | `POST` | `/api/v1/auth/forgot-password` | Request password reset email | 3/email/15m, 5/IP/15m |
| 8 | `POST` | `/api/v1/auth/reset-password` | Complete reset with token | 5/token/15m |
| 9 | `POST` | `/api/v1/auth/sso/google` | Google OAuth2 login | 10/IP/15m |
| 10 | `POST` | `/api/v1/auth/sso/microsoft` | Microsoft OAuth2 login | 10/IP/15m |
| 11 | `GET` | `/api/v1/auth/invite/{token}` | Validate invite token | 10/IP/15m |
| 12 | `POST` | `/api/v1/auth/invite/accept` | Accept invite, create account | 5/IP/15m |

#### Authentication (Bearer Required)

| # | Method | Path | Required Permission | Description |
|---|--------|------|-------------------|-------------|
| 13 | `GET` | `/api/v1/auth/profile` | — | Profile + orgs + permissions |
| 14 | `PUT` | `/api/v1/auth/profile` | — | Update profile (name, phone, avatar) |
| 15 | `POST` | `/api/v1/auth/change-password` | — | Change password (current required) |
| 16 | `GET` | `/api/v1/auth/password-policy` | — | Get org password policy |
| 17 | `GET` | `/api/v1/auth/orgs` | — | List user's orgs with roles |
| 18 | `POST` | `/api/v1/auth/orgs/{orgId}/switch` | — | Switch org, issue new JWT |

#### Sessions

| # | Method | Path | Required Permission | Description |
|---|--------|------|-------------------|-------------|
| 19 | `GET` | `/api/v1/auth/sessions` | — | List active sessions |
| 20 | `DELETE` | `/api/v1/auth/sessions/{id}` | — | Revoke specific session |
| 21 | `DELETE` | `/api/v1/auth/sessions/all` | — | Revoke all other sessions |

#### MFA

| # | Method | Path | Required Permission | Description |
|---|--------|------|-------------------|-------------|
| 22 | `POST` | `/api/v1/auth/mfa/setup/totp` | — | Generate TOTP secret + QR |
| 23 | `POST` | `/api/v1/auth/mfa/setup/totp/verify` | — | Verify TOTP setup |
| 24 | `POST` | `/api/v1/auth/mfa/setup/sms` | — | Register phone for SMS MFA |
| 25 | `POST` | `/api/v1/auth/mfa/setup/sms/verify` | — | Verify SMS MFA setup |
| 26 | `POST` | `/api/v1/auth/mfa/disable` | — | Disable MFA (requires password) |
| 27 | `GET` | `/api/v1/auth/mfa/backup-codes` | — | Get/regenerate backup codes |

#### Admin — User Management

| # | Method | Path | Required Permission | Description |
|---|--------|------|-------------------|-------------|
| 28 | `GET` | `/api/v1/admin/users` | users:read | List/search/filter users |
| 29 | `GET` | `/api/v1/admin/users/{id}` | users:read | User detail |
| 30 | `POST` | `/api/v1/admin/users/invite` | users:create | Invite via email |
| 31 | `POST` | `/api/v1/admin/users/invite/bulk` | users:create | Bulk CSV invite |
| 32 | `POST` | `/api/v1/admin/users/invite/{id}/resend` | users:create | Resend invite |
| 33 | `DELETE` | `/api/v1/admin/users/invite/{id}` | users:create | Revoke invite |
| 34 | `PUT` | `/api/v1/admin/users/{id}/status` | users:update | Activate/disable/lock |
| 35 | `DELETE` | `/api/v1/admin/users/{id}` | users:delete | Soft delete user |
| 36 | `GET` | `/api/v1/admin/users/{id}/audit-log` | audit_logs:read | User's audit trail |

#### Admin — RBAC

| # | Method | Path | Required Permission | Description |
|---|--------|------|-------------------|-------------|
| 37 | `GET` | `/api/v1/admin/rbac/roles` | roles:read | List roles with permissions |
| 38 | `POST` | `/api/v1/admin/rbac/roles` | roles:create | Create custom role |
| 39 | `PUT` | `/api/v1/admin/rbac/roles/{id}` | roles:update | Update role |
| 40 | `POST` | `/api/v1/admin/rbac/roles/{id}/clone` | roles:create | Clone role |
| 41 | `DELETE` | `/api/v1/admin/rbac/roles/{id}` | roles:delete | Delete custom role |
| 42 | `GET` | `/api/v1/admin/rbac/roles/{id}/permissions` | roles:read | Role's permissions |
| 43 | `PUT` | `/api/v1/admin/rbac/roles/{id}/permissions` | roles:update | Set role permissions |
| 44 | `GET` | `/api/v1/admin/rbac/matrix` | roles:read | Full permission matrix |
| 45 | `POST` | `/api/v1/admin/rbac/users/assign` | users:update | Assign role(s) to user(s) |
| 46 | `DELETE` | `/api/v1/admin/rbac/users/unassign` | users:update | Remove role from user |
| 47 | `GET` | `/api/v1/admin/rbac/users/{id}/roles` | users:read | User's roles |
| 48 | `GET` | `/api/v1/admin/rbac/users/{id}/permissions` | users:read | User's effective permissions |
| 49 | `GET` | `/api/v1/admin/rbac/permissions` | roles:read | List all permissions catalog |

#### Health & Diagnostics

| # | Method | Path | Auth | Description |
|---|--------|------|------|-------------|
| 50 | `GET` | `/health` | — | Health check |
| 51 | `GET` | `/health/ready` | — | Readiness check (DB, Redis, NATS) |
| 52 | `GET` | `/metrics` | — | Prometheus metrics |

### 4.2 Request/Response Schemas

#### POST /api/v1/auth/identify

**Request**:
```json
{
  "email": "rajesh@fleetcorp.com"
}
```

**Response 200**:
```json
{
  "exists": true,
  "email_verified": true,
  "organizations": [
    {
      "id": "org-001",
      "name": "FleetCorp India",
      "role": "org_admin",
      "vehicle_count": 2340,
      "theme_color": "#004bb5"
    },
    {
      "id": "org-002",
      "name": "FleetCorp Logistics",
      "role": "fleet_manager",
      "vehicle_count": 890,
      "theme_color": "#2d5016"
    }
  ],
  "mfa_required": true,
  "mfa_methods": ["totp", "sms"],
  "mfa_remembered": false,
  "sso_providers": ["google", "microsoft"],
  "account_locked": false
}
```

**Response 200 (user not found — anti-enumeration)**:
```json
{
  "exists": false,
  "organizations": [],
  "mfa_required": false,
  "mfa_methods": [],
  "mfa_remembered": false,
  "sso_providers": [],
  "account_locked": false
}
```

#### POST /api/v1/auth/login

**Request**:
```json
{
  "email": "rajesh@fleetcorp.com",
  "password": "CorrectHorseBatteryStaple!1",
  "org_id": "org-001",
  "remember_device": false
}
```

**Response 200 (no MFA)**:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "dGhpcyBpcyBhIHJlZnJl...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "usr-001",
    "email": "rajesh@fleetcorp.com",
    "name": "Rajesh Kumar",
    "avatar_url": "https://...",
    "role": "org_admin",
    "org_id": "org-001",
    "org_name": "FleetCorp India",
    "permissions": ["vehicles:read", "vehicles:write", "..."]
  }
}
```

**Response 200 (MFA required)**:
```json
{
  "mfa_required": true,
  "temp_token": "eyJ0eXAiOiJhY3Rpb24iLC...",
  "mfa_methods": ["totp"],
  "expires_in": 300
}
```

**Response 401**:
```json
{
  "error": "invalid_credentials",
  "message": "Invalid email or password.",
  "attempts_remaining": 3,
  "lockout_at": null
}
```

#### POST /api/v1/auth/mfa/verify

**Request**:
```json
{
  "temp_token": "eyJ0eXAiOiJhY3Rpb24iLC...",
  "code": "123456",
  "method": "totp",
  "remember_device": true
}
```

**Response 200**: Same as login success response.

#### POST /api/v1/auth/orgs/{orgId}/switch

**Response 200**:
```json
{
  "success": true,
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "dGhpcyBpcyBhIHJlZnJl...",
  "expires_in": 900,
  "organization": {
    "id": "org-002",
    "name": "FleetCorp Logistics",
    "role": "fleet_manager",
    "theme_color": "#2d5016"
  }
}
```

#### POST /api/v1/admin/users/invite

**Request**:
```json
{
  "email": "newuser@fleetcorp.com",
  "role_ids": ["role-001", "role-003"],
  "message": "Welcome to the team!"
}
```

**Response 201**:
```json
{
  "id": "inv-001",
  "email": "newuser@fleetcorp.com",
  "status": "pending",
  "expires_at": "2026-04-11T10:00:00Z",
  "invited_by": "usr-001",
  "invited_at": "2026-04-04T10:00:00Z"
}
```

#### GET /api/v1/admin/rbac/matrix

**Response 200**:
```json
{
  "roles": [
    {
      "id": "role-super",
      "name": "Super Admin",
      "is_system": true,
      "permissions": {
        "vehicles":   { "create": true, "read": true, "update": true, "delete": true, "export": true },
        "drivers":    { "create": true, "read": true, "update": true, "delete": true, "export": true },
        "users":      { "create": true, "read": true, "update": true, "delete": true, "export": false },
        "reports":    { "create": false, "read": true, "update": false, "delete": false, "export": true },
        "audit_logs": { "create": false, "read": true, "update": false, "delete": false, "export": true }
      }
    }
  ]
}
```

### 4.3 Error Response Format

```json
{
  "error": "error_code_snake_case",
  "message": "Human-readable message",
  "details": {},
  "request_id": "req-abc123"
}
```

**Error Codes**:

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `invalid_credentials` | 401 | Wrong email/password |
| `account_locked` | 423 | Account temporarily locked |
| `mfa_required` | 200 | MFA challenge needed (not an error) |
| `invalid_mfa_code` | 401 | Wrong MFA code |
| `mfa_expired` | 401 | MFA code expired |
| `invalid_token` | 401 | JWT invalid or expired |
| `user_not_found` | 404 | User doesn't exist |
| `role_not_found` | 404 | Role doesn't exist |
| `invite_expired` | 410 | Invitation token expired |
| `invite_already_used` | 409 | Invitation already accepted |
| `password_too_weak` | 400 | Password fails policy |
| `password_reused` | 400 | Password in history |
| `rate_limited` | 429 | Too many attempts |
| `permission_denied` | 403 | Insufficient permissions |
| `org_switch_failed` | 400 | Cannot switch to this org |
| `session_revoked` | 401 | Session was revoked |
| `internal_error` | 500 | Server error |

---

## 5. Authentication Flow

### 5.1 Complete Login Flow (Sequence)

```
┌──────┐     ┌──────┐     ┌──────────────┐     ┌──────┐     ┌───────┐
│Client│     │ BFF  │     │ Auth Service │     │Redis │     │PostgreSQL
└──┬───┘     └──┬───┘     └──────┬───────┘     └──┬───┘     └───┬───┘
   │            │               │                │              │
   │ 1.POST     │               │               │              │
   │ /identify  │               │               │              │
   │───────────>│               │               │              │
   │            │ 1a. Forward   │               │              │
   │            │──────────────>│               │              │
   │            │               │ 1b. GET user  │              │
   │            │               │─────────────────────────────>│
   │            │               │ 1c. user row  │              │
   │            │               │<─────────────────────────────│
   │            │               │ 1d. GET orgs  │              │
   │            │               │─────────────────────────────>│
   │            │               │ 1e. org list  │              │
   │            │               │<─────────────────────────────│
   │            │               │ 1f. GET mfa  │              │
   │            │               │ policy        │              │
   │            │               │─────────────────────────────>│
   │            │               │ 1g. mfa policy│              │
   │            │               │<─────────────────────────────│
   │            │ 1h. identify │               │              │
   │            │<──────────────│               │              │
   │ 1i. resp   │               │               │              │
   │<───────────│               │               │              │
   │            │               │               │              │
   │ 2.POST     │               │               │              │
   │ /login     │               │               │              │
   │───────────>│               │               │              │
   │            │ 2a. Forward   │               │              │
   │            │──────────────>│               │              │
   │            │               │ 2b. Rate check│              │
   │            │               │───────────────>│              │
   │            │               │ 2c. OK/denied │              │
   │            │               │<───────────────│              │
   │            │               │ 2d. GET user  │              │
   │            │               │─────────────────────────────>│
   │            │               │ 2e. bcrypt    │              │
   │            │               │   compare     │              │
   │            │               │               │              │
   │            │               │ [IF MFA]      │              │
   │            │               │ 2f. Generate  │              │
   │            │               │     temp JWT  │              │
   │            │               │ 2g. Return    │              │
   │            │               │     mfa_required              │
   │            │ 2h. resp     │               │              │
   │            │<──────────────│               │              │
   │ 2i. resp   │               │               │              │
   │<───────────│               │               │              │
   │            │               │               │              │
   │ 3.POST     │               │               │              │
   │ /mfa/verify│               │               │              │
   │───────────>│               │               │              │
   │            │ 3a. Forward   │               │              │
   │            │──────────────>│               │              │
   │            │               │ 3b. Verify TOTP                │
   │            │               │     (or SMS code in Redis)     │
   │            │               │               │              │
   │            │               │ 3c. Generate access + refresh  │
   │            │               │     tokens (JWT)               │
   │            │               │ 3d. Store session in Redis     │
   │            │               │───────────────>│              │
   │            │               │ 3e. Persist token in DB        │
   │            │               │─────────────────────────────>│
   │            │               │ 3f. Log audit event            │
   │            │               │─────────────────────────────>│
   │            │               │ 3g. Publish NATS event         │
   │            │ 3h. tokens   │               │              │
   │            │<──────────────│               │              │
   │ 3i. tokens │               │               │              │
   │<───────────│               │               │              │
```

### 5.2 Org Switch Flow

```
┌──────┐  ┌──────┐  ┌──────────────┐  ┌──────┐  ┌───────┐
│Client│  │ BFF  │  │ Auth Service │  │Redis │  │PostgreSQL
└──┬───┘  └──┬───┘  └──────┬───────┘  └──┬───┘  └───┬───┘
   │         │            │             │          │
   │ POST    │            │            │          │
   │ /orgs/:id/switch     │            │          │
   │────────>│            │            │          │
   │         │ Forward    │            │          │
   │         │───────────>│            │          │
   │         │            │ 1. Validate│          │
   │         │            │ user has   │          │
   │         │            │ access to  │          │
   │         │            │ target org │          │
   │         │            │─────────────────────>│
   │         │            │ org access │          │
   │         │            │<─────────────────────│
   │         │            │ 2. Get user's         │
   │         │            │ role in target org    │
   │         │            │─────────────────────>│
   │         │            │ role info  │          │
   │         │            │<─────────────────────│
   │         │            │ 3. Compute effective  │
   │         │            │ permissions           │
   │         │            │ 4. Generate new JWT   │
   │         │            │    (new org_id, role) │
   │         │            │ 5. Cache session       │
   │         │            │─────────>│            │
   │         │            │ 6. Log switch event    │
   │         │            │─────────────────────>│
   │         │            │ 7. Publish NATS event  │
   │         │ 200 OK     │            │          │
   │         │<───────────│            │          │
   │ 200 OK  │            │            │          │
   │<────────│            │            │          │
   │         │            │            │          │
   │ Dashboard refreshes with new org context      │
```

### 5.3 Password Reset Flow

```
┌──────┐  ┌──────┐  ┌──────────────┐  ┌──────┐  ┌────────┐
│Client│  │ BFF  │  │ Auth Service │  │Redis │  │Email   │
└──┬───┘  └──┬───┘  └──────┬───────┘  └──┬───┘  └───┬────┘
   │         │            │             │          │
   │ 1. POST /forgot-password            │          │
   │────────>│            │             │          │
   │         │ Forward    │             │          │
   │         │───────────>│             │          │
   │         │            │ 1a. Rate check (email+IP)
   │         │            │────────────>│          │
   │         │            │ 1b. GET user (silent fail if not found)
   │         │            │ 1c. Generate action token (30min)
   │         │            │ 1d. Queue email (async via NATS)
   │         │            │────────────────────────────────────>│
   │         │            │ 1e. Generic success response        │
   │         │ 200 OK     │             │          │
   │         │<───────────│             │          │
   │ 200 OK  │            │             │          │
   │<────────│            │             │          │
   │         │            │             │          │
   │ User clicks email link → /reset-password?token=xxx          │
   │         │            │             │          │
   │ 2. POST /reset-password             │          │
   │────────>│            │             │          │
   │         │ Forward    │             │          │
   │         │───────────>│             │          │
   │         │            │ 2a. Validate token (not expired, valid)
   │         │            │ 2b. Check password policy (strength)
   │         │            │ 2c. Check password history (last 5)
   │         │            │ 2d. Hash new password (bcrypt cost 12)
   │         │            │ 2e. Save to password_history         │
   │         │            │ 2f. Update user password_hash        │
   │         │            │ 2g. Revoke all sessions              │
   │         │            │────────────>│          │
   │         │            │ 2h. Log audit event                  │
   │         │ 200 OK     │             │          │
   │         │<───────────│             │          │
   │ 200 OK  │            │             │          │
   │<────────│            │             │          │
```

---

## 6. RBAC & Permission System

### 6.1 Dual Permission Model

The system uses **two complementary approaches**:

1. **Bitmask Permission** (fast path, for runtime checks)
   - `users.permission_mask` is a BIGINT (64 bits)
   - Each permission maps to a bit position (0-63)
   - O(1) bitwise check: `(permission_mask & (1 << bit_position)) != 0`
   - Used by middleware for hot-path authorization

2. **Role-Permission Tables** (admin path, for management)
   - `roles` → `role_permissions` → `permissions` normalized tables
   - Used for admin UI (matrix view, role CRUD, cloning)
   - Bitmask is derived from role_permissions on role change

### 6.2 Permission Resolution Algorithm

```
EffectivePermissions(user, org):
  1. Get all user_roles for (user_id, org_id)
  2. For each role:
     a. Get all role_permissions
     b. Accumulate permission set (union)
  3. Compute final bitmask from accumulated permissions
  4. Update user.permission_mask (cache)
  5. Return permission list
```

### 6.3 Bitmask Constants (Go)

```go
// pkg/security/permission_bits.go
const (
    PermVehiclesCreate  int64 = 1 << 0   // 1
    PermVehiclesRead    int64 = 1 << 1   // 2
    PermVehiclesUpdate  int64 = 1 << 2   // 4
    PermVehiclesDelete  int64 = 1 << 3   // 8
    PermVehiclesExport  int64 = 1 << 4   // 16
    PermDriversCreate   int64 = 1 << 5   // 32
    PermDriversRead     int64 = 1 << 6   // 64
    PermDriversUpdate   int64 = 1 << 7   // 128
    PermDriversDelete   int64 = 1 << 8   // 256
    PermDriversExport   int64 = 1 << 9   // 512
    // ... continue through bit 63
    PermAllPermissions  int64 = -1       // All bits set (super admin)
)

// HasPermission checks if a user's permission_mask includes the given permission
func HasPermission(mask int64, permission int64) bool {
    return (mask & permission) != 0
}

// AddPermission sets a permission bit in the mask
func AddPermission(mask int64, permission int64) int64 {
    return mask | permission
}

// RemovePermission clears a permission bit
func RemovePermission(mask int64, permission int64) int64 {
    return mask &^ permission
}
```

### 6.4 RBAC Middleware

```go
// middleware/rbac.go
func RBACMiddleware(requiredPermission string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract claims from context (set by AuthMiddleware)
            userID := middleware.GetUserID(r.Context())
            orgID := middleware.GetOrgID(r.Context())

            // Fast path: check bitmask
            mask, ok := middleware.GetPermissionMask(r.Context())
            if !ok {
                // Fallback: compute from roles (slower)
                mask = computeEffectivePermissions(userID, orgID)
            }

            bitPosition := permissionBits[requiredPermission]
            if !security.HasPermission(mask, 1<<bitPosition) {
                writeError(w, http.StatusForbidden, "PERMISSION_DENIED",
                    fmt.Sprintf("Requires %s permission", requiredPermission))
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### 6.5 Permission-to-Route Mapping

```go
// Route → permission mapping table
var RoutePermissions = map[string]string{
    "POST   /api/v1/admin/users":                "users:create",
    "GET    /api/v1/admin/users":                "users:read",
    "PUT    /api/v1/admin/users/{id}/status":    "users:update",
    "DELETE /api/v1/admin/users/{id}":           "users:delete",
    "GET    /api/v1/admin/rbac/roles":           "roles:read",
    "POST   /api/v1/admin/rbac/roles":           "roles:create",
    "PUT    /api/v1/admin/rbac/roles/{id}":      "roles:update",
    "DELETE /api/v1/admin/rbac/roles/{id}":      "roles:delete",
    "GET    /api/v1/admin/rbac/matrix":          "roles:read",
    "POST   /api/v1/admin/rbac/users/assign":    "users:update",
    // ... all admin routes
}
```

---

## 7. Session Management

### 7.1 Token Architecture

```
┌─────────────────────────────────────────────┐
│              Access Token (JWT)              │
│  Header: {"alg":"HS256","typ":"JWT"}       │
│  Payload:                                   │
│    sub: "user-uuid"                         │
│    user_id: "user-uuid"                     │
│    org_id: "org-uuid"                       │
│    role: "org_admin"                        │
│    permissions_mask: 9223372036854775807    │
│    type: "access"                           │
│    exp: 1712239200  (15 min)                │
│    iat: 1712238300                          │
│    jti: "token-uuid"                        │
│    device_fingerprint: "sha256-xxx"         │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│             Refresh Token (Opaque)          │
│  Format: cryptographically random 64 bytes  │
│  Storage: PostgreSQL + Redis (for rotation) │
│  TTL: 7 days                                │
│  Rotation: new refresh token on each use    │
│  Detection: if old token reused → revoke all│
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│            Action Token (JWT)               │
│  Used for: password reset, invite accept    │
│  TTL: 30 min (reset), 7 days (invite)       │
│  Claims: sub, type, org_id, jti, exp, iat   │
│  Single-use (jti tracked in Redis)          │
└─────────────────────────────────────────────┘
```

### 7.2 Session Storage

**Redis** (hot cache):
```
Key: session:{access_token_jti}
Value: {
  "user_id": "usr-001",
  "org_id": "org-001",
  "role": "org_admin",
  "device": "Chrome/Windows",
  "ip": "203.0.113.42",
  "created_at": "2026-04-04T10:00:00Z",
  "last_active": "2026-04-04T10:30:00Z"
}
TTL: 15 minutes (matches access token)

Key: refresh:{refresh_token_hash}
Value: {"user_id": "usr-001", "org_id": "org-001"}
TTL: 7 days

Key: lockout:{user_id}
Value: {"attempts": 5, "locked_until": "2026-04-04T10:15:00Z"}
TTL: 15 minutes

Key: mfa:{temp_token_jti}
Value: {"user_id": "usr-001", "org_id": "org-001"}
TTL: 5 minutes

Key: rate:login:{ip}
Value: count (integer)
TTL: 15 minutes
```

**PostgreSQL** (persistent):
```
auth_tokens table:
  - access_token (hashed, for lookup)
  - refresh_token (hashed)
  - device_id, device_info
  - ip_address, user_agent
  - expires_at
  - last_active_at
  - revoked_at (for soft revoke)
```

### 7.3 Token Rotation & Theft Detection

```
Client                    Auth Service              Redis
  │                          │                       │
  │  POST /refresh           │                       │
  │  {refresh_token: "A"}    │                       │
  │─────────────────────────>│                       │
  │                          │ 1. Check refresh:A     │
  │                          │───────────────────────>│
  │                          │ 2. EXISTS? Yes         │
  │                          │<───────────────────────│
  │                          │ 3. Check if already used│
  │                          │   (rotation marker)     │
  │                          │───────────────────────>│
  │                          │ 4. Already used?        │
  │                          │   YES → THEFT DETECTED  │
  │                          │   → Revoke ALL sessions │
  │                          │   → Return 401          │
  │                          │ 5. Not used → proceed   │
  │                          │ 6. Mark A as used       │
  │                          │───────────────────────>│
  │                          │ 7. Generate new tokens  │
  │                          │   access_token: B       │
  │                          │   refresh_token: C      │
  │                          │ 8. Store refresh:C       │
  │                          │───────────────────────>│
  │  {access: B, refresh: C}│                       │
  │<─────────────────────────│                       │
```

---

## 8. Email Delivery

### 8.1 Architecture

```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│ Auth Service│     │ NATS         │     │ Email Worker │
│             │     │ JetStream    │     │              │
│ ForgotPwd   │────>│ email_queue  │────>│ SendGrid/SMTP│
│ Invite      │     │ (durable)    │     │              │
│ Lockout     │     │              │     │              │
└─────────────┘     └──────────────┘     └──────────────┘
                                              │
                                              ▼
                                        User Inbox
```

**Why async?** Email delivery is slow (1-5s). Making it synchronous would block the login/reset flow. NATS JetStream provides:
- Durability (messages persist if worker is down)
- Retry with backoff (3 attempts: 1m, 5m, 15m)
- Dead letter queue (after 3 failures, log for manual review)

### 8.2 Email Provider Abstraction

```go
// email/provider.go
type EmailProvider interface {
    Send(ctx context.Context, to, subject string, bodyHTML, bodyText string) error
    SendTemplate(ctx context.Context, to, templateID string, data map[string]interface{}) error
}

// Use SendGrid in production, SMTP in dev/test
func NewEmailProvider(cfg *config.Config) EmailProvider {
    switch cfg.Email.Provider {
    case "sendgrid":
        return sendgrid.New(cfg.Email.SendGridAPIKey)
    case "smtp":
        return smtp.New(cfg.Email.SMTPHost, cfg.Email.SMTPPort, ...)
    default:
        return &mockEmailProvider{} // dev/test
    }
}
```

### 8.3 Email Templates

Templates stored as embedded Go files (`//go:embed`):

**Password Reset** (HTML + TXT):
```
Subject: Reset your FlyFleet password
Variables: {{.FirstName}}, {{.ResetURL}}, {{.ExpiryMinutes}}
```

**Invitation** (HTML + TXT):
```
Subject: You're invited to join {{.OrgName}} on FlyFleet
Variables: {{.Name}}, {{.InviterName}}, {{.OrgName}}, {{.AcceptURL}}, {{.ExpiryDays}}, {{.Message}}
```

**Account Locked** (HTML + TXT):
```
Subject: Your FlyFleet account has been locked
Variables: {{.Name}}, {{.LockedUntil}}, {{.UnlockURL}} (if MFA self-unlock enabled)
```

### 8.4 Anti-Spam Measures

| Measure | Implementation |
|---------|---------------|
| Rate limiting | 3/email/15m per email, 5/IP/15m |
| Generic response | Never confirm account existence |
| Link expiry | 30 min for reset, 7 days for invite |
| Single-use tokens | Token jti tracked in Redis |
| Unsubscribe | N/A (transactional emails only) |

---

## 9. Middleware Chain

### 9.1 Request Pipeline

```
Request
  │
  ▼
┌─────────────┐
│ 1. CORS     │     Cross-origin preflight, allowed origins
└──────┬──────┘
       ▼
┌─────────────┐
│ 2. Security │     CSP, HSTS, X-Frame-Options, X-Content-Type-Options
│    Headers  │     X-XSS-Protection, Referrer-Policy
└──────┬──────┘
       ▼
┌─────────────┐
│ 3. Rate     │     Token bucket per IP/endpoint
│    Limit    │     Skip: /health, /metrics
└──────┬──────┘
       ▼
┌─────────────┐
│ 4. Request  │     Correlation ID, request logging
│    ID       │
└──────┬──────┘
       ▼
  ┌──────────────────────┐
  │ Public endpoint?     │──── YES ──→ 5b. Skip to handler
  └──────────────────────┘
       │ NO
       ▼
┌─────────────┐
│ 5a. Auth    │     Parse Bearer token, verify signature
│    (JWT)    │     Extract claims → context
└──────┬──────┘
       ▼
┌─────────────┐
│ 6. Session  │     Check if session revoked (Redis)
│    Check    │     Update last_active timestamp
└──────┬──────┘
       ▼
┌─────────────┐
│ 7. RBAC     │     Check permission_mask or compute from roles
│    Check    │     Skip: profile endpoints (self-access)
└──────┬──────┘
       ▼
┌─────────────┐
│ 8. Org      │     Validate org_id in JWT matches route param
│    Scope    │     Inject org filter for data queries
└──────┬──────┘
       ▼
┌─────────────┐
│ 9. Handler  │     Business logic
└──────┬──────┘
       ▼
┌─────────────┐
│ 10. Audit   │     Log access event (async via NATS)
│    Log      │
└─────────────┘
```

### 9.2 Middleware Configuration

```go
// cmd/main.go - router setup
r := chi.NewRouter()

// Global middleware (all routes)
r.Use(middleware.RequestID)
r.Use(middleware.Logger)
r.Use(middleware.Recoverer)
r.Use(middleware.CORS(cfg.CORS))
r.Use(middleware.SecurityHeaders)

// Rate limiting (specific routes)
r.Group(func(r chi.Router) {
    r.Use(middleware.RateLimit(cfg.RateLimit.Login, "/api/v1/auth/login"))
    r.Post("/api/v1/auth/identify", authHandler.Identify)
    r.Post("/api/v1/auth/login", authHandler.Login)
    r.Post("/api/v1/auth/forgot-password", authHandler.ForgotPassword)
})

// Authenticated routes
r.Group(func(r chi.Router) {
    r.Use(middleware.AuthMiddleware(jwtManager))
    r.Use(middleware.SessionCheck(redisClient))
    r.Use(middleware.AuditLogger(natsPublisher))

    r.Get("/api/v1/auth/profile", profileHandler.GetProfile)
    r.Put("/api/v1/auth/profile", profileHandler.UpdateProfile)
    r.Get("/api/v1/auth/orgs", authHandler.ListOrgs)
    r.Post("/api/v1/auth/orgs/{orgId}/switch", authHandler.SwitchOrg)

    // Admin routes (require RBAC)
    r.Group(func(r chi.Router) {
        r.Use(middleware.RBACMiddleware("users:read")) // base requirement
        r.Get("/api/v1/admin/users", userHandler.ListUsers)
        r.Get("/api/v1/admin/users/{id}", userHandler.GetUser)
    })
})
```

---

## 10. Frontend Integration

### 10.1 Existing Frontend State

| Component | Path | Status | Changes Needed |
|-----------|------|--------|---------------|
| LoginPage.tsx | `apps/dashboard-web/src/pages/` | ✅ Ready | API endpoint paths, org switch flow |
| ForgotPassword.tsx | `apps/dashboard-web/src/pages/` | ✅ Ready | Connect to new backend |
| ResetPassword.tsx | `apps/dashboard-web/src/pages/` | ✅ Ready | Connect to new backend |
| Users.tsx | `apps/dashboard-web/src/pages/admin/` | ✅ Ready | Connect to admin user API |
| Roles.tsx | `apps/dashboard-web/src/pages/admin/rbac/` | ✅ Ready | Connect to RBAC API |
| RoleMatrix.tsx | `apps/dashboard-web/src/pages/admin/rbac/` | ✅ Ready | Connect to matrix API |
| EnterpriseAuthLayout | `apps/dashboard-web/src/components/auth/` | ✅ Ready | Minor tweaks for error states |
| AuthProvider | `apps/dashboard-web/src/context/` | ✅ Existing | Add org switching, session management |

### 10.2 AuthProvider Enhancement

```typescript
// apps/dashboard-web/src/context/AuthProvider.tsx
interface AuthContext {
  // Existing
  user: User | null;
  isAuthenticated: boolean;
  login: (accessToken: string, refreshToken: string, user: User) => void;
  logout: () => void;

  // NEW: Org switching
  organizations: Organization[];
  currentOrg: Organization | null;
  switchOrg: (orgId: string) => Promise<void>;

  // NEW: Session management
  sessions: Session[];
  revokeSession: (sessionId: string) => Promise<void>;
  revokeAllSessions: () => Promise<void>;

  // NEW: Profile
  updateProfile: (data: Partial<User>) => Promise<void>;
  changePassword: (current: string, newPwd: string) => Promise<void>;

  // NEW: MFA
  setupMFA: (method: 'totp' | 'sms') => Promise<void>;
  disableMFA: (password: string) => Promise<void>;
  backupCodes: string[] | null;
}
```

### 10.3 API Service Layer

```typescript
// apps/dashboard-web/src/services/authService.ts

export const authService = {
  // Public
  identify: (email: string) => api.post('/api/v1/auth/identify', { email }),
  login: (email: string, password: string, orgId: string) =>
    api.post('/api/v1/auth/login', { email, password, org_id: orgId }),
  verifyMFA: (tempToken: string, code: string) =>
    api.post('/api/v1/auth/mfa/verify', { temp_token: tempToken, code }),
  forgotPassword: (email: string) =>
    api.post('/api/v1/auth/forgot-password', { email }),
  resetPassword: (token: string, newPassword: string) =>
    api.post('/api/v1/auth/reset-password', { token, new_password: newPassword }),

  // Authenticated
  getProfile: () => api.get('/api/v1/auth/profile'),
  updateProfile: (data: Partial<User>) => api.put('/api/v1/auth/profile', data),
  changePassword: (currentPassword: string, newPassword: string) =>
    api.post('/api/v1/auth/change-password', { current_password: currentPassword, new_password: newPassword }),

  // Org switching
  getOrgs: () => api.get('/api/v1/auth/orgs'),
  switchOrg: (orgId: string) => api.post(`/api/v1/auth/orgs/${orgId}/switch`),

  // Sessions
  getSessions: () => api.get('/api/v1/auth/sessions'),
  revokeSession: (id: string) => api.delete(`/api/v1/auth/sessions/${id}`),
  revokeAllSessions: () => api.delete('/api/v1/auth/sessions/all'),

  // MFA
  setupTOTP: () => api.post('/api/v1/auth/mfa/setup/totp'),
  verifyTOTP: (code: string) => api.post('/api/v1/auth/mfa/setup/totp/verify', { code }),
  getBackupCodes: () => api.get('/api/v1/auth/mfa/backup-codes'),

  // Admin - Users
  listUsers: (params: UserFilterParams) => api.get('/api/v1/admin/users', { params }),
  getUser: (id: string) => api.get(`/api/v1/admin/users/${id}`),
  inviteUser: (email: string, roleIds: string[], message?: string) =>
    api.post('/api/v1/admin/users/invite', { email, role_ids: roleIds, message }),
  updateUserStatus: (id: string, status: 'active' | 'disabled' | 'locked', reason?: string) =>
    api.put(`/api/v1/admin/users/${id}/status`, { status, reason }),

  // Admin - RBAC
  getRoles: () => api.get('/api/v1/admin/rbac/roles'),
  createRole: (name: string, description: string) =>
    api.post('/api/v1/admin/rbac/roles', { name, description }),
  cloneRole: (id: string, name: string, description: string) =>
    api.post(`/api/v1/admin/rbac/roles/${id}/clone`, { name, description }),
  getPermissionMatrix: () => api.get('/api/v1/admin/rbac/matrix'),
  assignRole: (userId: string, roleId: string) =>
    api.post('/api/v1/admin/rbac/users/assign', { user_id: userId, role_id: roleId }),
  getUserPermissions: (userId: string) =>
    api.get(`/api/v1/admin/rbac/users/${userId}/permissions`),
};
```

### 10.4 Web-BFF Route Mapping

The web-bff (port 8081) must route auth endpoints to auth-service:

```javascript
// web-bff/routes/auth.js
router.post('/api/v1/auth/identify', proxy({ target: AUTH_SERVICE }));
router.post('/api/v1/auth/login', proxy({ target: AUTH_SERVICE }));
router.post('/api/v1/auth/mfa/verify', proxy({ target: AUTH_SERVICE }));
router.post('/api/v1/auth/refresh', proxy({ target: AUTH_SERVICE }));
router.post('/api/v1/auth/logout', proxy({ target: AUTH_SERVICE }));
router.get('/api/v1/auth/profile', proxy({ target: AUTH_SERVICE }));
// ... all auth routes

// web-bff/routes/admin.js
router.get('/api/v1/admin/users', proxy({ target: AUTH_SERVICE }));
router.get('/api/v1/admin/rbac/*', proxy({ target: AUTH_SERVICE }));
// ... all admin auth routes
```

---

## 11. Security Implementation

### 11.1 Threat Mitigation Matrix

| Threat | Attack Vector | Mitigation | Implementation |
|--------|--------------|-----------|----------------|
| **Credential stuffing** | Automated login with breached passwords | Rate limiting + breached password check | 5 attempts/IP/15m, HaveIBeenPwned API (Phase 3) |
| **User enumeration** | Timing analysis, error messages | Constant-time comparison, generic errors | `bcrypt.CompareHashAndPassword` on non-existent users, generic "if exists" message |
| **Token theft** | XSS, network interception | HTTP-only cookies, refresh token rotation, device fingerprinting | `SameSite=Strict`, rotate on each refresh, fingerprint in JWT |
| **Brute force** | Rapid password guessing | Account lockout, exponential backoff | Lock after 5 failures, 15m lockout, 2x increase on repeated |
| **CSRF** | Cross-site request forgery | Double-submit cookie, SameSite=Strict | `X-CSRF-Token` header required for state-changing requests |
| **XSS** | Script injection | CSP headers, input sanitization, HTTP-only | `Content-Security-Policy: default-src 'self'`, sanitize all inputs |
| **SQL injection** | Malicious query params | Parameterized queries | All DB queries use `pgx` parameterized queries |
| **Privilege escalation** | Forged JWT, tampered claims | Server-side RBAC + RLS | RBAC middleware + PostgreSQL RLS policies |
| **Session fixation** | Reuse session ID after login | New session ID on login, token rotation | Fresh JWT + refresh token on every login |
| **Password reset poisoning** | Intercepted reset link | Token bound to user_id + org_id, single-use, 30-min | Action JWT with `sub`, `type`, `exp` claims |
| **MFA bypass** | Client-side MFA toggle | Server-side enforced MFA state | MFA check in login flow, not client-trusted |
| **Org data leakage** | Switch org and access old data | RLS policies, org_id in every query | `ALTER TABLE ... ENABLE ROW LEVEL SECURITY` + `USING (org_id = current_setting('app.current_org')::uuid)` |
| **DoS (login)** | Volumetric attack on login | Edge rate limiting (nginx/BFF) + service-level | Token bucket at BFF, sliding window at service |
| **Password reuse** | Attacker forces old password | Password history (last 5) | `password_history` table, check before accepting new password |

### 11.2 Password Policy Engine

```go
// service/password_policy.go

type PasswordPolicy struct {
    MinLength       int      // default: 12
    RequireUpper    bool     // default: true
    RequireLower    bool     // default: true
    RequireDigit    bool     // default: true
    RequireSpecial  bool     // default: true
    MinStrengthScore int     // default: 75/100 (using zxcvbn-like scoring)
    HistoryLength   int      // default: 5
    BlockedPatterns []string // common passwords, org name, email localpart
}

type PasswordValidator struct {
    policy PasswordPolicy
    historyRepo *repository.PasswordHistoryRepo
}

func (v *PasswordValidator) Validate(ctx context.Context, userID uuid.UUID, password string) error {
    // 1. Length check
    if utf8.RuneCountInString(password) < v.policy.MinLength {
        return ErrPasswordTooShort
    }

    // 2. Character class checks
    if v.policy.RequireUpper && !hasUpper(password) {
        return ErrPasswordNoUpper
    }
    // ... lower, digit, special

    // 3. Strength score (simplified zxcvbn)
    score := v.calculateStrength(password)
    if score < v.policy.MinStrengthScore {
        return ErrPasswordTooWeak
    }

    // 4. Blocked patterns
    for _, pattern := range v.policy.BlockedPatterns {
        if strings.Contains(strings.ToLower(password), strings.ToLower(pattern)) {
            return ErrPasswordBlocked
        }
    }

    // 5. History check
    recentPasswords, err := v.historyRepo.GetRecent(ctx, userID, v.policy.HistoryLength)
    if err != nil {
        return err
    }
    for _, oldHash := range recentPasswords {
        if bcrypt.CompareHashAndPassword([]byte(oldHash), []byte(password)) == nil {
            return ErrPasswordReused
        }
    }

    return nil
}
```

### 11.3 Row-Level Security (PostgreSQL)

```sql
-- Enable RLS on all multi-tenant tables
ALTER TABLE roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE invitations ENABLE ROW LEVEL SECURITY;
ALTER TABLE org_switch_events ENABLE ROW LEVEL SECURITY;

-- RLS policies
CREATE POLICY org_isolation_roles ON roles
    USING (org_id = current_setting('app.current_org_id', true)::uuid
           OR current_setting('app.is_super_admin', true)::boolean = true);

CREATE POLICY org_isolation_user_roles ON user_roles
    USING (org_id = current_setting('app.current_org_id', true)::uuid
           OR current_setting('app.is_super_admin', true)::boolean = true);

CREATE POLICY org_isolation_invitations ON invitations
    USING (org_id = current_setting('app.current_org_id', true)::uuid
           OR current_setting('app.is_super_admin', true)::boolean = true);

-- Set org context per connection (called by middleware)
-- SET app.current_org_id = 'org-001';
-- SET app.is_super_admin = 'false';
```

---

## 12. Testing Strategy

### 12.1 Test Pyramid

```
                    ┌───────────┐
                   /  Manual   \    Security audit, pen testing
                  /   Testing   \   (1-2 days before launch)
                 ────────────────
                /    E2E Tests    \   Playwright: login, reset, org switch, RBAC
               /  (Integration)    \  (~30 scenarios)
              ──────────────────────
             /   Service Tests      \  Go tests: service layer logic, password policy,
            /   (Unit + DB)          \ token generation, MFA, RBAC resolution (~80 tests)
           ──────────────────────────
          /     Unit Tests            \ Pure function tests: bitmask ops, validation,
         /     (Pure Go)               \ email template rendering (~120 tests)
        ────────────────────────────────
```

### 12.2 Key Test Scenarios

**Unit Tests (pure Go)**:
- Bitmask permission operations (add, remove, check)
- Password policy validation (all failure modes)
- Email format validation
- Token claim generation and extraction
- Rate limiter token bucket logic

**Service Tests (with test DB)**:
- Login flow: success, wrong password, locked account, MFA required
- Password reset: valid, expired token, reused token, rate limited
- Org switch: valid, unauthorized org, non-existent org
- RBAC: create role, clone role, assign role, compute effective permissions
- Invitation: create, accept, expired, revoked, duplicate email

**E2E Tests (Playwright)**:
- Full login flow: identify → org select → password → MFA → dashboard
- Password reset: forgot → email link → set new password → login
- Org switch: dropdown → switch → dashboard refresh
- Admin: invite user → accept → appear in user list
- RBAC: create role → assign permissions → assign to user → verify access

### 12.3 Security Testing

| Test | Tool | Frequency |
|------|------|-----------|
| Dependency vulnerability scan | `govulncheck`, `trivy` | CI (every PR) |
| Static analysis | `gosec`, `staticcheck` | CI (every PR) |
| Secret detection | `gitleaks` | CI (every PR) |
| Penetration testing | Manual (security engineer) | Pre-launch |
| Rate limit testing | Custom load test script | Pre-launch |
| RLS policy verification | SQL test suite | CI (every migration) |

### 12.4 Load Testing Targets

| Scenario | Target | Tool |
|----------|--------|------|
| Concurrent logins/sec | 10,000 | `k6` or `vegeta` |
| Org switch P95 latency | < 200ms | `k6` |
| Password reset throughput | 1,000/min | `k6` |
| JWT generation | 50,000/sec | Go benchmark |
| Permission resolution | < 5ms (cached) | Go benchmark |

---

## 13. Observability

### 13.1 Structured Logging

```go
// Every log includes: correlation_id, service, version
logger.Info("login_attempt",
    zap.String("correlation_id", reqID),
    zap.String("user_email", security.MaskPII(email)),
    zap.String("ip", ip),
    zap.String("org_id", orgID.String()),
    zap.Bool("mfa_required", mfaRequired),
    zap.Bool("success", success),
)

// Security events at WARN level
logger.Warn("account_locked",
    zap.String("user_id", userID.String()),
    zap.Int("failed_attempts", attempts),
    zap.Time("locked_until", lockedUntil),
)
```

### 13.2 Metrics (Prometheus)

```go
var (
    loginAttempts = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "auth_login_attempts_total",
        Help: "Total login attempts",
    }, []string{"status", "org_id"}) // status: success, fail_mfa, fail_password, fail_locked

    loginDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "auth_login_duration_seconds",
        Help:    "Login request duration",
        Buckets: prometheus.DefBuckets,
    }, []string{"org_id"})

    passwordResets = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "auth_password_resets_total",
        Help: "Total password resets",
    }, []string{"method"}) // method: reset_link, change_password

    activeSessions = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "auth_active_sessions",
        Help: "Current active sessions",
    }, []string{"org_id"})

    mfaVerifications = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "auth_mfa_verifications_total",
        Help: "MFA verification attempts",
    }, []string{"method", "status"}) // method: totp, sms; status: success, fail

    orgSwitches = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "auth_org_switches_total",
        Help: "Organization context switches",
    }, []string{"org_id"})

    rateLimitHits = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "auth_rate_limit_hits_total",
        Help: "Rate limit exceeded count",
    }, []string{"endpoint", "key"})
)
```

### 13.3 Distributed Tracing

- **Correlation ID**: Generated at BFF edge, propagated via `X-Correlation-ID` header
- **Span propagation**: Auth service creates child spans for DB, Redis, NATS calls
- **Key spans**: `auth.identify`, `auth.login`, `auth.mfa.verify`, `auth.org.switch`, `auth.password.reset`

### 13.4 NATS Events

| Subject | Event | Payload |
|---------|-------|---------|
| `auth.login.success` | Successful login | `{user_id, org_id, ip, device, timestamp}` |
| `auth.login.failed` | Failed login | `{email_hash, ip, reason, attempts_remaining}` |
| `auth.account.locked` | Account lockout | `{user_id, email, ip, failed_attempts}` |
| `auth.password.reset` | Password reset | `{user_id, method, timestamp}` |
| `auth.session.revoked` | Session revoked | `{user_id, session_id, reason, by_user_id}` |
| `auth.org.switch` | Org context change | `{user_id, from_org_id, to_org_id, ip}` |
| `auth.user.invited` | User invited | `{email, org_id, invited_by, role_ids}` |
| `auth.user.created` | User created | `{user_id, org_id, created_by, method}` |
| `auth.mfa.enabled` | MFA enabled | `{user_id, method}` |
| `auth.mfa.failed` | MFA failure | `{user_id, method, ip, attempts_remaining}` |

---

## 14. Deployment & Rollout

### 14.1 Docker Configuration

```dockerfile
# services/auth-service/Dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o auth-service ./cmd

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/auth-service /auth-service
COPY --from=builder /app/config.toml /etc/flyfleet/config.toml
USER 1000
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1
ENTRYPOINT ["/auth-service"]
CMD ["-config", "/etc/flyfleet/config.toml"]
```

### 14.2 Docker Compose Addition

```yaml
# Add to docker-compose.yml
  auth-service:
    build:
      context: ./services/auth-service
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://fleet_user:${POSTGRES_PASSWORD}@postgres:5432/fleet_db
      - REDIS_URL=redis://redis:6379/0
      - NATS_URL=nats://nats:4222
      - JWT_SECRET=${JWT_SECRET}
      - SENDGRID_API_KEY=${SENDGRID_API_KEY}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      nats:
        condition: service_healthy
    restart: unless-stopped
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '1.0'
```

### 14.3 Phased Rollout

| Phase | Timeline | Scope | Success Criteria | Rollback Trigger |
|-------|----------|-------|-----------------|-----------------|
| **Alpha** | Week 4 | Internal devs only | All E2E tests pass | Any critical bug |
| **Closed Beta** | Week 10 | 50 opt-in customers | Zero P0 bugs, login success > 98% | Error rate > 2% |
| **Canary (10%)** | Week 11 | 10% of traffic via BFF routing | P95 login < 500ms, NPS > 40 | P95 > 1s for 10m |
| **Canary (50%)** | Week 12 | 50% of traffic | Reset completion > 90% | Any auth outage |
| **GA (100%)** | Week 12 | All users | All KPIs met | — |

### 14.4 Rollback Plan

```bash
# BFF routing control — switch auth routes back to legacy/disabled
# If auth-service fails, BFF returns 503 with maintenance page

# Feature flags (in config or Redis):
#   feature.login_v2 = false  → use old login
#   feature.org_switch = false → hide org switch UI
#   feature.password_reset = false → hide forgot password link

# Database: migrations are forward-only compatible (no DROP)
# No rollback needed for DB layer
```

---

## 15. Migration Plan

### 15.1 Migration Execution

```bash
# 1. Run new migrations (safe, all CREATE IF NOT EXISTS)
cd services/auth-service
go run cmd/main.go -migrate

# Or use existing migration tool:
./scripts/migrate.sh --up --path services/auth-service/migrations

# 2. Verify schema
psql -d fleet_db -c "\dt roles"
psql -d fleet_db -c "\dt permissions"
psql -d fleet_db -c "\dt role_permissions"
psql -d fleet_db -c "\dt user_roles"
psql -d fleet_db -c "\dt invitations"
psql -d fleet_db -c "\dt password_history"

# 3. Seed permissions
psql -d fleet_db -f services/auth-service/migrations/seed_permissions.sql

# 4. Verify RLS
psql -d fleet_db -c "SELECT tablename, policyname FROM pg_policies WHERE schemaname = 'public';"
```

### 15.2 Data Migration

**Existing data**: The `users` table already has `role` as a VARCHAR string. Migration to the new RBAC system:

```sql
-- Migration script: convert legacy string roles to role-based system
-- Run once during cutover

-- 1. Create system roles for each existing org
INSERT INTO roles (org_id, name, is_system, permission_mask)
SELECT DISTINCT
    uo.org_id,
    uo.role,  -- existing role string: 'super_admin', 'org_admin', etc.
    true,
    CASE
        WHEN uo.role = 'super_admin' THEN 9223372036854775807::bigint  -- all bits
        WHEN uo.role = 'org_admin' THEN -- specific bits
            (1::bigint << 0) | (1::bigint << 1) | (1::bigint << 2) | -- vehicles
            (1::bigint << 5) | (1::bigint << 6) | (1::bigint << 7) | -- drivers
            -- ... all admin bits
            0
        ELSE 0
    END
FROM user_organizations uo
LEFT JOIN roles r ON r.org_id = uo.org_id AND r.name = uo.role
WHERE r.id IS NULL;

-- 2. Create user_roles entries from user_organizations
INSERT INTO user_roles (user_id, role_id, org_id, assigned_at)
SELECT
    uo.user_id,
    r.id,
    uo.org_id,
    uo.created_at
FROM user_organizations uo
JOIN roles r ON r.org_id = uo.org_id AND r.name = uo.role
LEFT JOIN user_roles ur ON ur.user_id = uo.user_id AND ur.role_id = r.id AND ur.org_id = uo.org_id
WHERE ur.id IS NULL;
```

### 15.3 Cutover Strategy

```
Before:                    After:
  Frontend ──→ BFF          Frontend ──→ BFF ──→ auth-service (NEW)
              │                            │
              └──→ legacy (disabled)       └──→ settings-service (profile)
                                             └──→ fleet-core-service (data)
```

1. Deploy auth-service alongside existing services (no traffic)
2. Run migrations (additive only, no breaking changes)
3. Enable BFF routing to auth-service for alpha users (header-based)
4. Monitor for 48 hours
5. Enable for all users (feature flag flip)
6. After 2 weeks, remove legacy auth code

---

## 16. Open Decisions

### 16.1 Decisions to Make Before Build

| # | Decision | Options | Recommendation | Owner | Deadline |
|---|----------|---------|----------------|-------|----------|
| D1 | Email provider | SendGrid vs AWS SES vs SMTP relay | **SendGrid** (faster setup, better deliverability) | DevOps | Week 1 |
| D2 | JWT algorithm | HS256 (current) vs RS256 | **RS256** (asymmetric, supports key rotation, required for future SSO) | Backend Lead | Week 1 |
| D3 | Password history depth | 3 vs 5 vs configurable per org | **5 fixed** (SOC 2 recommendation) | Security | Week 1 |
| D4 | BFF routing or direct? | Route through BFF vs frontend calls auth-service directly | **Through BFF** (single origin, CORS simplicity, centralized rate limiting) | Architecture | Week 1 |
| D5 | SMS MFA provider | Twilio vs SNS vs disabled for MVP | **Disabled for MVP** (Phase 2: Twilio) | PM | Week 2 |
| D6 | SCIM timeline | In-scope vs Q3 2026 | **Q3 2026** (enterprise deals can wait) | PM | Week 2 |

### 16.2 Technical Risks

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|-----------|
| Bcrypt cost 12 too slow on ARM (Termux dev) | Medium | Low | Use cost 10 in dev, 12 in prod (config-driven) |
| Redis dependency adds operational complexity | Low | Medium | Fallback to DB queries if Redis unavailable (slower but functional) |
| RLS policy mistakes cause data leaks | Low | Critical | Automated RLS tests in CI, security review before beta |
| Email deliverability (spam filtering) | Medium | High | Use verified domain, SPF/DKIM setup, monitor bounce rate |
| Migration from string roles to RBAC tables | Low | High | Thorough migration script testing, dry run on staging first |
| NATS event ordering | Low | Low | Use JetStream with consumer groups, idempotent consumers |

### 16.3 Out of Scope (Future Phases)

| Feature | Phase | Estimated Effort |
|---------|-------|-----------------|
| SAML/OIDC SSO | Phase 2 (Q2 2026) | 3 weeks |
| SCIM 2.0 provisioning | Phase 3 (Q3 2026) | 4 weeks |
| Biometric auth (mobile) | Phase 4 (Q4 2026) | 2 weeks |
| Adaptive MFA (risk-based) | Phase 3 | 2 weeks |
| Bulk role assignment | Phase 2 | 1 week |
| Permission audit exports | Phase 2 | 1 week |
| Account lockout self-recovery via MFA | Phase 2 | 1 week |

---

## Appendix A: JWT Claim Schema

```json
{
  "sub": "user-uuid",
  "user_id": "user-uuid",
  "org_id": "org-uuid",
  "role": "org_admin",
  "permissions_mask": "9223372036854775807",
  "type": "access",
  "exp": 1712239200,
  "iat": 1712238300,
  "jti": "token-uuid",
  "iss": "flyfleet-auth-service",
  "aud": "flyfleet-api",
  "device_fingerprint": "sha256-abc123"
}
```

## Appendix B: Redis Key Schema

```
# Sessions
session:{jti}                    → Session metadata (TTL: 15m)
refresh:{sha256(token)}          → Refresh token record (TTL: 7d)
refresh_used:{sha256(token)}     → Rotation marker (TTL: 7d)

# Rate Limiting
rate:login:{ip}                  → Login attempt count (TTL: 15m)
rate:identify:{ip}               → Identify attempt count (TTL: 15m)
rate:reset:email:{email}         → Reset attempt count (TTL: 15m)
rate:reset:ip:{ip}               → Reset attempt count (TTL: 15m)

# Security
lockout:{user_id}                → Lockout state (TTL: 15m)
mfa:{temp_token_jti}             → Temp token state (TTL: 5m)
mfa:remember:{user_id}:{device}  → Remember device (TTL: 30d)
action_token_used:{jti}          → Single-use marker (TTL: 30m)

# Caching
user:{user_id}                   → User profile cache (TTL: 5m)
user_orgs:{user_id}              → User's orgs cache (TTL: 10m)
user_permissions:{user_id}:{org_id} → Effective permissions (TTL: 5m)
role:{role_id}                   → Role definition cache (TTL: 15m)
```

## Appendix C: NATS Event Schema

```json
// auth.login.success
{
  "event_type": "auth.login.success",
  "timestamp": "2026-04-04T10:00:00Z",
  "correlation_id": "req-abc123",
  "data": {
    "user_id": "usr-001",
    "org_id": "org-001",
    "ip": "203.0.113.42",
    "user_agent": "Mozilla/5.0...",
    "device": "Chrome/Windows",
    "mfa_method": "totp"
  }
}

// auth.account.locked
{
  "event_type": "auth.account.locked",
  "timestamp": "2026-04-04T10:00:00Z",
  "correlation_id": "req-abc123",
  "data": {
    "user_id": "usr-001",
    "email": "r***@fleetcorp.com",
    "ip": "203.0.113.42",
    "failed_attempts": 5,
    "locked_until": "2026-04-04T10:15:00Z"
  }
}
```

## Appendix D: Existing Code Reuse Plan

| Existing File | Reuse | Modification Needed |
|---------------|-------|-------------------|
| `packages/common/go/pkg/security/jwt.go` | ✅ Use as-is | Add RS256 support (decision D2) |
| `packages/common/go/pkg/security/rbac_policies.go` | ✅ Use as-is | Extend with new resources from product spec |
| `packages/common/go/pkg/security/tenant.go` | ✅ Use as-is | None |
| `packages/common/go/pkg/security/security.go` | ✅ Use as-is | Audit logger for email events |
| `services/_archive/auth-service-legacy/internal/handlers/auth_handlers.go` | ⚠️ Reference only | Patterns for input validation, error handling |
| `services/_archive/auth-service-legacy/internal/middleware/auth.go` | ⚠️ Reference only | Update for new JWT claim structure |
| `services/_archive/auth-service-legacy/internal/service/auth_service.go` | ⚠️ Reference only | Rewrite for new architecture (Redis, NATS) |
| `apps/dashboard-web/src/pages/LoginPage.tsx` | ✅ Use as-is | Minor API path updates |
| `apps/dashboard-web/src/pages/admin/rbac/RoleMatrix.tsx` | ✅ Use as-is | Connect to API |
| `apps/dashboard-web/src/pages/admin/rbac/Roles.tsx` | ✅ Use as-is | Connect to API |
| `apps/dashboard-web/src/pages/admin/Users.tsx` | ✅ Use as-is | Connect to API |

---

*Version 1.0 · 2026-04-04 · Next Review: 2026-04-11*
*Based on Product Spec: product-spec/IAM-USER-MANAGEMENT.md v1.0*
*Author: senior-architect · Status: Draft — Ready for Review*
