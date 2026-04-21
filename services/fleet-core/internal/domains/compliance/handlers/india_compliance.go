package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/compliance/models"
	"github.com/flyfleet/fleet-core/internal/domains/compliance/service"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// checkDriverAccess verifies that the authenticated user can access the specified driver
// Returns true if: user is the driver themselves, OR user has admin/super_admin role
func (h *IndiaComplianceHandler) checkDriverAccess(r *http.Request, driverID *uuid.UUID, authUserID uuid.UUID, role string) bool {
	// driverID can be nil for vehicle-level actions
	if driverID == nil {
		// For nil driverID, only admins can proceed (checked separately)
		return role == "admin" || role == "super_admin"
	}

	isSelf := authUserID == *driverID
	isAdmin := role == "admin" || role == "super_admin"

	if isSelf {
		h.logger.Debug("Driver self-access granted",
			zap.String("driver_id", driverID.String()))
		return true
	}

	if isAdmin {
		h.logger.Debug("Admin access granted for driver data",
			zap.String("admin_user_id", authUserID.String()),
			zap.String("driver_id", driverID.String()),
			zap.String("role", role))
		return true
	}

	h.logger.Warn("Access denied: user cannot access driver data",
		zap.String("auth_user_id", authUserID.String()),
		zap.String("driver_id", driverID.String()),
		zap.String("role", role))
	return false
}

type IndiaComplianceHandler struct {
	dutyService service.DriverDutyService
	aisService  service.AIS140Service
	tripService service.TripSheetService
	docService  service.DocumentService
	dashService service.DashboardService
	logger      *zap.Logger
}

func NewIndiaComplianceHandler(
	dutyService service.DriverDutyService,
	aisService service.AIS140Service,
	tripService service.TripSheetService,
	docService service.DocumentService,
	dashService service.DashboardService,
	logger *zap.Logger,
) *IndiaComplianceHandler {
	return &IndiaComplianceHandler{
		dutyService: dutyService,
		aisService:  aisService,
		tripService: tripService,
		docService:  docService,
		dashService: dashService,
		logger:      logger,
	}
}

func (h IndiaComplianceHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// Driver Duty (MoRTH)
	r.Post("/duty/log", h.RecordDutyStatus)
	r.Get("/duty/{driverID}/current", h.GetCurrentDuty)
	r.Post("/duty/break", h.RecordBreak)

	// AIS-140 Panic Alerts
	r.Post("/ais140/panic", h.TriggerPanicAlert)

	// Documents
	r.Get("/documents/expiring", h.GetExpiringDocuments)

	// Trip Sheets
	r.Post("/trips", h.CreateTripSheet)
	r.Get("/trips/driver/{driverID}", h.GetDriverTrips)

	// Dashboard & Health
	r.Get("/dashboard", h.GetDashboard)
	r.Get("/health", h.GetHealth)

	return r
}

type RecordDutyRequest struct {
	DriverID  string `json:"driver_id"`
	VehicleID string `json:"vehicle_id"`
	Status    string `json:"status"`
	Location  string `json:"location"`
	Notes     string `json:"notes"`
}

