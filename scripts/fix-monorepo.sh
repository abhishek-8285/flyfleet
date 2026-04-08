#!/usr/bin/env bash
# FlyFleet Monorepo Cleanup and Optimization Script
# This script fixes dependency bloat and optimizes the npm workspaces setup

set -euo pipefail

echo "=== FlyFleet Monorepo Cleanup ==="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if we're in the right directory
if [ ! -f "package.json" ] || ! grep -q '"workspaces"' package.json; then
  echo -e "${RED}Error: This script must be run from the flyfleet monorepo root${NC}"
  exit 1
fi

echo -e "${YELLOW}WARNING: This will:${NC}"
echo "  1. Remove ALL node_modules directories (root + workspaces)"
echo "  2. Clean package-lock.json"
echo "  3. Reinstall dependencies with proper hoisting"
echo "  4. Clean up ~930MB of duplicate dependencies"
echo ""
read -p "Continue? (yes/no): " -r
if [[ ! $REPLY =~ ^yes$ ]]; then
  echo "Aborted."
  exit 1
fi

# Step 1: Remove node_modules
echo ""
echo -e "${GREEN}[1/4] Removing node_modules directories...${NC}"
rm -rf node_modules
rm -rf apps/*/node_modules
rm -rf packages/*/node_modules
echo "✓ Removed all node_modules"

# Step 2: Clean lockfile
echo ""
echo -e "${GREEN}[2/4] Cleaning package-lock.json...${NC}"
rm -f package-lock.json
echo "✓ Removed package-lock.json"

# Step 3: Verify .npmrc exists
echo ""
echo -e "${GREEN}[3/4] Checking .npmrc configuration...${NC}"
if [ ! -f ".npmrc" ]; then
  echo "ERROR: .npmrc not found! Run the monorepo fix first."
  exit 1
fi
echo "✓ .npmrc present with hoisting config"

# Step 4: Reinstall dependencies
echo ""
echo -e "${GREEN}[4/4] Installing dependencies with hoisting...${NC}"
echo "This may take several minutes..."
npm install

echo ""
echo -e "${GREEN}=== Cleanup Complete ===${NC}"
echo ""
echo "Expected results:"
echo "  - Single hoisted node_modules at root (~500-600MB)"
echo "  - No node_modules in apps/* or packages/*"
echo "  - Faster installs and builds"
echo ""
echo "Verify with:"
echo "  du -sh node_modules"
echo "  find . -name node_modules -type d | wc -l  # should be 1"
echo ""
