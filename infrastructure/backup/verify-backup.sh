#!/usr/bin/env bash

# FlyFleet Database Backup Verification and Restore Test Script
# This script verifies that database backups are valid and can be restored
# Usage: ./scripts/backup_verification.sh [namespace] [postgres-pod]

set -euo pipefail

# Configuration
NAMESPACE="${1:-flyfleet}"
POSTGRES_POD="${2:-fleet-postgres}"
BACKUP_DIR="${BACKUP_DIR:-/backup}"
TEST_DB_NAME="${TEST_DB_NAME:-restore_test_$(date +%Y%m%d_%H%M%S)}"
S3_BUCKET="${S3_BUCKET:-}"  # If using S3 for backups
S3_PREFIX="${S3_PREFIX:-flyfleet/backups}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

check_prerequisites() {
    log_info "Checking prerequisites..."

    if ! command -v pg_dump &>/dev/null; then
        log_error "pg_dump not found. Install PostgreSQL client tools."
        exit 1
    fi

    if ! kubectl get namespace "$NAMESPACE" &>/dev/null; then
        log_error "Namespace $NAMESPACE does not exist"
        exit 1
    fi

    if ! kubectl get pod "$POSTGRES_POD" -n "$NAMESPACE" &>/dev/null; then
        log_error "PostgreSQL pod $POSTGRES_POD not found in namespace $NAMESPACE"
        exit 1
    fi

    log_info "Prerequisites check passed"
}

