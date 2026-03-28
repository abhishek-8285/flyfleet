# FlyFleet Self-Hosted Deployment Guide

**Target audience:** Developers, DevOps engineers, QA teams
**Difficulty:** Easy
**Time estimate:** 20-30 minutes
**Last updated:** 2026-03-28

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Architecture](#architecture)
4. [Installation Steps](#installation-steps)
5. [Configuration](#configuration)
6. [Operations](#operations)
7. [Networking](#networking)
8. [Storage](#storage)
9. [Troubleshooting](#troubleshooting)
10. [Cleanup](#cleanup)
11. [Comparison: Self-Hosted vs AWS](#comparison-self-hosted-vs-aws)

---

## Overview

This guide walks you through deploying FlyFleet entirely on your local machine using **k3s** (lightweight Kubernetes) and **Docker**. No AWS account required, zero cost, 100% open source.

### What You'll Get

- **Kubernetes cluster** running in Docker via k3s
- **Local Docker registry** on port 30000
- **PostgreSQL + Redis + NATS** via docker-compose (reuse existing)
- **FlyFleet API** accessible at `http://localhost:8084`
- **Helm-managed** deployments (same charts as production)

### Key Benefits

✅ Zero infrastructure cost
✅ Fast iteration (local dev loop)
✅ Identical to production (same Helm charts)
✅ Easy to tear down and rebuild
✅ No cloud vendor lock-in

### Limitations

⚠️ Single-node (no HA)
⚠️ Not suitable for production traffic
⚠️ Limited to your laptop/desktop resources
⚠️ Local Docker registry has no auth (dev only)

---

## Prerequisites

### Required Software

| Tool | Version | Install Command |
|------|---------|-----------------|
| Docker | 20.10+ | https://docs.docker.com/engine/install/ |
| docker-compose | 2.0+ | Included with Docker Desktop |
| k3s | Latest | `curl -sfL https://get.k3s.io \| sh -` |
| Helm | 3.12+ | https://helm.sh/docs/intro/install/ |
| kubectl | 1.28+ | https://kubernetes.io/docs/tasks/tools/ |
| Git | 2.0+ | Your package manager |
| Go | 1.23+ | https://go.dev/dl/ |
| Node.js | 20+ | https://nodejs.org/ |

**Verification:**
```bash
docker --version
docker-compose --version
k3s --version
helm version
kubectl version --client
go version
node --version
```

### System Resources

**Minimum:**
- CPU: 4 cores
- RAM: 8 GB (16 GB recommended)
- Disk: 20 GB free

**k3s resource usage:**
- ~2 GB RAM
- ~1 CPU core

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Your Local Machine                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                k3s (Kubernetes in Docker)          │   │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐   │   │
│  │  │ fleet-core │  │   postgres │  │    redis   │   │   │
│  │  │  service   │  │  (compose) │  │  (compose) │   │   │
│  │  └────────────┘  └────────────┘  └────────────┘   │   │
│  │         │               │               │           │   │
│  │         └───────────────┼───────────────┘           │   │
│  │                       │                             │   │
│  │            ┌──────────▼──────────┐                 │   │
│  │            │   Local Docker Reg   │  (port 30000)  │   │
│  │            │   (NodePort 30000)  │                 │   │
│  │            └─────────────────────┘                 │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  External Access: localhost:8084 (port-forward)           │
│                                                             │
└─────────────────────────────────────────────────────────────┘

docker-compose services (separate Docker network):
- fleet-postgres (port 5432)
- fleet-redis (port 6379)
- fleet-nats (port 4222)
```

**Data flow:**
1. Developer builds `fleet-core-service` Docker image
2. Image pushed to local Docker registry (`localhost:30000`)
3. Helm chart pulls image from registry
4. Pods connect to PostgreSQL/Redis/NATS via Kubernetes services
5. Port-forward exposes API to localhost

---

## Installation Steps

### Step 1: Start Dependencies (Postgres, Redis, NATS)

FlyFleet uses `docker-compose.yml` for stateful services. If you haven't copied it from the repo:

```bash
cd /data/data/com.termux/files/home/flyfleet

# Ensure docker-compose.yml exists
ls docker-compose.yml
# Should show: docker-compose.yml

# Start dependencies
docker-compose up -d postgres redis nats

# Verify
docker-compose ps
```

**Expected output:**
```
 NAME                IMAGE               STATUS          PORTS
 fleet-postgres      postgres:15         Up (healthy)    0.0.0.0:5432->5432/tcp
 fleet-redis         redis:7-alpine      Up (healthy)    0.0.0.0:6379->6379/tcp
 fleet-nats          nats:2.10           Up (healthy)    0.0.0.0:4222->4222/tcp
```

**Wait for PostgreSQL:**
```bash
# Wait until DB is ready to accept connections
until docker exec fleet-postgres pg_isready -U fleet_user -d fleet_db >/dev/null 2>&1; do
  echo "Waiting for PostgreSQL..."
  sleep 2
done
echo "PostgreSQL ready!"
```

---

### Step 2: Install k3s

k3s is a lightweight Kubernetes distribution that runs in a single Docker container.

```bash
# Install k3s (runs as root)
curl -sfL https://get.k3s.io | sh -

# Wait for k3s to start
sleep 10

# Copy kubeconfig to user directory (optional but convenient)
sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
chmod 600 ~/.kube/config
```

**Verify installation:**
```bash
k3s --version
# Output: k3s version v1.xx.x+x...

kubectl cluster-info
# Output: Kubernetes control plane is running at https://localhost:6443

kubectl get nodes
# Output:
# NAME    STATUS   ROLES           AGE   VERSION
# local   Ready    control-plane   30s   v1.28.5+k3s1
```

---

### Step 3: Deploy Everything (Automated)

The FlyFleet repo includes a Makefile target that automates the entire self-hosted setup:

```bash
cd /data/data/com.termux/files/home/flyfleet

# This will:
# 1. Deploy local Docker registry (infrastructure/self-hosted/k3s/registry.yaml)
# 2. Build fleet-core-service Docker image
# 3. Push image to local registry (localhost:30000/flyfleet/fleet-core-service:latest)
# 4. Deploy Helm chart with registry config
# 5. Wait for pods to be ready
make self-hosted
```

**Alternative manual steps** (if you want to understand what's happening):

```bash
# 1. Deploy local Docker registry to k3s
kubectl apply -f infrastructure/self-hosted/k3s/registry.yaml

# Wait for registry pod
kubectl -n registry wait --for=condition=ready pod -l app=registry --timeout=60s

# 2. Build Docker image for local registry
docker build \
  -t localhost:30000/flyfleet/fleet-core-service:latest \
  -f infrastructure/docker/GoService.Dockerfile \
  --build-arg SERVICE_NAME=fleet-core .

# 3. Push to local registry
docker push localhost:30000/flyfleet/fleet-core-service:latest

# 4. Deploy Helm chart with local registry config
helm upgrade --install flyfleet infrastructure/helm/flyfleet \
  --namespace flyfleet \
  --create-namespace \
  --set global.imageRegistry="localhost:30000/flyfleet" \
  --set global.imageTag="latest" \
  --set global.postgres.host="fleet-postgres" \
  --set global.redis.host="fleet-redis" \
  --set global.nats.url="nats://fleet-nats.flyfleet.svc.cluster.local:4222" \
  --set global.jwtSecret="dev-secret-change-in-production" \
  --set global.postgres.password="fleet_password" \
  --wait

# 5. Verify
kubectl get pods -n flyfleet
kubectl get svc -n flyfleet
kubectl get ingress -n flyfleet
```

---

### Step 4: Access the API

FlyFleet API runs in the k3s cluster. Expose it to your local machine:

```bash
# Port-forward the fleet-core-service
make self-hosted-portforward
# Or manually:
# kubectl port-forward -n flyfleet svc/flyfleet-fleet-core-service 8084:8084
```

This forwards cluster port 8084 to localhost:8084. Keep this terminal open.

**Test the API:**

```bash
# Health check
curl http://localhost:8084/health

# Expected output:
# {"status":"ok","timestamp":"2026-03-28T..."}

# Check API is serving
curl http://localhost:8084/api/v1/health

# View Swagger UI (if enabled)
# Open browser: http://localhost:8084/swagger
```

**Alternative: Use port-forward in background**
```bash
kubectl port-forward -n flyfleet svc/flyfleet-fleet-core-service 8084:8084 &
PF_PID=$!
sleep 2
curl http://localhost:8084/health
# When done: kill $PF_PID
```

---

## Configuration

### Helm Values Override

The default `infrastructure/helm/flyfleet/values.yaml` is configured for self-hosted deployment via environment variables:

```yaml
global:
  imageRegistry: "localhost:30000/flyfleet"  # Local Docker registry
  imageTag: "latest"
  environment: "development"

  postgres:
    host: "fleet-postgres"  # docker-compose service name
    port: 5432
    database: "fleet_db"
    username: "fleet_user"
    password: "fleet_password"  # Dev only - use secret in prod

  redis:
    host: "fleet-redis"
    port: 6379

  nats:
    url: "nats://fleet-nats.flyfleet.svc.cluster.local:4222"

  jwtSecret: "dev-secret-change-in-production"  # Change this!
```

**Override via command line:**
```bash
helm upgrade --install flyfleet infrastructure/helm/flyfleet \
  --namespace flyfleet \
  --set global.jwtSecret="$(openssl rand -base64 32)"
```

**Use custom values file:**
```bash
helm upgrade --install flyfleet infrastructure/helm/flyfleet \
  -f values-selfhosted.yaml \
  --namespace flyfleet
```

### Environment Variables

The following env vars can be set in the Helm chart under `fleet-core-service.env`:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | `postgres://...` | PostgreSQL connection |
| `REDIS_URL` | No | `redis://...` | Redis connection |
| `NATS_URL` | Yes | `nats://...` | NATS JetStream URL |
| `JWT_SECRET` | Yes | random | JWT signing key |
| `ENVIRONMENT` | No | `development` | App environment |

---

## Operations

### View Logs

```bash
# Pod logs (auto-follow)
make self-hosted-logs
# Or:
kubectl logs -n flyfleet deployment/flyfleet-fleet-core-service -f

# Specific container
kubectl logs -n flyfleet pod/flyfleet-fleet-core-service-xxxxx -c fleet-core-service -f

# Previous logs (if pod crashed)
kubectl logs -n flyfleet deployment/flyfleet-fleet-core-service -p
```

### Restart Service

```bash
# Rolling restart
kubectl rollout restart deployment/flyfleet-fleet-core-service -n flyfleet

# Watch rollout status
kubectl rollout status deployment/flyfleet-fleet-core-service -n flyfleet --watch

# Rollback if needed
kubectl rollout undo deployment/flyfleet-fleet-core-service -n flyfleet
```

### Scale Replicas

```bash
# Scale to 3 replicas
kubectl scale deployment flyfleet-fleet-core-service --replicas=3 -n flyfleet

# Autoscaling (if enabled in values.yaml)
kubectl get hpa -n flyfleet
```

### Port-Forward Multiple Services

```bash
# API
kubectl port-forward -n flyfleet svc/flyfleet-fleet-core-service 8084:8084 &

# PostgreSQL (if needed for debugging)
kubectl port-forward -n flyfleet svc/fleet-postgres 5432:5432 &

# Redis
kubectl port-forward -n flyfleet svc/fleet-redis 6379:6379 &
```

### Access PostgreSQL from Host

```bash
# The postgres service is from docker-compose, not k8s
# Connect directly to docker-compose container:
docker exec -it fleet-postgres psql -U fleet_user -d fleet_db

# Or from host (since port 5432 is published):
psql -h localhost -p 5432 -U fleet_user -d fleet_db
```

### Access NATS

```bash
# Using nats CLI (install: go install github.com/nats-io/natscli/nats@latest)
nats --server nats://localhost:4222 stream ls

# Using telnet
telnet localhost 4222
```

---

## Networking

### Service Discovery

**Kubernetes DNS:**
- Services reachable at `servicename.namespace.svc.cluster.local`
- Example: `fleet-postgres.flyfleet.svc.cluster.local` → cluster IP of postgres

**Port mapping:**
| Service | Cluster Port | Host Port | Access Method |
|---------|--------------|-----------|---------------|
| fleet-core-service | 8084 | 8084 (via port-forward) | `localhost:8084` |
| postgres | 5432 | 5432 (docker-compose) | `localhost:5432` |
| redis | 6379 | 6379 (docker-compose) | `localhost:6379` |
| nats | 4222 | 4222 (docker-compose) | `localhost:4222` |
| Docker registry | 30000 | 30000 (NodePort) | `localhost:30000` |

### Ingress

In self-hosted mode, ingress is configured but may not have a working load balancer. k3s uses **Traefik** as the default ingress controller.

**Check ingress:**
```bash
kubectl get ingress -n flyfleet
kubectl describe ingress flyfleet -n flyfleet
```

**Access via ingress host:**
The ingress is configured for `api.flyfleet.com`. To test locally, edit `/etc/hosts`:

```
127.0.0.1 api.flyfleet.com
```

Then:
```bash
curl http://api.flyfleet.com/health
```

---

## Storage

### Persistent Volume Claims

The self-hosted setup uses the default k3s storage class (`local-path`). PVCs are created by:

- PostgreSQL statefulset (if using the Helm chart version)
- But currently using docker-compose for postgres/redis/nats

**If deploying PostgreSQL via Helm** (future option):
```bash
kubectl get pvc -n flyfleet
kubectl get pv
```

### Docker Registry Storage

The local Docker registry uses `emptyDir` volume (ephemeral). Images are lost when pod restarts.

For persistent registry storage, modify `infrastructure/self-hosted/k3s/registry.yaml`:

```yaml
spec:
  volumes:
    - name: registry-storage
      persistentVolumeClaim:
        claimName: registry-pvc
```

Then create a PVC with appropriate storage class.

---

## Troubleshooting

### k3s fails to start

**Check Docker container:**
```bash
docker ps | grep k3s
docker logs k3s
```

**Common issues:**
- Port 6443 already in use → another Kubernetes cluster running
- Insufficient memory → increase swap or close apps
- Docker not running → `systemctl start docker`

**Fix:** Stop other clusters:
```bash
# If minikube or docker-desktop is running
minikube stop
# or
docker context use default  # Switch from docker-desktop
```

---

### Helm deployment fails

**Check pod status:**
```bash
kubectl get pods -n flyfleet -o wide
kubectl describe pod <pod-name> -n flyfleet
kubectl logs <pod-name> -n flyfleet
```

**Common failures:**

1. **ImagePullBackOff**
   - Registry not running → `kubectl get pods -n registry`
   - Image not pushed → Re-run `make self-hosted-build`
   - Wrong registry URL → Check `global.imageRegistry` value

2. **CrashLoopBackOff**
   - Check logs for panic: `kubectl logs <pod> -n flyfleet --previous`
   - Missing env vars → `kubectl describe pod <pod> -n flyfleet | grep -A5 "Environment"`
   - Database connection failure → verify docker-compose services are up

3. **Pending (unschedulable)**
   - Insufficient resources → `kubectl describe pod <pod>`
   - Node not ready → `kubectl get nodes`

---

### Docker registry not accessible

```bash
# Check registry pod
kubectl get pods -n registry

# Check service
kubectl get svc -n registry

# Test registry API from within cluster
kubectl run -it --rm --namespace flyfleet registry-test --image=curlimages/curl -- \
  curl http://registry:5000/v2/_catalog

# Test from host
curl http://localhost:30000/v2/_catalog
# Should return {"repositories":[]} or list of repos
```

**If registry service is ClusterIP (not NodePort):**
```bash
kubectl port-forward -n registry svc/registry 30000:5000
# Then test on localhost:30000
```

---

### Cannot connect to PostgreSQL from pod

```bash
# Check if postgres service is reachable from within cluster
kubectl run -it --rm --namespace flyfleet postgres-test --image=postgres:15 -- \
  bash -c "nc -zv fleet-postgres 5432 && echo OK"

# Check DNS resolution
kubectl run -it --rm --namespace flyfleet dns-test --image=busybox -- \
  nslookup fleet-postgres

# Check service exists
kubectl get svc -n flyfleet | grep postgres

# Verify docker-compose postgres is healthy
docker-compose ps postgres
docker-compose logs postgres
```

---

### Port-forward not working

```bash
# Check service has correct port
kubectl get svc -n flyfleet fleet-core-service -o wide

# Check pod is running
kubectl get pods -n flyfleet

# Port-forward with explicit pod if service broken
kubectl port-forward -n flyfleet pod/flyfleet-fleet-core-service-xxxx 8084:8084

# Check if port 8084 is already in use
lsof -i :8084
# Kill process using port: kill <PID>
```

---

### Out of disk space

k3s stores container images and logs in `/var/lib/rancher/k3s/`. Clean up:

```bash
# Remove unused images
docker image prune -a

# Clean k3s agent container logs
sudo journalctl -u k3s --vacuum-time=1h

# Remove unused PVCs (if any)
kubectl delete pvc --field-selector=status.phase=Failed
```

---

### PVCs stuck in `Pending`

k3s default storage class (`local-path`) may not be available. Check:

```bash
kubectl get storageclass
kubectl get pvc -n flyfleet

# If no default storage class, edit chart to disable PVCs or create storage class
# Edit release: helm edit flyfleet -n flyfleet
# Disable persistence: persistence.enabled=false
```

---

## Cleanup

### Stop everything (keep data)

```bash
# Stop port-forward (Ctrl+C in that terminal)

# Stop docker-compose services (but keep volumes)
docker-compose stop postgres redis nats

# Stop k3s
sudo systemctl stop k3s

# Keep kubectl context, data preserved on restart
```

### Remove everything (nuclear)

```bash
make self-hosted-clean
# Equivalent to:
helm uninstall flyfleet -n flyfleet 2>/dev/null || true
kubectl delete namespace flyfleet 2>/dev/null || true
kubectl delete -f infrastructure/self-hosted/k3s/registry.yaml 2>/dev/null || true
kubectl delete namespace registry 2>/dev/null || true
```

**Delete data:**
```bash
# WARNING: This deletes all data (Postgres DB, Redis cache, NATS streams)
docker-compose down -v  # Deletes volumes
sudo rm -rf /var/lib/rancher/k3s/agent/containerd  # Deletes all k8s images
```

### Restart from clean state

```bash
# Remove everything including volumes
docker-compose down -v
sudo systemctl stop k3s
sudo rm -rf /var/lib/rancher/k3s

# Restart fresh
docker-compose up -d postgres redis nats
curl -sfL https://get.k3s.io | sh -
sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
make self-hosted
```

---

## Comparison: Self-Hosted vs AWS

| Feature | Self-Hosted (k3s) | AWS (EKS) |
|---------|-------------------|-----------|
| **Cost** | $0 | $150-300+/month |
| **High Availability** | ❌ Single node | ✅ Multi-AZ |
| **Storage** | Local `emptyDir` / local-path | EBS / EFS |
| **Load Balancing** | NodePort / port-forward | AWS ALB / NLB |
| **Observability** | Manual setup | CloudWatch + Managed Prometheus |
| **Secrets Management** | Helm values (insecure) | AWS Secrets Manager |
| **Backup/Restore** | Manual `pg_dump` | RDS automated backups |
| **Auto-scaling** | Manual (edit deployment) | Cluster Autoscaler + HPA |
| **Updates** | Manual `k3s upgrade` | Managed EKS control plane |
| **Security** | Basic (localhost only) | VPC, KMS, IAM, SG |
| **Use Case** | Dev, Testing, Demos | Production |
| **Setup Time** | 20 minutes | 2-4 hours (Terraform) |

### When to use self-hosted:
- Local development
- CI/CD runners
- Proof of concepts
- Learning Kubernetes

### When to use AWS:
- Production traffic
- Need HA and SLA
- Customer-facing APIs
- Compliance requirements
- Large scale (10k+ requests/sec)

---

## Advanced Topics

### Custom Storage Class

If you need persistent volumes that survive pod restarts:

Create `infrastructure/self-hosted/k3s/storageclass.yaml`:
```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: local-path
provisioner: rancher.io/local-path
reclaimPolicy: Delete
volumeBindingMode: Immediate
allowVolumeExpansion: true
```

Apply:
```bash
kubectl apply -f infrastructure/self-hosted/k3s/storageclass.yaml
```

### Ingress with Traefik (k3s default)

k3s includes Traefik ingress controller by default.

**View ingress controller:**
```bash
kubectl get pods -n kube-system | grep traefik
```

**Test with custom host:**
Add to `/etc/hosts`:
```
127.0.0.1 api.flyfleet.local
```

Then port-forward Traefik service:
```bash
kubectl port-forward -n kube-system svc/traefik 8080:80
curl http://api.flyfleet.local/health --header "Host: api.flyfleet.local"
```

### Enable Metrics Server (for HPA)

```bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# Wait for metrics-server
kubectl wait --for=condition=ready pod -l k8s-app=metrics-server -n kube-system --timeout=60s

# Test
kubectl top pods -n flyfleet
```

---

## FAQ

**Q: Can I use self-hosted for production?**
A: Not recommended. Single node, no HA, no backups. Use AWS or another cloud provider for production.

**Q: How do I add more services?**
A: Create Helm chart subchart or add to `services.yaml` template. Then build Docker image and push to registry.

**Q: My pods are stuck in `CrashLoopBackOff`. Help!**
A: Run `kubectl logs <pod> -n flyfleet` to see error. Common causes: missing env vars, DB connection failure, OOM kill.

**Q: How do I update to latest code?**
A: Rebuild Docker image and push to registry:
```bash
docker build -t localhost:30000/flyfleet/fleet-core-service:latest .
docker push localhost:30000/flyfleet/fleet-core-service:latest
helm upgrade --install flyfleet infrastructure/helm/flyfleet --namespace flyfleet --set global.imageTag="latest"
```

**Q: Can I use k3s in a VM instead of local machine?**
A: Yes! Install k3s on any Linux VM. Follow same steps. Ensure ports 6443, 30000, 8084 are accessible.

**Q: How do I expose the API to my network?**
A: Change service type from `ClusterIP` to `NodePort` or configure Traefik ingress with your host's IP. Edit Helm values:
```yaml
fleet-core-service:
  service:
    type: NodePort
    nodePort: 30080  # Optional (30000-32767 range)
```

---

## Next Steps

After successful self-hosted deployment:

1. **Run tests:** `make test`
2. **Try the API:** Explore endpoints at `http://localhost:8084/swagger`
3. **Set up monitoring:** Deploy Prometheus stack
4. **Read the docs:** See `DEVOPS_FIXES_SUMMARY.md` for AWS deployment
5. **Contribute:** Fix bugs, add features, improve documentation

---

## Getting Help

- **Check logs first:** `kubectl logs -n flyfleet deployment/flyfleet-fleet-core-service -f`
- **GitHub Issues:** https://github.com/your-org/flyfleet/issues
- **Slack:** `#devops` or `#flyfleet` channels
- **Documentation:** `docs/` directory in repo

---

**Happy deploying! 🚀**
