package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/maintenance-service/internal/models"
	"github.com/flyfleet/maintenance-service/internal/repository"
)

// InventoryHandler handles HTTP requests for inventory
type InventoryHandler struct {
	repo   *repository.InventoryRepository
	logger *zap.Logger
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(repo *repository.InventoryRepository, logger *zap.Logger) *InventoryHandler {
	return &InventoryHandler{
		repo:   repo,
		logger: logger,
	}
}

// ListInventory godoc
// @Summary List inventory parts
// @Tags Inventory
// @Produce json
// @Param search query string false "Search term"
// @Param stock_location query string false "Stock location"
// @Param low_stock_only query bool false "Show only low stock items"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} models.InventoryListResponse
// @Router /inventory [get]
func (h *InventoryHandler) ListInventory(c *gin.Context) {
	orgID := getOrgID(c)

	filter := models.InventoryFilter{
		OrganizationID: orgID,
		Search:         c.Query("search"),
		StockLocation:  c.Query("stock_location"),
	}

	filter.LowStockOnly, _ = strconv.ParseBool(c.Query("low_stock_only"))
	filter.AvailableOnly, _ = strconv.ParseBool(c.Query("available_only"))
	filter.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	filter.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	parts, total, err := h.repo.List(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list inventory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.InventoryListResponse{
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
		Parts:    parts,
	})
}

// CreatePart godoc
// @Summary Create inventory part
// @Tags Inventory
// @Accept json
// @Produce json
// @Param request body models.CreateInventoryPartRequest true "Create request"
// @Success 201 {object} models.InventoryPart
// @Router /inventory [post]
func (h *InventoryHandler) CreatePart(c *gin.Context) {
	orgID := getOrgID(c)

	var req models.CreateInventoryPartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	part := &models.InventoryPart{
		OrganizationID:    orgID,
		PartNumber:        req.PartNumber,
		Name:              req.Name,
		Description:       req.Description,
		StockLocation:     req.StockLocation,
		QuantityAvailable: req.QuantityAvailable,
		ReorderLevel:      req.ReorderLevel,
		UnitCost:          req.UnitCost,
	}

	if err := h.repo.Create(c.Request.Context(), part); err != nil {
		h.logger.Error("failed to create part", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, part)
}

// GetPart godoc
// @Summary Get inventory part by ID
// @Tags Inventory
// @Produce json
// @Param id path string true "Part ID"
// @Success 200 {object} models.InventoryPart
// @Router /inventory/{id} [get]
func (h *InventoryHandler) GetPart(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid part ID"})
		return
	}

	part, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get part", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if part == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "part not found"})
		return
	}

	c.JSON(http.StatusOK, part)
}

// UpdatePart godoc
// @Summary Update inventory part
// @Tags Inventory
// @Accept json
// @Produce json
// @Param id path string true "Part ID"
// @Param request body models.UpdateInventoryPartRequest true "Update request"
// @Success 200 {object} models.InventoryPart
// @Router /inventory/{id} [put]
func (h *InventoryHandler) UpdatePart(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid part ID"})
		return
	}

	part, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if part == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "part not found"})
		return
	}

	var req models.UpdateInventoryPartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Apply updates
	if req.PartNumber != nil {
		part.PartNumber = *req.PartNumber
	}
	if req.Name != nil {
		part.Name = *req.Name
	}
	if req.Description != nil {
		part.Description = *req.Description
	}
	if req.StockLocation != nil {
		part.StockLocation = *req.StockLocation
	}
	if req.QuantityAvailable != nil {
		part.QuantityAvailable = *req.QuantityAvailable
	}
	if req.ReorderLevel != nil {
		part.ReorderLevel = *req.ReorderLevel
	}
	if req.UnitCost != nil {
		part.UnitCost = *req.UnitCost
	}

	if err := h.repo.Update(c.Request.Context(), part); err != nil {
		h.logger.Error("failed to update part", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, part)
}

// DeletePart godoc
// @Summary Delete inventory part
// @Tags Inventory
// @Produce json
// @Param id path string true "Part ID"
// @Success 204
// @Router /inventory/{id} [delete]
func (h *InventoryHandler) DeletePart(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid part ID"})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to delete part", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
