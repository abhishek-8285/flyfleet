#!/usr/bin/env python3
"""Frontend API hygiene checks.

Fails on hardcoded localhost API/WS URLs and reports prefix usage stats.
"""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

ROOT = Path("apps/dashboard-web/src")
EXTENSIONS = {".ts", ".tsx", ".js", ".jsx"}

ABSOLUTE_LOCAL_RE = re.compile(
    r"(?P<quote>['\"`])(?P<url>(?:https?://|wss?://)(?:localhost|127\.0\.0\.1)(?::\d+)?/[^'\"`\s]*)"
)
API_V1_RE = re.compile(r"['\"`](/api/v1(?:/[^'\"`]*)?)")
API_RE = re.compile(r"['\"`](/api/[^'\"`]*)")
BFF_RE = re.compile(r"['\"`](/bff/[^'\"`]*)")
FIRMWARE_MANAGER_PATH = Path("apps/dashboard-web/src/pages/devices/FirmwareManager.tsx")
WEB_CONTRACT_PATH = Path("apps/dashboard-web/src/generated/api/contract.json")


def scan_files() -> tuple[int, int, int, int, list[tuple[str, int, str]]]:
    files_scanned = 0
    api_v1_count = 0
    api_count = 0
    bff_count = 0
    violations: list[tuple[str, int, str]] = []

    for path in sorted(ROOT.rglob("*")):
        if not path.is_file() or path.suffix not in EXTENSIONS:
            continue
        files_scanned += 1

        try:
            content = path.read_text(encoding="utf-8")
        except UnicodeDecodeError:
            continue

        lines = content.splitlines()
        for i, line in enumerate(lines, start=1):
            for match in ABSOLUTE_LOCAL_RE.finditer(line):
                violations.append((str(path), i, match.group("url")))

            api_v1_count += len(API_V1_RE.findall(line))
            api_count += len(API_RE.findall(line))
            bff_count += len(BFF_RE.findall(line))

    return files_scanned, api_v1_count, api_count, bff_count, violations


def check_rollout_event_filter_alignment() -> tuple[bool, str]:
    if not WEB_CONTRACT_PATH.exists():
        return False, f"missing_generated_contract={WEB_CONTRACT_PATH}"
    if not FIRMWARE_MANAGER_PATH.exists():
        return False, f"missing_firmware_manager={FIRMWARE_MANAGER_PATH}"

    try:
        contract = json.loads(WEB_CONTRACT_PATH.read_text(encoding="utf-8"))
    except Exception as exc:
        return False, f"invalid_generated_contract_json={exc}"

    operations = contract.get("operations")
    if not isinstance(operations, list):
        return False, "invalid_generated_contract_operations"

    event_enum: set[str] | None = None
    for op in operations:
        if not isinstance(op, dict):
            continue
        if op.get("path") != "/firmware/rollouts/{id}/events" or op.get("method") != "GET":
            continue
        params = op.get("parameters")
        if not isinstance(params, list):
            continue
        for p in params:
            if not isinstance(p, dict):
                continue
            if p.get("name") == "event_type" and p.get("in") == "query":
                schema = p.get("schema")
                enum = schema.get("enum") if isinstance(schema, dict) else None
                if isinstance(enum, list):
                    event_enum = {str(v) for v in enum}
                break
        break

    if not event_enum:
        return False, "missing_event_type_enum_in_generated_contract"

    required_aliases = {"note", "transition", "created"}
    if not required_aliases.issubset(event_enum):
        return False, f"generated_contract_missing_aliases={sorted(required_aliases - event_enum)}"

    content = FIRMWARE_MANAGER_PATH.read_text(encoding="utf-8", errors="ignore")
    for alias in sorted(required_aliases):
        if f'<option value="{alias}">' not in content:
            return False, f"firmware_manager_missing_filter_option={alias}"

    if "params.set('event_type', eventFilter)" not in content:
        return False, "firmware_manager_missing_event_type_query_binding"

    return True, "frontend_rollout_filter_contract=PASS"


def main() -> int:
    if not ROOT.exists():
        print(f"error: source root not found: {ROOT}")
        return 2

    files_scanned, api_v1_count, api_count, bff_count, violations = scan_files()

    legacy_api_count = max(api_count - api_v1_count, 0)

    print(f"files_scanned={files_scanned}")
    print(f"api_v1_paths={api_v1_count}")
    print(f"api_paths_total={api_count}")
    print(f"api_paths_non_v1={legacy_api_count}")
    print(f"bff_paths={bff_count}")

    aligned, detail = check_rollout_event_filter_alignment()
    if not aligned:
        print(detail)
        print("frontend_api_hygiene=FAIL")
        return 1
    print(detail)

    if violations:
        print(f"hardcoded_local_urls={len(violations)}")
        for file_path, line_no, url in violations:
            print(f"violation {file_path}:{line_no} -> {url}")
        print("frontend_api_hygiene=FAIL")
        return 1

    print("hardcoded_local_urls=0")
    print("frontend_api_hygiene=PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
