# FlyFleet Incident Response Runbooks

This document outlines standardized recovery procedures for common platform incidents.

## 1. Telemetry Backlog Recovery

**Symptoms**: High worker queue depth, delayed tracking updates.
**Steps**:

1. Check NATS/Redis queue health.
2. Scale `diagnostics-worker` replicas:
   ```bash
   kubectl scale deployment diagnostics-worker --replicas=20 -n flyfleet
   ```
3. Monitor `telemetry_ingestion_rate` in Grafana.
4. Scale back once lag is cleared.

## 2. Database Failover (PostgreSQL)

**Symptoms**: Application "Connection Refused" errors.
**Steps**:

1. Verify Patroni/HA status.
2. If automatic failover failed, manually promote a replica:
   ```bash
   # Example for managed HA
   kubectl exec -it pg-0 -- patronictl promote
   ```
3. Verify new primary is accepting writes.
4. Check application logs for reconnection.

## 3. Worker Crash Recovery

**Symptoms**: Periodic "Device Offline" spikes, telemetry stop processing.
**Steps**:

1. Inspect worker logs for OOM or panic:
   ```bash
   kubectl logs -l app=diagnostics-worker -n flyfleet --tail=100
   ```
2. If OOM, increase memory limits in `infrastructure/k8s/worker-deployment.yaml`.
3. Force rollout of workers:
   ```bash
   kubectl rollout restart deployment diagnostics-worker -n flyfleet
   ```

## 4. Firmware Rollout Backlog

**Symptoms**: High rollout failure rate, devices stuck in `DOWNLOADING`.
**Steps**:

1. Pause the rollout via API.
2. Check CDN/Artifact storage availability.
3. Validate firmware signature integrity.
4. Resume with smaller batch size.

## 5. Global Platform Rollback

**Symptoms**: Critical regression after production deployment.
**Steps**:

1. Revert `fleet-core-service` to last stable tag:
   ```bash
   kubectl set image deployment/fleet-core-service *=flyfleet/core:vX.Y.Z-stable
   ```
2. Monitor API latencies and error rates.
3. Verify database migrations don't require rollback.
