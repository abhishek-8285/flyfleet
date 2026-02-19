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
	// Calculate pagination metadata
	totalPages := 0
	if size > 0 {
		totalPages = (total + size - 1) / size
	}

	writeJSON(w, map[string]interface{}{
		"data": sessions,
		"meta": map[string]interface{}{
			"total":       total,
			"page":        page,
			"limit":       size,
			"total_pages": totalPages,
			"has_next":    (page * size) < total,
			"has_prev":    page > 1,
		},
	}, http.StatusOK)
}

func (h *SecurityExtHandlers) RevokeSession(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := r.PathValue("id")
	if sessionIDStr == "" {
		// Fallback for query param or older router behavior if needed, 
		// but sticking to standard path param as per request.
		sessionIDStr = r.URL.Query().Get("id")
	}

	sessionID, err := uuid.Parse(sessionIDStr)
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
		zap.String("session_id", sessionIDStr),
		zap.String("admin_id", adminID.String()),
	)
	w.WriteHeader(http.StatusNoContent)
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


func (h *SecurityExtHandlers) GetSecurityTelemetry(w http.ResponseWriter, r *http.Request) {
telemetry, err := h.service.GetSecurityTelemetry(r.Context())
if err != nil {
h.logger.Error("Failed to fetch security telemetry", zap.Error(err))
http.Error(w, "Failed to fetch security telemetry", http.StatusInternalServerError)
return
}

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]interface{}{
"data": telemetry,
})
}

// GetSecurityMetrics provides comprehensive security metrics dashboard
func (h *SecurityExtHandlers) GetSecurityMetrics(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("org_id")
	var orgID uuid.UUID
	if orgIDStr != "" {
		var err error
		orgID, err = uuid.Parse(orgIDStr)
		if err != nil {
			http.Error(w, "Invalid organization ID", http.StatusBadRequest)
			return
		}
	}

	metrics, err := h.service.GetSecurityMetrics(r.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to fetch security metrics", zap.Error(err))
		http.Error(w, "Failed to fetch security metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// GetSecurityPostureScore calculates and returns the organization's security posture score
func (h *SecurityExtHandlers) GetSecurityPostureScore(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("org_id")
	var orgID uuid.UUID
	if orgIDStr != "" {
		var err error
		orgID, err = uuid.Parse(orgIDStr)
		if err != nil {
			http.Error(w, "Invalid organization ID", http.StatusBadRequest)
			return
		}
	}

	score, err := h.service.GetSecurityPostureScore(r.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to calculate security posture score", zap.Error(err))
		http.Error(w, "Failed to calculate security posture score", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"score": score,
		"max_score": 100,
		"description": "Security posture score based on various security factors",
	})
}

// ListSecurityEvents provides filtered security event listing
func (h *SecurityExtHandlers) ListSecurityEvents(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("org_id")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	eventType := r.URL.Query().Get("type")
	severity := r.URL.Query().Get("severity")

	var orgID uuid.UUID
	if orgIDStr != "" {
		var err error
		orgID, err = uuid.Parse(orgIDStr)
		if err != nil {
			http.Error(w, "Invalid organization ID", http.StatusBadRequest)
			return
		}
	}

	events, total, err := h.service.ListSecurityEvents(r.Context(), orgID, page, size, map[string]interface{}{
		"type": eventType,
		"severity": severity,
	})
	if err != nil {
		h.logger.Error("Failed to list security events", zap.Error(err))
		http.Error(w, "Failed to list security events", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"events": events,
		"total":  total,
		"page":   page,
		"size":   size,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSecurityComplianceReport generates compliance reports
func (h *SecurityExtHandlers) GetSecurityComplianceReport(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("org_id")
	reportType := r.URL.Query().Get("type") // e.g., "soc2", "gdpr", "iso27001"
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var orgID uuid.UUID
	if orgIDStr != "" {
		var err error
		orgID, err = uuid.Parse(orgIDStr)
		if err != nil {
			http.Error(w, "Invalid organization ID", http.StatusBadRequest)
			return
		}
	}

	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, _ = time.Parse("2006-01-02", startDateStr)
	}
	if endDateStr != "" {
		endDate, _ = time.Parse("2006-01-02", endDateStr)
	}

	report, err := h.service.GetSecurityComplianceReport(r.Context(), orgID, reportType, startDate, endDate)
	if err != nil {
		h.logger.Error("Failed to generate compliance report", zap.Error(err))
		http.Error(w, "Failed to generate compliance report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// GetVulnerabilityReport provides vulnerability management capabilities
func (h *SecurityExtHandlers) GetVulnerabilityReport(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("org_id")
	var orgID uuid.UUID
	if orgIDStr != "" {
		var err error
		orgID, err = uuid.Parse(orgIDStr)
		if err != nil {
			http.Error(w, "Invalid organization ID", http.StatusBadRequest)
			return
		}
	}

	report, err := h.service.GetVulnerabilityReport(r.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to generate vulnerability report", zap.Error(err))
		http.Error(w, "Failed to generate vulnerability report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// ManageSecuritySettings allows configuration of security settings
func (h *SecurityExtHandlers) ManageSecuritySettings(w http.ResponseWriter, r *http.Request) {
	var settings models.SecuritySettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	orgIDStr := r.URL.Query().Get("org_id")
	var orgID uuid.UUID
	if orgIDStr != "" {
		var err error
		orgID, err = uuid.Parse(orgIDStr)
		if err != nil {
			http.Error(w, "Invalid organization ID", http.StatusBadRequest)
			return
		}
	}

	if err := h.service.UpdateSecuritySettings(r.Context(), orgID, &settings); err != nil {
		h.logger.Error("Failed to update security settings", zap.Error(err))
		http.Error(w, "Failed to update security settings", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Security settings updated successfully"})
}

// GetSecuritySettings retrieves current security settings
func (h *SecurityExtHandlers) GetSecuritySettings(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("org_id")
	var orgID uuid.UUID
	if orgIDStr != "" {
		var err error
		orgID, err = uuid.Parse(orgIDStr)
		if err != nil {
			http.Error(w, "Invalid organization ID", http.StatusBadRequest)
			return
		}
	}

	settings, err := h.service.GetSecuritySettings(r.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to get security settings", zap.Error(err))
		http.Error(w, "Failed to get security settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}
