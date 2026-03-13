package reports

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core-service/internal/pkg/ctxkeys"
)

type Handler struct {
	service Service
	repo    Repository
	storage ObjectStorage
	metrics *PrometheusMetrics
	logger  *zap.Logger
}

func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		storage: NewLocalStorage(logger),
		metrics: NewPrometheusMetrics(logger, 4),
		logger:  logger,
	}
}

func NewHandlerWithStorage(service Service, repo Repository, storage ObjectStorage, metrics *PrometheusMetrics, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		repo:    repo,
		storage: storage,
		metrics: metrics,
		logger:  logger,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	// Reports
	r.Post("/", h.CreateReport)
	r.Get("/", h.ListReports)
	r.Get("/{id}", h.GetReportStatus)
	r.Get("/{id}/download", h.DownloadReport)
	r.Post("/{id}/cancel", h.CancelReport)
	r.Delete("/{id}", h.DeleteReport)

	// Templates
	r.Post("/templates", h.CreateTemplate)
	r.Get("/templates", h.ListTemplates)

	// Observability
	r.Get("/metrics", h.ServeMetrics)

	// DLQ operations
	r.Post("/dlq/{jobId}/requeue", h.RequeueDLQJob)
}

func (h *Handler) getTenantID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(ctxkeys.TenantIDKey); val != nil {
		if tenantIDStr, ok := val.(string); ok {
			id, _ := uuid.Parse(tenantIDStr)
			return id
		}
	}
	return uuid.Nil
}

func (h *Handler) getUserID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(ctxkeys.UserIDKey); val != nil {
		if userIDStr, ok := val.(string); ok {
			id, _ := uuid.Parse(userIDStr)
			return id
		}
	}
	return uuid.New()
}

func (h *Handler) CreateReport(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		http.Error(w, "Unauthorized: invalid tenant ID", http.StatusUnauthorized)
		return
	}
	userID := h.getUserID(r)

	var req CreateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	resp, err := h.service.CreateReport(r.Context(), req, tenantID, userID)
	if err != nil {
		h.logger.Error("Failed to create report", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.metrics.IncCreated()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetReportStatus(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		http.Error(w, "Unauthorized: invalid tenant ID", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid report ID format", http.StatusBadRequest)
		return
	}

	resp, err := h.service.GetReportStatus(r.Context(), id, tenantID)
	if err != nil {
		if err.Error() == "report not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) ListReports(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		http.Error(w, "Unauthorized: invalid tenant ID", http.StatusUnauthorized)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	reportsList, err := h.service.ListReports(r.Context(), tenantID, limit, offset)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reportsList)
}

// DownloadReport generates a signed URL with expiry and verifies tenant ownership.
// Security: tenant_id ownership check + signed URL (15-min expiry) + no direct file path exposure.
func (h *Handler) DownloadReport(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		http.Error(w, "Unauthorized: invalid tenant ID", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid report ID format", http.StatusBadRequest)
		return
	}

	// Ownership check: GetReport enforces tenant_id condition in query
	report, err := h.service.GetReport(r.Context(), id, tenantID)
	if err != nil {
		if err.Error() == "report not found" {
			// Return 404 for both not-found and unauthorized (avoid enumeration)
			http.Error(w, "Report not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if report.Status != ReportStatusCompleted {
		http.Error(w, "Report is not ready for download", http.StatusBadRequest)
		return
	}

	// Generate tenant-scoped storage key (tenant isolation in path)
	storageKey := tenantID.String() + "/" + report.ID.String() + "/export.csv.gz"

	// Issue a signed URL with 15-minute expiry
	signedURL, err := h.storage.SignedURL(r.Context(), storageKey, 15*time.Minute)
	if err != nil {
		h.logger.Error("Failed to generate signed URL",
			zap.String("report_id", report.ID.String()),
			zap.Error(err))
		http.Error(w, "Failed to generate download URL", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Signed download URL issued",
		zap.String("report_id", report.ID.String()),
		zap.String("tenant_id", tenantID.String()))

	http.Redirect(w, r, signedURL, http.StatusFound)
}

func (h *Handler) CancelReport(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.service.CancelReport(r.Context(), id, tenantID); err != nil {
		if err.Error() == "report not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "cannot cancel a completed report" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteReport(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var tmpl ReportTemplate
	if err := json.NewDecoder(r.Body).Decode(&tmpl); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if err := h.service.CreateTemplate(r.Context(), &tmpl); err != nil {
		h.logger.Error("Failed to create template", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tmpl)
}

func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.service.ListTemplates(r.Context())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

// ServeMetrics exposes Prometheus-format metrics at GET /reports/metrics.
func (h *Handler) ServeMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(h.metrics.ServeMetrics()))
}

// RequeueDLQJob allows ops to manually requeue a dead-letter job.
func (h *Handler) RequeueDLQJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobId")
	if h.repo == nil {
		http.Error(w, "DLQ requeue not configured", http.StatusServiceUnavailable)
		return
	}
	if err := RequeueDLQJob(r.Context(), h.repo, jobID, h.logger); err != nil {
		http.Error(w, "Failed to requeue job: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
