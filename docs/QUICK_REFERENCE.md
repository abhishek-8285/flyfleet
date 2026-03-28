# FlyFleet DevOps Quick Reference

**One-page cheat sheet for common commands**

Last updated: 2026-03-28

---

## 📦 Table of Contents

- [Terraform](#terraform)
- [Helm](#helm)
- [Kubernetes](#kubernetes)
- [Docker & ECR](#docker--ecr)
- [k3s](#k3s)
- [Makefile](#makefile)
- [Self-Hosted](#self-hosted)
- [Debugging](#debugging)

---

## Terraform

### Initialization

```bash
# Init with PostgreSQL backend
cd infrastructure/terraform
terraform init \
  -backend-config="conn_str=postgresql://fleet_user:fleet_password@localhost:5432/terraform_state?sslmode=disable"

# Verify backend
terraform init -backend-config=... -reconfigure
```

### Planning & Applying

```bash
# Validate syntax
terraform validate

# Format files
terraform fmt -recursive

# Plan (preview changes)
terraform plan -var-file=terraform.tfvars -out=tfplan

# Apply (execute changes)
terraform apply tfplan

# Destroy (tear down)
terraform destroy -var-file=terraform.tfvars
```

### State Management

```bash
# List resources in state
terraform state list

# Show specific resource
terraform state show module.networking.aws_vpc.main

# Pull latest state (before editing)
terraform state pull > state-backup.json

# Remove resource from state (keep in AWS)
terraform state rm aws_s3_bucket.mybucket

# Import existing resource
terraform import module.database.aws_db_instance.flyfleet_postgres i-1234567890abcdef0
```

### Outputs

```bash
# All outputs
terraform output

# Specific output
terraform output vpc_id
terraform output -json eks_cluster_endpoint

# Save outputs to file
terraform output -json > tf-outputs.json
```

### Modules

```bash
# List modules
find modules -type d -mindepth 1 -maxdepth 1

# Check module dependencies
terraform graph | dot -Tpng > graph.png
```

### Debugging

```bash
# Detailed logs
TF_LOG=DEBUG terraform plan

# Trace all HTTP requests
TF_LOG=TRACE terraform apply

# Enable backup before destructive ops
export TF_BACKUP_DIR=~/.terraform/backups
```

---

## Helm

### Repository Management

```bash
# Add repo
helm repo add flyfleet ./infrastructure/helm/flyfleet
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts

# Update repos
helm repo update

# List repos
helm repo list

# Search charts
helm search repo flyfleet
helm search repo prometheus
```

### Installation & Upgrades

```bash
# Install chart
helm install flyfleet infrastructure/helm/flyfleet \
  --namespace flyfleet \
  --create-namespace

# Install with values
helm install flyfleet infrastructure/helm/flyfleet \
  -f values.yaml \
  -f values-prod.yaml \
  --namespace flyfleet

# Upgrade (with wait)
helm upgrade flyfleet infrastructure/helm/flyfleet \
  --namespace flyfleet \
  --set global.imageTag=v1.2.3 \
  --wait --timeout 10m

# Rollback
helm rollback flyfleet 1  # To revision 1

# Uninstall (with cleanup)
helm uninstall flyfleet --keep-history
kubectl delete pvc -l app.kubernetes.io/instance=flyfleet  # Manual PVC cleanup
```

### Template & Lint

```bash
# Render templates (dry-run)
helm template flyfleet infrastructure/helm/flyfleet \
  -f values.yaml \
  --debug \
  --output-dir ./rendered

# Lint chart
helm lint infrastructure/helm/flyfleet
helm lint infrastructure/helm/flyfleet -f values.yaml

# Validate chart syntax
helm install --dry-run flyfleet infrastructure/helm/flyfleet
```

### Values Management

```bash
# Inspect values
helm show values infrastructure/helm/flyfleet

# Export current release values
helm get values flyfleet -n flyfleet -a > current-values.yaml

# Compare values against defaults
helm diff upgrade flyfleet infrastructure/helm/flyfleet -f values.yaml
# Requires: helm plugin install https://github.com/databus23/helm-diff.git
```

### History & Releases

```bash
# List releases
helm list -A
helm list -n flyfleet

# Get release status
helm status flyfleet -n flyfleet

# Get release history
helm history flyfleet -n flyfleet

# Get release notes
helm get notes flyfleet -n flyfleet

# Get all release resources
helm get all flyfleet -n flyfleet
```

---

## Kubernetes

### General

```bash
# Get cluster info
kubectl cluster-info
kubectl version --short

# Get nodes
kubectl get nodes -o wide
kubectl describe node <node-name>

# Get namespaces
kubectl get ns
kubectl get ns -o wide

# Config check
kubectl config view
kubectl config current-context
kubectl get contexts
kubectl config use-context <context>
```

### Pods

```bash
# List pods
kubectl get pods -n flyfleet
kubectl get pods -n flyfleet -o wide
kubectl get pods -n flyfleet -o jsonpath="{.items[*].status.phase}"

# Describe pod
kubectl describe pod <pod-name> -n flyfleet

# Pod logs
kubectl logs <pod-name> -n flyfleet
kubectl logs -f <pod-name> -n flyfleet          # Follow
kubectl logs -p <pod-name> -n flyfleet          # Previous instance
kubectl logs --tail=100 <pod-name> -n flyfleet

# Execute command
kubectl exec -it <pod-name> -n flyfleet -- /bin/bash
kubectl exec <pod-name> -n flyfleet -- curl http://localhost:8084/health

# Delete pod (will be recreated by deployment)
kubectl delete pod <pod-name> -n flyfleet

# Port-forward
kubectl port-forward -n flyfleet svc/fleet-core-service 8084:8084
kubectl port-forward -n flyfleet pod/<pod-name> 8084:8084
```

### Deployments

```bash
# List deployments
kubectl get deploy -n flyfleet
kubectl get deploy -n flyfleet -o wide

# Describe deployment
kubectl describe deployment <deployment-name> -n flyfleet

# Scale
kubectl scale deployment <deployment-name> --replicas=5 -n flyfleet

# Rollout status
kubectl rollout status deployment/<deployment-name> -n flyfleet

# Restart (roll new ReplicaSet)
kubectl rollout restart deployment/<deployment-name> -n flyfleet

# History
kubectl rollout history deployment/<deployment-name> -n flyfleet

# Rollback
kubectl rollout undo deployment/<deployment-name> -n flyfleet
kubectl rollout undo deployment/<deployment-name> --to-revision=2 -n flyfleet
```

### Services

```bash
# List services
kubectl get svc -n flyfleet
kubectl get svc -n flyfleet -o wide

# Describe service
kubectl describe svc <service-name> -n flyfleet

# Get service endpoints
kubectl get endpoints -n flyfleet

# Test service connectivity
kubectl run -it --rm --namespace flyfleet debug --image=busybox -- \
  wget -qO- http://<service-name>:8084/health
```

### ConfigMaps & Secrets

```bash
# List
kubectl get configmap -n flyfleet
kubectl get secret -n flyfleet

# Create from file
kubectl create configmap app-config --from-file=config.yaml -n flyfleet
kubectl create secret generic db-secret \
  --from-literal=password=secret123 \
  --from-file=ca.crt=./ca.crt \
  -n flyfleet

# View (secrets are base64 encoded)
kubectl get secret <secret-name> -n flyfleet -o yaml
kubectl get secret <secret-name> -n flyfleet -o jsonpath="{.data.password}" | base64 --decode

# Edit (caution!)
kubectl edit configmap <configmap-name> -n flyfleet
```

### Ingress

```bash
# List ingresses
kubectl get ingress -n flyfleet
kubectl get ingress -n flyfleet -o wide

# Describe
kubectl describe ingress <ingress-name> -n flyfleet

# View rules
kubectl get ingress <ingress-name> -n flyfleet -o jsonpath="{.spec.rules}"
```

### Events

```bash
# All events (sorted by time)
kubectl get events --sort-by='.lastTimestamp' -n flyfleet

# Watch events
kubectl get events -n flyfleet --watch

# Pod-specific events
kubectl get events -n flyfleet --field-selector involvedObject.name=<pod-name>

# All namespaces
kubectl get events --all-namespaces --sort-by='.lastTimestamp' | tail -50
```

### Resource Quotas & Limits

```bash
# List quotas
kubectl get resourcequota -n flyfleet
kubectl describe resourcequota <quota-name> -n flyfleet

# List limit ranges
kubectl get limitrange -n flyfleet
```

### Debugging

```bash
# Check pod scheduling issues
kubectl describe pod <pod-name> -n flyfleet | grep -A10 "Events"

# Check node resources
kubectl top nodes
kubectl top pods -n flyfleet

# Verify DNS
kubectl run -it --rm --namespace flyfleet dns-test --image=busybox -- \
  nslookup kubernetes.default

# Check service mesh (if applicable)
kubectl get pods -n istio-system
```

---

## Docker & ECR

### Local Docker

```bash
# Build image
docker build -t fleet-core-service:latest -f infrastructure/docker/GoService.Dockerfile .

# Build with build arg
docker build \
  --build-arg SERVICE_NAME=fleet-core \
  -t fleet-core-service:latest \
  -f infrastructure/docker/GoService.Dockerfile .

# Run container
docker run -p 8084:8084 --rm fleet-core-service:latest

# List images
docker images
docker image ls

# Remove image
docker rmi fleet-core-service:latest

# Prune unused
docker system prune -a
docker image prune -a
docker volume prune
```

### Docker Compose

```bash
# Start all
docker-compose up -d

# Start specific service
docker-compose up -d postgres redis nats

# View logs
docker-compose logs -f postgres
docker-compose logs -f --tail=100 fleet-core-service

# Stop services (keep volumes)
docker-compose stop

# Stop and remove everything (including volumes)
docker-compose down -v

# Exec into container
docker-compose exec postgres psql -U fleet_user -d fleet_db

# Check status
docker-compose ps
docker-compose ps postgres

# Rebuild image
docker-compose build --no-cache fleet-core-service
```

### Amazon ECR

```bash
# Login (using AWS CLI)
aws ecr get-login-password | docker login --username AWS --password-stdin <ACCOUNT>.dkr.ecr.us-east-1.amazonaws.com

# Or use GitHub Action: aws-actions/amazon-ecr-login@v2

# Create repository
aws ecr create-repository --repository-name flyfleet/fleet-core-service

# List repositories
aws ecr describe-repositories | jq '.repositories[]?.repositoryName'

# List images
aws ecr list-images --repository-name flyfleet/fleet-core-service

# Get image details
aws ecr describe-images --repository-name flyfleet/fleet-core-service

# Push tagged image
docker tag fleet-core-service:latest <ACCOUNT>.dkr.ecr.us-east-1.amazonaws.com/flyfleet/fleet-core-service:v1.0.0
docker push <ACCOUNT>.dkr.ecr.us-east-1.amazonaws.com/flyfleet/fleet-core-service:v1.0.0

# Delete image (careful!)
aws ecr batch-delete-image --repository-name flyfleet/fleet-core-service \
  --image-ids imageTag=latest

# Set lifecycle policy (auto-clean old images)
aws ecr put-lifecycle-policy --repository-name flyfleet/fleet-core-service \
  --lifecycle-policy-text file://lifecycle-policy.json
```

**Sample lifecycle policy (keep last 30 images):**
```json
{
  "rules": [
    {
      "rulePriority": 1,
      "description": "Keep last 30 images",
      "selection": {
        "tagStatus": "any",
        "countType": "imageCountMoreThan",
        "countNumber": 30
      },
      "action": {
        "type": "expire"
      }
    }
  ]
}
```

---

## k3s

### Installation

```bash
# Install
curl -sfL https://get.k3s.io | sh -

# Run in background
curl -sfL https://get.k3s.io | sh -s - --disable=traefik --flannel-backend=host-gateway

# Check status
sudo systemctl status k3s
sudo journalctl -u k3s -f

# Get kubeconfig
sudo cat /etc/rancher/k3s/k3s.yaml > ~/.kube/config
export KUBECONFIG=~/.kube/config
chmod 600 ~/.kube/config
```

### Uninstall

```bash
# Stop and remove
sudo systemctl stop k3s
sudo systemctl disable k3s
sudo rm -rf /etc/rancher/k3s
sudo rm -rf /var/lib/rancher/k3s
# Remove kubeconfig
rm ~/.kube/config
```

### Cluster Operations

```bash
# Check nodes
kubectl get nodes -o wide

# Drain node (for maintenance)
kubectl drain <node-name> --ignore-daemonsets --delete-emptydir-data

# Uncordon node
kubectl uncordon <node-name>

# Check k3s version
k3s --version
kubectl version --short

# Get k3s config
sudo cat /etc/rancher/k3s/k3s.yaml

# Disable traefik (if using custom ingress)
# Add to install: --disable=traefik
```

### Add-ons

```bash
# metrics-server (for hpa)
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# Wait for it
kubectl wait --for=condition=ready pod -l k8s-app=metrics-server -n kube-system --timeout=60s

# Verify
kubectl top nodes
kubectl top pods -n flyfleet
```

---

## Makefile

### Development

```bash
# Show all targets
make help
make  # Default target

# Bootstrap environment (start docker-compose)
make bootstrap

# Run tests
make test
make test:unit
make test:integration
make test:coverage
make coverage  # Alias with gate enforcement (70%)

# Lint
make lint
make lint:go
make lint:js

# Build
make build

# Full CI pipeline
make ci

# Clean
make clean
```

### Docker

```bash
make docker-build
make docker-push  # Requires DOCKER_REGISTRY env var
make docker-clean
```

### Governance Checks

```bash
make all-services-foundation-check
make identity-auth-foundation-check
make observability-foundation-check
make mobile-sync-foundation-check
make device-reliability-foundation-check
make workflow-coverage-check
make monolith-governance-check
```

*These are legacy monolith checks, may not apply to new service.*

### Self-Hosted

```bash
# Full setup (registry + build + deploy)
make self-hosted

# Individual steps
make self-hosted-setup      # Deploy registry
make self-hosted-build      # Build and push image
make self-hosted-deploy     # Helm install/upgrade
make self-hosted-registry   # Deploy registry only
make self-hosted-logs       # Tail logs
make self-hosted-portforward  # Expose API on localhost:8084
make self-hosted-clean      # Remove all
```

---

## Self-Hosted

### Quick Commands

```bash
# From flyfleet/ directory

# First time setup
make self-hosted  # Does everything

# After code changes
make self-hosted-build  # Rebuild and push image
kubectl rollout restart deployment/flyfleet-fleet-core-service -n flyfleet

# View logs
make self-hosted-logs

# Access API
make self-hosted-portforward  # Ctrl+C to stop

# Stop everything (keep data)
docker-compose stop postgres redis nats
sudo systemctl stop k3s

# Clean up (delete everything)
make self-hosted-clean
docker-compose down -v
```

### Manual Registry Operations

```bash
# Deploy registry
kubectl apply -f infrastructure/self-hosted/k3s/registry.yaml
kubectl -n registry get pods,svc

# Push image manually
docker tag myimage:tag localhost:30000/flyfleet/myimage:tag
docker push localhost:30000/flyfleet/myimage:tag

# List images in registry
curl http://localhost:30000/v2/_catalog

# Delete image from registry
curl -X DELETE http://localhost:30000/v2/<name>/manifests/<digest>
```

---

## Debugging

### Common Issues & Fixes

| Issue | Command to Diagnose | Fix |
|-------|---------------------|-----|
| Pod pending | `kubectl describe pod <pod> -n flyfleet` | Insufficient resources? Check `kubectl describe node` |
| ImagePullBackOff | `kubectl logs <pod> -n flyfleet` | Image not in registry, wrong registry URL |
| CrashLoopBackOff | `kubectl logs <pod> -p -n flyfleet` | Panic in app, check error logs |
| OOMKilled | `kubectl describe pod <pod> -n flyfleet` | Increase memory limits in values.yaml |
| Service unreachable | `kubectl get endpoints -n flyfleet` | Service has no endpoints, check pod labels |
| Ingress 404 | `kubectl describe ingress <ingress> -n flyfleet` | Host/path mismatch, ingress controller not running |
| DB connection refused | `kubectl exec -it <pod> -- nc -zv fleet-postgres 5432` | Postgres not running, wrong host, network policy |
| Port already in use | `lsof -i :8084` | Kill process: `kill <PID>` or change port-forward port |

### Log Commands

```bash
# All pods in namespace
for pod in $(kubectl get pods -n flyfleet -o jsonpath="{.items[*].metadata.name}"); do
  echo "=== $pod ==="
  kubectl logs $pod -n flyfleet --tail=50
done

# Events (all namespaces)
kubectl get events --all-namespaces --sort-by='.lastTimestamp' | tail -100

# Events for specific pod
kubectl get events -n flyfleet --field-selector involvedObject.name=<pod-name>

# Resource usage
kubectl top pods -n flyfleet
kubectl top nodes
```

### Connectivity Tests

```bash
# From within cluster (run debug pod)
kubectl run -it --rm --namespace flyfleet debug --image=busybox -- /bin/sh

# Inside debug pod:
nslookup fleet-postgres.flyfleet.svc.cluster.local
wget -qO- http://fleet-core-service.flyfleet.svc.cluster.local:8084/health
nc -zv fleet-postgres.flyfleet 5432
ping fleet-nats.flyfleet.svc.cluster.local
```

### Terraform Debugging

```bash
# Check state
terraform state list | grep database

# Validate configuration
terraform validate

# Check what will be changed
terraform plan -var-file=terraform.tfvars

# If apply fails, see what remains
terraform state pull > current-state.json
terraform refresh  # Sync state with real resources

# Unlock stuck lock (if terraform crashed)
terraform force-unlock <LOCK_ID>
```

---

## 📝 Quick Scripts

```bash
#!/bin/bash
# health-check.sh - Check FlyFleet health

API_URL=${1:-http://localhost:8084}

echo "Checking $API_URL/health..."
response=$(curl -s -w "\n%{http_code}" "$API_URL/health" 2>/dev/null)
body=$(echo "$response" | sed '$d')
code=$(echo "$response" | tail -n1)

if [ "$code" = "200" ]; then
  echo "✓ Healthy: $body"
  exit 0
else
  echo "✗ Unhealthy (HTTP $code): $body"
  exit 1
fi
```

```bash
#!/bin/bash
# deploy-self-hosted.sh - Quick self-hosted deploy

set -e

echo "Starting dependencies..."
docker-compose up -d postgres redis nats

echo "Waiting for PostgreSQL..."
until docker exec fleet-postgres pg_isready -U fleet_user -d fleet_db >/dev/null 2>&1; do
  sleep 2
done

echo "Deploying to k3s..."
cd infrastructure/self-hosted && ./scripts/setup.sh

echo "Deployment complete!"
echo "Access API: make self-hosted-portforward"
```

---

## 📚 Cheat Sheets

### kubectl shortcuts (add to ~/.bashrc or ~/.zshrc)

```bash
alias k=kubectl
alias kc='kubectl config use-context'
alias kn='kubectl config set-context --current --namespace'
alias kgp='kubectl get pods'
alias kgs='kubectl get svc'
alias kgd='kubectl get deploy'
alias kgi='kubectl get ingress'
alias kgn='kubectl get ns'
alias kcp='kubectl cp'
alias kl='kubectl logs -f'
alias kd='kubectl describe'
alias ke='kubectl exec -it'
alias kdel='kubectl delete'
alias kapp='kubectl apply -f'
```

Example usage:
```bash
kgp -n flyfleet          # Get pods
kn flyfleet              # Switch namespace to flyfleet
kd pod/myapp -n flyfleet # Describe pod
ke pod/myapp -- /bin/bash  # Exec into pod
```

### Helm shortcuts (zsh/bash)

```bash
alias h='helm'
alias hi='helm install'
alias hu='helm upgrade'
alias hd='helm delete'
alias hl='helm list'
alias hx='helm diff upgrade'
alias ht='helm template'
```

---

## 📞 Emergency Contacts

| Resource | URL / Command |
|----------|---------------|
| Kubernetes Docs | https://kubernetes.io/docs/ |
| Helm Docs | https://helm.sh/docs/ |
| Terraform Docs | https://developer.hashicorp.com/terraform/docs |
| EKS Best Practices | https://aws.github.io/aws-eks-best-practices/ |
| k3s Docs | https://docs.k3s.io/ |

---

**Last Updated:** 2026-03-28
**Keep this file handy!** Print it, bookmark it, or alias it.
