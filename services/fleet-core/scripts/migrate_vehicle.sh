#!/bin/bash
# Vehicle Domain Response Package Migration Script

cd /data/data/com.termux/files/home/flyfleet/services/fleet-core

echo "=== MIGRATING VEHICLE DOMAIN TO RESPONSE PACKAGE ==="

# Add response import to all handler files
for f in internal/domains/vehicle/handlers/*.go; do
  if [ -f "$f" ] && ! grep -q "platform/response" "$f"; then
    # Add import after chi import
    sed -i '/"github.com\/go-chi\/chi\/v5"/i\	"github.com/flyfleet/fleet-core-service/internal/platform/response"' "$f"
    echo "✅ Added import: $f"
  fi
done

# Update helper methods in each file
echo ""
echo "Updating helper methods..."

# vehicle_handlers.go
sed -i '/^func (h \*VehicleHandlers) writeJSONResponse/,/^}/c\
func (h *VehicleHandlers) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {\
	response.Write(w, statusCode, response.Success(data))\
}' internal/domains/vehicle/handlers/vehicle_handlers.go

sed -i '/^func (h \*VehicleHandlers) writeErrorResponse/,/^}/c\
func (h *VehicleHandlers) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {\
	response.Write(w, statusCode, response.Error("ERROR", message))\
}' internal/domains/vehicle/handlers/vehicle_handlers.go

# bulk_handlers.go
sed -i '/^func (h \*BulkHandlers) writeJSONResponse/,/^}/c\
func (h *BulkHandlers) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {\
	response.Write(w, statusCode, response.Success(data))\
}' internal/domains/vehicle/handlers/bulk_handlers.go

# compliance_handlers.go
sed -i '/^func (h \*ComplianceHandlers) writeJSON/,/^}/c\
func (h *ComplianceHandlers) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {\
	response.Write(w, statusCode, response.Success(data))\
}' internal/domains/vehicle/handlers/compliance_handlers.go

# documents_handlers.go
sed -i '/^func (h \*DocumentHandlers) writeJSON/,/^}/c\
func (h *DocumentHandlers) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {\
	response.Write(w, statusCode, response.Success(data))\
}' internal/domains/vehicle/handlers/documents_handlers.go

# lifecycle_handlers.go
sed -i '/^func (h \*LifecycleHandlers) writeJSON/,/^}/c\
func (h *LifecycleHandlers) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {\
	response.Write(w, statusCode, response.Success(data))\
}' internal/domains/vehicle/handlers/lifecycle_handlers.go

# trip_handlers.go
sed -i '/^func (h \*TripHandlers) writeJSON/,/^}/c\
func (h *TripHandlers) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {\
	response.Write(w, statusCode, response.Success(data))\
}' internal/domains/vehicle/handlers/trip_handlers.go

echo "✅ Helper methods updated"
echo ""
echo "Building vehicle domain..."
go build ./internal/domains/vehicle/... 2>&1 && echo "✅ VEHICLE DOMAIN BUILDS OK" || echo "❌ Build failed"
