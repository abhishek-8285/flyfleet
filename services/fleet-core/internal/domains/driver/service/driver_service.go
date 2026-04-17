package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/domains/driver/models"
)

// Repository defines the interface for driver persistence
type Repository interface {
	// Driver Methods
	Create(ctx context.Context, driver *models.Driver) error
	GetByID(ctx context.Context, id string) (*models.Driver, error)
	GetByEmail(ctx context.Context, email string) (*models.Driver, error)
	GetByLicenseNumber(ctx context.Context, licenseNumber string) (*models.Driver, error)
	List(ctx context.Context, filter models.DriverListFilter) ([]*models.Driver, int, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) error
	Delete(ctx context.Context, id string) error
	AssignVehicle(ctx context.Context, driverID, vehicleID string) error
	UnassignVehicle(ctx context.Context, driverID string) error
	UpdateBehaviorScore(ctx context.Context, driverID string, scoreChange int) error
	IncrementStats(ctx context.Context, driverID string, distanceKm, hoursDriven float64) error
	GetDriverAnalytics(ctx context.Context, driverID string) (*models.DriverAnalytics, error)
	GetActiveDriversForVehicle(ctx context.Context, vehicleID string) ([]*models.Driver, error)

	// Mobile App Methods
	CreateDutyStatusLog(ctx context.Context, log *models.DutyStatusLog) error
	GetLatestDutyStatus(ctx context.Context, driverID string) (*models.DutyStatusLog, error)
	CreateMessage(ctx context.Context, msg *models.DriverMessage) error
	GetMessages(ctx context.Context, driverID string, limit, offset int) ([]*models.DriverMessage, error)
	CreateExpense(ctx context.Context, expense *models.DriverExpense) error
	ListExpenses(ctx context.Context, driverID string) ([]*models.DriverExpense, error)

	// Behavior Scoring & Incident Methods
	LogBehaviorScoreChange(ctx context.Context, driverID, orgID uuid.UUID, oldScore, newScore int, reason, eventType string, userID uuid.UUID, userEmail string) error
	CreateBehaviorScoreUpdate(ctx context.Context, update *models.BehaviorScoreUpdate) error
	GetBehaviorRecords(ctx context.Context, driverID string, startDate, endDate time.Time) ([]*models.DriverBehavior, error)
	GetBehaviorSummary(ctx context.Context, driverID string) (*models.DriverBehavior, error)
	GetIncidents(ctx context.Context, driverID string, filter models.BehaviorReportRequest) ([]*models.DriverIncident, error)

	// Duty Status Methods
	EndDutyStatusLog(ctx context.Context, logID string, durationMinutes int) error
	GetDutyStatusLogs(ctx context.Context, driverID string, startDate, endDate time.Time) ([]*models.DutyStatusLog, error)

	// Attendance & Schedule Methods
	CreateAttendance(ctx context.Context, attendance *models.DriverAttendance) error
	GetAttendanceByID(ctx context.Context, id uuid.UUID) (*models.DriverAttendance, error)
	GetAttendanceByDriver(ctx context.Context, driverID, orgID uuid.UUID, from, to time.Time) ([]*models.DriverAttendance, error)
	UpdateCheckOut(ctx context.Context, id, orgID uuid.UUID, checkOut time.Time, remarks string) error

	CreateSchedule(ctx context.Context, schedule *models.DriverSchedule) error
	GetScheduleByID(ctx context.Context, id uuid.UUID) (*models.DriverSchedule, error)
	GetSchedulesByDriver(ctx context.Context, driverID, orgID uuid.UUID) ([]*models.DriverSchedule, error)
	UpdateSchedule(ctx context.Context, schedule *models.DriverSchedule) error
	DeleteSchedule(ctx context.Context, id, orgID uuid.UUID) error
}

type DriverService struct {
	repo       Repository
	logger     *zap.Logger
	nc         *nats.Conn
	dutyStatus *DutyStatusService
}

