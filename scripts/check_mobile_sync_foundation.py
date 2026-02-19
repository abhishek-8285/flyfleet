#!/usr/bin/env python3
"""Checks required mobile sync foundation artifacts."""

from __future__ import annotations

import json
import re
import subprocess
import sys
import tempfile
from pathlib import Path

RULES = {
    "apps/mobile/shared/sync/operation.schema.json": [r"Mobile Sync Operation", r"dead_letter"],
    "apps/mobile/shared/sync/sync_policy.json": [r"max_retries", r"server_wins_rebase"],
    "docs/architecture/mobile-sync.md": [r"Retry Policy", r"Conflict Policy", r"Acceptance Tests"],
    "scripts/mobile_sync_simulator.py": [r"mobile_sync_simulation=PASS"],
    "scripts/mobile_dead_letter_inspector.py": [r"mobile_dead_letter_inspection=PASS"],
}

POLICY_PATH = Path("apps/mobile/shared/sync/sync_policy.json")


def validate_policy_semantics() -> list[str]:
    failures: list[str] = []
    try:
        policy = json.loads(POLICY_PATH.read_text(encoding="utf-8"))
    except Exception as exc:
        return [f"invalid_sync_policy_json: {exc}"]

    retry = policy.get("retry")
    if not isinstance(retry, dict):
        return ["sync_policy_retry_missing"]

    base = retry.get("base_delay_seconds")
    max_delay = retry.get("max_delay_seconds")
    max_retries = retry.get("max_retries")
    jitter = retry.get("jitter_ratio")

    if not isinstance(base, int) or not (1 <= base <= 30):
        failures.append("sync_policy_invalid_base_delay_seconds")
    if not isinstance(max_delay, int) or not (60 <= max_delay <= 3600):
        failures.append("sync_policy_invalid_max_delay_seconds")
    if isinstance(base, int) and isinstance(max_delay, int) and max_delay <= base:
        failures.append("sync_policy_max_delay_not_greater_than_base")
    if not isinstance(max_retries, int) or not (1 <= max_retries <= 12):
        failures.append("sync_policy_invalid_max_retries")
    if not isinstance(jitter, (int, float)) or not (0 <= float(jitter) <= 0.5):
        failures.append("sync_policy_invalid_jitter_ratio")

    conflict = policy.get("conflict")
    if not isinstance(conflict, dict):
        failures.append("sync_policy_conflict_missing")
    else:
        if conflict.get("default_strategy") != "server_wins_rebase":
            failures.append("sync_policy_invalid_default_conflict_strategy")
        overrides = conflict.get("entity_overrides")
        if not isinstance(overrides, dict):
            failures.append("sync_policy_entity_overrides_missing")
        else:
            expected = {
                "trip": "server_authority",
                "note": "append_merge",
                "work_order": "field_level_merge",
            }
            for key, value in expected.items():
                if overrides.get(key) != value:
                    failures.append(f"sync_policy_entity_override_invalid:{key}")

    security = policy.get("security")
    if not isinstance(security, dict):
        failures.append("sync_policy_security_missing")
    else:
        if security.get("pause_on_auth_failure") is not True:
            failures.append("sync_policy_pause_on_auth_failure_not_true")
        if security.get("require_secure_storage") is not True:
            failures.append("sync_policy_require_secure_storage_not_true")

    ux = policy.get("ux")
    if not isinstance(ux, dict):
        failures.append("sync_policy_ux_missing")
    else:
        states = ux.get("states")
        if not isinstance(states, list):
            failures.append("sync_policy_ux_states_missing")
        else:
            required_states = {"up_to_date", "syncing", "action_required"}
            if not required_states.issubset({str(s) for s in states}):
                failures.append("sync_policy_ux_states_incomplete")

    return failures


def main() -> int:
    failures = []
    checked = 0

    for file_path, patterns in RULES.items():
        path = Path(file_path)
        if not path.exists():
            failures.append(f"missing_file: {file_path}")
            continue
        checked += 1
        content = path.read_text(encoding="utf-8", errors="ignore")
        for p in patterns:
            if not re.search(p, content):
                failures.append(f"missing_pattern: {file_path} -> /{p}/")

    print(f"mobile_sync_files_checked={checked}")

    policy_failures = validate_policy_semantics()
    if policy_failures:
        failures.extend(policy_failures)
    else:
        print("mobile_sync_policy_semantics=PASS")

    sim = subprocess.run(["python3", "scripts/mobile_sync_simulator.py"], capture_output=True, text=True)
    if sim.returncode != 0:
        print(sim.stdout, end="")
        print(sim.stderr, end="")
        failures.append("mobile_sync_simulator_failed")
    elif "mobile_sync_simulation=PASS" not in sim.stdout:
        failures.append("mobile_sync_simulator_missing_pass_marker")

    sample_xml = """<?xml version='1.0' encoding='utf-8' standalone='yes' ?>
<map>
  <string name="gps_dead_letter_ids">["evt-1","evt-2"]</string>
  <string name="gps_dead_letter_events">[{"reason":"max_retries_exceeded","id":"evt-1"},{"reason":"max_retries_exceeded","id":"evt-2"}]</string>
</map>
"""
    with tempfile.TemporaryDirectory(prefix="mobile-sync-check-") as tmpdir:
        sample_path = Path(tmpdir) / "prefs.xml"
        sample_path.write_text(sample_xml, encoding="utf-8")
        inspect = subprocess.run(
            ["python3", "scripts/mobile_dead_letter_inspector.py", str(sample_path)],
            capture_output=True,
            text=True,
        )
        if inspect.returncode != 0:
            print(inspect.stdout, end="")
            print(inspect.stderr, end="")
            failures.append("mobile_dead_letter_inspector_failed")
        elif "mobile_dead_letter_inspection=PASS" not in inspect.stdout:
            failures.append("mobile_dead_letter_inspector_missing_pass_marker")

    if failures:
        print(f"mobile_sync_foundation=FAIL ({len(failures)} issues)")
        for failure in failures:
            print(failure)
        return 1

    print("mobile_sync_foundation=PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
