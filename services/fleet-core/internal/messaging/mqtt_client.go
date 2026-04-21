package messaging

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

// MQTTClient wraps MQTT connection for fleet IoT
type MQTTClient struct {
	client mqtt.Client
	config MQTTConfig
	logger *zap.Logger
}

// MQTTConfig holds MQTT broker configuration
type MQTTConfig struct {
	BrokerURL      string        `json:"broker_url" yaml:"broker_url"`
	ClientID       string        `json:"client_id" yaml:"client_id"`
	Username       string        `json:"username" yaml:"username"`
	Password       string        `json:"password" yaml:"password"`
	QoS            byte          `json:"qos" yaml:"qos"`
	Retain         bool          `json:"retain" yaml:"retain"`
	CleanSession   bool          `json:"clean_session" yaml:"clean_session"`
	ConnectTimeout time.Duration `json:"connect_timeout" yaml:"connect_timeout"`
	KeepAlive      time.Duration `json:"keep_alive" yaml:"keep_alive"`
	TLSEnabled     bool          `json:"tls_enabled" yaml:"tls_enabled"`
}

// DefaultMQTTConfig returns sensible defaults for MQTT configuration
func DefaultMQTTConfig() MQTTConfig {
	return MQTTConfig{
		QoS:            1,
		Retain:         false,
		CleanSession:   false,
		ConnectTimeout: 30 * time.Second,
		KeepAlive:      60 * time.Second,
		TLSEnabled:     true,
	}
}

