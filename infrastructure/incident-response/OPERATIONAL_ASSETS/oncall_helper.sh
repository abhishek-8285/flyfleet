#!/bin/bash

# FlyFleet On-Call Helper Script
# Run this when you get paged to quickly open all necessary tools

set -e

echo "=================================================="
echo "  FlyFleet Incident Response - On-Call Helper"
echo "=================================================="
echo ""

# Check if we're in the right directory
if [ ! -d "infrastructure/incident-response" ]; then
    echo "Error: Run from flyfleet root directory"
    exit 1
fi

# Function to open URLs (works on Linux, macOS, Termux)
open_url() {
    if command -v xdg-open &> /dev/null; then
        xdg-open "$1" &
    elif command -v open &> /dev/null; then
        open "$1" &
    elif command -v termux-open-url &> /dev/null; then
        termux-open-url "$1" &
    else
        echo "Please open manually: $1"
    fi
}

# 1. Show quick reference
echo "📋 Quick Reference (keep this terminal open)"
echo "--------------------------------"
cat infrastructure/incident-response/SEVERITY_QUICK_REFERENCE.md | head -50
echo ""
echo "Full doc: infrastructure/incident-response/SEVERITY_QUICK_REFERENCE.md"
echo ""

# 2. Ask for incident details
read -p "Incident ticket number (INC-YYYY-NNNN): " INCIDENT_ID
read -p "Severity (SEV1/SEV2/SEV3/SEV4): " SEVERITY
read -p "Service affected (fleet-core, database, nats, etc.): " SERVICE

if [ -z "$INCIDENT_ID" ]; then
    INCIDENT_ID="TBD-$(date +%s)"
fi

# 3. Create war room channel reminder
echo ""
echo "=================================================="
echo "  ACTION REQUIRED"
echo "=================================================="
echo "1. Create Linear ticket: $INCIDENT_ID"
echo "2. Create war room: #incident-$INCIDENT_ID"
echo "3. Send initial notification (see templates/)"
echo "4. Assign IC and roles"
echo ""

# 4. Open tools based on service
echo "🔧 Opening tools for $SERVICE..."
echo ""

# Always open monitoring
echo "Opening Grafana..."
open_url "https://grafana.flyfleet.com/d/fleet-overview" || echo "Manual: https://grafana.flyfleet.com/d/fleet-overview"

# Open appropriate runbook
RUNBOOK_DIR="infrastructure/incident-response/runbooks"
case "$SERVICE" in
    "fleet-core"|"api")
        echo "Opening API outage runbook..."
        cat "$RUNBOOK_DIR/api-outage.md" | head -40
        ;;
    "telemetry")
        echo "Opening telemetry disruption runbook..."
        cat "$RUNBOOK_DIR/telemetry-disruption.md" | head -40
        ;;
    "database"|"postgres"|"db")
        echo "Opening database performance runbook..."
        cat "$RUNBOOK_DIR/database-performance.md" | head -40
        ;;
    "nats")
        echo "Opening NATS issues runbook..."
        cat "$RUNBOOK_DIR/nats-issues.md" | head -40
        ;;
    "auth")
        echo "Opening auth failure runbook..."
        cat "$RUNBOOK_DIR/auth-failure.md" | head -40
        ;;
    *)
        echo "No specific runbook for $SERVICE. Review all runbooks:"
        ls -1 "$RUNBOOK_DIR/"
        ;;
esac

# 5. Quick commands cheat sheet
echo ""
echo "=================================================="
echo "  QUICK COMMANDS"
echo "=================================================="
echo "kubectl get pods -n flyfleet -o wide"
echo "kubectl logs -f deployment/fleet-core -n flyfleet --since=10m"
echo "kubectl rollout status deployment/fleet-core -n flyfleet"
echo "kubectl exec -it postgres-0 -n flyfleet -- pg_isready"
echo "kubectl exec -it nats-0 -n flyfleet -- nats server report"
echo ""

# 6. Show checklist
echo "=================================================="
echo "  0-15 MINUTE CHECKLIST"
echo "=================================================="
echo "[] Acknowledge PagerDuty"
echo "[] Declare SEV${SEVERITY: -1}"
echo "[] Create Linear ticket: $INCIDENT_ID"
echo "[] Create war room: #incident-$INCIDENT_ID"
echo "[] Assign IC"
echo "[] Invite SMEs (@dba-team, @infra-sre, etc.)"
echo "[] Send initial notification (template)"
echo "[] Update status page"
echo "[] Open runbook and begin investigation"
echo "[] Send first update at: $(date -d '+15 minutes' +%H:%M UTC)"
echo ""

# 7. Open Linear if available
echo "Opening Linear..."
open_url "https://linear.app/flyfleet" || echo "Manual: https://linear.app/flyfleet"

# 8. Final reminder
echo "=================================================="
echo "  REMEMBER"
echo "=================================================="
echo "- Use quick reference: SEVERITY_QUICK_REFERENCE.md"
echo "- Follow runbook steps in order"
echo "- Update every 15/30 minutes"
echo "- Preserve logs for PIR"
echo "- Communication is 50% of the job"
echo ""
echo "Good luck! 🚀"
echo "=================================================="

# 9. Open full framework in browser (README)
# open_url "file://$(pwd)/infrastructure/incident-response/README.md"
