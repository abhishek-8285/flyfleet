#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
OUT_DIR="${1:-$ROOT_DIR/tmp}"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
SERVICE_DIR="$ROOT_DIR/services/fleet-core-service-legacy"

mkdir -p "$OUT_DIR"

RAW_FILE="$OUT_DIR/monolith_api_raw_${STAMP}.txt"
INV_FILE="$OUT_DIR/monolith_api_inventory_${STAMP}.csv"
DUP_FILE="$OUT_DIR/monolith_api_duplicates_${STAMP}.csv"
SUMMARY_FILE="$OUT_DIR/monolith_api_summary_${STAMP}.txt"

cd "$ROOT_DIR"

# Collect route declarations from monolith go code.
(
  rg -n --no-heading --glob '!**/*_test.go' --glob '!**/*.md' --glob '!**/*.json' -g '**/*.go' \
    "\\b(r|router)\\.(Get|Post|Put|Patch|Delete|Options|Head)\\(\"[^\"]+\"" "$SERVICE_DIR" || true
  rg -n --no-heading --glob '!**/*_test.go' --glob '!**/*.md' --glob '!**/*.json' -g '**/*.go' \
    "\\b(r|router)\\.Method\\(\"(GET|POST|PUT|PATCH|DELETE|OPTIONS|HEAD)\",\\s*\"[^\"]+\"" "$SERVICE_DIR" || true
) > "$RAW_FILE"

awk '
{
  line=$0

  if (match(line,/(r|router)\.(Get|Post|Put|Patch|Delete|Options|Head)\("[^"]+"/)) {
    tok=substr(line,RSTART,RLENGTH)
    method=tok
    sub(/^.*\./,"",method)
    sub(/\(.*/,"",method)
    method=toupper(method)

    path=tok
    sub(/^.*\.[A-Za-z]+\("/,"",path)
    sub(/"$/, "", path)

    print method "," path
    next
  }

  if (match(line,/(r|router)\.Method\("(GET|POST|PUT|PATCH|DELETE|OPTIONS|HEAD)",\s*"[^"]+"/)) {
    tok=substr(line,RSTART,RLENGTH)

    tmp=tok
    sub(/^.*\.Method\("/,"",tmp)
    method=tmp
    sub(/".*/,"",method)

    path=tok
    sub(/^.*\.Method\("[A-Z]+",\s*"/,"",path)
    sub(/"$/, "", path)

    print method "," path
  }
}
' "$RAW_FILE" | sort > "$OUT_DIR/monolith_api_method_path_${STAMP}.csv"

{
  echo "METHOD,PATH,OCCURRENCES"
  awk -F',' '{ key=$1","$2; c[key]++ } END { for (k in c) print k","c[k] }' "$OUT_DIR/monolith_api_method_path_${STAMP}.csv" | sort
} > "$INV_FILE"

{
  echo "METHOD,PATH,OCCURRENCES"
  awk -F',' 'NR>1 && $3+0>1 { print }' "$INV_FILE"
} > "$DUP_FILE"

declared_total=$(wc -l < "$OUT_DIR/monolith_api_method_path_${STAMP}.csv" | tr -d ' ')
unique_total=$(awk -F',' 'NR>1 { c++ } END { print c+0 }' "$INV_FILE")
duplicate_entries=$(awk -F',' 'NR>1 && $3+0>1 { c++ } END { print c+0 }' "$DUP_FILE")

cat > "$SUMMARY_FILE" <<SUMMARY
Generated: ${STAMP}
Service: services/fleet-core-service-legacy
Declared route declarations: ${declared_total}
Unique METHOD+PATH: ${unique_total}
Duplicate METHOD+PATH entries: ${duplicate_entries}

Files:
- ${INV_FILE}
- ${DUP_FILE}
SUMMARY

echo "$SUMMARY_FILE"
