package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/reports/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/reports/service"
	"github.com/flyfleet/fleet-core-service/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ReportHandler struct {
	svc      *service.ReportService
	logger   *zap.Logger
	validate *validator.Validate
}

func NewReportHandler(svc *service.ReportService, logger *zap.Logger) *ReportHandler {
	return &ReportHandler{
		svc:      svc,
		logger:   logger,
		validate: validator.New(),
	}
}

func (h *ReportHandler) RegisterRoutes(r chi.Router) {
	r.Post("/schedules", h.CreateReport)
	r.Get("/schedules", h.ListReports)
	r.Route("/schedules/{id}", func(r chi.Router) {
		r.Get("/", h.GetReport)
		r.Put("/", h.UpdateReport)
		r.Post("/pause", h.PauseReport)
		r.Post("/resume", h.ResumeReport)
	})
	r.Get("/executions", h.ListExecutions)
	r.Get("/settings", h.GetSettings)
	r.Put("/settings", h.UpdateSettings)
}

func (h *ReportHandler) decodeAndValidate(w http.ResponseWriter, r *http.Request, dest interface{}) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return false
	}
	if err := h.validate.Struct(dest); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed", err)
		return false
	}
	return true
}

func (h *ReportHandler) CreateReport(w http.ResponseWriter, r *http.Request) {
	var req models.ScheduledReport
	if !h.decodeAndValidate(w, r, &req) {
		return
	}

	// Get Organization ID and User ID from context (should be set by auth middleware)
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		h.logger.Warn("Missing org_id in context - request likely bypassed authentication middleware")
		h.writeErrorResponse(w, http.StatusUnauthorized, "organization context required", nil)
		return
	}
	req.OrganizationID = orgID

	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		h.logger.Warn("Missing user_id in context - request likely bypassed authentication middleware")
		h.writeErrorResponse(w, http.StatusUnauthorized, "user context required", nil)
		return
	}
	req.CreatedBy = userID

	id, err := h.svc.CreateReport(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create report", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to create report", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, map[string]interface{}{"id": id})
}

func (h *ReportHandler) ListReports(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	reports, total, err := h.svc.ListReports(r.Context(), page, limit)
	if err != nil {
		h.logger.Error("Failed to list reports", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to list reports", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":  reports,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *ReportHandler) GetReport(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid id used", err)
		return
	}

	report, err := h.svc.GetReport(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get report", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get report", err)
		return
	}
	if report == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "report not found", nil)
		return
	}
	h.writeJSONResponse(w, http.StatusOK, report)
}

func (h *ReportHandler) UpdateReport(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid id used", err)
		return
	}

	var req models.ScheduledReport
	if !h.decodeAndValidate(w, r, &req) {
		return
	}
	req.ID = id
	req.UpdatedBy = uuid.New()
	req.UpdatedAt = time.Now()

	if err := h.svc.UpdateReport(r.Context(), &req); err != nil {
		h.logger.Error("Failed to update report", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to update report", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ReportHandler) PauseReport(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid id used", err)
		return
	}
	if err := h.svc.PauseReport(r.Context(), id); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to pause", err)
		return
	}
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "paused"})
}

func (h *ReportHandler) ResumeReport(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid id used", err)
		return
	}
	if err := h.svc.ResumeReport(r.Context(), id); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to resume", err)
		return
	}
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "resumed"})
}

func (h *ReportHandler) ListExecutions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	reportIDStr := query.Get("report_id")
	var reportID uuid.UUID
	if reportIDStr != "" {
		reportID, _ = uuid.Parse(reportIDStr)
	}

	page, _ := strconv.Atoi(query.Get("page"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	execs, total, err := h.svc.ListExecutions(r.Context(), reportID, page, limit)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to list executions", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"executions": execs,
		"total":      total,
		"page":       page,
		"limit":      limit,
	})
}

func (h *ReportHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	s, err := h.svc.GetSettings(r.Context())
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to fetch settings", err)
		return
	}
	h.writeJSONResponse(w, http.StatusOK, s)
}

func (h *ReportHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req models.GlobalSettings
	if !h.decodeAndValidate(w, r, &req) {
		return
	}
	if err := h.svc.UpdateSettings(r.Context(), &req); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to update settings", err)
		return
	}
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ReportHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *ReportHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	h.writeJSONResponse(w, statusCode, map[string]string{"error": message})
}
