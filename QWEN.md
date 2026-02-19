# FlyFleet - Enterprise Fleet Management Platform

## Project Overview

FlyFleet is a comprehensive, enterprise-grade fleet management platform built as a microservices architecture. It provides end-to-end fleet management capabilities with advanced security features, role-based access control (RBAC), and integration with Indian government systems (FASTag, e-way bills).

### Architecture

**Backend**: Go (Golang) microservices (16 active services)
- **auth-service** (Port 8080) - Authentication, MFA, RBAC, security
- **fleet-core-service** (Port 8084) - Core fleet operations
- **tracking-service** (Port 8082) - GPS tracking, telemetry
- **web-bff** (Port 8081) - Backend-for-frontend aggregator
- **alert-service** (Port 8085) - Alerts and notifications
- **billing-service** (Port 8087) - Billing and payments
- **reports-service** (Port 8083) - Reporting and analytics
- **governance-service** (Port 8088) - Governance and compliance
- **analytics-service** (Port 8089) - Business intelligence
- **integrations-service** (Port 8090) - Third-party integrations
- **settings-service** (Port 8091) - User and org settings
- **maintenance-service** (Port 8092) - Maintenance scheduling
- **vehicle-service** (Port 8093) - Vehicle management
- **driver-service** (Port 8094) - Driver management
- **trip-service** (Port 8095) - Trip planning and tracking
- **india-integration-service** - Indian government integrations

**Frontend**: React + TypeScript + Tailwind CSS
- **dashboard-web** - Main admin dashboard
- **landing-page** - Marketing site
- **mobile** - Mobile applications

**Infrastructure**:
- **Database**: PostgreSQL with TimescaleDB (time-series data)
- **Cache**: Redis
- **Message Queue**: NATS JetStream
- **Containerization**: Docker & Docker Compose

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

# Start all services
docker-compose up -d

# Check service health
docker-compose ps

# View logs
docker-compose logs -f auth-service
```

### Manual Development Setup

```bash
# Build all Go services
cd services/auth-service && go build ./...
cd services/fleet-core-service && go build ./...
# ... repeat for other services

# Run database migrations
./migrate.sh

# Start individual service
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
# Run service tests
cd services/auth-service && go test ./...

# Run integration tests
./scripts/verify_all.sh

# Run comprehensive validation
./comprehensive_validation.sh
```

## Development Conventions

### Code Organization

**Go Services Structure**:
```
services/{service-name}/
├── cmd/
│   └── main.go           # Application entry point
├── internal/
│   ├── config/           # Configuration loading
│   ├── database/         # Database connection
│   ├── handlers/         # HTTP handlers
│   ├── middleware/       # HTTP middleware
│   ├── models/           # Data models
│   ├── repository/       # Data access layer
│   └── service/          # Business logic
├── migrations/           # Database migrations
├── Dockerfile
├── go.mod
└── config.toml
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

The system exposes approximately **535 API endpoints** across all services:

- **auth-service**: 74 endpoints
- **fleet-core-service**: 104 endpoints
- **web-bff**: 181 endpoints
- **tracking-service**: 9 endpoints
- **alert-service**: 8 endpoints
- **billing-service**: 13 endpoints
- **reports-service**: 11 endpoints
- **governance-service**: 10 endpoints
- **analytics-service**: 43 endpoints
- **integrations-service**: 4 endpoints
- **settings-service**: 12 endpoints
- **maintenance-service**: 18 endpoints
- **vehicle-service**: 25 endpoints
- **driver-service**: 19 endpoints
- **trip-service**: 17 endpoints

See `api_spec.json` and `openapi.yaml` for detailed API documentation.

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
├── apps/                    # Frontend applications
│   ├── dashboard-web/      # Admin dashboard
│   ├── landing-page/       # Marketing site
│   └── mobile/             # Mobile apps
├── services/               # Go microservices
│   ├── auth-service/
│   ├── fleet-core-service/
│   └── ...
├── packages/               # Shared packages
│   ├── common/            # Common Go packages
│   ├── config/            # Configuration packages
│   └── ui-components/     # Shared React components
├── infrastructure/         # Infrastructure code
│   ├── ci-cd/             # CI/CD pipelines
│   ├── database/          # Database migrations
│   ├── docker/            # Dockerfiles
│   ├── k8s/               # Kubernetes manifests
│   ├── monitoring/        # Monitoring setup
│   └── terraform/         # Terraform configs
├── docs/                   # Documentation
├── scripts/                # Utility scripts
├── backups/                # Database backups
└── firmware/               # IoT device firmware
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
