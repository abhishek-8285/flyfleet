# Self-Hosted FlyFleet Infrastructure

This directory contains configurations for running FlyFleet entirely self-hosted without cloud providers.

## Components

- **k3s**: Lightweight Kubernetes distribution (runs in Docker)
- **Local Docker Registry**: For storing container images
- **PostgreSQL**: Already in docker-compose
- **NATS**: Already in docker-compose
- **Redis**: Already in docker-compose

## Quick Start

### 1. Start k3s Cluster

```bash
# Install k3s (if not already)
curl -sfL https://get.k3s.io | sh -

# Wait for node to be Ready
k3s kubectl get nodes

# Export kubeconfig for regular kubectl
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
# Or copy to ~/.kube/config
sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
```

### 2. Set Up Local Docker Registry

```bash
# Create registry namespace and deploy
kubectl create namespace registry
kubectl apply -f registry/registry.yaml

# Port-forward for local access
kubectl port-forward -n registry svc/registry 5000:5000 &
```

### 3. Configure Docker to Trust Local Registry (insecure)

For development only:

```bash
# Create /etc/docker/daemon.json if not exists
sudo tee /etc/docker/daemon.json <<EOF
{
  "insecure-registries": ["localhost:5000"]
}
EOF

sudo systemctl restart docker
```

### 4. Build and Push Images Locally

```bash
# Build fleet-core-service
cd services/fleet-core
docker build -t localhost:5000/flyfleet/fleet-core-service:latest -f ../../infrastructure/docker/GoService.Dockerfile --build-arg SERVICE_NAME=fleet-core .

# Push to local registry
docker push localhost:5000/flyfleet/fleet-core-service:latest
```

### 5. Deploy with Helm

```bash
# Add local registry as Helm repo (optional)
# Or just use the images directly

# Update values.yaml to use localhost:5000 images
# Then deploy:
helm upgrade --install flyfleet ../../helm/flyfleet \
  --namespace flyfleet \
  --create-namespace \
  --set global.imageRegistry="localhost:5000/flyfleet" \
  --set global.imageTag="latest" \
  --set global.postgres.host="fleet-postgres" \
  --set global.redis.host="fleet-redis" \
  --set global.nats.url="nats://fleet-nats:4222" \
  --wait
```

### 6. Access Services

```bash
# Port-forward to access services locally
kubectl port-forward -n flyfleet svc/fleet-core-service 8084:8084
# Now http://localhost:8084/health should work
```

## Directory Structure

```
self-hosted/
├── k3s/               # k3s-specific configs
│   ├── registry.yaml  # Local Docker registry deployment
│   └── storageclass.yaml # Default storage class for k3s
├── registry/          # Helm chart for local registry (optional)
└── scripts/           # Helper scripts
    ├── setup.sh       # Full setup script
    ├── deploy.sh      # Deploy all services
    └── backup.sh      # Backup PostgreSQL database
```

## Differences from Cloud Deployment

- **No VPC/Networking**: k3s runs on local Docker network
- **No EBS/Storage**: Uses local path storage or hostPath
- **No IAM/IRSA**: ServiceAccount without IAM integration
- **No LoadBalancer**: Use NodePort or port-forward for access
- **No ECR**: Local Docker registry
- **No Terraform** (optional): Can use kubectl/helm directly

## Production Self-Hosted Considerations

For a real self-hosted production cluster (bare metal or VMs):

1. **Use proper storage**: Longhorn, Rook Ceph, or NFS
2. **Use Ingress Controller**: Traefik or NGINX Ingress (k3s includes Traefik by default)
3. **Use cert-manager**: For TLS certificates (Let's Encrypt)
4. **Use external PostgreSQL**: Consider RDS alternative like CrunchyDB on separate nodes
5. **High availability**: Multiple k3s server nodes with embedded etcd
6. **Backups**: Regular etcd backups, PostgreSQL dumps

## Troubleshooting

### k3s not starting
```bash
# Check logs
sudo journalctl -u k3s -f

# Reset if needed
sudo systemctl restart k3s
```

### Registry connection refused
```bash
# Check registry pod
kubectl get pods -n registry
kubectl logs -n registry -l app=registry

# Ensure Docker daemon trusts localhost:5000
```

### Images not found
```bash
# List images in local registry
curl http://localhost:5000/v2/_catalog

# Re-tag and push if needed
docker tag fleet-core-service localhost:5000/flyfleet/fleet-core-service:latest
docker push localhost:5000/flyfleet/fleet-core-service:latest
```

---

**Note**: This setup is for development/staging. For production self-hosting, use a multi-node k3s cluster with proper storage, networking, and monitoring.
