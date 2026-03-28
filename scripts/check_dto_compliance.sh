#!/bin/bash

# DTO Compliance Checker
# Fails if any handler decodes directly into models package

set -e

# Find all handler files in fleet-core
HANDLER_FILES=$(find services/fleet-core/internal/domains -name "*handler*.go" -type f 2>/dev/null)

VIOLATIONS=0

echo "Checking DTO compliance in handlers..."

for file in $HANDLER_FILES; do
    # Look for patterns: var something models.X followed by Decode(&something)
    # Using grep to find direct model decoding
    if grep -q "json.NewDecoder.*models\." "$file" || grep -q "ShouldBindJSON.*models\." "$file"; then
        echo "❌ VIOLATION: $file decodes directly into models package"
        VIOLATIONS=$((VIOLATIONS+1))
        # Show the offending lines
        echo "   Offending lines:"
        grep -n "json.NewDecoder.*models\\.\|ShouldBindJSON.*models\\." "$file" | sed 's/^/   /'
    fi
done

if [ $VIOLATIONS -gt 0 ]; then
    echo ""
    echo "FAILED: $VIOLATIONS handler(s) violate DTO pattern"
    echo "Fix: Create DTO types and decode into DTOs, then map to models."
    exit 1
fi

echo "✅ All handlers comply with DTO pattern"
exit 0
