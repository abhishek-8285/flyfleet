#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
OUT_DIR="${1:-$ROOT_DIR/tmp}"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"

mkdir -p "$OUT_DIR"
cd "$ROOT_DIR"

# 1) Run compatibility gate
bash scripts/check_monolith_compat_gate.sh >/dev/null
matrix_summary="$(ls -1t tmp/monolith_compat_matrix_summary_*.txt | head -n1)"
inventory_summary="$(ls -1t tmp/monolith_api_summary_*.txt | head -n1)"

# 2) Run backend deep scan
bash scripts/deep_backend_api_scan.sh "$OUT_DIR" >/dev/null
backend_summary="$(ls -1t "$OUT_DIR"/backend_api_deep_scan_*.txt | head -n1)"

# 3) Extract numbers
matrix_total="$(awk -F': ' '/Total web-bff routes/{print $2}' "$matrix_summary")"
matrix_covered="$(awk -F': ' '/Covered in monolith/{print $2}' "$matrix_summary")"
matrix_missing="$(awk -F': ' '/Missing in monolith/{print $2}' "$matrix_summary")"

mono_declared="$(awk -F': ' '/Declared route declarations/{print $2}' "$inventory_summary")"
mono_unique="$(awk -F': ' '/Unique METHOD\+PATH/{print $2}' "$inventory_summary")"
mono_dups="$(awk -F': ' '/Duplicate METHOD\+PATH entries/{print $2}' "$inventory_summary")"

backend_rest="$(awk -F'=' '/^TOTAL_REST=/{print $2}' "$backend_summary")"
backend_proto_rpc="$(awk -F'=' '/^GLOBAL_PROTO_RPC_TOTAL=/{print $2}' "$backend_summary")"
backend_total="$(awk -F'=' '/^GLOBAL_BACKEND_TOTAL=/{print $NF}' "$backend_summary")"

status_file="$OUT_DIR/monolith_status_bundle_${STAMP}.json"

cat > "$status_file" <<JSON
{
  "generated_at_utc": "$STAMP",
  "compatibility": {
    "web_bff_routes_total": ${matrix_total:-0},
    "web_bff_routes_covered": ${matrix_covered:-0},
    "web_bff_routes_missing": ${matrix_missing:-0},
    "compat_gate_passed": $([[ "${matrix_missing:-1}" == "0" ]] && echo true || echo false)
  },
  "monolith_inventory": {
    "declared_route_declarations": ${mono_declared:-0},
    "unique_method_path": ${mono_unique:-0},
    "duplicate_method_path_entries": ${mono_dups:-0}
  },
  "backend_inventory": {
    "rest_route_declarations": ${backend_rest:-0},
    "global_proto_rpc_total": ${backend_proto_rpc:-0},
    "global_backend_total": ${backend_total:-0}
  },
  "artifacts": {
    "compat_matrix_summary": "$matrix_summary",
    "monolith_inventory_summary": "$inventory_summary",
    "backend_deep_scan_summary": "$backend_summary"
  }
}
JSON

echo "$status_file"
