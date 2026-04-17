package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fleet-platform/shared/pkg/pagination"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"

	driverMiddleware "github.com/flyfleet/fleet-core/internal/domains/driver/middleware"
	"github.com/flyfleet/fleet-core/internal/domains/driver/models"
	"github.com/flyfleet/fleet-core/internal/domains/driver/service"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
)

// DTOs for request payloads (separate from models to prevent mass assignment)
type createDriverRequestDTO struct {
	FirstName        string                  `json:"first_name" validate:"required,max=50"`
	LastName         string                  `json:"last_name" validate:"required,max=50"`
	Email            string                  `json:"email" validate:"required,email"`
	Phone            string                  `json:"phone" validate:"required,e164"`
	LicenseNumber    string                  `json:"license_number" validate:"required,max=50"`
	LicenseType      string                  `json:"license_type" validate:"required,oneof=LMV HMV Transport"`
	LicenseExpiry    time.Time               `json:"license_expiry" validate:"required"`
	LicenseAuthority string                  `json:"license_authority" validate:"required,max=100"`
	LicenseIssueDate *time.Time              `json:"license_issue_date" validate:"omitempty"`
	DateOfBirth      time.Time               `json:"date_of_birth" validate:"required"`
	EmployeeCode     string                  `json:"employee_code" validate:"omitempty,max=50"`
	AadhaarNumber    string                  `json:"aadhaar_number" validate:"omitempty,len=12"`
	PanNumber        string                  `json:"pan_number" validate:"omitempty,len=10"`
	EmploymentType   string                  `json:"employment_type" validate:"required,oneof=FULL_TIME CONTRACT"`
	Address          models.Address          `json:"address" validate:"required"`
	EmergencyContact models.EmergencyContact `json:"emergency_contact" validate:"required"`
	ExperienceYears  int                     `json:"experience_years" validate:"gte=0,lte=50"`
	Metadata         map[string]interface{}  `json:"metadata"`
}

type updateDriverRequestDTO struct {
	FirstName        *string                  `json:"first_name,omitempty" validate:"omitempty,max=50"`
	LastName         *string                  `json:"last_name,omitempty" validate:"omitempty,max=50"`
	Email            *string                  `json:"email,omitempty" validate:"omitempty,email"`
	Phone            *string                  `json:"phone,omitempty" validate:"omitempty,e164"`
	LicenseNumber    *string                  `json:"license_number,omitempty" validate:"omitempty,max=50"`
	LicenseType      *string                  `json:"license_type,omitempty" validate:"omitempty,oneof=LMV HMV Transport"`
	LicenseExpiry    *time.Time               `json:"license_expiry,omitempty" validate:"omitempty"`
	LicenseAuthority *string                  `json:"license_authority,omitempty" validate:"omitempty,max=100"`
	LicenseIssueDate *time.Time               `json:"license_issue_date,omitempty" validate:"omitempty"`
	DateOfBirth      *time.Time               `json:"date_of_birth,omitempty" validate:"omitempty"`
	EmployeeCode     *string                  `json:"employee_code,omitempty" validate:"omitempty,max=50"`
	AadhaarNumber    *string                  `json:"aadhaar_number,omitempty" validate:"omitempty,len=12"`
	PanNumber        *string                  `json:"pan_number,omitempty" validate:"omitempty,len=10"`
	EmploymentType   *string                  `json:"employment_type,omitempty" validate:"omitempty,oneof=FULL_TIME CONTRACT"`
	Address          *models.Address          `json:"address,omitempty" validate:"omitempty"`
	EmergencyContact *models.EmergencyContact `json:"emergency_contact,omitempty" validate:"omitempty"`
	ExperienceYears  *int                     `json:"experience_years,omitempty" validate:"omitempty,gte=0,lte=50"`
	Metadata         *map[string]interface{}  `json:"metadata,omitempty"`
}

type assignVehicleRequestDTO struct {
	VehicleID string `json:"vehicle_id" validate:"required,uuid"`
}

type updateStatusRequestDTO struct {
	Status string `json:"status" validate:"required,oneof=active inactive suspended resigned"`
}

