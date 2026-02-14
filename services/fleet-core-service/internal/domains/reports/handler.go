package reports

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
	reports := rg.Group("/reports")
	{
		reports.POST("", h.CreateReport)
		reports.GET("", h.ListReports)
		reports.GET("/:id", h.GetReport)
	}
}

func (h *Handler) CreateReport(c *gin.Context) {
	var req CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.service.CreateReport(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create report"})
		return
	}

	c.JSON(http.StatusCreated, report)
}

func (h *Handler) GetReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	report, err := h.service.GetReport(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	c.JSON(http.StatusOK, report)
}

func (h *Handler) ListReports(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	reports, err := h.service.ListReports(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list reports", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list reports"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": reports})
}
