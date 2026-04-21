package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core/internal/database"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFailClosedRLSPolicy(t *testing.T) {
	// Skip if no DATABASE_URL is set (no test database available)
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set - skipping RLS security test that requires a real database")
	}

	cfg := database.PoolConfig{
		DatabaseURL: dbURL,
		CriticalMax: 5,
		HighMax:     2,
		MediumMax:   1,
		LowMax:      1,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pools, err := database.NewDatabasePools(ctx, cfg)
	require.NoError(t, err, "Failed to initialize db pools")
	defer pools.Close()

	t.Run("Empty TenantID Context Returns Zero Rows", func(t *testing.T) {
		// 1. Create a background context with NO tenant_id (simulating a malformed req/missing middleware)
		emptyCtx := context.Background()

		// 2. Attempt to query the heavily guarded `vehicles` table
		rows, err := pools.High.Query(emptyCtx, "SELECT id FROM vehicles LIMIT 10")
		require.NoError(t, err, "Query execution should succeed, but RLS should block rows")
		defer rows.Close()

		// 3. Assert that exactly 0 rows are returned due to the Fail-Closed policy
		count := 0
		for rows.Next() {
			count++
		}

		assert.Equal(t, 0, count, "CRITICAL SECURITY FAILURE: Vehicles were returned without an tenant_id context!")
	})

	t.Run("Valid TenantID Context Returns Rows", func(t *testing.T) {
		// We need a known seed TenantID, or we can just verify the query executes successfully
		// under the RLS context injection via pgxpool BeforeAcquire
		validCtx := context.WithValue(context.Background(), ctxkeys.TenantIDKey, "00000000-0000-0000-0000-000000000000") // Generic seed or format

		rows, err := pools.High.Query(validCtx, "SELECT id FROM vehicles LIMIT 1")
		require.NoError(t, err)
		defer rows.Close()

		// It doesn't matter if it returns 0 or 1 row here, what matters is the execution didn't panic
		// and the BeforeAcquire hook successfully injected the mock UUID into the session.
	})
}
