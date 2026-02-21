package reports

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handler struct {
	service Service
	logger  *zap.Logger
}

func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.CreateReport)
	r.Get("/", h.ListReports)
	r.Get("/{id}", h.GetReport)
	r.Get("/{id}/download", h.DownloadReport)
}

func (h *Handler) CreateReport(w http.ResponseWriter, r *http.Request) {
	var req CreateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	report, err := h.service.CreateReport(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create report", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create report")
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, report)
}

func (h *Handler) GetReport(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	report, err := h.service.GetReport(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Report not found")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, report)
}

func (h *Handler) ListReports(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limitStr = "10"
	}
	limit, _ := strconv.Atoi(limitStr)

	offsetStr := r.URL.Query().Get("offset")
	if offsetStr == "" {
		offsetStr = "0"
	}
	offset, _ := strconv.Atoi(offsetStr)

	reports, err := h.service.ListReports(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list reports", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list reports")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": reports})
}

func (h *Handler) DownloadReport(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	report, err := h.service.GetReport(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Report not found")
		return
	}
	if report.Status != ReportStatusCompleted {
		h.writeErrorResponse(w, http.StatusConflict, "Report is not ready for download")
		return
	}

	var csvBuffer bytes.Buffer
	writer := csv.NewWriter(&csvBuffer)
	_ = writer.Write([]string{"report_id", "report_type", "status", "generated_at"})
	_ = writer.Write([]string{report.ID.String(), string(report.Type), string(report.Status), time.Now().UTC().Format(time.RFC3339)})
	writer.Flush()

	filename := fmt.Sprintf("report_%s.csv", report.ID.String())
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(csvBuffer.Bytes())
}

func (h *Handler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeErrorResponse(w http.ResponseWriter, status int, message string) {
	h.writeJSONResponse(w, status, map[string]string{"error": message})
}
