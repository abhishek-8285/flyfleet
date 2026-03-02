#!/usr/bin/env python3
"""Sanity check for trace-id propagation wiring."""

from __future__ import annotations

import re
import sys
from pathlib import Path

RULES = {
    "apps/dashboard-web/src/services/apiClient.ts": [
        r"TRACE_ID_HEADER",
        r"REQUEST_ID_HEADER",
    ],
    "apps/dashboard-web/src/services/api.ts": [
        r"TRACE_ID_HEADER",
        r"REQUEST_ID_HEADER",
    ],
    "apps/dashboard-web/src/index.tsx": [
        r"initializeTraceFetch\(\)",
    ],
    "services/microservices/web-bff/internal/middleware/trace_id.go": [
        r"TraceIDMiddleware",
        r"GetTraceID",
        r"HeaderTraceID",
        r"HeaderRequestID",
        r"traceID := strings\.TrimSpace\(r\.Header\.Get\(HeaderTraceID\)\)",
        r"traceID = strings\.TrimSpace\(r\.Header\.Get\(HeaderRequestID\)\)",
        r"w\.Header\(\)\.Set\(HeaderTraceID, traceID\)",
        r"w\.Header\(\)\.Set\(HeaderRequestID, traceID\)",
    ],
    "services/microservices/web-bff/internal/middleware/cors.go": [
        r"X-Trace-ID",
        r"X-Request-ID",
    ],
    "services/microservices/web-bff/cmd/main.go": [
        r"TraceIDMiddleware",
        r"HeaderTraceID",
        r"HeaderRequestID",
        r"req\.Header\.Set\(bffMiddleware\.HeaderTraceID, traceID\)",
        r"req\.Header\.Set\(bffMiddleware\.HeaderRequestID, traceID\)",
        r"w\.Header\(\)\.Set\(bffMiddleware\.HeaderTraceID, traceID\)",
        r"w\.Header\(\)\.Set\(bffMiddleware\.HeaderRequestID, traceID\)",
    ],
    "services/microservices/web-bff/cmd/prefix_proxy_test.go": [
        r"seenTraceHeader",
        r"X-Trace-ID",
    ],
}


def main() -> int:
    failures = []
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

    print(f"trace_files_checked={checked}")
    if failures:
        print(f"trace_propagation=FAIL ({len(failures)} issues)")
        for item in failures:
            print(item)
        return 1

    print("trace_propagation=PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
