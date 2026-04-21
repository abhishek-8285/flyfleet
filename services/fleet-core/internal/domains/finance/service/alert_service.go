package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

const (
	eventFinanceAnomalyDetected = "finance.anomaly.detected"
	eventFuelTheftDetected      = "fuel.theft.detected"
	eventBudgetLimitExceeded    = "budget.limit.exceeded"
)

type AlertingService struct {
	nc     *nats.Conn
	logger *zap.Logger
}

type AlertEvent struct {
	EventType  string                 `json:"event_type"`
	TenantID      uuid.UUID              `json:"tenant_id"`
	OccurredAt time.Time              `json:"occurred_at"`
	Payload    map[string]interface{} `json:"payload"`
}

func NewAlertingService(nc *nats.Conn, logger *zap.Logger) *AlertingService {
	s := &AlertingService{nc: nc, logger: logger}
	s.ensureJetStream()
	return s
}

func (s *AlertingService) Publish(ctx context.Context, eventType string, tenantID uuid.UUID, payload map[string]interface{}) error {
	if s == nil || s.nc == nil {
		return nil
	}

	data, err := json.Marshal(AlertEvent{
		EventType:  eventType,
		TenantID:      tenantID,
		OccurredAt: time.Now().UTC(),
		Payload:    payload,
	})
	if err != nil {
		return err
	}

	js, err := s.nc.JetStream()
	if err == nil {
		_, pubErr := js.PublishAsync(eventType, data)
		if pubErr == nil {
			return nil
		}
		s.logger.Warn("jetstream publish failed, falling back to core nats",
			zap.String("event_type", eventType),
			zap.Error(pubErr))
	}

	return s.nc.Publish(eventType, data)
}

func (s *AlertingService) ensureJetStream() {
	if s == nil || s.nc == nil {
		return
	}
	js, err := s.nc.JetStream()
	if err != nil {
		return
	}
	_, infoErr := js.StreamInfo("FLYFLEET_ALERTS")
	if infoErr == nil {
		return
	}
	_, _ = js.AddStream(&nats.StreamConfig{
		Name:     "FLYFLEET_ALERTS",
		Subjects: []string{eventFinanceAnomalyDetected, eventFuelTheftDetected, eventBudgetLimitExceeded},
		Storage:  nats.FileStorage,
		MaxAge:   7 * 24 * time.Hour,
	})
}
