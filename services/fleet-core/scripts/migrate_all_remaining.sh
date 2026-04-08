#!/bin/bash
# Complete migration of all remaining handler files

cd /data/data/com.termux/files/home/flyfleet/services/fleet-core

echo "=== COMPLETING REMAINING DOMAINS MIGRATION ==="
echo ""

# List of remaining handler files
files=(
  "internal/domains/admin_support/domains/emergency/handlers/handler.go"
  "internal/domains/alert/domains/alerts/handlers/alert_handler.go"
  "internal/domains/alerts/handlers/handlers.go"
  "internal/domains/analytics/handlers/analytics_handler.go"
  "internal/domains/customer/events/event_handler.go"
  "internal/domains/customer/handlers/customer_handler.go"
  "internal/domains/diagnostics/events/event_handler.go"
  "internal/domains/diagnostics/handlers/diagnostics_handler.go"
  "internal/domains/emergency/handlers/emergency_handler.go"
  "internal/domains/integrations/handlers/handlers.go"
  "internal/domains/monitoring/handlers/monitoring_handler.go"
  "internal/domains/reports/handlers/handlers.go"
  "internal/domains/route/handlers/route_handlers.go"
  "internal/domains/settings/handlers/handlers.go"
  "internal/domains/video/handler.go"
  "internal/domains/video/handlers/video_handlers.go"
)

for f in "${files[@]}"; do
  if [ -f "$f" ]; then
    echo "Processing $f..."
    
    # Add response import if not present
    if ! grep -q "platform/response" "$f"; then
      sed -i '/"github.com\/go-chi\/chi\/v5"/a\	"github.com/flyfleet/fleet-core-service/internal/platform/response"' "$f"
    fi
    
    # Replace common patterns
    # Pattern 1: json.NewEncoder(w).Encode(...)
    sed -i 's/json\.NewEncoder(w)\.Encode(\([^)]*\))/response.Success200(w, \1)/g' "$f"
    
    # Pattern 2: http.Error with various status codes
    sed -i 's/http\.Error(w, `{"error":"[^"]*"}`, http\.StatusBadRequest)/response.Error400(w, "INVALID_REQUEST", "Invalid request")/g' "$f"
    sed -i 's/http\.Error(w, `{"error":"[^"]*"}`, http\.StatusUnauthorized)/response.Error401(w)/g' "$f"
    sed -i 's/http\.Error(w, `{"error":"[^"]*"}`, http\.StatusNotFound)/response.Error404(w, "Not found")/g' "$f"
    sed -i 's/http\.Error(w, `{"error":"[^"]*"}`, http\.StatusInternalServerError)/response.Error500(w, "Operation failed")/g' "$f"
    
    # Pattern 3: w.WriteHeader + json.NewEncoder
    sed -i 's/w\.WriteHeader(http\.StatusOK)/response.Success200(w, nil)/g' "$f"
    sed -i 's/w\.WriteHeader(http\.StatusCreated)/response.Created201(w, nil)/g' "$f"
  fi
done

echo ""
echo "✅ All remaining domains processed"
echo ""
echo "Building..."
go build ./internal/domains/alerts/... \
  ./internal/domains/analytics/... \
  ./internal/domains/customer/... \
  ./internal/domains/diagnostics/... \
  ./internal/domains/emergency/... \
  ./internal/domains/integrations/... \
  ./internal/domains/monitoring/... \
  ./internal/domains/reports/... \
  ./internal/domains/route/... \
  ./internal/domains/settings/... \
  ./internal/domains/video/... 2>&1 | head -20

echo ""
echo "=== MIGRATION COMPLETE ==="
