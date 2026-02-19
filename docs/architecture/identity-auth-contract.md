# Identity and Auth Contract (Web + Mobile + Backend + Device Edge)

## Purpose
Define one token/session contract and authorization baseline across all surfaces.

## Access Token Claims Contract
Canonical schema:
- `apps/mobile/shared/auth/token_claims.schema.json`

Required claims:
1. `org_id` (UUID): tenant scope.
2. `role` (string): effective RBAC role.
3. `sub` or `user_id`: user identity.

Optional operational claims:
1. `exp`, `iat`, `jti`.

## Compatibility Rule
During migration, services may read either:
1. `sub` (preferred)
2. `user_id` (legacy)

Tokens must carry at least one of the above.

## Session and Revocation Model
1. Auth service remains source of truth for token/session issuance and revocation.
2. Revocation events must invalidate access within platform SLA targets.
3. Mobile sync workers must pause on auth failures until refresh/login recovery.

## Tenant and RBAC Scoping
1. Backend mutating APIs require valid `org_id` scope.
2. Protected APIs require role checks that match UI route guards.
3. Cross-org access is forbidden unless explicitly allowed by super-admin policy.

## Enforcement
CI gate:
1. `scripts/check_identity_auth_contract.py`
2. `.github/workflows/identity-auth-foundation.yml`

The gate validates:
1. claim schema presence and required fields,
2. backend middleware usage of `org_id` and role claims,
3. auth-sensitive mobile sync policy presence.
