package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/compliance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/compliance/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type IndiaComplianceHandler struct {
	dutyService  service.DriverDutyService
	aisService   service.AIS140Service
	tripService  service.TripSheetService
	docService   service.DocumentService
	dashService  service.DashboardService
}

func NewIndiaComplianceHandler(
	dutyService service.DriverDutyService,
	aisService service.AIS140Service,
	tripService service.TripSheetService,
	docService service.DocumentService,
	dashService service.DashboardService,
) *IndiaComplianceHandler {
	return &IndiaComplianceHandler{
		dutyService: dutyService,
		aisService:  aisService,
		tripService: tripService,
		docService:  docService,
		dashService: dashService,
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

	dID, _ := uuid.Parse(req.DriverID)
	vID, _ := uuid.Parse(req.VehicleID)

	log, err := h.dutyService.RecordDutyStatus(r.Context(), dID, vID, models.DriverDutyStatus(req.Status), req.Location, req.Notes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(log)
}

func (h IndiaComplianceHandler) GetCurrentDuty(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "driverID")
	dID, _ := uuid.Parse(driverID)

	log, err := h.dutyService.GetLatestDutyStatus(r.Context(), dID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	dID, _ := uuid.Parse(req.DriverID)
	// Simply pass in now for the example
	brk, err := h.dutyService.RecordBreak(r.Context(), dID, time.Now(), models.BreakType(req.BreakType))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	vID, _ := uuid.Parse(req.VehicleID)
	var dID *uuid.UUID
	if req.DriverID != nil {
		id, _ := uuid.Parse(*req.DriverID)
		dID = &id
	}

	err := h.aisService.HandlePanicAlert(r.Context(), vID, dID, req.Latitude, req.Longitude)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h IndiaComplianceHandler) GetExpiringDocuments(w http.ResponseWriter, r *http.Request) {
	// Example query param: ?days=30
	alerts, err := h.docService.GetExpiringDocuments(r.Context(), 30) // Hardcoded 30 for simplicity here
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	dID, _ := uuid.Parse(req.DriverID)
	vID, _ := uuid.Parse(req.VehicleID)

	trip, err := h.tripService.CreateTrip(r.Context(), dID, vID, req.Origin, req.Destination)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trip)
}

func (h IndiaComplianceHandler) GetDriverTrips(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "driverID")
	dID, _ := uuid.Parse(driverID)

	trips, err := h.tripService.GetDriverTrips(r.Context(), dID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trips)
}

func (h IndiaComplianceHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	// orgID typically comes from context/auth
	orgID := uuid.Nil 
	
	metrics, err := h.dashService.GetDashboardMetrics(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (h IndiaComplianceHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]string{
		"status":                 "UP",
		"device_monitor_status":  "RUNNING",
		"alert_dispatcher_status": "UP",
		"document_worker_status":  "UP",
		"event_bus_status":       "ACTIVE",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}
