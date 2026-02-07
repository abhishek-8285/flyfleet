package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/maintenance-service/internal/models"
	"github.com/flyfleet/maintenance-service/internal/service"
)

type MaintenanceHandler struct {
	service *service.MaintenanceService
	logger  *zap.Logger
}

func NewMaintenanceHandler(svc *service.MaintenanceService, logger *zap.Logger) *MaintenanceHandler {
	return &MaintenanceHandler{
		service: svc,
		logger:  logger,
	}
}

// CreateSchedule godoc
// @Summary Create maintenance schedule
// @Description Create a new maintenance schedule for a vehicle
// @Tags Maintenance
// @Accept json
// @Produce json
// @Param schedule body models.CreateScheduleRequest true "Schedule creation request"
// @Success 201 {object} models.MaintenanceSchedule
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /maintenance/schedules [post]
func (h *MaintenanceHandler) CreateSchedule(c *gin.Context) {
	var req models.CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get organization ID from context
	orgIDStr := c.GetString("org_id")
	if orgIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization ID required"})
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	// Get user ID from context
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	schedule, err := h.service.CreateSchedule(c.Request.Context(), orgID, &req, userID)
	if err != nil {
		h.logger.Error("failed to create maintenance schedule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, schedule)
}

// GetSchedule godoc
// @Summary Get maintenance schedule
// @Description Retrieve a specific maintenance schedule by ID
// @Tags Maintenance
// @Produce json
// @Param id path string true "Schedule ID"
// @Success 200 {object} models.MaintenanceSchedule
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /maintenance/schedules/{id} [get]
func (h *MaintenanceHandler) GetSchedule(c *gin.Context) {
	id := c.Param("id")

	schedule, err := h.service.GetSchedule(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get maintenance schedule", zap.String("schedule_id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Maintenance schedule not found"})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

// ListSchedules godoc
// @Summary List maintenance schedules
// @Description List maintenance schedules with filtering and pagination
// @Tags Maintenance
// @Produce json
// @Param type query []string false "Filter by schedule type"
// @Param status query []string false "Filter by status"
// @Param priority query []string false "Filter by priority"
// @Param vehicle_id query string false "Filter by vehicle ID"
// @Param assigned_to query string false "Filter by assigned mechanic"
// @Param from_date query string false "From date (RFC3339)"
// @Param to_date query string false "To date (RFC3339)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(20)
// @Param order_by query string false "Order by field"
// @Param order_dir query string false "Order direction" Enums(asc, desc)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /maintenance/schedules [get]
func (h *MaintenanceHandler) ListSchedules(c *gin.Context) {
	// Get organization ID from context
	orgIDStr := c.GetString("org_id")
	if orgIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization ID required"})
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	// Parse query parameters
	filter := models.MaintenanceListFilter{
		OrgID:    orgID,
		Page:     1,
		PageSize: 20,
		OrderBy:  "due_date",
		OrderDir: "asc",
	}

	// Parse page and page_size
	if page := c.Query("page"); page != "" {
		if pageNum, err := strconv.Atoi(page); err == nil && pageNum >= 1 {
			filter.Page = pageNum
		}
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		if size, err := strconv.Atoi(pageSize); err == nil && size >= 1 && size <= 100 {
			filter.PageSize = size
		}
	}

	// Parse filters
	if types := c.QueryArray("type"); len(types) > 0 {
		filter.Type = types
	}

	if statuses := c.QueryArray("status"); len(statuses) > 0 {
		filter.Status = statuses
	}

	if priorities := c.QueryArray("priority"); len(priorities) > 0 {
		filter.Priority = priorities
	}

	if vehicleIDStr := c.Query("vehicle_id"); vehicleIDStr != "" {
		vehicleID, err := uuid.Parse(vehicleIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle_id"})
			return
		}
		filter.VehicleID = &vehicleID
	}

	if assignedToStr := c.Query("assigned_to"); assignedToStr != "" {
		assignedTo, err := uuid.Parse(assignedToStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assigned_to"})
			return
		}
		filter.AssignedTo = &assignedTo
	}

	// Parse date filters
	if fromDateStr := c.Query("from_date"); fromDateStr != "" {
		fromDate, err := time.Parse(time.RFC3339, fromDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from_date format"})
			return
		}
		filter.FromDate = fromDate
	}

	if toDateStr := c.Query("to_date"); toDateStr != "" {
		toDate, err := time.Parse(time.RFC3339, toDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to_date format"})
			return
		}
		filter.ToDate = toDate
	}

	schedules, totalCount, err := h.service.ListSchedules(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list maintenance schedules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list maintenance schedules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        schedules,
		"total":       totalCount,
		"page":        filter.Page,
		"page_size":   filter.PageSize,
		"total_pages": (totalCount + filter.PageSize - 1) / filter.PageSize,
	})
}

// UpdateSchedule godoc
// @Summary Update maintenance schedule
// @Description Update an existing maintenance schedule
// @Tags Maintenance
// @Accept json
// @Produce json
// @Param id path string true "Schedule ID"
// @Param schedule body models.UpdateScheduleRequest true "Schedule update request"
// @Success 200 {object} models.MaintenanceSchedule
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /maintenance/schedules/{id} [put]
func (h *MaintenanceHandler) UpdateSchedule(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	schedule, err := h.service.UpdateSchedule(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to update maintenance schedule", zap.String("schedule_id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Maintenance schedule not found"})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

// DeleteSchedule godoc
// @Summary Delete maintenance schedule
// @Description Delete a maintenance schedule
// @Tags Maintenance
// @Produce json
// @Param id path string true "Schedule ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /maintenance/schedules/{id} [delete]
func (h *MaintenanceHandler) DeleteSchedule(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteSchedule(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to delete maintenance schedule", zap.String("schedule_id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Maintenance schedule not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateLog godoc
// @Summary Create maintenance log
// @Description Create a new maintenance log entry
// @Tags Maintenance
// @Accept json
// @Produce json
// @Param log body models.CreateLogRequest true "Log creation request"
// @Success 201 {object} models.MaintenanceLog
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /maintenance/logs [post]
func (h *MaintenanceHandler) CreateLog(c *gin.Context) {
	var req models.CreateLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get organization ID from context
	orgIDStr := c.GetString("org_id")
	if orgIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization ID required"})
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	log, err := h.service.CreateLog(c.Request.Context(), orgID, &req)
	if err != nil {
		h.logger.Error("failed to create maintenance log", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, log)
}

// GetLog godoc
// @Summary Get maintenance log
// @Description Retrieve a specific maintenance log by ID
// @Tags Maintenance
// @Produce json
// @Param id path string true "Log ID"
// @Success 200 {object} models.MaintenanceLog
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /maintenance/logs/{id} [get]
func (h *MaintenanceHandler) GetLog(c *gin.Context) {
	id := c.Param("id")

	log, err := h.service.GetLog(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get maintenance log", zap.String("log_id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Maintenance log not found"})
		return
	}

	c.JSON(http.StatusOK, log)
}

// ListLogs godoc
// @Summary List maintenance logs
// @Description List maintenance logs with filtering and pagination
// @Tags Maintenance
// @Produce json
// @Param type query []string false "Filter by log type"
// @Param status query []string false "Filter by status"
// @Param vehicle_id query string false "Filter by vehicle ID"
// @Param performed_by query string false "Filter by mechanic"
// @Param from_date query string false "From date (RFC3339)"
// @Param to_date query string false "To date (RFC3339)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(20)
// @Param order_by query string false "Order by field"
// @Param order_dir query string false "Order direction" Enums(asc, desc)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /maintenance/logs [get]
func (h *MaintenanceHandler) ListLogs(c *gin.Context) {
	// Get organization ID from context
	orgIDStr := c.GetString("org_id")
	if orgIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization ID required"})
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	// Parse query parameters
	filter := models.MaintenanceListFilter{
		OrgID:    orgID,
		Page:     1,
		PageSize: 20,
		OrderBy:  "start_date",
		OrderDir: "desc",
	}

	// Parse page and page_size
	if page := c.Query("page"); page != "" {
		if pageNum, err := strconv.Atoi(page); err == nil && pageNum >= 1 {
			filter.Page = pageNum
		}
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		if size, err := strconv.Atoi(pageSize); err == nil && size >= 1 && size <= 100 {
			filter.PageSize = size
		}
	}

	// Parse filters
	if types := c.QueryArray("type"); len(types) > 0 {
		filter.Type = types
	}

	if statuses := c.QueryArray("status"); len(statuses) > 0 {
		filter.Status = statuses
	}

	if vehicleIDStr := c.Query("vehicle_id"); vehicleIDStr != "" {
		vehicleID, err := uuid.Parse(vehicleIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle_id"})
			return
		}
		filter.VehicleID = &vehicleID
	}

	if performedByStr := c.Query("performed_by"); performedByStr != "" {
		performedBy, err := uuid.Parse(performedByStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid performed_by"})
			return
		}
		filter.AssignedTo = &performedBy
	}

	// Parse date filters
	if fromDateStr := c.Query("from_date"); fromDateStr != "" {
		fromDate, err := time.Parse(time.RFC3339, fromDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from_date format"})
			return
		}
		filter.FromDate = fromDate
	}

	if toDateStr := c.Query("to_date"); toDateStr != "" {
		toDate, err := time.Parse(time.RFC3339, toDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to_date format"})
			return
		}
		filter.ToDate = toDate
	}

	logs, totalCount, err := h.service.ListLogs(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list maintenance logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list maintenance logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        logs,
		"total":       totalCount,
		"page":        filter.Page,
		"page_size":   filter.PageSize,
		"total_pages": (totalCount + filter.PageSize - 1) / filter.PageSize,
	})
}

// UpdateLog godoc
// @Summary Update maintenance log
// @Description Update an existing maintenance log
// @Tags Maintenance
// @Accept json
// @Produce json
// @Param id path string true "Log ID"
// @Param log body models.UpdateLogRequest true "Log update request"
// @Success 200 {object} models.MaintenanceLog
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /maintenance/logs/{id} [put]
func (h *MaintenanceHandler) UpdateLog(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	log, err := h.service.UpdateLog(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to update maintenance log", zap.String("log_id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Maintenance log not found"})
		return
	}

	c.JSON(http.StatusOK, log)
}

// DeleteLog godoc
// @Summary Delete maintenance log
// @Description Delete a maintenance log
// @Tags Maintenance
// @Produce json
// @Param id path string true "Log ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /maintenance/logs/{id} [delete]
func (h *MaintenanceHandler) DeleteLog(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteLog(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to delete maintenance log", zap.String("log_id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Maintenance log not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetVehicleMaintenanceHistory godoc
// @Summary Get vehicle maintenance history
// @Description Get maintenance history for a specific vehicle
// @Tags Maintenance
// @Produce json
// @Param vehicle_id path string true "Vehicle ID"
// @Param limit query int false "Number of records to return" default(20)
// @Success 200 {object} []models.MaintenanceLog
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /vehicles/{vehicle_id}/maintenance-history [get]
func (h *MaintenanceHandler) GetVehicleMaintenanceHistory(c *gin.Context) {
	vehicleIDStr := c.Param("vehicle_id")
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle ID"})
		return
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if limitVal, err := strconv.Atoi(limitStr); err == nil && limitVal > 0 && limitVal <= 100 {
			limit = limitVal
		}
	}

	logs, err := h.service.GetMaintenanceHistory(c.Request.Context(), vehicleID, limit)
	if err != nil {
		h.logger.Error("failed to get maintenance history", zap.String("vehicle_id", vehicleIDStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get maintenance history"})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// GetUpcomingMaintenance godoc
// @Summary Get upcoming maintenance tasks
// @Description Get upcoming maintenance schedules
// @Tags Maintenance
// @Produce json
// @Param days query int false "Number of days to look ahead" default(30)
// @Success 200 {object} []models.MaintenanceSchedule
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /maintenance/upcoming [get]
func (h *MaintenanceHandler) GetUpcomingMaintenance(c *gin.Context) {
	// Get organization ID from context
	orgIDStr := c.GetString("org_id")
	if orgIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization ID required"})
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	days := 30
	if daysStr := c.Query("days"); daysStr != "" {
		if daysVal, err := strconv.Atoi(daysStr); err == nil && daysVal > 0 && daysVal <= 365 {
			days = daysVal
		}
	}

	schedules, err := h.service.GetUpcomingSchedules(c.Request.Context(), orgID, days)
	if err != nil {
		h.logger.Error("failed to get upcoming maintenance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get upcoming maintenance"})
		return
	}

	c.JSON(http.StatusOK, schedules)
}

// GetOverdueMaintenance godoc
// @Summary Get overdue maintenance tasks
// @Description Get overdue maintenance schedules
// @Tags Maintenance
// @Produce json
// @Success 200 {object} []models.MaintenanceSchedule
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /maintenance/overdue [get]
func (h *MaintenanceHandler) GetOverdueMaintenance(c *gin.Context) {
	// Get organization ID from context
	orgIDStr := c.GetString("org_id")
	if orgIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization ID required"})
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	schedules, err := h.service.GetOverdueSchedules(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to get overdue maintenance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get overdue maintenance"})
		return
	}

	c.JSON(http.StatusOK, schedules)
}

// GetMaintenanceSummary godoc
// @Summary Get maintenance summary
// @Description Get maintenance statistics and summary
// @Tags Maintenance
// @Produce json
// @Param from_date query string true "From date (RFC3339)"
// @Param to_date query string true "To date (RFC3339)"
// @Success 200 {object} models.MaintenanceSummary
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /maintenance/summary [get]
func (h *MaintenanceHandler) GetMaintenanceSummary(c *gin.Context) {
	// Get organization ID from context
	orgIDStr := c.GetString("org_id")
	if orgIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization ID required"})
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	// Parse required date parameters
	fromDateStr := c.Query("from_date")
	toDateStr := c.Query("to_date")

	if fromDateStr == "" || toDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both from_date and to_date are required"})
		return
	}

	fromDate, err := time.Parse(time.RFC3339, fromDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from_date format"})
		return
	}

	toDate, err := time.Parse(time.RFC3339, toDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to_date format"})
		return
	}

	summary, err := h.service.GetMaintenanceSummary(c.Request.Context(), orgID, fromDate, toDate)
	if err != nil {
		h.logger.Error("failed to get maintenance summary", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get maintenance summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}
