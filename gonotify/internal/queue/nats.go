package queue

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	streamName    = "GN"
	streamSubject = "gn.>"
)

type NATSBus struct {
	nc     *nats.Conn
	js     nats.JetStreamContext
	logger *slog.Logger

	mu   sync.Mutex
	subs []*nats.Subscription
}

func NewNATSBus(url string, logger *slog.Logger) (*NATSBus, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, err
	}

	_, err = js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{streamSubject},
		Storage:  nats.FileStorage,
	})
	if err != nil {
		// If the stream already exists, that is acceptable.
		if _, infoErr := js.StreamInfo(streamName); infoErr != nil {
			nc.Close()
			return nil, err
		}
	}

	return &NATSBus{
		nc:     nc,
		js:     js,
		logger: logger,
	}, nil
}

func (b *NATSBus) Publish(ctx context.Context, subject string, msg Message) error {
	if msg.ID == "" {
		return errors.New("queue message id is required")
	}

	if msg.OccurredAt.IsZero() {
		msg.OccurredAt = time.Now().UTC()
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = b.js.Publish(subject, data, nats.MsgId(msg.ID))
	if err != nil {
		return err
	}

	b.logger.DebugContext(
		ctx,
		"published message",
		"subject", subject,
		"message_id", msg.ID,
		"tenant_id", msg.TenantID,
	)

	return nil
}

func (b *NATSBus) Subscribe(subject string, queueGroup string, handler Handler) error {
	sub, err := b.js.QueueSubscribe(
		subject,
		queueGroup,
		func(m *nats.Msg) {
			var msg Message

			if err := json.Unmarshal(m.Data, &msg); err != nil {
				b.logger.Error(
					"failed to unmarshal queue message",
					"subject", subject,
					"error", err,
				)

				// Poison message. Terminate redelivery.
				_ = m.Term()
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := handler(ctx, msg); err != nil {
				b.logger.ErrorContext(
					ctx,
					"queue handler failed",
					"subject", subject,
					"message_id", msg.ID,
					"tenant_id", msg.TenantID,
					"error", err,
				)

				_ = m.Nak()
				return
			}

			_ = m.Ack()
		},
		nats.Durable(queueGroup),
		nats.ManualAck(),
		nats.AckExplicit(),
		nats.MaxDeliver(5),
		nats.AckWait(30*time.Second),
	)
	if err != nil {
		return err
	}

	b.mu.Lock()
	b.subs = append(b.subs, sub)
	b.mu.Unlock()

	b.logger.Info(
		"subscribed to subject",
		"subject", subject,
		"queue_group", queueGroup,
	)

	return nil
}

func (b *NATSBus) PublishTransient(subject string, data []byte) error {
	if b.nc == nil {
		return errors.New("nats connection is not initialized")
	}

	return b.nc.Publish(subject, data)
}

func (b *NATSBus) SubscribeTransient(
	subject string,
	handler func(data []byte),
) (func(), error) {
	if b.nc == nil {
		return nil, errors.New("nats connection is not initialized")
	}

	sub, err := b.nc.Subscribe(subject, func(m *nats.Msg) {
		handler(m.Data)
	})
	if err != nil {
		return nil, err
	}

	return func() {
		_ = sub.Unsubscribe()
	}, nil
}

func (b *NATSBus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, sub := range b.subs {
		_ = sub.Unsubscribe()
	}

	b.nc.Close()
	return nil
}
