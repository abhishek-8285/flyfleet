#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../../.." && pwd)"

echo "🏠 FlyFleet Self-Hosted Setup"
echo "=============================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
info "Checking prerequisites..."
command -v docker >/dev/null 2>&1 || { error "Docker not found. Install Docker first."; exit 1; }
command -v kubectl >/dev/null 2>&1 || { warn "kubectl not found. Will check for k3s."; }
command -v helm >/dev/null 2>&1 || { warn "Helm not found. Will install."; }
command -v psql >/dev/null 2>&1 || { warn "psql not found. Database setup may need manual steps."; }

# Check k3s
if ! kubectl cluster-info >/dev/null 2>&1; then
    error "k3s cluster not found or not running."
    info "To install k3s, run:"
    info "  curl -sfL https://get.k3s.io | sh -"
    info "Then copy kubeconfig: sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config"
    exit 1
fi

CLUSTER_INFO=$(kubectl cluster-info | head -1)
info "Connected to cluster: ${CLUSTER_INFO}"

# Step 1: Ensure PostgreSQL database exists for Terraform state (if using Terraform)
info "Setting up PostgreSQL state database..."
if docker ps | grep -q fleet-postgres; then
    info "Creating terraform_state database in fleet-postgres..."
    docker exec fleet-postgres psql -U fleet_user -d fleet_db -c "CREATE DATABASE IF NOT EXISTS terraform_state;" 2>/dev/null || true
    info "✓ Database ready"
else
    warn "fleet-postgres container not running. Skipping state database creation."
    warn "Start docker-compose first: cd ${PROJECT_ROOT} && docker-compose up -d postgres redis nats"
fi

# Step 2: Deploy local Docker registry
info "Deploying local Docker registry to k3s..."
kubectl apply -f "${SCRIPT_DIR}/k3s/registry.yaml"

info "Waiting for registry to be ready..."
kubectl -n registry wait --for=condition=ready pod -l app=registry --timeout=60s || true

# Get node IP for registry access
NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}' 2>/dev/null || echo "localhost")
info "Registry will be accessible at: http://${NODE_IP}:30000"

# Configure Docker to trust local registry (Linux)
if [ -f /etc/docker/daemon.json ]; then
    if ! grep -q "localhost:5000" /etc/docker/daemon.json; then
        warn "Adding localhost:5000 to Docker insecure registries..."
        sudo tee /etc/docker/daemon.json > /dev/null <<'EOF'
{
  "insecure-registries": ["localhost:5000", "127.0.0.1:5000"]
}
EOF
        sudo systemctl restart docker
        info "Docker daemon restarted"
    fi
else
    warn "Docker daemon.json not found. You may need to manually add insecure-registries."
fi

# Step 3: Build and push fleet-core-service
read -p "Do you want to build fleet-core-service now? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    info "Building fleet-core-service Docker image..."
    cd "${PROJECT_ROOT}/services/fleet-core"

    # Build
    docker build -t localhost:5000/flyfleet/fleet-core-service:latest -f "${PROJECT_ROOT}/infrastructure/docker/GoService.Dockerfile" --build-arg SERVICE_NAME=fleet-core .

    # Push
    info "Pushing to local registry..."
    docker push localhost:5000/flyfleet/fleet-core-service:latest

    info "✓ Image built and pushed"
fi

# Step 4: Deploy FlyFleet with Helm
info "Deploying FlyFleet with Helm..."

# Ensure namespace exists
kubectl create namespace flyfleet --dry-run=client -o yaml | kubectl apply -f -

# Update values for local deployment
VALUES_FILE="${PROJECT_ROOT}/infrastructure/helm/flyfleet/values.yaml"
info "Using values from: ${VALUES_FILE}"

# Deploy
helm upgrade --install flyfleet "${PROJECT_ROOT}/infrastructure/helm/flyfleet" \
  --namespace flyfleet \
  --set global.imageRegistry="localhost:5000/flyfleet" \
  --set global.imageTag="latest" \
  --set global.postgres.host="fleet-postgres" \
  --set global.redis.host="fleet-redis" \
  --set global.nats.url="nats://fleet-nats.flyfleet.svc.cluster.local:4222" \
  --set global.jwtSecret="dev-secret-change-in-production" \
  --set global.postgres.password="fleet_password" \
  --wait \
  --timeout 5m

info "✓ FlyFleet deployed"

# Step 5: Show status
echo ""
info "Deployment Status:"
kubectl get pods -n flyfleet
kubectl get svc -n flyfleet

echo ""
info "🎉 Self-hosted FlyFleet is running!"
echo ""
echo "Next steps:"
echo "  1. Access services via port-forward:"
echo "     kubectl port-forward -n flyfleet svc/fleet-core-service 8084:8084"
echo "  2. Check health: http://localhost:8084/health"
echo "  3. View logs: kubectl logs -n flyfleet deployment/flyfleet-fleet-core-service -f"
echo ""
echo "To stop everything:"
echo "  helm uninstall flyfleet -n flyfleet"
echo "  kubectl delete namespace registry"
echo ""
