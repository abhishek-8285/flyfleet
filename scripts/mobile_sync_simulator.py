#!/usr/bin/env python3
"""Deterministic simulator for mobile offline sync retry behavior."""

from __future__ import annotations

import datetime as dt
import json
import random
import sys
from pathlib import Path

POLICY_PATH = Path("apps/mobile/shared/sync/sync_policy.json")


class SyncOp:
    def __init__(self, op_id: str):
        now = dt.datetime.now(dt.timezone.utc)
        self.op_id = op_id
        self.retry_count = 0
        self.status = "pending"
        self.next_retry_at = now


def next_delay_seconds(base: int, retries: int, max_delay: int, jitter_ratio: float) -> int:
    raw = min((2 ** retries) * base, max_delay)
    jitter = int(raw * jitter_ratio)
    return max(1, raw + random.randint(-jitter, jitter))


def simulate() -> tuple[int, int, int]:
    random.seed(42)
    policy = json.loads(POLICY_PATH.read_text(encoding="utf-8"))
    retry = policy["retry"]

    operations = [SyncOp(f"op-{i}") for i in range(1, 11)]

    applied = 0
    dead = 0
    conflicts = 0

    for op in operations:
        while op.retry_count <= retry["max_retries"]:
            roll = random.random()
            if roll < 0.70:
                op.status = "applied"
                applied += 1
                break
            if roll < 0.85:
                op.status = "conflict"
                conflicts += 1
                break

            op.retry_count += 1
            if op.retry_count > retry["max_retries"]:
                op.status = "dead_letter"
                dead += 1
                break

            delay = next_delay_seconds(
                retry["base_delay_seconds"],
                op.retry_count,
                retry["max_delay_seconds"],
                retry["jitter_ratio"],
            )
            op.next_retry_at = op.next_retry_at + dt.timedelta(seconds=delay)

    return applied, conflicts, dead


def main() -> int:
    if not POLICY_PATH.exists():
        print(f"missing_policy={POLICY_PATH}")
        return 1

    applied, conflicts, dead = simulate()
    print("simulated_ops=10")
    print(f"applied={applied}")
    print(f"conflicts={conflicts}")
    print(f"dead_letter={dead}")

    if applied + conflicts + dead != 10:
        print("mobile_sync_simulation=FAIL")
        return 1

    print("mobile_sync_simulation=PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
