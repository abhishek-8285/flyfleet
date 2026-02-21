package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core-service/internal/domains/driver/models"
)

// UpdateDutyStatus updates the driver's duty status (HOOS)
func (s *DriverService) UpdateDutyStatus(ctx context.Context, driverID string, req *models.DutyStatusLog) error {
	// Validate status transition if needed
	// For now, just log the new status
	
	// Create log entry
	req.DriverID = driverID
	req.StartedAt = time.Now()
	
	// Get previous status to set EndedAt
	lastLog, err := s.repo.GetLatestDutyStatus(ctx, driverID)
	if err == nil && lastLog != nil {
		// This would ideally update the previous log's EndedAt
		// For MVP we just create a new log
	}

	if err := s.repo.CreateDutyStatusLog(ctx, req); err != nil {
		s.logger.Error("failed to update duty status", 
			zap.String("driver_id", driverID), 
			zap.String("status", req.Status), 
			zap.Error(err))
		return fmt.Errorf("failed to update duty status: %w", err)
	}

	s.logger.Info("driver duty status updated", 
		zap.String("driver_id", driverID), 
		zap.String("status", req.Status))
	
	s.publishEvent("driver.duty_status_changed", map[string]interface{}{
		"driver_id": driverID,
		"status":    req.Status,
		"location":  req.Location,
		"timestamp": req.StartedAt,
	})

	return nil
}

// SendMessage sends a message from driver to dispatch
func (s *DriverService) SendMessage(ctx context.Context, driverID string, content, msgType, attachmentURL string) (*models.DriverMessage, error) {
	msg := &models.DriverMessage{
		DriverID:      driverID,
		Direction:     "INBOUND",
		Type:          msgType,
		Content:       content,
		AttachmentURL: attachmentURL,
		IsRead:        false,
	}

	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		s.logger.Error("failed to send message", 
			zap.String("driver_id", driverID), 
			zap.Error(err))
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	s.publishEvent("driver.message_received", msg)
	return msg, nil
}

// GetMessages retrieves messages for a driver
func (s *DriverService) GetMessages(ctx context.Context, driverID string, page, pageSize int) ([]*models.DriverMessage, error) {
	if page < 1 { page = 1 }
	if pageSize < 1 { pageSize = 20 }
	offset := (page - 1) * pageSize

	return s.repo.GetMessages(ctx, driverID, pageSize, offset)
}

// SubmitExpense submits a new expense claim
func (s *DriverService) SubmitExpense(ctx context.Context, driverID string, req *models.DriverExpense) (*models.DriverExpense, error) {
	req.DriverID = driverID
	req.Status = "PENDING"
	
	if err := s.repo.CreateExpense(ctx, req); err != nil {
		s.logger.Error("failed to submit expense", 
			zap.String("driver_id", driverID), 
			zap.Error(err))
		return nil, fmt.Errorf("failed to submit expense: %w", err)
	}

	s.publishEvent("driver.expense_submitted", req)
	return req, nil
}

// GetExpenses retrieves expenses for a driver
func (s *DriverService) GetExpenses(ctx context.Context, driverID string) ([]*models.DriverExpense, error) {
	return s.repo.ListExpenses(ctx, driverID)
}

// TriggerSOS triggers an emergency SOS alert
func (s *DriverService) TriggerSOS(ctx context.Context, driverID string, location *models.Location) error {
	s.logger.Warn("SOS ALERT TRIGGERED", 
		zap.String("driver_id", driverID),
		zap.Any("location", location))

	// In a real system, this would trigger immediate notifications via SMS/Push/Email
	// largely handled by the alert-service listening to this event
	s.publishEvent("driver.sos_alert", map[string]interface{}{
		"driver_id": driverID,
		"location":  location,
		"timestamp": time.Now(),
		"severity":  "CRITICAL",
	})

	return nil
}
