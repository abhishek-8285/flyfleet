#!/usr/bin/env bash
#
# FlyFleet Development Environment Bootstrap Script
# This script sets up a complete local development environment.
#
# Usage: ./scripts/bootstrap.sh [--skip-seed] [--help]
#   --skip-seed: Skip database seeding step
#   --help: Show this help message
#

set -euo pipefail

# Colors for output
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
RESET='\033[0m'

# Configuration
COMPOSE_FILE="docker-compose.yml"
SEED_DATA=true
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --skip-seed)
      SEED_DATA=false
      shift
      ;;
    --help|-h)
      echo "FlyFleet Bootstrap Script"
      echo ""
      echo "Usage: $0 [OPTIONS]"
      echo ""
      echo "Options:"
      echo "  --skip-seed    Skip database seeding"
      echo "  --help, -h     Show this help message"
      echo ""
      echo "This script will:"
      echo "  1. Check prerequisites (Go, Node, Docker)"
      echo "  2. Start PostgreSQL, Redis, and NATS"
      echo "  3. Wait for databases to be ready"
      echo "  4. Sync Go workspace and download dependencies"
      echo "  5. Install frontend dependencies (npm workspaces)"
      echo "  6. Optionally seed test data"
      echo ""
      echo "After running, you can start services with:"
      echo "  - make build && docker-compose up"
      echo "  - Or start individual services in separate terminals"
      exit 0
      ;;
    *)
      echo -e "${RED}Unknown option: $1${RESET}"
      echo "Use --help for usage information"
      exit 1
      ;;
  esac
done

# Helper functions
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

check_command() {
  if command -v "$1" >/dev/null 2>&1; then
    return 0
  else
    return 1
  fi
}

# Start banner
echo ""
echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${RESET}"
echo -e "${GREEN}║   FlyFleet Development Environment Bootstrap            ║${RESET}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${RESET}"
echo ""

# Change to project root
cd "$PROJECT_ROOT"
log_info "Project root: $PROJECT_ROOT"

# 1. Check prerequisites
log_info "Checking prerequisites..."

MISSING_DEPS=0

if ! check_command go; then
  log_error "Go is not installed. Please install Go 1.23+ from https://go.dev/dl/"
  MISSING_DEPS=1
else
  GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
  log_success "Found Go $GO_VERSION"
  # Check minimum version (1.23)
  if [[ $(printf '%s\n' "1.23" "$GO_VERSION" | sort -V | head -n1) != "1.23" ]]; then
    log_warning "Go version $GO_VERSION is older than recommended 1.23+"
  fi
fi

if ! check_command node; then
  log_error "Node.js is not installed. Please install Node.js 20+ from https://nodejs.org/"
  MISSING_DEPS=1
else
  NODE_VERSION=$(node --version)
  log_success "Found Node.js $NODE_VERSION"
fi

if ! check_command npm; then
  log_error "npm is not installed (comes with Node.js)"
  MISSING_DEPS=1
else
  log_success "Found npm $(npm --version)"
fi

if ! check_command docker; then
  log_error "Docker is not installed. Please install Docker Desktop or Docker Engine"
  MISSING_DEPS=1
else
  log_success "Found Docker $(docker --version | head -n1)"
fi

if ! check_command docker-compose; then
  # Try docker compose (v2)
  if docker compose version >/dev/null 2>&1; then
    log_success "Found Docker Compose (v2)"
    COMPOSE_CMD="docker compose"
  else
    log_error "Docker Compose is not installed"
    MISSING_DEPS=1
  fi
else
  log_success "Found Docker Compose $(docker-compose --version | head -n1)"
  COMPOSE_CMD="docker-compose"
fi

if [[ $MISSING_DEPS -eq 1 ]]; then
  log_error "Missing required dependencies. Please install them and try again."
  exit 1
fi

# 2. Start infrastructure services
log_info "Starting infrastructure services (Postgres, Redis, NATS)..."
if ! $COMPOSE_CMD up -d postgres redis nats; then
  log_error "Failed to start services. Check docker-compose.yml for errors."
  exit 1
fi

# 3. Wait for PostgreSQL
log_info "Waiting for PostgreSQL to become ready..."
MAX_WAIT=60
WAITED=0
until docker exec fleet-postgres pg_isready -U fleet_user -d fleet_db >/dev/null 2>&1; do
  if [[ $WAITED -ge $MAX_WAIT ]]; then
    log_error "PostgreSQL did not become ready within ${MAX_WAIT}s"
    log_error "Check logs: $COMPOSE_CMD logs postgres"
    exit 1
  fi
  printf '.'
  sleep 2
  WAITED=$((WAITED + 2))
