//go:build integration
// +build integration

package integrations

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
)

// mockConnector embeds BaseConnector for testing semaphore integration.
type mockConnector struct {
	BaseConnector
	activeCalls int
	mu          sync.Mutex
	maxSeen     int
}

func (m *mockConnector) SendEvent(ctx context.Context, payload interface{}) error {
	// Attempt to acquire semaphore just like real connector
	if err := m.sem.Acquire(ctx, 1); err != nil {
		return err
	}
	defer m.sem.Release(1)

	m.mu.Lock()
	m.activeCalls++
	if m.activeCalls > m.maxSeen {
		m.maxSeen = m.activeCalls
	}
	m.mu.Unlock()

	// Simulate long network call
	time.Sleep(50 * time.Millisecond)

	m.mu.Lock()
	m.activeCalls--
	m.mu.Unlock()

	return nil
}

func TestConnectorIsolation_ConcurrencyLimits(t *testing.T) {
	logger := zaptest.NewLogger(t)

	mock := &mockConnector{}
	mock.logger = logger
	
	// Create strict concurrency limit of 3
	cfg := ConnectorConfig{
		MaxConcurrency: 3,
		Timeout:        1 * time.Second,
	}
	_ = mock.Initialize(cfg)

	var wg sync.WaitGroup
	ctx := context.Background()

	// Launch 10 concurrent requests
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = mock.SendEvent(ctx, nil)
		}()
	}

	wg.Wait()

	// Assert that despite 10 goroutines attempting to run simultaneously,
	// the semaphore isolated them to a strict max of 3 at any given time.
	if mock.maxSeen != 3 {
		t.Errorf("expected max concurrency seen to be exactly 3, got %d", mock.maxSeen)
	}
}
