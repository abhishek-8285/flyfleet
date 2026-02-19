#!/usr/bin/env python3
"""Checks observability foundation artifacts for SLO dashboards and alerts."""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

import yaml

RULES = {
    "docs/operations/SLO_OPERATIONS.md": [r"SLO Catalog", r"Required Dashboards", r"Error Budget Policy"],
    "infrastructure/monitoring/slo-alert-rules.yml": [r"FlyFleetAuthAvailabilityCritical", r"FlyFleetOTASuccessCritical"],
    "infrastructure/monitoring/slo-journeys-dashboard.json": [r"FlyFleet SLO Journeys", r"OTA Success Rate by Ring"],
}

REQUIRED_ALERTS = {
    "FlyFleetAuthAvailabilityCritical",
    "FlyFleetCriticalAPILatencyCritical",
    "FlyFleetTelemetryFreshnessCritical",
    "FlyFleetCommandReliabilityCritical",
    "FlyFleetMobileSyncReliabilityCritical",
    "FlyFleetOTASuccessCritical",
}

REQUIRED_PANELS = {
    "Login Success Rate (30m)",
    "Dashboard P95 Usable Time",
    "Critical API P95 Latency",
    "Telemetry Freshness P95",
    "Command Success Rate (30m)",
    "Command Ack P95",
    "Mobile Sync Apply Success Rate",
    "OTA Success Rate by Ring",
}


def validate_alert_rules(path: Path) -> list[str]:
    failures: list[str] = []
    try:
        payload = yaml.safe_load(path.read_text(encoding="utf-8"))
    except Exception as exc:
        return [f"observability_alert_rules_parse_failed: {exc}"]

    groups = payload.get("groups") if isinstance(payload, dict) else None
    if not isinstance(groups, list):
        return ["observability_alert_rules_groups_missing"]

    seen: set[str] = set()
    for group in groups:
        if not isinstance(group, dict):
            continue
        rules = group.get("rules")
        if not isinstance(rules, list):
            continue
        for rule in rules:
            if not isinstance(rule, dict):
                continue
            alert_name = rule.get("alert")
            if isinstance(alert_name, str):
                seen.add(alert_name)
            labels = rule.get("labels")
            if not isinstance(labels, dict):
                failures.append(f"observability_alert_missing_labels:{alert_name}")
                continue
            if labels.get("severity") not in {"warning", "critical"}:
                failures.append(f"observability_alert_invalid_severity:{alert_name}")
            annotations = rule.get("annotations")
            if not isinstance(annotations, dict) or not annotations.get("runbook_url"):
                failures.append(f"observability_alert_missing_runbook:{alert_name}")

    missing = REQUIRED_ALERTS - seen
    for alert in sorted(missing):
        failures.append(f"observability_alert_missing:{alert}")
    return failures


def validate_dashboard(path: Path) -> list[str]:
    failures: list[str] = []
    try:
        payload = json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        return [f"observability_dashboard_parse_failed: {exc}"]

    dashboard = payload.get("dashboard")
    if not isinstance(dashboard, dict):
        return ["observability_dashboard_root_missing"]
    if dashboard.get("title") != "FlyFleet SLO Journeys":
        failures.append("observability_dashboard_title_invalid")

    panels = dashboard.get("panels")
    if not isinstance(panels, list):
        return failures + ["observability_dashboard_panels_missing"]

    titles = {str(panel.get("title")) for panel in panels if isinstance(panel, dict)}
    missing_panels = REQUIRED_PANELS - titles
    for panel in sorted(missing_panels):
        failures.append(f"observability_dashboard_panel_missing:{panel}")

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

    alert_rules = Path("infrastructure/monitoring/slo-alert-rules.yml")
    if alert_rules.exists():
        failures.extend(validate_alert_rules(alert_rules))
    dashboard = Path("infrastructure/monitoring/slo-journeys-dashboard.json")
    if dashboard.exists():
        failures.extend(validate_dashboard(dashboard))

    print(f"observability_files_checked={checked}")
    if failures:
        print(f"observability_foundation=FAIL ({len(failures)} issues)")
        for failure in failures:
            print(failure)
        return 1

    print("observability_foundation=PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
