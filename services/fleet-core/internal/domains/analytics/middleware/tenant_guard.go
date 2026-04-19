package middleware

import (
	"context"
	"net/http"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/flyfleet/fleet-core/internal/domains/analytics"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// DB is a subset of pgxpool.Pool methods needed for TenantGuard.
type DB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// TenantGuard ensures that the requested resource (vehicle, driver) belongs to the organization.
type TenantGuard struct {
	db      DB
	metrics *analytics.Metrics
	logger  *zap.Logger
}

// NewTenantGuard creates a new instance of TenantGuard.
func NewTenantGuard(db DB, metrics *analytics.Metrics, logger *zap.Logger) *TenantGuard {
	return &TenantGuard{
		db:      db,
		metrics: metrics,
		logger:  logger,
	}
}

// ResourceOwnershipGuard returns a middleware that checks resource ownership.
func (tg *TenantGuard) ResourceOwnershipGuard() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgID, ok := r.Context().Value(middleware.OrgIDKey).(uuid.UUID)
			if !ok {
				http.Error(w, "Organization context missing", http.StatusUnauthorized)
				return
			}

			// Check Vehicle ID if present in URL
			vehicleIDStr := chi.URLParam(r, "vehicleID")
			if vehicleIDStr == "" {
				vehicleIDStr = r.URL.Query().Get("vehicle_id")
			}
			if vehicleIDStr != "" {
				if vid, err := uuid.Parse(vehicleIDStr); err == nil {
					if !tg.verifyVehicleOwnership(r.Context(), vid, orgID) {
						tg.metrics.CrossTenantAccessTotal.WithLabelValues(orgID.String(), "vehicle").Inc()
						tg.logger.Warn("Cross-tenant vehicle access attempt blocked",
							zap.String("tenant_id", orgID.String()), zap.String("vehicle_id", vehicleIDStr))
						http.Error(w, "Forbidden: Vehicle does not belong to this organization", http.StatusForbidden)
						return
					}
				}
			}

			// Check Driver ID if present in URL
			driverIDStr := chi.URLParam(r, "driverID")
			if driverIDStr == "" {
				driverIDStr = r.URL.Query().Get("driver_id")
			}
			if driverIDStr != "" {
				if did, err := uuid.Parse(driverIDStr); err == nil {
					if !tg.verifyDriverOwnership(r.Context(), did, orgID) {
						tg.metrics.CrossTenantAccessTotal.WithLabelValues(orgID.String(), "driver").Inc()
						tg.logger.Warn("Cross-tenant driver access attempt blocked",
							zap.String("tenant_id", orgID.String()), zap.String("driver_id", driverIDStr))
						http.Error(w, "Forbidden: Driver does not belong to this organization", http.StatusForbidden)
						return
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (tg *TenantGuard) verifyVehicleOwnership(ctx context.Context, vehicleID, orgID uuid.UUID) bool {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM vehicles WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL)"
	err := tg.db.QueryRow(ctx, query, vehicleID, orgID).Scan(&exists)
	if err != nil {
		tg.logger.Error("Failed to verify vehicle ownership", zap.Error(err))
		return false
	}
	return exists
}

func (tg *TenantGuard) verifyDriverOwnership(ctx context.Context, driverID, orgID uuid.UUID) bool {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM drivers WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL)"
	err := tg.db.QueryRow(ctx, query, driverID, orgID).Scan(&exists)
	if err != nil {
		tg.logger.Error("Failed to verify driver ownership", zap.Error(err))
		return false
	}
	return exists
}
