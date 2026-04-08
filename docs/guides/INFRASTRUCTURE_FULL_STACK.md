# FlyFleet Full Observability & Analytics Stack

**100% Free & Open Source** - No licensing costs ever.

This stack extends FlyFleet with production-grade observability, search, and analytics capabilities inspired by Razorpay's architecture.

---

## 🚀 Quick Start

### 1. Prerequisites

- Docker + Docker Compose
- At least **8GB RAM** (16GB recommended)
- 50GB free disk space

### 2. Start Everything

```bash
# From flyfleet root directory:
docker-compose -f docker-compose.full.yml up -d

# Initialize Vault and verify services
./infrastructure/setup-full-stack.sh
```

This starts:
- ✅ PostgreSQL (TimescaleDB)
- ✅ Redis (cache)
- ✅ NATS JetStream (messaging)
- ✅ Fleet Core Backend (Go monolith)
- ✅ ML Service (Python/Go)
- ✅ OpenSearch (search engine)
- ✅ OpenSearch Dashboards (search UI)
- ✅ ClickHouse (analytics DB)
- ✅ Prometheus (metrics collector)
- ✅ Grafana (dashboards)
- ✅ Loki (log aggregation)
- ✅ Jaeger (distributed tracing)
- ✅ OpenTelemetry Collector (unified telemetry)
- ✅ Vault (secrets management)

### 3. Access UIs

| Service | URL | Purpose |
|---------|-----|---------|
| **Grafana** | http://localhost:3000 | Main monitoring dashboard (login: admin/admin) |
| **OpenSearch Dashboards** | http://localhost:5601 | Search & analytics (login: admin/admin) |
| **Jaeger** | http://localhost:16686 | Distributed traces UI |
| **Prometheus** | http://localhost:9090 | Metrics query/explorer |
| **Loki** | http://localhost:3100 | Logs (query via Grafana) |
| **Vault** | http://localhost:8200 | Secrets management (unseal required in prod) |
| **ClickHouse** | http://localhost:8123 | SQL analytics console |
| **FlyFleet API** | http://localhost:8081 | Backend API (via BFF) |
| **FlyFleet Swagger** | http://localhost:8081/swagger | API documentation |

---

## 📊 Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         FlyFleet Stack                             │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌────────────┐        ┌────────────┐      ┌──────────────┐      │
│  │   Frontend │───────▶│   BFF/API  │─────▶│  Fleet Core  │      │
│  │  (React)   │        │ (Go/Chi)   │      │  (Go Monolith)│      │
│  └────────────┘        └────────────┘      └──────────────┘      │
│                                                                     │
│                                │                                   │
│                                ▼                                   │
│                     ┌────────────────────┐                        │
│                     │  OpenTelemetry     │                        │
│                     │  Collector         │                        │
│                     └────────────────────┘                        │
│                                │                                   │
│                ┌───────────────┼───────────────┐                  │
│                ▼               ▼               ▼                  │
│         ┌───────────┐  ┌───────────┐  ┌───────────┐            │
│         │ Prometheus│  │    Jaeger │  │    Loki   │            │
│         │ (Metrics) │  │ (Traces)  │  │  (Logs)   │            │
│         └───────────┘  └───────────┘  └───────────┘            │
│                │                                           │      │
│                └───────────────┬───────────────────────────┘      │
│                                ▼                                   │
│                          ┌───────────┐                            │
│                          │  Grafana  │                            │
│                          │  Dashboards│                           │
│                          └───────────┘                            │
│                                                                     │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐                 │
│  │ PostgreSQL │  │  OpenSearch│  │ ClickHouse │                 │
│  │  (OLTP)    │  │  (Search)  │  │ (Analytics)│                 │
│  └────────────┘  └────────────┘  └────────────┘                 │
│                                                                     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐        │
│  │   Redis  │  │  NATS    │  │  Vault   │  │  ML Svc  │        │
│  │ (Cache)  │  │ (Events) │  │(Secrets) │  │ (ML/AI)  │        │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘        │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 🎯 What Each Service Does

### Core Application
| Service | Port | Description |
|---------|------|-------------|
| `fleet-core-service` | 8081 → 8084 | Main Go backend (modular monolith) |
| `fleet-ml-service` | 8000 | Machine learning service for predictions |