type updateBehaviorScoreRequestDTO struct {
	ScoreChange int    `json:"score_change" validate:"required,gte=-20,lte=20"`
	Reason      string `json:"reason" validate:"required,max=200"`
}

type uploadDocumentRequestDTO struct {
	DocumentType     string    `json:"document_type" validate:"required,oneof=license aadhaar pan medical training other"`
	DocumentNumber   string    `json:"document_number" validate:"required,max=100"`
	IssuingAuthority string    `json:"issuing_authority" validate:"required,max=100"`
	IssueDate        time.Time `json:"issue_date" validate:"required"`
	ExpiryDate       time.Time `json:"expiry_date" validate:"required"`
	FileURL          string    `json:"file_url" validate:"required,url"`
}

type markAttendanceRequestDTO struct {
	Date         time.Time  `json:"date" validate:"required"`
	Status       string     `json:"status" validate:"required,oneof=present absent leave holiday"`
	CheckInTime  *time.Time `json:"check_in_time,omitempty"`
	CheckOutTime *time.Time `json:"check_out_time,omitempty"`
	Remarks      string     `json:"remarks,omitempty" validate:"max=200"`
}

type setScheduleRequestDTO struct {
	DayOfWeek int    `json:"day_of_week" validate:"required,min=0,max=6"`
	StartTime string `json:"start_time" validate:"required"`
	EndTime   string `json:"end_time" validate:"required"`
	IsWorking bool   `json:"is_working"`
}

type ratingDTO struct {
	DriverID         string  `json:"driver_id" validate:"required,uuid"`
	TripID           string  `json:"trip_id" validate:"required,uuid"`
	PunctualityScore int     `json:"punctuality_score"`
	BehaviorScore    int     `json:"behavior_score"`
	DrivingScore     int     `json:"driving_score"`
	OverallRating    float64 `json:"overall_rating"`
	Comments         string  `json:"comments"`
}

type DriverHandler struct {
	service    *service.DriverService
	onboarding *service.OnboardingService
	shift      *service.ShiftService
	training   *service.TrainingService
	payroll    *service.PayrollService
	rating     *service.RatingService
	safety     *service.SafetyEventService
	mobileSync *service.MobileSyncService
	security   *driverMiddleware.DriverSecurityMiddleware
	logger     *zap.Logger
	validate   *validator.Validate
}

func NewDriverHandler(
	service *service.DriverService,
	onboarding *service.OnboardingService,
	shift *service.ShiftService,
	training *service.TrainingService,
	payroll *service.PayrollService,
	rating *service.RatingService,
	safety *service.SafetyEventService,
	mobileSync *service.MobileSyncService,
	security *driverMiddleware.DriverSecurityMiddleware,
	logger *zap.Logger,
) *DriverHandler {
	return &DriverHandler{
		service:    service,
		onboarding: onboarding,
		shift:      shift,
		training:   training,
		payroll:    payroll,
		rating:     rating,
		safety:     safety,
		mobileSync: mobileSync,
		security:   security,
		logger:     logger,
		validate:   validator.New(),
	}
}

