package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/flyfleet/admin-support-service/internal/domains/dr/models"
	"github.com/flyfleet/admin-support-service/internal/domains/dr/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DrHandlers struct {
	service service.DrService
	logger  *zap.Logger
}

func NewDrHandlers(service service.DrService, logger *zap.Logger) *DrHandlers {
	return &DrHandlers{service: service, logger: logger}
}

func (h *DrHandlers) RegisterRoutes(router *gin.RouterGroup) {
	dr := router.Group("/dr")
	{
		dr.POST("/backups", h.InitiateBackup)
		dr.GET("/backups", h.ListBackups)
		dr.GET("/backups/:id", h.GetBackup)
		dr.POST("/backups/:id/verify", h.VerifyBackup)

		dr.POST("/restore-jobs", h.InitiateRestore)
		dr.GET("/restore-jobs", h.ListRestoreJobs)
		dr.GET("/restore-jobs/:id", h.GetRestoreJob)
		dr.GET("/restore-jobs/:id/logs", h.GetRestoreJobLogs)

		dr.GET("/storage-pools", h.ListStoragePools)
	}
}

// Request Models
type InitiateBackupRequest struct {
	Type models.BackupType `json:"type" binding:"required"`
}

type InitiateRestoreRequest struct {
	BackupID          uuid.UUID `json:"backup_id" binding:"required"`
	TargetEnvironment string    `json:"target_environment" binding:"required"`
}

func (h *DrHandlers) InitiateBackup(c *gin.Context) {
	var req InitiateBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID := h.getUserID(c)
	backup, err := h.service.InitiateBackup(c.Request.Context(), req.Type, userID)
	if err != nil {
		h.logger.Error("Failed to initiate backup", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate backup"})
		return
	}

	c.JSON(http.StatusAccepted, backup)
}

func (h *DrHandlers) ListBackups(c *gin.Context) {
	limit, offset := h.getPaginationParams(c)
	backups, total, err := h.service.ListBackups(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list backups", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list backups"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  backups,
		"total": total,
		"limit": limit,
		"offset": offset,
	})
}

func (h *DrHandlers) GetBackup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid backup ID"})
		return
	}

	backup, err := h.service.GetBackup(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Backup not found"})
		return
	}

	c.JSON(http.StatusOK, backup)
}

func (h *DrHandlers) InitiateRestore(c *gin.Context) {
	var req InitiateRestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID := h.getUserID(c)
	job, err := h.service.InitiateRestore(c.Request.Context(), req.BackupID, req.TargetEnvironment, userID)
	if err != nil {
		h.logger.Error("Failed to initiate restore", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, job)
}

func (h *DrHandlers) ListRestoreJobs(c *gin.Context) {
	limit, offset := h.getPaginationParams(c)
	jobs, total, err := h.service.ListRestoreJobs(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list restore jobs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list restore jobs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  jobs,
		"total": total,
		"limit": limit,
		"offset": offset,
	})
}

func (h *DrHandlers) GetRestoreJob(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid restore job ID"})
		return
	}

	job, err := h.service.GetRestoreJob(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Restore job not found"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// Helpers
func (h *DrHandlers) getUserID(c *gin.Context) uuid.UUID {
	idVal, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		return uuid.Nil
	}

	idStr, ok := idVal.(string)
	if !ok {
		h.logger.Error("User ID in context is not a string")
		return uuid.Nil
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("Invalid User ID format", zap.Error(err))
		return uuid.Nil
	}
	return id
}

func (h *DrHandlers) getPaginationParams(c *gin.Context) (int, int) {
	limit := 10
	offset := 0
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	if o := c.Query("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}
	return limit, offset
}

func (h *DrHandlers) VerifyBackup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid backup ID"})
		return
	}

	userID := h.getUserID(c)
	verification, err := h.service.VerifyBackup(c.Request.Context(), id, userID)
	if err != nil {
		h.logger.Error("Failed to verify backup", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, verification)
}

func (h *DrHandlers) GetRestoreJobLogs(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid restore job ID"})
		return
	}

	job, err := h.service.GetRestoreJob(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Restore job not found"})
		return
	}

	// Parse logs into an array of log entries
	var logs []map[string]string
	if job.Logs != "" {
		logs = []map[string]string{
			{"level": "INFO", "message": job.Logs, "timestamp": job.StartedAt.Format(time.RFC3339)},
		}
		if job.CompletedAt != nil {
			logs = append(logs, map[string]string{
				"level":     "INFO",
				"message":   "Restore job completed",
				"timestamp": job.CompletedAt.Format(time.RFC3339),
			})
		}
	} else {
		logs = []map[string]string{}
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}

func (h *DrHandlers) ListStoragePools(c *gin.Context) {
	pools, err := h.service.ListStoragePools(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list storage pools", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list storage pools"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pools})
}
