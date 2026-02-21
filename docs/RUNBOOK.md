# FlyFleet Architecture Runbook

## Service Map

| Service                 | Port (HTTP) | Port (gRPC) | Database      | Purpose                              |
| ----------------------- | ----------- | ----------- | ------------- | ------------------------------------ |
| `auth-service`          | 8080        | 50051       | `fleet_db`    | Authentication, user management      |
| `fleet-core-service`    | 8084        | 50052       | `fleet_db`    | Vehicles, drivers, maintenance, fuel |
| `tracking-service`      | 8082        | 50053       | `tracking_db` | Telemetry ingestion, GPS tracking    |
| `web-bff`               | 8081        | —           | —             | API gateway, route aggregation       |
| `admin-support-service` | 8086        | —           | `support_db`  | Tickets, DR, incidents               |
| `billing-service`       | 8087        | —           | `billing_db`  | Invoicing, payment processing        |
| `alert-service`         | 8085        | —           | `alert_db`    | Notifications, preferences           |
| `settings-service`      | 8091        | —           | `settings_db` | Platform configuration               |

## Communication Patterns

```
┌─────────────┐     gRPC      ┌──────────────────┐
│   Web BFF   │──────────────▶│   Auth Service   │
│   :8081     │               │   :50051         │
│             │──── gRPC ────▶│ Fleet Core :50052│
│             │──── gRPC ────▶│ Tracking :50053  │
│             │── HTTP Proxy─▶│ Other Services   │
└─────────────┘               └──────────────────┘
                                      │
                                      ▼ NATS JetStream
                              ┌──────────────────┐
                              │   Event Bus      │
                              │   Stream: FLEET  │
                              └──────────────────┘
```

## Incident Response

### Service Unhealthy

1. Check pod status: `kubectl get pods -n flyfleet -l app=<service>`
2. Check logs: `kubectl logs -n flyfleet deployment/<service> --tail=100`
3. Check probes: `kubectl describe pod <pod-name> -n flyfleet`
4. Restart: `kubectl rollout restart deployment/<service> -n flyfleet`

### Database Connection Exhaustion

1. Check pool usage: `SELECT count(*) FROM pg_stat_activity WHERE datname='<db>';`
2. Kill idle connections: `SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state='idle' AND query_start < NOW() - INTERVAL '10 minutes';`
3. Restart affected service pods

### NATS Queue Backlog

1. Check stream info: `nats stream info FLEET`
2. Check consumer lag: `nats consumer info FLEET <consumer>`
3. If backlog > 100k: scale tracking-service replicas

### Telemetry Ingestion Latency

1. Target SLO: < 100ms p99 for GPS ingestion
2. Check: `curl -s http://tracking-service:8082/metrics | grep telemetry_latency`
3. Scale if needed: `kubectl scale deployment tracking-service --replicas=5 -n flyfleet`

## Data Retention Policies

| Table               | Hot (uncompressed) | Warm (compressed) | Purge    |
| ------------------- | ------------------ | ----------------- | -------- |
| `gps_locations`     | 30 days            | 90 days           | 90 days  |
| `obd_readings`      | 14 days            | 60 days           | 60 days  |
| `fuel_readings`     | —                  | —                 | 180 days |
| `device_heartbeats` | 7 days             | —                 | 7 days   |
| `audit_logs`        | —                  | —                 | 365 days |

## Secrets

All secrets are injected via environment variables. Source:

- **Local**: `docker-compose.yml` env section
- **K8s**: `flyfleet-secrets` Secret resource
- **Production**: HashiCorp Vault or AWS Secrets Manager

Required secrets per service:

- `JWT_SECRET` — shared across all services
- `DATABASE_URL` — per-service connection string
- `NATS_URL` — shared NATS connection
- `REDIS_ADDR` — shared Redis instance

## Deployment Checklist

- [ ] All CI checks pass (SAST, Trivy, tests)
- [ ] Database migrations applied
- [ ] Feature flags set for gradual rollout
- [ ] PDB ensures min 2 replicas
- [ ] Health probes verified
- [ ] Rollback plan documented
