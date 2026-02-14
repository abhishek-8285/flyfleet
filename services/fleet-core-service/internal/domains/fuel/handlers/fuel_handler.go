package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core-service/internal/domains/fuel/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/fuel/service"
)

type FuelHandler struct {
	svc    service.FuelService
	logger *zap.Logger
}

func NewFuelHandler(svc service.FuelService, logger *zap.Logger) *FuelHandler {
	return &FuelHandler{svc: svc, logger: logger}
}

// CreateLog godoc
// @Summary Create a fuel log
// @Description Create a new fuel log entry
// @Tags Fuel
// @Accept json
// @Produce json
// @Param X-Organization-ID header string true "Organization ID"
// @Param X-User-ID header string true "User ID"
// @Param request body models.CreateFuelLogRequest true "Create Fuel Log Request"
// @Success 201 {object} models.FuelLog
// @Router /fuel-logs [post]
func (h *FuelHandler) CreateLog(c *gin.Context) {
	orgID, _ := uuid.Parse(c.GetHeader("X-Organization-ID"))
	userID, _ := uuid.Parse(c.GetHeader("X-User-ID"))

	var req models.CreateFuelLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log, err := h.svc.CreateLog(c.Request.Context(), orgID, userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, log)
}

// GetLog godoc
// @Summary Get fuel log by ID
// @Tags Fuel
// @Param id path string true "Fuel Log ID"
// @Success 200 {object} models.FuelLog
// @Router /fuel-logs/{id} [get]
func (h *FuelHandler) GetLog(c *gin.Context) {
	orgID, _ := uuid.Parse(c.GetHeader("X-Organization-ID"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	log, err := h.svc.GetLog(c.Request.Context(), id, orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, log)
}

// ListLogs godoc
// @Summary List fuel logs
// @Tags Fuel
// @Param vehicle_id query string false "Vehicle ID"
// @Param driver_id query string false "Driver ID"
// @Param status query []string false "Status"
// @Param search query string false "Search query"
// @Success 200 {object} models.FuelLogListResponse
// @Router /fuel-logs [get]
func (h *FuelHandler) ListLogs(c *gin.Context) {
	orgID, _ := uuid.Parse(c.GetHeader("X-Organization-ID"))

	filter := models.FuelLogFilter{
		OrganizationID: orgID,
		Search:         c.Query("search"),
	}

	if vID := c.Query("vehicle_id"); vID != "" {
		if u, err := uuid.Parse(vID); err == nil {
			filter.VehicleID = &u
		}
	}

	if dID := c.Query("driver_id"); dID != "" {
		if u, err := uuid.Parse(dID); err == nil {
			filter.DriverID = &u
		}
	}

	status := c.QueryArray("status")
	for _, s := range status {
		filter.Status = append(filter.Status, models.FuelLogStatus(s))
	}

	c.JSON(http.StatusOK, func() interface{} {
		res, _ := h.svc.ListLogs(c.Request.Context(), filter)
		return res
	}())
}

// UpdateLog godoc
// @Summary Update fuel log
// @Tags Fuel
// @Param id path string true "Fuel Log ID"
// @Param request body models.UpdateFuelLogRequest true "Update Fuel Log Request"
// @Success 200 {object} models.FuelLog
// @Router /fuel-logs/{id} [put]
func (h *FuelHandler) UpdateLog(c *gin.Context) {
	orgID, _ := uuid.Parse(c.GetHeader("X-Organization-ID"))
	id, _ := uuid.Parse(c.Param("id"))

	var req models.UpdateFuelLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log, err := h.svc.UpdateLog(c.Request.Context(), id, orgID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, log)
}

// DeleteLog godoc
// @Summary Delete fuel log
// @Tags Fuel
// @Param id path string true "Fuel Log ID"
// @Success 204
// @Router /fuel-logs/{id} [delete]
func (h *FuelHandler) DeleteLog(c *gin.Context) {
	orgID, _ := uuid.Parse(c.GetHeader("X-Organization-ID"))
	id, _ := uuid.Parse(c.Param("id"))

	if err := h.svc.DeleteLog(c.Request.Context(), id, orgID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ChangeStatus godoc
// @Summary Approve or Reject fuel log
// @Tags Fuel
// @Param id path string true "Fuel Log ID"
// @Param request body models.ChangeFuelLogStatusRequest true "Status Change Request"
// @Success 200
// @Router /fuel-logs/{id}/status [patch]
func (h *FuelHandler) ChangeStatus(c *gin.Context) {
	orgID, _ := uuid.Parse(c.GetHeader("X-Organization-ID"))
	id, _ := uuid.Parse(c.Param("id"))

	var req models.ChangeFuelLogStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.ChangeStatus(c.Request.Context(), id, orgID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

// GetStats godoc
// @Summary Get fuel stats
// @Tags Fuel
// @Success 200 {object} models.FuelLogStats
// @Router /fuel/stats [get]
func (h *FuelHandler) GetStats(c *gin.Context) {
	orgID, _ := uuid.Parse(c.GetHeader("X-Organization-ID"))
	stats, err := h.svc.GetStats(c.Request.Context(), orgID, models.FuelLogFilter{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetSettings godoc
// @Summary Get fuel settings
// @Tags Fuel
// @Success 200 {object} models.FuelSettings
// @Router /fuel/settings [get]
func (h *FuelHandler) GetSettings(c *gin.Context) {
	orgID, _ := uuid.Parse(c.GetHeader("X-Organization-ID"))
	settings, err := h.svc.GetSettings(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

// UpdateSettings godoc
// @Summary Update fuel settings
// @Tags Fuel
// @Param request body models.UpdateFuelSettingsRequest true "Update Settings Request"
// @Success 200 {object} models.FuelSettings
// @Router /fuel/settings [put]
func (h *FuelHandler) UpdateSettings(c *gin.Context) {
	orgID, _ := uuid.Parse(c.GetHeader("X-Organization-ID"))

	var req models.UpdateFuelSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings, err := h.svc.UpdateSettings(c.Request.Context(), orgID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}
