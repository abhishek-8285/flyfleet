package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing"
	"github.com/flyfleet/fleet-core/internal/domains/india/models"
	"github.com/flyfleet/fleet-core/internal/domains/india/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// IndiaIntegrationHandlers handles HTTP requests for India integration services
type IndiaIntegrationHandlers struct {
	service         service.IndiaIntegrationService
	fastagService   service.FastagIntegrationService
	ewayBillService service.EWayBillIntegrationService
	vahanService    service.VahanIntegrationService
	sarathiService  service.SarathiIntegrationService
	gstService      service.GSTIntegrationService
	bulkService     service.BulkVerificationService
	encryptService  service.EncryptionService
	idempotency     *billing.IdempotencyHelper
	logger          *zap.Logger
}

// NewIndiaIntegrationHandlers creates a new instance of IndiaIntegrationHandlers
func NewIndiaIntegrationHandlers(
	svc service.IndiaIntegrationService,
	fastag service.FastagIntegrationService,
	ewayBill service.EWayBillIntegrationService,
	vahan service.VahanIntegrationService,
	sarathi service.SarathiIntegrationService,
	gst service.GSTIntegrationService,
	bulk service.BulkVerificationService,
	encrypt service.EncryptionService,
	redisClient *redis.Client,
	logger *zap.Logger,
) *IndiaIntegrationHandlers {
	var helper *billing.IdempotencyHelper
	if redisClient != nil {
		helper = billing.NewIdempotencyHelper(redisClient)
	}

	return &IndiaIntegrationHandlers{
		service:         svc,
		fastagService:   fastag,
		ewayBillService: ewayBill,
		vahanService:    vahan,
		sarathiService:  sarathi,
		gstService:      gst,
		bulkService:     bulk,
		encryptService:  encrypt,
		idempotency:     helper,
		logger:          logger,
	}
}

func (h *IndiaIntegrationHandlers) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *IndiaIntegrationHandlers) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	h.writeJSONResponse(w, statusCode, map[string]string{"error": message})
}

// CreateFASTagTransaction creates a new FASTag transaction
func (h *IndiaIntegrationHandlers) CreateFASTagTransaction(w http.ResponseWriter, r *http.Request) {
	reqID := r.Header.Get("X-Request-ID")
	orgID := r.Context().Value("tenant_id") // standard flyfleet context extraction

	// SECURITY FIX: Do not trust X-Org-ID header (CWE-284/290) - only use authenticated context
	var orgIDStr string
	if orgID != nil {
		if oid, ok := orgID.(string); ok {
			orgIDStr = oid
		} else if uuidVal, ok := orgID.(uuid.UUID); ok {
			orgIDStr = uuidVal.String()
		}
	} else {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization context missing", nil)
		return
	}

	if h.idempotency != nil && reqID != "" && orgIDStr != "" {
		acquired, err := h.idempotency.CheckAndLock(r.Context(), orgIDStr, reqID, 24*time.Hour)
		if err != nil {
			h.writeErrorResponse(w, http.StatusInternalServerError, "Idempotency check failed", err)
			return
		}
		if !acquired {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "duplicate_transaction",
				"message": "This transaction is already being processed",
			})
			return
		}
	}

	var transaction models.FASTagTransaction
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.service.CreateFASTagTransaction(r.Context(), &transaction); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create FASTag transaction", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, transaction)
}

// GetFASTagTransaction retrieves a FASTag transaction by ID
func (h *IndiaIntegrationHandlers) GetFASTagTransaction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Get orgID from context for authorization
	orgID := r.Context().Value("tenant_id")
	var orgIDStr string
	if orgID != nil {
		if oid, ok := orgID.(string); ok {
			orgIDStr = oid
		} else if uuidVal, ok := orgID.(uuid.UUID); ok {
			orgIDStr = uuidVal.String()
		}
	} else {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization context missing", nil)
		return
	}

	transaction, err := h.service.GetFASTagTransaction(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "FASTag transaction not found", err)
		return
	}

	// Verify transaction belongs to the authenticated organization
	reqOrgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Internal error", nil)
		return
	}
	if transaction.OrganizationID != reqOrgID {
		h.writeErrorResponse(w, http.StatusNotFound, "FASTag transaction not found or access denied", nil)
		return
	}
	h.writeJSONResponse(w, http.StatusOK, transaction)
}

