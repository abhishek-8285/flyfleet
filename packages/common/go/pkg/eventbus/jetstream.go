package eventbus

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// JetStreamClient wraps the NATS JetStream setup
type JetStreamClient struct {
	nc     *nats.Conn
	js     nats.JetStreamContext
	logger *zap.Logger
}

// NewJetStreamClient creates a new robust NATS JetStream client
func NewJetStreamClient(natsURL string, logger *zap.Logger) (*JetStreamClient, error) {
	nc, err := nats.Connect(natsURL, nats.RetryOnFailedConnect(true), nats.MaxReconnects(10), nats.ReconnectWait(2*time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to init JetStream: %w", err)
	}

	return &JetStreamClient{
		nc:     nc,
		js:     js,
		logger: logger,
	}, nil
}

// Publish is a helper to publish events robustly
func (c *JetStreamClient) Publish(ctx context.Context, subject string, msgId string, data []byte) error {
	msg := &nats.Msg{
		Subject: subject,
		Data:    data,
		Header:  nats.Header{},
	}
	if msgId != "" {
		msg.Header.Set(nats.MsgIdHdr, msgId) // Built-in JetStream deduplication on the server side
	}
	
	_, err := c.js.PublishMsg(msg, nats.Context(ctx))
	return err
}

// SubscribeIdempotent subscribes to a subject and prevents duplicate processing using Redis replay guards
func (c *JetStreamClient) SubscribeIdempotent(
	subject string,
	durableName string,
	redisClient *redis.Client,
	handler func(msg *nats.Msg) error,
) (*nats.Subscription, error) {

	cb := func(m *nats.Msg) {
		msgId := m.Header.Get(nats.MsgIdHdr)
		if msgId != "" {
			cacheKey := "event_processed:" + msgId
			
			// Try to set the key using SetNX. If it exists, it's a duplicate.
			// Replay Guard!
			ok, err := redisClient.SetNX(context.Background(), cacheKey, "1", 24*time.Hour).Result()
			if err != nil {
				c.logger.Error("Failed to check Redis for idempotency guard", zap.Error(err))
				m.Nak() // Try later
				return
			}
			
			if !ok {
				// Duplicate detected. Just ack it and drop it.
				c.logger.Debug("Dropping duplicate event", zap.String("msg_id", msgId))
				m.Ack()
				return
			}
		}

		// Execute handler
		if err := handler(m); err != nil {
			c.logger.Error("Event handler failed", zap.Error(err))
			m.Nak() // Nack triggers a redelivery attempt depending on JetStream consumer config
			
			// Remove from redis so it can be re-tried
			if msgId != "" {
				redisClient.Del(context.Background(), "event_processed:"+msgId)
			}
			return
		}
		
		// Success
		m.Ack()
	}

	return c.js.QueueSubscribe(subject, durableName+"-queue", cb, nats.Durable(durableName), nats.ManualAck())
}

// Close gracefully disconnects from NATS
func (c *JetStreamClient) Close() {
	if c.nc != nil {
		c.nc.Close()
	}
}
