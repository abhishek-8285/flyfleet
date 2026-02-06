package messaging

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type NATSClient struct {
	conn   *nats.Conn
	logger *zap.Logger
}

type MessageBroker interface {
	Publish(subject string, data interface{}) error
	Subscribe(subject string, handler func([]byte)) error
	Request(subject string, data interface{}, timeout time.Duration) ([]byte, error)
	Close() error
}

func NewNATSClient(url string, logger *zap.Logger) (MessageBroker, error) {
	conn, err := nats.Connect(url,
		nats.Name("FlyFleet"),
		nats.Timeout(10*time.Second),
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(5),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			logger.Warn("NATS disconnected", zap.Error(err))
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Info("NATS reconnected")
		}),
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	logger.Info("Connected to NATS", zap.String("url", url))

	return &NATSClient{
		conn:   conn,
		logger: logger,
	}, nil
}

func (nc *NATSClient) Publish(subject string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := nc.conn.Publish(subject, payload); err != nil {
		nc.logger.Error("Failed to publish message",
			zap.String("subject", subject),
			zap.Error(err))
		return err
	}

	nc.logger.Debug("Message published",
		zap.String("subject", subject),
		zap.Int("payload_size", len(payload)))

	return nil
}

func (nc *NATSClient) Subscribe(subject string, handler func([]byte)) error {
	_, err := nc.conn.Subscribe(subject, func(msg *nats.Msg) {
		nc.logger.Debug("Message received",
			zap.String("subject", subject),
			zap.Int("payload_size", len(msg.Data)))
		
		handler(msg.Data)
		
		// Acknowledge the message
		if err := msg.Ack(); err != nil {
			nc.logger.Error("Failed to acknowledge message", zap.Error(err))
		}
	})
	
	if err != nil {
		nc.logger.Error("Failed to subscribe to subject",
			zap.String("subject", subject),
			zap.Error(err))
		return err
	}

	nc.logger.Info("Subscribed to subject", zap.String("subject", subject))
	return nil
}

func (nc *NATSClient) Request(subject string, data interface{}, timeout time.Duration) ([]byte, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	msg, err := nc.conn.Request(subject, payload, timeout)
	if err != nil {
		nc.logger.Error("Request failed",
			zap.String("subject", subject),
			zap.Error(err))
		return nil, err
	}

	nc.logger.Debug("Request response received",
		zap.String("subject", subject),
		zap.Int("response_size", len(msg.Data)))

	return msg.Data, nil
}

func (nc *NATSClient) Close() error {
	nc.conn.Close()
	nc.logger.Info("NATS connection closed")
	return nil
}