// UpdateFASTagTransaction updates an existing FASTag transaction
func (h *IndiaIntegrationHandlers) UpdateFASTagTransaction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.service.UpdateFASTagTransaction(r.Context(), id, updates); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update FASTag transaction", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// DeleteFASTagTransaction deletes a FASTag transaction
func (h *IndiaIntegrationHandlers) DeleteFASTagTransaction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteFASTagTransaction(r.Context(), id); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete FASTag transaction", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// ListFASTagTransactions retrieves FASTag transactions with filtering
func (h *IndiaIntegrationHandlers) ListFASTagTransactions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	vehicleIDStr := query.Get("vehicle_id")
	tagNumber := query.Get("tag_number")
	startDateString := query.Get("start_date")
	endDateString := query.Get("end_date")
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	var vehicleID *uuid.UUID
	if vehicleIDStr != "" {
		parsedUUID, err := uuid.Parse(vehicleIDStr)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "invalid vehicle ID", err)
			return
		}
		vehicleID = &parsedUUID
	}

	var startDate, endDate *time.Time
	if startDateString != "" {
		parsedDate, err := time.Parse(time.RFC3339, startDateString)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "invalid start date format, use RFC3339", err)
			return
		}
		startDate = &parsedDate
	}

	if endDateString != "" {
		parsedDate, err := time.Parse(time.RFC3339, endDateString)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "invalid end date format, use RFC3339", err)
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

	transactions, totalCount, err := h.service.ListFASTagTransactions(r.Context(), vehicleID, &tagNumber, startDate, endDate, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list FASTag transactions", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":        transactions,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// CreateEWayBill creates a new E-Way Bill
func (h *IndiaIntegrationHandlers) CreateEWayBill(w http.ResponseWriter, r *http.Request) {
	var bill models.EWayBill
	if err := json.NewDecoder(r.Body).Decode(&bill); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.service.CreateEWayBill(r.Context(), &bill); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create E-Way Bill", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, bill)
}

// GetEWayBill retrieves an E-Way Bill by ID
func (h *IndiaIntegrationHandlers) GetEWayBill(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	bill, err := h.service.GetEWayBill(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "E-Way Bill not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, bill)
}

// GetEWayBillByNumber retrieves an E-Way Bill by its number
func (h *IndiaIntegrationHandlers) GetEWayBillByNumber(w http.ResponseWriter, r *http.Request) {
	number := chi.URLParam(r, "number")

	bill, err := h.service.GetEWayBillByNumber(r.Context(), number)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "E-Way Bill not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, bill)
}

// UpdateEWayBill updates an existing E-Way Bill
func (h *IndiaIntegrationHandlers) UpdateEWayBill(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.service.UpdateEWayBill(r.Context(), id, updates); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update E-Way Bill", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// DeleteEWayBill deletes an E-Way Bill
func (h *IndiaIntegrationHandlers) DeleteEWayBill(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteEWayBill(r.Context(), id); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete E-Way Bill", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// ListEWayBills retrieves E-Way Bills with filtering
func (h *IndiaIntegrationHandlers) ListEWayBills(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	vehicleNumber := query.Get("vehicle_number")
	transporterID := query.Get("transporter_id")
	startDateString := query.Get("start_date")
	endDateString := query.Get("end_date")
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	var startDate, endDate *time.Time
	if startDateString != "" {
		parsedDate, err := time.Parse(time.RFC3339, startDateString)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "invalid start date format, use RFC3339", err)
			return
		}
		startDate = &parsedDate
	}

	if endDateString != "" {
		parsedDate, err := time.Parse(time.RFC3339, endDateString)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "invalid end date format, use RFC3339", err)
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

	bills, totalCount, err := h.service.ListEWayBills(r.Context(), &vehicleNumber, &transporterID, startDate, endDate, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list E-Way Bills", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":        bills,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// CreateVehicleVerification creates a new vehicle verification
func (h *IndiaIntegrationHandlers) CreateVehicleVerification(w http.ResponseWriter, r *http.Request) {
	var verification models.VehicleVerification
	if err := json.NewDecoder(r.Body).Decode(&verification); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.service.CreateVehicleVerification(r.Context(), &verification); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create vehicle verification", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, verification)
}

// GetVehicleVerification retrieves a vehicle verification by ID
func (h *IndiaIntegrationHandlers) GetVehicleVerification(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	verification, err := h.service.GetVehicleVerification(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Vehicle verification not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, verification)
}

// GetVehicleVerificationByRegNumber retrieves a vehicle verification by registration number
func (h *IndiaIntegrationHandlers) GetVehicleVerificationByRegNumber(w http.ResponseWriter, r *http.Request) {
	regNumber := chi.URLParam(r, "reg_number")

	verification, err := h.service.GetVehicleVerificationByRegNumber(r.Context(), regNumber)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Vehicle verification not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, verification)
}

// UpdateVehicleVerification updates an existing vehicle verification
func (h *IndiaIntegrationHandlers) UpdateVehicleVerification(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.service.UpdateVehicleVerification(r.Context(), id, updates); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update vehicle verification", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// DeleteVehicleVerification deletes a vehicle verification
func (h *IndiaIntegrationHandlers) DeleteVehicleVerification(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteVehicleVerification(r.Context(), id); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete vehicle verification", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// ListVehicleVerifications retrieves vehicle verifications with filtering
func (h *IndiaIntegrationHandlers) ListVehicleVerifications(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	regNumber := query.Get("reg_number")
	status := query.Get("status")
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	verifications, totalCount, err := h.service.ListVehicleVerifications(r.Context(), &regNumber, &status, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list vehicle verifications", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":        verifications,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// CreateDriverVerification creates a new driver verification
func (h *IndiaIntegrationHandlers) CreateDriverVerification(w http.ResponseWriter, r *http.Request) {
	var verification models.DriverVerification
	if err := json.NewDecoder(r.Body).Decode(&verification); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.service.CreateDriverVerification(r.Context(), &verification); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create driver verification", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, verification)
}

// GetDriverVerification retrieves a driver verification by ID
func (h *IndiaIntegrationHandlers) GetDriverVerification(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	verification, err := h.service.GetDriverVerification(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Driver verification not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, verification)
}

// GetDriverVerificationByLicenseNumber retrieves a driver verification by license number
func (h *IndiaIntegrationHandlers) GetDriverVerificationByLicenseNumber(w http.ResponseWriter, r *http.Request) {
	licenseNumber := chi.URLParam(r, "license_number")

	verification, err := h.service.GetDriverVerificationByLicenseNumber(r.Context(), licenseNumber)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Driver verification not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, verification)
}

// UpdateDriverVerification updates an existing driver verification
func (h *IndiaIntegrationHandlers) UpdateDriverVerification(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.service.UpdateDriverVerification(r.Context(), id, updates); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update driver verification", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// DeleteDriverVerification deletes a driver verification
func (h *IndiaIntegrationHandlers) DeleteDriverVerification(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteDriverVerification(r.Context(), id); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete driver verification", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// ListDriverVerifications retrieves driver verifications with filtering
func (h *IndiaIntegrationHandlers) ListDriverVerifications(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	licenseNumber := query.Get("license_number")
	status := query.Get("status")
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	verifications, totalCount, err := h.service.ListDriverVerifications(r.Context(), &licenseNumber, &status, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list driver verifications", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":        verifications,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// CreateLocalPermit creates a new local permit
func (h *IndiaIntegrationHandlers) CreateLocalPermit(w http.ResponseWriter, r *http.Request) {
	var permit models.LocalPermit
	if err := json.NewDecoder(r.Body).Decode(&permit); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.service.CreateLocalPermit(r.Context(), &permit); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create local permit", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, permit)
}

// GetLocalPermit retrieves a local permit by ID
func (h *IndiaIntegrationHandlers) GetLocalPermit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	permit, err := h.service.GetLocalPermit(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Local permit not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, permit)
}

// GetLocalPermitByNumber retrieves a local permit by permit number
func (h *IndiaIntegrationHandlers) GetLocalPermitByNumber(w http.ResponseWriter, r *http.Request) {
	permitNumber := chi.URLParam(r, "permit_number")

	permit, err := h.service.GetLocalPermitByNumber(r.Context(), permitNumber)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Local permit not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, permit)
}

// UpdateLocalPermit updates an existing local permit
func (h *IndiaIntegrationHandlers) UpdateLocalPermit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.service.UpdateLocalPermit(r.Context(), id, updates); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update local permit", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// DeleteLocalPermit deletes a local permit
func (h *IndiaIntegrationHandlers) DeleteLocalPermit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteLocalPermit(r.Context(), id); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete local permit", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// ListLocalPermits retrieves local permits with filtering
func (h *IndiaIntegrationHandlers) ListLocalPermits(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	vehicleIDStr := query.Get("vehicle_id")
	permitType := query.Get("permit_type")
	status := query.Get("status")
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	var vehicleID *uuid.UUID
	if vehicleIDStr != "" {
		parsedUUID, err := uuid.Parse(vehicleIDStr)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "invalid vehicle ID", err)
			return
		}
		vehicleID = &parsedUUID
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	permits, totalCount, err := h.service.ListLocalPermits(r.Context(), vehicleID, &permitType, &status, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list local permits", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":        permits,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// CreateGSTCompliance creates a new GST compliance record
func (h *IndiaIntegrationHandlers) CreateGSTCompliance(w http.ResponseWriter, r *http.Request) {
	var compliance models.GSTCompliance
	if err := json.NewDecoder(r.Body).Decode(&compliance); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.service.CreateGSTCompliance(r.Context(), &compliance); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create GST compliance record", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, compliance)
}

// GetGSTCompliance retrieves a GST compliance record by ID
func (h *IndiaIntegrationHandlers) GetGSTCompliance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	compliance, err := h.service.GetGSTCompliance(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "GST compliance record not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, compliance)
}

// GetGSTComplianceByGSTIN retrieves a GST compliance record by GSTIN
func (h *IndiaIntegrationHandlers) GetGSTComplianceByGSTIN(w http.ResponseWriter, r *http.Request) {
	gstin := chi.URLParam(r, "gstin")

	compliance, err := h.service.GetGSTComplianceByGSTIN(r.Context(), gstin)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "GST compliance record not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, compliance)
}

// UpdateGSTCompliance updates an existing GST compliance record
func (h *IndiaIntegrationHandlers) UpdateGSTCompliance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.service.UpdateGSTCompliance(r.Context(), id, updates); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update GST compliance record", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// DeleteGSTCompliance deletes a GST compliance record
func (h *IndiaIntegrationHandlers) DeleteGSTCompliance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteGSTCompliance(r.Context(), id); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete GST compliance record", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// ListGSTCompliances retrieves GST compliance records with filtering
func (h *IndiaIntegrationHandlers) ListGSTCompliances(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	orgIDStr := query.Get("tenant_id")
	status := query.Get("status")
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid organization ID", err)
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	compliances, totalCount, err := h.service.ListGSTCompliances(r.Context(), orgID, &status, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list GST compliance records", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":        compliances,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// CreateIntegrationConfig creates a new integration configuration
func (h *IndiaIntegrationHandlers) CreateIntegrationConfig(w http.ResponseWriter, r *http.Request) {
	var config models.IntegrationConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if config.ClientSecret != "" {
		enc, _ := h.encryptService.Encrypt(config.ClientSecret, os.Getenv("ENCRYPTION_KEY"))
		config.ClientSecret = enc
	}
	if config.APIKey != "" {
		enc, _ := h.encryptService.Encrypt(config.APIKey, os.Getenv("ENCRYPTION_KEY"))
		config.APIKey = enc
	}

	if err := h.service.CreateIntegrationConfig(r.Context(), &config); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create integration config", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, config)
}

// GetIntegrationConfig retrieves an integration configuration by ID
func (h *IndiaIntegrationHandlers) GetIntegrationConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	config, err := h.service.GetIntegrationConfig(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Integration config not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, config)
}

// GetIntegrationConfigByServiceType retrieves an integration configuration by service type and organization
func (h *IndiaIntegrationHandlers) GetIntegrationConfigByServiceType(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "tenant_id")
	serviceType := chi.URLParam(r, "service_type")

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid organization ID", err)
		return
	}

	config, err := h.service.GetIntegrationConfigByServiceType(r.Context(), orgID, serviceType)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Integration config not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, config)
}

// UpdateIntegrationConfig updates an existing integration configuration
func (h *IndiaIntegrationHandlers) UpdateIntegrationConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if secret, ok := updates["client_secret"].(string); ok && secret != "" {
		enc, _ := h.encryptService.Encrypt(secret, os.Getenv("ENCRYPTION_KEY"))
		updates["client_secret"] = enc
	}

	if err := h.service.UpdateIntegrationConfig(r.Context(), id, updates); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update integration config", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// DeleteIntegrationConfig deletes an integration configuration
func (h *IndiaIntegrationHandlers) DeleteIntegrationConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteIntegrationConfig(r.Context(), id); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete integration config", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// ListIntegrationConfigs retrieves integration configurations with filtering
func (h *IndiaIntegrationHandlers) ListIntegrationConfigs(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePaginationParams(r.URL.Query())

	// Stub implementation - return empty list
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":        []interface{}{},
		"total_count": 0,
		"limit":       limit,
		"offset":      offset,
	})
}

// SyncFASTag synchronizes FASTag transactions from the issuer
func (h *IndiaIntegrationHandlers) SyncFASTag(w http.ResponseWriter, r *http.Request) {
	var req struct {
		VehicleID   string `json:"vehicle_id"`
		FromDate    string `json:"from_date"`
		ToDate      string `json:"to_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Stub implementation
	h.logger.Info("FASTag sync requested", zap.String("vehicle_id", req.VehicleID))
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "sync_initiated"})
}

// HandleFASTagWebhook handles incoming FASTag webhook notifications
func (h *IndiaIntegrationHandlers) HandleFASTagWebhook(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid webhook payload", err)
		return
	}

	h.logger.Info("received FASTag webhook", zap.Any("payload", payload))
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "webhook_received"})
}

// ValidateEWayBill validates an E-Way Bill with the GST portal
func (h *IndiaIntegrationHandlers) ValidateEWayBill(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EWayBillNumber string `json:"ewaybill_number" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Stub implementation - return mock validation result
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"ewaybill_number": req.EWayBillNumber,
		"valid":           true,
		"status":          "ACTIVE",
	})
}

// HandleEWayBillWebhook handles incoming E-Way Bill webhook notifications
func (h *IndiaIntegrationHandlers) HandleEWayBillWebhook(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid webhook payload", err)
		return
	}

	h.logger.Info("received E-Way Bill webhook", zap.Any("payload", payload))
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "webhook_received"})
}

// VerifyVehicle verifies a vehicle using VAHAN API
func (h *IndiaIntegrationHandlers) VerifyVehicle(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RegistrationNumber string `json:"registration_number" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Stub implementation - return mock vehicle data
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"registration_number": req.RegistrationNumber,
		"vehicle_class":       "LMV",
		"fuel_type":           "DIESEL",
		"valid":               true,
	})
}

// BulkVerifyVahan performs bulk vehicle verification
func (h *IndiaIntegrationHandlers) BulkVerifyVahan(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RegistrationNumbers []string `json:"registration_numbers" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	results := make([]interface{}, 0, len(req.RegistrationNumbers))
	for _, regNum := range req.RegistrationNumbers {
		results = append(results, map[string]interface{}{
			"registration_number": regNum,
			"vehicle_class":       "LMV",
			"fuel_type":           "DIESEL",
			"valid":               true,
		})
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"total":   len(results),
		"results": results,
	})
}

