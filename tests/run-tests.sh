#!/usr/bin/env bash

# FlyFleet Infrastructure Test Runner
# This script runs all infrastructure validation tests

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

run_tests() {
    local test_type="${1:-all}"

    case "$test_type" in
        terraform)
            log_info "Running Terraform validation tests..."
            pushd "$PROJECT_ROOT/infrastructure/terraform" > /dev/null

            log_info "1/3 Checking terraform format..."
            terraform fmt -check -recursive || {
                log_error "Terraform formatting issues found. Run: terraform fmt -recursive"
                exit 1
            }
            log_info "✅ Format check passed"

            log_info "2/3 Validating Terraform configuration..."
            terraform validate || {
                log_error "Terraform validation failed"
                exit 1
            }
            log_info "✅ Validation passed"

            log_info "3/3 Checking module structure..."
            if [ -f "variables.tf" ] && [ -f "outputs.tf" ]; then
                log_info "✅ Root module has variables and outputs"
            else
                log_error "Root module missing variables.tf or outputs.tf"
                exit 1
            fi

            popd > /dev/null
            log_info "All Terraform tests passed ✅"
            ;;

        smoke)
            log_info "Running smoke tests..."
            export SKIP_K8S_TESTS="${SKIP_K8S_TESTS:-true}"
            export SKIP_API_TESTS="${SKIP_API_TESTS:-true}"
            export SKIP_DB_TESTS="${SKIP_DB_TESTS:-true}"
            export SKIP_SECRETS_TESTS="${SKIP_SECRETS_TESTS:-true}"
            export SKIP_HELM_TESTS="${SKIP_HELM_TESTS:-true}"
            go test -v "$SCRIPT_DIR/infrastructure" -run "TestTerraform|TestRequired|TestModule"
            ;;

        integration)
            log_warn "Integration tests require live Kubernetes cluster"
            log_info "Set FLYFLEET_NAMESPACE and FLYFLEET_API_URL if not running in cluster"
            SKIP_K8S_TESTS=false SKIP_API_TESTS=false go test -v "$SCRIPT_DIR/infrastructure"
            ;;

        all)
            log_info "Running full test suite..."
            "$0" terraform
            "$0" smoke
            log_info "All test suites passed ✅"
            ;;

        *)
            echo "Usage: $0 [terraform|smoke|integration|all]"
            exit 1
            ;;
    esac
}

main() {
    case "${1:-}" in
        --help|-h)
            echo "FlyFleet Infrastructure Test Runner"
            echo ""
            echo "Usage: $0 [test-type]"
            echo ""
            echo "Test Types:"
            echo "  terraform    - Validate Terraform configuration (format, syntax, structure)"
            echo "  smoke        - Fast smoke tests (unit tests only, skip integration)"
            echo "  integration  - Full integration tests (requires live cluster)"
            echo "  all          - Run all test types (default)"
            echo ""
            echo "Environment Variables:"
            echo "  SKIP_*        - Skip specific test categories (true/false)"
            echo "  FLYFLEET_NAMESPACE - Kubernetes namespace"
            echo "  FLYFLEET_API_URL   - API endpoint for health checks"
            echo ""
            echo "Examples:"
            echo "  $0                    # Run all tests"
            echo "  $0 terraform         # Just Terraform checks"
            echo "  $0 smoke            # Fast smoke tests"
            echo "  SKIP_K8S_TESTS=false $0 integration  # Force k8s tests"
            exit 0
            ;;
        "")
            run_tests "all"
            ;;
        *)
            run_tests "$1"
            ;;
    esac
}

main "$@"
