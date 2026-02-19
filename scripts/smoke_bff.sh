#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${1:-http://localhost:8081}"
ALLOW_404="${ALLOW_404:-0}"

echo "# BFF smoke"
echo "base_url=$BASE_URL"

check() {
  local name="$1"
  local method="$2"
  local path="$3"
  local code

  code=$(curl -sS -o /tmp/bff_smoke_body.txt -w "%{http_code}" -X "$method" "$BASE_URL$path" || true)
  printf '%-24s %-6s %-48s -> %s\n' "$name" "$method" "$path" "$code"

  if [[ "$code" == "000" ]]; then
    return 1
  fi
  if [[ "$ALLOW_404" != "1" && "$code" == "404" ]]; then
    return 1
  fi
  return 0
}

fail=0
check "health" GET "/health" || fail=1
check "security" GET "/api/v1/security/sessions" || fail=1
check "governance" GET "/api/v1/admin/governance/requests" || fail=1
check "dr" GET "/api/v1/admin/dr/backups" || fail=1
check "reports" GET "/api/v1/admin/reports/schedules" || fail=1
check "settings" GET "/api/v1/admin/settings/platform" || fail=1
check "maintenance" GET "/api/v1/admin/work-orders" || fail=1
check "billing" GET "/api/v1/admin/billing/invoices" || fail=1
check "notifications" GET "/api/v1/admin/notifications/logs" || fail=1
check "search" GET "/api/search?q=test" || fail=1
check "timeline" GET "/api/timeline" || fail=1
check "devices-summary" GET "/api/devices/summary" || fail=1
check "devices-twins" GET "/api/devices/twins" || fail=1

if [[ "$fail" -ne 0 ]]; then
  echo "smoke_status=FAIL"
  exit 1
fi

echo "smoke_status=PASS"
