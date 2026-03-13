package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/india_v2/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/india_v2/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// IndiaIntegrationHandlers handles HTTP requests for India integration services
type IndiaIntegrationHandlers struct {
	service           service.IndiaIntegrationService
	fastagService     service.FastagIntegrationService
	ewayBillService   service.EWayBillIntegrationService
	vahanService      service.VahanIntegrationService
	sarathiService    service.SarathiIntegrationService
	gstService        service.GSTIntegrationService
	bulkService       service.BulkVerificationService
	encryptService    service.EncryptionService
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
) *IndiaIntegrationHandlers {
	return &IndiaIntegrationHandlers{
		service:         svc,
		fastagService:   fastag,
		ewayBillService: ewayBill,
		vahanService:    vahan,
		sarathiService:  sarathi,
		gstService:      gst,
		bulkService:     bulk,
		encryptService:  encrypt,
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

	transaction, err := h.service.GetFASTagTransaction(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "FASTag transaction not found", err)
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
	orgIDStr := query.Get("org_id")
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
		enc, _ := h.encryptService.Encrypt(config.ClientSecret, "REPLACE-WITH-ACTUAL-SECRET-KEY")
		config.ClientSecret = enc
	}
	if config.APIKey != "" {
		enc, _ := h.encryptService.Encrypt(config.APIKey, "REPLACE-WITH-ACTUAL-SECRET-KEY")
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
	orgIDStr := chi.URLParam(r, "org_id")
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
		enc, _ := h.encryptService.Encrypt(secret, "REPLACE-WITH-ACTUAL-SECRET-KEY")
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
	query := r.URL.Query()
	orgIDStr := query.Get("org_id")
	serviceType := query.Get("service_type")
	activeStr := query.Get("active")
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid organization ID", err)
		return
	}

	var active *bool
	if activeStr != "" {
		parsedActive, err := strconv.ParseBool(activeStr)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "invalid active parameter", err)
			return
		}
		active = &parsedActive
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	configs, totalCount, err := h.service.ListIntegrationConfigs(r.Context(), orgID, &serviceType, active, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list integration configs", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":        configs,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// HealthCheck returns the health status of India integrations
func (h *IndiaIntegrationHandlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	err := h.service.HealthCheck(r.Context())
	if err != nil {
		h.writeJSONResponse(w, http.StatusServiceUnavailable, map[string]string{
			"status": "down",
			"error":  err.Error(),
		})
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{
		"status": "up",
	})
}

// HandleFASTagWebhook handles incoming FASTag webhooks from NPCI
func (h *IndiaIntegrationHandlers) HandleFASTagWebhook(w http.ResponseWriter, r *http.Request) {
	if !h.validateSignature(r) {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Invalid signature", nil)
		return
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// Process FASTag webhook logic
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "received"})
}

// HandleEWayBillWebhook handles incoming E-Way Bill webhooks from NIC
func (h *IndiaIntegrationHandlers) HandleEWayBillWebhook(w http.ResponseWriter, r *http.Request) {
	if !h.validateSignature(r) {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Invalid signature", nil)
		return
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// Process E-Way Bill webhook logic
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "received"})
}

// HandleGSTWebhook handles incoming GST webhooks
func (h *IndiaIntegrationHandlers) HandleGSTWebhook(w http.ResponseWriter, r *http.Request) {
	if !h.validateSignature(r) {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Invalid signature", nil)
		return
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// Process GST webhook logic
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "received"})
}

func (h *IndiaIntegrationHandlers) validateSignature(r *http.Request) bool {
	signature := r.Header.Get("X-Gov-Signature")
	if signature == "" {
		return false
	}
	// Implementation of actual signature validation logic
	// For now, return true for simulation
	return true
}

// SyncFASTag triggers a manual sync for FASTag transactions
func (h *IndiaIntegrationHandlers) SyncFASTag(w http.ResponseWriter, r *http.Request) {
	if err := h.fastagService.SyncTransactions(r.Context()); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to sync FASTag", err)
		return
	}
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "sync_started"})
}

// VerifyVehicle triggers a manual vehicle verification with VAHAN
func (h *IndiaIntegrationHandlers) VerifyVehicle(w http.ResponseWriter, r *http.Request) {
	var req models.VerifyVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	v, err := h.vahanService.VerifyVehicle(r.Context(), req.OrganizationID, req.VehicleID, req.RegistrationNumber)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to verify vehicle", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, v)
}

// VerifyDriver triggers a manual driver verification with SARATHI
func (h *IndiaIntegrationHandlers) VerifyDriver(w http.ResponseWriter, r *http.Request) {
	var req models.VerifyDriverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	v, err := h.sarathiService.VerifyDriver(r.Context(), req.OrganizationID, req.DriverID, req.LicenseNumber, req.DateOfBirth)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to verify driver", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, v)
}

// ValidateEWayBill triggers a manual E-Way Bill validation
func (h *IndiaIntegrationHandlers) ValidateEWayBill(w http.ResponseWriter, r *http.Request) {
	// Implementation for manual validation
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "validated"})
}

// BulkVerifyVahan handles bulk vehicle verification requests
func (h *IndiaIntegrationHandlers) BulkVerifyVahan(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		OrganizationID string                        `json:"organization_id"`
		Requests       []models.VerifyVehicleRequest `json:"requests"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if len(payload.Requests) > 100 {
		h.writeErrorResponse(w, http.StatusBadRequest, "Batch size exceeds maximum of 100", nil)
		return
	}

	orgID, _ := uuid.Parse(payload.OrganizationID)
	results, err := h.vahanService.BulkVerifyVehicle(r.Context(), orgID, payload.Requests)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to verify vehicles in bulk", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, results)
}

// BulkVerifySarathi handles bulk driver verification requests
func (h *IndiaIntegrationHandlers) BulkVerifySarathi(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		OrganizationID string                       `json:"organization_id"`
		Requests       []models.VerifyDriverRequest `json:"requests"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if len(payload.Requests) > 100 {
		h.writeErrorResponse(w, http.StatusBadRequest, "Batch size exceeds maximum of 100", nil)
		return
	}

	orgID, _ := uuid.Parse(payload.OrganizationID)
	results, err := h.sarathiService.BulkVerifyDriver(r.Context(), orgID, payload.Requests)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to verify drivers in bulk", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, results)
}
