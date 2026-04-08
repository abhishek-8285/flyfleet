# FlyFleet Full Stack - Quick Reference

## 🚀 Start Everything

```bash
# Full observability stack (OpenSearch, ClickHouse, Grafana, etc.)
make full-stack-up

# Or directly:
docker-compose -f docker-compose.full.yml up -d
./infrastructure/setup-full-stack.sh
```

---

## 📍 Service URLs

| Service | URL | Login |
|---------|-----|-------|
| **FlyFleet API** | http://localhost:8081 | Bearer token |
| **API Docs (Swagger)** | http://localhost:8081/swagger | - |
| **Grafana** | http://localhost:3000 | admin / admin |
| **OpenSearch Dashboards** | http://localhost:5601 | admin / admin |
| **Jaeger (Traces)** | http://localhost:16686 | - |
| **Prometheus** | http://localhost:9090 | - |
| **Loki (Logs)** | http://localhost:3100 | - |
| **Vault** | http://localhost:8200 | dev-root-token |
| **ClickHouse** | http://localhost:8123 | flyfleet / flyfleet_password |
| **OpenSearch** | http://localhost:9200 | admin / admin |
| **PostgreSQL** | localhost:5433 | fleet_user / fleet_password |
| **Redis** | localhost:6380 | - |
| **NATS** | localhost:4222 | - |

---

## 📝 Useful Commands

```bash
# Check health of all services
make full-stack-health
./infrastructure/health-check.sh

# View logs
make full-stack-logs                    # All services
make full-stack-logs-grafana           # Specific service
docker-compose -f docker-compose.full.yml logs -f fleet-core-service

# Restart services
make full-stack-restart
docker-compose -f docker-compose.full.yml restart grafana

# Stop everything
make full-stack-down
docker-compose -f docker-compose.full.yml down -v  # With volumes

# Check status
make full-stack-status
docker-compose -f docker-compose.full.yml ps

# Rebuild backend (after code changes)
cd services/fleet-core && ~/go/bin/air
```

---

## 🔍 Searching Data

### OpenSearch (Full-Text Search)

```bash
# Search invoices
curl -XGET http://localhost:9200/invoices-*/_search -H 'Content-Type: application/json' -d '{
  "query": {
    "multi_match": {
      "query": "ABC Logistics",
      "fields": ["customer_name", "invoice_number"]
    }
  }
}'
```

**Index new document**:
```bash
curl -XPOST http://localhost:9200/invoices-2024-03/_doc/1 -H 'Content-Type: application/json' -d '{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "invoice_number": "INV-2024-0001",
  "customer_name": "ABC Logistics",
  "total_amount": 15000,
  "status": "paid",
  "created_at": "2024-03-15T10:30:00Z"
}'
```

---

## 📊 Analytics with ClickHouse

### Connect

```bash
# Via HTTP
curl -X POST http://localhost:8123 \
  -u flyfleet:flyfleet_password \
  -d "SELECT * FROM trip_analytics_hourly LIMIT 5"

# Via CLI (into container)
docker exec -it fleet-clickhouse clickhouse-client --database flyfleet_analytics
```

### Sample Queries

```sql
-- Daily revenue (last 30 days)
SELECT
    date,
    sum(total_revenue) as revenue,
    count(*) as invoices
FROM daily_revenue_summary
WHERE org_id = 'your-org-uuid'
GROUP BY date
ORDER BY date DESC
LIMIT 30;

-- Top vehicles by distance
SELECT vehicle_id, sum(total_distance_km) as km
FROM trip_analytics_hourly
GROUP BY vehicle_id
ORDER BY km DESC
LIMIT 10;

-- Payment method breakdown
SELECT payment_method, count(*) as cnt, sum(amount) as total
FROM payment_analytics
WHERE created_at >= now() - INTERVAL 30 DAY
GROUP BY payment_method
ORDER BY total DESC;
```

---

## 📈 Grafana Dashboards

### Login
- URL: http://localhost:3000
- User: `admin` | Password: `admin`

### Pre-loaded Dashboards
1. **FlyFleet Application Overview** - RPS, errors, latency, DB connections
2. (More to be added...)

