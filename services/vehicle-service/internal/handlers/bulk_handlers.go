package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/flyfleet/vehicle-service/internal/service"
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

// CreateBulkJob godoc
// @Summary Create bulk upload job
// @Router /api/v1/bulk-jobs [post]
func (h *BulkHandlers) CreateBulkJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	// Organization ID from header (set by Gateway/Middleware)
	orgIDStr := r.Header.Get("X-Organization-ID")
	if orgIDStr == "" {
		// Fallback for local testing if not behind gateway, or error
		// defaulting to a test org ID or error
		orgIDStr = "00000000-0000-0000-0000-000000000000"
	}
	
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid Organization ID", err)
		return
	}

	ctx := context.Background()
	job, err := h.service.CreateJob(ctx, orgID, header.Filename, file)
	if err != nil {
		h.logger.Error("Failed to create bulk job", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create job", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, job)
}

// ListBulkJobs godoc
// @Summary List bulk upload jobs
// @Router /api/v1/bulk-jobs [get]
func (h *BulkHandlers) ListBulkJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orgIDStr := r.Header.Get("X-Organization-ID")
	if orgIDStr == "" {
		orgIDStr = "00000000-0000-0000-0000-000000000000"
	}
	orgID, _ := uuid.Parse(orgIDStr)

	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	
	if page <= 0 { page = 1 }
	if limit <= 0 { limit = 20 }

	ctx := context.Background()
	jobs, total, err := h.service.ListJobs(ctx, orgID, status, page, limit)
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
// @Summary Get bulk job detail
// @Router /api/v1/bulk-jobs/{id} [get]
func (h *BulkHandlers) GetBulkJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID using Chi helper
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
	
	ctx := context.Background()
	job, err := h.service.GetJob(ctx, jobID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Job not found", err)
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, job)
}

// GetJobRecords godoc
// @Summary Get bulk job records (errors/success)
// @Router /api/v1/bulk-jobs/{id}/records [get]
func (h *BulkHandlers) GetJobRecords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := chi.URLParam(r, "id")
	status := r.URL.Query().Get("status")
	
	jobID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid Job ID", err)
		return
	}
	
	ctx := context.Background()
	records, err := h.service.GetJobRecords(ctx, jobID, status)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch records", err)
		return
	}
	
	h.writeJSONResponse(w, http.StatusOK, records)
}

// RetryBulkJob godoc
// @Summary Retry a failed bulk job
// @Router /api/v1/bulk-jobs/{id}/retry [post]
func (h *BulkHandlers) RetryBulkJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	orgIDStr := r.Header.Get("X-Organization-ID")
	if orgIDStr == "" {
		orgIDStr = "00000000-0000-0000-0000-000000000000"
	}
	orgID, _ := uuid.Parse(orgIDStr)

	ctx := context.Background()
	job, err := h.service.RetryJob(ctx, jobID, orgID)
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
		// h.logger.Debug("error details", zap.Error(err))
		response["details"] = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
