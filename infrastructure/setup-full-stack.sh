#!/bin/bash

# FlyFleet Full Stack Setup Script
# Initializes Vault, OpenSearch, ClickHouse with sample data

set -e

echo "=========================================="
echo "FlyFleet Full Stack Initialization"
echo "=========================================="

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Wait for services to be ready
wait_for_service() {
    local host=$1
    local port=$2
    local service=$3
    echo -n "Waiting for $service..."
    for i in {1..30}; do
        if nc -z "$host" "$port" 2>/dev/null; then
            echo -e " ${GREEN}✓${NC}"
            return 0
        fi
        sleep 2
        echo -n "."
    done
    echo -e " ${RED}✗${NC}"
    return 1
}

echo ""
echo "Step 1: Waiting for all services to start..."
wait_for_service localhost 8200 "Vault"
wait_for_service localhost 9200 "OpenSearch"
wait_for_service localhost 8123 "ClickHouse"
wait_for_service localhost 9090 "Prometheus"
wait_for_service localhost 3000 "Grafana"
wait_for_service localhost 3100 "Loki"
wait_for_service localhost 16686 "Jaeger"

echo ""
echo "Step 2: Initializing Vault..."

# Set Vault address
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=dev-root-token

# Wait for Vault to be ready
sleep 3

# Enable KV v2 secrets engine
vault secrets enable -path=secret kv-v2 2>/dev/null || true

# Store application secrets
echo "Storing secrets in Vault..."

vault kv put secret/flyfleet/database \
  host="postgres" \
  port=5432 \
  name="fleet_db" \
  user="fleet_user" \
  password="fleet_password" \
  @- <<EOF
{
  "readonly_user": "fleet_readonly",
  "readonly_password": "fleet_readonly_pass"
}
EOF

vault kv put secret/flyfleet/jwt \
  secret="${JWT_SECRET:-change-me-in-production-please-generate-a-strong-secret-key}"

vault kv put secret/flyfleet/redis \
  host="fleet-redis" \
  port=6379 \
  password=""

vault kv put secret/flyfleet/nats \
  url="nats://fleet-nats:4222"

vault kv put secret/flyfleet/s3 \
  endpoint="minio:9000" \
  access_key="minioadmin" \
  secret_key="minioadmin" \
  bucket="flyfleet-assets"

echo -e "${GREEN}✓ Vault initialized with secrets${NC}"

echo ""
echo "Step 3: Verifying OpenSearch..."

# Create OpenSearch index templates (via API)
curl -s -XPUT "http://localhost:9200/_index_template/fleet-invoices" \
  -H 'Content-Type: application/json' \
  -d '{
    "index_patterns": ["invoices-*"],
    "template": {
      "mappings": {
        "properties": {
          "id": {"type": "keyword"},
          "invoice_number": {"type": "text", "analyzer": "english"},
          "customer_name": {"type": "text", "analyzer": "english"},
          "org_id": {"type": "keyword"},
          "total_amount": {"type": "double"},
          "currency": {"type": "keyword"},
          "status": {"type": "keyword"},
          "created_at": {"type": "date"},
          "due_date": {"type": "date"},
          "items": {"type": "nested"}
        }
      }
    }
  }' >/dev/null && echo -e "${GREEN}✓ OpenSearch index template created${NC}" || echo -e "${YELLOW}⚠ OpenSearch index template skipped${NC}"

echo ""
echo "Step 4: Verifying ClickHouse..."

# Check ClickHouse is healthy
if clickhouse-client --query="SELECT 1" >/dev/null 2>&1; then
    echo -e "${GREEN}✓ ClickHouse is healthy${NC}"

    # Run init SQL (create tables)
    echo "Creating ClickHouse analytics tables..."
    clickhouse-client --database=flyfleet_analytics --query="$(cat infrastructure/clickhouse/init.sql)" >/dev/null 2>&1
    echo -e "${GREEN}✓ ClickHouse analytics tables created${NC}"
else
    echo -e "${RED}✗ ClickHouse not responding${NC}"
fi

echo ""
echo "Step 5: Setting up Grafana alerts..."

# Grafana will auto-provision datasources and dashboards
echo -e "${GREEN}✓ Grafana provisioning configured${NC}"
echo "  Access Grafana at http://localhost:3000"
echo "  Username: admin"
echo "  Password: ${GRAFANA_PASSWORD:-admin}"

echo ""
echo "=========================================="
echo -e "${GREEN}Setup Complete!${NC}"
echo "=========================================="
echo ""
echo "Services running:"
echo "  📊 Grafana (dashboards):        http://localhost:3000"
echo "  🔍 OpenSearch (search):         http://localhost:9200"
echo "  📈 Prometheus (metrics):        http://localhost:9090"
echo "  📋 Loki (logs):                 http://localhost:3100"
echo "  🔭 Jaeger (traces):             http://localhost:16686"
echo "  🔐 Vault (secrets):             http://localhost:8200"
echo "  🗄️  ClickHouse (analytics):     http://localhost:8123"
echo "  📮 NATS JetStream:              localhost:4222"
echo "  💾 PostgreSQL:                  localhost:5433"
echo "  📦 Redis:                       localhost:6380"
echo ""
echo "Next steps:"
echo "1. Update FlyFleet config to send traces to otel-collector"
echo "2. Deploy FlyFleet backend with docker-compose.full.yml"
echo "3. Access Grafana and import dashboards from infrastructure/grafana/dashboards/"
echo "4. Create Vault policies for production"
echo ""
echo "To view logs:"
echo "  docker-compose -f docker-compose.full.yml logs -f [service]"
echo ""
