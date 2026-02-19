#!/usr/bin/env python3
"""Runs all non-runtime enterprise foundation checks."""

from __future__ import annotations

import subprocess
import sys

CHECKS = [
    ["python3", "scripts/check_docs_governance.py"],
    ["python3", "scripts/check_operations_governance.py"],
    ["python3", "scripts/check_workflow_coverage.py"],
    ["python3", "scripts/check_generated_clients_freshness.py"],
    ["python3", "scripts/check_all_services_foundation.py"],
    ["python3", "scripts/check_frontend_api_hygiene.py"],
    ["python3", "scripts/check_bff_contract.py"],
    ["python3", "scripts/check_trace_propagation.py"],
    ["python3", "scripts/check_identity_auth_contract.py"],
    ["python3", "scripts/check_design_system_foundation.py"],
    ["python3", "scripts/check_observability_foundation.py"],
    ["python3", "scripts/check_mobile_sync_foundation.py"],
    ["python3", "scripts/check_firmware_ota_foundation.py"],
    ["python3", "scripts/check_device_reliability_foundation.py"],
    ["python3", "scripts/check_runtime_service_matrix.py"],
]


def main() -> int:
    for cmd in CHECKS:
        print(f"running={' '.join(cmd)}")
        run = subprocess.run(cmd)
        if run.returncode != 0:
            print(f"enterprise_foundation=FAIL command={' '.join(cmd)}")
            return run.returncode

    print("enterprise_foundation=PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
