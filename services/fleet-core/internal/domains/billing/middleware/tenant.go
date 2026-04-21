package middleware

import (
	"net/http"
)

func Tenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetTenantID(r.Context())
		if !ok {
			http.Error(w, "tenant_id missing or invalid", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
