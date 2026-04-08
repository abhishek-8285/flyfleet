# FlyFleet - Comprehensive Analysis & Improvement Opportunities

**Date**: 2026-03-26
**Analyst**: Claude Code
**Purpose**: Detailed review of FlyFleet enterprise fleet management platform with prioritized improvement recommendations

---

## Executive Summary

FlyFleet is an **enterprise-grade fleet management platform** built as a modern monorepo containing:
- **12+ Go microservices** (currently consolidating into fleet-core)
- **3 frontend applications** (Next.js landing, React dashboard, mobile)
- **26 business domains** handled across services
- **53+ frontend pages** in admin dashboard

The platform demonstrates strong engineering practices but has opportunities for improvement in developer experience, testing, observability, and long-term scalability.

**Overall Assessment**: B+ (Production-ready but needs maturation in DX, testing automation, and gradual decoupling)

---

## Current Architecture Snapshot

### Technology Stack
```
Backend:    Go 1.26, Gin/Chi, PostgreSQL (TimescaleDB), Redis, NATS JetStream
Frontend:   Next.js 14+, React, TypeScript, TailwindCSS (likely)
Infra:      Docker Compose, Kubernetes (Helm), Terraform
Observ:     OpenTelemetry, Prometheus, Grafana
Storage:    MinIO (S3), TimescaleDB (time-series)
```

### Key Components

1. **fleet-core-service** (Primary API Gateway)
   - Location: `services/fleet-core/`
   - Contains 26 domain modules under `internal/domains/`
   - Acts as BFF + reverse proxy for legacy services
   - ~4000 lines of Go code across domains
   - Proxies to: auth, alert, analytics, billing, compliance, driver, maintenance, reports, tracking, trip, vehicle, etc.

2. **Frontend Applications**
   - `apps/landing-page/` - Customer-facing (Next.js)
   - `apps/dashboard-web/` - Admin dashboard (React, ~53 pages, 17 for auth_v2 alone)
   - `apps/mobile/` - Cross-platform mobile apps (shared code)

3. **Shared Libraries**
   - `packages/common/go` - Shared Go utilities (logging, errors, middleware)
   - `packages/config` - White-label branding configuration

4. **Data & Messaging**
   - PostgreSQL (TimescaleDB HA) - Primary OLTP + time-series
   - Redis - Caching + session storage
   - NATS JetStream - Event streaming + command delivery
   - MinIO - Object storage (invoices, firmware, logs)

5. **Protocol Buffers**
   - 13 service contracts in `proto/` (v1 APIs)
   - Generated code used across services

---

## Detailed Improvement Recommendations

### Category 1: Developer Experience (HIGH IMPACT)

#### 1.1 Unified Bootstrap Process
**Problem**: New developers must manually:
- Install Go, Node.js, Docker
- Run `docker-compose up` in one terminal
- Run `go build ./...` in another
- Run `npm install` in multiple workspaces
- Find connection strings and credentials

**Solution**: Create `scripts/bootstrap.sh` that:
```bash
#!/usr/bin/env bash
set -e

# 1. Check prerequisites
go version > /dev/null || { echo "Install Go 1.23+"; exit 1; }
node --version > /dev/null || { echo "Install Node.js 20+"; exit 1; }
docker --version > /dev/null || { echo "Install Docker"; exit 1; }

# 2. Start dependencies
echo "Starting databases..."
docker-compose up -d postgres redis nats
sleep 10

# 3. Wait for DB
until docker exec fleet-postgres pg_isready -U fleet_user -d fleet_db; do
  echo "Waiting for PostgreSQL..."
  sleep 2
done

# 4. Go workspace sync
echo "Syncing Go workspace..."
go work sync
go mod download

# 5. NPM workspaces
echo "Installing frontend dependencies..."
npm install --workspaces

# 6. Generate code (proto, API clients)
echo "Generating protobuf code..."
./scripts/generate-proto.sh
echo "Generating API clients..."
npm run api:clients:generate

# 7. Seed test data (optional)
read -p "Seed test data? (y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  go run ./scripts/seed_dev_data.go
fi

echo "✓ Setup complete!"
echo "  - Dashboard: http://localhost:3000"
echo "  - API: http://localhost:8080"
echo "  - Swagger: http://localhost:8080/swagger"
echo "  - NATS: localhost:4222"
```

**Priority**: CRITICAL (impedes onboarding)
**Effort**: 2 hours
**Owner**: DevOps / Platform team

#### 1.2 Hot Reload & Debug Support
**Current**: Services require manual rebuilds

