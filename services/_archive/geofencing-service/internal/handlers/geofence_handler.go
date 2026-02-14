package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/geofencing-service/internal/models"
	"github.com/flyfleet/geofencing-service/internal/service"
)

type GeofenceHandler struct {
	service *service.GeofenceService
	logger  *zap.Logger
}

func NewGeofenceHandler(svc *service.GeofenceService, logger *zap.Logger) *GeofenceHandler {
	return &GeofenceHandler{
		service: svc,
		logger:  logger,
	}
}

// CreateGeofence handles creating a new geofence
func (h *GeofenceHandler) CreateGeofence(c *gin.Context) {
	var req models.CreateGeofenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body for creating geofence", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Get org_id and user_id from context (assumes JWT middleware has set these)
	orgIDStr, exists := c.Get("org_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization ID not found in context"})
		return
	}
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	orgID, ok := orgIDStr.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid organization ID format"})
		return
	}
	userID, ok := userIDStr.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	ctx := c.Request.Context()
	geofence, err := h.service.Create(ctx, orgID, userID, &req)
	if err != nil {
		h.logger.Error("Failed to create geofence", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, geofence)
}

// GetGeofence retrieves a geofence by ID
func (h *GeofenceHandler) GetGeofence(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("Invalid geofence ID", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid geofence ID"})
		return
	}

	ctx := c.Request.Context()
	geofence, err := h.service.GetByID(ctx, id)
	if err != nil {
		h.logger.Error("Failed to get geofence", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Geofence not found"})
		return
	}

	c.JSON(http.StatusOK, geofence)
}

// ListGeofences retrieves geofences with pagination and filtering
func (h *GeofenceHandler) ListGeofences(c *gin.Context) {
	var req models.ListGeofencesRequest
	
	// Bind query parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	req.Page = page
	
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}
	req.Limit = limit
	
	req.Search = c.Query("search")
	req.Type = c.Query("type")
	
	activeStr := c.Query("active")
	if activeStr != "" {
		activeBool := activeStr == "true"
		req.Active = &activeBool
	}
	
	req.Sort = c.Query("sort")

	// Get org_id from context
	orgIDStr, exists := c.Get("org_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization ID not found in context"})
		return
	}
	
	orgID, ok := orgIDStr.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid organization ID format"})
		return
	}

	ctx := c.Request.Context()
	response, err := h.service.List(ctx, orgID, &req)
	if err != nil {
		h.logger.Error("Failed to list geofences", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateGeofence updates an existing geofence
func (h *GeofenceHandler) UpdateGeofence(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("Invalid geofence ID", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid geofence ID"})
		return
	}

	var req models.UpdateGeofenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body for updating geofence", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	updatedGeofence, err := h.service.Update(ctx, id, &req)
	if err != nil {
		h.logger.Error("Failed to update geofence", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedGeofence)
}

// DeleteGeofence deletes a geofence (marks as inactive)
func (h *GeofenceHandler) DeleteGeofence(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("Invalid geofence ID", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid geofence ID"})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("Failed to delete geofence", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Geofence deleted successfully"})
}

// GetGeofenceEvents retrieves geofence events with filtering
func (h *GeofenceHandler) GetGeofenceEvents(c *gin.Context) {
	var filter models.GeofenceEventFilter
	
	// Bind query parameters
	geofenceIDStr := c.Query("geofence_id")
	if geofenceIDStr != "" {
		geofenceID, err := uuid.Parse(geofenceIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid geofence ID"})
			return
		}
		filter.GeofenceID = geofenceID
	}
	
	vehicleIDStr := c.Query("vehicle_id")
	if vehicleIDStr != "" {
		vehicleID, err := uuid.Parse(vehicleIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle ID"})
			return
		}
		filter.VehicleID = vehicleID
	}
	
	filter.EventType = c.Query("event_type")
	
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	filter.Page = page
	
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}
	filter.Limit = limit

	ctx := c.Request.Context()
	events, total, err := h.service.GetEvents(ctx, &filter)
	if err != nil {
		h.logger.Error("Failed to get geofence events", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"events": events,
		"total":  total,
		"page":   filter.Page,
		"limit":  filter.Limit,
		"has_more": int64((filter.Page * filter.Limit)) < total,
	}

	c.JSON(http.StatusOK, response)
}
