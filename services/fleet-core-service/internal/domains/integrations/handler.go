package integrations

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handler struct {
	service Service
	logger  *zap.Logger
}

func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	integrations := rg.Group("/integrations")
	{
		integrations.POST("", h.CreateIntegration)
		integrations.GET("", h.ListIntegrations)
		integrations.GET("/:id", h.GetIntegration)
		integrations.DELETE("/:id", h.DeleteIntegration)
	}
}

func (h *Handler) CreateIntegration(c *gin.Context) {
	var req CreateIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	integration, err := h.service.CreateIntegration(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create integration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create integration"})
		return
	}

	c.JSON(http.StatusCreated, integration)
}

func (h *Handler) GetIntegration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	integration, err := h.service.GetIntegration(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	c.JSON(http.StatusOK, integration)
}

func (h *Handler) ListIntegrations(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	integrations, err := h.service.ListIntegrations(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list integrations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list integrations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": integrations})
}

func (h *Handler) DeleteIntegration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := h.service.DeleteIntegration(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete integration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete integration"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
