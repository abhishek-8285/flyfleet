package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/driver/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DutyStatusService handles FMCSA Hours of Service (HOS) compliance
type DutyStatusService struct {
	repo   Repository
	logger *zap.Logger
}

// NewDutyStatusService creates a new duty status service
func NewDutyStatusService(repo Repository, logger *zap.Logger) *DutyStatusService {
	return &DutyStatusService{
		repo:   repo,
		logger: logger,
	}
}

// Valid duty statuses
const (
	StatusOffDuty      = "OFF_DUTY"
	StatusSleeperBerth = "SLEEPER_BERTH"
	StatusDriving      = "DRIVING"
	StatusOnDuty       = "ON_DUTY"
	StatusRest         = "REST"
	StatusBreak        = "BREAK"
)

// India HOOS Constants (MTW Act)
const (
	IndiaMaxContinuousDriveTime = 8.0  // 8 hours continuous
	IndiaMaxDutyHours           = 12.0 // 12 hours duty window
	IndiaMandatoryBreakTime     = 5.0  // Break after 5 hours
	IndiaRequiredBreakDuration  = 30   // 30 minutes
)

// FMCSA HOS limits
const (
	MaxDrivingHours     = 11 // Maximum driving time in a duty day
	MaxDutyHours        = 14 // Maximum on-duty time in a duty day
	RequiredRestBreak   = 30 // Required break after 8 hours driving
	MinSleeperBerthTime = 7  // Minimum sleeper berth time (with 3hr off-duty)
	MinOffDutyReset     = 10 // Minimum off-duty to reset 14-hour clock
	MaxOnDutyDays       = 60 // 60 hours in 7 days
	MaxOnDutyDays8Day   = 70 // 70 hours in 8 days
)

