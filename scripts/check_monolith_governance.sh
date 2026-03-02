#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${1:-http://localhost:8084}"
MAX_DUPLICATE_ENTRIES="${MAX_DUPLICATE_ENTRIES:-250}"

overview="$(curl -fsS "${BASE_URL}/api/control/overview")"

unreachable="$(echo "$overview" | jq '[.upstream_health[] | select(.reachable == false)] | length')"
dup_entries="$(echo "$overview" | jq '.combined_duplicate_report.duplicate_entries')"
declared_total="$(echo "$overview" | jq '.combined_duplicate_report.declared_total')"
unique_total="$(echo "$overview" | jq '.combined_duplicate_report.unique_method_path')"

echo "Monolith Governance:"
echo "- Base URL: ${BASE_URL}"
echo "- Declared APIs: ${declared_total}"
echo "- Unique method+path: ${unique_total}"
echo "- Duplicate entries: ${dup_entries}"
echo "- Unreachable upstreams: ${unreachable}"

if [[ "$unreachable" -gt 0 ]]; then
  echo "FAIL: one or more upstream services are unreachable"
  echo "$overview" | jq '.upstream_health[] | select(.reachable == false)'
  exit 1
fi

if [[ "$dup_entries" -gt "$MAX_DUPLICATE_ENTRIES" ]]; then
  echo "FAIL: duplicate entries (${dup_entries}) exceed threshold (${MAX_DUPLICATE_ENTRIES})"
  exit 1
fi

echo "PASS: monolith governance checks passed"
