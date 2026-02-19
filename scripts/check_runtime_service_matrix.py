#!/usr/bin/env python3
"""Enforces runtime service integration matrix coverage."""

from __future__ import annotations

import re
import sys
from pathlib import Path


RULES = {
    "scripts/runtime_integration_gate.sh": [
        r"service_health_matrix=PASS",
        r"SERVICE_HEALTH_ENDPOINTS=\(",
        r"/api/v1/admin/governance/requests\|200",
        r"/api/v1/admin/devices\?page=1&size=5\|200",
        r"/api/v1/admin/dr/tests\|200",
        r"/api/v1/admin/dr/tests/metrics\|200",
        r"/api/v1/admin/dr/tests/settings\|200",
        r"/api/v1/admin/reports/schedules\|200",
        r"/api/v1/admin/reports/executions\|200",
        r"/api/v1/admin/settings/platform\|200",
        r"/api/v1/admin/settings/audit\?limit=10\|200",
        r"/api/v1/admin/finance/expenses/reports\?limit=5\|200",
        r"/api/v1/admin/revenue/forecast\?route_id=BLR-MYS\|200",
        r"settings_export_code=\$settings_export_code",
        r"settings_export_semantics=PASS",
        r"settings_import_code=\$settings_import_code",
        r"settings_import_semantics=PASS",
        r"settings_import_invalid_code=\$settings_import_invalid_code",
        r"settings_audit_invalid_limit_code=\$settings_audit_invalid_limit_code",
        r"/api/v1/admin/settings/export",
        r"/api/v1/admin/settings/import",
        r"settings-service\|http://localhost:8091/health",
    ],
    ".github/workflows/service-integration-matrix.yml": [
        r"name:\s*Service Integration Matrix",
        r"bash scripts/runtime_integration_gate.sh",
    ],
}


def main() -> int:
    failures: list[str] = []
    checked = 0

    for file_path, patterns in RULES.items():
        path = Path(file_path)
        if not path.exists():
            failures.append(f"missing_file: {file_path}")
            continue
        checked += 1
        content = path.read_text(encoding="utf-8", errors="ignore")
        for pattern in patterns:
            if not re.search(pattern, content):
                failures.append(f"missing_pattern: {file_path} -> /{pattern}/")

    print(f"runtime_service_matrix_files_checked={checked}")
    if failures:
        print(f"runtime_service_matrix=FAIL ({len(failures)} issues)")
        for failure in failures:
            print(failure)
        return 1

    print("runtime_service_matrix=PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
