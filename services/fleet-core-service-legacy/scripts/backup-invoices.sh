#!/bin/bash
#
# FlyFleet Invoice Backup Script
# Backs up invoices to Google Drive using rclone
#
# Usage: ./backup-invoices.sh
#
# Setup:
# 1. Install rclone: curl https://rclone.org/install.sh | sudo bash
# 2. Configure Google Drive: rclone config
# 3. Name your remote: gdrive
# 4. Add to crontab: 0 2 * * * /opt/flyfleet/scripts/backup-invoices.sh
#

set -e

# Configuration
INVOICE_DIR="${INVOICE_DIR:-/opt/flyfleet/data/invoices}"
BACKUP_DIR="${BACKUP_DIR:-/opt/flyfleet/data/backups}"
RCLONE_REMOTE="${RCLONE_REMOTE:-gdrive:flyfleet-backup}"
RETENTION_DAYS="${RETENTION_DAYS:-90}"
LOG_FILE="${LOG_FILE:-/var/log/flyfleet/backup.log}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging function
log() {
    local level=$1
    local message=$2
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo -e "${timestamp} [${level}] ${message}" | tee -a "$LOG_FILE"
}

# Check if rclone is installed
check_rclone() {
    if ! command -v rclone &> /dev/null; then
        log "ERROR" "${RED}rclone is not installed. Please install it first.${NC}"
        log "ERROR" "Run: curl https://rclone.org/install.sh | sudo bash"
        exit 1
    fi
}

# Check if rclone is configured
check_rclone_config() {
    if ! rclone config file | grep -q "$RCLONE_REMOTE"; then
        log "ERROR" "${RED}rclone remote '$RCLONE_REMOTE' is not configured.${NC}"
        log "ERROR" "Run: rclone config"
        exit 1
    fi
}

# Create backup directory
create_backup_dir() {
    if [ ! -d "$BACKUP_DIR" ]; then
        log "INFO" "Creating backup directory: $BACKUP_DIR"
        mkdir -p "$BACKUP_DIR"
    fi
}

# Create daily backup
create_backup() {
    local DATE=$(date +%Y%m%d_%H%M%S)
    local BACKUP_FILE="$BACKUP_DIR/invoices_$DATE.tar.gz"
    
    log "INFO" "${GREEN}Starting backup...${NC}"
    log "INFO" "Source: $INVOICE_DIR"
    log "INFO" "Destination: $BACKUP_FILE"
    
    # Create tarball
    tar -czf "$BACKUP_FILE" -C "$(dirname $INVOICE_DIR)" "$(basename $INVOICE_DIR)" 2>/dev/null
    
    if [ $? -eq 0 ]; then
        local SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
        log "INFO" "${GREEN}Backup created: $BACKUP_FILE ($SIZE)${NC}"
    else
        log "ERROR" "${RED}Failed to create backup${NC}"
        exit 1
    fi
}

# Sync to Google Drive
sync_to_cloud() {
    log "INFO" "${GREEN}Syncing to Google Drive...${NC}"
    
    rclone sync "$BACKUP_DIR" "$RCLONE_REMOTE/invoices" \
        --progress \
        --transfers=4 \
        --checkers=8 \
        --tpslimit=10 \
        --retries=3 \
        2>&1 | tee -a "$LOG_FILE"
    
    if [ $? -eq 0 ]; then
        log "INFO" "${GREEN}Sync completed successfully${NC}"
    else
        log "ERROR" "${RED}Sync failed${NC}"
        exit 1
    fi
}

# Cleanup old local backups
cleanup_local() {
    log "INFO" "Cleaning up local backups older than $RETENTION_DAYS days..."
    
    local COUNT=$(find "$BACKUP_DIR" -name "invoices_*.tar.gz" -mtime +$RETENTION_DAYS | wc -l)
    
    if [ $COUNT -gt 0 ]; then
        find "$BACKUP_DIR" -name "invoices_*.tar.gz" -mtime +$RETENTION_DAYS -delete
        log "INFO" "Deleted $COUNT old backup files"
    else
        log "INFO" "No old backups to delete"
    fi
}

# Verify backup integrity
verify_backup() {
    log "INFO" "Verifying latest backup..."
    
    local LATEST=$(ls -t "$BACKUP_DIR"/invoices_*.tar.gz 2>/dev/null | head -1)
    
    if [ -z "$LATEST" ]; then
        log "ERROR" "${RED}No backup files found${NC}"
        return 1
    fi
    
    # Test tarball integrity
    if tar -tzf "$LATEST" > /dev/null 2>&1; then
        log "INFO" "${GREEN}Backup integrity verified${NC}"
        return 0
    else
        log "ERROR" "${RED}Backup integrity check failed${NC}"
        return 1
    fi
}