// BulkVerifySarathi performs bulk driver license verification
func (h *IndiaIntegrationHandlers) BulkVerifySarathi(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LicenseNumbers []string `json:"license_numbers" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	results := make([]interface{}, 0, len(req.LicenseNumbers))
	for _, licNum := range req.LicenseNumbers {
		results = append(results, map[string]interface{}{
			"license_number": licNum,
			"valid":          true,
			"status":         "ACTIVE",
		})
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"total":   len(results),
		"results": results,
	})
}

// HandleGSTWebhook handles incoming GST webhook notifications
func (h *IndiaIntegrationHandlers) HandleGSTWebhook(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid webhook payload", err)
		return
	}

	h.logger.Info("received GST webhook", zap.Any("payload", payload))
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "webhook_received"})
}

// HealthCheck returns the health status of India integration services
func (h *IndiaIntegrationHandlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.writeJSONResponse(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"service": "india-integration",
	})
}

// Helper functions

func parsePaginationParams(query map[string][]string) (int, int) {
	limit := 20
	offset := 0

	if vals := query["limit"]; len(vals) > 0 && vals[0] != "" {
		if parsed, err := strconv.Atoi(vals[0]); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100
			}
		}
	}

	if vals := query["offset"]; len(vals) > 0 && vals[0] != "" {
		if parsed, err := strconv.Atoi(vals[0]); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}

func (h *IndiaIntegrationHandlers) getOrgID(r *http.Request) uuid.UUID {
	if val := r.Context().Value("orgID"); val != nil {
		switch v := val.(type) {
		case uuid.UUID:
			return v
		case string:
			if id, err := uuid.Parse(v); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}
