# FlyFleet - Enterprise Fleet Management Platform

## Project Overview

FlyFleet is a comprehensive, enterprise-grade fleet management platform built as a **modular monolith** architecture. It provides end-to-end fleet management capabilities with advanced security features, role-based access control (RBAC), and integration with Indian government systems (FASTag, e-way bills).

### Architecture

**Backend**: Go (Golang) modular monolith — single deployable with clear internal domain boundaries
- **fleet-core** (Port 8084) — **The Monolith** — single binary containing all business domains
  - 27 internal domains: auth, tracking, geofencing, vehicle, driver, trip, billing, maintenance, alerts, analytics, governance, reports, settings, integrations, india, compliance, finance, emergency, admin_support, diagnostics, monitoring, route, fuel, customer, video, platform_compat, optimization
  - Each domain follows hexagonal architecture: handlers → service → repository → models
  - Shared infrastructure: middleware (JWKS auth, RBAC, tenant context, rate limit, CORS), platform observability, event bus
- **auth-service** (Port 8080) — Standalone auth service (being consolidated into fleet-core as `domains/iam/`)
  - 74 endpoints: login, refresh tokens, RBAC, MFA, sessions, user CRUD, invitations, password reset
  - 7 handlers, 8 services, 6 repositories, email providers (SendGrid/SMTP/NoOp)
- **web-bff** (Port 8081) — Backend-for-Frontend aggregator (React dashboard ↔ fleet-core)

**Archived Services** (in `services/_archive/` — no longer deployed):
- tracking-service, vehicle-service, driver-service, trip-service, billing-service, maintenance-service
- alert-service, analytics-service, reports-service, governance-service, settings-service
- integrations-service, india-integration-service, admin-support-service, fleet-ml-service, ingestion-service, optimization-service, worker-service

**Frontend**: React + TypeScript + Tailwind CSS
- **dashboard-web** — Main admin dashboard
- **landing-page** — Marketing site
- **mobile** — Mobile applications

**Infrastructure**:
- **Database**: PostgreSQL with TimescaleDB (time-series data)
- **Cache**: Redis
- **Message Queue**: NATS JetStream
- **Containerization**: Docker & Docker Compose

### Modular Monolith Structure

```
services/fleet-core/                    ← Single deployable
├── cmd/main.go                         ← Single entry point, Chi router
├── internal/
│   ├── domains/                        ← 27 business domains
│   │   ├── auth/                       ← Legacy auth v2 (being deprecated)
│   │   ├── tracking/                   ← GPS telemetry, live location
│   │   ├── geofencing/                 ← Spatial zones, entry/exit events
│   │   ├── vehicle/                    ← Vehicle lifecycle
│   │   ├── driver/                     ← Driver management
│   │   ├── billing/                    ← Invoicing, payments, PDF
│   │   ├── maintenance/                ← Preventive/corrective scheduling
│   │   ├── trip/                       ← Trip planning, state machine
│   │   └── ...                         ← (19 more domains)
│   ├── middleware/                     ← Shared: JWKS auth, RBAC, tenant, rate limit
│   ├── platform/                       ← Observability, EventBus, metrics
│   └── pkg/                            ← Shared utilities
├── migrations/                         ← Database migrations
├── go.mod / go.sum
└── Dockerfile                          ← Single container
```

### Key Features

1. **Security & Compliance**
   - Multi-Factor Authentication (MFA) with TOTP/SMS
   - Role-Based Access Control (RBAC) with granular permissions
   - IP-based access control and threat detection
   - Comprehensive audit trails
   - Session management with concurrent limits

2. **Fleet Operations**
   - Vehicle lifecycle management
   - Driver management with behavior scoring
   - Trip planning, routing, and optimization
   - Maintenance scheduling (preventive/corrective)
   - Fuel consumption tracking
   - Document management (licenses, insurance)
   - Geofencing with entry/exit alerts

3. **Real-time Capabilities**
   - GPS telemetry streaming
   - Live vehicle tracking
   - Driver behavior analytics
   - Route optimization
   - Predictive maintenance

4. **Indian Market Integration**
   - FASTag electronic toll collection
   - E-way bill generation
   - VAHAN/SARATHI integration
   - GST compliance

## Building and Running

### Prerequisites

- Go 1.25.7+
- Node.js (for frontend)
- Docker & Docker Compose
- PostgreSQL 15+ (or use Docker)
- Redis 7+ (or use Docker)
- NATS 2.10+ (or use Docker)

### Quick Start with Docker

