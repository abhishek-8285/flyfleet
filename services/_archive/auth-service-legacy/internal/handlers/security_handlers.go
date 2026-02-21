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
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, validate origin
	},
}

type SecurityHandlers struct {
	service *service.AuthService
	logger  *zap.Logger
}

func NewSecurityHandlers(s *service.AuthService, logger *zap.Logger) *SecurityHandlers {
	return &SecurityHandlers{service: s, logger: logger}
}

func (h *SecurityHandlers) GetMFADashboard(w http.ResponseWriter, r *http.Request) {
	orgID, _ := uuid.Parse(r.URL.Query().Get("org_id"))
	summary, err := h.service.GetMFADashboardSummary(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (h *SecurityHandlers) GetMFAAdoptionTrend(w http.ResponseWriter, r *http.Request) {
	orgID, _ := uuid.Parse(r.URL.Query().Get("org_id"))
	daysStr := r.URL.Query().Get("days")
	days, _ := strconv.Atoi(daysStr)
	if days == 0 { days = 30 }
	
	trend, err := h.service.GetMFAAdoptionTrend(r.Context(), orgID, days)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trend)
}

func (h *SecurityHandlers) GetMFAMethodDistribution(w http.ResponseWriter, r *http.Request) {
	orgID, _ := uuid.Parse(r.URL.Query().Get("org_id"))
	dist, err := h.service.GetMFAMethodDistribution(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dist)
}

func (h *SecurityHandlers) GetMFAEvents(w http.ResponseWriter, r *http.Request) {
	orgID, _ := uuid.Parse(r.URL.Query().Get("org_id"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 { page = 1 }
	size, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if size == 0 { size = 20 }
	
	events, total, err := h.service.GetMFAEvents(r.Context(), orgID, page, size, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"events": events, "total": total})
}

func (h *SecurityHandlers) GetTopFailureLocations(w http.ResponseWriter, r *http.Request) {
	orgID, _ := uuid.Parse(r.URL.Query().Get("org_id"))
	locs, err := h.service.GetTopFailureLocations(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if locs == nil {
		locs = []map[string]any{}
	}
	w.Header().Set("Content-Type", "application/json")
	// Standardize even for non-paginated lists if requested, but for "Top N" it's usually just data.
	// However, user requested "getMFAFailureLocations: Add meta".
	length := len(locs)
	json.NewEncoder(w).Encode(map[string]any{
		"data": locs,
		"meta": map[string]any{
			"total":       length,
			"page":        1,
			"limit":       length,
			"total_pages": 1,
			"has_next":    false,
			"has_prev":    false,
		},
	})
}

func (h *SecurityHandlers) GetTopFailureUsers(w http.ResponseWriter, r *http.Request) {
	orgID, _ := uuid.Parse(r.URL.Query().Get("org_id"))
	users, err := h.service.GetTopFailureUsers(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if users == nil {
		users = []map[string]any{}
	}
	w.Header().Set("Content-Type", "application/json")
	length := len(users)
	json.NewEncoder(w).Encode(map[string]any{
		"data": users,
		"meta": map[string]any{
			"total":       length,
			"page":        1,
			"limit":       length,
			"total_pages": 1,
			"has_next":    false,
			"has_prev":    false,
		},
	})
}

func (h *SecurityHandlers) GetLockedUsers(w http.ResponseWriter, r *http.Request) {
	orgID, _ := uuid.Parse(r.URL.Query().Get("org_id"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 { page = 1 }
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 { limit = 20 }

	users, total, err := h.service.GetLockedUsers(r.Context(), orgID, page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if users == nil {
		users = []models.User{}
	}
	
	totalPages := (total + limit - 1) / limit
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"data": users,
		"meta": map[string]any{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
		},
	})
}

func (h *SecurityHandlers) UnlockUser(w http.ResponseWriter, r *http.Request) {
	var req struct { UserID string `json:"userId"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	userID, _ := uuid.Parse(req.UserID)
	if err := h.service.UnlockUser(r.Context(), userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *SecurityHandlers) ResetMFA(w http.ResponseWriter, r *http.Request) {
	var req struct { UserID string `json:"userId"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	userID, _ := uuid.Parse(req.UserID)
	if err := h.service.ResetMFA(r.Context(), userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *SecurityHandlers) GetMFAPolicy(w http.ResponseWriter, r *http.Request) {
	orgID, _ := uuid.Parse(r.URL.Query().Get("org_id"))
	policy, err := h.service.GetMFAPolicy(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policy)
}

func (h *SecurityHandlers) UpdateMFAPolicy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Policy models.MFAPolicy `json:"policy"`
		Reason string           `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// In a real app, updaterID would come from context/session
	updaterID, _ := uuid.Parse(r.Header.Get("X-Admin-ID"))
	if updaterID == uuid.Nil {
		// Fallback for demo/testing
		updaterID = req.Policy.OrgID 
	}

	if err := h.service.UpdateMFAPolicy(r.Context(), &req.Policy, updaterID, req.Reason); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req.Policy)
}

func (h *SecurityHandlers) GetMFAPolicyHistory(w http.ResponseWriter, r *http.Request) {
	orgID, _ := uuid.Parse(r.URL.Query().Get("org_id"))
	history, err := h.service.GetMFAPolicyHistory(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (h *SecurityHandlers) RollbackMFAPolicy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		HistoryID string `json:"historyId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	hID, _ := uuid.Parse(req.HistoryID)
	updaterID, _ := uuid.Parse(r.Header.Get("X-Admin-ID"))

	if err := h.service.RollbackMFAPolicy(r.Context(), hID, updaterID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *SecurityHandlers) GetMFAOrgOverride(w http.ResponseWriter, r *http.Request) {
	orgID, _ := uuid.Parse(r.URL.Query().Get("org_id"))
	override, err := h.service.GetMFAOrgOverride(r.Context(), orgID)
	if err != nil {
		// Return empty override instead of error to facilitate UI
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.MFAOrgOverride{OrgID: orgID, Enabled: false})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(override)
}

func (h *SecurityHandlers) UpdateMFAOrgOverride(w http.ResponseWriter, r *http.Request) {
	var override models.MFAOrgOverride
	if err := json.NewDecoder(r.Body).Decode(&override); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	updaterID, _ := uuid.Parse(r.Header.Get("X-Admin-ID"))
	if err := h.service.UpdateMFAOrgOverride(r.Context(), &override, updaterID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(override)
}

func (h *SecurityHandlers) GetMFAExceptions(w http.ResponseWriter, r *http.Request) {
	orgID, _ := uuid.Parse(r.URL.Query().Get("org_id"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 { page = 1 }
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 { limit = 20 }

	excs, total, err := h.service.GetMFAExceptions(r.Context(), orgID, page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if excs == nil {
		excs = []models.MFAException{}
	}
	
	totalPages := (total + limit - 1) / limit
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"data": excs,
		"meta": map[string]any{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
		},
	})
}

func (h *SecurityHandlers) AddMFAException(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"userId"`
		OrgID  string `json:"orgId"`
		Reason string `json:"reason"`
		Days   int    `json:"days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	uid, _ := uuid.Parse(req.UserID)
	oid, _ := uuid.Parse(req.OrgID)
	if err := h.service.AddMFAException(r.Context(), uid, oid, uuid.Nil, req.Reason, req.Days); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *SecurityHandlers) DeleteMFAException(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	idStr := pathParts[len(pathParts)-1]
	id, _ := uuid.Parse(idStr)
	if err := h.service.DeleteMFAException(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SecurityHandlers) StreamMFAEvents(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade to websocket", zap.Error(err))
		return
	}
	defer conn.Close()

	orgID, _ := uuid.Parse(r.URL.Query().Get("org_id"))
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// In a real app, this would be reactive (NATS/Redis PubSub)
			// Mocking real-time feed by fetching last 5 events
			events, _, err := h.service.GetMFAEvents(r.Context(), orgID, 1, 5, nil)
			if err != nil { continue }
			if err := conn.WriteJSON(events); err != nil { return }
		case <-r.Context().Done():
			return
		}
	}
}