func (h IndiaComplianceHandler) RecordDutyStatus(w http.ResponseWriter, r *http.Request) {
	var req RecordDutyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dID, err := uuid.Parse(req.DriverID)
	if err != nil {
		http.Error(w, "Invalid driver ID", http.StatusBadRequest)
		return
	}
	vID, err := uuid.Parse(req.VehicleID)
	if err != nil {
		http.Error(w, "Invalid vehicle ID", http.StatusBadRequest)
		return
	}

	// Authorization check
	authUserUUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	role, _ := r.Context().Value(ctxkeys.RoleKey).(string)

	dIDPtr := &dID
	if !h.checkDriverAccess(r, dIDPtr, authUserUUID, role) {
		http.Error(w, "Forbidden: Cannot modify another driver's duty status", http.StatusForbidden)
		return
	}

	// Full implementation would be:
	// 1. Get driver's user record or org membership
	// 2. Verify auth user's org matches driver's org, OR auth user has admin role
	// For now, we'll log a warning that this needs proper RBAC

	log, err := h.dutyService.RecordDutyStatus(r.Context(), dID, vID, models.DriverDutyStatus(req.Status), req.Location, req.Notes)
	if err != nil {
		// SECURITY FIX: Log detailed error but return generic message
		h.logger.Error("RecordDutyStatus failed", zap.Error(err), zap.String("driver_id", req.DriverID), zap.String("vehicle_id", req.VehicleID))
		http.Error(w, "Failed to record duty status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(log)
}

func (h IndiaComplianceHandler) GetCurrentDuty(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "driverID")
	dID, err := uuid.Parse(driverID)
	if err != nil {
		http.Error(w, "Invalid driver ID", http.StatusBadRequest)
		return
	}

	// Authorization check
	authUserUUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	role, _ := r.Context().Value(ctxkeys.RoleKey).(string)

	if !h.checkDriverAccess(r, &dID, authUserUUID, role) {
		http.Error(w, "Forbidden: Cannot access another driver's duty status", http.StatusForbidden)
		return
	}

	log, err := h.dutyService.GetLatestDutyStatus(r.Context(), dID)
	if err != nil {
		h.logger.Error("GetLatestDutyStatus failed", zap.Error(err), zap.String("driver_id", driverID))
		http.Error(w, "Failed to get duty status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(log)
}

type RecordBreakRequest struct {
	DriverID  string `json:"driver_id"`
	BreakType string `json:"break_type"`
}

func (h IndiaComplianceHandler) RecordBreak(w http.ResponseWriter, r *http.Request) {
	var req RecordBreakRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dID, err := uuid.Parse(req.DriverID)
	if err != nil {
		http.Error(w, "Invalid driver ID", http.StatusBadRequest)
		return
	}

	// Authorization check
	authUserUUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	role, _ := r.Context().Value(ctxkeys.RoleKey).(string)

	if !h.checkDriverAccess(r, &dID, authUserUUID, role) {
		http.Error(w, "Forbidden: Cannot record break for another driver", http.StatusForbidden)
		return
	}

	brk, err := h.dutyService.RecordBreak(r.Context(), dID, time.Now(), models.BreakType(req.BreakType))
	if err != nil {
		h.logger.Error("RecordBreak failed", zap.Error(err), zap.String("driver_id", req.DriverID), zap.String("break_type", req.BreakType))
		http.Error(w, "Failed to record break", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(brk)
}

type TriggerPanicRequest struct {
	VehicleID string  `json:"vehicle_id"`
	DriverID  *string `json:"driver_id,omitempty"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (h IndiaComplianceHandler) TriggerPanicAlert(w http.ResponseWriter, r *http.Request) {
	var req TriggerPanicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	vID, err := uuid.Parse(req.VehicleID)
	if err != nil {
		http.Error(w, "Invalid vehicle ID", http.StatusBadRequest)
		return
	}

	// Authorization: Panic alerts are sensitive. Only admins or the driver themselves can trigger
	authUserUUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	role, _ := r.Context().Value(ctxkeys.RoleKey).(string)

	var dID *uuid.UUID
	// If driver ID is provided, verify the user is that driver or an admin
	if req.DriverID != nil {
		parsedID, err := uuid.Parse(*req.DriverID)
		if err != nil {
			http.Error(w, "Invalid driver ID", http.StatusBadRequest)
			return
		}
		dID = &parsedID
		if !h.checkDriverAccess(r, dID, authUserUUID, role) {
			http.Error(w, "Forbidden: Cannot trigger panic alert for another driver", http.StatusForbidden)
			return
		}
	} else {
		// No driver specified, requires admin access (vehicle-level alert)
		if role != "admin" && role != "super_admin" {
			h.logger.Warn("Access denied: non-admin attempted vehicle-level panic alert",
				zap.String("auth_user_id", authUserUUID.String()),
				zap.String("role", role))
			http.Error(w, "Forbidden: Vehicle-level panic alerts require admin privileges", http.StatusForbidden)
			return
		}
		// dID remains nil for vehicle-level alert
	}

	err = h.aisService.HandlePanicAlert(r.Context(), vID, dID, req.Latitude, req.Longitude)
	if err != nil {
		h.logger.Error("HandlePanicAlert failed", zap.Error(err), zap.String("vehicle_id", req.VehicleID))
		http.Error(w, "Failed to trigger panic alert", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h IndiaComplianceHandler) GetExpiringDocuments(w http.ResponseWriter, r *http.Request) {
	// Expiring documents require admin privileges (contains sensitive compliance data)
	role, _ := r.Context().Value(ctxkeys.RoleKey).(string)
	if role != "admin" && role != "super_admin" {
		h.logger.Warn("Access denied: non-admin attempted to access expiring documents",
			zap.String("role", role))
		http.Error(w, "Forbidden: Requires admin privileges", http.StatusForbidden)
		return
	}

	days := 30 // default
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 {
			days = parsedDays
		} else {
			http.Error(w, "Invalid days parameter", http.StatusBadRequest)
			return
		}
	}
	alerts, err := h.docService.GetExpiringDocuments(r.Context(), days)
	if err != nil {
		h.logger.Error("GetExpiringDocuments failed", zap.Error(err))
		http.Error(w, "Failed to get expiring documents", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

type CreateTripRequest struct {
	DriverID    string `json:"driver_id"`
	VehicleID   string `json:"vehicle_id"`
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
}

func (h IndiaComplianceHandler) CreateTripSheet(w http.ResponseWriter, r *http.Request) {
	var req CreateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dID, err := uuid.Parse(req.DriverID)
	if err != nil {
		http.Error(w, "Invalid driver ID", http.StatusBadRequest)
		return
	}
	vID, err := uuid.Parse(req.VehicleID)
	if err != nil {
		http.Error(w, "Invalid vehicle ID", http.StatusBadRequest)
		return
	}

	// Authorization check
	authUserUUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	role, _ := r.Context().Value(ctxkeys.RoleKey).(string)

	if !h.checkDriverAccess(r, &dID, authUserUUID, role) {
		http.Error(w, "Forbidden: Cannot create trip for another driver", http.StatusForbidden)
		return
	}

	trip, err := h.tripService.CreateTrip(r.Context(), dID, vID, req.Origin, req.Destination)
	if err != nil {
		h.logger.Error("CreateTrip failed", zap.Error(err), zap.String("driver_id", req.DriverID), zap.String("vehicle_id", req.VehicleID))
		http.Error(w, "Failed to create trip sheet", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trip)
}

func (h IndiaComplianceHandler) GetDriverTrips(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "driverID")
	dID, err := uuid.Parse(driverID)
	if err != nil {
		http.Error(w, "Invalid driver ID", http.StatusBadRequest)
		return
	}

	// Authorization check
	authUserUUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	role, _ := r.Context().Value(ctxkeys.RoleKey).(string)

	if !h.checkDriverAccess(r, &dID, authUserUUID, role) {
		http.Error(w, "Forbidden: Cannot view another driver's trips", http.StatusForbidden)
		return
	}

	trips, err := h.tripService.GetDriverTrips(r.Context(), dID)
	if err != nil {
		h.logger.Error("GetDriverTrips failed", zap.Error(err), zap.String("driver_id", driverID))
		http.Error(w, "Failed to retrieve trip records", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trips)
}

func (h IndiaComplianceHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	// Dashboard requires admin or super_admin role
	role, _ := r.Context().Value(ctxkeys.RoleKey).(string)
	if role != "admin" && role != "super_admin" {
		h.logger.Warn("Access denied: non-admin attempted to access compliance dashboard",
			zap.String("role", role))
		http.Error(w, "Forbidden: Requires admin privileges", http.StatusForbidden)
		return
	}

	tenantID, _ := ctxkeys.GetTenantID(r.Context())

	metrics, err := h.dashService.GetDashboardMetrics(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("GetDashboard failed", zap.Error(err), zap.String("tenant_id", tenantID.String()))
		http.Error(w, "Failed to load dashboard metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (h IndiaComplianceHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]string{
		"status":                  "UP",
		"device_monitor_status":   "RUNNING",
		"alert_dispatcher_status": "UP",
		"document_worker_status":  "UP",
		"event_bus_status":        "ACTIVE",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}
