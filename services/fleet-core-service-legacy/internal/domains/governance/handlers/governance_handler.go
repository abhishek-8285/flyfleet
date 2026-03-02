package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/flyfleet/fleet-core-service/internal/domains/governance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/governance/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type GovernanceHandler struct {
	svc service.GovernanceService
}

func NewGovernanceHandler(svc service.GovernanceService) *GovernanceHandler {
	return &GovernanceHandler{svc: svc}
}

func (h *GovernanceHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	orgID, _ := uuid.Parse(r.Header.Get("X-Organization-ID"))
	userID, _ := uuid.Parse(r.Header.Get("X-User-ID"))

	var req models.CreateGovernanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	gReq, err := h.svc.CreateRequest(r.Context(), orgID, userID, req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, gReq)
}

func (h *GovernanceHandler) GetRequest(w http.ResponseWriter, r *http.Request) {
	orgID, _ := uuid.Parse(r.Header.Get("X-Organization-ID"))
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request id")
		return
	}

	gReq, err := h.svc.GetRequest(r.Context(), id, orgID)
	if err != nil {
		if err == service.ErrRequestNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, gReq)
}

func (h *GovernanceHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	orgID, _ := uuid.Parse(r.Header.Get("X-Organization-ID"))
	
	query := r.URL.Query()
	filter := models.GovernanceRequestFilter{
		OrganizationID: orgID,
		Search:         query.Get("search"),
	}

	if t := query.Get("type"); t != "" {
		gt := models.GovernanceRequestType(t)
		filter.Type = &gt
	}

	if s := query.Get("status"); s != "" {
		filter.Status = []models.GovernanceRequestStatus{models.GovernanceRequestStatus(s)}
	}

	filter.Page, _ = strconv.Atoi(query.Get("page"))
	filter.PageSize, _ = strconv.Atoi(query.Get("page_size"))

	resp, err := h.svc.ListRequests(r.Context(), filter)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, resp)
}

func (h *GovernanceHandler) ApproveRequest(w http.ResponseWriter, r *http.Request) {
	orgID, _ := uuid.Parse(r.Header.Get("X-Organization-ID"))
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request id")
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.svc.ApproveRequest(r.Context(), id, orgID, req.Notes); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GovernanceHandler) RejectRequest(w http.ResponseWriter, r *http.Request) {
	orgID, _ := uuid.Parse(r.Header.Get("X-Organization-ID"))
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request id")
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.svc.RejectRequest(r.Context(), id, orgID, req.Notes); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GovernanceHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	orgID, _ := uuid.Parse(r.Header.Get("X-Organization-ID"))
	
	settings, err := h.svc.GetSettings(r.Context(), orgID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, settings)
}

func (h *GovernanceHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	orgID, _ := uuid.Parse(r.Header.Get("X-Organization-ID"))
	userID, _ := uuid.Parse(r.Header.Get("X-User-ID"))
	
	var settings models.GovernanceSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.UpdateSettings(r.Context(), orgID, userID, &settings); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, settings)
}

func (h *GovernanceHandler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *GovernanceHandler) writeErrorResponse(w http.ResponseWriter, status int, message string) {
	h.writeJSONResponse(w, status, map[string]string{"error": message})
}
