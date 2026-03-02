#!/usr/bin/env python3
"""Checks identity/auth contract artifacts and claim-parity wiring."""

from __future__ import annotations

import json
import sys
from pathlib import Path

import yaml


def require_file(path: str) -> Path:
    p = Path(path)
    if not p.exists():
        print(f"identity_auth=FAIL missing_file={path}")
        raise SystemExit(1)
    return p


def check_schema() -> None:
    schema_path = require_file("apps/mobile/shared/auth/token_claims.schema.json")
    data = json.loads(schema_path.read_text(encoding="utf-8"))

    required = set(data.get("required", []))
    if not {"org_id", "role"}.issubset(required):
        print("identity_auth=FAIL schema_required_claims_missing")
        raise SystemExit(1)

    props = set((data.get("properties") or {}).keys())
    expected = {"sub", "user_id", "org_id", "role"}
    if not expected.issubset(props):
        print("identity_auth=FAIL schema_properties_missing")
        raise SystemExit(1)

    print("identity_auth_schema=PASS")


def check_docs() -> None:
    doc = require_file("docs/architecture/identity-auth-contract.md").read_text(
        encoding="utf-8"
    )
    for token in [
        "Access Token Claims Contract",
        "Session and Revocation Model",
        "Tenant and RBAC Scoping",
        "Enforcement",
    ]:
        if token not in doc:
            print(f"identity_auth=FAIL missing_doc_section={token}")
            raise SystemExit(1)
    print("identity_auth_docs=PASS")


def check_backend_claim_usage() -> None:
    checks = {
        "services/microservices/auth-service/internal/service/auth_service.go": ["org_id", "role"],
        "services/microservices/auth-service/internal/middleware/auth.go": ["org_id"],
        "services/microservices/billing-service/internal/middleware/auth.go": ["org_id", "sub", "role"],
        "services/microservices/web-bff/internal/middleware/rbac.go": ["org_id", "role"],
    }
    for path_str, tokens in checks.items():
        text = require_file(path_str).read_text(encoding="utf-8")
        for token in tokens:
            if token not in text:
                print(
                    f"identity_auth=FAIL claim_token_missing file={path_str} token={token}"
                )
                raise SystemExit(1)
    print("identity_auth_backend_claim_parity=PASS")


def check_mobile_policy() -> None:
    policy_path = require_file("apps/mobile/shared/sync/sync_policy.json")
    policy = json.loads(policy_path.read_text(encoding="utf-8"))
    pause_on_auth_failure = (
        policy.get("safeguards", {}).get("pause_on_auth_failure")
        or policy.get("security", {}).get("pause_on_auth_failure")
    )
    if not pause_on_auth_failure:
        print("identity_auth=FAIL mobile_policy_pause_on_auth_failure_missing")
        raise SystemExit(1)
    print("identity_auth_mobile_policy=PASS")


def check_openapi_auth_contract() -> None:
    spec_path = require_file("docs/api/openapi.yaml")
    spec = yaml.safe_load(spec_path.read_text(encoding="utf-8"))

    security = (spec or {}).get("security")
    if not isinstance(security, list) or not security:
        print("identity_auth=FAIL openapi_security_missing")
        raise SystemExit(1)

    has_bearer = any(isinstance(entry, dict) and "bearerAuth" in entry for entry in security)
    if not has_bearer:
        print("identity_auth=FAIL openapi_bearer_security_missing")
        raise SystemExit(1)

    components = (spec or {}).get("components")
    schemes = components.get("securitySchemes") if isinstance(components, dict) else None
    bearer = schemes.get("bearerAuth") if isinstance(schemes, dict) else None
    if not isinstance(bearer, dict):
        print("identity_auth=FAIL openapi_bearer_scheme_missing")
        raise SystemExit(1)
    if bearer.get("type") != "http" or bearer.get("scheme") != "bearer":
        print("identity_auth=FAIL openapi_bearer_scheme_invalid")
        raise SystemExit(1)

    print("identity_auth_openapi_security=PASS")


def check_bff_identity_bridge() -> None:
    text = require_file("services/microservices/web-bff/cmd/main.go").read_text(encoding="utf-8")
    required_tokens = [
        "propagateIdentityHeaders(req, r)",
        'outgoing.Header.Get("X-Organization-ID")',
        'outgoing.Header.Get("X-User-ID")',
        'incoming.Header.Get("Authorization")',
        'claims["org_id"]',
        'claims["user_id"]',
        'claims["sub"]',
        'outgoing.Header.Set("X-Organization-ID", orgID)',
        'outgoing.Header.Set("X-User-ID", userID)',
    ]
    for token in required_tokens:
        if token not in text:
            print(f"identity_auth=FAIL bff_identity_bridge_missing_token={token}")
            raise SystemExit(1)

    print("identity_auth_bff_bridge=PASS")


def main() -> int:
    check_schema()
    check_docs()
    check_backend_claim_usage()
    check_mobile_policy()
    check_openapi_auth_contract()
    check_bff_identity_bridge()
    print("identity_auth=PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
