#!/usr/bin/env python3
import re
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
FRONTEND_DIR = ROOT / "apps" / "dashboard-web" / "src"
BFF_MAIN = ROOT / "services" / "web-bff" / "cmd" / "main.go"


FRONTEND_API_PATTERN = re.compile(r"""(?P<q>["'`])(?P<url>(?:https?://localhost:8081)?/(?:api|bff)[^"'`\s]*)""")
BFF_PREFIX_PATTERN = re.compile(r'registerPrefixProxy\(\s*r\s*,\s*"([^"]+)"')
BFF_DIRECT_PATH_PATTERN = re.compile(r'r\.(?:Get|Post|Put|Delete|Patch|Handle)\(\s*"([^"]+)"')
BFF_MAPPING_PATTERN = re.compile(
    r'registerPrefixProxy\(\s*r\s*,\s*"([^"]+)"\s*,\s*makePrefixProxy\(\s*logger\s*,\s*[^,]+\s*,\s*"([^"]+)"\s*,\s*"([^"]+)"',
    re.MULTILINE,
)

REQUIRED_PREFIXES = {
    "/api/firmware/rollouts",
    "/api/firmware",
    "/api/devices/summary",
    "/api/devices/twins",
    "/api/devices",
}

REQUIRED_MAPPINGS = {
    "/api/firmware/rollouts": "/api/v1/diagnostics/rollouts",
    "/api/firmware": "/api/v1/diagnostics/firmware",
    "/api/devices/summary": "/api/v1/diagnostics/summary",
    "/api/devices/twins": "/api/v1/diagnostics/twins",
    "/api/devices": "/api/v1/diagnostics/devices",
}


def normalize_frontend_path(raw: str) -> str:
    path = re.sub(r"^https?://localhost:8081", "", raw)
    path = path.split("?")[0]
    path = path.split("${")[0]
    if not path.startswith("/"):
        path = "/" + path
    return path


def load_frontend_paths() -> set[str]:
    paths: set[str] = set()
    for file in FRONTEND_DIR.rglob("*"):
        if file.suffix not in {".ts", ".tsx", ".js", ".jsx"}:
            continue
        text = file.read_text(encoding="utf-8", errors="ignore")
        for m in FRONTEND_API_PATTERN.finditer(text):
            path = normalize_frontend_path(m.group("url"))
            if path.startswith(("/api", "/bff")):
                paths.add(path)
    return paths


def load_bff_paths() -> tuple[set[str], set[str], dict[str, tuple[str, str]]]:
    text = BFF_MAIN.read_text(encoding="utf-8", errors="ignore")
    prefixes = set(BFF_PREFIX_PATTERN.findall(text))
    direct_paths = set(
        p for p in BFF_DIRECT_PATH_PATTERN.findall(text) if p.startswith(("/api", "/bff", "/health"))
    )
    mappings: dict[str, tuple[str, str]] = {}
    for registered_prefix, source_prefix, target_prefix in BFF_MAPPING_PATTERN.findall(text):
        mappings[registered_prefix] = (source_prefix, target_prefix)
    return prefixes, direct_paths, mappings


def matches(path: str, prefixes: set[str], direct_paths: set[str]) -> bool:
    if path in direct_paths:
        return True
    return any(path == prefix or path.startswith(prefix + "/") for prefix in prefixes)


def main() -> int:
    frontend_paths = load_frontend_paths()
    frontend_paths = {p for p in frontend_paths if p not in {"/api", "/api/v1"}}
    prefixes, direct_paths, mappings = load_bff_paths()

    missing = sorted(p for p in frontend_paths if not matches(p, prefixes, direct_paths))

    print(f"frontend_api_paths={len(frontend_paths)}")
    print(f"bff_prefix_routes={len(prefixes)}")
    print(f"bff_direct_routes={len(direct_paths)}")

    missing_required_prefixes = sorted(p for p in REQUIRED_PREFIXES if p not in prefixes)
    mapping_failures: list[str] = []
    for registered_prefix, expected_target in REQUIRED_MAPPINGS.items():
        mapping = mappings.get(registered_prefix)
        if mapping is None:
            mapping_failures.append(f"missing_mapping:{registered_prefix}")
            continue
        source_prefix, target_prefix = mapping
        if source_prefix != registered_prefix:
            mapping_failures.append(
                f"source_mismatch:{registered_prefix} source={source_prefix}"
            )
        if target_prefix != expected_target:
            mapping_failures.append(
                f"target_mismatch:{registered_prefix} target={target_prefix} expected={expected_target}"
            )

    if not missing and not missing_required_prefixes and not mapping_failures:
        print("contract_check=PASS")
        return 0

    print("contract_check=FAIL")
    if missing:
        print("missing_paths:")
        for p in missing:
            print(f"  - {p}")
    if missing_required_prefixes:
        print("missing_required_prefixes:")
        for p in missing_required_prefixes:
            print(f"  - {p}")
    if mapping_failures:
        print("mapping_failures:")
        for item in mapping_failures:
            print(f"  - {item}")
    return 1


if __name__ == "__main__":
    sys.exit(main())