```bash
# Copy environment template
cp .env.example .env

# Edit .env with your credentials
# Set POSTGRES_PASSWORD and JWT_SECRET

# Start all services (fleet-core monolith + infra)
docker-compose up -d

# Check service health
docker-compose ps

# View logs
docker-compose logs -f fleet-core
```

### Manual Development Setup

```bash
# Build the monolith
cd services/fleet-core && go build ./...

# Build auth-service (standalone, being consolidated)
cd services/auth-service && go build ./...

# Run database migrations
./migrate.sh

# Start fleet-core monolith
cd services/fleet-core
go run cmd/main.go

# Start auth-service (separate process)
cd services/auth-service
go run cmd/main.go
```

### NPM Scripts (Quality Gates)

```bash
# Build BFF
npm run bff:build

# Run contract checks
npm run bff:contract-check

# Run smoke tests
npm run bff:smoke

# Check frontend API integration
npm run frontend-api-check

# Generate API clients
npm run api:clients:generate

# Check API client freshness
npm run api:clients:check

# Check trace propagation
npm run trace:check

# Run all services foundation check
npm run services:all:check
```

### Makefile Targets

```bash
# Build BFF
make bff-build

# Contract checking
make bff-contract-check

# Smoke tests
make bff-smoke

# API client generation
make api-clients-generate

# Check client freshness
make api-clients-check

# Trace propagation check
make trace-propagation-check

# All services foundation check
make all-services-foundation-check

# Identity/auth foundation check
make identity-auth-foundation-check

# Observability foundation check
make observability-foundation-check

# Mobile sync simulation
make mobile-sync-simulate

# Firmware OTA validation
make firmware-ota-validate

# Device command simulation
make device-command-simulate

# Vehicle/device tests
make vehicle-device-tests

# Runtime integration gate
make runtime-integration-gate
```

### Testing

```bash
# Run fleet-core tests (all domains)
cd services/fleet-core && go test ./...

# Run auth-service tests
cd services/auth-service && go test ./...

# Run integration tests
./scripts/verify_all.sh

# Run comprehensive validation
./comprehensive_validation.sh
```

## Development Conventions

### Code Organization

**Modular Monolith Domain Structure**:
```
services/fleet-core/internal/domains/{domain-name}/
├── handlers/         # HTTP handlers (adapters)
├── service/          # Business logic
├── repository/       # Data access layer (implements interfaces)
├── models/           # Domain entities
└── *.go              # Domain-specific files (metrics, module.go, etc.)
```

**Auth-Service Structure** (being consolidated):
```
services/auth-service/
├── cmd/main.go           # Entry point (will be removed after consolidation)
├── internal/
│   ├── handlers/         # HTTP handlers (→ domains/iam/handlers/)
│   ├── service/          # Business logic (→ domains/iam/service/)
│   ├── repository/       # Data access (→ domains/iam/repository/)
│   ├── models/           # Domain models (→ domains/iam/models/)
│   ├── email/            # Email providers (→ domains/iam/email/)
│   ├── middleware/       # (subsumed by fleet-core middleware)
│   ├── redis/            # (subsumed by fleet-core shared pool)
│   └── nats/             # (subsumed by fleet-core EventBus)
├── migrations/           # DB migrations (merged into fleet-core/)
└── Dockerfile            # (will be removed after consolidation)
```

**Frontend Structure**:
```
apps/dashboard-web/
├── src/
│   ├── components/       # Reusable UI components
│   ├── pages/            # Page components
│   ├── services/         # API service layer
│   ├── store/            # Redux store
│   ├── context/          # React context providers
│   ├── hooks/            # Custom hooks
│   ├── types/            # TypeScript types
│   └── utils/            # Utility functions
├── public/
└── package.json
```

### Coding Standards

**Go**:
- Use `gofmt` for formatting
- Follow Go best practices (effective Go)
- Use `zap` for structured logging
- Implement proper error handling
- Use context for cancellation and timeouts
- Write unit tests with `testing` package

**TypeScript/React**:
- Use TypeScript strict mode
- Follow React hooks patterns
- Use Tailwind CSS for styling
- Implement proper error boundaries
- Write tests with Jest/React Testing Library

### API Design

- RESTful API design principles
- Versioned endpoints (`/api/v1/...`)
- JSON request/response format
- Proper HTTP status codes
- Consistent error response format
- OpenAPI/Swagger documentation

### Database Conventions

- Use UUIDs for primary keys
- Include `created_at` and `updated_at` timestamps
- Use soft deletes (`deleted_at` column)
- Implement proper indexing
- Use migrations for schema changes

