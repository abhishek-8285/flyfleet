# FlyFleet Observability Setup Guide

**Complete guide to deploying Prometheus, Grafana, and Alertmanager for FlyFleet monitoring**

Last updated: 2026-03-28

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Installation](#installation)
4. [Configuration](#configuration)
5. [Grafana Setup](#grafana-setup)
6. [Alertmanager Setup](#alertmanager-setup)
7. [Dashboards](#dashboards)
8. [Testing](#testing)
9. [Maintenance](#maintenance)
10. [Troubleshooting](#troubleshooting)

---

## Overview

This guide deploys the **kube-prometheus-stack** Helm chart, which includes:

| Component | Purpose | Default Port |
|-----------|---------|--------------|
| **Prometheus** | Metrics collection & alerting | 9090 |
| **Grafana** | Visualization & dashboards | 3000 |
| **Alertmanager** | Alert routing & notifications | 9093 |
| **Node Exporter** | Node-level metrics | 9100 |
| **kube-state-metrics** | Kubernetes resource metrics | 8080 |
| **Blackbox Exporter** | HTTP endpoint probing | 9115 |

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │ Prometheus   │  │ Alertmanager │  │   Grafana    │   │
│  │  (9090)      │  │  (9093)      │  │  (3000)      │   │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘   │
│         │                 │                 │           │
│         └─────────────────┼─────────────────┘           │
│                           │                             │
│         ┌─────────────────┼─────────────────┐           │
│         │                 │                 │           │
│  ┌──────▼───────┐  ┌────▼──────┐  ┌──────▼──────┐    │
│  │Node Exporter │  │kube-state │  │Blackbox     │    │
│  │  (9100)      │  │metrics    │  │Exporter     │    │
│  │              │  │  (8080)   │  │  (9115)     │    │
│  └──────────────┘  └───────────┘  └─────────────┘    │
│                                                             │
│  ┌──────────────────────────────────────────────────┐   │
│  │          FlyFleet Application Pods               │   │
│  │  ┌──────────────┐  ┌──────────────┐             │   │
│  │  │ fleet-core   │  │  postgres    │             │   │
│  │  │   (8084)     │  │   (5432)     │             │   │
│  │  └──────────────┘  └──────────────┘             │   │
│  └──────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Prerequisites

### Required

- ✅ Kubernetes cluster (EKS, k3s, or any other)
- ✅ Helm 3.12+ installed and configured
- ✅ kubectl configured with cluster access
- ✅ `kube-prometheus-stack` chart repository added

### Optional

- Ingress controller (for external Grafana access)
- TLS certificates (for secure access)
- Slack/PagerDuty integration (for alerts)

---

## Installation

### Step 1: Add Helm Repository

```bash
# Add prometheus-community repo
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Verify
helm search repo prometheus-community/kube-prometheus-stack
```

### Step 2: Create Monitoring Namespace

```bash
kubectl create namespace monitoring
```

### Step 3: Prepare Custom Values

```bash
# Copy the example values file
cp infrastructure/monitoring/prometheus-values.yaml infrastructure/monitoring/prometheus-values.custom.yaml

# Edit the values file:
# - Change grafana.adminPassword to a secure password
# - Adjust storage sizes (default 50Gi for Prometheus, 10Gi for Grafana)
# - Configure Alertmanager receivers (Slack, email, PagerDuty)
# - Set ingress.enabled to true if you have an ingress controller
# - Update notification channels (Slack webhook URLs, email credentials)
```

### Step 4: Install Helm Chart

```bash
# Basic installation
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  -f infrastructure/monitoring/prometheus-values.custom.yaml

# Wait for all pods to be ready (takes 2-5 minutes)
kubectl wait --for=condition=ready pod -l app.kubernetes.io/instance=prometheus -n monitoring --timeout=300s

# Check status
helm status prometheus -n monitoring
kubectl get pods -n monitoring
```

---

## Configuration

### Prometheus Configuration

#### Custom Scrape Configs

The `prometheusValues.yaml` includes a scrape config for FlyFleet API. To add more services:

```yaml
prometheus:
  prometheusSpec:
    additionalScrapeConfigs:
      - job_name: 'flyfleet-api'
        # ... existing config
      - job_name: 'nats'
        static_configs:
          - targets: ['fleet-nats.flyfleet.svc.cluster.local:8222']
        metrics_path: '/metricsz'
```

#### Recording Rules (Optional)

Create additional recording rules for frequently-queried metrics:

```yaml
prometheus:
  prometheusSpec:
    ruleSelector:
      matchLabels:
        role: alertmanager
    additionalAlertManagerConfigs:
      - <your-config>
```

### Grafana Configuration

#### Import Dashboards

Grafana automatically loads dashboards from:
1. The chart's default dashboards
2. ConfigMaps labeled with `grafana_dashboard`
3. JSON files in the `dashboards` folder (if sidecar enabled)

To import the FlyFleet dashboard:

```bash
# Create ConfigMap with FlyFleet dashboard
kubectl create configmap flyfleet-dashboard \
  --from-file=flyfleet-overview.json=infrastructure/monitoring/grafana-dashboard-flyfleet.json \
  -n monitoring

# Add label for sidecar
kubectl label configmap flyfleet-dashboard grafana_dashboard=flyfleet -n monitoring
```

#### Configure Data Sources

The `grafana-datasources.yaml` ConfigMap provides Prometheus datasource:

```bash
kubectl apply -f infrastructure/monitoring/grafana-datasources.yaml -n monitoring
```

### Alertmanager Configuration

The `alertmanager-config.yaml` provides:

- Email notifications
- Slack integration
- PagerDuty integration
- Routing based on severity

**Configure notification channels:**

1. **Slack:** Replace `YOUR/SLACK/WEBHOOK` with your Slack incoming webhook URL
2. **Email:** Update SMTP credentials in the config
3. **PagerDuty:** Replace integration keys

Apply the config:

```bash
kubectl create configmap alertmanager-config \
  --from-file=alertmanager-config.yaml=infrastructure/monitoring/alertmanager-config.yaml \
  -n monitoring

kubectl label configmap alertmanager-config alertmanager=config -n monitoring
```

---

## Grafana Setup

### Access Grafana

**Option A: Port-forward (quick access)**

```bash
kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80
# Open browser: http://localhost:3000
```

**Option B: Ingress (external access)**

If you have an ingress controller:

```yaml
# Add to prometheus-values.custom.yaml:
grafana:
  ingress:
    enabled: true
    hosts:
      - grafana.flyfleet.com
    annotations:
      kubernetes.io/ingress.class: nginx
      nginx.ingress.kubernetes.io/auth-type: basic
      nginx.ingress.kubernetes.io/auth-secret: grafana-basic-auth
      nginx.ingress.kubernetes.io/auth-realm: "Grafana Authentication"
```

Then:
```bash
kubectl get ingress -n monitoring
# Open: https://grafana.flyfleet.com
```

**Login Credentials:**
- Username: `admin`
- Password: The value you set in `grafana.adminPassword`

### Explore Dashboards

1. Click **Dashboards** in left sidebar
2. Select **FlyFleet** folder
3. Open **FlyFleet Overview** dashboard
4. You should see metrics from your cluster

### Create Additional Dashboards

```bash
# Export dashboard from Grafana UI (JSON)
# Save to infrastructure/monitoring/grafana-dashboard-*.json

# Create ConfigMap
kubectl create configmap my-dashboard \
  --from-file=my-dashboard.json=./my-dashboard.json \
  -n monitoring
kubectl label configmap my-dashboard grafana_dashboard=flyfleet -n monitoring

# Grafana will auto-reload (or restart the pod)
```

---

## Alertmanager Setup

### Test Alerts

```bash
# Port-forward Alertmanager
kubectl port-forward -n monitoring svc/prometheus-alertmanager 9093:9093 &
# Open: http://localhost:9093

# View active alerts
kubectl get alerts -n monitoring

# Manually fire a test alert (for testing purposes)
kubectl apply -f - <<EOF
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: test-alert
  namespace: monitoring
spec:
  groups:
    - name: test
      rules:
        - alert: TestAlert
          expr: vector(1)
          for: 1m
          labels:
            severity: critical
          annotations:
            summary: "This is a test alert"
            description: "Testing Alertmanager configuration"
EOF
```

After 1 minute, you should receive notifications per your Alertmanager config.

### Silence Alerts

From Alertmanager UI (`http://localhost:9093`):
1. Click **Alerts** tab
2. Find the alert
3. Click ** Silence ** → Set duration and reason
4. Click **Create Silence**

Or via `amtool`:

```bash
# Install amtool
brew install prometheus/alertmanager/amtool

# Add silence
amtool --alertmanager.url=http://localhost:9093 silence add --author="ops@flyfleet.com" --comment="Maintenance" --duration=30m alertname=TestAlert
```

---

## Dashboards

### Built-in Dashboards

The kube-prometheus-stack includes these pre-configured:

| Dashboard | Description |
|-----------|-------------|
| **Kubernetes / Compute Resources** | Node/Pod CPU, memory, I/O |
| **Kubernetes / API Server** | API server latency, request rate |
| **Kubernetes / Networking** | Network I/O, pod communications |
| **Kubernetes / etcd** | etcd performance |
| **Node Exporter / System** | Node-level metrics (disk, network, CPU) |

### FlyFleet Custom Dashboard

Import `grafana-dashboard-flyfleet.json` for application-specific metrics:

- Request rate & error rate
- API latency (p95, p99)
- Database connections
- NATS message rates
- Pod status & resource usage
- Redis memory usage

To customize:
1. Open dashboard in Grafana
2. Click **Settings** (gear icon)
3. Edit panels/queries
4. Save and **Export** JSON to repo

---

## Testing

### Verify Metrics Collection

```bash
# Query Prometheus directly
kubectl port-forward -n monitoring svc/prometheus-operated 9090:9090 &
curl 'http://localhost:9090/api/v1/query?query=up'

# Should return:
# {"status":"success","data":{"resultType":"vector","result":[{"metric":{...},"value":[timestamp,1]}]}}

# Check specific metrics
curl 'http://localhost:9090/api/v1/query?query=http_requests_total'
curl 'http://localhost:9090/api/v1/query?query=pg_stat_database_numbackends'
```

### Verify Grafana

```bash
kubectl get pods -l app.kubernetes.io/instance=prometheus -n monitoring
# All pods should be in Running state

kubectl logs -n monitoring -l app.kubernetes.io/component=grafana
# Should show no errors
```

### Verify Alertmanager

```bash
kubectl get pods -l app.kubernetes.io/instance=alertmanager -n monitoring

# Check config is loaded
kubectl exec -n monitoring -it deployment/prometheus-alertmanager -- cat /etc/alertmanager/alertmanager.yml | head -50
```

### Smoke Tests

Run the provided smoke test script:

```bash
# Test endpoints
./scripts/smoke_monitoring.sh

# Or manually:
curl -s http://localhost:9090/metrics | grep 'prometheus_tsdb_head_samples'
curl -s http://localhost:3000/api/health | jq .
curl -s http://localhost:9093/-/healthy
```

---

## Maintenance

### Updating Helm Release

```bash
# Update chart repository
helm repo update

# Check for available updates
helm upgrade --install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  -f infrastructure/monitoring/prometheus-values.custom.yaml \
  --dry-run

# Perform upgrade
helm upgrade prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  -f infrastructure/monitoring/prometheus-values.custom.yaml
```

### Backing Up Data

#### Prometheus

```bash
# Create snapshot
kubectl exec -n monitoring deploy/prometheus-prometheus-operated -- \
  wget -q -O- 'http://localhost:9090/api/v1/admin/tsdb/snapshot' | jq -r '.data.name' | xargs -I {} \
  kubectl exec -n monitoring deploy/prometheus-prometheus-operated -- tar czf /tmp/snapshots/{}.tar.gz /prometheus/snapshots/

# Copy snapshot from pod
kubectl cp monitoring/prometheus-prometheus-operated-xxxx:/tmp/snapshots/ ./
```

#### Grafana

Grafana dashboards are stored in ConfigMaps or PVC. Backup ConfigMaps:

```bash
kubectl get configmap -n monitoring -l grafana_dashboard -o yaml > grafana-dashboards-backup.yaml
kubectl get secret -n monitoring grafana -o yaml > grafana-secret-backup.yaml
```

### Scaling

#### Prometheus

```yaml
# In prometheus-values.custom.yaml:
prometheus:
  prometheusSpec:
    replicas: 3  # Increase for HA and performance
```

#### Grafana

```yaml
grafana:
  replicaCount: 2
```

Apply changes:

```bash
helm upgrade prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  -f infrastructure/monitoring/prometheus-values.custom.yaml
```

### Retention & Storage

Adjust retention in values:

```yaml
prometheus:
  prometheusSpec:
    retention: "90d"  # Keep 90 days of metrics
    storageSpec:
      volumeClaimTemplate:
        spec:
          resources:
            requests:
              storage: 200Gi  # Increase for longer retention
```

---

## Troubleshooting

### Pods not starting

```bash
# Check events
kubectl get events -n monitoring --sort-by='.lastTimestamp'

# Describe failing pod
kubectl describe pod <pod-name> -n monitoring

# Check logs
kubectl logs <pod-name> -n monitoring
```

### Prometheus not collecting metrics

```bash
# Check scrape configs
kubectl port-forward -n monitoring svc/prometheus-operated 9090:9090 &
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'

# Target should show "up" = 1
```

### Grafana shows "Data source is unavailable"

1. Verify Prometheus is running:
   ```bash
   kubectl get pods -l app.kubernetes.io/instance=prometheus -n monitoring
   ```

2. Check datasource config:
   ```bash
   kubectl get configmap -n monitoring grafana-datasources
   ```

3. Restart Grafana pod:
   ```bash
   kubectl rollout restart deployment/prometheus-grafana -n monitoring
   ```

### Alertmanager not sending notifications

1. Verify webhook URLs and credentials in Alertmanager config
2. Check Alertmanager logs:
   ```bash
   kubectl logs -n monitoring deploy/prometheus-alertmanager
   ```
3. Test notification manually:
   ```bash
   kubectl port-forward -n monitoring svc/prometheus-alertmanager 9093:9093
   curl -XPOST http://localhost:9093/api/v2/alerts -d '[{"labels":{"alertname":"TestAlert","severity":"critical"},"annotations":{"summary":"Test"}}]'
   ```

### Out of disk space

```bash
# Check PVC usage
kubectl get pvc -n monitoring

# Increase storage:
# Edit prometheus-values.custom.yaml → storage size
helm upgrade prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  -f infrastructure/monitoring/prometheus-values.custom.yaml
```

---

## Production Considerations

### High Availability

Enable multiple replicas:

```yaml
prometheus:
  prometheusSpec:
    replicas: 3
    volumeClaimTemplate:
      spec:
        storageClassName: "gp2"  # Use shared storage like EBS gp3
        accessModes: ["ReadWriteMany"]  # Not supported with local storage

alertmanager:
  replicas: 3

grafana:
  replicaCount: 2
```

### Security

1. **Enable authentication:**
   ```yaml
   grafana:
     grafana.ini:
       security:
         admin_user: "admin"
         admin_password: "SECURE_PASSWORD"
       auth.basic:
         enabled: true
   ```

2. **Use TLS:**
   ```yaml
   grafana:
     grafana.ini:
       server:
         root_url: "https://grafana.flyfleet.com"
         enforce_domain: true
         protocol: https
         cert_file: /etc/grafana/certs/tls.crt
         cert_key: /etc/grafana/certs/tls.key
   ```

3. **Restrict network access:**
   - Use security groups / network policies
   - Deploy only in private subnets
   - Use VPN or SSO for access

### Long-term Storage

For metrics retention >90 days, consider:

1. **Thanos** (S3/object storage)
2. **Cortex** (distributed long-term storage)
3. **Mimir** (Grafana Labs' managed long-term store)

---

## Uninstallation

```bash
# Delete Helm release
helm uninstall prometheus -n monitoring

# Delete namespace (with PVC cleanup)
kubectl delete namespace monitoring --wait=false

# Delete persistent volumes (WARNING: ERASES ALL DATA)
kubectl delete pvc -n monitoring --all
```

---

## Next Steps

After successful deployment:

1. **Import dashboards** from `grafana-dashboard-flyfleet.json`
2. **Configure alert receivers** (Slack, email, PagerDuty)
3. **Set up alert silencing** for maintenance windows
4. **Create custom dashboards** for business metrics
5. **Configure notification templates** for your team
6. **Set up retention policies** to manage storage costs
7. **Add additional scrape targets** for new services
8. **Enable TLS and authentication** for production

---

## Resources

- [kube-prometheus-stack Helm chart](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)
- [Prometheus documentation](https://prometheus.io/docs/)
- [Grafana documentation](https://grafana.com/docs/grafana/latest/)
- [Alertmanager documentation](https://prometheus.io/docs/alerting/latest/alertmanager/)
- [FlyFleet Monitoring Reference](https://github.com/your-org/flyfleet/infrastructure/monitoring)

---

**Last Updated:** 2026-03-28
**Maintainer:** FlyFleet Platform Team
