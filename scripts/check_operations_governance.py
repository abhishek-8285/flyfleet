#!/usr/bin/env python3
"""Enforces SLO/release/test-matrix governance docs and workflow coverage."""

from __future__ import annotations

import sys
from pathlib import Path

DOC_REQUIREMENTS = {
    "docs/operations/SLO_OPERATIONS.md": [
        "SLO Catalog",
        "Error Budget Policy",
        "Alert Routing",
        "Incident Operating Procedure",
        "Review Cadence",
    ],
    "docs/release/RELEASE_STRATEGY.md": [
        "Environment Model",
        "Release Trains",
        "Promotion Criteria",
        "Rollback Criteria",
        "Rollback Playbooks",
    ],
    "docs/quality/TEST_MATRIX.md": [
        "Test Layers",
        "Matrix",
        "Minimum Required Journeys",
        "Performance Gates",
        "Exit Criteria Per Release",
    ],
}

WORKFLOW_REQUIREMENTS = [
    ".github/workflows/api-contract-clients.yml",
    ".github/workflows/bff-gateway-check.yml",
    ".github/workflows/contract-tests.yml",
    ".github/workflows/performance-gates.yml",
    ".github/workflows/observability-foundation.yml",
    ".github/workflows/mobile-sync-foundation.yml",
    ".github/workflows/firmware-ota-foundation.yml",
    ".github/workflows/ota-simulation.yml",
    ".github/workflows/device-reliability-foundation.yml",
    ".github/workflows/vehicle-device-foundation.yml",
    ".github/workflows/runtime-integration.yml",
    ".github/workflows/trace-propagation.yml",
    ".github/workflows/service-integration-matrix.yml",
]


def check_docs() -> int:
    failures = 0
    for path_str, required_tokens in DOC_REQUIREMENTS.items():
        path = Path(path_str)
        if not path.exists():
            print(f"ops_governance=FAIL missing_doc={path_str}")
            failures += 1
            continue

        text = path.read_text(encoding="utf-8")
        missing = [token for token in required_tokens if token not in text]
        if missing:
            print(
                f"ops_governance=FAIL doc={path_str} missing_tokens={','.join(missing)}"
            )
            failures += 1
        else:
            print(f"ops_governance_doc=PASS doc={path_str}")
    return failures


def check_workflows() -> int:
    missing = [wf for wf in WORKFLOW_REQUIREMENTS if not Path(wf).exists()]
    print(f"ops_governance_required_workflows={len(WORKFLOW_REQUIREMENTS)}")
    if missing:
        print(f"ops_governance=FAIL missing_workflows={len(missing)}")
        for wf in missing:
            print(f"missing_workflow={wf}")
        return 1
    print("ops_governance_workflows=PASS")
    return 0


def main() -> int:
    failures = 0
    failures += check_docs()
    failures += check_workflows()

    if failures:
        print("ops_governance=FAIL")
        return 1

    print("ops_governance=PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
