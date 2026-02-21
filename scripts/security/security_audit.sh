#!/bin/bash
set -e

echo "======================================"
echo " Starting FlyFleet Security Audit "
echo "======================================"

# Determine base directory
BASE_DIR="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$BASE_DIR"

# 1. Check for Hardcoded Secrets using Trivy
echo "\n[1] Scanning for hardcoded secrets and vulnerabilities (Trivy)..."
if command -v trivy &> /dev/null; then
  # Scan filesystem, ignoring known tests/vendor, only reporting HIGH/CRITICAL
  trivy fs . --scanners vuln,secret --severity HIGH,CRITICAL --skip-dirs "node_modules,vendor,apps/dashboard-web/build"
else
  echo "\n[WARNING] 'trivy' is not installed. Skipping deep secret & vulnerability scan."
  echo "         Please install trivy: https://trivy.dev/latest/getting-started/installation"
fi

# 2. Run Go Security Checker (gosec)
echo "\n[2] Scanning Go services for security misconfigurations (Gosec)..."
if command -v gosec &> /dev/null; then
  # Find all go modules and run gosec on them
  # Ignore test folders
  gosec -exclude-dir=vendor -exclude-dir=test -severity HIGH -confidence HIGH ./...
else
  echo "\n[WARNING] 'gosec' is not installed. Skipping Go AST security scan."
  echo "         Please run: go install github.com/securego/gosec/v2/cmd/gosec@latest"
fi

echo "======================================"
echo " Security Audit Completed! "
echo "======================================"
