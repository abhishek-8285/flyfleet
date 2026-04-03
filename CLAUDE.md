# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with the FlyFleet fleet management system.

## Project Overview

FlyFleet is a comprehensive fleet management platform built with Go (backend) and modern web technologies. The system handles:
- Driver management and onboarding
- Vehicle tracking and telemetry
- Emergency incident response
- Financial operations and billing
- India compliance (FASTag, E-Way Bills, GST)
- Real-time GPS/OBD data ingestion

**Tech Stack:** Go, Chi router, PostgreSQL, NATS JetStream, Redis

## Important Directories

```
/data/data/com.termux/files/home/flyfleet/
├── services/fleet-core/          # Main backend service
│   ├── internal/domains/        # Domain-driven architecture
│   │   ├── driver/              # Driver management
│   │   ├── vehicle/             # Vehicle management
│   │   ├── tracking/            # GPS/telemetry ingestion
│   │   ├── emergency/           # SOS and incident response
│   │   ├── billing/             # Invoicing and payments
│   │   ├── finance/             # Financial operations
│   │   ├── india/               # India-specific compliance
│   │   ├── compliance/          # Regulatory compliance
│   │   └── ...                  # Other domains
├── packages/                    # Shared packages
├── proto/                       # Protocol buffer definitions
├── infrastructure/              # Docker, k8s configs
└── _archive/                    # ⚠️ DEPRECATED - DO NOT MODIFY
```

## ⚠️ CRITICAL RULES

### 1. DO NOT MODIFY `_archive` FOLDER

**Rule:** Never make any changes to files in the `_archive` directory or any subdirectory named `_archive`.

**Rationale:** The `_archive` folder contains deprecated, legacy, or historical code that should not be modified. All active development should occur in the current service directories (`services/fleet-core/`, etc.).

**Enforcement:**
- If you encounter a file in `_archive/`, do not edit it
- Do not create, move, or delete files in `_archive/`
- If a task requires changes to archived code, first check if there's an active equivalent in the main codebase
- When in doubt, ask the user before touching anything in `_archive/`

### 2. Security-First Mindset

This codebase handles sensitive data (driver PII, financial records, location data). Always:
- Validate all inputs rigorously
- Use parameterized queries (already implemented)
- Never log sensitive information
- Follow the principle of least privilege

### 3. Follow Existing Patterns

The codebase uses:
- **DTO pattern** for API boundaries (never decode directly into DB models)
- **Domain-driven design** with clear separation of concerns
- **Repository pattern** for data access
- **Context-based operations** with timeouts
- **Structured logging** with zap

## Build & Test

```bash
# Navigate to service
cd /data/data/com.termux/files/home/flyfleet/services/fleet-core

# Build
go build ./...

# Test
go test ./...

# Format
go fmt ./...
```

## Recent Security Fixes

The following vulnerabilities have been fixed (do not reintroduce):

1. ✅ Mass Assignment - Use DTOs, not direct model decoding
2. ✅ OS Command Injection - Use `exec.Command` without shell
3. ✅ Server-Side Template Injection - Comprehensive HTML sanitization
4. ✅ DoS Memory Exhaustion - Request body size limits (`MaxBytesReader`)
5. ✅ Cross-Tenant Message Injection - Validate org ownership via DB lookup
6. ✅ Credential JSON Exposure - Sensitive fields use `json:"-"`
7. ✅ Error Information Leakage - Generic error messages to clients

## Development Commands

```bash
# Bootstrap new development environment
./scripts/bootstrap.sh              # Full setup with databases
./scripts/bootstrap.sh --skip-seed  # Skip test data

# Build and test
make build                          # Build all services and apps
make test                           # Run all tests
make test:coverage:gate             # Run tests with 70% coverage enforcement
make coverage                       # Alias for above
make lint                           # Run all linters
make ci                             # Run full CI pipeline locally
make help                           # Show all commands

# Production readiness checks
make all-services-foundation-check  # Verify all service foundations
make identity-auth-foundation-check # Auth security checks
make observability-foundation-check # Observability verification
```

## Architecture Guidelines

### Modular Monolith Structure

FlyFleet uses a **modular monolith** with Domain-Driven Design:

- **27 domains** in `services/fleet-core/internal/domains/`
- Each domain follows hexagonal architecture:
  ```
  handlers/     → HTTP adapters
  service/      → Business logic
  repository/   → Data access (implements interface in repository/analytics_repository.go)
  models/       → Domain entities
  ```
- **Interface segregation**: Each domain has its own repository interface
- **Cross-cutting concerns**: `internal/platform/` for shared middleware, metrics, observability

### Event-Driven Communication

