# FlyFleet Enterprise Monorepo

Welcome to the unified FlyFleet Enterprise platform. This repository is organized as a high-performance monorepo supporting 12+ Go microservices and multiple Next.js/React frontend applications.

## Directory Structure

```
/flyfleet (Root)
├── apps/                # End-user applications (Frontends)
│   ├── landing-page/    # Next.js customer-facing landing page
│   └── dashboard-web/   # React/Vite administrative dashboard (migrated from CRA)
├── services/            # Backend Microservices (Go)
│   ├── auth-service/    # Identity and Access Management
│   ├── vehicle-service/ # Fleet and Asset Management
│   └── ...              # (12 total services)
├── packages/            # Shared libraries and configurations
│   ├── common/go        # Shared Go logic (fleet-platform/shared)
│   ├── config/          # Centralized whitelabel branding
│   └── ui-components/   # Shared design system (WIP)
├── infrastructure/      # IaC, CI/CD, and Orchestration
├── docs/                # Architecture, API, and Migration logs
└── scripts/             # Internal tooling and scripts
```

## Getting Started

### Prerequisites
- Go 1.23+
- Node.js 20+
- Docker & Docker Compose

### Development

1. **Backend**: Services are managed via `go.work`. Run `go build ./...` from the root to verify all services.
2. **Frontend**: Apps are managed via NPM Workspaces. Run `npm install` from the root to link dependencies.
3. **Branding**: Identity is managed in `packages/config/branding/default.json`.

## Architecture Principles

- **Zero Hardcoding**: All branding assets must be consumed from `@flyfleet/config`.
- **Stateless Services**: All microservices are designed for horizontal scalability.
- **Unified Logging**: Standardized Zap logging implemented in `packages/common/go`.

---
© 2026 Fleet Platform. Confidential and Proprietary.
