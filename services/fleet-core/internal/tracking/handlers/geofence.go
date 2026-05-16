package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/middleware"
	"github.com/flyfleet/fleet-core/internal/tracking/models"
	"github.com/flyfleet/fleet-core/internal/tracking/repository"
	"github.com/go-chi/chi/v5"
)

// GeofenceHandler handles GET/POST /api/v1/geofences
type GeofenceHandler struct {
	geoRepo repository.GeofenceRepositoryer
	logger  *zap.Logger
}

func NewGeofenceHandler(repo repository.GeofenceRepositoryer, logger *zap.Logger) *GeofenceHandler {
	return &GeofenceHandler{geoRepo: repo, logger: logger}
}

func (h *GeofenceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPost:
		h.handlePost(w, r)
	case http.MethodDelete:
		h.handleDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *GeofenceHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	// FIX: GetTenantID takes context.Context, returns google/uuid — convert to gofrs
	middlewareTenantID := middleware.GetTenantID(r.Context())
	tenantID, _ := uuid.FromString(middlewareTenantID.String())

	geofences, err := h.geoRepo.GetOrgGeofences(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("failed to list geofences", zap.Error(err))
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	if geofences == nil {
		geofences = []models.Geofence{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": geofences,
	})
}

type createGeofenceRequest struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"` // circle | polygon
	CenterLat  *float64 `json:"center_lat,omitempty"`
	CenterLng  *float64 `json:"center_lng,omitempty"`
	RadiusM    *float64 `json:"radius_m,omitempty"`
	PolygonWKT string   `json:"polygon_wkt,omitempty"`
}

func (h *GeofenceHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	var req createGeofenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}

	// FIX: GetTenantID takes context.Context, returns google/uuid — convert to gofrs
	middlewareTenantID := middleware.GetTenantID(r.Context())
	tenantID, _ := uuid.FromString(middlewareTenantID.String())

	g := &models.Geofence{
		ID:          uuid.Must(uuid.NewV4()),
		TenantID:       tenantID,
		Name:        req.Name,
		Type:        req.Type,
		CenterLat:   req.CenterLat,
		CenterLng:   req.CenterLng,
		RadiusM:     req.RadiusM,
		PolygonWKT:  req.PolygonWKT,
	}

	if err := h.geoRepo.CreateGeofence(r.Context(), g); err != nil {
		h.logger.Error("failed to create geofence", zap.Error(err))
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": g,
	})
}

func (h *GeofenceHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, `{"error":"id is required"}`, http.StatusBadRequest)
		return
	}

	id, err := uuid.FromString(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid uuid"}`, http.StatusBadRequest)
		return
	}

	if err := h.geoRepo.DeleteGeofence(r.Context(), id); err != nil {
		h.logger.Error("failed to delete geofence", zap.Error(err))
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

