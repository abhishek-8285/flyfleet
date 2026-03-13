package middleware

import (
	"net/http"

	"github.com/flyfleet/fleet-core-service/internal/platform/billing/service"
	iam_middleware "github.com/flyfleet/fleet-core-service/internal/platform/iam/middleware"
	"github.com/google/uuid"
)

func PlanLimitMiddleware(billingSvc service.BillingService, metric string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgID, ok := r.Context().Value(iam_middleware.OrgIDKey).(uuid.UUID)
			if !ok {
				http.Error(w, "organization context missing", http.StatusForbidden)
				return
			}

			allowed, err := billingSvc.CheckLimit(r.Context(), orgID, metric)
			if err != nil {
				http.Error(w, "failed to check plan limits", http.StatusInternalServerError)
				return
			}

			if !allowed {
				http.Error(w, "plan limit reached: "+metric, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