func NewDriverService(repo Repository, logger *zap.Logger, opts ...func(*DriverService)) *DriverService {
	s := &DriverService{
		repo:       repo,
		logger:     logger,
		dutyStatus: NewDutyStatusService(repo, logger),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// WithNATS sets the NATS connection for event publishing
func WithNATS(nc *nats.Conn) func(*DriverService) {
	return func(s *DriverService) {
		s.nc = nc
	}
}

// publishEvent publishes a NATS event (best-effort, non-blocking)
func (s *DriverService) publishEvent(subject string, data interface{}) {
	if s.nc == nil {
		s.logger.Warn("NATS connection is NIL in DriverService, skipping event publish", zap.String("subject", subject))
		return
	}
	payload, err := json.Marshal(data)
	if err != nil {
		s.logger.Warn("failed to marshal NATS event", zap.String("subject", subject), zap.Error(err))
		return
	}
	if err := s.nc.Publish(subject, payload); err != nil {
		s.logger.Warn("failed to publish NATS event", zap.String("subject", subject), zap.Error(err))
	} else {
		s.logger.Info("NATS event published successfully", zap.String("subject", subject))
	}
}

// Create creates a new driver
func (s *DriverService) Create(ctx context.Context, orgID string, req *models.CreateDriverRequest) (*models.Driver, error) {
	// Validate license expiry date
	if req.LicenseExpiry.Before(time.Now()) {
		return nil, fmt.Errorf("license has expired")
	}

	// Check for unique email
	existing, err := s.repo.GetByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("driver with email %s already exists", req.Email)
	}

	// Check for unique license number
	existingLic, err := s.repo.GetByLicenseNumber(ctx, req.LicenseNumber)
	if err == nil && existingLic != nil {
		return nil, fmt.Errorf("driver with license number %s already exists", req.LicenseNumber)
	}

	driver := &models.Driver{
		ID:               uuid.New().String(),
		OrgID:            orgID,
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		Email:            req.Email,
		Phone:            req.Phone,
		LicenseNumber:    req.LicenseNumber,
		LicenseType:      &req.LicenseType,
		LicenseIssueDate: req.LicenseIssueDate,
		LicenseExpiry:    &req.LicenseExpiry,
		LicenseAuthority: &req.LicenseAuthority,
		EmployeeCode:     &req.EmployeeCode,
		AadhaarNumber:    &req.AadhaarNumber,
		PanNumber:        &req.PanNumber,
		DateOfBirth:      &req.DateOfBirth,
		Address:          req.Address,
		EmergencyContact: req.EmergencyContact,
		Status:           "active",
		EmploymentType:   &req.EmploymentType,
		JoiningDate:      time.Now(),
		ExperienceYears:  req.ExperienceYears,
		Rating:           0.0,
		TotalTrips:       0,
		TotalDistanceKm:  0.0,
		TotalHoursDriven: 0.0,
		BehaviorScore:    100, // Start with perfect score
		Metadata:         req.Metadata,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.repo.Create(ctx, driver); err != nil {
		s.logger.Error("failed to create driver", zap.Error(err))
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}

	s.logger.Info("driver created", zap.String("driver_id", driver.ID))
	s.publishEvent("driver.created", driver)
	return driver, nil
}

// GetByID retrieves a driver by ID
func (s *DriverService) GetByID(ctx context.Context, id string) (*models.Driver, error) {
	driver, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get driver", zap.String("driver_id", id), zap.Error(err))
		return nil, fmt.Errorf("driver not found: %w", err)
	}
	return driver, nil
}

// List retrieves drivers with filtering and pagination
func (s *DriverService) List(ctx context.Context, filter models.DriverListFilter) ([]*models.Driver, int, error) {
	drivers, totalCount, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list drivers", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list drivers: %w", err)
	}
	return drivers, totalCount, nil
}

