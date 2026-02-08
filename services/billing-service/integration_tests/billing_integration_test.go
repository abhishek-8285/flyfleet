package integration_tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/flyfleet/billing-service/internal/models"
	"github.com/flyfleet/billing-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBillingRepositoryIntegration(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://fleet:fleet@localhost:5432/fleetdb?sslmode=disable"
		t.Log("DATABASE_URL not set, using default")
	}

	// Connect to database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("Skipping integration test: failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Verify connection
	err = pool.Ping(ctx)
	if err != nil {
		t.Skipf("Skipping integration test: database not reachable: %v", err)
	}

	repo := repository.NewBillingRepository(pool)
	orgID := uuid.New()

	t.Run("Create and Get Invoice", func(t *testing.T) {
		inv := &models.Invoice{
			OrgID:        orgID,
			CustomerName: "Integration Test Corp",
			InvoiceDate:  time.Now(),
			DueDate:      time.Now().Add(14 * 24 * time.Hour),
			Status:       models.StatusPending,
			Currency:     "USD",
			TotalAmount:  1000.0,
			BalanceDue:   1000.0,
			Items: []models.InvoiceItem{
				{
					Description: "Test Service",
					Quantity:    1,
					UnitPrice:   1000.0,
					TaxPercent:   0,
					Total:       1000.0,
				},
			},
		}

		err := repo.CreateInvoice(ctx, inv)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, inv.ID)

		// Get it back
		fetched, err := repo.GetInvoice(ctx, inv.ID)
		require.NoError(t, err)
		require.NotNil(t, fetched)
		assert.Equal(t, inv.CustomerName, fetched.CustomerName)
		assert.Equal(t, 1, len(fetched.Items))
		assert.Equal(t, "Test Service", fetched.Items[0].Description)
	})

	t.Run("Billing Settings Persistence", func(t *testing.T) {
		settings := &models.BillingSettings{
			OrgID:                 orgID,
			DefaultTaxRate:        12.5,
			BaseCurrency:          "EUR",
			GracePeriodDays:       30,
			AutoInvoicingEnabled: true,
		}

		err := repo.UpdateSettings(ctx, settings)
		require.NoError(t, err)

		fetched, err := repo.GetSettings(ctx, orgID)
		require.NoError(t, err)
		require.NotNil(t, fetched)
		assert.Equal(t, 12.5, fetched.DefaultTaxRate)
		assert.Equal(t, "EUR", fetched.BaseCurrency)
		assert.True(t, fetched.AutoInvoicingEnabled)
	})

	t.Run("Payment Recording", func(t *testing.T) {
		inv := &models.Invoice{
			OrgID:        orgID,
			CustomerName: "Payment Test",
			TotalAmount:  500.0,
			BalanceDue:   500.0,
			Status:       models.StatusPending,
		}
		repo.CreateInvoice(ctx, inv)

		err := repo.RecordPayment(ctx, inv.ID, 200.0)
		require.NoError(t, err)

		fetched, err := repo.GetInvoice(ctx, inv.ID)
		require.NoError(t, err)
		assert.Equal(t, 200.0, fetched.AmountPaid)
		assert.Equal(t, 300.0, fetched.BalanceDue)
		assert.Equal(t, models.StatusPartial, fetched.Status)
	})
}
