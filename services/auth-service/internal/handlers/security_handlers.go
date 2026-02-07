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
	json.NewEncoder(w).Encode(locs)
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
	json.NewEncoder(w).Encode(users)
}

func (h *SecurityHandlers) GetLockedUsers(w http.ResponseWriter, r *http.Request) {
	orgID, _ := uuid.Parse(r.URL.Query().Get("org_id"))
	users, err := h.service.GetLockedUsers(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
if users == nil {
		users = []models.User{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
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
	var policy models.MFAPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := h.service.UpdateMFAPolicy(r.Context(), &policy); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policy)
}

func (h *SecurityHandlers) GetMFAExceptions(w http.ResponseWriter, r *http.Request) {
	orgID, _ := uuid.Parse(r.URL.Query().Get("org_id"))
	excs, err := h.service.GetMFAExceptions(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(excs)
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

