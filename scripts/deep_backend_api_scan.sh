#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
OUT_DIR="${1:-$ROOT_DIR/tmp}"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
OUT_FILE="$OUT_DIR/backend_api_deep_scan_${STAMP}.txt"

mkdir -p "$OUT_DIR"

cd "$ROOT_DIR"

mapfile -t svcs < <(
  find services -mindepth 1 -maxdepth 1 -type d \
    ! -name "_archive" \
    ! -name "microservices" \
    ! -name "dummy*" \
    ! -name "*mock*" \
    | sort
)

{
  echo "Generated: ${STAMP}"
  echo
  echo "Rules: exclude services/_archive, services/microservices, services/dummy*, any *mock* paths; ignore .md/.json sources for counting"
  echo
  echo "SERVICE,REST_ROUTE_DECLARATIONS,GRPC_RPCS"

  rest_total=0
  grpc_total=0

  for s in "${svcs[@]}"; do
    rest1=$( (rg -n --no-heading --glob '!**/*_test.go' --glob '!**/*.md' --glob '!**/*.json' -g '**/*.go' "\\b(r|router)\\.(Get|Post|Put|Patch|Delete|Options|Head)\\(\"[^\"]+\"" "$s" || true) | wc -l | tr -d ' ')
    rest2=$( (rg -n --no-heading --glob '!**/*_test.go' --glob '!**/*.md' --glob '!**/*.json' -g '**/*.go' "\\b(r|router)\\.Method\\(\"(GET|POST|PUT|PATCH|DELETE|OPTIONS|HEAD)\",\\s*\"[^\"]+\"" "$s" || true) | wc -l | tr -d ' ')
    rest=$((rest1 + rest2))

    grpc=$( (rg -n --no-heading --glob '!**/*.md' --glob '!**/*.json' -g '**/*.proto' "^\\s*rpc\\s+[A-Za-z0-9_]+" "$s" || true) | wc -l | tr -d ' ')

    printf "%s,%s,%s\n" "${s#services/}" "$rest" "$grpc"

    rest_total=$((rest_total + rest))
    grpc_total=$((grpc_total + grpc))
  done

  echo "TOTAL_REST=${rest_total}"
  echo "TOTAL_GRPC=${grpc_total}"
  echo "TOTAL_APIS=${rest_total}+${grpc_total}=$((rest_total + grpc_total))"
  echo

  proto_rpc_total=$( (find proto -type f -name '*.proto' -print0 | xargs -0 rg -n '^\s*rpc\s+[A-Za-z0-9_]+' || true) | wc -l | tr -d ' ')
  echo "GLOBAL_PROTO_RPC_TOTAL=${proto_rpc_total}"
  echo "GLOBAL_BACKEND_TOTAL=TOTAL_REST(${rest_total})+GLOBAL_PROTO_RPC_TOTAL(${proto_rpc_total})=$((rest_total + proto_rpc_total))"
  echo

  echo "PROTO_FILE,RPC_COUNT"
  for f in $(find proto -type f -name '*.proto' | sort); do
    c=$(rg -n '^\s*rpc\s+[A-Za-z0-9_]+' "$f" | wc -l | tr -d ' ')
    printf "%s,%s\n" "$f" "$c"
  done
} > "$OUT_FILE"

echo "$OUT_FILE"