func (h *DriverHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.ListDrivers)
	r.Post("/", h.CreateDriver)
	r.Route("/onboarding", func(r chi.Router) {
		r.Post("/start", h.StartOnboarding)
	})
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.GetDriver)
		r.Put("/", h.UpdateDriver)
		r.Delete("/", h.DeleteDriver)
		r.Post("/assign-vehicle", h.AssignVehicle)
		r.Post("/unassign-vehicle", h.UnassignVehicle)
		r.Post("/update-status", h.UpdateStatus)
		r.Get("/behavior-score", h.GetBehaviorScore)
		r.Post("/update-behavior", h.UpdateBehaviorScore)
		r.Get("/analytics", h.GetAnalytics)
		r.Route("/onboarding", func(r chi.Router) {
			r.Post("/advance", h.AdvanceOnboarding)
			r.Get("/status", h.GetOnboardingStatus)
		})
		r.Route("/shifts", func(r chi.Router) {
			r.Post("/", h.CreateShift)
			r.Get("/", h.GetShifts)
		})
		r.Route("/payroll", func(r chi.Router) {
			r.Get("/earnings", h.GetEarnings)
			r.Get("/payouts", h.GetPayouts)
		})
		r.Route("/ratings", func(r chi.Router) {
			r.Post("/", h.SubmitRating)
			r.Get("/", h.GetRatings)
		})
		r.Route("/documents", func(r chi.Router) {
			r.Get("/", h.ListDocuments)
			r.Post("/", h.UploadDocument)
			r.Delete("/{doc_id}", h.DeleteDocument)
		})
		r.Route("/attendance", func(r chi.Router) {
			r.Get("/", h.GetAttendance)
			r.Post("/", h.MarkAttendance)
		})
		r.Route("/schedule", func(r chi.Router) {
			r.Get("/", h.GetSchedule)
			r.Post("/", h.SetSchedule)
		})

		// Mobile App Routes with Security Middleware
		r.Group(func(r chi.Router) {
			r.Use(h.security.DeviceFingerprintMiddleware)

			r.Post("/duty-status", h.UpdateDutyStatus)
			r.Post("/sos", h.TriggerSOS)
			r.Post("/sync", h.SyncMobileData)
			r.Route("/messages", func(r chi.Router) {
				r.Post("/", h.SendMessage)
				r.Get("/", h.GetMessages)
			})
			r.Route("/expenses", func(r chi.Router) {
				r.Post("/", h.SubmitExpense)
				r.Get("/", h.GetExpenses)
			})
			r.Get("/performance", h.GetPerformance)
		})
	})
}

