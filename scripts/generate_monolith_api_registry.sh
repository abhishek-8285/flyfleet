#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${1:-http://localhost:8084}"
OUT_DIR="${2:-/home/abhishek/Desktop/flyfleet/tmp}"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"

mkdir -p "$OUT_DIR"

COUNTS_FILE="$OUT_DIR/monolith_api_counts_${STAMP}.json"
SPEC_FILE="$OUT_DIR/monolith_openapi_combined_${STAMP}.json"
INDEX_FILE="$OUT_DIR/monolith_api_registry_latest.json"

curl -fsS "${BASE_URL}/api/control/apis/counts" > "$COUNTS_FILE"
curl -fsS "${BASE_URL}/swagger/openapi.combined.json" > "$SPEC_FILE"

cat > "$INDEX_FILE" <<EOF
{
  "generated_at_utc": "${STAMP}",
  "base_url": "${BASE_URL}",
  "counts_file": "${COUNTS_FILE}",
  "combined_openapi_file": "${SPEC_FILE}"
}
EOF

echo "Wrote:"
echo "- $COUNTS_FILE"
echo "- $SPEC_FILE"
echo "- $INDEX_FILE"
