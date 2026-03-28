# FlyFleet DevOps - Quick Reference

## 🎯 Choose Your Deployment Mode

| Mode | Use When | Command |
|------|----------|---------|
| **Self-Hosted** | Local dev, no cloud, fully open source | `make self-hosted` |
| **AWS Cloud** | Production, need managed services | `cd infrastructure/terraform && terraform apply` |

---

## 🏠 Self-Hosted (k3s) - Quick Start

### Prerequisites Check
```bash
docker --version    # Docker installed
kubectl cluster-info # k3s running
helm version        # Helm 3+
go version          # Go 1.26+
```

### One-Command Setup
```bash
make self-hosted
```

That's it! The script will:
1. Verify docker-compose services (Postgres/Redis/NATS)
2. Deploy local Docker registry to k3s
3. Build fleet-core-service Docker image
4. Push to local registry
5. Deploy via Helm

### Access the API
```bash
make self-hosted-portforward
# Then: curl http://localhost:8084/health
```

### View Logs
```bash
make self-hosted-logs
```

### Clean Up
```bash
make self-hosted-clean
```

---

## ☁️ AWS Cloud - Quick Start

### 1. Initialize State Backend
```bash
cd infrastructure/terraform

# Create terraform_state database in existing postgres
docker exec fleet-postgres psql -U fleet_user -d fleet_db -c "CREATE DATABASE terraform_state;"

# Copy and edit variables
cp terraform.tfvars.example terraform.tfvars
# Edit: set db_password, availability_zones, etc.

# Initialize Terraform
terraform init \
  -backend-config="conn_str=postgresql://fleet_user:fleet_password@localhost:5432/terraform_state?sslmode=disable"
```

### 2. Deploy Infrastructure
```bash
terraform plan -var-file=terraform.tfvars
terraform apply -var-file=terraform.tfvars
# Type 'yes' to confirm
```

**Wait 15-30 minutes** (EKS takes time)

### 3. Configure kubectl
```bash
aws eks update-kubeconfig --name flyfleet-production --region us-east-1
```

### 4. Build and Push Images
```bash
# Build fleet-core
docker build -t <account-id>.dkr.ecr.us-east-1.amazonaws.com/flyfleet/fleet-core-service:v1.0.0 \
  -f infrastructure/docker/GoService.Dockerfile \
  --build-arg SERVICE_NAME=fleet-core \
  services/fleet-core

# Push to ECR (Terraform output shows ECR URLs)
aws ecr get-login-password | docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com
docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/flyfleet/fleet-core-service:v1.0.0
```

### 5. Deploy to EKS
```bash
helm upgrade --install flyfleet infrastructure/helm/flyfleet \
  --namespace flyfleet \
  --create-namespace \
  --set global.imageRegistry="<account-id>.dkr.ecr.us-east-1.amazonaws.com/flyfleet" \
  --set global.imageTag="v1.0.0" \
  --set global.postgres.host="<rds-endpoint>" \
  --set global.postgres.password="your-db-password" \
  --set global.jwtSecret="your-jwt-secret" \
  --wait
```

### 6. Access
```bash
kubectl get svc -n flyfleet
# Look for service with LoadBalancer type, get EXTERNAL-IP
# Or port-forward:
kubectl port-forward -n flyfleet svc/fleet-core-service 8084:8084
```

---

## 📁 Key Files

| Purpose | File |
|---------|------|
| Terraform config | `infrastructure/terraform/main-new.tf` |
| Helm chart | `infrastructure/helm/flyfleet/` |
| Self-hosted setup | `infrastructure/self-hosted/scripts/setup.sh` |
| CI/CD pipeline | `.github/workflows/deploy.yml` |
| Full docs | `SELF_HOSTED_GUIDE.md` |
| Summary | `DEVOPS_FIXES_SUMMARY.md` |

---

## 🆘 Common Commands

