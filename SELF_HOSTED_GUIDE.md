# FlyFleet Self-Hosted Deployment Guide

**100% Open Source, No Cloud Dependencies**

This guide walks you through deploying FlyFleet entirely on your own infrastructure using k3s (lightweight Kubernetes) and locally-run services.

---

## 🎯 What You'll Deploy

```
┌─────────────────────────────────────────────────────────────┐
│                    Self-Hosted Stack                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              k3s (Local Kubernetes)                  │  │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐    │  │
│  │  │fleet-core  │  │   fleet-   │  │   fleet-   │    │  │
│  │  │  service   │  │   ml-      │  │   other    │    │  │
│  │  └────────────┘  │  service   │  │  services  │    │  │
│  │                   └────────────┘  └────────────┘    │  │
│  └──────────────────────────────────────────────────────┘  │
│                            ▲                                │
│                            │                                │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Docker Compose (Databases)              │  │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐    │  │
│  │  │ PostgreSQL │  │    NATS    │  │    Redis   │    │  │
│  │  │  (Times-   │  │  JetStream │  │            │    │  │
│  │  │   caleDB)  │  │            │  │            │    │  │
│  │  └────────────┘  └────────────┘  └────────────┘    │  │
│  └──────────────────────────────────────────────────────┘  │
│                            ▲                                │
│  ┌──────────────────────────────────────────────────────┐  │
│  │        Local Docker Registry (localhost:5000)        │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 📋 Prerequisites

- **Linux/macOS** (WSL2 on Windows also works)
- **Docker** and **docker-compose** installed
- **k3s** (lightweight Kubernetes) installed
- **kubectl** configured to access k3s
- **Helm** 3.x installed
- **Go** 1.26+ (to build the service)
- **Git**

### Install Quick Links

```bash
# k3s
curl -sfL https://get.k3s.io | sh -
sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
export KUBECONFIG=~/.kube/config

# Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Docker: https://docs.docker.com/engine/install/
```

---

## 🚀 One-Command Deployment

We've created a single script that does everything:

```bash
make self-hosted
# Or directly:
cd infrastructure/self-hosted && ./scripts/setup.sh
```

This script will:
1. ✅ Check all prerequisites
2. ✅ Verify docker-compose Postgres/Redis/NATS are running
3. ✅ Create Terraform state database (if using Terraform)
4. ✅ Deploy local Docker registry to k3s
5. ✅ Configure Docker to trust localhost:5000
6. ✅ Build fleet-core-service Docker image
7. ✅ Push image to local registry
8. ✅ Deploy FlyFleet via Helm to k3s
9. ✅ Show deployment status

---

## 🛠️ Manual Step-by-Step

If you prefer manual control, follow these steps:

### Step 1: Start Docker Compose Services

```bash
cd /data/data/com.termux/files/home/flyfleet
docker-compose up -d postgres redis nats

# Verify they're running
docker ps | grep -E "postgres|redis|nats"
```

### Step 2: Install/Verify k3s

```bash
# Check if k3s is running
kubectl cluster-info

# If not installed:
curl -sfL https://get.k3s.io | sh -
sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
export KUBECONFIG=~/.kube/config
```

### Step 3: Deploy Local Docker Registry

```bash
kubectl apply -f infrastructure/self-hosted/k3s/registry.yaml

# Wait for it
kubectl -n registry wait --for=condition=ready pod -l app=registry

# Get access URL
NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
echo "Registry: http://${NODE_IP}:30000"
```

### Step 4: Build and Push Image

```bash
# Build fleet-core-service
docker build -t localhost:5000/flyfleet/fleet-core-service:latest \
  -f infrastructure/docker/GoService.Dockerfile \
  --build-arg SERVICE_NAME=fleet-core \
  services/fleet-core

# Push to local registry
docker push localhost:5000/flyfleet/fleet-core-service:latest
```

### Step 5: Deploy with Helm

```bash
# Create namespace
kubectl create namespace flyfleet

# Deploy
helm upgrade --install flyfleet infrastructure/helm/flyfleet \
  --namespace flyfleet \
  --set global.imageRegistry="localhost:5000/flyfleet" \
  --set global.imageTag="latest" \
  --set global.postgres.host="fleet-postgres" \
  --set global.redis.host="fleet-redis" \
  --set global.nats.url="nats://fleet-nats.flyfleet.svc.cluster.local:4222" \
  --set global.jwtSecret="dev-secret" \
  --set global.postgres.password="fleet_password" \
  --wait

# Check status
kubectl get pods -n flyfleet
```

### Step 6: Access the Application

```bash
# Port-forward to access API
kubectl port-forward -n flyfleet svc/fleet-core-service 8084:8084

# In another terminal, test health
curl http://localhost:8084/health
```

---

## 🎛️ Makefile Shortcuts

After setup, use these convenient commands:

```bash
make self-hosted-logs           # Tail logs from fleet-core-service
make self-hosted-portforward    # Expose service on localhost:8084
make self-hosted-clean          # Remove all deployed resources
make self-hosted-deploy         # Re-deploy after image rebuild
```

---

## 🔧 Configuration

### Environment Variables

The Helm chart uses these values (override via `--set`):

| Parameter | Default | Description |
|-----------|---------|-------------|
| `global.imageRegistry` | (none) | Docker registry host:port |
| `global.imageTag` | latest | Image tag to deploy |
| `global.postgres.host` | fleet-postgres | PostgreSQL service hostname |
| `global.redis.host` | fleet-redis | Redis service hostname |
| `global.nats.url` | nats://fleet-nats:4222 | NATS URL |
| `global.jwtSecret` | (required) | JWT signing secret |
| `global.postgres.password` | fleet_password | DB password |

### Custom values.yaml

You can create `infrastructure/helm/flyfleet/values-selfhosted.yaml`:

```yaml
global:
  imageRegistry: "localhost:5000/flyfleet"
  imageTag: "latest"
  postgres:
    host: "fleet-postgres"
    password: "your-password"
  jwtSecret: "your-jwt-secret"

