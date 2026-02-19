#!/usr/bin/env python3
"""Checks shared design token and cross-surface design contract artifacts."""

from __future__ import annotations

import json
import sys
from pathlib import Path


def require_file(path: str) -> Path:
    p = Path(path)
    if not p.exists():
        print(f"design_system=FAIL missing_file={path}")
        raise SystemExit(1)
    return p


def check_token_files() -> None:
    common = json.loads(
        require_file("packages/common/design/design_tokens.json").read_text(
            encoding="utf-8"
        )
    )
    mobile = json.loads(
        require_file("apps/mobile/shared/design/design_tokens.json").read_text(
            encoding="utf-8"
        )
    )

    for key in ["version", "color", "spacing", "radius"]:
        if key == "version":
            if common.get("version") != mobile.get("version"):
                print("design_system=FAIL token_version_mismatch")
                raise SystemExit(1)
            continue

        common_family = common.get(key) or {}
        mobile_family = (mobile.get("tokens") or {}).get(key) or {}
        if not common_family or not mobile_family:
            print(f"design_system=FAIL missing_token_family={key}")
            raise SystemExit(1)

    print("design_system_tokens=PASS")


def check_web_binding() -> None:
    css = require_file("apps/dashboard-web/src/styles/design-tokens.css").read_text(
        encoding="utf-8"
    )
    index_tsx = require_file("apps/dashboard-web/src/index.tsx").read_text(
        encoding="utf-8"
    )
    required_vars = [
        "--ff-color-primary",
        "--ff-space-md",
        "--ff-radius-md",
    ]
    for token in required_vars:
        if token not in css:
            print(f"design_system=FAIL missing_css_variable={token}")
            raise SystemExit(1)
    if "./styles/design-tokens.css" not in index_tsx:
        print("design_system=FAIL missing_web_token_import")
        raise SystemExit(1)
    print("design_system_web_binding=PASS")


def check_contract_doc() -> None:
    doc = require_file("docs/architecture/design-system-contract.md").read_text(
        encoding="utf-8"
    )
    for section in ["Source of Truth", "Required Token Families", "Enforcement"]:
        if section not in doc:
            print(f"design_system=FAIL missing_doc_section={section}")
            raise SystemExit(1)
    print("design_system_docs=PASS")


def main() -> int:
    check_token_files()
    check_web_binding()
    check_contract_doc()
    print("design_system=PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