### Kubernetes
```bash
kubectl get pods -n flyfleet                    # List pods
kubectl logs -n flyfleet deployment/flyfleet-fleet-core-service -f  # Tail logs
kubectl describe pod <pod-name> -n flyfleet     # Debug pod
kubectl exec -it -n flyfleet <pod-name> -- sh   # Shell into pod
kubectl port-forward -n flyfleet svc/fleet-core-service 8084:8084  # Expose
```

### Helm
```bash
helm list -n flyfleet                            # List releases
helm status flyfleet -n flyfleet                 # Show release status
helm upgrade --install flyfleet ...              # Deploy/upgrade
helm uninstall flyfleet -n flyfleet              # Remove
helm get values flyfleet -n flyfleet             # Show applied values
```

### Docker
```bash
docker build -t <registry>/<image>:<tag> .      # Build
docker push <registry>/<image>:<tag>            # Push
docker images | grep flyfleet                   # List images
docker system prune -a                          # Clean up
```

### Terraform
```bash
terraform init                                  # Initialize backend
terraform plan -var-file=terraform.tfvars      # Preview changes
terraform apply -var-file=terraform.tfvars     # Apply changes
terraform destroy -var-file=terraform.tfvars   # Destroy all
terraform state list                            # Show state resources
```

---

## 🐛 Troubleshooting

### Self-Hosted: Registry connection refused
```bash
# Add to /etc/docker/daemon.json:
{
  "insecure-registries": ["localhost:5000"]
}
sudo systemctl restart docker
```

### Self-Hosted: Pods stuck in Pending
```bash
kubectl apply -f infrastructure/self-hosted/k3s/storageclass.yaml
kubectl get storageclass  # Should show k3s-local-path as default
```

### AWS: Terraform init fails with S3 error
**We're using PostgreSQL now, not S3.** Check:
- PostgreSQL is running and `terraform_state` database exists
- Connection string in `backend.tf` is correct
- Use `-backend-config="conn_str=..."` during init

### AWS: EKS cluster creation times out
- Check AWS Console → EKS → Cluster → "Creation failed" for details
- Common cause: hitting service quota (EKS clusters per region)
- Common cause: VPC/subnet configuration issues

### AWS: ImagePullBackOff
```bash
# Check image exists in ECR
aws ecr list-images --repository-name fleet-core-service

# Verify ECR login
aws ecr get-login-password | docker login --username AWS --password-stdin <account>.dkr.ecr.region.amazonaws.com

# Check IAM role (IRSA) permissions
kubectl describe sa fleet-core-service -n flyfleet
```

---

## 💡 Tips

1. **Use Workspaces** for dev/staging/prod environments:
   ```bash
   cd infrastructure/terraform
   terraform workspace new staging
   terraform apply -var-file="staging.tfvars"
   ```

2. **Keep Terraform state safe**: Regularly backup the `terraform_state` database:
   ```bash
   docker exec fleet-postgres pg_dump -U fleet_user terraform_state > backup_$(date +%Y%m%d).sql
   ```

3. **Use semantic image tags** for production: `v1.0.0`, `v1.0.1`, not `latest`

4. **Monitor costs**: Set up AWS Budgets or local resource monitoring

5. **Secrets management**: For production, replace `jwtSecret` in values.yaml with Kubernetes Secret:
   ```bash
   kubectl create secret generic flyfleet-secrets -n flyfleet \
     --from-literal=jwt-secret='your-secret' \
     --from-literal=postgres-password='your-password'
   ```

---

## 📚 More Documentation

- **Full Guide**: `SELF_HOSTED_GUIDE.md` (self-hosted)
- **Architecture**: `ARCHITECTURE_ASSESSMENT.md`
- **Dev Guide**: `CLAUDE.md`
- **Terraform README**: `infrastructure/terraform/README.md`
- **Helm README**: `infrastructure/self-hosted/k3s/README.md`

---

**Last Updated:** 2026-03-27
**Status:** ✅ All DevOps infrastructure in place
