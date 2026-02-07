package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/auth-service/internal/middleware"
	"github.com/flyfleet/auth-service/internal/service"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AdminHandlers struct {
	service *service.AuthService
	logger  *zap.Logger
}

func NewAdminHandlers(s *service.AuthService, logger *zap.Logger) *AdminHandlers {
	return &AdminHandlers{
		service: s,
		logger:  logger,
	}
}

// GET /api/v1/admin/organizations
func (h *AdminHandlers) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	pageCtx := r.Context().Value("pagination").(map[string]int)
	page, size := pageCtx["page"], pageCtx["size"]

	// In a real scenario, check if the requester is a Super Admin here

	orgs, total, err := h.service.ListOrganizations(r.Context(), page, size)
	if err != nil {
		h.logger.Error("Failed to list organizations", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	middleware.WriteSuccess(w, orgs, page, size, total)
}

// DELETE /api/v1/admin/organizations/{id}
func (h *AdminHandlers) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id") // Or extraction from path params if router supports it
	if idStr == "" {
		// Fallback for simple query param usage for now
		http.Error(w, `{"error": "Missing ID"}`, http.StatusBadRequest)
		return
	}
	
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.SoftDeleteOrganization(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete organization", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	middleware.WriteSuccess(w, nil, 0, 0, 0)
}

// GET /api/v1/admin/users
func (h *AdminHandlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	pageCtx := r.Context().Value("pagination").(map[string]int)
	page, size := pageCtx["page"], pageCtx["size"]
	
	orgID, ok := r.Context().Value("org_id").(uuid.UUID)
	if !ok {
		// Should be caught by middleware, but safe check
		http.Error(w, `{"error": "Organization Context Missing"}`, http.StatusForbidden)
		return
	}

	users, total, err := h.service.ListUsers(r.Context(), orgID, page, size)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	middleware.WriteSuccess(w, users, page, size, total)
}

// DELETE /api/v1/admin/users/{id}
func (h *AdminHandlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
		return
	}

	// Verify user belongs to org? Omitted for brevity, but crucial in real app

	if err := h.service.SoftDeleteUser(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete user", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	middleware.WriteSuccess(w, nil, 0, 0, 0)
}

// GET /api/v1/admin/vehicles
func (h *AdminHandlers) ListVehicles(w http.ResponseWriter, r *http.Request) {
	pageCtx := r.Context().Value("pagination").(map[string]int)
	page, size := pageCtx["page"], pageCtx["size"]
	orgID := r.Context().Value("org_id").(uuid.UUID)

	vehicles, total, err := h.service.ListVehicles(r.Context(), orgID, page, size)
	if err != nil {
		h.logger.Error("Failed to list vehicles", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	middleware.WriteSuccess(w, vehicles, page, size, total)
}

// POST /api/v1/admin/vehicles
func (h *AdminHandlers) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	orgID := r.Context().Value("org_id").(uuid.UUID)
	
	var req struct {
		Name         string `json:"name"`
		LicensePlate string `json:"license_plate"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid Body"}`, http.StatusBadRequest)
		return
	}

	vehicle, err := h.service.CreateVehicle(r.Context(), req.Name, req.LicensePlate, orgID)
	if err != nil {
		h.logger.Error("Failed to create vehicle", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	middleware.WriteSuccess(w, vehicle, 1, 1, 1)
}

// DELETE /api/v1/admin/vehicles
func (h *AdminHandlers) DeleteVehicle(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.SoftDeleteVehicle(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete vehicle", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	middleware.WriteSuccess(w, nil, 0, 0, 0)
}

// Drivers Handlers
func (h *AdminHandlers) ListDrivers(w http.ResponseWriter, r *http.Request) {
	pageCtx := r.Context().Value("pagination").(map[string]int)
	orgID := r.Context().Value("org_id").(uuid.UUID)
	drivers, total, err := h.service.ListDrivers(r.Context(), orgID, pageCtx["page"], pageCtx["size"])
	if err != nil { 
		h.logger.Error("Failed to list drivers", zap.Error(err))
		http.Error(w, `{"error": "Internal Error"}`, 500)
		return 
	}
	middleware.WriteSuccess(w, drivers, pageCtx["page"], pageCtx["size"], total)
}

func (h *AdminHandlers) CreateDriver(w http.ResponseWriter, r *http.Request) {
	orgID := r.Context().Value("org_id").(uuid.UUID)
	var req struct { Name string `json:"name"`; Phone string `json:"phone"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { 
		http.Error(w, `{"error":"Invalid Body"}`, 400)
		return 
	}
	driver, err := h.service.CreateDriver(r.Context(), req.Name, req.Phone, orgID)
	if err != nil { 
		h.logger.Error("Failed to create driver", zap.Error(err))
		http.Error(w, `{"error":"Internal Error"}`, 500)
		return 
	}
	middleware.WriteSuccess(w, driver, 1, 1, 1)
}

func (h *AdminHandlers) DeleteDriver(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil { http.Error(w, `{"error":"Invalid ID"}`, 400); return }
	
	if err := h.service.SoftDeleteDriver(r.Context(), id); err != nil { 
		h.logger.Error("Failed to delete driver", zap.Error(err))
		http.Error(w, `{"error":"Error"}`, 500)
		return 
	}
	middleware.WriteSuccess(w, nil, 0, 0, 0)
}

// Devices Handlers
func (h *AdminHandlers) ListDevices(w http.ResponseWriter, r *http.Request) {
	pageCtx := r.Context().Value("pagination").(map[string]int)
	orgID := r.Context().Value("org_id").(uuid.UUID)
	devices, total, err := h.service.ListDevices(r.Context(), orgID, pageCtx["page"], pageCtx["size"])
	if err != nil { 
		h.logger.Error("Failed to list devices", zap.Error(err))
		http.Error(w, `{"error": "Internal Error"}`, 500)
		return 
	}
	middleware.WriteSuccess(w, devices, pageCtx["page"], pageCtx["size"], total)
}

func (h *AdminHandlers) CreateDevice(w http.ResponseWriter, r *http.Request) {
	orgID := r.Context().Value("org_id").(uuid.UUID)
	var req struct { Serial string `json:"serial_number"`; Type string `json:"type"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { 
		http.Error(w, `{"error":"Invalid Body"}`, 400)
		return 
	}
	device, err := h.service.CreateDevice(r.Context(), req.Serial, req.Type, orgID)
	if err != nil { 
		h.logger.Error("Failed to create device", zap.Error(err))
		http.Error(w, `{"error":"Internal Error"}`, 500)
		return 
	}
	middleware.WriteSuccess(w, device, 1, 1, 1)
}

func (h *AdminHandlers) DeleteDevice(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil { http.Error(w, `{"error":"Invalid ID"}`, 400); return }

	if err := h.service.SoftDeleteDevice(r.Context(), id); err != nil { 
		h.logger.Error("Failed to delete device", zap.Error(err))
		http.Error(w, `{"error":"Error"}`, 500)
		return 
	}
	middleware.WriteSuccess(w, nil, 0, 0, 0)
}
// Trips Handlers
func (h *AdminHandlers) ListTrips(w http.ResponseWriter, r *http.Request) {
	pageCtx := r.Context().Value("pagination").(map[string]int)
	orgID := r.Context().Value("org_id").(uuid.UUID)
	trips, total, err := h.service.ListTrips(r.Context(), orgID, pageCtx["page"], pageCtx["size"])
	if err != nil { 
		h.logger.Error("Failed to list trips", zap.Error(err))
		http.Error(w, `{"error": "Internal Error"}`, 500)
		return 
	}
	middleware.WriteSuccess(w, trips, pageCtx["page"], pageCtx["size"], total)
}

func (h *AdminHandlers) CreateTrip(w http.ResponseWriter, r *http.Request) {
	// TODO: Parse trip body properly
	http.Error(w, "Not implemented in gateway yet - direct service call required", http.StatusNotImplemented)
}

// Alerts Handlers
func (h *AdminHandlers) ListAlerts(w http.ResponseWriter, r *http.Request) {
	pageCtx := r.Context().Value("pagination").(map[string]int)
	orgID := r.Context().Value("org_id").(uuid.UUID)
	alerts, total, err := h.service.ListAlerts(r.Context(), orgID, pageCtx["page"], pageCtx["size"])
	if err != nil { 
		h.logger.Error("Failed to list alerts", zap.Error(err))
		http.Error(w, `{"error": "Internal Error"}`, 500)
		return 
	}
	middleware.WriteSuccess(w, alerts, pageCtx["page"], pageCtx["size"], total)
}

func (h *AdminHandlers) GetAlertSummary(w http.ResponseWriter, r *http.Request) {
	orgID := r.Context().Value("org_id").(uuid.UUID)
	summary, err := h.service.GetAlertSummary(r.Context(), orgID)
	if err != nil { 
		h.logger.Error("Failed to get alert summary", zap.Error(err))
		http.Error(w, `{"error": "Internal Error"}`, 500)
		return 
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// Analytics Handlers
func (h *AdminHandlers) GetDriverStats(w http.ResponseWriter, r *http.Request) {
	orgID := r.Context().Value("org_id").(uuid.UUID)
	stats, err := h.service.GetDriverStats(r.Context(), orgID)
	if err != nil { 
		h.logger.Error("Failed to get driver stats", zap.Error(err))
		http.Error(w, `{"error": "Internal Error"}`, 500)
		return 
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *AdminHandlers) GetTripStats(w http.ResponseWriter, r *http.Request) {
	orgID := r.Context().Value("org_id").(uuid.UUID)
	stats, err := h.service.GetTripStats(r.Context(), orgID)
	if err != nil { 
		h.logger.Error("Failed to get trip stats", zap.Error(err))
		http.Error(w, `{"error": "Internal Error"}`, 500)
		return 
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
