package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/flyfleet/auth-service/internal/models"
	"github.com/flyfleet/auth-service/internal/service"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SecurityExtHandlers struct {
	service *service.AuthService
	logger  *zap.Logger
}

func NewSecurityExtHandlers(s *service.AuthService, logger *zap.Logger) *SecurityExtHandlers {
	return &SecurityExtHandlers{service: s, logger: logger}
}

// Helper to extract admin ID from request
func getAdminID(r *http.Request) uuid.UUID {
	// Try header first (set by BFF/gateway)
	if id := r.Header.Get("X-User-ID"); id != "" {
		if uid, err := uuid.Parse(id); err == nil {
			return uid
		}
	}
	// Fallback to X-Admin-ID
	if id := r.Header.Get("X-Admin-ID"); id != "" {
		if uid, err := uuid.Parse(id); err == nil {
			return uid
		}
	}
	return uuid.Nil
}

// Helper to extract client IP
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (from reverse proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	// Fall back to RemoteAddr
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

// writeError writes a JSON error response
func writeError(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	})
}

// writeJSON writes a JSON success response
func writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// IP Rules

func (h *SecurityExtHandlers) ListIPRules(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("org_id")
	var orgID uuid.UUID
	if orgIDStr != "" {
		parsed, err := uuid.Parse(orgIDStr)
		if err != nil {
			writeError(w, service.ErrInvalidOrgID, http.StatusBadRequest)
			return
		}
		orgID = parsed
	} else {
		orgID = uuid.Nil
	}

	rules, err := h.service.ListIPRules(r.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to list IP rules", zap.Error(err), zap.String("org_id", orgIDStr))
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	if rules == nil {
		rules = []models.IPRule{}
	}
	writeJSON(w, rules, http.StatusOK)
}

func (h *SecurityExtHandlers) CreateIPRule(w http.ResponseWriter, r *http.Request) {
	var rule models.IPRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	// Get org from query if not in body
	if rule.OrgID == uuid.Nil {
		rule.OrgID, _ = uuid.Parse(r.URL.Query().Get("org_id"))
	}

	adminID := getAdminID(r)
	clientIP := getClientIP(r)
	rule.AddedBy = adminID

	if err := h.service.CreateIPRule(r.Context(), &rule, adminID, clientIP); err != nil {
		h.logger.Warn("Failed to create IP rule", 
			zap.Error(err), 
			zap.String("cidr", rule.CIDR),
			zap.String("admin_id", adminID.String()),
		)
		statusCode := http.StatusBadRequest
		if err == service.ErrDuplicateIPRule {
			statusCode = http.StatusConflict
		}
		writeError(w, err, statusCode)
		return
	}

	h.logger.Info("IP rule created", 
		zap.String("id", rule.ID.String()),
		zap.String("cidr", rule.CIDR),
		zap.String("action", rule.Action),
		zap.String("admin_id", adminID.String()),
	)
	writeJSON(w, rule, http.StatusCreated)
}

func (h *SecurityExtHandlers) DeleteIPRule(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	idStr := pathParts[len(pathParts)-1]
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	adminID := getAdminID(r)
	clientIP := getClientIP(r)

	if err := h.service.DeleteIPRule(r.Context(), id, adminID, clientIP); err != nil {
		h.logger.Error("Failed to delete IP rule", zap.Error(err), zap.String("rule_id", idStr))
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	h.logger.Info("IP rule deleted", zap.String("id", idStr), zap.String("admin_id", adminID.String()))
	w.WriteHeader(http.StatusNoContent)
}

// Access Reviews

func (h *SecurityExtHandlers) ListAccessReviews(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("org_id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil || orgID == uuid.Nil {
		writeError(w, service.ErrInvalidOrgID, http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	reviews, total, err := h.service.ListAccessReviews(r.Context(), orgID, page, size)
	if err != nil {
		h.logger.Error("Failed to list access reviews", zap.Error(err))
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	if reviews == nil {
		reviews = []models.AccessReview{}
	}
	writeJSON(w, map[string]interface{}{
		"reviews": reviews,
		"total":   total,
		"page":    page,
		"size":    size,
	}, http.StatusOK)
}

func (h *SecurityExtHandlers) GetAccessReview(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	idStr := pathParts[len(pathParts)-1]
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	review, err := h.service.GetAccessReviewByID(r.Context(), id)
	if err != nil {
		writeError(w, err, http.StatusNotFound)
		return
	}

	writeJSON(w, review, http.StatusOK)
}

func (h *SecurityExtHandlers) CreateAccessReview(w http.ResponseWriter, r *http.Request) {
	var review models.AccessReview
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	if review.OrgID == uuid.Nil {
		review.OrgID, _ = uuid.Parse(r.URL.Query().Get("org_id"))
	}

	adminID := getAdminID(r)
	clientIP := getClientIP(r)
	review.CreatedBy = adminID

	if err := h.service.CreateAccessReview(r.Context(), &review, adminID, clientIP); err != nil {
		h.logger.Warn("Failed to create access review", zap.Error(err))
		writeError(w, err, http.StatusBadRequest)
		return
	}

	h.logger.Info("Access review created", 
		zap.String("id", review.ID.String()),
		zap.String("title", review.Title),
		zap.String("admin_id", adminID.String()),
	)
	writeJSON(w, review, http.StatusCreated)
}

func (h *SecurityExtHandlers) UpdateAccessReviewStatus(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	idStr := pathParts[len(pathParts)-2] // /reviews/{id}/status
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	adminID := getAdminID(r)
	clientIP := getClientIP(r)

	if err := h.service.UpdateAccessReviewStatus(r.Context(), id, req.Status, adminID, clientIP); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	writeJSON(w, map[string]string{"status": "updated"}, http.StatusOK)
}

func (h *SecurityExtHandlers) UpdateReviewItem(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ItemID     string `json:"item_id"`
		Status     string `json:"status"`
		ReviewerID string `json:"reviewer_id"`
		Notes      string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	itemID, err := uuid.Parse(req.ItemID)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	reviewerID := getAdminID(r)
	if req.ReviewerID != "" {
		if parsed, err := uuid.Parse(req.ReviewerID); err == nil {
			reviewerID = parsed
		}
	}

	clientIP := getClientIP(r)

	if err := h.service.UpdateReviewItem(r.Context(), itemID, req.Status, reviewerID, req.Notes, clientIP); err != nil {
		h.logger.Warn("Failed to update review item", zap.Error(err))
		writeError(w, err, http.StatusBadRequest)
		return
	}

	h.logger.Info("Review item updated", 
		zap.String("item_id", req.ItemID),
		zap.String("status", req.Status),
		zap.String("reviewer_id", reviewerID.String()),
	)
	writeJSON(w, map[string]string{"status": "updated"}, http.StatusOK)
}

// Sessions

func (h *SecurityExtHandlers) ListSessions(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("org_id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil || orgID == uuid.Nil {
		writeError(w, service.ErrInvalidOrgID, http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	sessions, total, err := h.service.ListAllSessions(r.Context(), orgID, page, size)
	if err != nil {
		h.logger.Error("Failed to list sessions", zap.Error(err))
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	if sessions == nil {
		sessions = []models.AuthToken{}
	}
	writeJSON(w, map[string]interface{}{
		"sessions": sessions,
		"total":    total,
		"page":     page,
		"size":     size,
	}, http.StatusOK)
}

func (h *SecurityExtHandlers) RevokeSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		writeError(w, service.ErrInvalidSessionID, http.StatusBadRequest)
		return
	}

	adminID := getAdminID(r)
	clientIP := getClientIP(r)

	if err := h.service.RevokeSession(r.Context(), sessionID, adminID, clientIP); err != nil {
		h.logger.Error("Failed to revoke session", zap.Error(err))
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	h.logger.Info("Session revoked", 
		zap.String("session_id", req.SessionID),
		zap.String("admin_id", adminID.String()),
	)
	writeJSON(w, map[string]string{"status": "revoked"}, http.StatusOK)
}

func (h *SecurityExtHandlers) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("org_id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil || orgID == uuid.Nil {
		writeError(w, service.ErrInvalidOrgID, http.StatusBadRequest)
		return
	}

	adminID := getAdminID(r)
	clientIP := getClientIP(r)

	if err := h.service.RevokeAllSessionsForOrg(r.Context(), orgID, adminID, clientIP); err != nil {
		h.logger.Error("Failed to revoke all sessions", zap.Error(err))
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	h.logger.Info("All sessions revoked", 
		zap.String("org_id", orgIDStr),
		zap.String("admin_id", adminID.String()),
	)
	writeJSON(w, map[string]string{"status": "all_revoked"}, http.StatusOK)
}

// Login Events

func (h *SecurityExtHandlers) ListLoginEvents(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("org_id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil || orgID == uuid.Nil {
		writeError(w, service.ErrInvalidOrgID, http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	status := r.URL.Query().Get("status")

	events, total, err := h.service.ListLoginEvents(r.Context(), orgID, page, size, status)
	if err != nil {
		h.logger.Error("Failed to list login events", zap.Error(err))
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{
		"events": events,
		"total":  total,
		"page":   page,
		"size":   size,
	}, http.StatusOK)
}

// Threat Alerts

func (h *SecurityExtHandlers) ListThreatAlerts(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("org_id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil || orgID == uuid.Nil {
		writeError(w, service.ErrInvalidOrgID, http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	status := r.URL.Query().Get("status")

	alerts, total, err := h.service.ListThreatAlerts(r.Context(), orgID, page, size, status)
	if err != nil {
		h.logger.Error("Failed to list threat alerts", zap.Error(err))
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{
		"alerts": alerts,
		"total":  total,
		"page":   page,
		"size":   size,
	}, http.StatusOK)
}

func (h *SecurityExtHandlers) UpdateThreatAlertStatus(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	alertIDStr := pathParts[len(pathParts)-2] // /alerts/{id}/status
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateThreatAlertStatus(r.Context(), alertID, req.Status); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	writeJSON(w, map[string]string{"status": "updated"}, http.StatusOK)
}

// MFA Users

func (h *SecurityExtHandlers) ListMFAUsers(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("org_id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil || orgID == uuid.Nil {
		writeError(w, service.ErrInvalidOrgID, http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	users, total, err := h.service.ListMFAUsers(r.Context(), orgID, page, size)
	if err != nil {
		h.logger.Error("Failed to list MFA users", zap.Error(err))
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{
		"users": users,
		"total": total,
		"page":  page,
		"size":  size,
	}, http.StatusOK)
}