// UpdateDutyStatus updates driver's duty status with validation
func (s *DutyStatusService) UpdateDutyStatus(ctx context.Context, driverID, tenantID uuid.UUID, newStatus string, location *models.Location, remarks string) (*models.DutyStatusLog, error) {
	// Validate status
	validStatuses := map[string]bool{
		StatusOffDuty:      true,
		StatusSleeperBerth: true,
		StatusDriving:      true,
		StatusOnDuty:       true,
		StatusRest:         true,
		StatusBreak:        true,
	}

	if !validStatuses[newStatus] {
		return nil, ErrInvalidDutyStatus
	}

	// Get driver's latest duty status
	latestLog, err := s.repo.GetLatestDutyStatus(ctx, driverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest duty status: %w", err)
	}

	// Validate status transition
	if latestLog != nil {
		if err := s.validateStatusTransition(latestLog.Status, newStatus); err != nil {
			return nil, err
		}

		// Check HOS compliance before allowing driving
		if newStatus == StatusDriving {
			// Determine which rules to use. For now, we default to India logic if requested or org-based.
			// Ideally, this is determined by org config.
			isIndiaRule := true // Defaulting to India as per project focus
			if isIndiaRule {
				if err := s.checkIndiaDrivingEligibility(ctx, driverID); err != nil {
					return nil, err
				}
			} else {
				if err := s.checkDrivingEligibility(ctx, driverID); err != nil {
					return nil, err
				}
			}
		}
	}

	// Create new duty status log
	log := &models.DutyStatusLog{
		ID:        uuid.New(),
		DriverID:  driverID,
		Status:    newStatus,
		Location:  location,
		StartedAt: time.Now(),
		Remarks:   remarks,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateDutyStatusLog(ctx, log); err != nil {
		return nil, fmt.Errorf("failed to create duty status log: %w", err)
	}

	// End previous log if exists
	if latestLog != nil && latestLog.EndedAt == nil {
		duration := int(log.StartedAt.Sub(latestLog.StartedAt).Minutes())
		if err := s.repo.EndDutyStatusLog(ctx, latestLog.ID, duration); err != nil {
			s.logger.Warn("Failed to end previous duty status log", zap.Error(err))
		}
	}

	s.logger.Info("Duty status updated",
		zap.String("driver_id", driverID.String()),
		zap.String("new_status", newStatus),
		zap.String("previous_status", latestLog.Status),
	)

	return log, nil
}

// validateStatusTransition validates FMCSA duty status transitions
func (s *DutyStatusService) validateStatusTransition(fromStatus, toStatus string) error {
	// Most transitions are allowed, but log warnings for unusual ones
	switch fromStatus {
	case StatusDriving:
		if toStatus == StatusDriving {
			// Continuous driving - check if exceeding 11 hours
			return nil // Will be checked in checkDrivingEligibility
		}
	case StatusOnDuty:
		if toStatus == StatusDriving {
			// On-duty to driving is normal
			return nil
		}
	}

	return nil
}

// checkDrivingEligibility checks if driver is eligible to drive per FMCSA rules
func (s *DutyStatusService) checkDrivingEligibility(ctx context.Context, driverID uuid.UUID) error {
	// Get driver's duty logs for current duty day
	now := time.Now()
	dutyDayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	logs, err := s.repo.GetDutyStatusLogs(ctx, driverID, dutyDayStart, now)
	if err != nil {
		return fmt.Errorf("failed to get duty logs: %w", err)
	}

	// Calculate driving hours in current duty day
	drivingHours := 0.0
	for _, log := range logs {
		if log.Status == StatusDriving {
			duration := int(now.Sub(log.StartedAt).Minutes())
			log.DurationMinutes = &duration
		}
		if log.DurationMinutes != nil {
			drivingHours += float64(*log.DurationMinutes) / 60.0
		}
	}

	// Check 11-hour driving limit
	if drivingHours >= MaxDrivingHours {
		return fmt.Errorf("driver has exceeded %d-hour driving limit", MaxDrivingHours)
	}

	// Check 14-hour on-duty limit
	onDutyHours := 0.0
	for _, log := range logs {
		if log.Status == StatusDriving || log.Status == StatusOnDuty {
			if log.DurationMinutes == nil {
				duration := int(now.Sub(log.StartedAt).Minutes())
				log.DurationMinutes = &duration
			}
			onDutyHours += float64(*log.DurationMinutes) / 60.0
		}
	}

	if onDutyHours >= MaxDutyHours {
		return fmt.Errorf("driver has exceeded %d-hour on-duty limit", MaxDutyHours)
	}

	// Check 30-minute break requirement
	lastBreakTime, err := s.getLastBreakTime(ctx, driverID)
	if err != nil {
		return fmt.Errorf("failed to get last break time: %w", err)
	}

	timeSinceBreak := now.Sub(lastBreakTime).Hours()
	if timeSinceBreak > 8 {
		return fmt.Errorf("driver requires 30-minute break after 8 hours")
	}

	return nil
}

// getLastBreakTime gets the time of driver's last qualifying break
func (s *DutyStatusService) getLastBreakTime(ctx context.Context, driverID uuid.UUID) (time.Time, error) {
	// Get recent duty logs
	now := time.Now()
	since := now.Add(-24 * time.Hour)

	logs, err := s.repo.GetDutyStatusLogs(ctx, driverID, since, now)
	if err != nil {
		return time.Time{}, err
	}

	// Find last qualifying break (30+ minutes off-duty or sleeper berth)
	for i := len(logs) - 1; i >= 0; i-- {
		log := logs[i]
		if (log.Status == StatusOffDuty || log.Status == StatusSleeperBerth) &&
			log.DurationMinutes != nil && *log.DurationMinutes >= RequiredRestBreak {
			return log.StartedAt, nil
		}
	}

	// No qualifying break found, return start of duty day
	now = time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
}

// GetDriverHOSStatus gets driver's current Hours of Service status
func (s *DutyStatusService) GetDriverHOSStatus(ctx context.Context, driverID uuid.UUID) (*HOSStatus, error) {
	now := time.Now()
	dutyDayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := dutyDayStart.AddDate(0, 0, -7)

	logs, err := s.repo.GetDutyStatusLogs(ctx, driverID, weekStart, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get duty logs: %w", err)
	}

	status := &HOSStatus{
		DriverID: driverID,
		AsOf:     now,
	}

	// Calculate hours for current duty day
	for _, log := range logs {
		if log.StartedAt.After(dutyDayStart) {
			duration := 0.0
			if log.DurationMinutes != nil {
				duration = float64(*log.DurationMinutes) / 60.0
			} else if log.Status == StatusDriving || log.Status == StatusOnDuty {
				// Current session
				duration = now.Sub(log.StartedAt).Hours()
			}

			switch log.Status {
			case StatusDriving:
				status.DrivingHoursToday += duration
			case StatusOnDuty:
				status.OnDutyHoursToday += duration
			}
		}

		// Calculate 7-day total
		if log.Status == StatusDriving || log.Status == StatusOnDuty {
			duration := 0.0
			if log.DurationMinutes != nil {
				duration = float64(*log.DurationMinutes) / 60.0
			}
			status.OnDutyHours7Days += duration
		}
	}

	// Calculate remaining hours
	status.DrivingHoursRemaining = MaxDrivingHours - status.DrivingHoursToday
	status.OnDutyHoursRemaining = MaxDutyHours - status.OnDutyHoursToday
	status.OnDutyHours7DaysRemaining = MaxOnDutyDays - status.OnDutyHours7Days

	// Determine compliance status
	if status.DrivingHoursRemaining <= 0 || status.OnDutyHoursRemaining <= 0 {
		status.ComplianceStatus = "violation"
	} else if status.DrivingHoursRemaining < 2 || status.OnDutyHoursRemaining < 2 {
		status.ComplianceStatus = "warning"
	} else {
		status.ComplianceStatus = "compliant"
	}

	return status, nil
}

// HOSStatus represents driver's Hours of Service status
type HOSStatus struct {
	DriverID                  uuid.UUID `json:"driver_id"`
	AsOf                      time.Time `json:"as_of"`
	DrivingHoursToday         float64   `json:"driving_hours_today"`
	DrivingHoursRemaining     float64   `json:"driving_hours_remaining"`
	OnDutyHoursToday          float64   `json:"on_duty_hours_today"`
	OnDutyHoursRemaining      float64   `json:"on_duty_hours_remaining"`
	OnDutyHours7Days          float64   `json:"on_duty_hours_7_days"`
	OnDutyHours7DaysRemaining float64   `json:"on_duty_hours_7_days_remaining"`
	ComplianceStatus          string    `json:"compliance_status"` // compliant, warning, violation
}

// GetDutyStatusHistory retrieves driver's duty status history
func (s *DutyStatusService) GetDutyStatusHistory(ctx context.Context, driverID uuid.UUID, startDate, endDate time.Time) ([]*models.DutyStatusLog, error) {
	return s.repo.GetDutyStatusLogs(ctx, driverID, startDate, endDate)
}

// GenerateDailyHOSReport generates a daily HOS report for a driver
func (s *DutyStatusService) GenerateDailyHOSReport(ctx context.Context, driverID uuid.UUID, date time.Time) (*HOSReport, error) {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	dayEnd := dayStart.Add(24 * time.Hour)

	logs, err := s.repo.GetDutyStatusLogs(ctx, driverID, dayStart, dayEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get duty logs: %w", err)
	}

	report := &HOSReport{
		DriverID: driverID,
		Date:     date,
		Logs:     logs,
	}

	// Calculate totals
	for _, log := range logs {
		duration := 0.0
		if log.DurationMinutes != nil {
			duration = float64(*log.DurationMinutes) / 60.0
		}

		switch log.Status {
		case StatusDriving:
			report.TotalDrivingHours += duration
		case StatusOnDuty:
			report.TotalOnDutyHours += duration
		case StatusOffDuty:
			report.TotalOffDutyHours += duration
		case StatusSleeperBerth:
			report.TotalSleeperBerthHours += duration
		}
	}

	// Check violations
	if report.TotalDrivingHours > MaxDrivingHours {
		report.Violations = append(report.Violations, fmt.Sprintf("Exceeded %d-hour driving limit", MaxDrivingHours))
	}
	if report.TotalOnDutyHours > MaxDutyHours {
		report.Violations = append(report.Violations, fmt.Sprintf("Exceeded %d-hour on-duty limit", MaxDutyHours))
	}

	return report, nil
}

// HOSReport represents a daily HOS report
type HOSReport struct {
	DriverID               uuid.UUID               `json:"driver_id"`
	Date                   time.Time               `json:"date"`
	Logs                   []*models.DutyStatusLog `json:"logs"`
	TotalDrivingHours      float64                 `json:"total_driving_hours"`
	TotalOnDutyHours       float64                 `json:"total_on_duty_hours"`
	TotalOffDutyHours      float64                 `json:"total_off_duty_hours"`
	TotalSleeperBerthHours float64                 `json:"total_sleeper_berth_hours"`
	Violations             []string                `json:"violations,omitempty"`
}

// checkIndiaDrivingEligibility checks if driver is eligible to drive per Indian MTW Act rules
func (s *DutyStatusService) checkIndiaDrivingEligibility(ctx context.Context, driverID uuid.UUID) error {
	now := time.Now()
	dutyDayStart := time.Now().Add(-24 * time.Hour)

	logs, err := s.repo.GetDutyStatusLogs(ctx, driverID, dutyDayStart, now)
	if err != nil {
		return fmt.Errorf("failed to get duty logs: %w", err)
	}

	for _, log := range logs {
		if log.Status == StatusDriving {
			driveDuration := now.Sub(log.StartedAt).Hours()
			if driveDuration >= IndiaMaxContinuousDriveTime {
				return fmt.Errorf("safety violation: mandatory rest required after 8h driving")
			}
			if driveDuration >= IndiaMandatoryBreakTime {
				hasBreak := false
				for _, b := range logs {
					if (b.Status == StatusRest || b.Status == StatusBreak || b.Status == StatusOffDuty) &&
						b.StartedAt.After(log.StartedAt) && b.DurationMinutes != nil && *b.DurationMinutes >= IndiaRequiredBreakDuration {
						hasBreak = true
						break
					}
				}
				if !hasBreak {
					return fmt.Errorf("safety violation: 30-minute break required after 5h driving")
				}
			}
		}

		if log.Status != StatusOffDuty {
			dutyDuration := now.Sub(log.StartedAt).Hours()
			if dutyDuration >= IndiaMaxDutyHours {
				return fmt.Errorf("safety violation: duty window exceeded 12h")
			}
		}
	}

	return nil
}
