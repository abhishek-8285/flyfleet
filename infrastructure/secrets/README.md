# FlyFleet Secrets Management with AWS Secrets Manager

**Production-grade secret management using External Secrets Operator**

---

## 📋 Overview

This setup replaces plain-text secrets in Helm values with **AWS Secrets Manager** integration using the [External Secrets Operator](https://external-secrets.io/) (ESO).

### Benefits

✅ **No secrets in Git** - All secrets stored encrypted in AWS Secrets Manager
✅ **Automatic rotation** - Secrets can be rotated without pod restarts
✅ **Fine-grained IAM** - Least-privilege permissions via IRSA
✅ **Audit trail** - All secret accesses logged in CloudTrail
✅ **Encryption at rest** - KMS encryption for all secrets
✅ **Consistent across environments** - Same mechanism for staging/production

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │           External Secrets Operator                │   │
│  │  ┌─────────────────────────────────────────────┐  │   │
│  │  │ Monitors ExternalSecret CRDs                │  │   │
│  │  │ • Watch for new/updated resources           │  │   │
│  │  │ • Fetch from AWS Secrets Manager            │  │   │
│  │  │ • Create/Update K8s Secrets                  │  │   │
│  │  └─────────────────────────────────────────────┘  │   │
│  └──────────────────────┬────────────────────────────┘   │
│                         │                                 │
│                         │ IRSA (IAM Role)                │
│                         ▼                                 │
│  ┌─────────────────────────────────────────────────────┐   │
│  │           AWS Secrets Manager                      │   │
│  │  ┌─────────────────────────────────────────────┐  │   │
│  │  │ /flyfleet/production/database               │  │   │
│  │  │   { "url": "...", "password": "***" }      │  │   │
│  │  │                                             │  │   │
│  │  │ /flyfleet/production/auth                  │  │   │
│  │  │   { "jwt_secret": "***" }                  │  │   │
│  │  │                                             │  │   │
│  │  │ /flyfleet/production/redis                 │  │   │
│  │  │   { "url": "redis://..." }                 │  │   │
│  │  └─────────────────────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  K8s Secret: flyfleet-secrets                              │
│  ├── DATABASE_URL = postgresql://...                       │
│  ├── DB_PASSWORD = ********                                │
│  ├── REDIS_URL = redis://...                               │
│  ├── NATS_URL = nats://...                                 │
│  └── JWT_SECRET = ********                                 │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Prerequisites

- ✅ EKS cluster with **IRSA** (OIDC provider enabled)
- ✅ kubectl configured with cluster access
- ✅ Helm 3.12+
- ✅ AWS CLI configured with appropriate permissions
- ✅ Secrets already created in AWS Secrets Manager (see below)

---

## Step 1: Create Secrets in AWS Secrets Manager

Before deploying ESO, create your secrets in AWS Secrets Manager:

```bash
# Create database secret
aws secretsmanager create-secret \
  --name "flyfleet/production/database" \
  --secret-string '{"url":"postgresql://fleet_user:YOUR_PASSWORD@your-db-host:5432/fleet_db","password":"YOUR_PASSWORD"}' \
  --region us-east-1 \
  --tags Key=Environment,Value=production Key=Project,Value=FlyFleet

# Create auth secret
aws secretsmanager create-secret \
  --name "flyfleet/production/auth" \
  --secret-string '{"jwt_secret":"YOUR_JWT_SECRET"}' \
  --region us-east-1

# Create redis secret
aws secretsmanager create-secret \
  --name "flyfleet/production/redis" \
  --secret-string '{"url":"redis://fleet-redis:6379"}' \
  --region us-east-1

# Create NATS secret
aws secretsmanager create-secret \
  --name "flyfleet/production/nats" \
  --secret-string '{"url":"nats://fleet-nats:4222"}' \
  --region us-east-1
```

**Important:**
- Use JSON format for flexibility
- Include all keys that your application expects
- Enable automatic rotation if desired (via Lambda)
- Set appropriate secret policy to allow access only from your IAM role

---

## Step 2: Create IAM Role with IRSA

Deploy the Terraform module to create the IAM role:

```bash
cd infrastructure/terraform

# Add external-secrets module to main-new.tf (see next step)

# Or apply standalone:
terraform apply -target=module.external-secrets -var-file=terraform.tfvars
```

### Update `main-new.tf` to include external-secrets module:

Add after the `module "iam"` block:

```hcl
module "external-secrets" {
  source             = "./modules/external-secrets"
  aws_region         = var.aws_region
  environment        = var.environment
  eks_oidc_issuer_url = module.kubernetes.oidc_issuer_url
  tags               = var.tags

  depends_on = [module.kubernetes]
}
```

Save the role ARN output:

```bash
terraform output external-secrets_role_arn
# Copy this value for Step 4
```

---

## Step 3: Install External Secrets Operator

```bash
# Add Helm repository
helm repo add external-secrets https://charts.external-secrets.io
helm repo update

# Create monitoring namespace (if not exists)
kubectl create namespace monitoring

# Install ESO
helm install external-secrets external-secrets/external-secrets \
  --namespace monitoring \
  --create-namespace \
  -f infrastructure/secrets/external-secrets-values.yaml \
  --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"="<EXTERNAL_SECRETS_ROLE_ARN>"
```

Replace `<EXTERNAL_SECRETS_ROLE_ARN>` with the ARN from Step 2.

Wait for pods to be ready:

```bash
kubectl wait --for=condition=ready pod -l app.kubernetes.io/instance=external-secrets -n monitoring --timeout=120s
```

---

## Step 4: Configure ClusterSecretStore

Apply the ClusterSecretStore resource:

```bash
# Edit clostersecretstore.yaml to include your AWS account ID and region
# Replace ${AWS_ACCOUNT_ID} and ${ENVIRONMENT}

kubectl apply -f infrastructure/secrets/clustersecretstore.yaml
```

**Verify:**

```bash
kubectl get clustersecretstore -n monitoring
# Should show: aws-secretsmanager   Ready
```

---

## Step 5: Deploy ExternalSecret Resources

Apply the ExternalSecret resources for FlyFleet:

```bash
kubectl apply -f infrastructure/secrets/externalsecrets-flyfleet.yaml
```

This creates Kubernetes secret `flyfleet-secrets` in the `flyfleet` namespace.

**Verify:**

```bash
# Check ExternalSecret status
kubectl get externalsecret -n flyfleet

# Should show: READY = True, AGE

# Check the generated K8s secret
kubectl get secret flyfleet-secrets -n flyfleet -o yaml

# View secret data (base64 decoded)
kubectl get secret flyfleet-secrets -n flyfleet -o jsonpath="{.data.DATABASE_URL}" | base64 --decode
```

---

## Step 6: Update FlyFleet Deployment

Update your Helm deployment to **remove** the plaintext secrets and use the new K8s secret:

### Before (old values.yaml):

```yaml
global:
  postgres:
    password: "fleet_password"  # ❌ Plain text in Git!
  jwtSecret: "dev-secret"       # ❌ Plain text in Git!
```

### After (new values.yaml):

```yaml
global:
  # Remove plaintext passwords

  # Reference K8s secret instead
  database:
    secretName: "flyfleet-secrets"

  # The application will read env vars from the pod spec
  # (injected via deployment template)

fleet-core-service:
  env:
    # These will be populated from the secret
    - name: DATABASE_URL
      valueFrom:
        secretKeyRef:
          name: flyfleet-secrets
          key: DATABASE_URL
    - name: JWT_SECRET
      valueFrom:
        secretKeyRef:
          name: flyfleet-secrets
          key: JWT_SECRET
```

**Or**, if using the existing `services.yaml` template, it already supports `database.secretName`. Update your values:

```yaml
global:
  database:
    secretName: "flyfleet-secrets"
    # Remove host/port/password/username from here

  jwtSecret: ""  # Remove plaintext
```

Then redeploy:

```bash
helm upgrade --install flyfleet infrastructure/helm/flyfleet \
  --namespace flyfleet \
  --set global.database.secretName="flyfleet-secrets" \
  --set global.jwtSecret="" \
  ...
```

---

## Step 7: Verify Application Works

```bash
# Check pods are running with secrets mounted
kubectl get pods -n flyfleet
kubectl describe pod <flyfleet-pod> -n flyfleet | grep -A5 "Environment"

# Check logs (no password leakage!)
kubectl logs -n flyfleet deployment/flyfleet-fleet-core-service

# Test health endpoint
kubectl port-forward -n flyfleet svc/flyfleet-fleet-core-service 8084:8084 &
curl http://localhost:8084/health

# Should return: {"status":"ok"}
```

---

## Rotation & Maintenance

### Rotate Secrets

1. Update secret in AWS Secrets Manager:

```bash
aws secretsmanager put-secret-value \
  --secret-id flyfleet/production/database \
  --secret-string '{"url":"postgresql://fleet_user:NEW_PASSWORD@...","password":"NEW_PASSWORD"}' \
  --region us-east-1
```

2. External Secrets Operator will automatically sync within `refreshInterval` (default 1h).

3. To force immediate refresh:

```bash
# Annotate the ExternalSecret to trigger refresh
kubectl annotate externalsecret flyfleet-secrets -n flyfleet \
  external-secrets.io/refresh="now"

# Or delete pods to force reload (not recommended for production)
kubectl rollout restart deployment/flyfleet-fleet-core-service -n flyfleet
```

### View Secret History

```bash
aws secretsmanager describe-secret --secret-id flyfleet/production/database
aws secretsmanager get-secret-value --secret-id flyfleet/production/database --version-stage AWSCURRENT
aws secretsmanager get-secret-value --secret-id flyfleet/production/database --version-stage AWSPREVIOUS
```

### Recover from Failed Rotation

If a secret update breaks the app:

```bash
# Revert to previous version
kubectl edit externalsecret flyfleet-secrets -n flyfleet
# Change property to use AWSPREVIOUS (requires custom fetch logic)

# Or manually set secret value back
aws secretsmanager put-secret-value --secret-id ... --secret-string '<previous-value>'
```

---

## Advanced Configuration

### Different Secrets Per Environment

Create separate ExternalSecret resources for staging/production:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: flyfleet-secrets-staging
  namespace: flyfleet-staging
spec:
  refreshInterval: "1h"
  secretStoreRef:
    name: aws-secretsmanager
    kind: ClusterSecretStore
  target:
    name: flyfleet-secrets
  data:
    - secretKey: DATABASE_URL
      remoteRef:
        key: "flyfleet/staging/database"
        property: "url"
```

### Using Parameter Store (SSM) instead of Secrets Manager

Update ClusterSecretStore:

```yaml
spec:
  provider:
    aws:
      service: ParameterStore
      region: us-east-1
```

Then use SSM parameter path:

```yaml
data:
  - secretKey: DATABASE_URL
    remoteRef:
      key: "/flyfleet/production/database/url"  # SSM parameter name
```

### Secret Templates (Transformations)

Apply transformations before creating K8s secret:

```yaml
target:
  template:
    data:
      # Transform db_url property to DATABASE_URL key
      DATABASE_URL: "{{ .db_url | toString }}"
      # Obfuscate password in logs
      DB_PASSWORD: "{{ .password | b64enc }}"
```

---

## Security Best Practices

1. **Restrict IAM permissions** - Only grant access to specific secret paths
2. **Enable CloudTrail logging** - Audit all Secrets Manager access
3. **Use KMS encryption keys** with key policies limiting decryption
4. **Rotate secrets regularly** (90 days or less)
5. **Use different secrets per environment** (staging vs production)
6. **Never log secret values** - Ensure application doesn't print them
7. **Enable secret deletion protection** in AWS Secrets Manager
8. **Set up SCP/IAM boundaries** to prevent secret exfiltration

---

## Troubleshooting

### ExternalSecret not syncing

```bash
# Check ExternalSecret status
kubectl describe externalsecret flyfleet-secrets -n flyfleet

# Look for conditions:
# - Ready: True/False
# - SecretSynced: True/False
# - Last synchronization

# Check ESO pod logs
kubectl logs -n monitoring -l app.kubernetes.io/instance=external-secrets

# Common errors:
# - AccessDenied: IAM role missing permissions
# - ResourceNotFoundException: Secret doesn't exist in AWS
# - InvalidRequest: KMS key not accessible
```

### IAM permissions denied

```bash
# Verify role is assumed correctly
kubectl get pods -n monitoring -o yaml | grep -A5 "serviceAccountName"

# Test IAM role (from within cluster)
kubectl run -it --rm -n monitoring secrets-test --image=amazonlinux -- \
  bash -c "yum install -y aws-cli && aws sts get-caller-identity"

# Check role permissions
aws iam simulate-principal-policy \
  --policy-source-arn arn:aws:iam::<ACCOUNT>:role/flyfleet-external-secrets-production \
  --action-names secretsmanager:GetSecretValue
```

### Secret not updating in pod

External Secrets Operator updates the K8s secret, but **pods don't automatically reload**.

Solutions:

1. **Use environment variable references** (k8s 1.22+): Pods automatically get updated when secret changes

2. **Restart deployment** on secret update:

```bash
# Add annotation to ExternalSecret:
# metadata:
#   annotations:
#     external-secrets.io/refresh: "always"

# Or use post-sync hook:
kubectl annotate externalsecret flyfleet-secrets -n flyfleet \
  "helm.sh/hook": post-sync,post-delete
```

3. **Use secret volume mounts** with `optional: true` and periodic reload

---

## Migration from Plaintext Secrets

### Phase 1: Deploy ESO and create AWS secrets (without removing old ones)

1. Follow Steps 1-5 above
2. Verify `flyfleet-secrets` K8s secret exists
3. Deploy with **both** old and new secret references
4. Application picks up new secrets gradually (first from old, then new if missing)

### Phase 2: Update application to use new secret

1. Update values.yaml to use `flyfleet-secrets`
2. Redeploy Helm
3. Verify application still works
4. Test one secret at a time (e.g., only JWT_SECRET from new secret)

### Phase 3: Remove old plaintext values

1. Remove all plaintext `postgres.password`, `jwtSecret`, etc. from values.yaml
2. Final Helm upgrade
3. Test thoroughly

### Phase 4: Cleanup (optional)

1. Delete old kubernetes secrets that referenced plaintext values
2. Remove old values from Git configuration

---

## Monitoring ESO Health

```bash
# Check ESO deployment status
kubectl get deployment external-secrets -n monitoring

# Check metrics (if enabled)
kubectl port-forward -n monitoring svc/prometheus-operated 9090:9090 &
curl http://localhost:9090/api/v1/query?query=external_secrets_sync_total

# Grafana dashboard: Import monitoring dashboards from:
# https://github.com/external-secrets/external-secrets/tree/main/deploy/ GrafanaDashboard
```

---

## Uninstall

```bash
# Delete ExternalSecret resources first (keeps AWS secrets)
kubectl delete externalsecret -n flyfleet --all

# Delete ClusterSecretStore
kubectl delete clustersecretstore aws-secretsmanager

# Uninstall Helm chart (keeps K8s secrets)
helm uninstall external-secrets -n monitoring

# Delete K8s secrets created by ESO (if desired)
kubectl delete secret flyfleet-secrets -n flyfleet

# Destroy IAM role (Terraform)
terraform destroy -target=module.external-secrets -var-file=terraform.tfvars

# (Optional) Delete AWS Secrets Manager secrets
aws secretsmanager delete-secret --secret-id flyfleet/production/database --force-delete-without-recovery
```

---

## Next Steps

After Secrets Management is production-ready:

1. **Set up secret rotation** - Lambda functions for automatic rotation
2. **Enable audit logging** - CloudTrail insights for secret access
3. **Implement secret scanning** - Prevent accidental commits
4. **Create runbook** - Document secret recovery procedures
5. **Add to CI/CD** - Automated secret validation tests

---

**Last Updated:** 2026-03-28
**Maintainer:** FlyFleet Platform Team
