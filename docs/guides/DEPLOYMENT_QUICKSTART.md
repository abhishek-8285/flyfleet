# FlyFleet Production Deployment Quickstart

**One-page guide to deploy FlyFleet to production AWS**

---

## 📦 What You're Deploying

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Infrastructure** | Terraform (modular) | AWS resources (VPC, EKS, RDS, IAM) |
| **Container Orchestration** | EKS (Kubernetes) | Application deployment |
| **Database** | RDS PostgreSQL 15 | Primary data store |
| **Messaging** | NATS JetStream | Event streaming |
| **Cache** | ElastiCache Redis | Session/cache layer |
| **CI/CD** | GitHub Actions | Automated deployments |
| **Monitoring** | Prometheus + Grafana | Metrics & dashboards |
| **Secrets** | AWS Secrets Manager | Encrypted secret storage |
| **Backup** | CronJob + S3 | Automated daily backups + restore testing |

---

## ⏱️ Deployment Timeline

| Phase | Duration | Dependencies |
|-------|----------|--------------|
| Terraform Init & Apply | 15-30 min | AWS account, PostgreSQL state DB |
| Build Docker Images | 10-20 min | Go build, test pass |
| Helm Deploy | 5-10 min | ECR images pushed |
| Monitoring Setup | 5 min | EKS cluster running |
| Secrets Migration | 30 min | AWS Secrets Manager populated |
| Verification | 10 min | All above complete |
| **Total** | **~1 hour** | - |

---

## 📋 Pre-Flight Checklist

### AWS Account

- [ ] AWS account with appropriate permissions (Administrator or custom policy)
- [ ] Region selected (e.g., `us-east-1`)
- [ ] Access keys configured (`aws configure`)
- [ ] S3 bucket for Terraform state (optional - can use PostgreSQL)
- [ ] S3 bucket for backups (recommended)

### Infrastructure

- [ ] PostgreSQL database for Terraform state (can use existing `fleet_postgres`)
- [ ] `terraform.tfvars` file created with secure values
- [ ] KMS keys will be auto-created

### Application

- [ ] `fleet-core-service` builds successfully: `make build`
- [ ] All tests passing: `make test`
- [ ] Docker image ready: `docker build -t fleet-core .`

### GitHub Repository

- [ ] Repository secrets configured:
  - `AWS_ACCESS_KEY_ID` (or use OIDC)
  - `AWS_SECRET_ACCESS_KEY` (or use OIDC)
  - `JWT_SECRET` (random 32+ char string)
  - `DB_PASSWORD` (secure password)
- [ ] OIDC role created in IAM (optional but recommended)

---

## 🚀 Step-by-Step Deployment

### Step 1: Prepare Terraform State

```bash
cd infrastructure/terraform

# If using existing PostgreSQL for state:
docker exec fleet-postgres psql -U fleet_user -d fleet_db -c "CREATE DATABASE terraform_state;"

# Or use separate RDS instance:
aws rds create-database-instance --db-instance-identifier terraform-state ...
```

---

### Step 2: Configure & Initialize Terraform

```bash
# Copy and edit terraform.tfvars
cp terraform.tfvars.example terraform.tfvars

# Edit these variables:
# - db_password = (generate: openssl rand -base64 32)
# - admin_cidr = (your IP, e.g., "203.0.113.0/24")
# - availability_zones = ["us-east-1a", "us-east-1b"]

# Initialize with PostgreSQL backend
terraform init \
  -backend-config="conn_str=postgresql://fleet_user:fleet_password@localhost:5432/terraform_state?sslmode=disable"
```

---

### Step 3: Deploy AWS Infrastructure

```bash
terraform plan -var-file=terraform.tfvars -out=tfplan
terraform apply tfplan
```

**Time:** 15-30 minutes

**Save outputs:**
```bash
terraform output -json > ../tf-outputs.json
# Key outputs: vpc_id, eks_cluster_endpoint, db_endpoint, ecr_repository_urls
```

---

### Step 4: Configure kubectl for EKS

```bash
aws eks update-kubeconfig \
  --region us-east-1 \
  --name flyfleet-production  # or flyfleet-staging

# Verify connection
kubectl get nodes
# Should show your EKS node group nodes
```

---

### Step 5: Build & Push Docker Images

```bash
# Login to ECR
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
aws ecr get-login-password | docker login --username AWS --password-stdin ${ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com

# Build
docker build -t ${ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/flyfleet/fleet-core-service:v1.0.0 \
  -f infrastructure/docker/GoService.Dockerfile \
  --build-arg SERVICE_NAME=fleet-core .

# Push
docker push ${ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/flyfleet/fleet-core-service:v1.0.0

# Tag as latest
docker tag ${ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/flyfleet/fleet-core-service:v1.0.0 \
  ${ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/flyfleet/fleet-core-service:latest
docker push ${ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/flyfleet/fleet-core-service:latest
```

---

### Step 6: Deploy FlyFleet Application