**Add**:
- Go hot reload with [`air`](https://github.com/cosmtrek/air):
  ```bash
  # In services/fleet-core/
  air init  # creates .air.toml
  air      # watches and rebuilds
  ```
- Next.js Turbopack for dashboard-web (already in Next.js 14+)
- Debug configurations:
  ```json
  // .vscode/launch.json
  {
    "name": "Debug fleet-core",
    "type": "debugpy",
    "request": "launch",
    "mode": "debug",
    "program": "${workspaceRoot}/services/fleet-core/cmd/main.go",
    "env": {
      "GO_ENV": "development"
    }
  }
  ```

**Priority**: HIGH (speeds up dev iteration)
**Effort**: 1 hour per service
**Dependencies**: Add `air` to dev dependencies

#### 1.3 Root-Level Makefile Enforcement
**Current**: `Makefile` exists but needs expansion

**Add targets**:
```makefile
.PHONY: help
help:  ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

dev: bootstrap ## Start development environment
.PHONY: dev

test: ## Run all tests
	@echo "Running Go tests..."
	go test ./...
	@echo "Running frontend tests..."
	npm run test --workspaces
.PHONY: test

lint: ## Run linters
	golangci-lint run ./...
	npm run lint --workspaces
.PHONY: lint

ci: lint test build ## Run full CI pipeline locally
.PHONY: ci

build: ## Build all services and apps
	go build ./...
	npm run build --workspaces
.PHONY: build
```

**Priority**: HIGH (standardizes workflows)
**Effort**: 1 hour

---

### Category 2: Testing & Quality Gates (HIGH IMPACT)

#### 2.1 Root Cause: No Test Executon in package.json
**Current**: Root `package.json` has:
```json
"scripts": {
  "test": "echo \"Error: no test specified\" && exit 1"
}
```

**Fix**: Add actual test orchestration:
```json
{
  "scripts": {
    "test": "npm run test --workspaces --if-present",
    "test:unit": "go test ./... -short",
    "test:integration": "go test ./... -tags=integration",
    "test:e2e": "npm run test:e2e --workspace=apps/dashboard-web",
    "test:coverage": "go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out"
  }
}
```

**Priority**: CRITICAL (testing not enforced)
**Effort**: 15 minutes
**Issue**: Tests may exist but aren't being run by CI

#### 2.2 Contract Testing Between BFF and Microservices
**Problem**: fleet-core proxies to other services but no guarantee that OpenAPI spec matches implementation

**Solution**: Use [`hermetic`](https://github.com/invopop/hermetic) or `pact`:
- Generate OpenAPI from each microservice code (via `gin-swagger` or similar)
- In CI, compare with canonical OpenAPI (`openapi.yaml` in root)
- Fail build if mismatch detected
- Already mentioned in docs → needs implementation

**Implementation**:
```bash
# In CI pipeline
go run ./scripts/verify-openapi-contracts.go
# Should:
# 1. Generate OpenAPI from each service's code
# 2. Compare with docs/api/openapi.yaml
# 3. Report differences
```

**Priority**: HIGH (prevents contract drift)
**Effort**: 4-8 hours
**Owner**: API Governance team

#### 2.3 Integration Test Matrix
**Current**: Likely unit tests exist but integration coverage unclear

**Create** `scripts/run_integration_tests.sh`:
```bash
#!/usr/bin/env bash
set -e

# Start full stack
docker-compose -f docker-compose.yml -f docker-compose.test.yml up -d

# Wait for health
./scripts/wait_for_services.sh

# Run integration tests
go test ./tests/integration/... -v -timeout=10m

# Tear down
docker-compose -f docker-compose.yml -f docker-compose.test.yml down -v
```

Use `testcontainers-go` (already in go.mod) to spin up isolated test environments.

**Priority**: HIGH (catches runtime issues)
**Effort**: 2 days
**Owner**: QA / Platform team

#### 2.4 Mutation Testing
**Why**: Ensure tests actually catch bugs, not just exercise code

**Tool**: `go-mutesting` or `mutation-testing-go`

**Action**: Run on critical paths (auth, billing, device commands). Generate report showing:
- Killed mutations (tests caught the bug)
- Survived mutations (tests missed the bug) → need more tests

**Priority**: MEDIUM (advanced quality metric)
**Effort**: 1 day setup, then ongoing
**Owner**: Engineering Enablement

---

### Category 3: Backend Architecture (MEDIUM-HIGH IMPACT)

#### 3.1 Decouple Proxy Logic from fleet-core
**Current State**: fleet-core contains upstream proxy configuration:
```go
// Pseudo-code
if route.Migrated {
  serveLocalHandler(route)
} else {
  proxyToUpstream(route, upstreamURLs[route.Service])
}
```

**Problems**:
- One service handles routing for all domains → blast radius
- Scaling issues: CPU-bound routing logic competes with business logic
- Deployment coupling: changing proxy config requires fleet-core deploy

**Migration Path**:

**Phase 1** (Short-term):
- Extract proxy logic into separate `edge-proxy` service
- Use Envoy or Traefik as API gateway instead of application-level proxy
- fleet-core focuses on migrated domains only

**Phase 2** (Medium-term):
- Deploy each domain as independent microservice
- Use service mesh (Istio/Linkerd) for inter-service communication
- Remove upstream proxy variables from fleet-core environment

**Already Migrated** (from README):
```
settings, governance, integrations, reports, maintenance,
analytics, trip, vehicle, billing, admin, alert, auth,
compliance, tracking
```

**Remaining in fleet-core** (from internal/domains/):
```
admin_support, alerts (duplicate?), analytics, billing (duplicate?),
compliance, customer, diagnostics, driver, emergency, finance,
fuel, geofencing, india, maintenance, monitoring, platform_compat,
reports (duplicate?), route, settings (duplicate?), tracking (duplicate?),
trip (duplicate?), vehicle (duplicate?), video
```

**Action Item**: Audit which domains are actually still proxied vs. locally served.

**Priority**: MEDIUM (architectural cleanup, reduces complexity)
**Effort**: 2-4 weeks
**Owner**: Platform Architecture

#### 3.2 Improve Database Connection Pool Isolation
**Problem**: All domains share single DB connection pool → noisy neighbor problem

**Solution**:
- Create separate `sql.DB` instances for high-traffic domains (tracking, telemetry)
- Set per-domain connection limits in config:
  ```yaml
  database:
    max_open: 100
    max_idle: 20
    domain_limits:
      tracking: 40
      billing: 20
      auth: 10
  ```
- Use connection pool metrics to detect starvation

**Priority**: MEDIUM (prevents cascading failures)
**Effort**: 2 days
**Owner**: Backend team

#### 3.3 Domain Health Check Granularity
**Current**: Single `/health` endpoint for entire service

**Improve**:
```
GET /health/live  # Process alive
GET /health/ready # Dependencies ready (DB, Redis, NATS)
GET /health/ready?domain=billing  # Domain-specific readiness
GET /health/metrics # Prometheus metrics
```

Add dependency health tracking:
```go
type HealthStatus struct {
  Status      string            `json:"status"`
  Timestamp   time.Time         `json:"timestamp"`
  Duration    time.Duration     `json:"duration"`
  Checks      map[string]Check  `json:"checks,omitempty"`
  Version     string            `json:"version"`
}

type Check struct {
  Status  string    `json:"status"`  // pass, warn, fail
  Latency time.Duration `json:"latency"`
  Details  string   `json:"details,omitempty"`
}
```

**Priority**: MEDIUM (improves observability)
**Effort**: 1 day
**Owner**: Backend team

---

### Category 4: Frontend Architecture (MEDIUM-HIGH IMPACT)

#### 4.1 Generated API Client Currency
**Verify**:
```bash
# Check if generated code matches OpenAPI spec
cd apps/dashboard-web
npm run api:clients:check  # Should exist
```

If not implemented, add with `orval` or `openapi-generator`:
```bash
# orval.config.js
module.exports = {
  api: {
    input: {
      target: '../openapi.json',
    },
    output: {
      target: './src/generated/api/index.ts',
      style: 'typescript-axios',
    },
  },
};
```

Add to `package.json`:
```json
{
  "scripts": {
    "api:clients:generate": "orval",
    "api:clients:check": "npm run api:clients:generate -- --dry-run && git diff --exit-code src/generated/api/"
  }
}
```

**Enforce in CI**: PR fails if generated code is stale.

**Priority**: HIGH (prevents client-server drift)
**Effort**: 1 day
**Owner**: Frontend platform team

#### 4.2 Standardize Server State Management
**Current**: Likely using Context API + manual `fetch` in components

**Recommendation**: Adopt TanStack Query (React Query) for:
- Automatic caching & deduplication
- Background refetching
- Optimistic updates
- Retry logic
- Pagination helpers

**Migration Path**:
1. Install: `npm install @tanstack/react-query`
2. Wrap app in `QueryClientProvider`
3. Replace manual fetching:
   ```typescript
   // Before
   const [vehicles, setVehicles] = useState<Vehicle[]>([]);
   useEffect(() => {
     fetch('/api/v1/vehicles').then(r => r.json()).then(setVehicles);
   }, []);

   // After
   const { data: vehicles } = useQuery({
     queryKey: ['vehicles'],
     queryFn: fetchVehicles,
   });
   ```

**Priority**: HIGH (dramatically simplifies data layer)
**Effort**: 1-2 weeks for full migration
**Owner**: Frontend team

#### 4.3 Error Boundaries & Error Handling
**Add**:
- Root-level Error Boundary catching render errors
- Global error modal for API errors
- Retry button for failed requests
- Sentry or similar for error aggregation

**Example**:
```typescript
// src/components/ErrorBoundary.tsx
export class ErrorBoundary extends React.Component {
  state = { hasError: false };

  static getDerivedStateFromError() {
    return { hasError: true };
  }

  componentDidCatch(error, info) {
    // Log to monitoring service
    console.error(error, info);
  }

  render() {
    if (this.state.hasError) {
      return <ErrorFallback onRetry={() => this.setState({ hasError: false })} />;
    }
    return this.props.children;
  }
}
```

**Priority**: HIGH (improves UX on errors)
**Effort**: 2 days
**Owner**: Frontend team

#### 4.4 Component Library Acceleration
**Status**: `packages/ui-components` marked as WIP

**Accelerate with**:
- Storybook for component development:
  ```bash
  cd packages/ui-components
  npx storybook init
  ```
- Chromatic for visual regression testing
- Radix UI primitives for accessibility
- Tailwind CSS for styling

**Priority**: MEDIUM (consistency across apps)
**Effort**: 1 week to establish foundation
**Owner**: Design Systems team

---

### Category 5: Observability & Operations (HIGH IMPACT)

#### 5.1 Structured Logging Standardization
**Goal**: Every log line includes:
```json
{
  "timestamp": "2025-03-26T12:34:56Z",
  "level": "info",
  "service": "fleet-core",
  "version": "v1.2.3",
  "trace_id": "abc123",
  "org_id": "org_456",
  "user_id": "user_789",
  "message": "Vehicle updated",
  "details": {...}
}
```

**Implementation**:
1. Middleware to extract/generate trace ID:
   ```go
   func TraceIDMiddleware() gin.HandlerFunc {
     return func(c *gin.Context) {
       traceID := c.GetHeader("X-Trace-ID")
       if traceID == "" {
         traceID = uuid.New().String()
       }
       c.Set("trace_id", traceID)
       c.Header("X-Trace-ID", traceID) // propagate to response
       c.Next()
     }
   }
   ```
2. Zap logger with structured fields:
   ```go
   logger.With(
     zap.String("trace_id", traceID),
     zap.String("org_id", orgID),
     zap.String("user_id", userID),
   ).Info("message")
   ```
3. Loki/Elastic log aggregation rules

**Priority**: HIGH (essential for debugging)
**Effort**: 2 days (audit existing logging + add missing fields)
**Owner**: Observability team

#### 5.2 HTTP Metrics (RED Method)
**Track per endpoint**:
- **Rate** (requests/sec)
- **Errors** (4xx, 5xx counts)
- **Duration** (p50, p95, p99 histograms)

**Implementation** (Prometheus client already in go.mod):
```go
import "github.com/prometheus/client_golang/prometheus"

var (
  httpRequests = prometheus.NewCounterVec(
    prometheus.CounterOpts{
      Name: "http_requests_total",
      Help: "Total HTTP requests",
    },
    []string{"method", "path", "status", "service"},
  )

  httpDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
      Name:    "http_request_duration_seconds",
      Help:    "HTTP request latency",
      Buckets: prometheus.DefBuckets,
    },
    []string{"method", "path", "service"},
  )
)

// In middleware
httpRequests.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(resp.Status), serviceName).Inc()
```

**Priority**: HIGH (SLO tracking)
**Effort**: 1 day
**Owner**: Observability team

#### 5.3 Business Metrics
**Track**:
- `organizations_created_total`
- `vehicles_online_gauge`
- `active_sessions_gauge`
- `invoices_generated_total`
- `firmware_rollouts_total{status="success|failed"}`
- `commands_sent_total{type="lock|unlock|ping"}`

**Implementation**: Create `internal/metrics` package with counters/gauges/histograms. Increment at business logic boundaries.

**Priority**: MEDIUM (product insights)
**Effort**: 2 days
**Owner**: Product Engineering

#### 5.4 Distributed Tracing Completeness
**Current**: OpenTelemetry in go.mod → verify instrumentation

**Checklist**:
- [ ] HTTP inbound/outbound spans
- [ ] Database query spans (pgx)
- [ ] Redis spans (go-redis)
- [ ] NATS spans (nats.go)
- [ ] External API calls (Stripe, SendGrid, FASTag, etc.)
- [ ] WebSocket connections (track frame send/receive)
- [ ] gRPC calls (if any)

**Auto-instrumentation**: Use `go.opentelemetry.io/contrib/instrumentation` packages.

**Add to Jaeger/Tempo**:
- Service dependency graph
- Trace to logs correlation (trace_id in logs)
- Slow trace alerts (p95 > 1s)

**Priority**: HIGH (debugging distributed systems)
**Effort**: 1 week (audit + add missing instrumentation)
**Owner**: Observability team

#### 5.5 Health Check Separation
**Implement**:
```go
// liveness probe
app.GET("/healthz", func(c *gin.Context) {
  c.JSON(200, HealthStatus{Status: "pass"})
})

// readiness probe
app.GET("/ready", func(c *gin.Context) {
  checks := map[string]interface{}{
    "postgres": checkPostgres(),
    "redis":    checkRedis(),
    "nats":     checkNATS(),
  }
  // Return 503 if any critical check fails
  c.JSON(200, HealthStatus{Status: "pass", Checks: checks})
})

// individual domain readiness
app.GET("/ready/:domain", func(c *gin.Context) {
  domain := c.Param("domain")
  if !isDomainReady(domain) {
    c.JSON(503, HealthStatus{Status: "fail", Details: "domain not ready"})
    return
  }
  c.JSON(200, HealthStatus{Status: "pass"})
})
```

Update `docker-compose.yml`:
```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/ready"]
  interval: 30s
  timeout: 10s
  retries: 3
```

**Priority**: HIGH (Kubernetes readiness, zero-downtime deploys)
**Effort**: 1 day
**Owner**: DevOps team

---

### Category 6: Security Hardening (HIGH IMPACT)

#### 6.1 Secrets Management Audit
**Current**: Using `.env` files locally (good)
**Production Gap**: Need Vault/AWS Secrets Manager/ Azure Key Vault

**Action**:
1. Document secrets required (create `docs/operations/SECRETS_MANIFEST.md`)
2. Implement Vault integration:
   ```go
   secret, err := vault.Read("secret/data/flyfleet/postgres")
   ```
3. Remove fallback to env vars in production code
4. Add secret rotation procedures (JWT secret, API keys, third-party credentials)
5. Add secret scanning to CI:
   ```yaml
   # .github/workflows/secret-scan.yml
   - name: Scan for secrets
     uses: trufflesecurity/trufflehog@main
     with:
       path: ./
       base: main
   ```

**Priority**: CRITICAL (secrets leakage risk)
**Effort**: 3 days
**Owner**: Security team

#### 6.2 Rate Limiting Granularity
**Current**: Per-IP + per-user (from README)

**Enhance**:
```go
// Multi-level rate limiting
limiter := rate.NewLimiter(rate.Every(time.Second), 100) // global
userLimiter := perUserStore.Get(userID)                  // per user
orgLimiter := perOrgStore.Get(orgID)                    // per tenant
endpointLimiter := perEndpointStore.Get(route)          // per endpoint

// Apply based on sensitivity
if route.Sensitive {
  if !orgLimiter.Allow() { return 429 }
}
```

Use Redis for distributed rate limiting (already have Redis).

**Priority**: HIGH (prevent abuse)
**Effort**: 2 days
**Owner**: Backend team

#### 6.3 Audit Logging for Security Events
**Log** every:
- Login attempt (success/failure)
- Password change
- MFA enable/disable
- Role/permission changes
- Data export (especially PII)
- OTA firmware approvals
- Billing disputes/refunds
- User deletion
- API key creation/revocation

**Format** (immutable, structured):
```json
{
  "audit_event": "user.role_changed",
  "timestamp": "2025-03-26T12:34:56Z",
  "actor": {"type": "user", "id": "user_123"},
  "target": {"type": "user", "id": "user_456"},
  "changes": {
    "old_roles": ["viewer"],
    "new_roles": ["admin", "editor"]
  },
  "ip": "203.0.113.42",
  "user_agent": "...",
  "trace_id": "abc123"
}
```

Store in separate append-only table or audit log service.

**Priority**: HIGH (compliance, incident response)
**Effort**: 3 days
**Owner**: Security team

#### 6.4 API Security Headers
**Add** middleware:
```go
headers := map[string]string{
  "X-Content-Type-Options": "nosniff",
  "X-Frame-Options": "DENY",
  "X-XSS-Protection": "1; mode=block",
  "Referrer-Policy": "strict-origin-when-cross-origin",
  "Permissions-Policy": "geolocation=(), microphone=()",
  "Content-Security-Policy": "default-src 'self'", // refine as needed
}
```

For cookie-based sessions:
```go
cookie.Secure = true   // HTTPS only
cookie.HttpOnly = true // No JS access
cookie.SameSite = http.SameSiteStrictMode
```

**Priority**: HIGH (OWASP Top 10)
**Effort**: 1 day
**Owner**: Backend team

---

### Category 7: Data & Database (MEDIUM IMPACT)

#### 7.1 Migration Safety
**Current**: `migrations/` folder exists but process unclear

**Improve**:
1. Use `golang-migrate` with checksums:
   ```bash
   migrate -path migrations -database "$POSTGRES_URL" version
   ```
2. Add `down` migrations for rollback (current SQL files may only have `up`)
3. CI validation:
   ```yaml
   - name: Validate migrations
     run: |
       migrate -path migrations -database "$TEST_DATABASE_URL" validate
   ```
4. Dry-run in staging before production apply
5. Keep migration history in version control, never modify committed migrations

**Priority**: MEDIUM (prevents DB corruption)
**Effort**: 2 days (audit existing migrations + add down paths)
**Owner**: Database team

#### 7.2 TimescaleDB Hypertables & Retention
**Verify**: Time-series data (telemetry, events) uses hypertables.

**Action**:
```sql
-- Check existing hypertables
SELECT * FROM timescaledb_information.hypertables;

-- Create if missing
SELECT create_hypertable('telemetry_events', 'timestamp');

-- Add retention policy (keep 90 days)
SELECT add_retention_policy('telemetry_events', INTERVAL '90 days');
```

**Priority**: MEDIUM (performance + storage cost)
**Effort**: 1 day
**Owner**: Database team

#### 7.3 Slow Query Detection
**Already enabled**: `pg_stat_statements`

**Add automated alerting**:
```sql
-- Find queries > 1s avg
SELECT query, calls, total_time, mean_time, rows
FROM pg_stat_statements
WHERE mean_time > 1000
ORDER BY mean_time DESC
LIMIT 10;
```

Create nightly job that emails/slack if slow queries detected. Add missing indexes based on findings.

**Priority**: MEDIUM (performance)
**Effort**: 2 days
**Owner**: Database team

---

### Category 8: Infrastructure & DevOps (HIGH IMPACT)

#### 8.1 Kubernetes Production Readiness
**Current**: Helm charts exist in `infrastructure/helm/`

**Checklist**:
- [ ] HPA (Horizontal Pod Autoscaler) for all deployments
- [ ] PDB (PodDisruptionBudget) for HA
- [ ] Resource limits/requests defined
- [ ] Readiness/Liveness probes configured
- [ ] `values-prod.yaml`, `values-staging.yaml` separated
- [ ] Ingress with TLS termination
- [ ] ServiceMonitors for Prometheus Operator
- [ ] ConfigMaps/Secrets externalized (not baked in images)
- [ ] Init containers for DB migrations

**Example PDB**:
```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: fleet-core-pdb
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: fleet-core
```

**Priority**: HIGH (if deploying to K8s)
**Effort**: 3 days
**Owner**: DevOps team

#### 8.2 CI/CD Pipeline Completeness
**Current**: `.pre-commit-config.yaml` exists but CI workflow unclear

**Create** `.github/workflows/ci.yml`:
```yaml
name: CI

on:
  pull_request:
  push:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with: { go-version: '1.26' }
      - run: make lint

  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: timescale/timescaledb-ha:pg15
        env:
          POSTGRES_DB: test_db
          POSTGRES_USER: test_user
          POSTGRES_PASSWORD: test_pass
        ports: [5432:5432]
        options: >-
          --health-cmd "pg_isready -U test_user -d test_db"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      redis:
        image: redis:7-alpine
        ports: [6379:6379]
      nats:
        image: nats:2.10-alpine
        ports: [4222:4222]
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with: { go-version: '1.26' }
      - run: make test

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: |
        # Run gosec
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
        gosec ./...
      - uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          format: 'sarif'
          output: 'trivy-results.sarif'
      - uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: 'trivy-results.sarif'

  build:
    needs: [lint, test, security]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with: { go-version: '1.26' }
      - run: make build
      - run: docker build -t flyfleet/fleet-core:${GITHUB_SHA} services/fleet-core
      - run: docker build -t flyfleet/dashboard-web:${GITHUB_SHA} apps/dashboard-web
```

**Priority**: HIGH (prevents broken code from merging)
**Effort**: 2 days
**Owner**: DevOps team

#### 8.3 Backup & Restore Automation
**Current**: Manual or unclear backup strategy

**Create** `scripts/backup.sh`:
```bash
#!/usr/bin/env bash
set -e

BACKUP_DIR="/backups/$(date +%Y-%m-%d_%H-%M-%S)"
mkdir -p "$BACKUP_DIR"

# PostgreSQL backup
docker exec fleet-postgres pg_dump -U fleet_user fleet_db \
  | gzip > "$BACKUP_DIR/postgres.sql.gz"

# Redis dump
docker exec fleet-redis redis-cli --rdb /data/dump.rdb
docker cp fleet-redis:/data/dump.rdb "$BACKUP_DIR/redis.rdb"

# Upload to S3/MinIO (optional)
# aws s3 cp "$BACKUP_DIR" s3://flyfleet-backups/ --recursive

echo "Backup complete: $BACKUP_DIR"
```

**Create** `scripts/restore.sh`:
```bash
#!/usr/bin/env bash
BACKUP_PATH=$1
if [ -z "$BACKUP_PATH" ]; then
  echo "Usage: $0 <backup_directory>"
  exit 1
fi

# Stop services
docker-compose down

# Restore PostgreSQL
gunzip -c "$BACKUP_PATH/postgres.sql.gz" | docker exec -i fleet-postgres psql -U fleet_user fleet_db

# Restore Redis
docker cp "$BACKUP_PATH/redis.rdb" fleet-redis:/data/dump.rdb
docker restart fleet-redis

docker-compose up -d
```

Add to cron:
```
0 2 * * * /data/flyfleet/scripts/backup.sh >> /var/log/backup.log 2>&1
```

**Priority**: HIGH (disaster recovery)
**Effort**: 2 days
**Owner**: DevOps team

#### 8.4 Monitoring Dashboards Completion
**Verify** `infrastructure/monitoring/` contains:
- [ ] Service health dashboard (up/down, response times)
- [ ] Infrastructure dashboard (CPU, memory, disk, network)
- [ ] Database dashboard (connections, slow queries, replication lag)
- [ ] Redis dashboard (memory usage, hit rate, commands/sec)
- [ ] NATS dashboard (connections, messages, latency)
- [ ] Business metrics dashboard (vehicles, trips, invoices)

**Action**: Import Grafana dashboards from `infrastructure/monitoring/` to running instance. Add to `README` how to access.

**Priority**: MEDIUM (operational visibility)
**Effort**: 1 day
**Owner**: Observability team

---

### Category 9: Documentation (MEDIUM-HIGH IMPACT)

#### 9.1 Architecture Decision Records (ADRs)
**Missing**: No `docs/adr/` directory found

**Create** ADRs for key decisions:
- ADR-001: BFF Pattern Adoption
- ADR-002: Monolith-to-Microservices Migration Strategy
- ADR-003: TimescaleDB for Telemetry Storage
- ADR-004: NATS JetStream for Event Streaming
- ADR-005: OpenAPI as Single Source of Truth
- ADR-006: White-Label Branding Architecture

**Template** (from `docs/architecture/`):
```markdown
# ADR-XXX: Decision Title

**Status**: Accepted
**Date**: 2025-03-26
**Deciders**: Platform Team

## Context
Problem statement, constraints, assumptions.

## Decision
What was decided, with technical details.

## Consequences
Positive: What improved?
Negative: New problems/risks introduced.
```

**Priority**: HIGH (knowledge retention)
**Effort**: 2 days (write 5-10 ADRs)
**Owner**: Platform Architecture team

#### 9.2 Runbook Enhancement
**Check** `docs/runbooks/` → ensure covers:

**Critical incidents**:
- [ ] Database outage (connection pool exhaustion, replication failure, disk full)
- [ ] Redis outage (memory pressure, eviction)
- [ ] NATS outage (JetStream full, consumer lag)
- [ ] High error rate (>1% 5xx)
- [ ] Memory leak (OOM kills, GC pressure)
- [ ] Slow API responses (p95 > 2s)
- [ ] File storage full (MinIO)
- [ ] Firmware rollout failure

**Each runbook should include**:
```
1. Detection: How we know it's happening (alerts, metrics)
2. Impact: Which users/services affected
3. Diagnosis: Steps to identify root cause (queries, commands)
4. Mitigation: Immediate actions to restore service
5. Escalation: Who to call, when
6. Rollback: How to revert if recent deploy caused it
7. Post-mortem: Template for follow-up
```

**Priority**: HIGH (on-call effectiveness)
**Effort**: 3 days
**Owner**: SRE / Operations team

#### 9.3 API Documentation Synchronization
**Ensure**:
- `openapi.yaml` matches actual code (enforced by CI contract testing)
- Swagger UI accessible at `/swagger` (fleet-core has it)
- Generate client libraries automatically in CI
- Document breaking changes and deprecations

**Add** `docs/api/VERSIONING_POLICY.md`:
- Version in URL: `/api/v1/...`, `/api/v2/...`
- Support 2 versions concurrently, deprecate old after 6 months
- Backward-compatible changes only within version (add fields, new endpoints)
- Breaking changes require major version bump

**Priority**: MEDIUM (developer experience)
**Effort**: 1 day
**Owner**: API Governance team

---

### Category 10: Configuration & Environment (MEDIUM IMPACT)

#### 10.1 Configuration Validation at Startup
**Use Viper effectively**:

Create `internal/config` package:
```go
type Config struct {
  Server struct {
    Port         int    `mapstructure:"PORT" validate:"required,min=1024,max=65535"`
    Env          string `mapstructure:"ENV" validate:"required,oneof=development staging production"`
    Host         string `mapstructure:"HOST"`
  } `mapstructure:"SERVER"`

  Database struct {
    Host     string `mapstructure:"POSTGRES_HOST" validate:"required"`
    Port     int    `mapstructure:"POSTGRES_PORT" validate:"required"`
    User     string `mapstructure:"POSTGRES_USER" validate:"required"`
    Password string `mapstructure:"POSTGRES_PASSWORD" validate:"required"`
    Name     string `mapstructure:"POSTGRES_DB" validate:"required"`
    SSLMode  string `mapstructure:"POSTGRES_SSLMODE"`
  } `mapstructure:"DATABASE"`

  // ... Redis, NATS, etc.
}

func Load() (*Config, error) {
  viper.AutomaticEnv()
  var cfg Config
  if err := viper.Unmarshal(&cfg); err != nil {
    return nil, err
  }
  if err := validator.New().Validate(cfg); err != nil {
    return nil, fmt.Errorf("config validation failed: %w", err)
  }
  return &cfg, nil
}
```

**Benefits**: Fail fast on missing/invalid config.

**Priority**: MEDIUM (reuces runtime errors)
**Effort**: 2 days
**Owner**: Backend team

#### 10.2 Multi-Environment Configuration
**Structure**:
```
config/
├── base.yaml          # Common defaults
├── development.yaml   # Dev overrides
├── staging.yaml       # Staging overrides
└── production.yaml    # Production overrides
```

Load with:
```bash
# Development
ENV=development go run cmd/main.go

# Staging
ENV=staging go run cmd/main.go
```

**Priority**: MEDIUM (environment isolation)
**Effort**: 1 day
**Owner**: DevOps team

---

## Quick Wins (Effort < 4 hours)

| # | Improvement | Impact | Effort |
|---|-------------|--------|--------|
| 1 | Add `make help` target | Low | 15min |
| 2 | Fix root `package.json` test script | High | 15min |
| 3 | Add `docker-compose.override.yml` with hot-reload mounts | High | 1hr |
| 4 | Add trace ID middleware if missing | High | 1hr |
| 5 | Add `make lint` with staticcheck + eslint | Medium | 1hr |
| 6 | Add `scripts/seed_dev_data.sh` with sample org/vehicles | Medium | 2hrs |
| 7 | Add root `Makefile` targets: `test`, `build`, `lint` | High | 1hr |
| 8 | Verify health checks on all docker-compose services | High | 1hr |
| 9 | Add pre-commit hooks (if not active) | Medium | 1hr |
| 10 | Generate docs site with mkdocs/docusaurus | Low | 2hrs |
| 11 | Add `go vet` + `staticcheck` to CI | High | 30min |
| 12 | Add dependency vulnerability scan (Trivy) to CI | High | 1hr |
| 13 | Add `TESTCONTAINERS_RYUK_DISABLED=true` for CI tests | Medium | 15min |
| 14 | Document environment variables in `.env.example` thoroughly | Medium | 1hr |
| 15 | Add `scripts/verify_integration.sh` connectivity check | Medium | 2hrs |

---

## Technical Debt Items

1. **Remove hardcoded values** (see `HARDCODED_VALUES_FIX_PLAN.md`)
2. **Delete unused services** in `services/_archive/` or document why kept
3. **Standardize error envelope** across all domains (verify consistency)
4. **Upgrade dependencies**: `go get -u ./...`, `npm update` (check for vulnerabilities)
5. **Review large files** in fleet-core; split domains >500LOC
6. **Implement pagination** on all list endpoints (verify cursor-based)
7. **Add idempotency keys** for write operations (POST/PUT/DELETE)
8. **Remove duplicate domain packages** (`alerts` vs `alert`, `reports` duplicate?)
9. **Fix TODOs/FIXMEs** across codebase (search for them)
10. **Review and refactor** `internal/domains` structure (some may belong in separate services)

---

## Prioritization Matrix

| Priority | Category                              | Effort    | Impact   |
|----------|---------------------------------------|-----------|----------|
| P0       | Developer Experience (bootstrap)      | 2 hrs     | Critical |
| P0       | Testing (package.json test script)    | 15 min    | Critical |
| P0       | Security (secrets scanning)           | 1 day     | Critical |
| P0       | CI/CD (security scanning)             | 1 day     | High     |
| P1       | Observability (structured logging)   | 2 days    | High     |
| P1       | API Contract Testing                  | 1 day     | High     |
| P1       | Frontend API client generation        | 1 day     | High     |
| P1       | Health Check Separation               | 1 day     | High     |
| P1       | Infrastructure Backups               | 2 days    | High     |
| P2       | Database Connection Isolation        | 2 days    | Medium   |
| P2       | Decouple Proxy Logic                  | 2-4 wks   | Medium   |
| P2       | ADRs & Runbooks                      | 3 days    | Medium   |
| P2       | Business Metrics                     | 2 days    | Medium   |
| P3       | Component Library (Storybook)         | 1 week    | Medium   |
| P3       | Mutation Testing                     | 1 day     | Low      |

---

## Owner Assignment Suggestions

| Area                          | Suggested Owner        |
|-------------------------------|------------------------|
| Bootstrap / Dev Experience    | DevOps / Platform Team |
| Testing & Quality Gates       | QA / Engineering Enab |
| Backend Architecture          | Platform Architecture  |
| Frontend DX                   | Frontend Platform Team |
| Observability                 | SRE / Observability   |
| Security Hardening            | Security Team         |
| Database                      | Database Team / DBA   |
| Infrastructure / K8s          | DevOps Team           |
| Documentation                 | Tech Writers + Arch   |

---

## Metrics to Track Progress

- **Developer onboarding time** (target: <1 day)
- **Test coverage** (target: >80% statement, critical paths 100%)
- **Mean time to detect (MTTD)** incidents
- **Mean time to recover (MTTR)** incidents
- **P95 API latency** (target: <200ms)
- **Error rate** (target: <0.1%)
- **Security scan findings** (target: 0 high/critical)
- **Contract drift incidents** (target: 0)
- **Build time** (target: <15 min)

---

## Next Actions

1. **Share this document** with engineering leadership
2. **Prioritize P0 items** in next sprint planning
3. **Assign owners** for each category
4. **Create GitHub issues** for each actionable item
5. **Set up weekly** improvement triage meetings
6. **Track metrics** in dashboard (Grafana or spreadsheet)
7. **Revisit quarterly** to assess progress and adjust priorities

---

## Conclusion

FlyFleet is a **sophisticated, production-grade platform** with solid foundations. The recommended improvements focus on:

1. **Operational excellence** (backups, monitoring, runbooks)
2. **Developer velocity** (bootstrap, hot reload, testing)
3. **Quality gates** (contract testing, CI enforcement)
4. **Security hardening** (secrets, audit logging, headers)
5. **Architectural clarity** (decoupling proxy, domain isolation)
6. **Observability** (tracing, metrics, structured logs)

By addressing the **P0/P1 items** in the next 2-3 months, FlyFleet will significantly improve reliability, developer experience, and operational readiness for scale.

---

**Document Status**: Ready for review
**Last Updated**: 2026-03-26
**Maintained By**: Engineering Leadership
