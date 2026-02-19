#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${1:-http://localhost:8081}"
MAX_MS="${2:-300}"

# Endpoint list focuses on critical user journeys.
ENDPOINTS=(
  "/health"
  "/api/v1/search?q=test"
  "/api/v1/search/facets"
  "/api/v1/timeline"
  "/api/v1/timeline/stats"
)

echo "# API latency smoke"
echo "base_url=$BASE_URL"
echo "max_ms=$MAX_MS"

fail=0
samples=0
sum_ms=0

for path in "${ENDPOINTS[@]}"; do
  out=$(curl -sS -o /tmp/api_latency_body.txt -w "%{http_code} %{time_total}" "$BASE_URL$path" || true)
  code=$(awk '{print $1}' <<<"$out")
  secs=$(awk '{print $2}' <<<"$out")

  if [[ -z "$secs" ]]; then
    secs="999"
  fi

  ms=$(awk -v s="$secs" 'BEGIN { printf("%.0f", s * 1000) }')
  printf '%-42s code=%-4s latency_ms=%s\n' "$path" "$code" "$ms"

  # Ignore auth status but fail if unreachable or over latency budget.
  if [[ "$code" == "000" ]]; then
    fail=1
  fi
  if (( ms > MAX_MS )); then
    fail=1
  fi

  samples=$((samples + 1))
  sum_ms=$((sum_ms + ms))
done

avg_ms=$((sum_ms / samples))
echo "samples=$samples"
echo "avg_ms=$avg_ms"

if [[ "$fail" -ne 0 ]]; then
  echo "api_latency_smoke=FAIL"
  exit 1
fi

echo "api_latency_smoke=PASS"