// Update modifies an existing driver
func (s *DriverService) Update(ctx context.Context, id string, req *models.UpdateDriverRequest) (*models.Driver, error) {
	// Validate that driver exists
	driver, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.Email != nil && *req.Email != driver.Email {
		// Check for unique email when changing
		existing, err := s.repo.GetByEmail(ctx, *req.Email)
		if err == nil && existing != nil {
			return nil, fmt.Errorf("driver with email %s already exists", *req.Email)
		}
		updates["email"] = *req.Email
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.LicenseNumber != nil {
		updates["license_number"] = *req.LicenseNumber
	}
	if req.LicenseType != nil {
		updates["license_type"] = *req.LicenseType
	}
	if req.LicenseExpiry != nil {
		if req.LicenseExpiry.Before(time.Now()) {
			return nil, fmt.Errorf("license has expired")
		}
		updates["license_expiry"] = *req.LicenseExpiry
	}
	if req.LicenseAuthority != nil {
		updates["license_authority"] = *req.LicenseAuthority
	}
	if req.DateOfBirth != nil {
		updates["date_of_birth"] = *req.DateOfBirth
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.EmergencyContact != nil {
		updates["emergency_contact"] = *req.EmergencyContact
	}
	if req.ExperienceYears != nil {
		updates["experience_years"] = *req.ExperienceYears
	}
	if req.EmployeeCode != nil {
		updates["employee_code"] = *req.EmployeeCode
	}
	if req.AadhaarNumber != nil {
		updates["aadhaar_number"] = *req.AadhaarNumber
	}
	if req.PanNumber != nil {
		updates["pan_number"] = *req.PanNumber
	}
	if req.LicenseIssueDate != nil {
		updates["license_issue_date"] = *req.LicenseIssueDate
	}
	if req.EmploymentType != nil {
		updates["employment_type"] = *req.EmploymentType
	}
	if req.Metadata != nil {
		updates["metadata"] = *req.Metadata
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()

		if err := s.repo.Update(ctx, id, updates); err != nil {
			s.logger.Error("failed to update driver", zap.String("driver_id", id), zap.Error(err))
			return nil, fmt.Errorf("failed to update driver: %w", err)
		}
	}

	updated, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.publishEvent("drivers.updated", map[string]interface{}{
		"driver_id":  id,
		"updates":    updates,
		"old_values": driver,
	})
	return updated, nil
}

// Delete performs soft delete of a driver
func (s *DriverService) Delete(ctx context.Context, id string) error {
	// Check if driver has active assignments
	driver, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if driver.CurrentVehicleID != nil {
		return fmt.Errorf("cannot delete driver with active vehicle assignment")
	}

	if driver.CurrentTripID != nil {
		return fmt.Errorf("cannot delete driver with active trip")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete driver", zap.String("driver_id", id), zap.Error(err))
		return fmt.Errorf("failed to delete driver: %w", err)
	}

	s.logger.Info("driver deleted", zap.String("driver_id", id))
	s.publishEvent("drivers.deleted", map[string]string{"driver_id": id})
	return nil
}

// AssignVehicle assigns a vehicle to a driver
func (s *DriverService) AssignVehicle(ctx context.Context, driverID, vehicleID string) error {
	// Check if driver exists and is active
	driver, err := s.GetByID(ctx, driverID)
	if err != nil {
		return err
	}

	if driver.Status != "active" {
		return fmt.Errorf("driver is not active")
	}

	// Check if driver already has a vehicle assigned
	if driver.CurrentVehicleID != nil {
		return fmt.Errorf("driver already has a vehicle assigned")
	}

	if err := s.repo.AssignVehicle(ctx, driverID, vehicleID); err != nil {
		s.logger.Error("failed to assign vehicle to driver",
			zap.String("driver_id", driverID),
			zap.String("vehicle_id", vehicleID),
			zap.Error(err))
		return fmt.Errorf("failed to assign vehicle: %w", err)
	}

	s.logger.Info("vehicle assigned to driver",
		zap.String("driver_id", driverID),
		zap.String("vehicle_id", vehicleID))
	return nil
}

// UnassignVehicle removes vehicle assignment from a driver
func (s *DriverService) UnassignVehicle(ctx context.Context, driverID string) error {
	// Check if driver exists
	driver, err := s.GetByID(ctx, driverID)
	if err != nil {
		return err
	}

	// Check if driver has a vehicle assigned
	if driver.CurrentVehicleID == nil {
		return fmt.Errorf("driver has no vehicle assigned")
	}

	if err := s.repo.UnassignVehicle(ctx, driverID); err != nil {
		s.logger.Error("failed to unassign vehicle from driver",
			zap.String("driver_id", driverID),
			zap.Error(err))
		return fmt.Errorf("failed to unassign vehicle: %w", err)
	}

	s.logger.Info("vehicle unassigned from driver", zap.String("driver_id", driverID))
	return nil
}

// UpdateStatus updates driver status
func (s *DriverService) UpdateStatus(ctx context.Context, driverID string, status string) error {
	validStatuses := map[string]bool{
		"active": true, "inactive": true, "suspended": true, "resigned": true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	// If resigning, set resignation date
	if status == "resigned" {
		updates["resignation_date"] = time.Now()
		// Also unassign vehicle and trip if any
		driver, err := s.GetByID(ctx, driverID)
		if err == nil {
			if driver.CurrentVehicleID != nil {
				s.repo.UnassignVehicle(ctx, driverID)
			}
			if driver.CurrentTripID != nil {
				// Would need to integrate with trip service to properly end trip
				updates["current_trip_id"] = nil
			}
		}
	}

	if err := s.repo.Update(ctx, driverID, updates); err != nil {
		s.logger.Error("failed to update driver status",
			zap.String("driver_id", driverID),
			zap.String("status", status),
			zap.Error(err))
		return fmt.Errorf("failed to update driver status: %w", err)
	}

	s.logger.Info("driver status updated",
		zap.String("driver_id", driverID),
		zap.String("status", status))
	return nil
}

// UpdateBehaviorScore updates driver's behavior score
func (s *DriverService) UpdateBehaviorScore(ctx context.Context, driverID string, scoreChange int, reason string) error {
	// Validate score change range
	if scoreChange < -20 || scoreChange > 20 {
		return fmt.Errorf("score change must be between -20 and 20")
	}

	// Check if driver exists
	driver, err := s.GetByID(ctx, driverID)
	if err != nil {
		return err
	}

	newScore := driver.BehaviorScore + scoreChange

	// Ensure score stays within bounds
	if newScore < 0 {
		newScore = 0
	} else if newScore > 100 {
		newScore = 100
	}

	scoreChangeActual := newScore - driver.BehaviorScore

	if err := s.repo.UpdateBehaviorScore(ctx, driverID, scoreChangeActual); err != nil {
		s.logger.Error("failed to update behavior score",
			zap.String("driver_id", driverID),
			zap.Int("score_change", scoreChange),
			zap.String("reason", reason),
			zap.Error(err))
		return fmt.Errorf("failed to update behavior score: %w", err)
	}

	s.logger.Info("driver behavior score updated",
		zap.String("driver_id", driverID),
		zap.Int("old_score", driver.BehaviorScore),
		zap.Int("new_score", newScore),
		zap.Int("score_change", scoreChangeActual),
		zap.String("reason", reason))

	s.publishEvent("driver.behavior_updated", map[string]interface{}{
		"driver_id":    driverID,
		"old_score":    driver.BehaviorScore,
		"new_score":    newScore,
		"score_change": scoreChangeActual,
		"reason":       reason,
		"timestamp":    time.Now(),
	})

	return nil
}

// GetBehaviorScore retrieves driver's current behavior score
func (s *DriverService) GetBehaviorScore(ctx context.Context, driverID string) (int, error) {
	driver, err := s.GetByID(ctx, driverID)
	if err != nil {
		return 0, err
	}
	return driver.BehaviorScore, nil
}

// GetDriverAnalytics retrieves driver analytics
func (s *DriverService) GetDriverAnalytics(ctx context.Context, driverID string) (*models.DriverAnalytics, error) {
	analytics, err := s.repo.GetDriverAnalytics(ctx, driverID)
	if err != nil {
		s.logger.Error("failed to get driver analytics",
			zap.String("driver_id", driverID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get driver analytics: %w", err)
	}
	return analytics, nil
}

// IncrementDriverStats increments driver statistics after trip completion
func (s *DriverService) IncrementDriverStats(ctx context.Context, driverID string, distanceKm, hoursDriven float64) error {
	if err := s.repo.IncrementStats(ctx, driverID, distanceKm, hoursDriven); err != nil {
		s.logger.Error("failed to increment driver stats",
			zap.String("driver_id", driverID),
			zap.Float64("distance_km", distanceKm),
			zap.Float64("hours_driven", hoursDriven),
			zap.Error(err))
		return fmt.Errorf("failed to increment driver stats: %w", err)
	}
	return nil
}

// GetByLicenseNumber retrieves driver by license number
func (s *DriverService) GetByLicenseNumber(ctx context.Context, licenseNumber string) (*models.Driver, error) {
	driver, err := s.repo.GetByLicenseNumber(ctx, licenseNumber)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("driver with license number %s not found", licenseNumber)
		}
		return nil, fmt.Errorf("failed to get driver by license: %w", err)
	}
	return driver, nil
}

// GetActiveDriversForVehicle retrieves active drivers for a vehicle
func (s *DriverService) GetActiveDriversForVehicle(ctx context.Context, vehicleID string) ([]*models.Driver, error) {
	return s.repo.GetActiveDriversForVehicle(ctx, vehicleID)
}

// Publish satisfies the EventPublisher interface for workers
func (s *DriverService) Publish(ctx context.Context, topic string, payload interface{}) error {
	s.publishEvent(topic, payload)
	return nil
}

// GetActiveDriversDutyHours satisfies the DutyStatusProvider interface for workers
func (s *DriverService) GetActiveDriversDutyHours(ctx context.Context) ([]models.DriverDutyHours, error) {
	// This is a simplified implementation. In a real system, we would query the database
	// for active duty status logs and calculate hours.
	return nil, nil // Placeholder
}

// GetDriverPerformanceSummary retrieves a comprehensive performance summary for a driver
func (s *DriverService) GetDriverPerformanceSummary(ctx context.Context, driverID, orgID string) (map[string]interface{}, error) {
	driver, err := s.repo.GetByID(ctx, driverID)
	if err != nil {
		return nil, err
	}
	if driver == nil {
		return nil, fmt.Errorf("driver not found")
	}

	// Enforce organization isolation
	if driver.OrgID != orgID {
		s.logger.Warn("unauthorized access attempt to driver performance summary",
			zap.String("requested_driver", driverID),
			zap.String("authorized_org", driver.OrgID),
			zap.String("requested_org", orgID))
		return nil, fmt.Errorf("unauthorized access to driver data")
	}

	// Get behavior summary
	behavior, err := s.repo.GetBehaviorSummary(ctx, driverID)
	if err != nil {
		s.logger.Warn("failed to get behavior summary", zap.Error(err))
	}

	// Get latest rating
	rating, err := s.repo.GetDriverAnalytics(ctx, driverID)
	if err != nil {
		s.logger.Warn("failed to get driver analytics", zap.Error(err))
	}

	summary := map[string]interface{}{
		"driver_id":      driverID,
		"behavior_score": driver.BehaviorScore,
		"total_distance": driver.TotalDistanceKm,
		"total_hours":    driver.TotalHoursDriven,
		"status":         driver.Status,
	}

	if behavior != nil {
		summary["safety_score"] = behavior.SafetyScore
		summary["efficiency_score"] = behavior.EfficiencyScore
		summary["risk_level"] = behavior.RiskLevel
		summary["incidents"] = map[string]int{
			"speeding":           behavior.SpeedingIncidents,
			"harsh_braking":      behavior.HarshBrakings,
			"harsh_acceleration": behavior.HarshAccelerations,
		}
	}

	if rating != nil {
		summary["average_rating"] = rating.AverageRating
		summary["total_trips"] = rating.TotalTrips
	}

	return summary, nil
}

// ============================================================================
// Attendance & Schedule Methods
// ============================================================================

// MarkAttendance marks attendance for a driver
func (s *DriverService) MarkAttendance(ctx context.Context, driverID, orgID uuid.UUID, date time.Time, status string, checkIn, checkOut *time.Time, remarks string) error {
	// Verify driver belongs to org
	driver, err := s.repo.GetByID(ctx, driverID.String())
	if err != nil {
		return fmt.Errorf("failed to verify driver: %w", err)
	}
	if driver == nil || driver.OrgID != orgID.String() {
		return fmt.Errorf("driver not found or unauthorized")
	}

	att := &models.DriverAttendance{
		ID:             uuid.New(),
		DriverID:       driverID,
		OrgID:          orgID.String(),
		AttendanceDate: date,
		Status:         status,
		CheckInTime:    checkIn,
		CheckOutTime:   checkOut,
		Remarks:        remarks,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.CreateAttendance(ctx, att); err != nil {
		return fmt.Errorf("failed to mark attendance: %w", err)
	}

	// Publish event for attendance marking
	s.publishEvent("driver.attendance.marked", map[string]interface{}{
		"driver_id": driverID,
		"org_id":    orgID,
		"date":      date.Format("2006-01-02"),
		"status":    status,
	})

	return nil
}

// GetAttendance retrieves attendance records for a driver
func (s *DriverService) GetAttendance(ctx context.Context, driverID, orgID uuid.UUID, from, to time.Time) ([]*models.DriverAttendance, error) {
	// Verify driver belongs to org
	driver, err := s.repo.GetByID(ctx, driverID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to verify driver: %w", err)
	}
	if driver == nil || driver.OrgID != orgID.String() {
		return nil, fmt.Errorf("driver not found or unauthorized")
	}

	attendance, err := s.repo.GetAttendanceByDriver(ctx, driverID, orgID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendance: %w", err)
	}

	return attendance, nil
}

// SetSchedule sets or updates a driver's weekly schedule
func (s *DriverService) SetSchedule(ctx context.Context, driverID, orgID uuid.UUID, dayOfWeek int, startTime, endTime string, isWorking bool) error {
	// Verify driver belongs to org
	driver, err := s.repo.GetByID(ctx, driverID.String())
	if err != nil {
		return fmt.Errorf("failed to verify driver: %w", err)
	}
	if driver == nil || driver.OrgID != orgID.String() {
		return fmt.Errorf("driver not found or unauthorized")
	}

	// Check if schedule already exists for this day
	schedules, err := s.repo.GetSchedulesByDriver(ctx, driverID, orgID)
	if err != nil {
		return fmt.Errorf("failed to check existing schedule: %w", err)
	}

	// Find existing schedule for this day
	for _, sch := range schedules {
		if sch.DayOfWeek == dayOfWeek {
			// Update existing
			sch.DayOfWeek = dayOfWeek
			sch.StartTime = startTime
			sch.EndTime = endTime
			sch.IsWorking = isWorking
			sch.UpdatedAt = time.Now()

			if err := s.repo.UpdateSchedule(ctx, sch); err != nil {
				return fmt.Errorf("failed to update schedule: %w", err)
			}
			return nil
		}
	}

	// Create new schedule
	sch := &models.DriverSchedule{
		ID:        uuid.New(),
		DriverID:  driverID,
		OrgID:     orgID.String(),
		DayOfWeek: dayOfWeek,
		StartTime: startTime,
		EndTime:   endTime,
		IsWorking: isWorking,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateSchedule(ctx, sch); err != nil {
		return fmt.Errorf("failed to create schedule: %w", err)
	}

	// Publish event for schedule update
	s.publishEvent("driver.schedule.updated", map[string]interface{}{
		"driver_id":  driverID,
		"org_id":     orgID,
		"day":        dayOfWeek,
		"start_time": startTime,
		"end_time":   endTime,
		"is_working": isWorking,
	})

	return nil
}

// GetSchedule retrieves all schedules for a driver
func (s *DriverService) GetSchedule(ctx context.Context, driverID, orgID uuid.UUID) ([]*models.DriverSchedule, error) {
	// Verify driver belongs to org
	driver, err := s.repo.GetByID(ctx, driverID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to verify driver: %w", err)
	}
	if driver == nil || driver.OrgID != orgID.String() {
		return nil, fmt.Errorf("driver not found or unauthorized")
	}

	schedules, err := s.repo.GetSchedulesByDriver(ctx, driverID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedules: %w", err)
	}

	return schedules, nil
}

// DeleteSchedule deletes a driver's schedule entry
func (s *DriverService) DeleteSchedule(ctx context.Context, driverID, orgID, scheduleID uuid.UUID) error {
	// Verify driver belongs to org
	driver, err := s.repo.GetByID(ctx, driverID.String())
	if err != nil {
		return fmt.Errorf("failed to verify driver: %w", err)
	}
	if driver == nil || driver.OrgID != orgID.String() {
		return fmt.Errorf("driver not found or unauthorized")
	}

	if err := s.repo.DeleteSchedule(ctx, scheduleID, orgID); err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}

	return nil
}
