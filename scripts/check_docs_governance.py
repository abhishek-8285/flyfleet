#!/usr/bin/env python3
"""Docs governance checks for mandatory platform spec pack."""

from __future__ import annotations

import re
import sys
from pathlib import Path

REQUIRED_FILES = {
    "docs/README.md": [
        r"FlyFleet Documentation Index",
        r"Detailed Specs \(No Code Dependency\)",
    ],
    "docs/CROSS_PLATFORM_BLUEPRINT.md": [
        r"Cross-Platform Blueprint",
        r"Target Architecture",
    ],
    "docs/CROSS_PLATFORM_BACKLOG.md": [
        r"Cross-Platform Implementation Backlog",
        r"Execution Order",
    ],
    "docs/api/API_GOVERNANCE.md": [
        r"API Governance Specification",
        r"CI Gates",
    ],
    "docs/api/openapi.yaml": [
        r"^openapi:\s*3\.0\.3",
        r"^info:",
    ],
    "docs/architecture/mobile-sync.md": [
        r"Mobile Offline Sync Specification",
        r"Conflict Policy",
    ],
    "docs/firmware/OTA_SPEC.md": [
        r"Firmware OTA Specification",
        r"Device OTA State Machine",
    ],
    "docs/firmware/ota-manifest.schema.json": [
        r"\"title\":\s*\"FlyFleet OTA Manifest\"",
    ],
    "docs/device/DEVICE_RELIABILITY_SPEC.md": [
        r"Device Reliability Specification",
        r"Command Lifecycle",
    ],
    "docs/device/device_twin.schema.json": [
        r"\"title\":\s*\"Device Twin\"",
    ],
    "docs/device/command_lifecycle.schema.json": [
        r"\"title\":\s*\"Command Lifecycle Event\"",
    ],
    "docs/quality/TEST_MATRIX.md": [
        r"End-to-End Test Matrix",
        r"Performance Gates",
    ],
    "docs/operations/SLO_OPERATIONS.md": [
        r"SLO and Operations Specification",
        r"SLO Catalog",
    ],
    "docs/release/RELEASE_STRATEGY.md": [
        r"Release Strategy",
        r"Rollback Criteria",
    ],
    "docs/SERVICE_OWNERSHIP.md": [
        r"Service Ownership and Source-of-Truth Map",
        r"Domain Ownership",
    ],
    "docs/EXECUTION_STATUS.md": [
        r"Execution Status",
        r"Implemented",
    ],
}


def main() -> int:
    failures: list[str] = []
    checked = 0

    for file_path, patterns in REQUIRED_FILES.items():
        path = Path(file_path)
        if not path.exists():
            failures.append(f"missing_file: {file_path}")
            continue

        checked += 1
        content = path.read_text(encoding="utf-8", errors="ignore")
        for pattern in patterns:
            if not re.search(pattern, content, flags=re.MULTILINE):
                failures.append(f"missing_pattern: {file_path} -> /{pattern}/")

    print(f"docs_files_checked={checked}")
    print(f"docs_rules_total={sum(len(v) for v in REQUIRED_FILES.values())}")

    if failures:
        print(f"docs_governance=FAIL ({len(failures)} issues)")
        for failure in failures:
            print(failure)
        return 1

    print("docs_governance=PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