### Databases
| Service | Port | Description |
|---------|------|-------------|
| `postgres` | 5433 | Primary transactional database (TimescaleDB) |
| `redis` | 6380 | Caching, rate limiting, sessions |
| `open search` | 9200 | Full-text search (Elasticsearch fork) |
| `clickhouse` | 8123 | Columnar analytics database |
| `nats` | 4222/8222 | Event streaming with JetStream |

### Observability
| Service | Port | Description |
|---------|------|-------------|
| `prometheus` | 9090 | Metrics collection & alerting |
| `grafana` | 3000 | Dashboards & visualization |
| `loki` | 3100 | Log aggregation & query |
| `jaeger` | 16686 | Distributed tracing UI |
| `otel-collector` | 4317 | OpenTelemetry data collector |

### Security & Ops
| Service | Port | Description |
|---------|------|-------------|
| `vault` | 8200 | Secrets management (passwords, API keys) |

---

## 📈 Grafana Dashboards

Pre-configured dashboards:
1. **FlyFleet Application Overview** - Request rate, errors, latency, DB connections
2. **Infrastructure Monitoring** - CPU, memory, disk, network
3. **Database Performance** - PostgreSQL query stats, connection pool
4. **Business Metrics** - Invoices created, payments processed, active vehicles
5. **NATS JetStream** - Message rates, pending, consumer lag
6. **Full-Text Search** - OpenSearch query latency, index size

Access dashboards: http://localhost:3000 (login: `admin` / `admin`)

---

## 🔍 Using OpenSearch for Search

### Index Data from FlyFleet

```go
// In your code, send documents to OpenSearch:
client, _ := opensearch.NewClient(opensearch.Config{
    Addresses: []string{"http://localhost:9200"},
    Username: "admin",
    Password: "admin",
})

// Index an invoice
_, err := client.Index("invoices-2024-03",
    strings.NewReader(`{
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "invoice_number": "INV-2024-0001",
        "customer_name": "ABC Logistics",
        "total_amount": 15000,
        "status": "paid"
    }`))
```

### Search Example

```bash
# Search invoices
curl -XGET http://localhost:9200/invoices-*/_search -H 'Content-Type: application/json' -d '{
  "query": {
    "multi_match": {
      "query": "ABC",
      "fields": ["customer_name", "invoice_number"]
    }
  }
}'
```

---

## 📊 Using ClickHouse for Analytics

### Connect

```bash
# CLI
docker exec -it fleet-clickhouse clickhouse-client --database flyfleet_analytics

# Or via HTTP
curl -X POST http://localhost:8123 \
  -u flyfleet:flyfleet_password \
  -d "SELECT * FROM trip_analytics_hourly LIMIT 10"
```

### Sample Queries

```sql
-- Daily revenue
SELECT
    date,
    sum(total_revenue) as revenue,
    avg(avg_balance_due) as avg_balance
FROM daily_revenue_summary
WHERE org_id = 'your-org-uuid'
GROUP BY date
ORDER BY date DESC
LIMIT 30;

-- Top vehicles by distance
SELECT
    vehicle_id,
    sum(total_distance_km) as km
FROM trip_analytics_hourly
GROUP BY vehicle_id
ORDER BY km DESC
LIMIT 10;

-- Payment method distribution
SELECT
    payment_method,
    count(*) as count,
    sum(amount) as total
FROM payment_analytics
WHERE created_at >= now() - INTERVAL 30 DAY
GROUP BY payment_method
ORDER BY total DESC;
```

---

## 🔐 Vault: Secrets Management

### Initialize Vault (Production)

```bash
# Development: Already auto-initialized with docker-compose
# Production: Manual initialization required

vault operator init \
  -key-shares=5 \
  -key-threshold=3

# Save unseal keys and root token securely!

# Unseal
vault operator unseal <key1>
vault operator unseal <key2>
vault operator unseal <key3>

# Login
vault login <root-token>
```

### Store Secrets

```bash
vault kv put secret/flyfleet/database \
  host="postgres.internal" \
  password="super-secret-password"

# Applications retrieve via Vault API or Vault Agent
```

### Use in FlyFleet

```go
// Read secret from Vault
secret, err := vault.Logical().Read("secret/data/flyfleet/database")
if err != nil {
    return err
}
password := secret.Data["password"].(string)
```

---

## 📝 Tracing with Jaeger

### Instrument Your Code

FlyFleet already has OpenTelemetry. Configure it to export to Jaeger:

