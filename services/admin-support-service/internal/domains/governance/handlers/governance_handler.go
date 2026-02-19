package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/flyfleet/admin-support-service/internal/domains/governance/models"
	"github.com/flyfleet/admin-support-service/internal/domains/governance/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/fleet-platform/shared/pkg/middleware"
)

type GovernanceHandler struct {
	svc service.GovernanceService
}

func NewGovernanceHandler(svc service.GovernanceService) *GovernanceHandler {
	return &GovernanceHandler{svc: svc}
}

func (h *GovernanceHandler) RegisterRoutes(r chi.Router) {
	r.Route("/governance", func(r chi.Router) {
		r.Post("/requests", h.CreateRequest)
		r.Get("/requests", h.ListRequests)
		r.Get("/requests/{id}", h.GetRequest)
		r.Post("/requests/{id}/approve", h.ApproveRequest)
		r.Post("/requests/{id}/reject", h.RejectRequest)

		r.Get("/settings", h.GetSettings)
		r.Put("/settings", h.UpdateSettings)
	})
}

func (h *GovernanceHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)

	var req models.CreateGovernanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	gReq, err := h.svc.CreateRequest(r.Context(), orgID, userID, req)
	if err != nil {
		h.errorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, gReq, http.StatusCreated)
}

func (h *GovernanceHandler) GetRequest(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorJSON(w, "invalid request id", http.StatusBadRequest)
		return
	}

	gReq, err := h.svc.GetRequest(r.Context(), id, orgID)
	if err != nil {
		if err == service.ErrRequestNotFound {
			h.errorJSON(w, err.Error(), http.StatusNotFound)
			return
		}
		h.errorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, gReq, http.StatusOK)
}

func (h *GovernanceHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	
	var filter models.GovernanceRequestFilter
	query := r.URL.Query()
	
	if t := query.Get("type"); t != "" {
		gt := models.GovernanceRequestType(t)
		filter.Type = &gt
	}
	
	if s := query.Get("status"); s != "" {
		filter.Status = []models.GovernanceRequestStatus{models.GovernanceRequestStatus(s)}
	}
	
	if p := query.Get("page"); p != "" {
		filter.Page, _ = strconv.Atoi(p)
	}
	if ps := query.Get("page_size"); ps != "" {
		filter.PageSize, _ = strconv.Atoi(ps)
	}
	
	filter.OrganizationID = orgID

	resp, err := h.svc.ListRequests(r.Context(), filter)
	if err != nil {
		h.errorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, resp, http.StatusOK)
}

func (h *GovernanceHandler) ApproveRequest(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorJSON(w, "invalid request id", http.StatusBadRequest)
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.svc.ApproveRequest(r.Context(), id, orgID, req.Notes); err != nil {
		h.errorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GovernanceHandler) RejectRequest(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorJSON(w, "invalid request id", http.StatusBadRequest)
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.svc.RejectRequest(r.Context(), id, orgID, req.Notes); err != nil {
		h.errorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GovernanceHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	
	settings, err := h.svc.GetSettings(r.Context(), orgID)
	if err != nil {
		h.errorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, settings, http.StatusOK)
}

func (h *GovernanceHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	
	var settings models.GovernanceSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		h.errorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.svc.UpdateSettings(r.Context(), orgID, &settings); err != nil {
		h.errorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, settings, http.StatusOK)
}

// Helpers
func (h *GovernanceHandler) json(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *GovernanceHandler) errorJSON(w http.ResponseWriter, message string, status int) {
	h.json(w, map[string]string{"error": message}, status)
}

func (h *GovernanceHandler) getOrgID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.OrgIDKey); val != nil {
		if orgIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(orgIDStr); err == nil {
				return id
			}
		}
	}
	// Fallback to Header if not in context (e.g. legacy or internal calls)
	if headerID := r.Header.Get("X-Organization-ID"); headerID != "" {
		if id, err := uuid.Parse(headerID); err == nil {
			return id
		}
	}
	return uuid.Nil
}

func (h *GovernanceHandler) getUserID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.UserIDKey); val != nil {
		if userIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(userIDStr); err == nil {
				return id
			}
		}
	}
	if headerID := r.Header.Get("X-User-ID"); headerID != "" {
		if id, err := uuid.Parse(headerID); err == nil {
			return id
		}
	}
	return uuid.Nil
}
