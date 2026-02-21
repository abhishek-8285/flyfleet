package messaging

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// JetStream handles standardized event publish/subscribe
type JetStream struct {
	Conn nats.JetStreamContext
}

// NewJetStream creates a new JetStream wrapper
func NewJetStream(url string) (*JetStream, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	return &JetStream{Conn: js}, nil
}

// Publish publishes a versioned event to a subject
func (j *JetStream) Publish(ctx context.Context, subject string, data []byte, msgID string) error {
	_, err := j.Conn.Publish(subject, data, nats.MsgId(msgID))
	return err
}

// CreateStream ensures a stream exists with specified retention
func (j *JetStream) CreateStream(name string, subjects []string) error {
	_, err := j.Conn.AddStream(&nats.StreamConfig{
		Name:     name,
		Subjects: subjects,
		Storage:  nats.FileStorage,
		Retention: nats.LimitsPolicy,
		MaxAge:   30 * 24 * time.Hour, // 30 days retention
	})
	return err
}

// SubscribeDurably creates a durable consumer for a stream
func (j *JetStream) SubscribeDurably(subject, queue, durable string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return j.Conn.QueueSubscribe(subject, queue, handler, 
		nats.Durable(durable),
		nats.DeliverAll(),
		nats.AckExplicit(),
	)
}
