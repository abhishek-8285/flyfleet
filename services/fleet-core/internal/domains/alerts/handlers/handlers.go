package handlers

import (
	"net/http"

	"github.com/flyfleet/fleet-core-service/internal/domains/alerts/dto"
	"github.com/flyfleet/fleet-core-service/internal/domains/alerts/repository"
	"github.com/flyfleet/fleet-core-service/internal/domains/alerts/service"
	"github.com/flyfleet/fleet-core-service/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AlertsHandlers struct {
	service   *service.AlertsService
	repo      repository.Repository
	logger    *zap.Logger
	validator *validator.Validate
}

func NewAlertsHandlers(service *service.AlertsService, repo repository.Repository, logger *zap.Logger) *AlertsHandlers {
	return &AlertsHandlers{
		service:   service,
		repo:      repo,
		logger:    logger,
		validator: validator.New(),
	}
}

// getOrgID extracts organization ID from context and returns error if missing
func (h *AlertsHandlers) getOrgID(c *gin.Context) uuid.UUID {
	orgID := middleware.GetOrgID(c.Request.Context())
	if orgID == uuid.Nil {
		h.logger.Warn("Missing org_id in context - request likely bypassed authentication middleware",
			zap.String("path", c.Request.URL.Path))
	}
	return orgID
}

// requireOrgAccess checks that the resource belongs to the user's organization
func (h *AlertsHandlers) requireOrgAccess(c *gin.Context, resourceOrgID uuid.UUID) bool {
	orgID := h.getOrgID(c)
	if orgID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization ID required"})
		return false
	}
	if resourceOrgID != orgID {
		h.logger.Warn("unauthorized cross-org access attempt",
			zap.String("requested_org", orgID.String()),
			zap.String("resource_org", resourceOrgID.String()),
			zap.String("path", c.Request.URL.Path))
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
		return false
	}
	return true
}

// TriggerAlert handles POST /v1/alerts/trigger
func (h *AlertsHandlers) TriggerAlert(c *gin.Context) {
	orgID := h.getOrgID(c)
	if orgID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization context required"})
		return
	}

	var req dto.AlertRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate DTO
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed: " + err.Error()})
		return
	}

	// Convert DTO to model with org_id injected
	modelReq := req.ToModel(orgID)

	if err := h.service.TriggerAlert(c.Request.Context(), modelReq); err != nil {
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

	orgID := h.getOrgID(c)
	if orgID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization ID required"})
		return
	}

	prefs, err := h.repo.GetUserPreferences(c.Request.Context(), userID, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch preferences"})
		return
	}

	// Convert to DTOs
	response := make([]dto.UserPreferenceResponseDTO, len(prefs))
	for i, p := range prefs {
		response[i] = dto.ToUserPreferenceResponseDTO(p)
	}

	c.JSON(http.StatusOK, response)
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

	orgID := h.getOrgID(c)
	if orgID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization ID required"})
		return
	}

	var req dto.UserPreferenceUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed: " + err.Error()})
		return
	}

	// Convert DTO to model
	pref := req.ToModel(userID, orgID)

	if err := h.repo.UpsertUserPreference(c.Request.Context(), &pref); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update preference"})
		return
	}

	// Return response DTO
	c.JSON(http.StatusOK, dto.ToUserPreferenceResponseDTO(pref))
}

// CreateTemplate handles POST /v1/alerts/templates
func (h *AlertsHandlers) CreateTemplate(c *gin.Context) {
	orgID := h.getOrgID(c)
	if orgID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization ID required"})
		return
	}

	var req dto.CreateTemplateRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed: " + err.Error()})
		return
	}

	// Convert DTO to model
	tmpl := req.ToModel(orgID)

	if err := h.repo.CreateTemplate(c.Request.Context(), &tmpl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
		return
	}

	// Return response DTO
	c.JSON(http.StatusCreated, dto.ToTemplateResponseDTO(tmpl))
}
