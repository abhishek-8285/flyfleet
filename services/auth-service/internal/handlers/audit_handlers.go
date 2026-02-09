package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flyfleet/auth-service/internal/models"
	"github.com/flyfleet/auth-service/internal/service"
	"github.com/google/uuid"
)

type AuditHandlers struct {
	service *service.AuthService
}

func NewAuditHandlers(s *service.AuthService) *AuditHandlers {
	return &AuditHandlers{service: s}
}

func (h *AuditHandlers) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("org_id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "invalid org_id", http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 { page = 1 }
	size, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if size == 0 { size = 50 }

	filters := make(map[string]any)
	if action := r.URL.Query().Get("action"); action != "" {
		filters["action"] = action
	}
	if severity := r.URL.Query().Get("severity"); severity != "" {
		filters["severity"] = severity
	}
	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		if uid, err := uuid.Parse(userIDStr); err == nil {
			filters["user_id"] = uid
		}
	}

	logs, total, err := h.service.ListAuditLogs(r.Context(), orgID, page, size, filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"logs":  logs,
		"total": total,
	})
}

func (h *AuditHandlers) GetAuditLogByID(w http.ResponseWriter, r *http.Request) {
	// Path: /api/v1/admin/audit/logs/{id}
	pathParts := strings.Split(r.URL.Path, "/")
	idStr := pathParts[len(pathParts)-1]
	
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid audit log id", http.StatusBadRequest)
		return
	}

	log, err := h.service.GetAuditLogByID(r.Context(), id)
	if err != nil {
		http.Error(w, "audit log not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(log)
}

func (h *AuditHandlers) GetAuditSettings(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("org_id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "invalid org_id", http.StatusBadRequest)
		return
	}

	settings, err := h.service.GetAuditSettings(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

func (h *AuditHandlers) UpdateAuditSettings(w http.ResponseWriter, r *http.Request) {
	var settings models.AuditSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateAuditSettings(r.Context(), &settings); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuditHandlers) ListEntityHistory(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("type")
	entityID := r.URL.Query().Get("id")

	if entityType == "" || entityID == "" {
		http.Error(w, "missing type or id", http.StatusBadRequest)
		return
	}

	history, err := h.service.ListEntityAuditHistory(r.Context(), entityType, entityID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (h *AuditHandlers) CreateEvidenceExport(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrgID uuid.UUID `json:"org_id"`
		Module string `json:"module"`
		Start time.Time `json:"date_range_start"`
		End time.Time `json:"date_range_end"`
		Format string `json:"format"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	// Extract requestedBy from context
	requestedBy, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		requestedBy = uuid.Nil
	}

	export, err := h.service.CreateEvidenceExport(r.Context(), req.OrgID, req.Module, req.Start, req.End, req.Format, requestedBy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(export)
}

func (h *AuditHandlers) ListEvidenceExports(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("org_id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "invalid org_id", http.StatusBadRequest)
		return
	}

	exports, err := h.service.ListEvidenceExports(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exports)
}

func (h *AuditHandlers) GetEvidenceExport(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	idStr := pathParts[len(pathParts)-1]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	export, err := h.service.GetEvidenceExportByID(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(export)
}

func (h *AuditHandlers) GetAuditPolicy(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("org_id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "invalid org_id", http.StatusBadRequest)
		return
	}

	policy, err := h.service.GetAuditPolicy(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policy)
}

func (h *AuditHandlers) UpdateAuditPolicy(w http.ResponseWriter, r *http.Request) {
	var policy models.AuditPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateAuditPolicy(r.Context(), &policy); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
