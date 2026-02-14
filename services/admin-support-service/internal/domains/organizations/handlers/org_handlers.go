package handlers

import (
	"net/http"
	"strconv"

	"github.com/flyfleet/admin-support-service/internal/domains/organizations/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OrganizationHandlers struct {
	service service.OrganizationService
	logger  *zap.Logger
}

func NewOrganizationHandlers(service service.OrganizationService, logger *zap.Logger) *OrganizationHandlers {
	return &OrganizationHandlers{
		service: service,
		logger:  logger,
	}
}

func (h *OrganizationHandlers) RegisterRoutes(r *gin.RouterGroup) {
	orgs := r.Group("/organizations")
	{
		orgs.GET("", h.ListOrganizations)
		orgs.GET("/stats", h.GetOrganizationStats)
	}
	
	r.GET("/dashboard/stats", h.GetDashboardStats)
}

func (h *OrganizationHandlers) ListOrganizations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	
	orgs, total, err := h.service.GetOrganizations(c.Request.Context(), page, limit, search)
	if err != nil {
		h.logger.Error("Failed to list organizations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch organizations"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data": orgs,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

func (h *OrganizationHandlers) GetOrganizationStats(c *gin.Context) {
	stats, err := h.service.GetOrganizationStats(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get organization stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"data": stats})
}

func (h *OrganizationHandlers) GetDashboardStats(c *gin.Context) {
	stats, err := h.service.GetDashboardStats(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get dashboard stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dashboard stats"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"data": stats})
}
