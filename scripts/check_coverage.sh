#!/usr/bin/env bash
#
# FlyFleet Test Coverage Gate
# Fails if test coverage is below the threshold (default: 70%)
#
# Usage: ./scripts/check_coverage.sh [threshold]
#   threshold: minimum coverage percentage (default: 70)

set -euo pipefail

# Colors
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
RESET='\033[0m'

# Configuration
THRESHOLD="${1:-70}"  # Default 70%
COVERAGE_PROFILE="${COVERAGE_PROFILE:-coverage.out}"
COVERAGE_HTML="${COVERAGE_HTML:-coverage.html}"
MODULES="./..."

log_info() {
  echo -e "${BLUE}[INFO]${RESET} $1"
}

log_success() {
  echo -e "${GREEN}[SUCCESS]${RESET} $1"
}

log_warning() {
  echo -e "${YELLOW}[WARN]${RESET} $1"
}

log_error() {
  echo -e "${RED}[ERROR]${RESET} $1"
}

# Check that coverage profile exists
if [[ ! -f "$COVERAGE_PROFILE" ]]; then
  log_error "Coverage profile not found: $COVERAGE_PROFILE"
  log_info "Run 'go test ./... -coverprofile=$COVERAGE_PROFILE -covermode=atomic' first"
  exit 1
fi

log_info "Checking test coverage (threshold: ${THRESHOLD}%)..."

# Extract total coverage percentage from coverage.out
# The format is: <filename>:<line>.<column>,<line>.<column> <count> <status>
# We need the total percentage which is printed by go tool cover -func
TOTAL_COV=$(go tool cover -func="$COVERAGE_PROFILE" | tail -1 | awk '{print $3}' | sed 's/%//')

log_info "Total coverage: ${TOTAL_COV}%"

# Compare with threshold (handle floating point)
COMPARE_RESULT=$(echo "$TOTAL_COV >= $THRESHOLD" | bc -l 2>/dev/null || echo "0")

if [[ "$COMPARE_RESULT" -eq 1 ]]; then
  log_success "Coverage meets threshold (${TOTAL_COV}% >= ${THRESHOLD}%)"

  # Also generate HTML report if requested
  if [[ "${GENERATE_HTML:-1}" == "1" ]]; then
    log_info "Generating HTML coverage report: $COVERAGE_HTML"
    go tool cover -html="$COVERAGE_PROFILE" -o "$COVERAGE_HTML"
    log_success "HTML report generated: $COVERAGE_HTML"
  fi

  exit 0
else
  log_error "Coverage below threshold (${TOTAL_COV}% < ${THRESHOLD}%)"
  echo ""
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "Coverage gate failed!"
  echo ""
  echo "To fix:"
  echo "  1. Write tests for uncovered code"
  echo "  2. Run: go test ./... -coverprofile=coverage.out -covermode=atomic"
  echo "  3. Review coverage: go tool cover -html=coverage.out"
  echo ""
  echo "Modules with low coverage:"
  go tool cover -func="$COVERAGE_PROFILE" | grep -v "total:" | sort -k3 -n | tail -10
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  exit 1
fi
