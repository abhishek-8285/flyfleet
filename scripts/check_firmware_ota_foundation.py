#!/usr/bin/env python3
"""Checks firmware OTA foundation artifacts and validator."""

from __future__ import annotations

import re
import subprocess
import sys
from pathlib import Path

import yaml

RULES = {
    "docs/firmware/OTA_SPEC.md": [r"Firmware OTA Specification", r"Device OTA State Machine", r"Rollback Strategy"],
    "docs/firmware/ota-manifest.schema.json": [r"FlyFleet OTA Manifest", r"max_failure_rate"],
    "firmware/releases/sample/ota-manifest.json": [r"firmware_version", r"rollout_policy"],
    "firmware/releases/sample/ota-rollout-scenarios.json": [r"healthy_progressive_rollout", r"pause_and_rollback"],
    "firmware/keys/trusted_keys.json": [r"flyfleet-sample-ed25519-2026q1", r"public_key_path"],
    "scripts/firmware_ota_validator.py": [r"ota_manifest_validation=PASS", r"cryptographic_signature_verification=PASS"],
    "scripts/ota_rollout_orchestrator.py": [r"ota_rollout_orchestration_simulation=PASS", r"pause_and_rollback"],
    "scripts/runtime_integration_gate.sh": [
        r"rollout_note_events_all_note",
        r"rollout_note_events_canonical_code",
        r"rollout_transition_events_all_transition",
        r"rollout_transition_events_canonical_code",
        r"rollout_created_events_all_created",
        r"rollout_created_events_canonical_code",
        r"rollout_invalid_event_filter_code",
        r"rollout_invalid_limit_code",
        r"rollout_limit_max_boundary_code",
        r"rollout_limit_too_large_code",
        r"rollout_invalid_offset_code",
        r"rollout_invalid_limit_format_code",
        r"rollout_invalid_offset_format_code",
        r"ota_orchestration_runtime=PASS",
    ],
}

EXPECTED_EVENT_FILTER_ENUM = {
    "note",
    "transition",
    "created",
    "rollout.note",
    "rollout.transition",
    "rollout.created",
}

EXPECTED_EVENT_RESPONSE_ENUM = {"rollout.note", "rollout.transition", "rollout.created"}


def validate_rollout_events_openapi(path: Path) -> list[str]:
    failures: list[str] = []
    try:
        spec = yaml.safe_load(path.read_text(encoding="utf-8"))
    except Exception as exc:
        return [f"openapi_parse_failed: {path} ({exc})"]

    paths = (spec or {}).get("paths")
    if not isinstance(paths, dict):
        return [f"openapi_paths_missing: {path}"]

    route = paths.get("/firmware/rollouts/{id}/events")
    if not isinstance(route, dict):
        return [f"openapi_route_missing: {path} /firmware/rollouts/{{id}}/events"]

    get_op = route.get("get")
    if not isinstance(get_op, dict):
        return [f"openapi_get_missing: {path} /firmware/rollouts/{{id}}/events"]

    params = get_op.get("parameters")
    if not isinstance(params, list):
        return [f"openapi_parameters_missing: {path} rollout events get"]

    by_key: dict[tuple[str, str], dict] = {}
    for p in params:
        if not isinstance(p, dict):
            continue
        name = p.get("name")
        location = p.get("in")
        if isinstance(name, str) and isinstance(location, str):
            by_key[(name, location)] = p

    limit = by_key.get(("limit", "query"))
    if not isinstance(limit, dict):
        failures.append(f"openapi_limit_param_missing: {path}")
    else:
        schema = limit.get("schema")
        if not isinstance(schema, dict):
            failures.append(f"openapi_limit_schema_missing: {path}")
        else:
            if schema.get("minimum") != 1:
                failures.append(f"openapi_limit_min_invalid: {path}")
            if schema.get("maximum") != 500:
                failures.append(f"openapi_limit_max_invalid: {path}")
            if schema.get("default") != 50:
                failures.append(f"openapi_limit_default_invalid: {path}")

    offset = by_key.get(("offset", "query"))
    if not isinstance(offset, dict):
        failures.append(f"openapi_offset_param_missing: {path}")
    else:
        schema = offset.get("schema")
        if not isinstance(schema, dict):
            failures.append(f"openapi_offset_schema_missing: {path}")
        else:
            if schema.get("minimum") != 0:
                failures.append(f"openapi_offset_min_invalid: {path}")
            if schema.get("default") != 0:
                failures.append(f"openapi_offset_default_invalid: {path}")

    event_type = by_key.get(("event_type", "query"))
    if not isinstance(event_type, dict):
        failures.append(f"openapi_event_type_param_missing: {path}")
    else:
        schema = event_type.get("schema")
        if not isinstance(schema, dict):
            failures.append(f"openapi_event_type_schema_missing: {path}")
        else:
            enum = schema.get("enum")
            if not isinstance(enum, list):
                failures.append(f"openapi_event_type_enum_missing: {path}")
            elif set(enum) != EXPECTED_EVENT_FILTER_ENUM:
                failures.append(f"openapi_event_type_enum_invalid: {path}")

    responses = get_op.get("responses")
    if not isinstance(responses, dict):
        failures.append(f"openapi_responses_missing: {path}")
        return failures

    if "400" not in responses and 400 not in responses:
        failures.append(f"openapi_400_response_missing: {path}")

    r200 = responses.get("200")
    if not isinstance(r200, dict):
        failures.append(f"openapi_200_response_missing: {path}")
        return failures

    try:
        enum = (
            r200["content"]["application/json"]["schema"]["items"]["properties"]["event_type"]["enum"]  # type: ignore[index]
        )
    except Exception:
        failures.append(f"openapi_response_event_type_enum_missing: {path}")
        return failures

    if not isinstance(enum, list) or set(enum) != EXPECTED_EVENT_RESPONSE_ENUM:
        failures.append(f"openapi_response_event_type_enum_invalid: {path}")

    return failures


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

    failures.extend(validate_rollout_events_openapi(Path("docs/api/openapi.yaml")))

    print(f"ota_files_checked={checked}")

    run = subprocess.run(["python3", "scripts/firmware_ota_validator.py"], capture_output=True, text=True)
    if run.returncode != 0:
        print(run.stdout, end="")
        print(run.stderr, end="")
        failures.append("validator_execution_failed")

    orchestrator = subprocess.run(["python3", "scripts/ota_rollout_orchestrator.py"], capture_output=True, text=True)
    if orchestrator.returncode != 0:
        print(orchestrator.stdout, end="")
        print(orchestrator.stderr, end="")
        failures.append("orchestrator_execution_failed")
    elif "ota_rollout_orchestration_simulation=PASS" not in orchestrator.stdout:
        failures.append("orchestrator_missing_pass_marker")

    if failures:
        print(f"firmware_ota_foundation=FAIL ({len(failures)} issues)")
        for f in failures:
            print(f)
        return 1

    print("firmware_ota_foundation=PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
