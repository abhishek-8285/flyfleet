# Maintenance Service

The Maintenance Service is a Go-based microservice for FlyFleet that manages vehicle maintenance schedules, work orders, and parts tracking.

## Features
- Multi-tenant isolation (org_id mandated)
- Soft-delete strategy for all data
- Standardized API responses (RFC 7807)
- Event-driven architecture (NATS)
- Background job processing (Asynq/Redis)

## Setup & Running

### Requirements
- Go 1.24+
- Docker & Docker Compose
- Postgres 16+
- Redis 7+
- NATS 2.10+

### Local Development
```bash
make docker-up
make run
```

### Build
```bash
make build
make docker-build
```

### Testing
```bash
make test
```

## API Documentation
The API documentation is available in `api/openapi.yaml` and follows the OpenAPI 3.1.0 specification.
soul
