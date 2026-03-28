# Metrics Endpoint Verification Guide

## Quick Verification Steps

After the observability enhancement fix, follow these steps to verify all metrics are properly exposed.

### Prerequisites

Ensure infrastructure is running:
- PostgreSQL (on port 5432 or unix socket)
- Redis (default port 6379 or 6380)
- NATS JetStream (port 4222)

### Option 1: Using startup script (Termux native)

```bash
cd /data/data/com.termux/files/home/flyfleet

# Build the backend (if not already built)
cd services/fleet-core
go build -o main ./cmd

# Start infrastructure (Postgres already running; start Redis and NATS)
redis-server --port 6380 --daemonize yes
nats-server -js -p 4222 > /dev/null 2>&1 &
go run scripts/setup_nats.go

# Return to project root
cd /data/data/com.termux/files/home/flyfleet

# Start backend with proper env
export DATABASE_URL="host=/data/data/com.termux/files/usr/tmp user=u0_a412 dbname=fleet_db sslmode=disable"
./services/fleet-core/main > /tmp/fleet-core.log 2>&1 &
```

### Option 2: Using Docker Compose

```bash
cd /data/data/com.termux/files/home/flyfleet
docker-compose up -d postgres redis nats
cd services/fleet-core
go build -o main ./cmd
make run  # or however the service is started in docker mode
```

### Verify Metrics Endpoint

Once the service is running:

```bash
# Check that the server is up
curl -s http://localhost:8081/health | jq .

# Fetch metrics
curl -s http://localhost:8081/metrics | head -50

# Count total metric lines (excluding comments and HELP/TYPE)
curl -s http://localhost:8081/metrics | grep -E '^[^#]' | wc -l

# Search for specific critical metrics
curl -s http://localhost:8081/metrics | grep -E '^(telemetry_events_received_total|device_command_events_total|active_trips|trips_started_total|auth_login_success_total)' | head -10
```

### Expected Results

- **No 500 errors** from `/metrics` endpoint
- **At least 600+ metrics** should be listed (the exact count depends on label combinations)
- **Key SLO metrics** must be present:
  - `telemetry_events_received_total`
  - `telemetry_ingestion_latency_seconds`
  - `telemetry_gaps_total`
  - `telemetry_expected_total`
  - `device_command_events_total`
  - `device_command_ack_latency_seconds`
  - `auth_login_success_total`
  - `auth_login_attempts_total`
  - `trips_started_total`
  - `active_trips`
  - `organizations_active_total`
- **All 27 domains** should contribute at least one metric (look for domain prefixes)

### Troubleshooting

#### Binary panics on startup with `ExponentialBuckets needs a positive start value`

This indicates a histogram bucket configuration error. Common fixes:
- Ensure all `prometheus.ExponentialBuckets` calls have a positive start value (>0)
- For metrics that need to handle negative values, use `prometheus.LinearBuckets` instead.
- Check domains: `route` (ETA accuracy) and `trip` (waypoint arrival variance) have been fixed.

#### Database connection errors

Verify `DATABASE_URL` is set correctly and PostgreSQL is running:
```bash
echo $DATABASE_URL
pg_isready -h /data/data/com.termux/files/usr/tmp -p 5432 -U u0_a412 -d fleet_db
```

#### Metrics not appearing

Ensure all domain packages are imported in `cmd/main.go` as blank imports. Check build output for any warnings about missing packages.

#### Port already in use

Change port via environment variable `PORT=8082` or edit the server configuration.

### Automated Check Script

You can run this quick check:

```bash
#!/bin/bash
set -e
echo "Checking /metrics endpoint..."
METRICS=$(curl -s http://localhost:8081/metrics)
echo "Sample metrics:"
echo "$METRICS" | grep -E '^(telemetry_events_received_total|device_command_events_total|auth_login_success_total|trips_started_total|active_trips)' | head -5
echo ""
echo "Total non-comment lines: $(echo "$METRICS" | grep -E '^[^#]' | wc -l)"
```
