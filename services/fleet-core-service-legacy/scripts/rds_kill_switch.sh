#!/bin/bash
# scripts/rds_kill_switch.sh
# Automated Kill Switch triggered by Prometheus AlertManager.
# Used when RDS CPU > 85%.

set -euo pipefail

LOG_FILE="/var/log/flyfleet/rds_killswitch.log"
DATE=$(date +'%Y-%m-%dT%H:%M:%SZ')

echo "[$DATE] [CRITICAL] RDS CPU exceeded 85%. EXECUTING HEADROOM KILL SWITCH." | tee -a $LOG_FILE

# 1. Scale down analytics (heavy read)
kubectl scale deployment analytics-worker --replicas=0 -n flyfleet-prod
echo "[$DATE] analytics-worker scaled to 0" | tee -a $LOG_FILE

# 2. Scale down reports (heavy read)
kubectl scale deployment reports-worker --replicas=0 -n flyfleet-prod
echo "[$DATE] reports-worker scaled to 0" | tee -a $LOG_FILE

# 3. Scale down maintenance (heavy read)
kubectl scale deployment maintenance-worker --replicas=0 -n flyfleet-prod
echo "[$DATE] maintenance-worker scaled to 0" | tee -a $LOG_FILE

# We intentionally LEAVE 'tracking', 'auth', and 'emergency' (The Critical pools) running
# to ensure SOS triggers and live vehicle locations continue uninterrupted.

echo "[$DATE] [SUCCESS] Low-Priority domains suspended. Headroom preserved." | tee -a $LOG_FILE
exit 0
