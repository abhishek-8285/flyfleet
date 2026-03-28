# FlyFleet Database Backup Verification Guide

**Automated backup validation and restore testing for PostgreSQL**

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Setup Instructions](#setup-instructions)
4. [Running Verification Manually](#running-verification-manually)
5. [Automated Scheduled Tests](#automated-scheduled-tests)
6. Monitoring & Alerting
7. Troubleshooting
8. Best Practices

---

## Overview

Backup verification ensures that your database backups are:
- ✅ **Valid** (not corrupted)
- ✅ **Restorable** (can be successfully restored)
- ✅ **Complete** (contains all expected data)

This implementation provides:
- **Automatic integrity checks** on backup files
- **Test restore** to temporary database
- **Scheduled verification** via Kubernetes CronJob
- **Alerting** on failure (integration with Alertmanager)
- **Audit trail** of verification results

### Workflow

```
┌─────────────────────────────────────────────────────────────┐
│                  Backup Creation (daily 2:00 AM)            │
│  PostgreSQL → pg_dump → /backup/flyfleet_20260328.sql.gz  │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              Backup Verification (daily 3:00 AM)           │
│   CronJob triggers → Download backup → Verify integrity → │
│   → Restore to test DB → Spot-check data → Cleanup        │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   Results & Notifications                  │
│  Success: Log to stdout (collected by Loki)               │
│  Failure: Alertmanager → Slack/Email/PagerDuty            │
└─────────────────────────────────────────────────────────────┘
```

---

## Prerequisites

### Required

- ✅ PostgreSQL database running in Kubernetes
- ✅ Backup cronjob configured (see `infrastructure/database/backup-cronjob.yaml`)
- ✅ Backups stored either:
  - **Local PVC** mounted at `/backup` in PostgreSQL pod
  - **S3 bucket** (recommended for production)
- ✅ kubectl access to cluster
- ✅ Helm 3+ (for deploying CronJob)

### Optional

- AWS S3 bucket (for off-cluster backup storage)
- Alertmanager configured for failure notifications
- Loki/Elastic for log aggregation

---

## Setup Instructions

### Step 1: Configure S3 Bucket (Optional but Recommended)

If storing backups in S3 (recommended for production):

```bash
# Create S3 bucket
aws s3 mb s3://flyfleet-backups-$(aws sts get-caller-identity --query Account --output text) --region us-east-1

# Enable versioning (for recovery)
aws s3api put-bucket-versioning \
  --bucket flyfleet-backups-<ACCOUNT_ID> \
  --versioning-configuration Status=Enabled

# Set lifecycle policy (optional) - delete after 90 days
aws s3api put-bucket-lifecycle-configuration \
  --bucket flyfleet-backups-<ACCOUNT_ID> \
  --lifecycle-configuration file://lifecycle.json
```

`lifecycle.json`:
```json
{
  "Rules": [
    {
      "ID": "Delete old backups",
      "Status": "Enabled",
      "Expiration": { "Days": 90 }
    }
  ]
}
```

---

### Step 2: Create Backup PVC (if using local storage)

If your backups are stored on a PersistentVolume, ensure the PVC exists:

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-backup-pvc
  namespace: flyfleet
spec:
  accessModes:
    - ReadWriteMany
  storageClassName: "standard"  # Adjust to your storage class
  resources:
    requests:
      storage: 100Gi  # Size based on your backup retention
EOF
```

---

### Step 3: Deploy IAM Role (for S3 access)

If using S3, create the IAM role via Terraform:

```bash
cd infrastructure/terraform

# Add module to main-new.tf:
# module "backup-verifier" {
#   source            = "./modules/backup-verifier"
#   aws_region        = var.aws_region
#   environment       = var.environment
#   backup_s3_bucket  = "flyfleet-backups-${AWS_ACCOUNT_ID}"
#   eks_oidc_issuer_url = module.kubernetes.oidc_issuer_url
#   tags              = var.tags
#   depends_on        = [module.kubernetes]
# }

# Or apply standalone:
terraform apply -target=module.backup-verifier -var-file=terraform.tfvars
```

Save the role ARN:

```bash
terraform output -module=backup-verifier backup_verifier_role_arn
# Copy this value for Step 4
```

---

### Step 4: Deploy Verification CronJob

Create ConfigMap with verification script:

```bash
# Create ConfigMap
kubectl create configmap backup-verification-scripts \
  --from-file=verify-backup.sh=infrastructure/backup/verify-backup.sh \
  -n flyfleet
```

Update `infrastructure/backup/restore-test-cron.yaml`:
1. Replace `${AWS_ACCOUNT_ID}` with your AWS account ID
2. Replace `${ENVIRONMENT}` with `production` or `staging`
3. Set `backup_s3_bucket` environment variable if using S3

Apply the CronJob:

```bash
kubectl apply -f infrastructure/backup/restore-test-cron.yaml
```

---

### Step 5: Verify Setup

```bash
# Check CronJob exists
kubectl get cronjob backup-verification -n flyfleet

# Check ServiceAccount has IAM role annotation
kubectl get serviceaccount flyfleet-backup-verifier -n flyfleet -o yaml | grep iam

# List pods (should be none until next scheduled run)
kubectl get pods -n flyfleet -l component=backup-verification
```

---

## Running Verification Manually

### Run Full Verification

```bash
# Using default namespace and pod
kubectl create job --from=cronjob/backup-verification backup-verification-manual -n flyfleet

# Watch logs
kubectl logs -f job/backup-verification-manual -n flyfleet

# Check job status
kubectl get job backup-verification-manual -n flyfleet
```

### Run Against Specific Backup

```bash
# Get latest backup filename from S3 or pod
BACKUP_FILE="flyfleet_20260328_0200.sql.gz"

# Create one-off job with specific backup
kubectl run backup-test-$(date +%s) -n flyfleet --rm -i --tty \
  --image=postgres:15-alpine \
  --env="BACKUP_S3_BUCKET=flyfleet-backups-123456789012" \
  --env="S3_PREFIX=flyfleet/backups" \
  -- /bin/sh -c "
    apk add --no-cache aws-cli
    aws s3 cp s3://flyfleet-backups-123456789012/flyfleet/backups/$BACKUP_FILE /tmp/backup.sql.gz
    gunzip -c /tmp/backup.sql.gz | psql -h fleet-postgres -U fleet_user -d fleet_db -c 'DROP DATABASE IF EXISTS test_restore; CREATE DATABASE test_restore;'
    gunzip -c /tmp/backup.sql.gz | psql -h fleet-postgres -U fleet_user -d test_restore
    echo '✅ Restore successful'
  "
```

---

## Automated Scheduled Tests

The CronJob is configured to run daily at 3:00 AM (after backups complete at 2:00 AM).

### Cron Schedule

Edit `restore-test-cron.yaml`:

```yaml
spec:
  schedule: "0 3 * * *"  # Daily at 3:00 AM
  # Other options:
  # "0 3 * * 1"       # Mondays at 3:00 AM
  # "0 */6 * * *"     # Every 6 hours
  # "0 0 1 * *"       # 1st of each month
```

Apply the change:

```bash
kubectl apply -f infrastructure/backup/restore-test-cron.yaml
```

---

## Monitoring & Alerting

### Prometheus Alert Rule

Add to `infrastructure/monitoring/prometheus-values.yaml`:

```yaml
additionalPrometheusRulesMap:
  flyfleet-rules:
    groups:
      - name: backup
        rules:
          - alert: BackupVerificationFailed
            expr: increase(backup_verification_failed_total[1h]) > 0
            for: 5m
            labels:
              severity: critical
              component: backup
            annotations:
              summary: "Backup verification failed"
              description: "Database backup verification has failed within the last hour. Last error: {{ $labels.error }}"
```

If using the provided verification script with Prometheus metrics endpoint, ensure it's scraped by Prometheus.

---

### Alertmanager Notification

Configure Alertmanager to notify on `BackupVerificationFailed` alerts (see `infrastructure/monitoring/alertmanager-config.yaml`).

---

### Check Verification History

```bash
# List CronJob runs
kubectl get jobs -n flyfleet -l component=backup-verification --sort-by=.metadata.creationTimestamp

# Get logs from last successful/failed run
kubectl logs -l component=backup-verification -n flyfleet --tail=100

# Check pod exit code
kubectl get pods -n flyfleet -l component=backup-verification -o wide
```

---

## Troubleshooting

### CronJob not running

```bash
# Check CronJob status
kubectl describe cronjob backup-verification -n flyfleet

# Check if CronJob controller is running
kubectl get pods -n kube-system | grep cronjob

# Test schedule by creating manual job
kubectl create job --from=cronjob/backup-verification test-manual -n flyfleet
kubectl logs -f job/test-manual -n flyfleet
```

### Permission denied (S3 access)

```bash
# Check IAM role annotation on service account
kubectl get serviceaccount flyfleet-backup-verifier -n flyfleet -o yaml

# Verify IAM role exists and has correct policy
aws iam get-role --role-name flyfleet-backup-verifier-production
aws iam list-attached-role-policies --role-name flyfleet-backup-verifier-production

# Test from within pod (if using AWS CLI)
kubectl run -it --rm -n flyfleet s3-test --image=amazonlinux -- \
  bash -c "yum install -y aws-cli && aws s3 ls s3://flyfleet-backups-<ACCOUNT_ID>/"
```

### Restore test fails (corrupted backup)

The verification script will mark the test as failed and trigger an alert. Investigate:

1. Check backup file integrity manually
2. Try restore manually to see detailed error
3. If backup is truly corrupted, restore from previous backup
4. Notify DBA/team immediately

### Insufficient disk space for restore

The test restores to a temporary database, which uses shared storage. Ensure PostgreSQL has enough space:

```bash
# Check disk usage
kubectl exec -n flyfleet fleet-postgres -- df -h /var/lib/postgresql/data

# Clean up old test databases (script should do this automatically)
kubectl exec -n flyfleet fleet-postgres -- psql -U fleet_user -d fleet_db -c "DROP DATABASE IF EXISTS restore_test_*;"
```

---

## Best Practices

### 1. Test Frequency

| Environment | Frequency | Time |
|-------------|-----------|------|
| Production | Daily | 3:00 AM |
| Staging | Weekly | Sunday 3:00 AM |
| Development | On-demand | Manual |

### 2. Alert Response Playbook

When `BackupVerificationFailed` alert fires:

1. **Acknowledge** alert in PagerDuty/Slack
2. **Check logs** of verification job:
   ```bash
   kubectl logs -l component=backup-verification -n flyfleet --tail=50
   ```
3. **Determine cause**:
   - ❌ Backup file missing/corrupted → Notify DBA, restore from previous backup
   - ❌ Database connection issue → Check PostgreSQL pod status
   - ❌ Permission error → Check IAM role/S3 access
4. **Initiate manual restore** if needed (to dev environment first)
5. **Document** incident in runbook
6. **Fix root cause** (backup process, storage, IAM, etc.)

### 3. Retention & Cleanup

The script automatically drops the test database after verification. Ensure:

- Test databases don't accumulate (check `SELECT datname FROM pg_database WHERE datname LIKE 'restore_test_%';`)
- If leftovers exist, schedule cleanup job or adjust script

---

## Advanced: Custom Verification Script

To extend verification beyond basic restore:

```bash
#!/bin/bash
# Add to verify-backup.sh after restore_backup() completes

after_restore_checks() {
    local db_name="$1"

    # 1. Row count comparison (if you have baseline)
    local expected_drivers=1000
    local actual_drivers=$(PGPASSWORD="$DB_PASSWORD" psql -h ... -c "SELECT COUNT(*) FROM drivers" | head -2 | tail -1)
    if [ "$actual_drivers" -lt "$expected_drivers" ]; then
        echo "ERROR: Driver count mismatch (expected >$expected_drivers, got $actual_drivers)"
        return 1
    fi

    # 2. Schema validation
    PGPASSWORD="$DB_PASSWORD" psql -h ... -c "\dt" | grep -q "^ drivers |" || {
        echo "ERROR: Missing drivers table"
        return 1
    }

    # 3. Foreign key integrity
    local broken_fk=$(PGPASSWORD="$DB_PASSWORD" psql -h ... -c "
      SELECT COUNT(*) FROM information_schema.referential_constraints rc
      JOIN pg_constraint c ON rc.constraint_name = c.conname
      WHERE NOT EXISTS (
        SELECT 1 FROM pg_class t
        JOIN pg_namespace n ON t.relnamespace = n.oid
        WHERE t.relname = (string_to_array(c.conrelid::regclass::text, '.'))[2]
      );
    " | head -2 | tail -1)
    [ "$broken_fk" -eq 0 ] || {
        echo "ERROR: $broken_fk broken foreign keys"
        return 1
    }

    return 0
}

after_restore_checks "$TEST_DB_NAME" || return 1
```

---

## Uninstall

```bash
# Delete CronJob (keeps ServiceAccount and ConfigMap)
kubectl delete cronjob backup-verification -n flyfleet

# Delete all components
kubectl delete -f infrastructure/backup/restore-test-cron.yaml

# Delete ConfigMap
kubectl delete configmap backup-verification-scripts -n flyfleet

# Remove IAM role (Terraform)
terraform destroy -target=module.backup-verifier -var-file=terraform.tfvars
```

---

## Related Documentation

- **Backup cronjob:** `infrastructure/database/backup-cronjob.yaml`
- **Monitoring:** `docs/MONITORING_SETUP.md`
- **Runbook:** `docs/RUNBOOK.md` (incident response procedures)
- **Terraform IAM:** `infrastructure/terraform/modules/backup-verifier/`

---

## Support

If verification consistently fails:

1. **Check backup jobs:** Ensure daily backups are running successfully
2. **Test restore manually:** `./infrastructure/backup/verify-backup.sh`
3. **Review storage:** Ensure sufficient disk space for restore
4. **Contact DBA:** Database corruption may require point-in-time recovery

---

**Last Updated:** 2026-03-28
**Maintainer:** FlyFleet Platform Team
