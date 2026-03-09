package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/admin_support_v2/domains/dr/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/admin_support_v2/domains/dr/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/fleet-platform/shared/pkg/middleware"
)

type DrHandlers struct {
	service service.DrService
	logger  *zap.Logger
}

func NewDrHandlers(service service.DrService, logger *zap.Logger) *DrHandlers {
	return &DrHandlers{service: service, logger: logger}
}

func (h *DrHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/dr", func(r chi.Router) {
		r.Post("/backups", h.InitiateBackup)
		r.Get("/backups", h.ListBackups)
		r.Get("/backups/{id}", h.GetBackup)
		r.Post("/backups/{id}/verify", h.VerifyBackup)

		r.Post("/restore-jobs", h.InitiateRestore)
		r.Get("/restore-jobs", h.ListRestoreJobs)
		r.Get("/restore-jobs/{id}", h.GetRestoreJob)
		r.Get("/restore-jobs/{id}/logs", h.GetRestoreJobLogs)

		r.Get("/storage-pools", h.ListStoragePools)

		r.Post("/tests", h.CreateDRTest)
		r.Get("/tests", h.ListDRTests)
		r.Get("/tests/metrics", h.GetDRTestMetrics)
		r.Get("/tests/settings", h.GetDRTestSettings)
		r.Put("/tests/settings", h.UpdateDRTestSettings)
		r.Get("/tests/{id}", h.GetDRTestDetail)
		r.Post("/tests/{id}/abort", h.AbortDRTest)
	})
}

// Request Models
type InitiateBackupRequest struct {
	Type models.BackupType `json:"type" validate:"required,oneof=FULL INCREMENTAL SNAPSHOT"`
}

type InitiateRestoreRequest struct {
	BackupID          uuid.UUID `json:"backup_id" validate:"required"`
	TargetEnvironment string    `json:"target_environment" validate:"required"`
}

type CreateDRTestRequest struct {
	Name            string     `json:"name" validate:"required"`
	Description     string     `json:"description"`
	TestType        string     `json:"test_type" validate:"required,oneof=FAILOVER RESTORE BACKUP_VERIFY"`
	TargetEnv       string     `json:"target_env" validate:"required"`
	SuccessCriteria string     `json:"success_criteria"`
	ScheduledAt     *time.Time `json:"scheduled_at"`
}

func (h *DrHandlers) InitiateBackup(w http.ResponseWriter, r *http.Request) {
	var req InitiateBackupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := ValidateRequest(&req); err != nil {
		h.errorJSON(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)
	backup, err := h.service.InitiateBackup(r.Context(), req.Type, userID)
	if err != nil {
		h.logger.Error("Failed to initiate backup", zap.Error(err))
		h.errorJSON(w, "Failed to initiate backup", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]interface{}{"data": backup}, http.StatusAccepted)
}

func (h *DrHandlers) ListBackups(w http.ResponseWriter, r *http.Request) {
	limit, offset := h.getPaginationParams(r)
	backups, total, err := h.service.ListBackups(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list backups", zap.Error(err))
		h.errorJSON(w, "Failed to list backups", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]interface{}{
		"data":   backups,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}, http.StatusOK)
}

func (h *DrHandlers) GetBackup(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorJSON(w, "Invalid backup ID", http.StatusBadRequest)
		return
	}

	backup, err := h.service.GetBackup(r.Context(), id)
	if err != nil {
		h.errorJSON(w, "Backup not found", http.StatusNotFound)
		return
	}

	h.json(w, backup, http.StatusOK)
}

func (h *DrHandlers) InitiateRestore(w http.ResponseWriter, r *http.Request) {
	var req InitiateRestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := ValidateRequest(&req); err != nil {
		h.errorJSON(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)
	job, err := h.service.InitiateRestore(r.Context(), req.BackupID, req.TargetEnvironment, userID)
	if err != nil {
		h.logger.Error("Failed to initiate restore", zap.Error(err))
		h.errorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]interface{}{"data": job}, http.StatusAccepted)
}

func (h *DrHandlers) ListRestoreJobs(w http.ResponseWriter, r *http.Request) {
	limit, offset := h.getPaginationParams(r)
	jobs, total, err := h.service.ListRestoreJobs(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list restore jobs", zap.Error(err))
		h.errorJSON(w, "Failed to list restore jobs", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]interface{}{
		"data":   jobs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}, http.StatusOK)
}

func (h *DrHandlers) GetRestoreJob(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorJSON(w, "Invalid restore job ID", http.StatusBadRequest)
		return
	}

	job, err := h.service.GetRestoreJob(r.Context(), id)
	if err != nil {
		h.errorJSON(w, "Restore job not found", http.StatusNotFound)
		return
	}

	h.json(w, job, http.StatusOK)
}

func (h *DrHandlers) VerifyBackup(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorJSON(w, "Invalid backup ID", http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)
	verification, err := h.service.VerifyBackup(r.Context(), id, userID)
	if err != nil {
		h.logger.Error("Failed to verify backup", zap.Error(err))
		h.errorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, verification, http.StatusOK)
}

