#!/bin/bash
# Complete maintenance domain migration

cd /data/data/com.termux/files/home/flyfleet/services/fleet-core

echo "=== MIGRATING MAINTENANCE DOMAIN ==="

# Add response import to all files
for f in internal/domains/maintenance/handlers/*.go; do
  if [ -f "$f" ] && ! grep -q "platform/response" "$f"; then
    sed -i '/"github.com\/go-chi\/chi\/v5"/i\	"github.com/flyfleet/fleet-core/internal/platform/response"' "$f"
  fi
done

# Update analytics_handler.go
sed -i '/^func (h \*AnalyticsHandler) writeJSONResponse/,/^}/c\
func (h *AnalyticsHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {\
	response.Write(w, statusCode, response.Success(data))\
}' internal/domains/maintenance/handlers/analytics_handler.go

sed -i '/^func (h *AnalyticsHandler) writeErrorResponse/,/^}/c\
func (h *AnalyticsHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {\
	response.Write(w, statusCode, response.Error("ERROR", message))\
}' internal/domains/maintenance/handlers/analytics_handler.go

# Update inspection_handler.go
sed -i '/^func (h \*InspectionHandler) writeJSONResponse/,/^}/c\
func (h *InspectionHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {\
	response.Write(w, statusCode, response.Success(data))\
}' internal/domains/maintenance/handlers/inspection_handler.go

sed -i '/^func (h \*InspectionHandler) writeErrorResponse/,/^}/c\
func (h *InspectionHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {\
	response.Write(w, statusCode, response.Error("ERROR", message))\
}' internal/domains/maintenance/handlers/inspection_handler.go

# Update inventory_handler.go
sed -i '/^func (h \*InventoryHandler) writeJSONResponse/,/^}/c\
func (h *InventoryHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {\
	response.Write(w, statusCode, response.Success(data))\
}' internal/domains/maintenance/handlers/inventory_handler.go

sed -i '/^func (h \*InventoryHandler) writeErrorResponse/,/^}/c\
func (h *InventoryHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {\
	response.Write(w, statusCode, response.Error("ERROR", message))\
}' internal/domains/maintenance/handlers/inventory_handler.go

# Update maintenance_handler.go
sed -i '/^func (h \*MaintenanceHandler) writeJSONResponse/,/^}/c\
func (h *MaintenanceHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {\
	response.Write(w, statusCode, response.Success(data))\
}' internal/domains/maintenance/handlers/maintenance_handler.go

sed -i '/^func (h \*MaintenanceHandler) writeErrorResponse/,/^}/c\
func (h *MaintenanceHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {\
	response.Write(w, statusCode, response.Error("ERROR", message))\
}' internal/domains/maintenance/handlers/maintenance_handler.go

# Update technician_handler.go
sed -i '/^func (h \*TechnicianHandler) writeJSONResponse/,/^}/c\
func (h *TechnicianHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {\
	response.Write(w, statusCode, response.Success(data))\
}' internal/domains/maintenance/handlers/technician_handler.go

sed -i '/^func (h \*TechnicianHandler) writeErrorResponse/,/^}/c\
func (h *TechnicianHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {\
	response.Write(w, statusCode, response.Error("ERROR", message))\
}' internal/domains/maintenance/handlers/technician_handler.go

# Update vendor_handler.go
sed -i '/^func (h \*VendorHandler) writeJSONResponse/,/^}/c\
func (h *VendorHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {\
	response.Write(w, statusCode, response.Success(data))\
}' internal/domains/maintenance/handlers/vendor_handler.go

sed -i '/^func (h \*VendorHandler) writeErrorResponse/,/^}/c\
func (h *VendorHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {\
	response.Write(w, statusCode, response.Error("ERROR", message))\
}' internal/domains/maintenance/handlers/vendor_handler.go

# Update work_order_handler.go
sed -i '/^func (h \*WorkOrderHandler) writeJSONResponse/,/^}/c\
func (h *WorkOrderHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {\
	response.Write(w, statusCode, response.Success(data))\
}' internal/domains/maintenance/handlers/work_order_handler.go

sed -i '/^func (h \*WorkOrderHandler) writeErrorResponse/,/^}/c\
func (h *WorkOrderHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {\
	response.Write(w, statusCode, response.Error("ERROR", message))\
}' internal/domains/maintenance/handlers/work_order_handler.go

echo "✅ Maintenance domain migrated"
go build ./internal/domains/maintenance/... 2>&1 && echo "✅ MAINTENANCE DOMAIN BUILDS OK" || echo "❌ Build issues"
