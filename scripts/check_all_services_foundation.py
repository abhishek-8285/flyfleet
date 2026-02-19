#!/usr/bin/env python3
"""Enforces baseline readiness for all active service directories."""

from __future__ import annotations

import re
import sys
from pathlib import Path

SERVICES_ROOT = Path("services")
COMPOSE_PATH = Path("docker-compose.yml")

JWT_ALLOWED_BY_SERVICE = {
    "auth-service": {"${JWT_SECRET:-}", "${JWT_SECRET:-runtime-integration-secret}"},
    "fleet-core-service": {"${JWT_SECRET:-}", "${JWT_SECRET:-runtime-integration-secret}"},
}


def service_dirs() -> list[Path]:
    return sorted(
        [
            p
            for p in SERVICES_ROOT.iterdir()
            if p.is_dir() and p.name.endswith("service") and p.parent.name != "_archive"
        ]
    )


def check_service_files(service: Path) -> list[str]:
    errors: list[str] = []
    required = [
        service / "go.mod",
        service / "cmd" / "main.go",
        service / "README.md",
    ]
    for path in required:
        if not path.exists():
            errors.append(f"missing_file={path}")

    main_path = service / "cmd" / "main.go"
    if main_path.exists():
        text = main_path.read_text(encoding="utf-8", errors="ignore")
        if "/health" not in text:
            errors.append(f"missing_health_endpoint={main_path}")
    return errors


def parse_compose_envs() -> dict[str, list[str]]:
    env_by_service: dict[str, list[str]] = {}
    if not COMPOSE_PATH.exists():
        return env_by_service

    current_service: str | None = None
    in_services = False

    for raw in COMPOSE_PATH.read_text(encoding="utf-8").splitlines():
        if raw.strip() == "services:":
            in_services = True
            continue
        if not in_services:
            continue

        service_match = re.match(r"^  ([a-zA-Z0-9_-]+):\s*$", raw)
        if service_match:
            current_service = service_match.group(1)
            env_by_service.setdefault(current_service, [])
            continue

        env_match = re.match(r"^\s*-\s*JWT_SECRET=(.+)\s*$", raw)
        if current_service and env_match:
            env_by_service[current_service].append(env_match.group(1))

    return env_by_service


def check_compose_jwt(service_name: str, env_by_service: dict[str, list[str]]) -> list[str]:
    errors: list[str] = []
    if service_name not in env_by_service:
        return errors

    jwt_values = env_by_service.get(service_name, [])
    if not jwt_values:
        errors.append(f"compose_missing_jwt_secret service={service_name}")
        return errors

    if service_name in JWT_ALLOWED_BY_SERVICE:
        allowed = JWT_ALLOWED_BY_SERVICE[service_name]
    else:
        allowed = {"${JWT_SECRET:-runtime-integration-secret}"}

    if not any(v in allowed for v in jwt_values):
        errors.append(
            f"compose_unexpected_jwt_secret service={service_name} values={','.join(jwt_values)}"
        )
    return errors


def main() -> int:
    services = service_dirs()
    env_by_service = parse_compose_envs()
    failures: list[str] = []

    for service in services:
        failures.extend(check_service_files(service))
        failures.extend(check_compose_jwt(service.name, env_by_service))

    print(f"services_checked={len(services)}")
    if failures:
        print(f"all_services_foundation=FAIL count={len(failures)}")
        for failure in failures:
            print(f"failure={failure}")
        return 1

    print("all_services_foundation=PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
