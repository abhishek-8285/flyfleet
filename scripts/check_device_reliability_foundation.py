#!/usr/bin/env python3
"""Checks digital twin and command reliability foundation artifacts."""

from __future__ import annotations

import re
import subprocess
import sys
from pathlib import Path

RULES = {
    "docs/device/DEVICE_RELIABILITY_SPEC.md": [r"Digital Twin Model", r"Command Lifecycle", r"SLA Targets"],
    "docs/device/device_twin.schema.json": [r"Device Twin", r"drift_status"],
    "docs/device/command_lifecycle.schema.json": [r"Command Lifecycle Event", r"timed_out"],
    "scripts/device_command_simulator.py": [r"device_command_simulation=PASS"],
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

    print(f"device_reliability_files_checked={checked}")

    run = subprocess.run(["python3", "scripts/device_command_simulator.py"], capture_output=True, text=True)
    if run.returncode != 0:
        print(run.stdout, end="")
        print(run.stderr, end="")
        failures.append("device_simulator_failed")

    if failures:
        print(f"device_reliability_foundation=FAIL ({len(failures)} issues)")
        for f in failures:
            print(f)
        return 1

    print("device_reliability_foundation=PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
