# FlyFleet Enterprise Migration Log

This log documents the structural changes made to the FlyFleet codebase during the Enterprise Reorganization.

## Phase 1: Frontend Consolidation
**Date**: 2026-02-06
- **Moved**: `ui/landing-page` -> `apps/landing-page`
- **Moved**: `ui/dashboard` -> `apps/dashboard`
- **Moved**: `web-admin` -> `apps/admin-web`
- **Rationale**: Groups all user-facing applications into a unified `apps/` directory for better discoverability and workspace management.

## Phase 2: Backend Consolidation
**Date**: 2026-02-06
- **Moved**: `backend/services/*` -> `services/*`
- **Rationale**: Lifts microservices to the top level, emphasizing their role as independent deployment units in a monorepo.

## Phase 3: Infrastructure and Packages (In Progress)
**Date**: 2026-02-06
- **Created**: `packages/config`, `packages/common`, `packages/ui-components`, `infrastructure`, `docs`, `scripts`
- **Rationale**: Sets up the foundation for shared logic, configuration-driven branding, and automated deployment pipelines.

## Phase 4: Workspace Standardization
**Date**: 2026-02-06
- **Updated**: `go.work` to include all 12 microservices and the new `packages/common/go`.
- **Consolidated**: `apps/admin-web` and `apps/dashboard` (doc-only) into `apps/dashboard-web`.
- **Rationale**: Standardizes frontend naming and removes redundant documentation placeholders to ensure a clean source-of-truth.

## Phase 5: Branding Cleanup (In Progress)
**Date**: 2026-02-06
- **Scrubbed**: Hardcoded "FlyFleet" strings from `apps/landing-page` and `apps/dashboard-web`.
- **Initialized**: `packages/config` workspace package for centralized identity management.
- **Linked**: `@flyfleet/config` to all frontend applications via NPM Workspaces.
- **Cleanup**: Removed legacy `package-lock.json` and `node_modules` to ensure clean workspace linking.
- **Refactored**: Updated all 12 Go microservice `go.mod` files and imports to use the new monorepo hierarchy.
- **Unified**: Replaced legacy `github.com/flyfleet/backend` paths with standardized `github.com/flyfleet` and `github.com/fleet-platform/shared` paths.
- **Rationale**: Standardizes backend dependency management and follows Go best practices for monorepo development.