### Security Practices

- JWT tokens for authentication
- MFA for sensitive operations
- RBAC for authorization
- Input validation on all endpoints
- SQL injection prevention (parameterized queries)
- Rate limiting on auth endpoints
- CORS configuration
- Security headers middleware

### Logging & Observability

- Structured logging with Zap (Go)
- Correlation IDs for request tracing
- Metrics collection (Prometheus)
- Health check endpoints
- Distributed tracing ready

## API Endpoints Summary

The system exposes approximately **182 API endpoints** across active services:

- **fleet-core**: ~104 endpoints (27 domains)
- **auth-service**: ~74 endpoints (IAM domain, being consolidated)
- **web-bff**: ~4 endpoints (passthrough/aggregator)

*Note: The original 535 endpoints included archived microservices now consolidated into fleet-core.*

## Configuration

### Environment Variables

Key environment variables (see `.env.example`):

```bash
# Database
POSTGRES_USER=fleet_user
POSTGRES_PASSWORD=<secure>
POSTGRES_DB=fleet_db

# Security
JWT_SECRET=<secure-32-char-min>

# Service Ports
AUTH_SERVICE_PORT=8080
FLEET_CORE_SERVICE_PORT=8084
# ... etc
```

### Config Files

- `config.toml` - Main application configuration
- `docker-compose.yml` - Development Docker setup
- `docker-compose.prod.yml` - Production Docker setup

## Project Structure

```
flyfleet/
├── apps/                          # Frontend applications
│   ├── dashboard-web/             # Admin dashboard (React)
│   ├── landing-page/              # Marketing site
│   └── mobile/                    # Mobile apps
├── services/                      # Backend
│   ├── fleet-core/                # ← THE MONOLITH (single deployable)
│   │   ├── internal/domains/      # 27 business domains
│   │   ├── internal/middleware/   # Shared: auth, RBAC, tenant
│   │   ├── internal/platform/     # Observability, EventBus
│   │   └── cmd/main.go            # Single entry point
│   ├── auth-service/              # ← Being consolidated into fleet-core/domains/iam/
│   ├── _archive/                  # Archived microservices (DO NOT MODIFY)
│   └── monolith/                  # Scaffold (empty)
├── packages/                      # Shared packages
│   ├── common/                    # Common Go packages
│   ├── config/                    # Configuration packages
│   └── ui-components/             # Shared React components
├── infrastructure/                # Infrastructure code
│   ├── ci-cd/                     # CI/CD pipelines
│   ├── database/                  # Database migrations
│   ├── docker/                    # Dockerfiles
│   ├── k8s/                       # Kubernetes manifests
│   ├── monitoring/                # Monitoring setup
│   └── terraform/                 # Terraform configs
├── docs/                          # Documentation
├── scripts/                       # Utility scripts
├── backups/                       # Database backups
├── product-spec/                  # Product specifications
└── firmware/                      # IoT device firmware
```

## Quality Gates

The project includes comprehensive quality gates:

1. **Contract Testing**: API contract validation
2. **Smoke Testing**: Service health checks
3. **Integration Testing**: End-to-end workflows
4. **Performance Testing**: API latency checks
5. **Security Testing**: RBAC and auth validation
6. **Code Quality**: Linting and formatting
7. **Documentation**: API and operations docs

## Contributing

1. Create a feature branch
2. Make changes following coding conventions
3. Write/update tests
4. Run quality gates (`npm run services:all:check`)
5. Submit a pull request

## Support & Documentation

- API Documentation: `docs/api/`, `api_spec.json`, `openapi.yaml`
- Architecture Docs: `docs/architecture/`
- Deployment Guide: `docs/deployment/`
- Security Docs: `docs/security/`
- Operations Guide: `docs/operations/`
- Product Specs: `product-spec/`

## Qwen Added Memories
- **FlyFleet is a MODULAR MONOLITH**, not microservices. Single deployable: `services/fleet-core/` with 27 internal domains.
- **Auth-service is being consolidated** into fleet-core as `domains/iam/`. All 74 auth endpoints are functional.
- **NEVER create new parallel domains** — always extend existing domain modules. E.g., tracking code lives in `internal/domains/tracking/`, not a new directory.
- Key existing tracking files: `domains/tracking/service/ingestion_service.go`, `telemetry_query_service.go`, `repository/live_cache_repository.go`, `repository/telemetry_repository.go`, `handlers/tracking_query_handler.go`, `ingest_handler.go`.
- **`_archive/` is READ-ONLY** — contains deprecated microservice code. Never modify.
