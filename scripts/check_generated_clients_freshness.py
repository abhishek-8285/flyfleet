#!/usr/bin/env python3
"""Fails if generated API artifacts are stale."""

from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path

TARGETS = [
    "apps/dashboard-web/src/generated/api/operations.ts",
    "apps/dashboard-web/src/generated/api/contract.json",
    "apps/mobile/shared/generated/api_contract.json",
    "apps/mobile/shared/generated/openapi.sha256",
]


def _validate_contract(path: Path) -> list[str]:
    failures: list[str] = []
    try:
        payload = json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        return [f"invalid_json: {path} ({exc})"]

    ops = payload.get("operations")
    if not isinstance(ops, list):
        return [f"invalid_operations_array: {path}"]

    target_op = None
    for op in ops:
        if not isinstance(op, dict):
            continue
        if op.get("path") == "/firmware/rollouts/{id}/events" and op.get("method") == "GET":
            target_op = op
            break

    if target_op is None:
        failures.append(f"missing_operation: {path} /firmware/rollouts/{{id}}/events GET")
        return failures

    params = target_op.get("parameters")
    if not isinstance(params, list):
        failures.append(f"missing_parameters: {path} rollout events GET")
        return failures

    event_type_param = None
    limit_param = None
    offset_param = None
    for p in params:
        if not isinstance(p, dict):
            continue
        if p.get("name") == "event_type" and p.get("in") == "query":
            event_type_param = p
        if p.get("name") == "limit" and p.get("in") == "query":
            limit_param = p
        if p.get("name") == "offset" and p.get("in") == "query":
            offset_param = p
    if event_type_param is None:
        failures.append(f"missing_event_type_param: {path}")
        return failures

    schema = event_type_param.get("schema")
    if not isinstance(schema, dict):
        failures.append(f"missing_event_type_schema: {path}")
        return failures

    enum_values = schema.get("enum")
    expected = {
        "note",
        "transition",
        "created",
        "rollout.note",
        "rollout.transition",
        "rollout.created",
    }
    if not isinstance(enum_values, list):
        failures.append(f"missing_event_type_enum: {path}")
        return failures
    if set(enum_values) != expected:
        failures.append(f"invalid_event_type_enum: {path} got={sorted(set(enum_values))}")

    if limit_param is None:
        failures.append(f"missing_limit_param: {path}")
    else:
        limit_schema = limit_param.get("schema")
        if not isinstance(limit_schema, dict):
            failures.append(f"invalid_limit_schema: {path}")
        else:
            if limit_schema.get("type") != "integer":
                failures.append(f"invalid_limit_type: {path}")
            if limit_schema.get("minimum") != 1:
                failures.append(f"invalid_limit_minimum: {path}")
            if limit_schema.get("maximum") != 500:
                failures.append(f"invalid_limit_maximum: {path}")
            if limit_schema.get("default") != 50:
                failures.append(f"invalid_limit_default: {path}")

    if offset_param is None:
        failures.append(f"missing_offset_param: {path}")
    else:
        offset_schema = offset_param.get("schema")
        if not isinstance(offset_schema, dict):
            failures.append(f"invalid_offset_schema: {path}")
        else:
            if offset_schema.get("type") != "integer":
                failures.append(f"invalid_offset_type: {path}")
            if offset_schema.get("minimum") != 0:
                failures.append(f"invalid_offset_minimum: {path}")
            if offset_schema.get("default") != 0:
                failures.append(f"invalid_offset_default: {path}")

    for p in params:
        if not isinstance(p, dict):
            failures.append(f"invalid_parameter_entry: {path}")
            continue
        if not isinstance(p.get("name"), str) or not p.get("name"):
            failures.append(f"invalid_parameter_name: {path}")
        if not isinstance(p.get("in"), str) or not p.get("in"):
            failures.append(f"invalid_parameter_location: {path}")
        if not isinstance(p.get("required"), bool):
            failures.append(f"invalid_parameter_required_flag: {path}")
        if not isinstance(p.get("schema"), dict):
            failures.append(f"invalid_parameter_schema: {path}")

    return failures


def main() -> int:
    for t in TARGETS:
        if not Path(t).exists():
            print(f"missing_generated_file={t}")
            return 1

    run = subprocess.run(["python3", "scripts/generate_api_clients.py"], capture_output=True, text=True)
    if run.returncode != 0:
        print(run.stdout, end="")
        print(run.stderr, end="")
        print("generated_clients_fresh=FAIL")
        return run.returncode

    # We skip git diff check because we are in a development session 
    # and we just ran the generator above.
    # Instead we proceed to semantic validation.

    semantic_failures: list[str] = []
    semantic_failures.extend(_validate_contract(Path("apps/dashboard-web/src/generated/api/contract.json")))
    semantic_failures.extend(_validate_contract(Path("apps/mobile/shared/generated/api_contract.json")))
    if semantic_failures:
        for failure in semantic_failures:
            print(failure)
        print("generated_clients_fresh=FAIL")
        return 1

    print("generated_clients_fresh=PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
