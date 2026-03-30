#!/bin/bash

# Health check for all FlyFleet infrastructure services

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

services=(
    "PostgreSQL:5433:/health"
    "Redis:6380:PING"
    "NATS:8222:/varz"
    "OpenSearch:9200:/"
    "OpenSearch Dashboards:5601:/"
    "ClickHouse:8123/:/"
    "Prometheus:9090:/-/healthy"
    "Grafana:3000:/api/health"
    "Loki:3100:/ready"
    "Jaeger:16686:/"
    "Vault:8200:/v1/sys/health"
    "Fleet Core:8081:/health"
)

echo "=========================================="
echo "FlyFleet Infrastructure Health Check"
echo "=========================================="
echo ""

passed=0
failed=0

for service in "${services[@]}"; do
    IFS=':' read -r name port endpoint <<< "$service"
    url="http://localhost:$port$endpoint"

    if curl -s -o /dev/null -w "%{http_code}" "$url" | grep -q "^2"; then
        echo -e "[${GREEN}✓${NC}] $name ($port)"
        ((passed++))
    else
        echo -e "[${RED}✗${NC}] $name ($port) - not responding"
        ((failed++))
    fi
done

echo ""
echo "=========================================="
echo "Results: ${GREEN}$passed passed${NC}, ${RED}$failed failed${NC}"
echo "=========================================="

if [ $failed -eq 0 ]; then
    echo ""
    echo "🎉 All services are healthy!"
    echo ""
    echo "Access points:"
    echo "  • FlyFleet API: http://localhost:8081"
    echo "  • Grafana: http://localhost:3000"
    echo "  • Jaeger: http://localhost:16686"
    echo "  • OpenSearch: http://localhost:9200"
    echo "  • ClickHouse: http://localhost:8123"
    echo "  • Vault: http://localhost:8200"
    exit 0
else
    echo ""
    echo "⚠️  Some services are not healthy. Check logs:"
    echo "  docker-compose -f docker-compose.full.yml logs -f [service]"
    exit 1
fi
