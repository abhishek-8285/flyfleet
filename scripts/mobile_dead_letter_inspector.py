#!/usr/bin/env python3
"""Inspect mobile GPS dead-letter state from Android shared_prefs XML."""

from __future__ import annotations

import argparse
import json
import sys
import xml.etree.ElementTree as ET
from pathlib import Path


def load_pref_string(xml_path: Path, key: str) -> str | None:
    tree = ET.parse(xml_path)
    root = tree.getroot()
    for child in root.findall("string"):
        if child.attrib.get("name") == key:
            return child.text or ""
    return None


def inspect(xml_path: Path) -> int:
    if not xml_path.exists():
        print(f"missing_file={xml_path}")
        return 1

    dead_ids_raw = load_pref_string(xml_path, "gps_dead_letter_ids")
    dead_events_raw = load_pref_string(xml_path, "gps_dead_letter_events")

    if dead_ids_raw is None and dead_events_raw is None:
        print("mobile_dead_letter_state=NOT_FOUND")
        print("hint=No gps_dead_letter_* keys found in this shared_prefs file.")
        return 0

    dead_ids = []
    dead_events = []

    if dead_ids_raw:
        try:
            dead_ids = json.loads(dead_ids_raw)
        except Exception:
            print("dead_letter_ids_parse=FAIL")
            print(f"dead_letter_ids_raw={dead_ids_raw}")
            return 1

    if dead_events_raw:
        try:
            dead_events = json.loads(dead_events_raw)
        except Exception:
            print("dead_letter_events_parse=FAIL")
            print(f"dead_letter_events_raw={dead_events_raw}")
            return 1

    print("mobile_dead_letter_state=FOUND")
    print(f"dead_letter_id_count={len(dead_ids)}")
    print(f"dead_letter_event_count={len(dead_events)}")

    reason_counts: dict[str, int] = {}
    for event in dead_events:
        reason = str(event.get("reason", "unknown"))
        reason_counts[reason] = reason_counts.get(reason, 0) + 1

    for reason, count in sorted(reason_counts.items(), key=lambda x: (-x[1], x[0])):
        print(f"reason_count[{reason}]={count}")

    if dead_events:
        latest = dead_events[-1]
        print("latest_event=" + json.dumps(latest, ensure_ascii=True))

    print("mobile_dead_letter_inspection=PASS")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "prefs_xml",
        type=Path,
        help="Path to shared_prefs XML (example: /data/data/com.flyfleet.driver/shared_prefs/flyfleet.xml)",
    )
    args = parser.parse_args()
    return inspect(args.prefs_xml)


if __name__ == "__main__":
    sys.exit(main())
