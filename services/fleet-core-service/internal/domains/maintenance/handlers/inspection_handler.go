package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core-service/internal/domains/maintenance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/maintenance/repository"
)

// InspectionHandler handles HTTP requests for digital inspections
type InspectionHandler struct {
	repo   *repository.InspectionRepository
	logger *zap.Logger
}

// NewInspectionHandler creates a new inspection handler
func NewInspectionHandler(repo *repository.InspectionRepository, logger *zap.Logger) *InspectionHandler {
	return &InspectionHandler{
		repo:   repo,
		logger: logger,
	}
}

func (h *InspectionHandler) RegisterRoutes(r chi.Router) {
	r.Route("/templates", func(r chi.Router) {
		r.Get("/", h.ListTemplates)
		r.Post("/", h.CreateTemplate)
		r.Get("/{id}", h.GetTemplate)
	})
	r.Route("/reports", func(r chi.Router) {
		r.Post("/", h.CreateInspection)
		r.Get("/{id}", h.GetInspection)
		r.Get("/vehicle/{vehicleId}", h.ListByVehicle)
	})
}

// --- Templates ---

func (h *InspectionHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	vehicleType := r.URL.Query().Get("vehicle_type")

	templates, err := h.repo.ListTemplates(r.Context(), orgID, vehicleType)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, templates)
}

func (h *InspectionHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)

	var template models.InspectionTemplate
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	template.OrganizationID = orgID

	if err := h.repo.CreateTemplate(r.Context(), &template); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, template)
}

func (h *InspectionHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid template ID")
		return
	}

	template, err := h.repo.GetTemplateByID(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if template == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "template not found")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, template)
}

// --- Inspections ---

func (h *InspectionHandler) CreateInspection(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)

	var inspection models.VehicleInspection
	if err := json.NewDecoder(r.Body).Decode(&inspection); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	inspection.OrganizationID = orgID
	inspection.InspectorID = userID

	if err := h.repo.CreateInspection(r.Context(), &inspection); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, inspection)
}

func (h *InspectionHandler) GetInspection(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid inspection ID")
		return
	}

	inspection, err := h.repo.GetInspectionByID(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if inspection == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "inspection not found")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, inspection)
}

func (h *InspectionHandler) ListByVehicle(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := uuid.Parse(chi.URLParam(r, "vehicleId"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid vehicle ID")
		return
	}

	inspections, err := h.repo.ListByVehicle(r.Context(), vehicleID, 10)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, inspections)
}

// Helpers
func (h *InspectionHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *InspectionHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	h.writeJSONResponse(w, statusCode, map[string]string{"error": message})
}

func (h *InspectionHandler) getOrgID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.OrgIDKey); val != nil {
		if orgIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(orgIDStr); err == nil {
				return id
			}
		}
	}
	return uuid.MustParse("00000000-0000-0000-0000-000000000001")
}

func (h *InspectionHandler) getUserID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.UserIDKey); val != nil {
		if userIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(userIDStr); err == nil {
				return id
			}
		}
	}
	return uuid.MustParse("00000000-0000-0000-0000-000000000001")
}
