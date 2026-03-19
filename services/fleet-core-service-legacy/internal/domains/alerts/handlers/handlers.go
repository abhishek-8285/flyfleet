package handlers

import (
	"net/http"

	"github.com/flyfleet/fleet-core-service/internal/domains/alerts/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/alerts/repository"
	"github.com/flyfleet/fleet-core-service/internal/domains/alerts/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AlertsHandlers struct {
	service *service.AlertsService
	repo    repository.Repository
}

func NewAlertsHandlers(service *service.AlertsService, repo repository.Repository) *AlertsHandlers {
	return &AlertsHandlers{
		service: service,
		repo:    repo,
	}
}

// TriggerAlert handles POST /v1/alerts/trigger
func (h *AlertsHandlers) TriggerAlert(c *gin.Context) {
	var req models.AlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.TriggerAlert(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to trigger alert"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "Alert processing initiated"})
}

// GetPreferences handles GET /v1/alerts/preferences
func (h *AlertsHandlers) GetPreferences(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid User ID"})
		return
	}

	prefs, err := h.repo.GetUserPreferences(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch preferences"})
		return
	}

	c.JSON(http.StatusOK, prefs)
}

// UpdatePreference handles PUT /v1/alerts/preferences
func (h *AlertsHandlers) UpdatePreference(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid User ID"})
		return
	}

	var pref models.UserPreference
	if err := c.ShouldBindJSON(&pref); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pref.UserID = userID

	if err := h.repo.UpsertUserPreference(c.Request.Context(), &pref); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update preference"})
		return
	}

	c.JSON(http.StatusOK, pref)
}

// CreateTemplate handles POST /v1/alerts/templates
func (h *AlertsHandlers) CreateTemplate(c *gin.Context) {
	var tmpl models.Template
	if err := c.ShouldBindJSON(&tmpl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.CreateTemplate(c.Request.Context(), &tmpl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
		return
	}

	c.JSON(http.StatusCreated, tmpl)
}
