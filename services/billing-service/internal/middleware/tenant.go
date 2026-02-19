package middleware

import (
	"net/http"

	"github.com/flyfleet/billing-service/pkg/errors"
)

func Tenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetOrgID(r.Context())
		if !ok {
			errors.WriteProblem(w, errors.Forbidden("org_id missing or invalid"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

