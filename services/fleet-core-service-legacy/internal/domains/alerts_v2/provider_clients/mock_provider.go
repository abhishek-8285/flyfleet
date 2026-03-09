package providers

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/alerts_v2/models"
)

type mockProvider struct {
	name string
}

func NewMockProvider(name string) ProviderClient {
	return &mockProvider{name: name}
}

func (p *mockProvider) Name() string {
	return p.name
}

func (p *mockProvider) Send(ctx context.Context, log *models.NotificationLog) (map[string]any, error) {
	fmt.Printf("[%s] Sending to %s (Channel: %s): %s\n", p.name, log.Recipient, log.Channel, log.Subject)
	
	// Simulate some latency
	time.Sleep(100 * time.Millisecond)

	// Simulate flaky network (10% failure)
	if rand.Float32() < 0.1 {
		return nil, fmt.Errorf("simulated %s provider error", p.name)
	}

	return map[string]any{
		"message_id": fmt.Sprintf("mock-%d", time.Now().UnixNano()),
		"provider":   p.name,
		"status":     "delivered",
	}, nil
}
