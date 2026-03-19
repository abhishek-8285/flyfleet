#!/bin/bash
set -e

# Base directory for the monolith
BASE_DIR="/home/abhishek/Desktop/flyfleet/services/fleet-core-service-legacy"
DOMAINS_DIR="$BASE_DIR/internal/domains"

# Domains with legacy versions to move
LEGACY_DOMAINS=("admin_support" "billing" "customer" "diagnostics" "integrations" "reports")

# All V2 domains to rename (the prefix part)
V2_DOMAINS=(
    "admin_support"
    "alert"
    "alerts"
    "auth"
    "billing"
    "compliance"
    "customer"
    "diagnostics"
    "india"
    "integrations"
    "reports"
    "tracking"
)

echo "Shadowing legacy domains..."
for domain in "${LEGACY_DOMAINS[@]}"; do
    if [ -d "$DOMAINS_DIR/$domain" ]; then
        mv "$DOMAINS_DIR/$domain" "$DOMAINS_DIR/${domain}_v1"
        echo "Moved $domain to ${domain}_v1"
    fi
done

echo "Renaming V2 domains to canonical names..."
for domain in "${V2_DOMAINS[@]}"; do
    if [ -d "$DOMAINS_DIR/${domain}_v2" ]; then
        mv "$DOMAINS_DIR/${domain}_v2" "$DOMAINS_DIR/$domain"
        echo "Renamed ${domain}_v2 to $domain"
    fi
done

echo "Performing global search and replace for imports and packages..."
# We use a broad search across the entire legacy service directory
SEARCH_DIR="$BASE_DIR"

for domain in "${V2_DOMAINS[@]}"; do
    old_pkg="${domain}_v2"
    new_pkg="${domain}"
    
    echo "Updating references for $domain ($old_pkg -> $new_pkg)..."
    
    # 1. Update internal package declarations (e.g., "package auth_v2" -> "package auth")
    # Using find/sed for better cross-platform compatibility and to avoid arg list limits
    if [ -d "$DOMAINS_DIR/$domain" ]; then
        find "$DOMAINS_DIR/$domain" -type f -name "*.go" -exec sed -i "s/package $old_pkg/package $new_pkg/g" {} +
    fi
    
    # 2. Update all imports project-wide
    # Pattern: "internal/domains/auth_v2" -> "internal/domains/auth"
    # We search the whole monolith directory
    find "$SEARCH_DIR" -type f \( -name "*.go" -o -name "go.mod" -o -name "Makefile" \) -exec sed -i "s|internal/domains/$old_pkg|internal/domains/$new_pkg|g" {} +
done

echo "Refactor complete."
