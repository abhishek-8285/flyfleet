package events

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gonotify/gonotify/internal/observability"
	"github.com/gonotify/gonotify/internal/queue"
)

type PublishRequest struct {
	TenantID       string
	EventType      string
	UserID         string
	IdempotencyKey string
	Data           map[string]any
	Channels       []string
	Metadata       map[string]any
}

type PublishResponse struct {
	EventID string `json:"event_id"`
	Status  string `json:"status"`
}

type Service struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewService(db *pgxpool.Pool, logger *slog.Logger) *Service {
	return &Service{
		db:     db,
		logger: logger,
	}
}

func (s *Service) Publish(ctx context.Context, req PublishRequest) (PublishResponse, error) {
	if req.EventType == "" {
		return PublishResponse{}, errors.New("event type is required")
	}

	if req.TenantID == "" {
		return PublishResponse{}, errors.New("tenant id is required")
	}

	if req.Data == nil {
		req.Data = map[string]any{}
	}

	if req.Metadata == nil {
		req.Metadata = map[string]any{}
	}

	if req.IdempotencyKey != "" {
		existing, err := s.findByIdempotencyKey(ctx, req.TenantID, req.IdempotencyKey)
		if err == nil {
			observability.EventsIngested.WithLabelValues(req.TenantID, req.EventType, "deduplicated").Inc()

			return PublishResponse{
				EventID: existing.ID,
				Status:  existing.Status,
			}, nil
		}

		if !errors.Is(err, pgx.ErrNoRows) {
			return PublishResponse{}, err
		}
	}

	payload, err := json.Marshal(req.Data)
	if err != nil {
		return PublishResponse{}, err
	}

	metadata, err := json.Marshal(req.Metadata)
	if err != nil {
		return PublishResponse{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return PublishResponse{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var eventID string

	err = tx.QueryRow(
		ctx,
		`
		INSERT INTO events (
			tenant_id,
			event_type,
			user_id,
			idempotency_key,
			payload,
			metadata,
			status,
			received_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, 'received', now())
		RETURNING id::text
		`,
		req.TenantID,
		req.EventType,
		req.UserID,
		req.IdempotencyKey,
		payload,
		metadata,
	).Scan(&eventID)
	if err != nil {
		return PublishResponse{}, err
	}

	queuePayload, err := json.Marshal(req.Data)
	if err != nil {
		return PublishResponse{}, err
	}

	queueMessage := queue.Message{
		ID:         eventID,
		TenantID:   req.TenantID,
		Type:       "event.received",
		OccurredAt: time.Now().UTC(),
		Payload:    queuePayload,
		Metadata: map[string]string{
			"event_type": req.EventType,
			"user_id":    req.UserID,
		},
	}

	if len(req.Channels) > 0 {
		queueMessage.Metadata["channels"] = strings.Join(req.Channels, ",")
	}

	queueMessageBytes, err := json.Marshal(queueMessage)
	if err != nil {
		return PublishResponse{}, err
	}

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO outbox_messages (
			tenant_id,
			aggregate_type,
			aggregate_id,
			topic,
			payload
		)
		VALUES ($1, 'event', $2, 'gn.events.received', $3)
		`,
		req.TenantID,
		eventID,
		queueMessageBytes,
	)
	if err != nil {
		return PublishResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return PublishResponse{}, err
	}

	observability.EventsIngested.WithLabelValues(req.TenantID, req.EventType, "accepted").Inc()

	s.logger.InfoContext(
		ctx,
		"event accepted",
		"event_id", eventID,
		"tenant_id", req.TenantID,
		"event_type", req.EventType,
	)

	return PublishResponse{
		EventID: eventID,
		Status:  "accepted",
	}, nil
}

type existingEvent struct {
	ID     string
	Status string
}

func (s *Service) findByIdempotencyKey(
	ctx context.Context,
	tenantID string,
	idempotencyKey string,
) (existingEvent, error) {
	var event existingEvent

	err := s.db.QueryRow(
		ctx,
		`
		SELECT
			id::text,
			status
		FROM events
		WHERE tenant_id = $1
		  AND idempotency_key = $2
		`,
		tenantID,
		idempotencyKey,
	).Scan(
		&event.ID,
		&event.Status,
	)

	return event, err
}