```go
import "go.opentelemetry.io/otel"

func init() {
    otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
    if otlpEndpoint == "" {
        otlpEndpoint = "http://localhost:4317"
    }

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(otlptrace.New(
            otlptrace.WithEndpoint(otlpEndpoint),
            otlptrace.WithInsecure(), // Use TLS in prod
        )),
    )
    otel.SetTracerProvider(tp)
}
```

### View Traces

1. Go to http://localhost:16686
2. Select service: `flyfleet`
3. Click "Find Traces"
4. Click a trace to see spans

---

## 📋 Logging with Loki

### Ship Application Logs

FlyFleet uses zap logger with structured JSON. Loki collects via:

1. **Docker logging driver** (if running in Docker) - already configured
2. **Filelog receiver** (OTel Collector) - reads container logs
3. **Direct SDK** (optional) - use OpenTelemetry Log SDK

### Query Logs in Grafana

1. Open Grafana → Explore
2. Select **Loki** datasource
3. Query: `{service="fleet-core"} |= "ERROR"`
4. Logs appear with labels

---

## 🔗 Data Pipelines

### PostgreSQL → ClickHouse (Analytics Sync)

Use **Debezium** for real-time CDC (Change Data Capture):

```yaml
# infrastructure/debezium/connector.json
{
  "name": "flyfleet-connector",
  "config": {
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "database.hostname": "postgres",
    "database.port": "5432",
    "database.user": "debezium",
    "database.password": "dbz",
    "database.dbname": "fleet_db",
    "table.include.list": "public.invoices,public.payments",
    "topic.prefix": "flyfleet"
  }
}
```

Or use **Airbyte** (GUI-based ETL): http://localhost:8000

### NATS → OpenSearch (Event Indexing)

```go
// Subscribe to events and index in OpenSearch
nc, _ := nats.Connect(nats.DefaultURL)
js, _ := nc.JetStream()

_, err := js.Subscribe("fleet.events.>", func(msg *nats.Msg) {
    var event map[string]interface{}
    json.Unmarshal(msg.Data, &event)

    // Index in OpenSearch
    client.Index("events-"+time.Now().Format("2006-01-02"),
        strings.NewReader(string(msg.Data)))
})
```

---

## 🧪 Testing the Stack

### Verify All Services are Healthy

```bash
./infrastructure/setup-full-stack.sh
```

### Run a Sample Query

```bash
# 1. Create a test invoice via API
curl -X POST http://localhost:8081/api/v1/billing/invoices \
  -H "Authorization: Bearer $(TOKEN)" \
  -H "Content-Type: application/json" \
  -d '{
    "org_id": "org-uuid",
    "customer_name": "Test Corp",
    "items": [{"description": "Service", "quantity": 1, "unit_price": 1000}]
  }'

# 2. Check logs appear in Loki
curl 'http://localhost:3100/loki/api/v1/query?query={service="fleet-core"}&limit=10'

# 3. Check metrics in Prometheus
curl http://localhost:9090/metrics | grep http_requests

# 4. View trace in Jaeger (if instrumented)
# Go to http://localhost:16686

# 5. Search invoice in OpenSearch
curl -XGET http://localhost:9200/invoices-*/_search -H 'Content-Type: application/json' -d '{"query":{"match":{"customer_name":"Test"}}}'
```

---

## 📦 What's Running (Resource Usage)

| Service | RAM | Disk | Purpose |
|---------|-----|------|---------|
| PostgreSQL | 512MB | 10GB | Primary DB |
| Redis | 128MB | 1GB | Cache |
| NATS | 64MB | 1GB | Messaging |
| OpenSearch | 2GB | 10GB | Search |
| ClickHouse | 2GB | 10GB | Analytics |
| Prometheus | 512MB | 2GB | Metrics |
| Grafana | 256MB | 1GB | Dashboards |
| Loki | 512MB | 5GB | Logs |
| Jaeger | 512MB | 1GB | Traces |
| OTel Collector | 256MB | 0 | Telemetry |
| Vault | 128MB | 0 | Secrets |
| Fleet Core | 512MB | 0 | App |
| **Total** | **~7.5GB** | **~40GB** | - |

**Minimum specs**: 8GB RAM, 50GB disk
**Recommended**: 16GB RAM, 100GB SSD

---

## 🛠️ Maintenance