# Enable additional services
driver-service:
  enabled: true
vehicle-service:
  enabled: true
# ... etc
```

Then deploy with:
```bash
helm upgrade --install flyfleet infrastructure/helm/flyfleet \
  -f values-selfhosted.yaml \
  --namespace flyfleet
```

---

## 🧪 Testing

### Health Check

```bash
# After port-forward
curl http://localhost:8084/health
# Expected: {"status":"ok"}
```

### View Logs

```bash
make self-hosted-logs
# Or:
kubectl logs -n flyfleet deployment/flyfleet-fleet-core-service -f
```

### Access Database

```bash
# Connect to Postgres
docker exec fleet-postgres psql -U fleet_user -d fleet_db

# Query data
SELECT * FROM information_schema.tables LIMIT 5;
```

---

## 🧹 Cleanup

To remove everything:

```bash
# Delete Helm release
helm uninstall flyfleet -n flyfleet
kubectl delete namespace flyfleet

# Delete local registry
kubectl delete -f infrastructure/self-hosted/k3s/registry.yaml
kubectl delete namespace registry

# Stop docker-compose services
docker-compose down

# Optional: Remove Docker images
docker rmi localhost:5000/flyfleet/fleet-core-service:latest
```

### Full cleanup (make):
```bash
make self-hosted-clean
```

---

## 🐛 Troubleshooting

### Registry connection refused
```bash
# Check Docker daemon config
cat /etc/docker/daemon.json
# Must include: "insecure-registries": ["localhost:5000"]

# Restart Docker
sudo systemctl restart docker
```

### Pods stuck in Pending
```bash
# Check if storage class exists
kubectl get storageclass

# If missing:
kubectl apply -f infrastructure/self-hosted/k3s/storageclass.yaml
```

### ImagePullBackOff
```bash
# Verify image exists in registry
curl http://localhost:30000/v2/_catalog

# Re-tag and push
docker tag fleet-core-service localhost:5000/flyfleet/fleet-core-service:latest
docker push localhost:5000/flyfleet/fleet-core-service:latest
```

### Can't connect to Postgres from pod
```bash
# Check service exists
kubectl get svc -n flyfleet fleet-postgres

# Test connectivity from a pod
kubectl run -it --rm debug --image=alpine/curl --restart=Never -- sh
# Inside pod:
apk add postgresql-client
pg_isready -h fleet-postgres -p 5432
```

---

## 📊 What's Different from Cloud

| Feature | Cloud (AWS) | Self-Hosted (k3s) |
|---------|-------------|-------------------|
| **Kubernetes** | EKS | k3s (local Docker) |
| **Database** | RDS PostgreSQL | Docker container |
| **Cache** | ElastiCache Redis | Docker container |
| **Messaging** | NATS on EC2 | Docker container |
| **Registry** | ECR | Local Docker registry |
| **Storage** | EBS | hostPath (local) |
| **Load Balancer** | ELB | NodePort/port-forward |
| **Secrets** | Secrets Manager | K8s Secrets (base64) |
| **State Locking** | DynamoDB | PostgreSQL (local) |

---

## ⚠️ Production Considerations

For production self-hosting on bare metal/VMs:

1. **Use proper storage**: Replace hostPath with Longhorn or Ceph RBD
2. **High availability**: Run 3+ k3s server nodes with embedded etcd
3. **Ingress**: Deploy NGINX Ingress or Traefik with real DNS
4. **TLS**: Install cert-manager for Let's Encrypt certificates
5. **Backups**: Regular PostgreSQL dumps and etcd backups
6. **Monitoring**: Deploy Prometheus + Grafana stack
7. **Registry**: Use a proper registry with S3 backend and TLS
8. **Disaster recovery**: Off-site backups, documented recovery procedures

---

## 🔄 Migrating Between Cloud and Self-Hosted

### From Cloud → Self-Hosted

1. Export data from RDS: `pg_dump -h <rds-endpoint> -U fleet_user fleet_db > backup.sql`
2. Import to local: `docker exec -i fleet-postgres psql -U fleet_user -d fleet_db < backup.sql`
3. Update Helm values to use local endpoints
4. Re-deploy

### From Self-Hosted → Cloud

1. Export local DB: `docker exec fleet-postgres pg_dump -U fleet_user fleet_db > backup.sql`
2. Import to RDS: `psql -h <rds-endpoint> -U fleet_user fleet_db < backup.sql`
3. Update Helm values to use RDS endpoint
4. Re-deploy with cloud Terraform

---

## 📚 Resources

- **k3s Docs**: https://docs.k3s.io/
- **Helm Docs**: https://helm.sh/docs/
- **FlyFleet Architecture**: `ARCHITECTURE_ASSESSMENT.md`
- **Development Guide**: `CLAUDE.md`

---

**Ready to deploy?** Run `make self-hosted` or `./infrastructure/self-hosted/scripts/setup.sh`

---

**Last Updated:** 2026-03-27
**Maintainer:** FlyFleet Platform Team
