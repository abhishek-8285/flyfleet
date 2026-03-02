#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

echo "[1/4] web-bff rewrite parity"
python3 scripts/check_web_bff_rewrite_parity.py

echo "[2/4] public route parity"
python3 scripts/check_bff_public_route_parity.py

echo "[3/4] compatibility matrix"
python3 scripts/generate_monolith_compat_matrix.py
latest_summary="$(ls -1t tmp/monolith_compat_matrix_summary_*.txt | head -n1)"
missing="$(awk -F': ' '/Missing in monolith/{print $2}' "$latest_summary")"
if [[ "${missing:-1}" != "0" ]]; then
  echo "FAIL: compatibility matrix has missing mappings"
  cat "$latest_summary"
  exit 1
fi

echo "[4/4] monolith inventory"
bash scripts/monolith_api_inventory.sh "$ROOT_DIR/tmp" >/dev/null
latest_inventory="$(ls -1t tmp/monolith_api_summary_*.txt | head -n1)"

echo "PASS: monolith compatibility gate"
echo "- matrix: $latest_summary"
echo "- inventory: $latest_inventory"
