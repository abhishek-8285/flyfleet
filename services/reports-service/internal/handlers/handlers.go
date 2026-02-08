package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/flyfleet/reports-service/internal/models"
	"github.com/flyfleet/reports-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ReportHandler struct {
	svc    *service.ReportService
	logger *zap.Logger
}

func NewReportHandler(svc *service.ReportService, logger *zap.Logger) *ReportHandler {
	return &ReportHandler{
		svc:    svc,
		logger: logger,
	}
}

func (h *ReportHandler) bindStrictJSON(c *gin.Context, obj any) error {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return binding.Validator.ValidateStruct(obj)
}

func (h *ReportHandler) CreateReport(c *gin.Context) {
	var req models.ScheduledReport
	if err := h.bindStrictJSON(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": err.Error()})
		return
	}

	// Set Org ID (From Context/JWT in real app)
	req.OrganizationID = uuid.New() // Placeholder
	req.CreatedBy = uuid.New()      // Placeholder

	id, err := h.svc.CreateReport(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create report"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *ReportHandler) ListReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	reports, total, err := h.svc.ListReports(c.Request.Context(), page, limit)
	if err != nil {
		h.logger.Error("Failed to list reports", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list reports"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  reports,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *ReportHandler) GetReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id used"})
		return
	}

	report, err := h.svc.GetReport(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get report"})
		return
	}
	if report == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}
	c.JSON(http.StatusOK, report)
}

func (h *ReportHandler) UpdateReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id used"})
		return
	}

	var req models.ScheduledReport
	if err := h.bindStrictJSON(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": err.Error()})
		return
	}
	req.ID = id
	temp := uuid.New()
	req.UpdatedBy = &temp
	now := time.Now()
	req.UpdatedAt = &now

	if err := h.svc.UpdateReport(c.Request.Context(), &req); err != nil {
		h.logger.Error("Failed to update report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update report"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *ReportHandler) PauseReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id used"})
		return
	}
	if err := h.svc.PauseReport(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to pause"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "paused"})
}

func (h *ReportHandler) ResumeReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id used"})
		return
	}
	if err := h.svc.ResumeReport(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resume"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "resumed"})
}

func (h *ReportHandler) ListExecutions(c *gin.Context) {
	reportIDStr := c.Query("report_id")
	var reportID uuid.UUID
	if reportIDStr != "" {
		reportID, _ = uuid.Parse(reportIDStr)
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	execs, total, err := h.svc.ListExecutions(c.Request.Context(), reportID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list executions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"executions": execs,
		"total":      total,
		"page":       page, 
		"limit":      limit,
	})
}

func (h *ReportHandler) GetSettings(c *gin.Context) {
	s, err := h.svc.GetSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch settings"})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *ReportHandler) UpdateSettings(c *gin.Context) {
	var req models.GlobalSettings
	if err := h.bindStrictJSON(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
		return
	}
	if err := h.svc.UpdateSettings(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update settings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}
