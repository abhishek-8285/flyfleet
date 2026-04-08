# Instrumentation Task Checklist

**Priority Order:** Critical SLO metrics first, then business metrics, then operational metrics.

## Critical SLO Metrics (Must-Have for Production)

### 1. Telemetry Ingestion Domain ✅/❌
**Related SLOs:** telemetry-throughput, telemetry-latency, telemetry-data-loss

- [ ] **`telemetry_events_received_total`** - Increment in `IngestionService.Ingest()` after successful processing
- [ ] **`telemetry_ingestion_latency_seconds`** - Record histogram in `IngestionService.Ingest()`: time from receipt to persistence
- [ ] **`telemetry_gaps_total`** - Detect and increment when sequence gaps detected (gap detection logic)
- [ ] **`telemetry_expected_total`** - Increment based on expected sequence numbers per device/session

**File:** `services/fleet-core/internal/domains/tracking/ingestion_service.go`

### 2. Trip Domain ✅/❌
**Related SLOs:** trips-active, trips-completion-rate

- [ ] **`trips_started_total`** - Increment in `TripService.Start()` when trip begins
- [ ] **`active_trips` gauge** - Periodically update (every 30s?) from in-memory cache or DB query
  - Options: cache in Redis with TTL, or maintain in-memory counter with atomic operations

**Files:**
- `services/fleet-core/internal/domains/trip/service.go` (TripService.Start)
- `services/fleet-core/internal/domains/trip/` (new helper or periodic updater)

### 3. Command & Control Domain ✅/❌
**Related SLOs:** command-success-rate, command-latency, command-queue-depth

_Status: Already defined in platform metrics. Need to verify instrumentation:_

- [ ] **`device_command_events_total`** - Verify increment in `device_command_service.go` on command sent/attempted
- [ ] **`device_command_ack_latency_seconds`** - Verify histogram recording when ack received
- [ ] **`device_command_queue_depth`** - Gauge for current queue depth (NATS JetStream backlog)

**File:** `services/fleet-core/internal/platform/device_command_service.go`

### 4. Auth Domain ✅/❌
**Related SLOs:** api-availability, api-errors (indirect)

- [ ] **`auth_login_success_total`** - Increment on successful login
- [ ] **`auth_login_attempts_total`** - Increment on every login attempt
- [ ] **`auth_login_failures_total`** - Increment on failed login (could be derived from attempts - success)

**File:** `services/fleet-core/internal/domains/auth/service.go`

### 5. Database Availability ✅/❌
**Related SLOs:** database-availability, database-latency, database-connection-utilization

- [ ] **`database_up` gauge** - Set to 1 when DB connection pool healthy, 0 when down (periodic health check)
- [ ] **`database_connection_utilization_percentage`** - Gauge: (active_connections / max_pool_size) * 100
- [ ] **`database_latency_seconds`** - Histogram for query execution time (wrap DB calls)

**File:** `services/fleet-core/internal/platform/database.go` or connection pool wrapper

---

## High-Value Business Metrics

### 6. Organizations Domain
- [ ] **`organizations_active_total`** - Already defined in platform metrics. Increment when org status becomes active, decrement when inactive.

**File:** `services/fleet-core/internal/domains/customer/organization_repository.go` or service

### 7. Driver Domain
- [ ] **`drivers_active_total`** - Gauge for currently active (online) drivers
- [ ] `drivers_onboarding_total` - Counter for driver registrations
- [ ] `driver_behavior_score` - Gauge for average driver behavior score

**File:** `services/fleet-core/internal/domains/driver/`

### 8. Vehicle Domain
- [ ] **`vehicles_active_total`** - Gauge for vehicles currently active/in-use
- [ ] `vehicles_available_total` - Gauge for available vehicles
- [ ] `vehicle_telemetry_messages_total` - Already defined? Check `services/fleet-core/internal/domains/vehicle/metrics.go`

**File:** `services/fleet-core/internal/domains/vehicle/`

### 9. Billing Domain
**Related SLO:** invoices-generation-rate

- [ ] **`invoices_generated_total`** - Increment when invoice created
- [ ] `invoice_amount_total` - Histogram or summary for invoice amounts
- [ ] `payment_success_total` / `payment_failure_total` - Payment processing metrics

**File:** `services/fleet-core/internal/domains/billing/`

### 10. Emergency/SOS Domain
- [ ] **`sos_alerts_total`** - Counter for SOS triggers
- [ ] `sos_response_time_seconds` - Time from SOS to first response
- [ ] `emergency_incidents_total` - Incidents created

**File:** `services/fleet-core/internal/domains/emergency/`

---

## Observability & Platform Metrics (Verification Needed)

### 11. NATS JetStream metrics
- [ ] **`nats_consumer_lag`** - Already defined? Ensure consumer lag is being reported
- [ ] **`nats_storage_utilization_bytes`** - JetStream storage usage
- [ ] **`nats_throughput_bytes_total`** - Bytes published/consumed

**File:** `services/fleet-core/internal/platform/nats/`

### 12. Cost Monitoring Metrics
- [ ] **`infrastructure_cost_usd`** - Gauge updated periodically (daily) from cloud billing API
- [ ] `cost_per_trip` - Calculated metric

**File:** New package or `services/fleet-core/internal/platform/cost_monitor.go`

---

## Implementation Strategy

### Phase 1: SLO-Critical (1-2 days)
Complete all items in **Critical SLO Metrics** section to enable SLO monitoring and alerting.

### Phase 2: Core Business (2-3 days)
Complete high-value business metrics for operations visibility.

### Phase 3: Operational (1-2 days)
Finish platform/infrastructure metrics for full observability.

### Integration Steps Per Metric

1. **Locate** the domain's `metrics.go` file and verify metric definition exists (should be already defined after enhancement)
2. **Find** the appropriate service/handler method where business event occurs
3. **Add** metric recording code:
   - For counters: `metric.Inc()` or `metric.With(labels).Inc()`
   - For gauges: `metric.Set(value)` or `metric.Add(delta)`
   - For histograms: `metric.Observe(value)`
4. **Test** by triggering the event and checking `/metrics` endpoint
5. **Verify** in Prometheus that metric appears with correct labels

### Notes

- Many metrics are already defined in the 27 domain `metrics.go` files but lack increment calls.
- Use domain-specific helpers (e.g., `helpers.RecordCommandSent(ctx, ...)`) if available.
- Ensure all metric operations are non-blocking and thread-safe (Prometheus client is safe for concurrent use).
- Avoid high-cardinality labels; use label values with bounded cardinality (e.g., org_id, event_type, not user_id).
- For gauges that need periodic updates, consider a background goroutine or a scheduled task.

---

**Reference:**
- SLO definitions: `infrastructure/monitoring/slo-recording-rules.yml`
- Alert rules: `infrastructure/monitoring/slo-alert-rules.yml`
- OBSERVABILITY_DESIGN.md
