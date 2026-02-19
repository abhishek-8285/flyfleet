#!/usr/bin/env python3
"""Checks dashboard-web JS bundle size against enforceable budgets.

Enforcement budgets (gzipped):
- Main bundle <= 550 KB
- Total JS bundles <= 1200 KB

Target (non-blocking roadmap):
- Main bundle <= 300 KB
"""

from __future__ import annotations

import gzip
import sys
from pathlib import Path

BUILD_JS_DIR = Path("apps/dashboard-web/build/static/js")
MAIN_BUDGET_KB = 550
TOTAL_BUDGET_KB = 1200


def gzipped_size(path: Path) -> int:
    data = path.read_bytes()
    return len(gzip.compress(data, compresslevel=9))


def main() -> int:
    if not BUILD_JS_DIR.exists():
        print(f"missing_build_dir={BUILD_JS_DIR}")
        print("hint=run `cd apps/dashboard-web && npm run build` first")
        return 2

    files = sorted(p for p in BUILD_JS_DIR.glob("*.js") if p.is_file())
    if not files:
        print("bundle_files=0")
        print("bundle_budget=FAIL")
        return 1

    main_candidates = [p for p in files if p.name.startswith("main.")]
    if not main_candidates:
        print("main_bundle_found=false")
        print("bundle_budget=FAIL")
        return 1

    main_file = main_candidates[0]
    main_gz = gzipped_size(main_file)

    total_gz = 0
    for p in files:
        total_gz += gzipped_size(p)

    main_kb = main_gz / 1024
    total_kb = total_gz / 1024

    print(f"bundle_files={len(files)}")
    print(f"main_bundle_file={main_file.name}")
    print(f"main_bundle_gzip_kb={main_kb:.2f}")
    print(f"total_js_gzip_kb={total_kb:.2f}")
    print(f"main_budget_kb={MAIN_BUDGET_KB}")
    print(f"total_budget_kb={TOTAL_BUDGET_KB}")

    failed = False
    if main_kb > MAIN_BUDGET_KB:
        print("budget_violation=main_bundle")
        failed = True
    if total_kb > TOTAL_BUDGET_KB:
        print("budget_violation=total_js")
        failed = True

    if failed:
        print("bundle_budget=FAIL")
        return 1

    print("bundle_budget=PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
