package handlers

import (
	"net/http"

	"github.com/flyfleet/admin-support-service/internal/domains/governance/models"
	"github.com/flyfleet/admin-support-service/internal/domains/governance/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GovernanceHandler struct {
	svc service.GovernanceService
}

func NewGovernanceHandler(svc service.GovernanceService) *GovernanceHandler {
	return &GovernanceHandler{svc: svc}
}

func (h *GovernanceHandler) RegisterRoutes(router *gin.RouterGroup) {
	governance := router.Group("/governance")
	{
		governance.POST("/requests", h.CreateRequest)
		governance.GET("/requests", h.ListRequests)
		governance.GET("/requests/:id", h.GetRequest)
		governance.POST("/requests/:id/approve", h.ApproveRequest)
		governance.POST("/requests/:id/reject", h.RejectRequest)

		governance.GET("/settings", h.GetSettings)
		governance.PUT("/settings", h.UpdateSettings)
	}
}

func (h *GovernanceHandler) CreateRequest(c *gin.Context) {
	orgID, _ := uuid.Parse(c.GetHeader("X-Organization-ID"))
	userID, _ := uuid.Parse(c.GetHeader("X-User-ID"))

	var req models.CreateGovernanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	gReq, err := h.svc.CreateRequest(c.Request.Context(), orgID, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gReq)
}

func (h *GovernanceHandler) GetRequest(c *gin.Context) {
	orgID, _ := uuid.Parse(c.GetHeader("X-Organization-ID"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request id"})
		return
	}

	gReq, err := h.svc.GetRequest(c.Request.Context(), id, orgID)
	if err != nil {
		if err == service.ErrRequestNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gReq)
}

func (h *GovernanceHandler) ListRequests(c *gin.Context) {
	orgID, _ := uuid.Parse(c.GetHeader("X-Organization-ID"))
	
	var filter models.GovernanceRequestFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	filter.OrganizationID = orgID

	resp, err := h.svc.ListRequests(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *GovernanceHandler) ApproveRequest(c *gin.Context) {
	orgID, _ := uuid.Parse(c.GetHeader("X-Organization-ID"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request id"})
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	_ = c.ShouldBindJSON(&req)

	if err := h.svc.ApproveRequest(c.Request.Context(), id, orgID, req.Notes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *GovernanceHandler) RejectRequest(c *gin.Context) {
	orgID, _ := uuid.Parse(c.GetHeader("X-Organization-ID"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request id"})
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	_ = c.ShouldBindJSON(&req)

	if err := h.svc.RejectRequest(c.Request.Context(), id, orgID, req.Notes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *GovernanceHandler) GetSettings(c *gin.Context) {
	orgID, _ := uuid.Parse(c.GetHeader("X-Organization-ID"))
	
	settings, err := h.svc.GetSettings(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *GovernanceHandler) UpdateSettings(c *gin.Context) {
	orgID, _ := uuid.Parse(c.GetHeader("X-Organization-ID"))
	
	var settings models.GovernanceSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdateSettings(c.Request.Context(), orgID, &settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}