### Add Your Own Dashboard
Place JSON files in `infrastructure/grafana/dashboards/`. Grafana auto-imports them on startup.

---

## 🔭 Tracing (Jaeger)

### Instrument Your Code

FlyFleet already has OpenTelemetry. Ensure `OTEL_EXPORTER_OTLP_ENDPOINT` is set:

```go
import "go.opentelemetry.io/otel"

func init() {
    otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
    if otlpEndpoint == "" {
        otlpEndpoint = "http://localhost:4317"
    }
    // ... configure tracer provider
}
```

### View Traces
1. Go to http://localhost:16686
2. Select service: `flyfleet`
3. Click "Find Traces"
4. Click a trace to see spans

---

## 🔐 Vault Secrets

### Access
- URL: http://localhost:8200
- Root Token: `dev-root-token` (dev only)

### CLI

```bash
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=dev-root-token

# List secrets
vault kv list secret/flyfleet/

# Get database creds
vault kv get secret/flyfleet/database

# Update a secret
vault kv put secret/flyfleet/jwt secret="new-secret-key"
```

### In Application

```go
import "github.com/hashicorp/vault/api"

client, _ := api.NewClient(&api.Config{Address: "http://localhost:8200"})
client.SetToken("dev-root-token")

secret, _ := client.KVv2("secret").Get(context.Background(), "flyfleet/database")
password := secret.Data["password"].(string)
```

---

## 📋 Logs with Loki

### Query via HTTP

```bash
# Search logs
curl -G 'http://localhost:3100/loki/api/v1/query' \
  --data-urlencode 'query={service="fleet-core"} |= "ERROR"' \
  --data-urlencode 'limit=20'

# Logs for a specific trace
curl -G 'http://localhost:3100/loki/api/v1/query' \
  --data-urlencode 'query={trace_id="abc123"}' \
  --data-urlencode 'limit=50'
```

### In Grafana
- Open Explore
- Select **Loki** datasource
- Query: `{service="fleet-core"} | json | line_format "{{.message}}"`

---

## 🔧 Troubleshooting

### Services not starting?

```bash
# Check logs
docker-compose -f docker-compose.full.yml logs [service-name]

# Restart specific service
docker-compose -f docker-compose.full.yml restart [service]

# Check resource usage (maybe out of RAM?)
docker stats
```

### Port already in use?

```bash
# Find process
lsof -i :9200
# Kill it
kill -9 <PID>

# Or change port in docker-compose.full.yml
```

### Grafana can't connect to Prometheus?

```bash
# Verify Prometheus is up
curl http://localhost:9090/-/healthy

# In Grafana, go to Configuration → Data Sources → Prometheus
# Test connection
```

### OpenSearch out of memory?

Increase heap in docker-compose.full.yml:
```yaml
environment:
  - "OPENSEARCH_JAVA_OPTS=-Xms2g -Xmx4g"
```

---

## 📦 What's Running

| Service | Version | Purpose |
|---------|---------|---------|
| PostgreSQL | 15 (TimescaleDB) | Primary database |
| Redis | 7 | Cache, rate limiting |
| NATS | 2.10 | Event streaming |
| OpenSearch | 2.14 | Search engine |
| ClickHouse | 24.3 | Analytics DB |
| Prometheus | Latest | Metrics |
| Grafana | Latest | Dashboards |
| Loki | Latest | Logs |
| Jaeger | Latest | Traces |
| OTel Collector | 0.102.0 | Telemetry pipeline |
| Vault | 1.16 | Secrets |
| Fleet Core | Go 1.26 | Application |

**ALL 100% FREE & Open Source**

---

## 📚 More Information

See **INFRASTRUCTURE_FULL_STACK.md** for comprehensive documentation.

---

## 💾 Resource Usage (Approximate)

- **RAM**: ~7.5GB total
- **Disk**: ~40GB allocated (grow based on data)
- **CPU**: 2-4 cores recommended

**Minimum requirements**: 8GB RAM, 50GB disk
**Recommended**: 16GB RAM, 100GB SSD

---

**Last Updated**: 2026-03-29