# Send notification (optional - webhook/email)
send_notification() {
    local STATUS=$1
    local MESSAGE=$2
    
    # Webhook notification (optional)
    if [ -n "$NOTIFICATION_WEBHOOK" ]; then
        curl -X POST "$NOTIFICATION_WEBHOOK" \
            -H "Content-Type: application/json" \
            -d "{\"status\": \"$STATUS\", \"message\": \"$MESSAGE\", \"timestamp\": \"$(date -Iseconds)\"}" \
            > /dev/null 2>&1 || true
    fi
    
    # Log notification
    log "INFO" "Notification: [$STATUS] $MESSAGE"
}

# Main execution
main() {
    log "INFO" "=========================================="
    log "INFO" "FlyFleet Invoice Backup Started"
    log "INFO" "=========================================="
    
    # Pre-flight checks
    check_rclone
    check_rclone_config
    create_backup_dir
    
    # Create backup
    create_backup
    
    # Verify backup
    if verify_backup; then
        # Sync to cloud
        sync_to_cloud
        
        # Cleanup old backups
        cleanup_local
        
        # Send success notification
        send_notification "SUCCESS" "Backup completed successfully"
        
        log "INFO" "${GREEN}=========================================="
        log "INFO" "Backup completed successfully"
        log "INFO" "==========================================${NC}"
    else
        # Send failure notification
        send_notification "FAILURE" "Backup verification failed"
        
        log "ERROR" "${RED}=========================================="
        log "ERROR" "Backup failed verification"
        log "ERROR" "==========================================${NC}"
        exit 1
    fi
}

# Show help
show_help() {
    echo "FlyFleet Invoice Backup Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  backup    Run full backup (default)"
    echo "  sync      Sync existing backups to cloud"
    echo "  cleanup   Clean up old local backups"
    echo "  verify    Verify latest backup integrity"
    echo "  status    Show backup status"
    echo "  help      Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  INVOICE_DIR      Invoice directory (default: /opt/flyfleet/data/invoices)"
    echo "  BACKUP_DIR       Backup directory (default: /opt/flyfleet/data/backups)"
    echo "  RCLONE_REMOTE    rclone remote name (default: gdrive:flyfleet-backup)"
    echo "  RETENTION_DAYS   Local backup retention (default: 90)"
    echo "  LOG_FILE         Log file path (default: /var/log/flyfleet/backup.log)"
    echo "  NOTIFICATION_WEBHOOK  Webhook URL for notifications"
}

# Show status
show_status() {
    echo "FlyFleet Backup Status"
    echo "======================"
    echo ""
    echo "Invoice Directory: $INVOICE_DIR"
    echo "Backup Directory: $BACKUP_DIR"
    echo "rclone Remote: $RCLONE_REMOTE"
    echo "Retention: $RETENTION_DAYS days"
    echo ""
    
    if [ -d "$INVOICE_DIR" ]; then
        local INVOICE_COUNT=$(find "$INVOICE_DIR" -name "*.pdf" | wc -l)
        local INVOICE_SIZE=$(du -sh "$INVOICE_DIR" | cut -f1)
        echo "Invoices: $INVOICE_COUNT files ($INVOICE_SIZE)"
    else
        echo "Invoices: Directory not found"
    fi
    
    if [ -d "$BACKUP_DIR" ]; then
        local BACKUP_COUNT=$(ls "$BACKUP_DIR"/invoices_*.tar.gz 2>/dev/null | wc -l)
        local BACKUP_SIZE=$(du -sh "$BACKUP_DIR" 2>/dev/null | cut -f1)
        echo "Backups: $BACKUP_COUNT files ($BACKUP_SIZE)"
        
        local LATEST=$(ls -t "$BACKUP_DIR"/invoices_*.tar.gz 2>/dev/null | head -1)
        if [ -n "$LATEST" ]; then
            echo "Latest: $(basename $LATEST) ($(stat -c %y "$LATEST" 2>/dev/null | cut -d' ' -f1))"
        fi
    else
        echo "Backups: Directory not found"
    fi
}

# Parse command
case "${1:-backup}" in
    backup)
        main
        ;;
    sync)
        check_rclone
        check_rclone_config
        sync_to_cloud
        ;;
    cleanup)
        cleanup_local
        ;;
    verify)
        verify_backup
        ;;
    status)
        show_status
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo "Unknown command: $1"
        show_help
        exit 1
        ;;
esac
