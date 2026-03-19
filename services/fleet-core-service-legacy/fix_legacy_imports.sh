#!/bin/bash
set -e

BASE_DIR="/home/abhishek/Desktop/flyfleet/services/fleet-core-service-legacy"
DOMAINS_DIR="$BASE_DIR/internal/domains"

# Mapping: domain -> unique symbol
declare -A SYMBOLS=(
    ["admin_support"]="SearchResponse"
    ["billing"]="BillingService"
    ["customer"]="ShipmentListFilter"
    ["diagnostics"]="DeviceTelemetry"
    ["integrations"]="EventBus"
    ["reports"]="StreamingCSVWriter"
)

echo "Updating package declarations in _v1 folders..."
for domain in "${!SYMBOLS[@]}"; do
    if [ -d "$DOMAINS_DIR/${domain}_v1" ]; then
        find "$DOMAINS_DIR/${domain}_v1" -type f -name "*.go" -exec sed -i "s/package $domain/package ${domain}_v1/g" {} +
        echo "Updated package declarations for ${domain}_v1"
    fi
done

echo "Updating imports based on symbol detection..."
# Any file that uses the symbol and imports the canonical path should be redirected to _v1
for domain in "${!SYMBOLS[@]}"; do
    symbol="${SYMBOLS[$domain]}"
    echo "Processing $domain (Symbol: $symbol)..."
    
    # 1. Find files that import the canonical path
    # 2. Check if they contain the unique legacy symbol
    # 3. If yes, replace the import path with _v1
    
    # We use a combined grep to find potential files
    grep -rl "internal/domains/$domain" "$BASE_DIR" | while read -r file; do
        if grep -q "$symbol" "$file"; then
            # But wait, we must NOT update the ones in the canonical domain itself (the new one)
            # unless it's a legacy test or something. Usually, new domains (formerly v2)
            # should NOT have these symbols.
            if [[ "$file" != *"/internal/domains/$domain/"* ]]; then
                sed -i "s|internal/domains/$domain|internal/domains/${domain}_v1|g" "$file"
                echo "  Redirected legacy import in $file"
            fi
        fi
    done
done

# Special case: fuel_service.go used TopicFuel which is in integrations_v1
echo "Handling special case: integrations -> integrations_v1 in fuel_service..."
sed -i "s|internal/domains/integrations_v1|internal/domains/integrations_v1_v1|g" "$BASE_DIR/internal/domains/fuel/service/fuel_service.go" || true

echo "Legacy repair complete."