// MQTTMessage represents a standardized MQTT message
type MQTTMessage struct {
	Topic     string                 `json:"topic"`
	Payload   []byte                 `json:"payload"`
	QoS       byte                   `json:"qos"`
	Retained  bool                   `json:"retained"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewMQTTClient creates a new MQTT client with auto-reconnect
func NewMQTTClient(cfg MQTTConfig, logger *zap.Logger) (*MQTTClient, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.BrokerURL)
	opts.SetClientID(cfg.ClientID)
	opts.SetUsername(cfg.Username)
	opts.SetPassword(cfg.Password)
	opts.SetCleanSession(cfg.CleanSession)
	opts.SetConnectTimeout(cfg.ConnectTimeout)
	opts.SetKeepAlive(cfg.KeepAlive)

	// Auto-reconnect configuration
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.SetMaxReconnectInterval(30 * time.Second)
	opts.SetOrderMatters(false) // Allow parallel message processing

	// Default message handler (for unexpected messages)
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		if logger != nil {
			logger.Debug("received unexpected mqtt message",
				zap.String("topic", msg.Topic()),
				zap.ByteString("payload", msg.Payload()))
		}
	})

	// TLS configuration
	if cfg.TLSEnabled {
		opts.SetTLSConfig(&tls.Config{
			MinVersion: tls.VersionTLS12,
		})
	}

	// Connection handlers
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		if logger != nil {
			logger.Info("mqtt client connected",
				zap.String("client_id", cfg.ClientID),
				zap.String("broker", cfg.BrokerURL))
		}
	})

	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		if logger != nil {
			logger.Warn("mqtt connection lost",
				zap.String("client_id", cfg.ClientID),
				zap.Error(err))
		}
	})

	opts.SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
		if logger != nil {
			logger.Info("mqtt client reconnecting",
				zap.String("client_id", cfg.ClientID))
		}
	})

	client := mqtt.NewClient(opts)

	// Connect to broker
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to mqtt broker: %w", token.Error())
	}

	// Wait for connection to be established
	time.Sleep(100 * time.Millisecond)

	return &MQTTClient{
		client: client,
		config: cfg,
		logger: logger,
	}, nil
}

// Publish publishes a message to a topic
func (c *MQTTClient) Publish(topic string, payload []byte) error {
	return c.PublishWithQoS(topic, payload, c.config.QoS, c.config.Retain)
}

// PublishWithQoS publishes a message with specific QoS and retain settings
func (c *MQTTClient) PublishWithQoS(topic string, payload []byte, qos byte, retain bool) error {
	token := c.client.Publish(topic, qos, retain, payload)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("failed to publish to %s: %w", topic, token.Error())
	}

	if c.logger != nil {
		c.logger.Debug("mqtt message published",
			zap.String("topic", topic),
			zap.Int("payload_size", len(payload)),
			zap.Uint8("qos", qos))
	}

	return nil
}

// PublishJSON publishes a JSON-serialized message
func (c *MQTTClient) PublishJSON(topic string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	return c.Publish(topic, payload)
}

// Subscribe subscribes to a topic with the default handler
func (c *MQTTClient) Subscribe(topic string, handler func([]byte)) error {
	return c.SubscribeWithQoS(topic, handler, c.config.QoS)
}

// SubscribeWithQoS subscribes to a topic with specific QoS
func (c *MQTTClient) SubscribeWithQoS(topic string, handler func([]byte), qos byte) error {
	token := c.client.Subscribe(topic, qos, func(client mqtt.Client, msg mqtt.Message) {
		handler(msg.Payload())
	})
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", topic, token.Error())
	}

	if c.logger != nil {
		c.logger.Info("mqtt subscription established",
			zap.String("topic", topic),
			zap.Uint8("qos", qos))
	}

	return nil
}

// SubscribeWithContext subscribes with context for cancellation
func (c *MQTTClient) SubscribeWithContext(ctx context.Context, topic string, handler func(ctx context.Context, payload []byte)) error {
	return c.SubscribeWithQoS(topic, func(payload []byte) {
		handler(ctx, payload)
	}, c.config.QoS)
}

// Unsubscribe unsubscribes from a topic
func (c *MQTTClient) Unsubscribe(topic string) error {
	token := c.client.Unsubscribe(topic)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("failed to unsubscribe from %s: %w", topic, token.Error())
	}

	if c.logger != nil {
		c.logger.Debug("mqtt subscription removed", zap.String("topic", topic))
	}
	return nil
}

// IsConnected returns true if the client is connected to the broker
func (c *MQTTClient) IsConnected() bool {
	return c.client.IsConnected()
}

// Close closes the MQTT connection gracefully
func (c *MQTTClient) Close() {
	if c.logger != nil {
		c.logger.Info("closing mqtt connection", zap.String("client_id", c.config.ClientID))
	}
	c.client.Disconnect(1000) // Wait 1 second for pending messages
}

// GetClient returns the underlying MQTT client for advanced operations
func (c *MQTTClient) GetClient() mqtt.Client {
	return c.client
}

// GetConfig returns the MQTT configuration
func (c *MQTTClient) GetConfig() MQTTConfig {
	return c.config
}

// PublishVehicleTelemetry publishes vehicle telemetry data
func (c *MQTTClient) PublishVehicleTelemetry(tenantID, vehicleID string, telemetry interface{}) error {
	topic := fmt.Sprintf("fleet/%s/vehicles/%s/telemetry", tenantID, vehicleID)
	return c.PublishJSON(topic, telemetry)
}

// PublishVehicleStatus publishes vehicle status update
func (c *MQTTClient) PublishVehicleStatus(tenantID, vehicleID string, status interface{}) error {
	topic := fmt.Sprintf("fleet/%s/vehicles/%s/status", tenantID, vehicleID)
	return c.PublishJSON(topic, status)
}

// PublishVehicleAlert publishes a vehicle alert
func (c *MQTTClient) PublishVehicleAlert(tenantID, vehicleID string, alert interface{}) error {
	topic := fmt.Sprintf("fleet/%s/vehicles/%s/alerts", tenantID, vehicleID)
	return c.PublishWithQoS(topic, nil, 2, false) // QoS 2 for alerts
}

// SendCommand publishes a command to a vehicle
func (c *MQTTClient) SendCommand(tenantID, vehicleID string, command interface{}) error {
	topic := fmt.Sprintf("fleet/%s/commands/%s", tenantID, vehicleID)
	return c.PublishWithQoS(topic, nil, 2, false) // QoS 2 for commands
}

// SubscribeToCommands subscribes to command responses for a vehicle
func (c *MQTTClient) SubscribeToCommands(tenantID, vehicleID string, handler func([]byte)) error {
	topic := fmt.Sprintf("fleet/%s/responses/%s", tenantID, vehicleID)
	return c.Subscribe(topic, handler)
}