```bash
# Get IAM role ARN for fleet-core-service
export ENVIRONMENT=production
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
ROLE_ARN="arn:aws:iam::${AWS_ACCOUNT_ID}:role/flyfleet-fleet-core-service-sa-role-${ENVIRONMENT}"

# Deploy Helm chart
helm upgrade --install flyfleet infrastructure/helm/flyfleet \
  --namespace flyfleet \
  --create-namespace \
  --set global.imageRegistry="${ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/flyfleet" \
  --set global.imageTag="v1.0.0" \
  --set global.environment="${ENVIRONMENT}" \
  --set global.jwtSecret="$(openssl rand -base64 32)" \
  --set global.postgres.password="YOUR_DB_PASSWORD" \
  --set iam.annotations."eks\.amazonaws\.com/role-arn"="${ROLE_ARN}" \
  --timeout 10m \
  --wait
```

---

### Step 7: Verify Deployment

```bash
# Check pods
kubectl get pods -n flyfleet
# All should be Running

# Check services
kubectl get svc -n flyfleet

# Check ingress (if configured)
kubectl get ingress -n flyfleet

# Wait for rollout to complete
kubectl rollout status deployment/flyfleet-fleet-core-service -n flyfleet --timeout=300s

# Test health endpoint
kubectl port-forward -n flyfleet svc/flyfleet-fleet-core-service 8084:8084 &
sleep 3
curl http://localhost:8084/health
# Should return: {"status":"ok","timestamp":"..."}
```

---

### Step 8: Deploy Observability Stack

```bash
# Add Prometheus repo
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Deploy kube-prometheus-stack
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring --create-namespace \
  -f infrastructure/monitoring/prometheus-values.custom.yaml

# Wait for pods
kubectl wait --for=condition=ready pod -l app.kubernetes.io/instance=prometheus -n monitoring --timeout=300s

# Access Grafana
kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80
# Open http://localhost:3000, login: admin / (password from values.yaml)

# Import FlyFleet dashboard
# In Grafana: Dashboards → Import → paste contents of grafana-dashboard-flyfleet.json
```

---

### Step 9: Set Up Secrets Management (Optional but Recommended)

```bash
# 1. Create IAM role via Terraform
cd infrastructure/terraform
terraform apply -target=module.external-secrets -var-file=terraform.tfvars
ROLE_ARN=$(terraform output -module=external-secrets external_secrets_role_arn)

# 2. Install External Secrets Operator
helm install external-secrets external-secrets/external-secrets \
  --namespace monitoring \
  -f infrastructure/secrets/external-secrets-values.yaml \
  --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"="${ROLE_ARN}"

# 3. Create secrets in AWS Secrets Manager
aws secretsmanager create-secret --name "flyfleet/production/database" --secret-string '{"url":"postgresql://...","password":"..."}'
aws secretsmanager create-secret --name "flyfleet/production/auth" --secret-string '{"jwt_secret":"..."}'
# ... etc

# 4. Apply ClusterSecretStore and ExternalSecret
kubectl apply -f infrastructure/secrets/

# 5. Update Helm to use K8s secrets
helm upgrade flyfleet infrastructure/helm/flyfleet \
  --namespace flyfleet \
  --set global.database.secretName="flyfleet-secrets" \
  --set global.jwtSecret="" \
  --wait
```

---

### Step 10: Deploy Backup Verification

```bash
# 1. Create S3 bucket for backups (if using S3)
aws s3 mb s3://flyfleet-backups-${AWS_ACCOUNT_ID}
aws s3api put-bucket-versioning --bucket flyfleet-backups-${AWS_ACCOUNT_ID} --versioning-configuration Status=Enabled

# 2. Create backup PVC (if using local storage)
kubectl apply -f - <<EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-backup-pvc
  namespace: flyfleet
spec:
  accessModes: ["ReadWriteMany"]
  storageClassName: "standard"
  resources:
    requests:
      storage: 100Gi
EOF

# 3. Deploy IAM role (Terraform)
terraform apply -target=module.backup-verifier -var-file=terraform.tfvars

# 4. Deploy verification CronJob
kubectl create configmap backup-verification-scripts \
  --from-file=verify-backup.sh=infrastructure/backup/verify-backup.sh \
  -n flyfleet

kubectl apply -f infrastructure/backup/restore-test-cron.yaml

# 5. Manual test run
kubectl create job --from=cronjob/backup-verification backup-test-manual -n flyfleet
kubectl logs -f job/backup-test-manual -n flyfleet
```

---

### Step 11: Enable GitHub Actions CI/CD (Optional)

Update `.github/workflows/deploy.yml`:

```yaml
- name: Configure AWS credentials
  uses: aws-actions/configure-aws-credentials@v4
  with:
    role-to-assume: arn:aws:iam::${{ secrets.AWS_ACCOUNT_ID }}:role/flyfleet-github-actions-deploy-production
    aws-region: us-east-1
```

Create OIDC role in IAM (Terraform already creates this when `github_org` and `github_repo` are set).

---

## ✅ Post-Deployment Checklist

### Health Checks

