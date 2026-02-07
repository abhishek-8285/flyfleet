package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/flyfleet/india-integration-service/internal/models"
	"github.com/flyfleet/india-integration-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// IndiaIntegrationHandlers handles HTTP requests for India integration services
type IndiaIntegrationHandlers struct {
	service service.IndiaIntegrationService
}

// NewIndiaIntegrationHandlers creates a new instance of IndiaIntegrationHandlers
func NewIndiaIntegrationHandlers(svc service.IndiaIntegrationService) *IndiaIntegrationHandlers {
	return &IndiaIntegrationHandlers{
		service: svc,
	}
}

// CreateFASTagTransaction creates a new FASTag transaction
func (h *IndiaIntegrationHandlers) CreateFASTagTransaction(c *gin.Context) {
	var transaction models.FASTagTransaction
	if err := c.ShouldBindJSON(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.CreateFASTagTransaction(ctx, &transaction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, transaction)
}

// GetFASTagTransaction retrieves a FASTag transaction by ID
func (h *IndiaIntegrationHandlers) GetFASTagTransaction(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	transaction, err := h.service.GetFASTagTransaction(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// UpdateFASTagTransaction updates an existing FASTag transaction
func (h *IndiaIntegrationHandlers) UpdateFASTagTransaction(c *gin.Context) {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.UpdateFASTagTransaction(ctx, id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// DeleteFASTagTransaction deletes a FASTag transaction
func (h *IndiaIntegrationHandlers) DeleteFASTagTransaction(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	if err := h.service.DeleteFASTagTransaction(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// ListFASTagTransactions retrieves FASTag transactions with filtering
func (h *IndiaIntegrationHandlers) ListFASTagTransactions(c *gin.Context) {
	vehicleIDStr := c.Query("vehicle_id")
	tagNumber := c.Query("tag_number")
	startDateString := c.Query("start_date")
	endDateString := c.Query("end_date")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	var vehicleID *uuid.UUID
	if vehicleIDStr != "" {
		parsedUUID, err := uuid.Parse(vehicleIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vehicle ID"})
			return
		}
		vehicleID = &parsedUUID
	}

	var startDate, endDate *time.Time
	if startDateString != "" {
		parsedDate, err := time.Parse(time.RFC3339, startDateString)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format, use RFC3339"})
			return
		}
		startDate = &parsedDate
	}

	if endDateString != "" {
		parsedDate, err := time.Parse(time.RFC3339, endDateString)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format, use RFC3339"})
			return
		}
		endDate = &parsedDate
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	ctx := c.Request.Context()
	transactions, totalCount, err := h.service.ListFASTagTransactions(ctx, vehicleID, &tagNumber, startDate, endDate, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        transactions,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// CreateEWayBill creates a new E-Way Bill
func (h *IndiaIntegrationHandlers) CreateEWayBill(c *gin.Context) {
	var bill models.EWayBill
	if err := c.ShouldBindJSON(&bill); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.CreateEWayBill(ctx, &bill); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, bill)
}

// GetEWayBill retrieves an E-Way Bill by ID
func (h *IndiaIntegrationHandlers) GetEWayBill(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	bill, err := h.service.GetEWayBill(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bill)
}

// GetEWayBillByNumber retrieves an E-Way Bill by its number
func (h *IndiaIntegrationHandlers) GetEWayBillByNumber(c *gin.Context) {
	number := c.Param("number")

	ctx := c.Request.Context()
	bill, err := h.service.GetEWayBillByNumber(ctx, number)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bill)
}

// UpdateEWayBill updates an existing E-Way Bill
func (h *IndiaIntegrationHandlers) UpdateEWayBill(c *gin.Context) {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.UpdateEWayBill(ctx, id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// DeleteEWayBill deletes an E-Way Bill
func (h *IndiaIntegrationHandlers) DeleteEWayBill(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	if err := h.service.DeleteEWayBill(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// ListEWayBills retrieves E-Way Bills with filtering
func (h *IndiaIntegrationHandlers) ListEWayBills(c *gin.Context) {
	vehicleNumber := c.Query("vehicle_number")
	transporterID := c.Query("transporter_id")
	startDateString := c.Query("start_date")
	endDateString := c.Query("end_date")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	var startDate, endDate *time.Time
	if startDateString != "" {
		parsedDate, err := time.Parse(time.RFC3339, startDateString)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format, use RFC3339"})
			return
		}
		startDate = &parsedDate
	}

	if endDateString != "" {
		parsedDate, err := time.Parse(time.RFC3339, endDateString)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format, use RFC3339"})
			return
		}
		endDate = &parsedDate
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	ctx := c.Request.Context()
	bills, totalCount, err := h.service.ListEWayBills(ctx, &vehicleNumber, &transporterID, startDate, endDate, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        bills,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// CreateVehicleVerification creates a new vehicle verification
func (h *IndiaIntegrationHandlers) CreateVehicleVerification(c *gin.Context) {
	var verification models.VehicleVerification
	if err := c.ShouldBindJSON(&verification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.CreateVehicleVerification(ctx, &verification); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, verification)
}

// GetVehicleVerification retrieves a vehicle verification by ID
func (h *IndiaIntegrationHandlers) GetVehicleVerification(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	verification, err := h.service.GetVehicleVerification(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, verification)
}

// GetVehicleVerificationByRegNumber retrieves a vehicle verification by registration number
func (h *IndiaIntegrationHandlers) GetVehicleVerificationByRegNumber(c *gin.Context) {
	regNumber := c.Param("reg_number")

	ctx := c.Request.Context()
	verification, err := h.service.GetVehicleVerificationByRegNumber(ctx, regNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, verification)
}

// UpdateVehicleVerification updates an existing vehicle verification
func (h *IndiaIntegrationHandlers) UpdateVehicleVerification(c *gin.Context) {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.UpdateVehicleVerification(ctx, id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// DeleteVehicleVerification deletes a vehicle verification
func (h *IndiaIntegrationHandlers) DeleteVehicleVerification(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	if err := h.service.DeleteVehicleVerification(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// ListVehicleVerifications retrieves vehicle verifications with filtering
func (h *IndiaIntegrationHandlers) ListVehicleVerifications(c *gin.Context) {
	regNumber := c.Query("reg_number")
	status := c.Query("status")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	if limitStr == "" {
		limitStr = "10"
	}
	if offsetStr == "" {
		offsetStr = "0"
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	ctx := c.Request.Context()
	verifications, totalCount, err := h.service.ListVehicleVerifications(ctx, &regNumber, &status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        verifications,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// CreateDriverVerification creates a new driver verification
func (h *IndiaIntegrationHandlers) CreateDriverVerification(c *gin.Context) {
	var verification models.DriverVerification
	if err := c.ShouldBindJSON(&verification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.CreateDriverVerification(ctx, &verification); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, verification)
}

// GetDriverVerification retrieves a driver verification by ID
func (h *IndiaIntegrationHandlers) GetDriverVerification(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	verification, err := h.service.GetDriverVerification(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, verification)
}

// GetDriverVerificationByLicenseNumber retrieves a driver verification by license number
func (h *IndiaIntegrationHandlers) GetDriverVerificationByLicenseNumber(c *gin.Context) {
	licenseNumber := c.Param("license_number")

	ctx := c.Request.Context()
	verification, err := h.service.GetDriverVerificationByLicenseNumber(ctx, licenseNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, verification)
}

// UpdateDriverVerification updates an existing driver verification
func (h *IndiaIntegrationHandlers) UpdateDriverVerification(c *gin.Context) {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.UpdateDriverVerification(ctx, id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// DeleteDriverVerification deletes a driver verification
func (h *IndiaIntegrationHandlers) DeleteDriverVerification(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	if err := h.service.DeleteDriverVerification(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// ListDriverVerifications retrieves driver verifications with filtering
func (h *IndiaIntegrationHandlers) ListDriverVerifications(c *gin.Context) {
	licenseNumber := c.Query("license_number")
	status := c.Query("status")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	if limitStr == "" {
		limitStr = "10"
	}
	if offsetStr == "" {
		offsetStr = "0"
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	ctx := c.Request.Context()
	verifications, totalCount, err := h.service.ListDriverVerifications(ctx, &licenseNumber, &status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        verifications,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// CreateLocalPermit creates a new local permit
func (h *IndiaIntegrationHandlers) CreateLocalPermit(c *gin.Context) {
	var permit models.LocalPermit
	if err := c.ShouldBindJSON(&permit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.CreateLocalPermit(ctx, &permit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, permit)
}

// GetLocalPermit retrieves a local permit by ID
func (h *IndiaIntegrationHandlers) GetLocalPermit(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	permit, err := h.service.GetLocalPermit(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permit)
}

// GetLocalPermitByNumber retrieves a local permit by permit number
func (h *IndiaIntegrationHandlers) GetLocalPermitByNumber(c *gin.Context) {
	permitNumber := c.Param("permit_number")

	ctx := c.Request.Context()
	permit, err := h.service.GetLocalPermitByNumber(ctx, permitNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permit)
}

// UpdateLocalPermit updates an existing local permit
func (h *IndiaIntegrationHandlers) UpdateLocalPermit(c *gin.Context) {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.UpdateLocalPermit(ctx, id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// DeleteLocalPermit deletes a local permit
func (h *IndiaIntegrationHandlers) DeleteLocalPermit(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	if err := h.service.DeleteLocalPermit(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// ListLocalPermits retrieves local permits with filtering
func (h *IndiaIntegrationHandlers) ListLocalPermits(c *gin.Context) {
	vehicleIDStr := c.Query("vehicle_id")
	permitType := c.Query("permit_type")
	status := c.Query("status")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	var vehicleID *uuid.UUID
	if vehicleIDStr != "" {
		parsedUUID, err := uuid.Parse(vehicleIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vehicle ID"})
			return
		}
		vehicleID = &parsedUUID
	}

	if limitStr == "" {
		limitStr = "10"
	}
	if offsetStr == "" {
		offsetStr = "0"
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	ctx := c.Request.Context()
	permits, totalCount, err := h.service.ListLocalPermits(ctx, vehicleID, &permitType, &status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        permits,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// CreateGSTCompliance creates a new GST compliance record
func (h *IndiaIntegrationHandlers) CreateGSTCompliance(c *gin.Context) {
	var compliance models.GSTCompliance
	if err := c.ShouldBindJSON(&compliance); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.CreateGSTCompliance(ctx, &compliance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, compliance)
}

// GetGSTCompliance retrieves a GST compliance record by ID
func (h *IndiaIntegrationHandlers) GetGSTCompliance(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	compliance, err := h.service.GetGSTCompliance(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, compliance)
}

// GetGSTComplianceByGSTIN retrieves a GST compliance record by GSTIN
func (h *IndiaIntegrationHandlers) GetGSTComplianceByGSTIN(c *gin.Context) {
	gstin := c.Param("gstin")

	ctx := c.Request.Context()
	compliance, err := h.service.GetGSTComplianceByGSTIN(ctx, gstin)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, compliance)
}

// UpdateGSTCompliance updates an existing GST compliance record
func (h *IndiaIntegrationHandlers) UpdateGSTCompliance(c *gin.Context) {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.UpdateGSTCompliance(ctx, id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// DeleteGSTCompliance deletes a GST compliance record
func (h *IndiaIntegrationHandlers) DeleteGSTCompliance(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	if err := h.service.DeleteGSTCompliance(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// ListGSTCompliances retrieves GST compliance records with filtering
func (h *IndiaIntegrationHandlers) ListGSTCompliances(c *gin.Context) {
	orgIDStr := c.Query("org_id")
	status := c.Query("status")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}

	if limitStr == "" {
		limitStr = "10"
	}
	if offsetStr == "" {
		offsetStr = "0"
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	ctx := c.Request.Context()
	compliances, totalCount, err := h.service.ListGSTCompliances(ctx, orgID, &status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        compliances,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// CreateIntegrationConfig creates a new integration configuration
func (h *IndiaIntegrationHandlers) CreateIntegrationConfig(c *gin.Context) {
	var config models.IntegrationConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.CreateIntegrationConfig(ctx, &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, config)
}

// GetIntegrationConfig retrieves an integration configuration by ID
func (h *IndiaIntegrationHandlers) GetIntegrationConfig(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	config, err := h.service.GetIntegrationConfig(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// GetIntegrationConfigByServiceType retrieves an integration configuration by service type and organization
func (h *IndiaIntegrationHandlers) GetIntegrationConfigByServiceType(c *gin.Context) {
	orgIDStr := c.Param("org_id")
	serviceType := c.Param("service_type")

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}

	ctx := c.Request.Context()
	config, err := h.service.GetIntegrationConfigByServiceType(ctx, orgID, serviceType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateIntegrationConfig updates an existing integration configuration
func (h *IndiaIntegrationHandlers) UpdateIntegrationConfig(c *gin.Context) {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.UpdateIntegrationConfig(ctx, id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// DeleteIntegrationConfig deletes an integration configuration
func (h *IndiaIntegrationHandlers) DeleteIntegrationConfig(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	if err := h.service.DeleteIntegrationConfig(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// ListIntegrationConfigs retrieves integration configurations with filtering
func (h *IndiaIntegrationHandlers) ListIntegrationConfigs(c *gin.Context) {
	orgIDStr := c.Query("org_id")
	serviceType := c.Query("service_type")
	activeStr := c.Query("active")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}

	var active *bool
	if activeStr != "" {
		parsedActive, err := strconv.ParseBool(activeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid active parameter"})
			return
		}
		active = &parsedActive
	}

	if limitStr == "" {
		limitStr = "10"
	}
	if offsetStr == "" {
		offsetStr = "0"
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	ctx := c.Request.Context()
	configs, totalCount, err := h.service.ListIntegrationConfigs(ctx, orgID, &serviceType, active, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        configs,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// HealthCheck performs a health check on the service
func (h *IndiaIntegrationHandlers) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	if err := h.service.HealthCheck(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "unhealthy", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
func (h *IndiaIntegrationHandlers) SyncFASTag(c *gin.Context) {
c.JSON(http.StatusOK, gin.H{
"message": "FASTag sync initiated",
"data":    gin.H{"status": "processing"},
})
}

func (h *IndiaIntegrationHandlers) ValidateEWayBill(c *gin.Context) {
c.JSON(http.StatusOK, gin.H{
"message": "E-Way bill validation initiated",
"data":    gin.H{"status": "processing"},
})
}

func (h *IndiaIntegrationHandlers) VerifyVehicle(c *gin.Context) {
c.JSON(http.StatusOK, gin.H{
"message": "Vehicle verification initiated",
"data":    gin.H{"status": "processing"},
})
}

func (h *IndiaIntegrationHandlers) VerifyDriver(c *gin.Context) {
c.JSON(http.StatusOK, gin.H{
"message": "Driver verification initiated",
"data":    gin.H{"status": "processing"},
})
}
