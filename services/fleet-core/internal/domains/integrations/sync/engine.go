package sync

import (
	"context"
	"fmt"

	"github.com/flyfleet/fleet-core/internal/domains/integrations/connectors"
)

type SyncEngine struct {
	connectors map[string]connectors.ERPConnector
}

func NewSyncEngine() *SyncEngine {
	return &SyncEngine{
		connectors: make(map[string]connectors.ERPConnector),
	}
}

func (e *SyncEngine) RegisterConnector(name string, connector connectors.ERPConnector) {
	e.connectors[name] = connector
}

func (e *SyncEngine) TriggerOrderSync(ctx context.Context, connectorName, orderID string) error {
	connector, ok := e.connectors[connectorName]
	if !ok {
		return fmt.Errorf("connector %s not found", connectorName)
	}

	// In a real scenario, fetch order data from DB here
	data := map[string]interface{}{
		"order_id": orderID,
		"amount":   1000.00,
		"currency": "USD",
	}

	fmt.Printf("Starting sync for Order %s using %s...\n", orderID, connectorName)
	result, err := connector.PushOrder(ctx, orderID, data)
	if err != nil {
		return fmt.Errorf("sync error: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("sync failed: %s", result.Message)
	}

	fmt.Printf("Sync successful: %s\n", result.Message)
	return nil
}