Use NATS JetStream for loose coupling:
- Publish events: `eventBus.Publish(ctx, "subject.name", data, orgID)`
- Subscribe: `eventBus.Subscribe("subject.name", handler)`
- Always include `org_id` in event payload for multi-tenancy

### Multi-Tenancy Enforcement

**MANDATORY**: All database queries MUST filter by `org_id`:
```go
// In repository methods
query := "SELECT * FROM table WHERE org_id = $1 AND id = $2"
// NEVER omit org_id filter
```

### Code Quality Standards

- **File size**: Keep Go files < 1000 lines. If a file exceeds 1000 lines, it's a "god class" and must be refactored.
- **Test coverage**: Minimum 70% overall. Critical domains (auth, billing) should target 80%+.
- **Security**: Use parameterized queries exclusively. Never use string concatenation for SQL.
- **Logging**: Use structured logging with zap. Include `org_id` in all log entries.
- **Timeouts**: All database and external calls must use context timeouts (typical: 30s).

### DTO Pattern Enforcement

**Never expose database models directly via API**. Create separate DTOs for:
- Request payloads (input)
- Response payloads (output)
- API boundaries (handlers ↔ service layer)

Domain models (`models.go`) are for internal use only. DTOs live in separate files (e.g., `dto/login_request.go`).

## Infrastructure & Deployment

- **Development**: Use `docker-compose.yml` for local PostgreSQL, Redis, NATS
- **Production**: Kubernetes (Helm charts in `infrastructure/helm/`)
- **Observability**: OpenTelemetry + Prometheus + Grafana (already instrumented)
- **CI/CD**: GitHub Actions with security scanning, contract tests, coverage enforcement

## Security Scanning

All code changes must pass:
- `govulncheck` for Go vulnerabilities (automated in CI)
- `npm audit` for frontend vulnerabilities
- Container scanning with Trivy

Run locally before pushing:
```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
npm audit --workspaces
```

## Important Warnings

1. **Never modify `_archive/`** - Deprecated legacy code
2. **Never hardcode secrets** - Use environment variables
3. **Never bypass multi-tenancy** - org_id must be present in all data operations
4. **Never skip tests** - New code must have corresponding tests
5. **Avoid god classes** - If a file exceeds 1000 lines, split it before review

---

**Last Updated:** 2026-03-27
**Maintainer:** FlyFleet Architecture Team
**Source:** Based on comprehensive architecture review (ARCHITECTURE_REVIEW_SUMMARY.md)

## Billing System

The billing domain (`internal/domains/billing/`) is fully implemented and wired into `main.go`.

### Architecture
- **Handler**: `handlers/handlers.go` — `BillingHandler` with 11+ endpoints
- **Service**: `service/service.go` — `BillingService` + `InvoiceService`
- **Repository**: `repository/repository.go` — `BillingRepository` (main CRUD)
- **PDF Engine**: `pdf/pdf_generator.go` — chromedp HTML→PDF rendering
- **Routes**: Registered at `/api/v1/billing/*` in `cmd/main.go`

### Key Endpoints
| Endpoint | Method | Handler |
|----------|--------|---------|
| `/api/v1/billing/invoices` | GET, POST | ListInvoices, CreateInvoice |
| `/api/v1/billing/invoices/{id}` | GET, PUT, DELETE | GetInvoice, UpdateInvoice, DeleteInvoice |
| `/api/v1/billing/invoices/{id}/payments` | POST | RecordPayment |
| `/api/v1/billing/invoices/{id}/pdf` | GET | DownloadInvoicePDF |
| `/api/v1/billing/settings` | GET, PUT | GetSettings, UpdateSettings |
| `/api/v1/billing/metrics` | GET | GetMetrics |
| `/api/v1/billing/health` | GET | HealthCheck |

### Database Schema
The `invoices` table uses `custom_fields` (JSONB), NOT `metadata`. It does NOT have `created_by` or `updated_by` columns. Always verify column names with `\d invoices` before writing SQL.

### Termux Note
On Android/Termux, invoice storage uses `$PREFIX/tmp/flyfleet/invoices` instead of `/tmp/flyfleet/invoices`.

## Database

- Primary: PostgreSQL
- Connection pooling via `pgxpool`
- All queries should be parameterized
- Use context timeouts on all DB operations

## Messaging

- NATS JetStream for telemetry persistence
- Redis for caching and session management

## Environment Variables

Key configuration:
- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - JWT signing key (required)
- `REDIS_URL` - Redis connection (optional but recommended)
- `FINANCE_BACKUP_DIR` - Backup storage location

## Notes

- Telemetry endpoints are high-throughput (1000+ req/sec). Avoid blocking operations.
- Emergency services (SOS, rollover, fire) require immediate processing.
- India compliance features have strict regulatory requirements.

---

**Last Updated:** 2026-03-26
**Maintainer:** FlyFleet Security Team
