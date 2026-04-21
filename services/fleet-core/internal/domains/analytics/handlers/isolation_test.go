//go:build integration
// +build integration

package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/flyfleet/fleet-core/internal/domains/analytics"
	analyticsMiddleware "github.com/flyfleet/fleet-core/internal/domains/analytics/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestTenantGuard_Ownership(t *testing.T) {
	logger := zap.NewNop()
	metrics := analytics.NewMetrics()

	// Create pgxmock
	mockDB, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mockDB.Close()

	guard := analyticsMiddleware.NewTenantGuard(mockDB, metrics, logger)
	mockService := new(MockAnalyticsService)
	handler := NewAnalyticsHandler(mockService, guard, logger)

	tenantID := uuid.New()
	otherTenantID := uuid.New()
	vehicleID := uuid.New()

	t.Run("Authorized Vehicle Access", func(t *testing.T) {
		// Mock DB check: vehicle exists for this org
		mockDB.ExpectQuery("SELECT EXISTS").
			WithArgs(vehicleID, tenantID).
			WillReturnRows(mockDB.NewRows([]string{"exists"}).AddRow(true))

		req, _ := http.NewRequest("GET", "/stats/vehicle/"+vehicleID.String(), nil)
		// Inject org ID into context (as the authentication middleware would)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, tenantID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		r := chi.NewRouter()
		handler.RegisterRoutes(r)

		mockService.On("GetVehicleDailyStats", mock.Anything, vehicleID, tenantID, mock.Anything, mock.Anything).
			Return(nil, nil).Once()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})

	t.Run("Unauthorized Vehicle Access (Cross-Tenant)", func(t *testing.T) {
		// Mock DB check: vehicle does NOT belong to this org
		mockDB.ExpectQuery("SELECT EXISTS").
			WithArgs(vehicleID, otherTenantID).
			WillReturnRows(mockDB.NewRows([]string{"exists"}).AddRow(false))

		req, _ := http.NewRequest("GET", "/stats/vehicle/"+vehicleID.String(), nil)
		ctx := context.WithValue(req.Context(), middleware.TenantIDKey, otherTenantID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		r := chi.NewRouter()
		handler.RegisterRoutes(r)

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "Forbidden: Vehicle does not belong to this organization")
		assert.NoError(t, mockDB.ExpectationsWereMet())
	})
}
