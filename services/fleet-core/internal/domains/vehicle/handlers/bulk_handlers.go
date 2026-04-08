package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/flyfleet/fleet-core/internal/domains/vehicle/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BulkHandlers struct {
	service *service.BulkService
	logger  *zap.Logger
}

func NewBulkHandlers(service *service.BulkService, logger *zap.Logger) *BulkHandlers {
	return &BulkHandlers{
		service: service,
		logger:  logger,
	}
}

func (h *BulkHandlers) RegisterRoutes(r chi.Router) {
	r.Post("/", h.CreateBulkJob)
	r.Get("/", h.ListBulkJobs)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.GetBulkJob)
		r.Get("/records", h.GetJobRecords)
		r.Post("/retry", h.RetryBulkJob)
	})
}

// CreateBulkJob godoc
func (h *BulkHandlers) CreateBulkJob(w http.ResponseWriter, r *http.Request) {
	// Parse Multipart Form (Max 50MB)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "File too large or invalid form", err)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Missing file part", err)
		return
	}
	defer file.Close()

	// Organization ID from context
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID not found", nil)
		return
	}

	job, err := h.service.CreateJob(r.Context(), orgID, header.Filename, file)
	if err != nil {
		h.logger.Error("Failed to create bulk job", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create job", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, job)
}

// ListBulkJobs godoc
func (h *BulkHandlers) ListBulkJobs(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)

	query := r.URL.Query()
	status := query.Get("status")
	page, _ := strconv.Atoi(query.Get("page"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	jobs, total, err := h.service.ListJobs(r.Context(), orgID, status, page, limit)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list jobs", err)
		return
	}

	response := map[string]interface{}{
		"jobs":  jobs,
		"total": total,
		"page":  page,
		"limit": limit,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetBulkJob godoc
func (h *BulkHandlers) GetBulkJob(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Missing Job ID", nil)
		return
	}

	jobID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid Job ID", err)
		return
	}

	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID not found", nil)
		return
	}

	job, err := h.service.GetJob(r.Context(), jobID, orgID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Job not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, job)
}

// GetJobRecords godoc
func (h *BulkHandlers) GetJobRecords(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	status := r.URL.Query().Get("status")

	jobID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid Job ID", err)
		return
	}

	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID not found", nil)
		return
	}

	records, err := h.service.GetJobRecords(r.Context(), jobID, orgID, status)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch records", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, records)
}

// RetryBulkJob godoc
func (h *BulkHandlers) RetryBulkJob(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Missing Job ID", nil)
		return
	}

	jobID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid Job ID", err)
		return
	}

	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID not found", nil)
		return
	}

	job, err := h.service.RetryJob(r.Context(), jobID, orgID)
	if err != nil {
		h.logger.Error("Failed to retry bulk job", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, job)
}

// Helpers

func (h *BulkHandlers) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

func (h *BulkHandlers) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	response := map[string]interface{}{
		"error": message,
	}

	if err != nil {
		var errDetail string
		if statusCode >= 500 {
			errDetail = "Internal server error"
		} else {
			errDetail = err.Error()
		}
		response["details"] = errDetail
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func (h *BulkHandlers) getOrgID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.OrgIDKey); val != nil {
		if orgIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(orgIDStr); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}