func (h *DrHandlers) GetRestoreJobLogs(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorJSON(w, "Invalid restore job ID", http.StatusBadRequest)
		return
	}

	job, err := h.service.GetRestoreJob(r.Context(), id)
	if err != nil {
		h.errorJSON(w, "Restore job not found", http.StatusNotFound)
		return
	}

	// Parse logs into an array of log entries
	var logs []map[string]string
	if job.Logs != "" {
		logs = []map[string]string{
			{"level": "INFO", "message": job.Logs, "timestamp": job.StartedAt.Format(time.RFC3339)},
		}
		if job.CompletedAt != nil {
			logs = append(logs, map[string]string{
				"level":     "INFO",
				"message":   "Restore job completed",
				"timestamp": job.CompletedAt.Format(time.RFC3339),
			})
		}
	} else {
		logs = []map[string]string{}
	}

	h.json(w, map[string]interface{}{"data": logs}, http.StatusOK)
}

func (h *DrHandlers) ListStoragePools(w http.ResponseWriter, r *http.Request) {
	pools, err := h.service.ListStoragePools(r.Context())
	if err != nil {
		h.logger.Error("Failed to list storage pools", zap.Error(err))
		h.errorJSON(w, "Failed to list storage pools", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]interface{}{"data": pools}, http.StatusOK)
}

func (h *DrHandlers) CreateDRTest(w http.ResponseWriter, r *http.Request) {
	var req CreateDRTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := ValidateRequest(&req); err != nil {
		h.errorJSON(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r).String()
	test := &models.DRTest{
		Name:            req.Name,
		Description:     req.Description,
		TestType:        req.TestType,
		TargetEnv:       req.TargetEnv,
		SuccessCriteria: req.SuccessCriteria,
		ScheduledAt:     req.ScheduledAt,
		InitiatedBy:     userID,
	}
	created, err := h.service.CreateDRTest(r.Context(), test)
	if err != nil {
		h.logger.Error("failed to create dr test", zap.Error(err))
		h.errorJSON(w, "Failed to create DR test", http.StatusInternalServerError)
		return
	}
	h.json(w, map[string]any{"data": created}, http.StatusCreated)
}

func (h *DrHandlers) ListDRTests(w http.ResponseWriter, r *http.Request) {
	limit, offset := h.getPaginationParams(r)
	tests, total, err := h.service.ListDRTests(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("failed to list dr tests", zap.Error(err))
		h.errorJSON(w, "Failed to list DR tests", http.StatusInternalServerError)
		return
	}
	h.json(w, map[string]any{"data": tests, "total": total, "limit": limit, "offset": offset}, http.StatusOK)
}

func (h *DrHandlers) GetDRTestMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.service.GetDRTestMetrics(r.Context())
	if err != nil {
		h.logger.Error("failed to get dr test metrics", zap.Error(err))
		h.errorJSON(w, "Failed to load metrics", http.StatusInternalServerError)
		return
	}
	h.json(w, map[string]any{"data": metrics}, http.StatusOK)
}

func (h *DrHandlers) GetDRTestSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.GetDRTestSettings(r.Context())
	if err != nil {
		h.logger.Error("failed to get dr test settings", zap.Error(err))
		h.errorJSON(w, "Failed to load settings", http.StatusInternalServerError)
		return
	}
	h.json(w, map[string]any{"data": settings}, http.StatusOK)
}

func (h *DrHandlers) UpdateDRTestSettings(w http.ResponseWriter, r *http.Request) {
	var settings models.DRTestSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		h.errorJSON(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.service.UpdateDRTestSettings(r.Context(), &settings); err != nil {
		h.logger.Error("failed to update dr test settings", zap.Error(err))
		h.errorJSON(w, "Failed to save settings", http.StatusInternalServerError)
		return
	}
	h.json(w, map[string]any{"data": settings}, http.StatusOK)
}

func (h *DrHandlers) GetDRTestDetail(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorJSON(w, "Invalid test ID", http.StatusBadRequest)
		return
	}
	test, steps, logs, err := h.service.GetDRTestDetails(r.Context(), id)
	if err != nil {
		h.errorJSON(w, "Test not found", http.StatusNotFound)
		return
	}
	h.json(w, map[string]any{
		"data":  test,
		"steps": steps,
		"logs":  logs,
	}, http.StatusOK)
}

func (h *DrHandlers) AbortDRTest(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorJSON(w, "Invalid test ID", http.StatusBadRequest)
		return
	}
	if err := h.service.AbortDRTest(r.Context(), id); err != nil {
		h.logger.Error("failed to abort dr test", zap.Error(err))
		h.errorJSON(w, "Failed to abort DR test", http.StatusInternalServerError)
		return
	}
	h.json(w, map[string]string{"status": "aborted"}, http.StatusOK)
}

// Helpers
func (h *DrHandlers) json(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *DrHandlers) errorJSON(w http.ResponseWriter, message string, status int) {
	h.json(w, map[string]string{"error": message}, status)
}

func (h *DrHandlers) getUserID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.UserIDKey); val != nil {
		if userIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(userIDStr); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}

func (h *DrHandlers) getPaginationParams(r *http.Request) (int, int) {
	limit := 10
	offset := 0
	query := r.URL.Query()
	if l := query.Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	if o := query.Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}
	return limit, offset
}
