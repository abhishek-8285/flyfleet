#!/usr/bin/env bash

ROOT="/home/abhishek/Desktop/flyfleet"
OUT="$ROOT/all_apis.json"

echo "[" > "$OUT"

############################
# REST - GO (clean filter)
############################
grep -Rho --exclude-dir=_archive --include=*.go -E 'HandleFunc\(".*"|Path\(".*"' "$ROOT/services" \
| sed -E 's/.*\("([^"]+)".*/\1/' \
| grep '^/' \
| sort -u \
| while read route; do
  echo "{\"type\":\"REST\",\"language\":\"Go\",\"route\":\"$route\"}," >> "$OUT"
done

############################
# REST - PYTHON
############################
grep -Rho --exclude-dir=_archive --include=*.py -E '@app\.(get|post|put|patch|delete)\(".*"' "$ROOT/services" \
| sed -E 's/.*\("([^"]+)".*/\1/' \
| grep '^/' \
| sort -u \
| while read route; do
  echo "{\"type\":\"REST\",\"language\":\"Python\",\"route\":\"$route\"}," >> "$OUT"
done

############################
# RPC - correct parsing
############################
find "$ROOT/proto" -name "*.proto" -not -path "*/_archive/*" | while read file; do
  service=$(grep -m1 "^service " "$file" | sed -E 's/service ([A-Za-z0-9]+).*/\1/')

  grep "^  rpc " "$file" | sed -E 's/rpc ([A-Za-z0-9]+).*/\1/' | while read method; do
    echo "{\"type\":\"RPC\",\"service\":\"$service\",\"method\":\"$method\"}," >> "$OUT"
  done
done

############################
# finalize json
############################
sed -i '$ s/,$//' "$OUT"
echo "]" >> "$OUT"

############################
# counts
############################
echo ""
echo "========= API COUNT ========="
REST_COUNT=$(grep -c '"type":"REST"' "$OUT")
RPC_COUNT=$(grep -c '"type":"RPC"' "$OUT")
echo "Total REST APIs: $REST_COUNT"
echo "Total RPC APIs: $RPC_COUNT"
echo "TOTAL APIs: $((REST_COUNT + RPC_COUNT))"
