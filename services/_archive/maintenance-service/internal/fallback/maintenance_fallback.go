package fallback

import (
	"context"
	"errors"
	"time"

	"github.com/flyfleet/maintenance-service/internal/models"
)

// FallbackService provides fallback mechanisms for maintenance operations
type FallbackService struct {
	// In-memory storage for fallback operations
	fallbackSchedules map[string]*models.MaintenanceSchedule
	fallbackLogs      map[string]*models.MaintenanceLog
}

func NewFallbackService() *FallbackService {
	return &FallbackService{
		fallbackSchedules: make(map[string]*models.MaintenanceSchedule),
		fallbackLogs:      make(map[string]*models.MaintenanceLog),
	}
}

// FallbackCreateSchedule handles schedule creation when primary service is unavailable
func (f *FallbackService) FallbackCreateSchedule(ctx context.Context, schedule *models.MaintenanceSchedule) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Store in fallback storage
		f.fallbackSchedules[schedule.ID] = schedule
		return nil
	}
}

// FallbackCreateLog handles log creation when primary service is unavailable
func (f *FallbackService) FallbackCreateLog(ctx context.Context, log *models.MaintenanceLog) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Store in fallback storage
		f.fallbackLogs[log.ID] = log
		return nil
	}
}

// FallbackGetSchedule retrieves schedule from fallback storage
func (f *FallbackService) FallbackGetSchedule(ctx context.Context, id string) (*models.MaintenanceSchedule, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		schedule, exists := f.fallbackSchedules[id]
		if !exists {
			return nil, errors.New("schedule not found in fallback storage")
		}
		return schedule, nil
	}
}

// FallbackGetLog retrieves log from fallback storage
func (f *FallbackService) FallbackGetLog(ctx context.Context, id string) (*models.MaintenanceLog, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		log, exists := f.fallbackLogs[id]
		if !exists {
			return nil, errors.New("log not found in fallback storage")
		}
		return log, nil
	}
}

// HasSchedule checks if a schedule exists in fallback storage
func (f *FallbackService) HasSchedule(id string) bool {
	_, exists := f.fallbackSchedules[id]
	return exists
}

// HasLog checks if a log exists in fallback storage
func (f *FallbackService) HasLog(id string) bool {
	_, exists := f.fallbackLogs[id]
	return exists
}

// GetFallbackSchedules returns all stored schedules
func (f *FallbackService) GetFallbackSchedules() []*models.MaintenanceSchedule {
	var schedules []*models.MaintenanceSchedule
	for _, schedule := range f.fallbackSchedules {
		schedules = append(schedules, schedule)
	}
	return schedules
}

// GetFallbackLogs returns all stored logs
func (f *FallbackService) GetFallbackLogs() []*models.MaintenanceLog {
	var logs []*models.MaintenanceLog
	for _, log := range f.fallbackLogs {
		logs = append(logs, log)
	}
	return logs
}

// CleanupOldEntries removes entries older than specified duration
func (f *FallbackService) CleanupOldEntries(maxAge time.Duration) {
	now := time.Now()
	
	// Clean up old schedules
	for id, schedule := range f.fallbackSchedules {
		if now.Sub(schedule.CreatedAt) > maxAge {
			delete(f.fallbackSchedules, id)
		}
	}
	
	// Clean up old logs
	for id, log := range f.fallbackLogs {
		if now.Sub(log.CreatedAt) > maxAge {
			delete(f.fallbackLogs, id)
		}
	}
}

// RetryOperation attempts to execute an operation with retry logic
func (f *FallbackService) RetryOperation(operation func() error, maxRetries int, delay time.Duration) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = operation()
		if err == nil {
			return nil
		}
		
		if i < maxRetries-1 {
			time.Sleep(delay)
		}
	}
	return err
}
