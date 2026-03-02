package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/admin_support_v2/domains/emergency/models"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type EmergencyService interface {
	GlobalHalt(ctx context.Context, reason string, issuedBy string) (*models.HaltCommand, error)
	GetStatus(ctx context.Context) (*models.EmergencyStatus, error)
}

type emergencyService struct {
	nc     *nats.Conn
	logger *zap.Logger
}

func NewEmergencyService(nc *nats.Conn, logger *zap.Logger) EmergencyService {
	return &emergencyService{nc: nc, logger: logger}
}

func (s *emergencyService) GlobalHalt(ctx context.Context, reason string, issuedBy string) (*models.HaltCommand, error) {
	command := &models.HaltCommand{
		CommandID: fmt.Sprintf("HALT-%d", time.Now().Unix()),
		Type:      "GLOBAL_HALT",
		IssuedAt:  time.Now(),
		Reason:    reason,
		Issuer:    issuedBy,
	}

	data, err := json.Marshal(command)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal halt command: %w", err)
	}

	// Broadcast to fleet.cmd.halt
	// In a real system we might use JetStream, but standard Pub/Sub works for immediate global actions.
	if err := s.nc.Publish("fleet.cmd.halt", data); err != nil {
		s.logger.Error("Failed to publish halt command to NATS", zap.Error(err))
		return nil, fmt.Errorf("failed to broadcast halt command: %w", err)
	}

	s.logger.Warn("GLOBAL FLEET HALT ISSUED", 
		zap.String("id", command.CommandID),
		zap.String("issuer", issuedBy),
		zap.String("reason", reason))

	return command, nil
}

func (s *emergencyService) GetStatus(ctx context.Context) (*models.EmergencyStatus, error) {
	// For now, return a default status. In a real system, this would be backed by DB or Redis.
	return &models.EmergencyStatus{
		Active: false,
	}, nil
}