done
echo ""
log_success "PostgreSQL is ready"

# Wait for Redis
log_info "Waiting for Redis to become ready..."
WAITED=0
until docker exec fleet-redis redis-cli ping >/dev/null 2>&1; do
  if [[ $WAITED -ge $MAX_WAIT ]]; then
    log_error "Redis did not become ready within ${MAX_WAIT}s"
    log_error "Check logs: $COMPOSE_CMD logs redis"
    exit 1
  fi
  printf '.'
  sleep 2
  WAITED=$((WAITED + 2))
done
echo ""
log_success "Redis is ready"

# Wait for NATS
log_info "Waiting for NATS to become ready..."
WAITED=0
until curl -s http://localhost:8222/varz >/dev/null 2>&1; do
  if [[ $WAITED -ge $MAX_WAIT ]]; then
    log_error "NATS did not become ready within ${MAX_WAIT}s"
    log_error "Check logs: $COMPOSE_CMD logs nats"
    exit 1
  fi
  printf '.'
  sleep 2
  WAITED=$((WAITED + 2))
done
echo ""
log_success "NATS is ready"

# 4. Go workspace setup
log_info "Syncing Go workspace..."
if ! go work sync; then
  log_error "go work sync failed. Check go.work file."
  exit 1
fi

log_info "Downloading Go dependencies..."
if ! go mod download; then
  log_error "Failed to download Go dependencies"
  exit 1
fi
log_success "Go dependencies ready"

# 5. NPM workspaces setup
log_info "Installing frontend dependencies (this may take a few minutes)..."
if ! npm install --workspaces; then
  log_error "npm install failed"
  exit 1
fi
log_success "Frontend dependencies installed"

# 6. Optional data seeding
if [[ "$SEED_DATA" == "true" ]]; then
  log_info "Checking for seed data scripts..."

  # Look for seed scripts
  if [[ -f "scripts/seed_dev_data.sh" ]]; then
    log_info "Running seed data script..."
    if bash scripts/seed_dev_data.sh; then
      log_success "Seed data loaded"
    else
      log_warning "Seed data script failed (this is optional)"
    fi
  elif [[ -f "scripts/seed_dev_data.go" ]]; then
    log_info "Running Go seed data program..."
    if go run scripts/seed_dev_data.go; then
      log_success "Seed data loaded"
    else
      log_warning "Seed data script failed (this is optional)"
    fi
  else
    log_warning "No seed data script found. Skipping."
    log_info "To add seed data, create scripts/seed_dev_data.sh or scripts/seed_dev_data.go"
  fi
fi

# Success summary
echo ""
echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${RESET}"
echo -e "${GREEN}║   Bootstrap Complete!                                    ║${RESET}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${RESET}"
echo ""
echo "Services running:"
echo "  - PostgreSQL:  localhost:5433 (user: fleet_user, db: fleet_db)"
echo "  - Redis:       localhost:6380"
echo "  - NATS:        localhost:4222 (JetStream enabled)"
echo ""
echo "To start development:"
echo ""
echo "  Option A - Build and run with Docker Compose:"
echo "    ${BLUE}docker-compose up${RESET}  (or ${BLUE}make docker-build && docker-compose up${RESET})"
echo ""
echo "  Option B - Run services individually:"
echo "    Terminal 1: ${BLUE}cd services/fleet-core && go run cmd/main.go${RESET}"
echo "    Terminal 2: ${BLUE}cd apps/dashboard-web && npm run dev${RESET}"
echo ""
echo "  Option C - Build all:"
echo "    ${BLUE}make build${RESET}"
echo ""
echo "Useful commands:"
echo "  ${BLUE}make help${RESET}       - Show all available make targets"
echo "  ${BLUE}make test${RESET}      - Run all tests"
echo "  ${BLUE}make lint${RESET}      - Run linters"
echo "  ${BLUE}make ci${RESET}        - Run full CI pipeline locally"
echo ""
echo "Documentation:"
echo "  - README: file://$PROJECT_ROOT/README.md"
echo "  - Architecture: file://$PROJECT_ROOT/FULL_SPECTRUM_SYSTEM_SYNTHESIS.md"
echo "  - Improvement Analysis: file://$PROJECT_ROOT/IMPROVEMENT_ANALYSIS.md"
echo ""
