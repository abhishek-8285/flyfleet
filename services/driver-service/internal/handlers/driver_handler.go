package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"

	"github.com/flyfleet/driver-service/internal/models"
	"github.com/flyfleet/driver-service/internal/service"
)

type DriverHandler struct {
	service *service.DriverService
	logger  *zap.Logger
	validate *validator.Validate
}

func NewDriverHandler(service *service.DriverService, logger *zap.Logger) *DriverHandler {
	return &DriverHandler{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}

// CreateDriver godoc
// @Summary Create a new driver
// @Description Create a new driver with personal and license information
// @Tags drivers
// @Accept json
// @Produce json
// @Param driver body models.CreateDriverRequest true "Driver information"
// @Success 201 {object} models.Driver
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drivers [post]
func (h *DriverHandler) CreateDriver(c *gin.Context) {
	var req models.CreateDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind create driver request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("validation failed for create driver request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orgID := c.GetHeader("X-Org-ID")
	if orgID == "" {
		orgID = "default-org"
	}

	driver, err := h.service.Create(c.Request.Context(), orgID, &req)
	if err != nil {
		h.logger.Error("failed to create driver", zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err.Error() == "license has expired" || err.Error() == "driver with license number "+req.LicenseNumber+" already exists" {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, driver)
}

// GetDriver godoc
// @Summary Get driver by ID
// @Description Get driver details by driver ID
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Success 200 {object} models.Driver
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drivers/{id} [get]
func (h *DriverHandler) GetDriver(c *gin.Context) {
	driverID := c.Param("id")
	
	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver ID format"})
		return
	}

	driver, err := h.service.GetByID(c.Request.Context(), driverID)
	if err != nil {
		h.logger.Error("failed to get driver", zap.String("driver_id", driverID), zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err.Error() == "driver not found: no rows in result set" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, driver)
}

// ListDrivers godoc
// @Summary List drivers
// @Description Get list of drivers with optional filtering and pagination
// @Tags drivers
// @Accept json
// @Produce json
// @Param status query string false "Filter by status (active, inactive, suspended, resigned)"
// @Param vehicle_id query string false "Filter by vehicle ID"
// @Param name query string false "Filter by name (partial match)"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20, max: 100)"
// @Param order_by query string false "Order by field (created_at, full_name, rating, total_trips, behavior_score, status)"
// @Param order_dir query string false "Order direction (asc, desc)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drivers [get]
func (h *DriverHandler) ListDrivers(c *gin.Context) {
	var filter models.DriverListFilter
	
	// Parse query parameters
	filter.Status = c.Query("status")
	filter.VehicleID = c.Query("vehicle_id")
	filter.Name = c.Query("name")
	
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil || page < 1 {
		page = 1
	}
	filter.Page = page
	
	pageSize, err := strconv.Atoi(c.Query("page_size"))
	if err != nil || pageSize < 1 {
		pageSize = 20
	} else if pageSize > 100 {
		pageSize = 100
	}
	filter.PageSize = pageSize
	
	filter.OrderBy = c.Query("order_by")
	filter.OrderDir = c.Query("order_dir")

	drivers, totalCount, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list drivers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"data":       drivers,
		"total":      totalCount,
		"page":       filter.Page,
		"page_size":  filter.PageSize,
		"total_pages": (totalCount + filter.PageSize - 1) / filter.PageSize,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateDriver godoc
// @Summary Update driver
// @Description Update driver information
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param driver body models.UpdateDriverRequest true "Driver update information"
// @Success 200 {object} models.Driver
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drivers/{id} [put]
func (h *DriverHandler) UpdateDriver(c *gin.Context) {
	driverID := c.Param("id")
	
	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver ID format"})
		return
	}

	var req models.UpdateDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind update driver request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("validation failed for update driver request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	driver, err := h.service.Update(c.Request.Context(), driverID, &req)
	if err != nil {
		h.logger.Error("failed to update driver", zap.String("driver_id", driverID), zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err.Error() == "driver not found: no rows in result set" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, driver)
}

// DeleteDriver godoc
// @Summary Delete driver
// @Description Soft delete a driver (mark as inactive)
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drivers/{id} [delete]
func (h *DriverHandler) DeleteDriver(c *gin.Context) {
	driverID := c.Param("id")
	
	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver ID format"})
		return
	}

	err := h.service.Delete(c.Request.Context(), driverID)
	if err != nil {
		h.logger.Error("failed to delete driver", zap.String("driver_id", driverID), zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err.Error() == "driver not found: no rows in result set" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "cannot delete driver with active vehicle assignment" ||
		          err.Error() == "cannot delete driver with active trip" {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// AssignVehicle godoc
// @Summary Assign vehicle to driver
// @Description Assign a vehicle to a driver
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param assignment body models.AssignVehicleRequest true "Vehicle assignment"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drivers/{id}/assign-vehicle [post]
func (h *DriverHandler) AssignVehicle(c *gin.Context) {
	driverID := c.Param("id")
	
	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver ID format"})
		return
	}

	var req models.AssignVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind assign vehicle request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("validation failed for assign vehicle request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate vehicle ID format
	if _, err := uuid.Parse(req.VehicleID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle ID format"})
		return
	}

	err := h.service.AssignVehicle(c.Request.Context(), driverID, req.VehicleID)
	if err != nil {
		h.logger.Error("failed to assign vehicle to driver", 
			zap.String("driver_id", driverID), 
			zap.String("vehicle_id", req.VehicleID), 
			zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err.Error() == "driver not found: no rows in result set" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "driver is not active" ||
		          err.Error() == "driver already has a vehicle assigned" {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Vehicle assigned successfully",
		"driver_id": driverID,
		"vehicle_id": req.VehicleID,
	})
}

// UnassignVehicle godoc
// @Summary Unassign vehicle from driver
// @Description Remove vehicle assignment from a driver
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drivers/{id}/unassign-vehicle [post]
func (h *DriverHandler) UnassignVehicle(c *gin.Context) {
	driverID := c.Param("id")
	
	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver ID format"})
		return
	}

	err := h.service.UnassignVehicle(c.Request.Context(), driverID)
	if err != nil {
		h.logger.Error("failed to unassign vehicle from driver", 
			zap.String("driver_id", driverID), 
			zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err.Error() == "driver not found: no rows in result set" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "driver has no vehicle assigned" {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Vehicle unassigned successfully",
		"driver_id": driverID,
	})
}

// UpdateStatus godoc
// @Summary Update driver status
// @Description Update driver status (active, inactive, suspended, resigned)
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param status body models.UpdateStatusRequest true "Status update"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drivers/{id}/update-status [post]
func (h *DriverHandler) UpdateStatus(c *gin.Context) {
	driverID := c.Param("id")
	
	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver ID format"})
		return
	}

	var req models.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind update status request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("validation failed for update status request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.UpdateStatus(c.Request.Context(), driverID, req.Status)
	if err != nil {
		h.logger.Error("failed to update driver status", 
			zap.String("driver_id", driverID), 
			zap.String("status", req.Status), 
			zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err.Error() == "driver not found: no rows in result set" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "invalid status: "+req.Status {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Driver status updated successfully",
		"driver_id": driverID,
		"status": req.Status,
	})
}

// GetBehaviorScore godoc
// @Summary Get driver behavior score
// @Description Get current behavior score for a driver
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drivers/{id}/behavior-score [get]
func (h *DriverHandler) GetBehaviorScore(c *gin.Context) {
	driverID := c.Param("id")
	
	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver ID format"})
		return
	}

	score, err := h.service.GetBehaviorScore(c.Request.Context(), driverID)
	if err != nil {
		h.logger.Error("failed to get behavior score", 
			zap.String("driver_id", driverID), 
			zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err.Error() == "driver not found: no rows in result set" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"driver_id": driverID,
		"behavior_score": score,
		"max_score": 100,
	})
}

// UpdateBehaviorScore godoc
// @Summary Update driver behavior score
// @Description Update driver's behavior score based on driving events
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param request body models.UpdateBehaviorScoreRequest true "Score update request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drivers/{id}/update-behavior [post]
func (h *DriverHandler) UpdateBehaviorScore(c *gin.Context) {
	driverID := c.Param("id")
	
	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver ID format"})
		return
	}

	var req models.UpdateBehaviorScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind update behavior score request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("validation failed for update behavior score request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.UpdateBehaviorScore(c.Request.Context(), driverID, req.ScoreChange, req.Reason)
	if err != nil {
		h.logger.Error("failed to update behavior score", 
			zap.String("driver_id", driverID), 
			zap.Int("score_change", req.ScoreChange),
			zap.String("reason", req.Reason),
			zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err.Error() == "driver not found: no rows in result set" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "score change must be between -20 and 20" {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	// Get updated score
	newScore, err := h.service.GetBehaviorScore(c.Request.Context(), driverID)
	if err != nil {
		newScore = 0 // fallback
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Behavior score updated successfully",
		"driver_id": driverID,
		"score_change": req.ScoreChange,
		"new_score": newScore,
		"reason": req.Reason,
	})
}

// GetAnalytics godoc
// @Summary Get driver analytics
// @Description Get performance analytics for a driver
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Success 200 {object} models.DriverAnalytics
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drivers/{id}/analytics [get]
func (h *DriverHandler) GetAnalytics(c *gin.Context) {
	driverID := c.Param("id")
	
	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver ID format"})
		return
	}

	analytics, err := h.service.GetDriverAnalytics(c.Request.Context(), driverID)
	if err != nil {
		h.logger.Error("failed to get driver analytics", 
			zap.String("driver_id", driverID), 
			zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err.Error() == "driver not found: no rows in result set" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// UploadDocument godoc
// @Summary Upload driver document
// @Description Upload and store driver document
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param document body models.UploadDocumentRequest true "Document information"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drivers/{id}/documents [post]
func (h *DriverHandler) UploadDocument(c *gin.Context) {
	driverID := c.Param("id")
	
	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver ID format"})
		return
	}

	var req models.UploadDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind upload document request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("validation failed for upload document request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// In a real implementation, this would store the document and update the driver record
	// For now, we'll just return a success response
	docID := uuid.New().String()
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "Document uploaded successfully",
		"driver_id": driverID,
		"document_id": docID,
		"document_type": req.DocumentType,
		"document_number": req.DocumentNumber,
	})
}

// ListDocuments godoc
// @Summary List driver documents
// @Description Get list of documents for a driver
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drivers/{id}/documents [get]
func (h *DriverHandler) ListDocuments(c *gin.Context) {
	driverID := c.Param("id")
	
	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver ID format"})
		return
	}

	// In a real implementation, this would fetch documents from the database
	// For now, return empty list
	c.JSON(http.StatusOK, gin.H{
		"driver_id": driverID,
		"documents": []interface{}{},
		"count": 0,
	})
}

// DeleteDocument godoc
// @Summary Delete driver document
// @Description Delete a document from driver's records
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param doc_id path string true "Document ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drivers/{id}/documents/{doc_id} [delete]
func (h *DriverHandler) DeleteDocument(c *gin.Context) {
	driverID := c.Param("id")
	docID := c.Param("doc_id")
	
	// Validate UUID formats
	if _, err := uuid.Parse(driverID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver ID format"})
		return
	}
	
	if _, err := uuid.Parse(docID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID format"})
		return
	}

	// In a real implementation, this would delete the document from storage and database
	// For now, just return success
	c.JSON(http.StatusOK, gin.H{
		"message": "Document deleted successfully",
		"driver_id": driverID,
		"document_id": docID,
	})
}

// MarkAttendance godoc
// @Summary Mark driver attendance
// @Description Mark attendance for a driver on a specific date
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param attendance body models.MarkAttendanceRequest true "Attendance record"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drivers/{id}/attendance [post]
func (h *DriverHandler) MarkAttendance(c *gin.Context) {
	driverID := c.Param("id")
	
	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver ID format"})
		return
	}

	var req models.MarkAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind mark attendance request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("validation failed for mark attendance request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// In a real implementation, this would store the attendance record
	// For now, just return success
	recordID := uuid.New().String()
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "Attendance marked successfully",
		"driver_id": driverID,
		"record_id": recordID,
		"date": req.Date.Format("2006-01-02"),
		"status": req.Status,
	})
}

// GetAttendance godoc
// @Summary Get driver attendance
// @Description Get attendance records for a driver
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drivers/{id}/attendance [get]
func (h *DriverHandler) GetAttendance(c *gin.Context) {
	driverID := c.Param("id")
	
	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver ID format"})
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	
	var startDate, endDate time.Time
	var err error
	
	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
			return
		}
	} else {
		startDate = time.Now().AddDate(0, 0, -30) // Last 30 days default
	}
	
	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
	} else {
		endDate = time.Now()
	}

	// In a real implementation, this would fetch attendance records from database
	// For now, return empty list
	c.JSON(http.StatusOK, gin.H{
		"driver_id": driverID,
		"attendance_records": []interface{}{},
		"start_date": startDate.Format("2006-01-02"),
		"end_date": endDate.Format("2006-01-02"),
		"count": 0,
	})
}

// SetSchedule godoc
// @Summary Set driver schedule
// @Description Set work schedule for a driver
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param schedule body models.SetScheduleRequest true "Work schedule"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drivers/{id}/schedule [post]
func (h *DriverHandler) SetSchedule(c *gin.Context) {
	driverID := c.Param("id")
	
	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver ID format"})
		return
	}

	var req models.SetScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind set schedule request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("validation failed for set schedule request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// In a real implementation, this would store the schedule
	// For now, just return success
	scheduleID := uuid.New().String()
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "Schedule set successfully",
		"driver_id": driverID,
		"schedule_id": scheduleID,
		"day_of_week": req.DayOfWeek,
		"start_time": req.StartTime,
		"end_time": req.EndTime,
		"is_working": req.IsWorking,
	})
}

// GetSchedule godoc
// @Summary Get driver schedule
// @Description Get work schedule for a driver
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drivers/{id}/schedule [get]
func (h *DriverHandler) GetSchedule(c *gin.Context) {
	driverID := c.Param("id")
	
	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver ID format"})
		return
	}

	// In a real implementation, this would fetch schedules from database
	// For now, return empty list
	c.JSON(http.StatusOK, gin.H{
		"driver_id": driverID,
		"schedules": []interface{}{},
		"count": 0,
	})
}
