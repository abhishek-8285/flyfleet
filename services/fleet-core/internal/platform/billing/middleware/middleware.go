package middleware

import (
	"net/http"

	"github.com/flyfleet/fleet-core/internal/platform/billing/service"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/google/uuid"
)

func PlanLimitMiddleware(billingSvc service.BillingService, metric string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tenantID uuid.UUID
			var ok bool
			if val := r.Context().Value(ctxkeys.TenantIDKey); val != nil {
				if id, isUUID := val.(uuid.UUID); isUUID {
					tenantID = id
					ok = true
				} else if s, isStr := val.(string); isStr {
					if id, err := uuid.Parse(s); err == nil {
						tenantID = id
						ok = true
					}
				}
			}

			if !ok {
				http.Error(w, "organization context missing", http.StatusForbidden)
				return
			}

			allowed, err := billingSvc.CheckLimit(r.Context(), tenantID, metric)
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
