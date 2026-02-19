#!/usr/bin/env python3
"""Deterministic command lifecycle simulator for CI gate."""

from __future__ import annotations

import random
import sys

ALLOWED_TRANSITIONS = {
    "created": {"queued"},
    "queued": {"delivered", "failed"},
    "delivered": {"acked", "timed_out"},
    "acked": {"applied", "failed"},
    "applied": set(),
    "failed": set(),
    "timed_out": set(),
}


def simulate_one() -> list[str]:
    random.seed(7)
    state = "created"
    path = [state]

    while ALLOWED_TRANSITIONS[state]:
        options = sorted(ALLOWED_TRANSITIONS[state])
        pick = options[0] if len(options) == 1 else options[0 if random.random() < 0.8 else 1]
        state = pick
        path.append(state)
        if state in {"applied", "failed", "timed_out"}:
            break

    return path


def validate_path(path: list[str]) -> bool:
    for i in range(len(path) - 1):
        if path[i + 1] not in ALLOWED_TRANSITIONS[path[i]]:
            return False
    return True


def main() -> int:
    path = simulate_one()
    print(f"simulated_path={' -> '.join(path)}")

    if not validate_path(path):
        print("device_command_simulation=FAIL")
        return 1

    terminal = path[-1]
    if terminal not in {"applied", "failed", "timed_out"}:
        print("device_command_simulation=FAIL")
        return 1

    print(f"terminal_state={terminal}")
    print("device_command_simulation=PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
