package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/fleet-platform/shared/pkg/pagination"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"

	driverMiddleware "github.com/flyfleet/fleet-core-service/internal/domains/driver/middleware"
	"github.com/flyfleet/fleet-core-service/internal/domains/driver/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/driver/service"
)

type DriverHandler struct {
	service  *service.DriverService
	security *driverMiddleware.DriverSecurityMiddleware
	logger   *zap.Logger
	validate *validator.Validate
}

func NewDriverHandler(service *service.DriverService, security *driverMiddleware.DriverSecurityMiddleware, logger *zap.Logger) *DriverHandler {
	return &DriverHandler{
		service:  service,
		security: security,
		logger:   logger,
		validate: validator.New(),
	}
}

func (h *DriverHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.ListDrivers)
	r.Post("/", h.CreateDriver)
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
			r.Route("/messages", func(r chi.Router) {
				r.Post("/", h.SendMessage)
				r.Get("/", h.GetMessages)
			})
			r.Route("/expenses", func(r chi.Router) {
				r.Post("/", h.SubmitExpense)
				r.Get("/", h.GetExpenses)
			})
		})
	})
}

// CreateDriver godoc
func (h *DriverHandler) CreateDriver(w http.ResponseWriter, r *http.Request) {
	var req models.CreateDriverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode create driver request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("validation failed for create driver request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID not found", nil)
		return
	}

	driver, err := h.service.Create(r.Context(), orgID.String(), &req)
	if err != nil {
		h.logger.Error("failed to create driver", zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err.Error() == "license has expired" || err.Error() == "driver with license number "+req.LicenseNumber+" already exists" {
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

	driver, err := h.service.GetByID(r.Context(), driverID)
	if err != nil {
		h.logger.Error("failed to get driver", zap.String("driver_id", driverID), zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err.Error() == "driver not found: no rows in result set" {
			statusCode = http.StatusNotFound
		}
		h.writeErrorResponse(w, statusCode, err.Error(), nil)
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

	var req models.UpdateDriverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode update driver request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("validation failed for update driver request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	driver, err := h.service.Update(r.Context(), driverID, &req)
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

	var req models.AssignVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode assign vehicle request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("validation failed for assign vehicle request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Validate vehicle ID format
	if _, err := uuid.Parse(req.VehicleID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid vehicle ID format", nil)
		return
	}

	err := h.service.AssignVehicle(r.Context(), driverID, req.VehicleID)
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
		h.writeErrorResponse(w, statusCode, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Vehicle assigned successfully",
		"driver_id": driverID,
		"vehicle_id": req.VehicleID,
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
		"message": "Vehicle unassigned successfully",
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

	var req models.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode update status request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("validation failed for update status request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err := h.service.UpdateStatus(r.Context(), driverID, req.Status)
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
		h.writeErrorResponse(w, statusCode, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Driver status updated successfully",
		"driver_id": driverID,
		"status": req.Status,
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
		"driver_id": driverID,
		"behavior_score": score,
		"max_score": 100,
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

	var req models.UpdateBehaviorScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode update behavior score request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("validation failed for update behavior score request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err := h.service.UpdateBehaviorScore(r.Context(), driverID, req.ScoreChange, req.Reason)
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
		h.writeErrorResponse(w, statusCode, err.Error(), nil)
		return
	}

	// Get updated score
	newScore, err := h.service.GetBehaviorScore(r.Context(), driverID)
	if err != nil {
		newScore = 0 // fallback
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Behavior score updated successfully",
		"driver_id": driverID,
		"score_change": req.ScoreChange,
		"new_score": newScore,
		"reason": req.Reason,
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

	var req models.UploadDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode upload document request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("validation failed for upload document request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// In a real implementation, this would call service.UploadDocument
	// Since we are refactoring existing gin code which had mock behavior, we maintain it or improve if service is ready
	docID := uuid.New().String()
	
	h.writeJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"message": "Document uploaded successfully",
		"driver_id": driverID,
		"document_id": docID,
		"document_type": req.DocumentType,
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

	// Mock behavior from original file
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"driver_id": driverID,
		"documents": []interface{}{},
		"count": 0,
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

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Document deleted successfully",
		"driver_id": driverID,
		"document_id": docID,
	})
}

// MarkAttendance godoc
func (h *DriverHandler) MarkAttendance(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")
	
	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid driver ID format", nil)
		return
	}

	var req models.MarkAttendanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode mark attendance request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("validation failed for mark attendance request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	recordID := uuid.New().String()
	
	h.writeJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"message": "Attendance marked successfully",
		"driver_id": driverID,
		"record_id": recordID,
		"date": req.Date.Format("2006-01-02"),
		"status": req.Status,
	})
}

// GetAttendance godoc
func (h *DriverHandler) GetAttendance(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")
	
	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid driver ID format", nil)
		return
	}

	query := r.URL.Query()
	startDateStr := query.Get("start_date")
	endDateStr := query.Get("end_date")
	
	var startDate, endDate time.Time
	var err error
	
	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid start date format", err)
			return
		}
	} else {
		startDate = time.Now().AddDate(0, 0, -30) // Last 30 days default
	}
	
	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid end date format", err)
			return
		}
	} else {
		endDate = time.Now()
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"driver_id": driverID,
		"attendance_records": []interface{}{},
		"start_date": startDate.Format("2006-01-02"),
		"end_date": endDate.Format("2006-01-02"),
		"count": 0,
	})
}

// SetSchedule godoc
func (h *DriverHandler) SetSchedule(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")
	
	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid driver ID format", nil)
		return
	}

	var req models.SetScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode set schedule request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("validation failed for set schedule request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	scheduleID := uuid.New().String()
	
	h.writeJSONResponse(w, http.StatusCreated, map[string]interface{}{
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
func (h *DriverHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "id")
	
	// Validate UUID format
	if _, err := uuid.Parse(driverID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid driver ID format", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"driver_id": driverID,
		"schedules": []interface{}{},
		"count": 0,
	})
}

// Helper methods

func (h *DriverHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", zap.Error(err))
	}
}

func (h *DriverHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	resp := map[string]interface{}{
		"error": message,
	}
	if err != nil && statusCode == http.StatusInternalServerError {
		resp["details"] = err.Error()
	}
	h.writeJSONResponse(w, statusCode, resp)
}

func (h *DriverHandler) getOrgID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.OrgIDKey); val != nil {
		if orgIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(orgIDStr); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}
