# Billing Service

Microservice for managing financial operations, including invoicing and payment tracking.

## Features
- RFC 7807 Standard Error Responses
- JWT Authentication & Tenant Isolation
- NATS Event Publishing
- Background Job Processing (Asynq)
- OpenAPI 3.1.0 Documentation

## Setup
```bash
make docker-up
make run
```

## Testing
```bash
make test
```