```bash
# API health
curl http://<your-domain>/health

# Pods running
kubectl get pods -n flyfleet
# All pods: Running, 0/1 restarts

# Database connectivity
kubectl exec -n flyfleet deploy/flyfleet-fleet-core-service -- \
  pg_isready -h fleet-postgres -p 5432 -U fleet_user -d fleet_db

# Prometheus targets
curl http://<prometheus-url>/api/v1/targets | jq '.data.activeTargets[] | .labels.job + ":" + .health'
# All should be "up"

# Grafana dashboards
# Open Grafana → FlyFleet Overview → All panels showing data
```

### Monitoring

- [ ] All Prometheus targets up
- [ ] Grafana dashboards populated
- [ ] Alertmanager notifications configured (send test alert)
- [ ] Loki/Elasticsearch for logs (optional)
- [ ] Metrics being scraped from fleet-core-service

### Security

- [ ] All secrets in AWS Secrets Manager (no plaintext in Git)
- [ ] IAM roles using IRSA (no static credentials in pods)
- [ ] EKS private endpoint only
- [ ] RDS in private subnets
- [ ] Security groups least privilege
- [ ] KMS encryption enabled for all sensitive data
- [ ] VPC Flow Logs enabled

### Reliability

- [ ] Database backups running (check S3 or PVC)
- [ ] Backup verification CronJob succeeded at least once
- [ ] HPA enabled and scaling metrics working
- [ ] Multi-AZ deployment verified (simulate AZ failure)
- [ ] Rollback tested: `helm rollback flyfleet 1`

### Operations

- [ ] Runbooks documented for common incidents
- [ ] PagerDuty/Opsgenie configured
- [ ] On-call rotation established
- [ ] Access logs being stored (CloudFront/ELB/Nginx)
- [ ] Cost monitoring alerts configured (AWS Budgets)

---

## 🆘 Troubleshooting

### Terraform fails to apply

- Check IAM permissions: `aws sts get-caller-identity`
- Verify PostgreSQL backend DB exists and user has privileges
- Check variable values in `terraform.tfvars`

### Helm deployment stuck

- Check pod events: `kubectl describe pod <pod> -n flyfleet`
- Verify image exists in ECR: `aws ecr describe-images --repository-name flyfleet/fleet-core-service`
- Ensure service account role exists: `kubectl get sa -n flyfleet`

### API not responding

- Check pod logs: `kubectl logs -n flyfleet deployment/flyfleet-fleet-core-service`
- Verify environment variables: `kubectl exec -n flyfleet <pod> -- env | grep -E "DATABASE|JWT|NATS"`
- Check database connectivity: `kubectl exec -n flyfleet <pod> -- nc -zv fleet-postgres 5432`

### Prometheus no metrics

- Verify ServiceMonitor exists: `kubectl get servicemonitor -n monitoring`
- Check Prometheus targets: `kubectl port-forward -n monitoring svc/prometheus-operated 9090:9090`, then open http://localhost:9090/targets
- Ensure application has Prometheus annotations: `prometheus.io/scrape: "true"`

---

## 📞 Emergency Contacts

| Issue | First Step | Escalation |
|-------|------------|------------|
| API down | `kubectl logs -n flyfleet deploy/flyfleet-fleet-core-service` | Rollback: `kubectl rollout undo deployment/flyfleet-fleet-core-service -n flyfleet` |
| DB down | Check RDS console → reboot if needed | Restore from backup (see RUNBOOK.md) |
| Deploy failed | `helm status flyfleet -n flyfleet` | `helm rollback flyfleet 1` |
| Monitoring down | `kubectl get pods -n monitoring` | Restart Prometheus pods |
| Secrets broken | Check ExternalSecret status: `kubectl get externalsecret -n flyfleet` | Roll back secret values in AWS Secrets Manager |

---

## 📚 Documentation Index

| Document | Purpose |
|----------|---------|
| `docs/DEVOPS_FIXES_SUMMARY.md` | Complete infrastructure overview and fixes |
| `docs/SELF_HOSTED_GUIDE.md` | Local development deployment |
| `docs/QUICK_REFERENCE.md` | Command cheat sheet |
| `docs/MONITORING_SETUP.md` | Observability deployment and operations |
| `infrastructure/secrets/README.md` | Secrets management guide |
| `infrastructure/backup/README.md` | Backup and restore procedures |
| `tests/README.md` | Infrastructure testing framework |
| `docs/RUNBOOK.md` | Incident response procedures |

---

## 🎉 Success Criteria

Your production deployment is complete when:

- ✅ EKS cluster running with 2+ nodes
- ✅ RDS PostgreSQL accessible from pods
- ✅ fleet-core-service pods running and healthy
- ✅ API responds to `/health` with 200
- ✅ Prometheus scraping metrics (targets: up)
- ✅ Grafana dashboard shows data
- ✅ Secrets loaded from AWS Secrets Manager (no plaintext in Git)
- ✅ Daily backups verified with restore testing
- ✅ CI/CD pipeline auto-deploying on main branch
- ✅ Alertmanager sending test notifications
- ✅ All infrastructure tests passing

---

**Good luck with your production deployment! 🚀**

If you encounter issues, consult the troubleshooting sections in each guide or create a GitHub issue.

---

**Last Updated:** 2026-03-28
**Maintainers:** FlyFleet Platform Team
