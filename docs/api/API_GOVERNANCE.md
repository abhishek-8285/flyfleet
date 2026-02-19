# API Governance Specification

## Scope
Applies to all externally consumed APIs: web, mobile, partner integrations, device control APIs.

## Canonical Artifacts
1. Primary OpenAPI spec: `docs/api/openapi.yaml`
2. Domain-level supplements (optional): `services/*/api/openapi/*.yaml`
3. Error contract: `{ code, message, details, trace_id }`

## Versioning Policy
1. Public routes must be versioned (`/api/v1/...`, `/api/v2/...`).
2. Additive changes are allowed in a minor release:
   - new endpoint
   - new optional request field
   - new optional response field
3. Breaking changes require major version introduction:
   - remove/rename field
   - change field type
   - change auth semantics
   - remove endpoint

## Deprecation Policy
1. Deprecated operations must include:
   - `deprecated: true`
   - `x-replaced-by: <new operation path>`
   - `x-sunset-date: YYYY-MM-DD`
2. Minimum deprecation window: 90 days before removal.
3. Changelog entry required in release notes for each deprecated operation.

## Auth and Tenancy Requirements
1. Protected operations must declare bearer auth.
2. Every org-scoped operation must document org scoping behavior.
3. RBAC expectations must be described in operation description or extension field.

## API Quality Requirements
1. Every operation must define:
   - summary
   - tags
   - success response schema
   - error response schema for auth and validation errors
2. Pagination endpoints must expose deterministic page/limit or cursor semantics.
3. Idempotent write operations must define idempotency key behavior.

## Generated Clients
1. Web and mobile consumers use generated clients for new API usage.
2. Manual HTTP calls are allowed only for migration windows and must be tracked.
3. Generation output paths:
   - Web: `apps/dashboard-web/src/generated/api/`
   - Mobile: `apps/mobile/shared/generated/`
4. Generation command:
   - `python3 scripts/generate_api_clients.py`
5. Generated contract shape (`contract.json`, `api_contract.json`) must include for every operation:
   - `operationId`, `method`, `path`, `summary`, `tags`
   - `parameters[]` entries with: `name`, `in`, `required`, `schema`
6. Parameter enums in generated contracts are source-of-truth for cross-surface filters.
   - Example: rollout events `event_type` supports:
     - `note`, `transition`, `created`
     - `rollout.note`, `rollout.transition`, `rollout.created`
7. CI enforces generated contract freshness and semantic parameter parity for critical operations.

## CI Gates
1. OpenAPI syntax validation.
2. Governance rule validation (deprecated operation metadata, operation completeness).
3. Contract tests for implementation drift.
4. Generated client freshness check.
5. Workflow: `.github/workflows/api-contract-clients.yml`

## PR Checklist (Required)
1. Spec updated for behavior changes.
2. Contract tests added/updated.
3. Generated clients refreshed if contract changed.
4. Deprecation metadata added where relevant.
5. Migration notes added for breaking changes.
