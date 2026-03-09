#!/bin/bash
#
# FlyFleet Invoice System - Quick Test
# Run this to verify everything works
#

set -e

BASE_URL="${BASE_URL:-http://localhost:8084}"
TOKEN="${TOKEN:-test-token}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=========================================="
echo "FlyFleet Invoice System - Quick Test"
echo "=========================================="
echo ""

# Test 1: Check if server is running
echo -n "1. Checking if server is running... "
if curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health" | grep -q "200"; then
    echo -e "${GREEN}✓ PASS${NC}"
else
    echo -e "${RED}✗ FAIL${NC} (Is the server running?)"
    exit 1
fi

# Test 2: Check Chrome is installed
echo -n "2. Checking Chrome/Chromium... "
if command -v chromium-browser &> /dev/null || command -v chromium &> /dev/null || command -v google-chrome &> /dev/null; then
    echo -e "${GREEN}✓ PASS${NC}"
else
    echo -e "${YELLOW}⚠ WARNING${NC} (Chrome not found - PDF generation will fail)"
fi

# Test 3: Check directories exist
echo -n "3. Checking invoice directories... "
if [ -d "/opt/flyfleet/data/invoices" ] || [ -d "./data/invoices" ]; then
    echo -e "${GREEN}✓ PASS${NC}"
else
    echo -e "${YELLOW}⚠ WARNING${NC} (Directory doesn't exist - will be created on first use)"
fi

# Test 4: Health check endpoint
echo -n "4. Testing billing health endpoint... "
HEALTH=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/v1/billing/health")
if echo "$HEALTH" | grep -q "healthy"; then
    echo -e "${GREEN}✓ PASS${NC}"
    echo "   Response: $(echo $HEALTH | jq -r '.status')"
else
    echo -e "${RED}✗ FAIL${NC}"
    echo "   Response: $HEALTH"
fi

# Test 5: Get billing stats
echo -n "5. Testing billing stats endpoint... "
STATS=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/v1/billing/stats")
if echo "$STATS" | grep -q "storage"; then
    echo -e "${GREEN}✓ PASS${NC}"
    echo "   Cache entries: $(echo $STATS | jq -r '.cache.entry_count')"
    echo "   Storage files: $(echo $STATS | jq -r '.storage.file_count')"
else
    echo -e "${YELLOW}⚠ WARNING${NC} (No stats available yet)"
fi

# Test 6: List invoices
echo -n "6. Testing list invoices endpoint... "
INVOICES=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/v1/billing/invoices")
if echo "$INVOICES" | grep -q "data"; then
    echo -e "${GREEN}✓ PASS${NC}"
    COUNT=$(echo $INVOICES | jq -r '.total // 0')
    echo "   Total invoices: $COUNT"
else
    echo -e "${YELLOW}⚠ WARNING${NC} (No invoices found)"
fi

echo ""
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo ""
echo "If you see mostly ${GREEN}GREEN${NC} checks, you're good to go!"
echo ""
echo "Next steps:"
echo "1. Create a test invoice via API"
echo "2. Download PDF: GET /api/v1/billing/invoices/{id}/pdf"
echo "3. Send email: POST /api/v1/billing/invoices/{id}/email"
echo "4. Run backup: ./scripts/backup-invoices.sh"
echo ""
