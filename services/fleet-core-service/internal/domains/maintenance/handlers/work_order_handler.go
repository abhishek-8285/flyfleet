package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core-service/internal/domains/maintenance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/maintenance/service"
)

// WorkOrderHandler handles HTTP requests for work orders
type WorkOrderHandler struct {
	service *service.WorkOrderService
	logger  *zap.Logger
}

// NewWorkOrderHandler creates a new work order handler
func NewWorkOrderHandler(svc *service.WorkOrderService, logger *zap.Logger) *WorkOrderHandler {
	return &WorkOrderHandler{
		service: svc,
		logger:  logger,
	}
}

// ListWorkOrders godoc
// @Summary List work orders
// @Tags WorkOrders
// @Produce json
// @Param status query []string false "Filter by status"
// @Param priority query []string false "Filter by priority"
// @Param search query string false "Search term"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} models.WorkOrderListResponse
// @Router /work-orders [get]
func (h *WorkOrderHandler) ListWorkOrders(c *gin.Context) {
	orgID := getOrgID(c)

	filter := models.WorkOrderFilter{
		OrganizationID: orgID,
		Search:         c.Query("search"),
		SortBy:         c.DefaultQuery("sort_by", "created_at"),
		SortDir:        c.DefaultQuery("sort_dir", "desc"),
	}

	// Parse status filter
	if statuses := c.QueryArray("status"); len(statuses) > 0 {
		for _, s := range statuses {
			filter.Status = append(filter.Status, models.WorkOrderStatus(s))
		}
	}

	// Parse priority filter
	if priorities := c.QueryArray("priority"); len(priorities) > 0 {
		for _, p := range priorities {
			filter.Priority = append(filter.Priority, models.WorkOrderPriority(p))
		}
	}

	// Parse pagination
	filter.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	filter.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.service.ListWorkOrders(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list work orders", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreateWorkOrder godoc
// @Summary Create work order
// @Tags WorkOrders
// @Accept json
// @Produce json
// @Param request body models.CreateWorkOrderRequest true "Create request"
// @Success 201 {object} models.WorkOrder
// @Router /work-orders [post]
func (h *WorkOrderHandler) CreateWorkOrder(c *gin.Context) {
	orgID := getOrgID(c)
	userID := getUserID(c)

	var req models.CreateWorkOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wo, err := h.service.CreateWorkOrder(c.Request.Context(), orgID, userID, &req)
	if err != nil {
		h.logger.Error("failed to create work order", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, wo)
}

// GetWorkOrder godoc
// @Summary Get work order by ID
// @Tags WorkOrders
// @Produce json
// @Param id path string true "Work Order ID"
// @Success 200 {object} models.WorkOrder
// @Router /work-orders/{id} [get]
func (h *WorkOrderHandler) GetWorkOrder(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid work order ID"})
		return
	}

	wo, err := h.service.GetWorkOrder(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get work order", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if wo == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "work order not found"})
		return
	}

	c.JSON(http.StatusOK, wo)
}

// UpdateWorkOrder godoc
// @Summary Update work order
// @Tags WorkOrders
// @Accept json
// @Produce json
// @Param id path string true "Work Order ID"
// @Param request body models.UpdateWorkOrderRequest true "Update request"
// @Success 200 {object} models.WorkOrder
// @Router /work-orders/{id} [put]
func (h *WorkOrderHandler) UpdateWorkOrder(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid work order ID"})
		return
	}

	var req models.UpdateWorkOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wo, err := h.service.UpdateWorkOrder(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to update work order", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if wo == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "work order not found"})
		return
	}

	c.JSON(http.StatusOK, wo)
}

// ChangeStatus godoc
// @Summary Change work order status
// @Tags WorkOrders
// @Accept json
// @Produce json
// @Param id path string true "Work Order ID"
// @Param request body models.ChangeStatusRequest true "Status request"
// @Success 200 {object} map[string]string
// @Router /work-orders/{id}/status [patch]
func (h *WorkOrderHandler) ChangeStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid work order ID"})
		return
	}

	var req models.ChangeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ChangeStatus(c.Request.Context(), id, req.Status); err != nil {
		h.logger.Error("failed to change status", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

// CloseWorkOrder godoc
// @Summary Close work order
// @Tags WorkOrders
// @Produce json
// @Param id path string true "Work Order ID"
// @Success 200 {object} map[string]string
// @Router /work-orders/{id}/close [post]
func (h *WorkOrderHandler) CloseWorkOrder(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid work order ID"})
		return
	}

	if err := h.service.CloseWorkOrder(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to close work order", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "work order closed"})
}

// GetWorkOrderParts godoc
// @Summary Get parts assigned to work order
// @Tags WorkOrders
// @Produce json
// @Param id path string true "Work Order ID"
// @Success 200 {array} models.WorkOrderPart
// @Router /work-orders/{id}/parts [get]
func (h *WorkOrderHandler) GetWorkOrderParts(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid work order ID"})
		return
	}

	wo, err := h.service.GetWorkOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if wo == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "work order not found"})
		return
	}

	c.JSON(http.StatusOK, wo.Parts)
}

// AssignPart godoc
// @Summary Assign part to work order
// @Tags WorkOrders
// @Accept json
// @Produce json
// @Param id path string true "Work Order ID"
// @Param request body models.AssignPartRequest true "Assign request"
// @Success 201 {object} models.WorkOrderPart
// @Router /work-orders/{id}/parts [post]
func (h *WorkOrderHandler) AssignPart(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid work order ID"})
		return
	}

	var req models.AssignPartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	part, err := h.service.AssignPart(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to assign part", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, part)
}

// RemovePart godoc
// @Summary Remove part from work order
// @Tags WorkOrders
// @Produce json
// @Param id path string true "Work Order ID"
// @Param partId path string true "Part ID"
// @Success 200 {object} map[string]string
// @Router /work-orders/{id}/parts/{partId} [delete]
func (h *WorkOrderHandler) RemovePart(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid work order ID"})
		return
	}

	partID, err := uuid.Parse(c.Param("partId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid part ID"})
		return
	}

	if err := h.service.RemovePart(c.Request.Context(), id, partID); err != nil {
		h.logger.Error("failed to remove part", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "part removed"})
}

// GetStats godoc
// @Summary Get work order statistics
// @Tags WorkOrders
// @Produce json
// @Success 200 {object} models.WorkOrderStats
// @Router /work-orders/stats [get]
func (h *WorkOrderHandler) GetStats(c *gin.Context) {
	orgID := getOrgID(c)

	stats, err := h.service.GetStats(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to get stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Helper functions to get org/user from context
func getOrgID(c *gin.Context) uuid.UUID {
	if id, exists := c.Get("org_id"); exists {
		if orgID, ok := id.(uuid.UUID); ok {
			return orgID
		}
	}
	// Default for development
	return uuid.MustParse("00000000-0000-0000-0000-000000000001")
}

func getUserID(c *gin.Context) uuid.UUID {
	if id, exists := c.Get("user_id"); exists {
		if userID, ok := id.(uuid.UUID); ok {
			return userID
		}
	}
	// Default for development
	return uuid.MustParse("00000000-0000-0000-0000-000000000001")
}
