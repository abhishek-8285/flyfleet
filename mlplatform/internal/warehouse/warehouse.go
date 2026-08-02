package warehouse

import "context"

// Warehouse is the offline analytics store interface.
// The concrete implementation lives in ClickHouseStore.
type Warehouse interface {
	Migrate(ctx context.Context) error
}
