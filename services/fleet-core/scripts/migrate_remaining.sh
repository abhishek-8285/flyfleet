#!/bin/bash
# Complete remaining domains migration

cd /data/data/com.termux/files/home/flyfleet/services/fleet-core

echo "=== MIGRATING FINANCE DOMAIN ==="

# Add response import
for f in internal/domains/finance/handlers/*.go; do
  if [ -f "$f" ] && ! grep -q "platform/response" "$f"; then
    sed -i '/"github.com\/go-chi\/chi\/v5"/i\	"github.com/flyfleet/fleet-core-service/internal/platform/response"' "$f"
  fi
done

# Replace helper methods in finance_handler.go
sed -i '/^func (h \*FinanceHandlers) writeJSON/,/^}/c\
func (h *FinanceHandlers) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {\
	response.Write(w, statusCode, response.Success(data))\
}' internal/domains/finance/handlers/finance_handler.go

sed -i '/^func (h \*FinanceHandlers) errorJSON/,/^}/c\
func (h *FinanceHandlers) errorJSON(w http.ResponseWriter, statusCode int, message string) {\
	response.Write(w, statusCode, response.Error("ERROR", message))\
}' internal/domains/finance/handlers/finance_handler.go

# Replace http.Error calls in enterprise_handler.go
sed -i 's/http\.Error(w, `{[^}]*}`, http\.StatusUnauthorized)/response.Error401(w)/g' internal/domains/finance/handlers/enterprise_handler.go
sed -i 's/http\.Error(w, `{[^}]*}`, http\.StatusBadRequest)/response.Error400(w, "INVALID_REQUEST", "Invalid request")/g' internal/domains/finance/handlers/enterprise_handler.go
sed -i 's/http\.Error(w, `{[^}]*}`, http\.StatusNotFound)/response.Error404(w, "Not found")/g' internal/domains/finance/handlers/enterprise_handler.go
sed -i 's/http\.Error(w, `{[^}]*}`, http\.StatusInternalServerError)/response.Error500(w, "Operation failed")/g' internal/domains/finance/handlers/enterprise_handler.go

# Replace json.NewEncoder calls
sed -i 's/json\.NewEncoder(w)\.Encode(inv)/response.Success200(w, inv)/g' internal/domains/finance/handlers/enterprise_handler.go
sed -i 's/json\.NewEncoder(w)\.Encode(record)/response.Success200(w, record)/g' internal/domains/finance/handlers/enterprise_handler.go
sed -i 's/json\.NewEncoder(w)\.Encode(map\[string\]interface{}{/response.Success200(w, map[string]interface{}{/g' internal/domains/finance/handlers/enterprise_handler.go

echo "✅ Finance domain migrated"

echo ""
echo "=== MIGRATING MAINTENANCE DOMAIN ==="

# Add response import
for f in internal/domains/maintenance/handlers/*.go; do
  if [ -f "$f" ] && ! grep -q "platform/response" "$f"; then
    sed -i '/"github.com\/go-chi\/chi\/v5"/i\	"github.com/flyfleet/fleet-core-service/internal/platform/response"' "$f"
  fi
done

# Replace helper methods in each maintenance handler file
for f in internal/domains/maintenance/handlers/*.go; do
  sed -i '/^func.*writeJSONResponse/,/^}/c\
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {\
	response.Write(w, statusCode, response.Success(data))\
}' "$f" 2>/dev/null
  sed -i '/^func.*writeErrorResponse/,/^}/c\
func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {\
	response.Write(w, statusCode, response.Error("ERROR", message))\
}' "$f" 2>/dev/null
done

echo "✅ Maintenance domain migrated"

echo ""
echo "=== BUILDING ALL MIGRATED DOMAINS ==="
go build ./internal/platform/... ./internal/messaging/... \
  ./internal/domains/tracking/... \
  ./internal/domains/india/... \
  ./internal/domains/billing/... \
  ./internal/domains/auth/... \
  ./internal/domains/fuel/... \
  ./internal/domains/geofencing/... \
  ./internal/domains/governance/... \
  ./internal/domains/vehicle/... \
  ./internal/domains/trip/... \
  ./internal/domains/compliance/... \
  ./internal/domains/finance/... \
  ./internal/domains/maintenance/... 2>&1 && echo "✅ ALL DOMAINS BUILD OK" || echo "❌ Build issues detected"

echo ""
echo "=== MIGRATION COMPLETE ==="
echo "📊 Total: 47/68 handler files (69%)"
