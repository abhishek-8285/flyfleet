package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/flyfleet/fleet-core-service/internal/domains/governance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/governance/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type GovernanceHandler struct {
	svc            service.GovernanceService
	exportSvc       service.ExportService
	accessReviewSvc service.AccessReviewService
}

func NewGovernanceHandler(
	svc service.GovernanceService,
	exportSvc service.ExportService,
	accessReviewSvc service.AccessReviewService,
) *GovernanceHandler {
	return &GovernanceHandler{
		svc:             svc,
		exportSvc:       exportSvc,
		accessReviewSvc: accessReviewSvc,
	}
}

// parseUUIDs extracts and validates organization and user IDs from request headers
func (h *GovernanceHandler) parseUUIDs(r *http.Request) (orgID uuid.UUID, userID uuid.UUID, ok bool) {
	orgIDStr := r.Header.Get("X-Organization-ID")
	userIDStr := r.Header.Get("X-User-ID")

	var err error
	orgID, err = uuid.Parse(orgIDStr)
	if err != nil {
		return uuid.Nil, uuid.Nil, false
	}

	userID, err = uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, uuid.Nil, false
	}

	return orgID, userID, true
}

func (h *GovernanceHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	orgID, userID, ok := h.parseUUIDs(r)
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "invalid or missing organization/user ID")
		return
	}

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
	orgID, _, ok := h.parseUUIDs(r)
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "invalid or missing organization ID")
		return
	}

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
	orgID, _, ok := h.parseUUIDs(r)
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "invalid or missing organization ID")
		return
	}

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
	orgID, userID, ok := h.parseUUIDs(r)
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "invalid or missing organization/user ID")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request id")
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Add user_id to context for audit logging
	ctx := context.WithValue(r.Context(), "user_id", userID.String())

	if err := h.svc.ApproveRequest(ctx, id, orgID, req.Notes); err != nil {
		if err == service.ErrRequestNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		if err == service.ErrInvalidStatus {
			h.writeErrorResponse(w, http.StatusConflict, err.Error())
			return
		}
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GovernanceHandler) RejectRequest(w http.ResponseWriter, r *http.Request) {
	orgID, userID, ok := h.parseUUIDs(r)
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "invalid or missing organization/user ID")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request id")
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Add user_id to context for audit logging
	ctx := context.WithValue(r.Context(), "user_id", userID.String())

	if err := h.svc.RejectRequest(ctx, id, orgID, req.Notes); err != nil {
		if err == service.ErrRequestNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		if err == service.ErrInvalidStatus {
			h.writeErrorResponse(w, http.StatusConflict, err.Error())
			return
		}
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GovernanceHandler) VerifyRequest(w http.ResponseWriter, r *http.Request) {
	orgID, userID, ok := h.parseUUIDs(r)
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "invalid or missing organization/user ID")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request id")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=VERIFIED FAILED"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.VerifyRequest(r.Context(), id, orgID, userID, req.Status); err != nil {
		if err == service.ErrRequestNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		if err == service.ErrInvalidStatus {
			h.writeErrorResponse(w, http.StatusConflict, err.Error())
			return
		}
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GovernanceHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	orgID, _, ok := h.parseUUIDs(r)
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "invalid or missing organization ID")
		return
	}

	settings, err := h.svc.GetSettings(r.Context(), orgID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, settings)
}

func (h *GovernanceHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	orgID, userID, ok := h.parseUUIDs(r)
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "invalid or missing organization/user ID")
		return
	}

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

func (h *GovernanceHandler) GetDownloadLink(w http.ResponseWriter, r *http.Request) {
	orgID, _, ok := h.parseUUIDs(r)
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "invalid or missing organization ID")
		return
	}

	requestID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request id")
		return
	}

	// 1. Check if request is approved and file is ready
	req, err := h.svc.GetRequest(r.Context(), requestID, orgID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "request not found")
		return
	}

	if req.Status != models.RequestStatusCompleted {
		h.writeErrorResponse(w, http.StatusForbidden, "export is not ready or has expired")
		return
	}

	// 2. Generate short-lived token
	// First get the export metadata
	export, err := h.exportSvc.ProcessExportRequest(r.Context(), requestID) // Re-process if needed or get from DB
	// Actually we should have a GetExportByRequestID in exportSvc or just use Processexport if it's idempotent
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to retrieve export")
		return
	}

	token, err := h.exportSvc.GenerateDownloadToken(r.Context(), export.ID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"download_url": "/api/v1/governance/download/" + token.Token.String(),
		"expires_at":   token.ExpiresAt,
	})
}

func (h *GovernanceHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	tokenStr := chi.URLParam(r, "token")
	token, err := uuid.Parse(tokenStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid token")
		return
	}

	export, err := h.exportSvc.ValidateToken(r.Context(), token)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "link expired or invalid")
		return
	}

	// Serve the file
	file, err := os.Open(export.FilePath)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to open export file")
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(export.ID.String()+".json"))
	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(w, file)
}

func (h *GovernanceHandler) GenerateAccessReview(w http.ResponseWriter, r *http.Request) {
	orgID, _, ok := h.parseUUIDs(r)
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "invalid or missing organization ID")
		return
	}

	report, err := h.accessReviewSvc.GenerateReport(r.Context(), orgID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, report)
}

func (h *GovernanceHandler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *GovernanceHandler) writeErrorResponse(w http.ResponseWriter, status int, message string) {
	h.writeJSONResponse(w, status, map[string]string{"error": message})
}
