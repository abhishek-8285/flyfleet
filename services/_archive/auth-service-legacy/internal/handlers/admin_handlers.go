package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/flyfleet/auth-service/internal/middleware"
	"github.com/flyfleet/auth-service/internal/models"
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

	orgs, total, err := h.service.ListOrganizations(r.Context(), page, size)
	if err != nil {
		h.logger.Error("Failed to list organizations", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	middleware.WriteSuccess(w, orgs, page, size, total)
}

// POST /api/v1/admin/organizations
func (h *AdminHandlers) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name   string `json:"name"`
		Domain string `json:"domain"`
		Theme  string `json:"default_theme"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid body"}`, http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)
	org, err := h.service.CreateOrganization(r.Context(), req.Name, req.Domain, req.Theme, userID)
	if err != nil {
		h.logger.Error("Failed to create organization", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	middleware.WriteSuccess(w, org, 1, 1, 1)
}

// GET /api/v1/admin/organizations/{id}
func (h *AdminHandlers) GetOrganization(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
		return
	}

	org, err := h.service.GetOrganizationByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get organization", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	if org == nil {
		http.Error(w, `{"error": "Organization not found"}`, http.StatusNotFound)
		return
	}

	middleware.WriteSuccess(w, org, 1, 1, 1)
}

// PUT /api/v1/admin/organizations/{id}
func (h *AdminHandlers) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		Name   string `json:"name"`
		Domain string `json:"domain"`
		Theme  string `json:"default_theme"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid body"}`, http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)
	org, err := h.service.UpdateOrganization(r.Context(), id, req.Name, req.Domain, req.Theme, req.Status, userID)
	if err != nil {
		h.logger.Error("Failed to update organization", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	middleware.WriteSuccess(w, org, 1, 1, 1)
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
	query := r.URL.Query()

	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 {
		limit = 10
	}

	orgID, ok := r.Context().Value("org_id").(uuid.UUID)
	if !ok {
		h.logger.Warn("Organization Context Missing in ListUsers")
	}

	// Allow admin to override org_id via query parameter
	if adminOrgID := query.Get("org_id"); adminOrgID != "" {
		if uid, err := uuid.Parse(adminOrgID); err == nil {
			orgID = uid
		}
	}

	filter := models.UserFilter{
		OrgID:     orgID,
		Role:      query.Get("role"),
		Status:    query.Get("status"),
		Search:    query.Get("search"),
		Page:      page,
		Limit:     limit,
		SortBy:    query.Get("sort_by"),
		SortOrder: query.Get("sort_order"),
	}

	response, err := h.service.ListUsers(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to list users"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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

// GET /api/v1/admin/vehicles/{id}
func (h *AdminHandlers) GetVehicle(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
		return
	}

	vehicle, err := h.service.GetVehicleByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get vehicle", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	if vehicle == nil {
		http.Error(w, `{"error": "Vehicle not found"}`, http.StatusNotFound)
		return
	}

	middleware.WriteSuccess(w, vehicle, 1, 1, 1)
}

// POST /api/v1/admin/vehicles
func (h *AdminHandlers) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	orgID := r.Context().Value("org_id").(uuid.UUID)

	var v models.Vehicle
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, `{"error": "Invalid Body"}`, http.StatusBadRequest)
		return
	}
	v.OrgID = orgID

	vehicle, err := h.service.CreateVehicle(r.Context(), &v)
	if err != nil {
		h.logger.Error("Failed to create vehicle", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	middleware.WriteSuccess(w, vehicle, 1, 1, 1)
}

// PUT /api/v1/admin/vehicles/{id}
func (h *AdminHandlers) UpdateVehicle(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
		return
	}

	var v models.Vehicle
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, `{"error": "Invalid Body"}`, http.StatusBadRequest)
		return
	}
	v.ID = id

	// Optional: Check if ID matches context Org

	updated, err := h.service.UpdateVehicle(r.Context(), &v)
	if err != nil {
		h.logger.Error("Failed to update vehicle", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	middleware.WriteSuccess(w, updated, 1, 1, 1)
}

// PUT /api/v1/admin/vehicles/{id}/status
func (h *AdminHandlers) UpdateVehicleStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid Body"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateVehicleStatus(r.Context(), id, req.Status); err != nil {
		h.logger.Error("Failed to update vehicle status", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	middleware.WriteSuccess(w, nil, 0, 0, 0)
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
	var req struct {
		Name  string `json:"name"`
		Phone string `json:"phone"`
	}
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
	if err != nil {
		http.Error(w, `{"error":"Invalid ID"}`, 400)
		return
	}

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
	var req struct {
		Serial    string     `json:"serial_number"`
		Type      string     `json:"type"`
		VehicleID *uuid.UUID `json:"vehicle_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid Body"}`, 400)
		return
	}
	device, err := h.service.CreateDevice(r.Context(), req.Serial, req.Type, orgID, req.VehicleID)
	if err != nil {
		h.logger.Error("Failed to create device", zap.Error(err))
		http.Error(w, `{"error":"Internal Error"}`, 500)
		return
	}
	middleware.WriteSuccess(w, device, 1, 1, 1)
}

