package handlers

import (
	"net/http"

	"github.com/flyfleet/dr-service/internal/models"
	"github.com/flyfleet/dr-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DRHandlers struct {
	svc    *service.DRService
	logger *zap.Logger
}

func NewDRHandlers(svc *service.DRService, logger *zap.Logger) *DRHandlers {
	return &DRHandlers{svc: svc, logger: logger}
}

func (h *DRHandlers) ListBackups(c *gin.Context) {
	backups, err := h.svc.ListBackups(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list backups", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch backups"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": backups})
}

func (h *DRHandlers) CreateBackup(c *gin.Context) {
	var req models.CreateBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	backup, err := h.svc.CreateBackup(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create backup", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate backup"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": backup, "message": "Backup job initiated"})
}

func (h *DRHandlers) ListStoragePools(c *gin.Context) {
	pools, err := h.svc.ListStoragePools(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list storage pools", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch storage pools"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pools})
}

func (h *DRHandlers) GetBackup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid backup ID"})
		return
	}
	b, err := h.svc.GetBackupDetails(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Backup not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": b})
}

func (h *DRHandlers) VerifyBackup(c *gin.Context) {
	// For now just return OK as simulated verification happens automatically
	c.JSON(http.StatusOK, gin.H{"message": "Verification scheduled"})
}

// --- Restore Operations ---

func (h *DRHandlers) ListRestores(c *gin.Context) {
	restores, err := h.svc.ListRestores(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list restores", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch restores"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": restores})
}

func (h *DRHandlers) CreateRestore(c *gin.Context) {
	var req models.CreateRestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	restore, err := h.svc.InitiateRestore(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to initiate restore", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate restore"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": restore, "message": "Restore job initiated"})
}

func (h *DRHandlers) GetRestore(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid restore ID"})
		return
	}
	r, err := h.svc.GetRestoreDetails(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Restore not found"})
		return
	}
	
	// Fetch logs too
	logs, _ := h.svc.GetRestoreLogs(c.Request.Context(), id)
	
	c.JSON(http.StatusOK, gin.H{"data": r, "logs": logs})
}

func (h *DRHandlers) GetRestoreLogs(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid restore ID"})
		return
	}
	logs, err := h.svc.GetRestoreLogs(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": logs})
}

func (h *DRHandlers) CancelRestore(c *gin.Context) {
	// Placeholder
	c.JSON(http.StatusOK, gin.H{"status": "CANCELLED"})
}

// ========== DR Testing Handlers ==========

func (h *DRHandlers) ListDRTests(c *gin.Context) {
	tests, err := h.svc.ListDRTests(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list DR tests", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch DR tests"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tests})
}

func (h *DRHandlers) CreateDRTest(c *gin.Context) {
	var req models.CreateDRTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	test, err := h.svc.CreateDRTest(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create DR test", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule DR test"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": test, "message": "DR test scheduled"})
}

func (h *DRHandlers) GetDRTest(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid test ID"})
		return
	}
	test, steps, err := h.svc.GetDRTestDetails(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "DR test not found"})
		return
	}

	logs, _ := h.svc.GetDRTestLogs(c.Request.Context(), id)

	c.JSON(http.StatusOK, gin.H{"data": test, "steps": steps, "logs": logs})
}

func (h *DRHandlers) GetDRTestLogs(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid test ID"})
		return
	}
	logs, err := h.svc.GetDRTestLogs(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": logs})
}

func (h *DRHandlers) AbortDRTest(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid test ID"})
		return
	}
	if err := h.svc.AbortDRTest(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to abort test"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ABORTED"})
}

func (h *DRHandlers) GetDRTestMetrics(c *gin.Context) {
	metrics, err := h.svc.GetDRTestMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch metrics"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": metrics})
}

func (h *DRHandlers) GetDRTestSettings(c *gin.Context) {
	settings, err := h.svc.GetDRTestSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": settings})
}

func (h *DRHandlers) UpdateDRTestSettings(c *gin.Context) {
	var req models.UpdateDRTestSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdateDRTestSettings(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Settings updated"})
}