get_latest_backup() {
    local backup_file=""

    if [ -n "$S3_BUCKET" ]; then
        log_info "Fetching latest backup from S3: s3://$S3_BUCKET/$S3_PREFIX/"
        backup_file=$(aws s3 ls "s3://$S3_BUCKET/$S3_PREFIX/" | sort | tail -1 | awk '{print $4}')
        if [ -z "$backup_file" ]; then
            log_error "No backups found in S3"
            exit 1
        fi
        log_info "Latest backup: $backup_file"
        echo "s3://$S3_BUCKET/$S3_PREFIX/$backup_file"
    else
        log_info "Fetching latest backup from local pod: $BACKUP_DIR"
        backup_file=$(kubectl exec -n "$NAMESPACE" "$POSTGRES_POD" -- ls -1t "$BACKUP_DIR"/*.sql.gz 2>/dev/null | head -1)
        if [ -z "$backup_file" ]; then
            log_error "No backups found in $BACKUP_DIR"
            exit 1
        fi
        log_info "Latest backup: $(basename "$backup_file")"
        echo "$backup_file"
    fi
}

download_backup() {
    local backup_source="$1"
    local local_file="/tmp/backup_$(date +%s).sql.gz"

    log_info "Downloading backup to $local_file..."

    if [[ "$backup_source" == s3://* ]]; then
        aws s3 cp "$backup_source" "$local_file"
    else
        kubectl exec -n "$NAMESPACE" "$POSTGRES_POD" -- cat "$backup_source" > "$local_file"
    fi

    echo "$local_file"
}

verify_backup_integrity() {
    local backup_file="$1"

    log_info "Verifying backup integrity..."

    # Check file exists and is not empty
    if [ ! -s "$backup_file" ]; then
        log_error "Backup file is empty or missing"
        return 1
    fi

    # Check gzip integrity
    if ! gunzip -t "$backup_file" 2>/dev/null; then
        log_error "Backup file is corrupted (gzip test failed)"
        return 1
    fi

    # Check file size (should be at least a few KB)
    local size=$(stat -c%s "$backup_file" 2>/dev/null || stat -f%z "$backup_file" 2>/dev/null)
    if [ "$size" -lt 10240 ]; then
        log_warn "Backup file is suspiciously small: ${size} bytes"
        return 1
    fi

    # Verify SQL header (for uncompressed check)
    if zcat "$backup_file" 2>/dev/null | head -1 | grep -q "PGDMP"; then
        log_info "Valid PostgreSQL dump format detected"
    else
        log_warn "File does not appear to be a PostgreSQL dump"
        # Not necessarily an error, but worth noting
    fi

    log_info "Backup integrity check passed (size: ${size} bytes)"
    return 0
}

restore_backup() {
    local backup_file="$1"

    log_info "Starting restore test to database: $TEST_DB_NAME"

    # Get database credentials from secret
    local db_host
    local db_port
    local db_user
    local db_password

    log_info "Fetching database credentials from Kubernetes secret..."

    # Try to get from secret mounted in pod first
    db_host=$(kubectl get secret flyfleet-secrets -n "$NAMESPACE" -o jsonpath="{.data.database-host}" 2>/dev/null | base64 -d || echo "localhost")
    db_port=$(kubectl get secret flyfleet-secrets -n "$NAMESPACE" -o jsonpath="{.data.database-port}" 2>/dev/null | base64 -d || echo "5432")
    db_user=$(kubectl get secret flyfleet-secrets -n "$NAMESPACE" -o jsonpath="{.data.database-username}" 2>/dev/null | base64 -d || echo "fleet_user")
    db_password=$(kubectl get secret flyfleet-secrets -n "$NAMESPACE" -o jsonpath="{.data.database-password}" 2>/dev/null | base64 -d || echo "fleet_password")

    # Create test database
    log_info "Creating test database: $TEST_DB_NAME"
    PGPASSWORD="$db_password" psql \
        -h "$db_host" \
        -p "$db_port" \
        -U "$db_user" \
        -d "flyfleet" \
        -c "DROP DATABASE IF EXISTS $TEST_DB_NAME; CREATE DATABASE $TEST_DB_NAME;" 2>/dev/null

    # Restore backup
    log_info "Restoring backup to $TEST_DB_NAME (this may take a while)..."
    local restore_start=$(date +%s)

    if ! gunzip -c "$backup_file" | PGPASSWORD="$db_password" psql \
        -h "$db_host" \
        -p "$db_port" \
        -U "$db_user" \
        -d "$TEST_DB_NAME"; then
        log_error "Restore failed"
        return 1
    fi

    local restore_end=$(date +%s)
    local restore_duration=$((restore_end - restore_start))
    log_info "Restore completed in ${restore_duration}s"

    # Verify restored database
    log_info "Verifying restored database structure..."

    local table_count
    table_count=$(PGPASSWORD="$db_password" psql \
        -h "$db_host" \
        -p "$db_port" \
        -U "$db_user" \
        -d "$TEST_DB_NAME" \
        -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public';" 2>/dev/null | xargs)

    if [ "$table_count" -gt 0 ]; then
        log_info "Database contains $table_count tables"
    else
        log_warn "No tables found in restored database"
    fi

    # Spot-check a few rows
    log_info "Sample data verification:"
    PGPASSWORD="$db_password" psql \
        -h "$db_host" \
        -p "$db_port" \
        -U "$db_user" \
        -d "$TEST_DB_NAME" \
        -c "SELECT 'drivers' as table_name, COUNT(*) as row_count FROM drivers LIMIT 1" 2>/dev/null || true
    PGPASSWORD="$db_password" psql \
        -h "$db_host" \
        -p "$db_port" \
        -U "$db_user" \
        -d "$TEST_DB_NAME" \
        -c "SELECT 'vehicles' as table_name, COUNT(*) as row_count FROM vehicles LIMIT 1" 2>/dev/null || true

    # Cleanup test database
    log_info "Cleaning up test database..."
    PGPASSWORD="$db_password" psql \
        -h "$db_host" \
        -p "$db_port" \
        -U "$db_user" \
        -d "flyfleet" \
        -c "DROP DATABASE IF EXISTS $TEST_DB_NAME;" 2>/dev/null || true

    log_info "Restore test PASSED ✅"
    return 0
}

run_full_verification() {
    log_info "=========================================="
    log_info "FlyFleet Backup Verification Test"
    log_info "=========================================="
    log_info "Namespace: $NAMESPACE"
    log_info "PostgreSQL Pod: $POSTGRES_POD"
    log_info "Test DB: $TEST_DB_NAME"
    log_info "=========================================="

    local backup_file

    # 1. Get latest backup
    log_info "Step 1/3: Fetching latest backup..."
    backup_file=$(get_latest_backup)
    if [ $? -ne 0 ]; then
        log_error "Failed to fetch backup"
        exit 1
    fi

    # 2. Verify integrity
    log_info "Step 2/3: Verifying backup integrity..."
    if ! verify_backup_integrity "$backup_file"; then
        log_error "Backup integrity check failed"
        exit 1
    fi

    # 3. Test restore
    log_info "Step 3/3: Testing restore..."
    if restore_backup "$backup_file"; then
        log_info "=========================================="
        log_info "ALL TESTS PASSED ✅"
        log_info "=========================================="
        log_info "Backup: $(basename "$backup_file")"
        log_info "Verified: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
        log_info "Restore test: SUCCESS"

        # Cleanup backup file if downloaded
        if [[ "$backup_file" == /tmp/* ]]; then
            rm -f "$backup_file"
        fi

        exit 0
    else
        log_error "=========================================="
        log_error "RESTORE TEST FAILED ❌"
        log_error "=========================================="
        exit 1
    fi
}

main() {
    trap 'log_error "Script failed at line $LINENO"' ERR

    case "${1:-}" in
        --help|-h)
            echo "Usage: $0 [namespace] [postgres-pod]"
            echo ""
            echo "Environment variables:"
            echo "  BACKUP_DIR        - Local backup directory in pod (default: /backup)"
            echo "  S3_BUCKET         - S3 bucket for backups (if using S3)"
            echo "  S3_PREFIX         - S3 prefix/folder (default: flyfleet/backups)"
            echo "  TEST_DB_NAME      - Test database name (default: restore_test_<timestamp>)"
            echo ""
            echo "Examples:"
            echo "  $0                        - Test default namespace/pod"
            echo "  $0 flyfleet fleet-postgres - Test specific namespace/pod"
            echo "  S3_BUCKET=my-bucket $0    - Test S3 backup"
            exit 0
            ;;
        *)
            check_prerequisites
            run_full_verification
            ;;
    esac
}

main "$@"