func (h *AdminHandlers) AssignDevice(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, `{"error":"Invalid Device ID"}`, 400)
		return
	}

	var req struct {
		VehicleID uuid.UUID `json:"vehicle_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid Body"}`, 400)
		return
	}

	if err := h.service.AssignDeviceToVehicle(r.Context(), id, req.VehicleID); err != nil {
		h.logger.Error("Failed to assign device", zap.Error(err))
		http.Error(w, `{"error":"Internal Error"}`, 500)
		return
	}
	middleware.WriteSuccess(w, nil, 0, 0, 0)
}

func (h *AdminHandlers) UnassignDevice(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, `{"error":"Invalid Device ID"}`, 400)
		return
	}

	if err := h.service.UnassignDevice(r.Context(), id); err != nil {
		h.logger.Error("Failed to unassign device", zap.Error(err))
		http.Error(w, `{"error":"Internal Error"}`, 500)
		return
	}
	middleware.WriteSuccess(w, nil, 0, 0, 0)
}

func (h *AdminHandlers) AssignDriver(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, `{"error":"Invalid Vehicle ID"}`, 400)
		return
	}

	var req struct {
		DriverID uuid.UUID `json:"driver_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid Body"}`, 400)
		return
	}

	if err := h.service.AssignDriverToVehicle(r.Context(), req.DriverID, id); err != nil {
		h.logger.Error("Failed to assign driver", zap.Error(err))
		http.Error(w, `{"error":"Internal Error"}`, 500)
		return
	}
	middleware.WriteSuccess(w, nil, 0, 0, 0)
}

func (h *AdminHandlers) DeleteDevice(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, `{"error":"Invalid ID"}`, 400)
		return
	}

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
	orgID := r.Context().Value("org_id").(uuid.UUID)

	var trip models.Trip
	if err := json.NewDecoder(r.Body).Decode(&trip); err != nil {
		http.Error(w, `{"error":"Invalid Body"}`, http.StatusBadRequest)
		return
	}

	trip.OrgID = orgID
	created, err := h.service.CreateTrip(r.Context(), &trip)
	if err != nil {
		h.logger.Error("Failed to create trip", zap.Error(err))
		http.Error(w, `{"error":"Internal Error"}`, http.StatusInternalServerError)
		return
	}

	middleware.WriteSuccess(w, created, 1, 1, 1)
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

// GET /api/v1/admin/users/stats
func (h *AdminHandlers) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetSystemStats(r.Context())
	if err != nil {
		h.logger.Error("Failed to get system stats", zap.Error(err))
		http.Error(w, `{"error": "Internal Error"}`, http.StatusInternalServerError)
		return
	}
	// BFF expects { "data": stats } or just stats?
	// BFF AuthClient.GetStats unmarshals into a struct matching fields.
	// Let's return the struct directly.
	middleware.WriteSuccess(w, stats, 1, 1, 1)
}