### Backup Databases

```bash
# Postgres (already automated via docker-compose)
docker exec fleet-db-backup /app/backup.sh

# ClickHouse
docker exec fleet-clickhouse clickhouse-client --query="BACKUP TABLE analytics.* TO 'backup/'"

# OpenSearch
curl -XPOST http://localhost:9200/_snapshot/my_backup/snapshot_$(date +%s) -H 'Content-Type: application/json'
```

### Update Services

```bash
# Pull latest images
docker-compose -f docker-compose.full.yml pull

# Recreate containers
docker-compose -f docker-compose.full.yml up -d --force-recreate

# Migrate ClickHouse schema if changed
docker exec -i fleet-clickhouse clickhouse-client < infrastructure/clickhouse/init.sql
```

### View Logs

```bash
# All services
docker-compose -f docker-compose.full.yml logs -f

# Specific service
docker-compose -f docker-compose.full.yml logs -f prometheus

# Grafana logs (most useful for debugging)
docker-compose -f docker-compose.full.yml logs grafana
```

### Restart a Service

```bash
docker-compose -f docker-compose.full.yml restart grafana
```

---

## 🐛 Troubleshooting

### Services not starting?

```bash
# Check logs
docker-compose -f docker-compose.full.yml logs [service-name]

# Check resource usage (maybe out of RAM?)
docker stats

# Restart everything
docker-compose -f docker-compose.full.yml down
docker-compose -f docker-compose.full.yml up -d
```

### Port already in use?

Check what's using the port and stop it:

```bash
lsof -i :9200  # Find process
kill -9 <PID>  # Kill it
```

Or change ports in `docker-compose.full.yml`.

### OpenSearch out of memory?

OpenSearch needs at least 2GB heap. Increase in docker-compose:

```yaml
environment:
  - "OPENSEARCH_JAVA_OPTS=-Xms2g -Xmx2g"
```

### Grafana can't connect to Prometheus?

Check Prometheus is running:

```bash
curl http://localhost:9090/-/healthy
```

If Prometheus down: `docker-compose -f docker-compose.full.yml restart prometheus`

### Vault not initialized?

Run the setup script again:

```bash
./infrastructure/setup-full-stack.sh
```

---

## 🌐 Production Deployment

### Move to Kubernetes

You have Helm charts! Use them:

```bash
# Create namespace
kubectl create namespace flyfleet

# Install stack (manual chart creation needed or use bitnami subcharts)
helm install prometheus prometheus-community/kube-prometheus-stack -n flyfleet
helm install loki grafana/loki-stack -n flyfleet

# Deploy FlyFleet
helm install flyfleet infrastructure/helm/flyfleet -n flyfleet
```

### Secure Communications

1. Enable TLS for all services (Vault, OpenSearch, Grafana)
2. Use Vault for secrets (not environment variables)
3. Enable OpenSearch security plugin (xpack)
4. Configure OTel Collector with mTLS
5. Enable audit logging

### High Availability

- Run 3+ replicas of each service
- Use external storage (AWS RDS, S3)
- Configure PostgreSQL read replicas
- Set up NATS clustering
- Deploy across multiple availability zones

---

## 📚 Additional Resources

| Tool | Documentation |
|------|---------------|
| OpenSearch | https://opensearch.org/docs/ |
| ClickHouse | https://clickhouse.com/docs/ |
| Grafana | https://grafana.com/docs/ |
| Prometheus | https://prometheus.io/docs/ |
| Loki | https://grafana.com/docs/loki/ |
| Jaeger | https://www.jaegertracing.io/docs/ |
| OpenTelemetry | https://opentelemetry.io/docs/ |
| Vault | https://developer.hashicorp.com/vault/docs |

---

## 💡 What's Next?

1. **Instrument FlyFleet** - Ensure OTel SDK is embedded in all services
2. **Set up alerts** - Configure Prometheus alerts (disk full, error rate spiking)
3. **Create dashboards** - Build business-specific dashboards in Grafana
4. **Data pipelines** - Set up CDC from Postgres to ClickHouse
5. **Search integration** - Index invoices, drivers, vehicles in OpenSearch
6. **Vault integration** - Move secrets from .env to Vault
7. **Backup strategy** - Automated backups for all data stores
8. **Security review** - Harden configurations for production

---

**Total software cost: $0** 🎉

All tools are 100% free and open source.
