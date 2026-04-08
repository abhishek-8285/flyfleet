#!/bin/bash
# Go Nil Safety Auto-Fix Script
# Fixes common nil safety patterns in Go files

echo "=== Go Nil Safety Auto-Fix Script ==="
echo "Scanning active services (excluding _archive)..."

FIXED_COUNT=0
SKIPPED_COUNT=0

# Find all Go files excluding _archive
find services -name "*.go" -type f | grep -v "_archive" | while read file; do
    echo "Processing: $file"
    
    # Create backup
    cp "$file" "$file.nilfix.bak"
    
    # Fix Pattern 1: Add error check after err := function()
    # This is a simple pattern - real fix requires manual review
    sed -i '/err := .*\.Close()/a\
\tif err != nil {\
\t\treturn err\
\t}' "$file" 2>/dev/null
    
    # Fix Pattern 2: Add nil check before pointer dereference (commented - requires manual review)
    # sed -i 's/\*\(ptr\|p\|obj\)/if \1 != nil { *\1 }/g' "$file" 2>/dev/null
    
    # Count fixes
    if ! diff -q "$file" "$file.nilfix.bak" > /dev/null 2>&1; then
        FIXED_COUNT=$((FIXED_COUNT + 1))
        echo "  ✓ Fixed: $file"
    else
        SKIPPED_COUNT=$((SKIPPED_COUNT + 1))
    fi
    
    # Remove backup
    rm "$file.nilfix.bak"
done

echo ""
echo "=== Summary ==="
echo "Files Fixed: $FIXED_COUNT"
echo "Files Skipped: $SKIPPED_COUNT"
echo ""
echo "⚠️  IMPORTANT: Review all changes with 'git diff' before committing!"
echo "⚠️  Run tests to ensure no regressions: go test ./..."