// CreateDriver godoc
func (h *DriverHandler) CreateDriver(w http.ResponseWriter, r *http.Request) {
	var dto createDriverRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.logger.Error("failed to decode create driver request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate request
	if err := h.validate.Struct(dto); err != nil {
		h.logger.Error("validation failed for create driver request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Map DTO to model
	req := &models.CreateDriverRequest{
		FirstName:        dto.FirstName,
		LastName:         dto.LastName,
		Email:            dto.Email,
		Phone:            dto.Phone,
		LicenseNumber:    dto.LicenseNumber,
		LicenseType:      dto.LicenseType,
		LicenseExpiry:    dto.LicenseExpiry,
		LicenseAuthority: dto.LicenseAuthority,
		LicenseIssueDate: dto.LicenseIssueDate,
		DateOfBirth:      dto.DateOfBirth,
		EmployeeCode:     dto.EmployeeCode,
		AadhaarNumber:    dto.AadhaarNumber,
		PanNumber:        dto.PanNumber,
		EmploymentType:   dto.EmploymentType,
		Address:          dto.Address,
		EmergencyContact: dto.EmergencyContact,
		ExperienceYears:  dto.ExperienceYears,
		Metadata:         dto.Metadata,
	}

	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID not found", nil)
		return
	}

	driver, err := h.service.Create(r.Context(), orgID.String(), req)
	if err != nil {
		h.logger.Error("failed to create driver", zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err.Error() == "license has expired" || err.Error() == "driver with license number "+dto.LicenseNumber+" already exists" {
			statusCode = http.StatusBadRequest
		}
		h.writeErrorResponse(w, statusCode, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, map[string]interface{}{"data": driver})
}

// GetDriver godoc
func (h *DriverHandler) GetDriver(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")

	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid driver ID format", nil)
		return
	}

	driver, ok := h.verifyDriverOwnership(w, r, driverID)
	if !ok {
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": driver})
}

// ListDrivers godoc
func (h *DriverHandler) ListDrivers(w http.ResponseWriter, r *http.Request) {
	p := pagination.GetParams(r)

	query := r.URL.Query()
	var filter models.DriverListFilter
	filter.Status = query.Get("status")
	filter.VehicleID = query.Get("vehicle_id")
	filter.Name = query.Get("name")
	filter.Page = p.Page
	filter.PageSize = p.Limit
	filter.OrderBy = query.Get("order_by")
	filter.OrderDir = query.Get("order_dir")

	orgID := h.getOrgID(r)
	filter.OrgID = orgID.String()

	drivers, totalCount, err := h.service.List(r.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list drivers", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list drivers", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, pagination.NewResponse(drivers, totalCount, p.Page, p.Limit))
}

// UpdateDriver godoc
func (h *DriverHandler) UpdateDriver(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")

	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid driver ID format", nil)
		return
	}

	// Verify driver belongs to org before update (security check)
	_, ok := h.verifyDriverOwnership(w, r, driverID)
	if !ok {
		return
	}

	var dto updateDriverRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.logger.Error("failed to decode update driver request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate request
	if err := h.validate.Struct(dto); err != nil {
		h.logger.Error("validation failed for update driver request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Map DTO to model
	req := &models.UpdateDriverRequest{
		FirstName:        dto.FirstName,
		LastName:         dto.LastName,
		Email:            dto.Email,
		Phone:            dto.Phone,
		LicenseNumber:    dto.LicenseNumber,
		LicenseType:      dto.LicenseType,
		LicenseExpiry:    dto.LicenseExpiry,
		LicenseAuthority: dto.LicenseAuthority,
		LicenseIssueDate: dto.LicenseIssueDate,
		DateOfBirth:      dto.DateOfBirth,
		EmployeeCode:     dto.EmployeeCode,
		AadhaarNumber:    dto.AadhaarNumber,
		PanNumber:        dto.PanNumber,
		EmploymentType:   dto.EmploymentType,
		Address:          dto.Address,
		EmergencyContact: dto.EmergencyContact,
		ExperienceYears:  dto.ExperienceYears,
		Metadata:         dto.Metadata,
	}

	driver, err := h.service.Update(r.Context(), driverID, req)
	if err != nil {
		h.logger.Error("failed to update driver", zap.String("driver_id", driverID), zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err.Error() == "driver not found: no rows in result set" {
			statusCode = http.StatusNotFound
		}
		h.writeErrorResponse(w, statusCode, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": driver})
}

// DeleteDriver godoc
func (h *DriverHandler) DeleteDriver(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")

	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid driver ID format", nil)
		return
	}

	// Verify driver belongs to org before deletion (security check)
	_, ok := h.verifyDriverOwnership(w, r, driverID)
	if !ok {
		return
	}

	err := h.service.Delete(r.Context(), driverID)
	if err != nil {
		h.logger.Error("failed to delete driver", zap.String("driver_id", driverID), zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err.Error() == "driver not found: no rows in result set" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "cannot delete driver with active vehicle assignment" ||
			err.Error() == "cannot delete driver with active trip" {
			statusCode = http.StatusConflict
		}
		h.writeErrorResponse(w, statusCode, err.Error(), nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AssignVehicle godoc
func (h *DriverHandler) AssignVehicle(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")

	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid driver ID format", nil)
		return
	}

	// Verify driver belongs to org
	if _, ok := h.verifyDriverOwnership(w, r, driverID); !ok {
		return
	}

	var dto assignVehicleRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.logger.Error("failed to decode assign vehicle request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate request
	if err := h.validate.Struct(dto); err != nil {
		h.logger.Error("validation failed for assign vehicle request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Validate vehicle ID format
	if _, err := uuid.Parse(dto.VehicleID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid vehicle ID format", nil)
		return
	}

	err := h.service.AssignVehicle(r.Context(), driverID, dto.VehicleID)
	if err != nil {
		h.logger.Error("failed to assign vehicle to driver",
			zap.String("driver_id", driverID),
			zap.String("vehicle_id", dto.VehicleID),
			zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err.Error() == "driver not found: no rows in result set" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "driver is not active" ||
			err.Error() == "driver already has a vehicle assigned" {
			statusCode = http.StatusConflict
		}
		h.writeErrorResponse(w, statusCode, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":    "Vehicle assigned successfully",
		"driver_id":  driverID,
		"vehicle_id": dto.VehicleID,
	})
}

// UnassignVehicle godoc
func (h *DriverHandler) UnassignVehicle(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")

	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid driver ID format", nil)
		return
	}

	// Verify driver belongs to org
	if _, ok := h.verifyDriverOwnership(w, r, driverID); !ok {
		return
	}

	err := h.service.UnassignVehicle(r.Context(), driverID)
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
		h.writeErrorResponse(w, statusCode, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":   "Vehicle unassigned successfully",
		"driver_id": driverID,
	})
}

// UpdateStatus godoc
func (h *DriverHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")

	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid driver ID format", nil)
		return
	}

	// Verify driver belongs to org
	if _, ok := h.verifyDriverOwnership(w, r, driverID); !ok {
		return
	}

	var dto updateStatusRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.logger.Error("failed to decode update status request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate request
	if err := h.validate.Struct(dto); err != nil {
		h.logger.Error("validation failed for update status request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err := h.service.UpdateStatus(r.Context(), driverID, dto.Status)
	if err != nil {
		h.logger.Error("failed to update driver status",
			zap.String("driver_id", driverID),
			zap.String("status", dto.Status),
			zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err.Error() == "driver not found: no rows in result set" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "invalid status: "+dto.Status {
			statusCode = http.StatusBadRequest
		}
		h.writeErrorResponse(w, statusCode, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":   "Driver status updated successfully",
		"driver_id": driverID,
		"status":    dto.Status,
	})
}

// GetBehaviorScore godoc
func (h *DriverHandler) GetBehaviorScore(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")

	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid driver ID format", nil)
		return
	}

	// Verify driver belongs to org
	if _, ok := h.verifyDriverOwnership(w, r, driverID); !ok {
		return
	}

	score, err := h.service.GetBehaviorScore(r.Context(), driverID)
	if err != nil {
		h.logger.Error("failed to get behavior score",
			zap.String("driver_id", driverID),
			zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err.Error() == "driver not found: no rows in result set" {
			statusCode = http.StatusNotFound
		}
		h.writeErrorResponse(w, statusCode, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"driver_id":      driverID,
		"behavior_score": score,
		"max_score":      100,
	})
}

// UpdateBehaviorScore godoc
func (h *DriverHandler) UpdateBehaviorScore(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")

	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid driver ID format", nil)
		return
	}

	// Verify driver belongs to org
	if _, ok := h.verifyDriverOwnership(w, r, driverID); !ok {
		return
	}

	var dto updateBehaviorScoreRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.logger.Error("failed to decode update behavior score request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate request
	if err := h.validate.Struct(dto); err != nil {
		h.logger.Error("validation failed for update behavior score request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err := h.service.UpdateBehaviorScore(r.Context(), driverID, dto.ScoreChange, dto.Reason)
	if err != nil {
		h.logger.Error("failed to update behavior score",
			zap.String("driver_id", driverID),
			zap.Int("score_change", dto.ScoreChange),
			zap.String("reason", dto.Reason),
			zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err.Error() == "driver not found: no rows in result set" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "score change must be between -20 and 20" {
			statusCode = http.StatusBadRequest
		}
		h.writeErrorResponse(w, statusCode, err.Error(), nil)
		return
	}

	// Get updated score
	newScore, err := h.service.GetBehaviorScore(r.Context(), driverID)
	if err != nil {
		newScore = 0 // fallback
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":      "Behavior score updated successfully",
		"driver_id":    driverID,
		"score_change": dto.ScoreChange,
		"new_score":    newScore,
		"reason":       dto.Reason,
	})
}

// GetAnalytics godoc
func (h *DriverHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")

	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid driver ID format", nil)
		return
	}

	// Verify driver belongs to org
	if _, ok := h.verifyDriverOwnership(w, r, driverID); !ok {
		return
	}

	analytics, err := h.service.GetDriverAnalytics(r.Context(), driverID)
	if err != nil {
		h.logger.Error("failed to get driver analytics",
			zap.String("driver_id", driverID),
			zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err.Error() == "driver not found: no rows in result set" {
			statusCode = http.StatusNotFound
		}
		h.writeErrorResponse(w, statusCode, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, analytics)
}

// UploadDocument godoc
func (h *DriverHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")

	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid driver ID format", nil)
		return
	}

	// Verify driver belongs to org
	if _, ok := h.verifyDriverOwnership(w, r, driverID); !ok {
		return
	}

	var dto uploadDocumentRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.logger.Error("failed to decode upload document request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate request
	if err := h.validate.Struct(dto); err != nil {
		h.logger.Error("validation failed for upload document request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Map DTO to model
	req := &models.UploadDocumentRequest{
		DocumentType:     dto.DocumentType,
		DocumentNumber:   dto.DocumentNumber,
		IssuingAuthority: dto.IssuingAuthority,
		IssueDate:        dto.IssueDate,
		ExpiryDate:       dto.ExpiryDate,
		FileURL:          dto.FileURL,
	}

	// In a real implementation, this would call service.UploadDocument
	// Since we are refactoring existing gin code which had mock behavior, we maintain it or improve if service is ready
	docID := uuid.New().String()

	h.writeJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"message":         "Document uploaded successfully",
		"driver_id":       driverID,
		"document_id":     docID,
		"document_type":   req.DocumentType,
		"document_number": req.DocumentNumber,
	})
}

// ListDocuments godoc
func (h *DriverHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")

	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid driver ID format", nil)
		return
	}

	// Verify driver belongs to org
	if _, ok := h.verifyDriverOwnership(w, r, driverID); !ok {
		return
	}

	// Mock behavior from original file
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"driver_id": driverID,
		"documents": []interface{}{},
		"count":     0,
	})
}

// DeleteDocument godoc
func (h *DriverHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")
	docID := chi.URLParam(r, "doc_id")

	// Validate UUID formats
	if _, err := uuid.Parse(driverID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid driver ID format", nil)
		return
	}

	if _, err := uuid.Parse(docID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid document ID format", nil)
		return
	}

	// Verify driver belongs to org
	if _, ok := h.verifyDriverOwnership(w, r, driverID); !ok {
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":     "Document deleted successfully",
		"driver_id":   driverID,
		"document_id": docID,
	})
}

// MarkAttendance godoc
func (h *DriverHandler) MarkAttendance(w http.ResponseWriter, r *http.Request) {
	driverIDStr := chi.URLParam(r, "id")
	driverID, err := uuid.Parse(driverIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid driver ID format", nil)
		return
	}

	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "organization ID required", nil)
		return
	}

	// Verify driver belongs to org
	if _, ok := h.verifyDriverOwnership(w, r, driverIDStr); !ok {
		return
	}

	var dto markAttendanceRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.logger.Error("failed to decode mark attendance request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate request
	if err := h.validate.Struct(dto); err != nil {
		h.logger.Error("validation failed for mark attendance request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var checkIn, checkOut *time.Time
	if dto.CheckInTime != nil {
		checkIn = dto.CheckInTime
	}
	if dto.CheckOutTime != nil {
		checkOut = dto.CheckOutTime
	}

	if err := h.service.MarkAttendance(r.Context(), driverID, orgID, dto.Date, dto.Status, checkIn, checkOut, dto.Remarks); err != nil {
		h.logger.Error("failed to mark attendance", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"message":   "Attendance marked successfully",
		"driver_id": driverID,
		"date":      dto.Date.Format("2006-01-02"),
		"status":    dto.Status,
	})
}

// GetAttendance godoc
func (h *DriverHandler) GetAttendance(w http.ResponseWriter, r *http.Request) {
	driverIDStr := chi.URLParam(r, "id")
	driverID, err := uuid.Parse(driverIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid driver ID format", nil)
		return
	}

	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "organization ID required", nil)
		return
	}

	// Verify driver belongs to org
	if _, ok := h.verifyDriverOwnership(w, r, driverIDStr); !ok {
		return
	}

	query := r.URL.Query()
	startDateStr := query.Get("start_date")
	endDateStr := query.Get("end_date")

	var startDate, endDate time.Time
	var parseErr error

	if startDateStr != "" {
		startDate, parseErr = time.Parse("2006-01-02", startDateStr)
		if parseErr != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid start date format", parseErr)
			return
		}
	} else {
		startDate = time.Now().AddDate(0, 0, -30) // Last 30 days default
	}

	if endDateStr != "" {
		endDate, parseErr = time.Parse("2006-01-02", endDateStr)
		if parseErr != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid end date format", parseErr)
			return
		}
	} else {
		endDate = time.Now()
	}

	attendance, err := h.service.GetAttendance(r.Context(), driverID, orgID, startDate, endDate)
	if err != nil {
		h.logger.Error("failed to get attendance", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"driver_id":          driverID,
		"attendance_records": attendance,
		"start_date":         startDate.Format("2006-01-02"),
		"end_date":           endDate.Format("2006-01-02"),
		"count":              len(attendance),
	})
}

// StartOnboarding handles starting the driver onboarding process
func (h *DriverHandler) StartOnboarding(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DriverID string `json:"driver_id" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.onboarding.StartOnboarding(r.Context(), req.DriverID); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "onboarding_started"})
}

// AdvanceOnboarding handles advancing the driver to the next onboarding step
func (h *DriverHandler) AdvanceOnboarding(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DriverID string `json:"driver_id" validate:"required"`
		StepName string `json:"step_name" validate:"required"`
		Status   string `json:"status" validate:"required"`
		Remarks  string `json:"remarks"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.onboarding.AdvanceOnboarding(r.Context(), req.DriverID, req.StepName, req.Status, req.Remarks); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "onboarding_advanced"})
}

// GetOnboardingStatus retrieves the driver's onboarding status
func (h *DriverHandler) GetOnboardingStatus(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")
	steps, err := h.onboarding.GetOnboardingStatus(r.Context(), driverID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"driver_id": driverID,
		"steps":     steps,
	})
}

// CreateShift handles creating a new shift for a driver
func (h *DriverHandler) CreateShift(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DriverID  string    `json:"driver_id" validate:"required"`
		ShiftStart time.Time `json:"shift_start" validate:"required"`
		ShiftEnd   time.Time `json:"shift_end" validate:"required"`
		ShiftType  string    `json:"shift_type"`
		VehicleID *string   `json:"vehicle_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	shift := &models.Shift{
		ID:         uuid.New().String(),
		DriverID:   req.DriverID,
		ShiftStart: req.ShiftStart,
		ShiftEnd:   req.ShiftEnd,
		ShiftType:  req.ShiftType,
		Status:     "SCHEDULED",
	}

	if err := h.shift.CreateShift(r.Context(), shift); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	if req.VehicleID != nil {
		if err := h.shift.AssignToVehicle(r.Context(), shift.ID, *req.VehicleID); err != nil {
			h.logger.Warn("failed to assign vehicle to shift", zap.Error(err))
		}
	}

	h.writeJSONResponse(w, http.StatusCreated, shift)
}

// GetShifts retrieves shifts for a driver
func (h *DriverHandler) GetShifts(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	var start, end time.Time
	var err error

	if startDate != "" {
		start, err = time.Parse("2006-01-02", startDate)
	} else {
		start = time.Now().AddDate(0, -1, 0)
	}

	if endDate != "" {
		end, err = time.Parse("2006-01-02", endDate)
	} else {
		end = time.Now()
	}

	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid date format", err)
		return
	}

	shifts, err := h.shift.GetShiftsByDriverID(r.Context(), driverID, start, end)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"driver_id": driverID,
		"shifts":    shifts,
		"count":     len(shifts),
	})
}

// GetEarnings retrieves trip earnings for a driver
func (h *DriverHandler) GetEarnings(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	var start, end time.Time
	var err error

	if startDate != "" {
		start, err = time.Parse("2006-01-02", startDate)
	} else {
		start = time.Now().AddDate(0, -1, 0)
	}

	if endDate != "" {
		end, err = time.Parse("2006-01-02", endDate)
	} else {
		end = time.Now()
	}

	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid date format", err)
		return
	}

	earnings, err := h.payroll.GetTripEarnings(r.Context(), driverID, start, end)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"driver_id": driverID,
		"earnings":  earnings,
		"count":     len(earnings),
	})
}

// GetPayouts retrieves payouts for a driver
func (h *DriverHandler) GetPayouts(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")
	// For now, return empty payouts - can be implemented later
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"driver_id": driverID,
		"payouts":   []interface{}{},
	})
}

// SubmitRating handles submitting a rating for a driver
func (h *DriverHandler) SubmitRating(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DriverID        string  `json:"driver_id" validate:"required"`
		TripID          string  `json:"trip_id" validate:"required"`
		PunctualityScore int    `json:"punctuality_score"`
		BehaviorScore   int     `json:"behavior_score"`
		DrivingScore    int     `json:"driving_score"`
		OverallRating   float64 `json:"overall_rating"`
		Comments        string  `json:"comments"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	rating := &models.Rating{
		DriverID:         req.DriverID,
		TripID:           req.TripID,
		PunctualityScore: req.PunctualityScore,
		BehaviorScore:    req.BehaviorScore,
		DrivingScore:     req.DrivingScore,
		OverallRating:    req.OverallRating,
		Comments:         req.Comments,
	}

	if err := h.rating.SubmitRating(r.Context(), rating); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, rating)
}

// GetRatings retrieves ratings for a driver
func (h *DriverHandler) GetRatings(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")
	avgRating, err := h.rating.GetDriverRating(r.Context(), driverID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"driver_id":      driverID,
		"average_rating": avgRating,
	})
}

// GetSchedule retrieves the schedule for a driver
func (h *DriverHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")
	// For now, return empty schedule - can be implemented later
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"driver_id": driverID,
		"schedule":  []interface{}{},
	})
}

// SetSchedule handles setting a driver's schedule
func (h *DriverHandler) SetSchedule(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")
	var req struct {
		DayOfWeek int    `json:"day_of_week"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		IsWorking bool   `json:"is_working"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	// Stub implementation
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"driver_id": driverID,
		"schedule":  req,
	})
}

// SyncMobileData handles mobile data synchronization
func (h *DriverHandler) SyncMobileData(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	// Stub implementation
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"driver_id": driverID,
		"synced":    true,
	})
}

// Helper methods

func (h *DriverHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *DriverHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
	if err != nil {
		h.logger.Debug("error details", zap.Error(err))
	}
}

func (h *DriverHandler) getOrgID(r *http.Request) uuid.UUID {
	ctx := r.Context()
	
	// Check standardized OrgIDKey first
	if val := ctx.Value(ctxkeys.OrgIDKey); val != nil {
		switch v := val.(type) {
		case uuid.UUID:
			return v
		case string:
			if id, err := uuid.Parse(v); err == nil {
				return id
			}
		}
	}

	// Check legacy/string literal fallbacks
	for _, key := range []interface{}{"orgID", "org_id", ctxkeys.TenantIDKey} {
		if val := ctx.Value(key); val != nil {
			switch v := val.(type) {
			case uuid.UUID:
				return v
			case string:
				if id, err := uuid.Parse(v); err == nil {
					return id
				}
			}
		}
	}

	return uuid.Nil
}

func (h *DriverHandler) verifyDriverOwnership(w http.ResponseWriter, r *http.Request, driverID string) (*models.Driver, bool) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "organization ID required", nil)
		return nil, false
	}

	driver, err := h.service.GetByID(r.Context(), driverID)
	if err != nil {
		h.logger.Error("failed to get driver for ownership check", zap.String("driver_id", driverID), zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err.Error() == "driver not found: no rows in result set" {
			statusCode = http.StatusNotFound
		}
		h.writeErrorResponse(w, statusCode, "Driver not found", err)
		return nil, false
	}

	if driver.OrgID != orgID.String() {
		h.logger.Warn("unauthorized access attempt to driver",
			zap.String("driver_id", driverID),
			zap.String("requester_org_id", orgID.String()),
			zap.String("driver_org_id", driver.OrgID))
		h.writeErrorResponse(w, http.StatusForbidden, "Access denied", nil)
		return nil, false
	}

	return driver, true
}
