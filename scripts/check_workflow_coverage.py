#!/usr/bin/env python3
"""Ensures core enterprise workflow files exist."""

from __future__ import annotations

import sys
from pathlib import Path

REQUIRED = [
    ".github/workflows/api-contract-clients.yml",
    ".github/workflows/all-services-foundation.yml",
    ".github/workflows/bff-gateway-check.yml",
    ".github/workflows/contract-tests.yml",
    ".github/workflows/docs-governance.yml",
    ".github/workflows/operations-governance.yml",
    ".github/workflows/performance-gates.yml",
    ".github/workflows/observability-foundation.yml",
    ".github/workflows/trace-propagation.yml",
    ".github/workflows/identity-auth-foundation.yml",
    ".github/workflows/design-system-foundation.yml",
    ".github/workflows/mobile-sync-foundation.yml",
    ".github/workflows/firmware-ota-foundation.yml",
    ".github/workflows/ota-simulation.yml",
    ".github/workflows/device-reliability-foundation.yml",
    ".github/workflows/vehicle-device-foundation.yml",
    ".github/workflows/enterprise-foundation.yml",
    ".github/workflows/runtime-integration.yml",
    ".github/workflows/service-integration-matrix.yml",
]


def main() -> int:
    missing = [p for p in REQUIRED if not Path(p).exists()]
    print(f"required_workflows={len(REQUIRED)}")

    if missing:
        print(f"workflow_coverage=FAIL missing={len(missing)}")
        for path in missing:
            print(f"missing_workflow={path}")
        return 1

    print("workflow_coverage=PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
