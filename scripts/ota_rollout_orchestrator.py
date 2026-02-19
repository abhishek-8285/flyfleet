#!/usr/bin/env python3
"""Deterministic OTA rollout orchestration simulator.

Evaluates ring progression decisions from a manifest rollout policy:
- promote when all observed rates are within thresholds
- pause_and_rollback when any stop condition is breached
"""

from __future__ import annotations

import json
import sys
from pathlib import Path


def _rate(numerator: int, denominator: int) -> float:
    if denominator <= 0:
        return 0.0
    return float(numerator) / float(denominator)


def decide_action(stats: dict, thresholds: dict) -> tuple[str, dict]:
    attempted = int(stats.get("attempted", 0))
    failed = int(stats.get("failed", 0))
    boot_failed = int(stats.get("boot_failed", 0))
    silent = int(stats.get("silent", 0))

    rates = {
        "failure_rate": _rate(failed, attempted),
        "boot_failure_rate": _rate(boot_failed, attempted),
        "silence_rate": _rate(silent, attempted),
    }

    if rates["failure_rate"] > float(thresholds["max_failure_rate"]):
        return "pause_and_rollback", rates
    if rates["boot_failure_rate"] > float(thresholds["max_boot_failure_rate"]):
        return "pause_and_rollback", rates
    if rates["silence_rate"] > float(thresholds["max_silence_rate"]):
        return "pause_and_rollback", rates
    return "promote", rates


def run_scenario(manifest: dict, scenario: dict) -> tuple[bool, str]:
    policy = manifest.get("rollout_policy", {})
    rings = policy.get("rings", [])
    thresholds = policy.get("auto_pause_thresholds", {})
    if not isinstance(rings, list) or not isinstance(thresholds, dict):
        return False, "invalid_rollout_policy"

    ring_observations = scenario.get("rings", {})
    if not isinstance(ring_observations, dict):
        return False, "invalid_scenario_rings"

    name = str(scenario.get("name", "unknown"))
    expected_action = str(scenario.get("expected_terminal_action", ""))
    expected_ring = str(scenario.get("expected_terminal_ring", ""))

    terminal_action = "promote"
    terminal_ring = ""
    evaluated = 0

    for ring in rings:
        ring_name = str(ring.get("name", ""))
        if not ring_name:
            return False, f"scenario={name} invalid_ring_name"
        stats = ring_observations.get(ring_name)
        if not isinstance(stats, dict):
            continue
        action, rates = decide_action(stats, thresholds)
        evaluated += 1
        print(
            f"ota_ring_decision scenario={name} ring={ring_name} action={action} "
            f"failure_rate={rates['failure_rate']:.4f} "
            f"boot_failure_rate={rates['boot_failure_rate']:.4f} "
            f"silence_rate={rates['silence_rate']:.4f}"
        )
        terminal_action = action
        terminal_ring = ring_name
        if action == "pause_and_rollback":
            break

    if evaluated == 0:
        return False, f"scenario={name} no_rings_evaluated"
    if terminal_action != expected_action:
        return (
            False,
            f"scenario={name} action_mismatch expected={expected_action} actual={terminal_action}",
        )
    if terminal_ring != expected_ring:
        return (
            False,
            f"scenario={name} ring_mismatch expected={expected_ring} actual={terminal_ring}",
        )
    return True, f"scenario={name} result=PASS"


def main() -> int:
    manifest_path = Path(sys.argv[1]) if len(sys.argv) > 1 else Path("firmware/releases/sample/ota-manifest.json")
    scenarios_path = (
        Path(sys.argv[2]) if len(sys.argv) > 2 else Path("firmware/releases/sample/ota-rollout-scenarios.json")
    )

    if not manifest_path.exists():
        print(f"missing_manifest={manifest_path}")
        return 1
    if not scenarios_path.exists():
        print(f"missing_scenarios={scenarios_path}")
        return 1

    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    scenarios = json.loads(scenarios_path.read_text(encoding="utf-8"))
    if not isinstance(scenarios, list) or not scenarios:
        print("invalid_scenarios_file")
        return 1

    failures: list[str] = []
    passed = 0
    for scenario in scenarios:
        if not isinstance(scenario, dict):
            failures.append("invalid_scenario_entry")
            continue
        ok, message = run_scenario(manifest, scenario)
        if ok:
            passed += 1
            print(message)
        else:
            failures.append(message)

    print(f"ota_orchestration_scenarios_total={len(scenarios)}")
    print(f"ota_orchestration_scenarios_passed={passed}")
    if failures:
        print(f"ota_rollout_orchestration_simulation=FAIL ({len(failures)} issues)")
        for failure in failures:
            print(failure)
        return 1

    print("ota_rollout_orchestration_simulation=PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
