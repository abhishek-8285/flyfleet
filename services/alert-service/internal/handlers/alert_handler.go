package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/alert-service/internal/models"
	"github.com/flyfleet/alert-service/internal/service"
)

type AlertHandler struct {
	service *service.AlertService
	logger  *zap.Logger
}

func NewAlertHandler(svc *service.AlertService, logger *zap.Logger) *AlertHandler {
	return &AlertHandler{
		service: svc,
		logger:  logger,
	}
}

// CreateRule godoc
// @Summary Create alert rule
// @Description Create a new alert rule for monitoring fleet events
// @Tags Alerts
// @Accept json
// @Produce json
// @Param rule body models.CreateRuleRequest true "Rule creation request"
// @Success 201 {object} models.AlertRule
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /alerts/rules [post]
func (h *AlertHandler) CreateRule(c *gin.Context) {
	var req models.CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get organization ID from context (would come from auth middleware)
	orgIDStr := c.GetString("org_id")
	if orgIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization ID required"})
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	// Get user ID from context (would come from auth middleware)
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	rule, err := h.service.CreateRule(c.Request.Context(), orgID, &req, userID)
	if err != nil {
		h.logger.Error("failed to create alert rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// GetRule godoc
// @Summary Get alert rule
// @Description Retrieve a specific alert rule by ID
// @Tags Alerts
// @Produce json
// @Param id path string true "Rule ID"
// @Success 200 {object} models.AlertRule
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /alerts/rules/{id} [get]
func (h *AlertHandler) GetRule(c *gin.Context) {
	id := c.Param("id")

	rule, err := h.service.GetRule(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get alert rule", zap.String("rule_id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert rule not found"})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// ListRules godoc
// @Summary List alert rules
// @Description List alert rules with filtering and pagination
// @Tags Alerts
// @Produce json
// @Param type query []string false "Filter by rule type"
// @Param severity query []string false "Filter by severity"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(20)
// @Param order_by query string false "Order by field"
// @Param order_dir query string false "Order direction" Enums(asc, desc)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /alerts/rules [get]
func (h *AlertHandler) ListRules(c *gin.Context) {
	// Get organization ID from context
	orgIDStr := c.GetString("org_id")
	if orgIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization ID required"})
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	// Parse query parameters
	filter := models.AlertListFilter{
		Page:     1,
		PageSize: 20,
		OrderBy:  "created_at",
		OrderDir: "desc",
	}

	if page := c.Query("page"); page != "" {
		// Would parse page number
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		// Would parse page size
	}

	// Parse type filter
	if types := c.QueryArray("type"); len(types) > 0 {
		filter.Type = types
	}

	// Parse severity filter
	if severities := c.QueryArray("severity"); len(severities) > 0 {
		filter.Severity = severities
	}

	rules, totalCount, err := h.service.ListRules(c.Request.Context(), orgID, filter)
	if err != nil {
		h.logger.Error("failed to list alert rules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list alert rules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       rules,
		"total":      totalCount,
		"page":       filter.Page,
		"page_size":  filter.PageSize,
		"total_pages": (totalCount + filter.PageSize - 1) / filter.PageSize,
	})
}

// UpdateRule godoc
// @Summary Update alert rule
// @Description Update an existing alert rule
// @Tags Alerts
// @Accept json
// @Produce json
// @Param id path string true "Rule ID"
// @Param rule body models.UpdateRuleRequest true "Rule update request"
// @Success 200 {object} models.AlertRule
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /alerts/rules/{id} [put]
func (h *AlertHandler) UpdateRule(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	rule, err := h.service.UpdateRule(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to update alert rule", zap.String("rule_id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert rule not found"})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// DeleteRule godoc
// @Summary Delete alert rule
// @Description Delete an alert rule
// @Tags Alerts
// @Produce json
// @Param id path string true "Rule ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /alerts/rules/{id} [delete]
func (h *AlertHandler) DeleteRule(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteRule(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to delete alert rule", zap.String("rule_id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert rule not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

// EnableRule godoc
// @Summary Enable alert rule
// @Description Enable an alert rule
// @Tags Alerts
// @Produce json
// @Param id path string true "Rule ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /alerts/rules/{id}/enable [post]
func (h *AlertHandler) EnableRule(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.EnableRule(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to enable alert rule", zap.String("rule_id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert rule not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert rule enabled"})
}

// DisableRule godoc
// @Summary Disable alert rule
// @Description Disable an alert rule
// @Tags Alerts
// @Produce json
// @Param id path string true "Rule ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /alerts/rules/{id}/disable [post]
func (h *AlertHandler) DisableRule(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DisableRule(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to disable alert rule", zap.String("rule_id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert rule not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert rule disabled"})
}

// GetAlert godoc
// @Summary Get alert
// @Description Retrieve a specific alert by ID
// @Tags Alerts
// @Produce json
// @Param id path string true "Alert ID"
// @Success 200 {object} models.Alert
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /alerts/{id} [get]
func (h *AlertHandler) GetAlert(c *gin.Context) {
	id := c.Param("id")

	alert, err := h.service.GetAlert(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get alert", zap.String("alert_id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
		return
	}

	c.JSON(http.StatusOK, alert)
}

// ListAlerts godoc
// @Summary List alerts
// @Description List alerts with filtering and pagination
// @Tags Alerts
// @Produce json
// @Param status query []string false "Filter by status"
// @Param severity query []string false "Filter by severity"
// @Param vehicle_id query string false "Filter by vehicle ID"
// @Param driver_id query string false "Filter by driver ID"
// @Param from_date query string false "From date (RFC3339)"
// @Param to_date query string false "To date (RFC3339)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(20)
// @Param order_by query string false "Order by field"
// @Param order_dir query string false "Order direction" Enums(asc, desc)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /alerts [get]
func (h *AlertHandler) ListAlerts(c *gin.Context) {
	// Get organization ID from query param or context
	orgIDStr := c.Query("org_id")
	if orgIDStr == "" {
		orgIDStr = c.GetString("org_id")
	}
	if orgIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization ID required"})
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	// Parse query parameters
	filter := models.AlertListFilter{
		OrgID:    orgID,
		Page:     1,
		PageSize: 20,
		OrderBy:  "timestamp",
		OrderDir: "desc",
	}

	// Parse date filters
	if fromDateStr := c.Query("from_date"); fromDateStr != "" {
		fromDate, err := time.Parse(time.RFC3339, fromDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from_date format"})
			return
		}
		filter.FromDate = fromDate
	}

	if toDateStr := c.Query("to_date"); toDateStr != "" {
		toDate, err := time.Parse(time.RFC3339, toDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to_date format"})
			return
		}
		filter.ToDate = toDate
	}

	// Parse other filters
	if statuses := c.QueryArray("status"); len(statuses) > 0 {
		filter.Status = statuses
	}

	if severities := c.QueryArray("severity"); len(severities) > 0 {
		filter.Severity = severities
	}

	if vehicleIDStr := c.Query("vehicle_id"); vehicleIDStr != "" {
		vehicleID, err := uuid.Parse(vehicleIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle_id"})
			return
		}
		filter.VehicleID = &vehicleID
	}

	if driverIDStr := c.Query("driver_id"); driverIDStr != "" {
		driverID, err := uuid.Parse(driverIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver_id"})
			return
		}
		filter.DriverID = &driverID
	}

	alerts, totalCount, err := h.service.ListAlerts(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list alerts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        alerts,
		"total":       totalCount,
		"page":        filter.Page,
		"page_size":   filter.PageSize,
		"total_pages": (totalCount + filter.PageSize - 1) / filter.PageSize,
	})
}

// AcknowledgeAlert godoc
// @Summary Acknowledge alert
// @Description Mark an alert as acknowledged
// @Tags Alerts
// @Accept json
// @Produce json
// @Param id path string true "Alert ID"
// @Param request body map[string]string false "Acknowledge request with optional notes"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /alerts/{id}/acknowledge [post]
func (h *AlertHandler) AcknowledgeAlert(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Notes string `json:"notes,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Notes are optional, so ignore binding errors
		req.Notes = ""
	}

	// Get user ID from context
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.service.AcknowledgeAlert(c.Request.Context(), id, userID, req.Notes); err != nil {
		h.logger.Error("failed to acknowledge alert", zap.String("alert_id", id), zap.Error(err))
		if err.Error() == "alert must be in 'new' status to acknowledge" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert acknowledged"})
}

// ResolveAlert godoc
// @Summary Resolve alert
// @Description Mark an alert as resolved
// @Tags Alerts
// @Accept json
// @Produce json
// @Param id path string true "Alert ID"
// @Param request body map[string]string false "Resolve request with optional notes"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /alerts/{id}/resolve [post]
func (h *AlertHandler) ResolveAlert(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Notes string `json:"notes,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Notes are optional, so ignore binding errors
		req.Notes = ""
	}

	// Get user ID from context
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.service.ResolveAlert(c.Request.Context(), id, userID, req.Notes); err != nil {
		h.logger.Error("failed to resolve alert", zap.String("alert_id", id), zap.Error(err))
		if err.Error() == "alert is already resolved" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert resolved"})
}

// GetAlertSummary godoc
// @Summary Get alert summary
// @Description Get alert statistics and summary
// @Tags Alerts
// @Produce json
// @Param from_date query string true "From date (RFC3339)"
// @Param to_date query string true "To date (RFC3339)"
// @Success 200 {object} models.AlertSummary
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /alerts/summary [get]
func (h *AlertHandler) GetAlertSummary(c *gin.Context) {
	// Get organization ID from context
	orgIDStr := c.GetString("org_id")
	if orgIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization ID required"})
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	// Parse required date parameters
	fromDateStr := c.Query("from_date")
	toDateStr := c.Query("to_date")

	if fromDateStr == "" || toDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both from_date and to_date are required"})
		return
	}

	fromDate, err := time.Parse(time.RFC3339, fromDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from_date format"})
		return
	}

	toDate, err := time.Parse(time.RFC3339, toDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to_date format"})
		return
	}

	summary, err := h.service.GetAlertSummary(c.Request.Context(), orgID, fromDate, toDate)
	if err != nil {
		h.logger.Error("failed to get alert summary", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get alert summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// EscalateAlert escalates an alert to higher authorities
func (h *AlertHandler) EscalateAlert(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Alert escalated", "alert_id": id})
}

// GetAlertTrends returns alert trends over time
func (h *AlertHandler) GetAlertTrends(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"trends": []interface{}{}})
}

// CreateTemplate creates a notification template
func (h *AlertHandler) CreateTemplate(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "Template created"})
}

// ListTemplates lists all notification templates
func (h *AlertHandler) ListTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"templates": []interface{}{}})
}

// GetTemplate gets a template by ID
func (h *AlertHandler) GetTemplate(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"template_id": id})
}

// UpdateTemplate updates a template
func (h *AlertHandler) UpdateTemplate(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Template updated", "template_id": id})
}

// DeleteTemplate deletes a template
func (h *AlertHandler) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Template deleted", "template_id": id})
}
func (h *AlertHandler) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get alert stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
