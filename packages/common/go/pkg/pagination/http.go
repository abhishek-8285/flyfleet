package pagination

import (
	"context"
	"net/http"
	"strconv"
)

type contextKey string

const PaginationKey contextKey = "pagination"

// Middleware extracts and normalizes pagination params from query string.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := Params{Page: 1, Limit: 20}

		query := r.URL.Query()
		if v, err := strconv.Atoi(query.Get("page")); err == nil {
			p.Page = v
		}

		if v, err := strconv.Atoi(query.Get("limit")); err == nil {
			p.Limit = v
		} else if v, err := strconv.Atoi(query.Get("page_size")); err == nil {
			p.Limit = v
			w.Header().Set("X-Deprecation-Notice", "page_size is deprecated, use limit")
		}

		p.Normalize()
		ctx := context.WithValue(r.Context(), PaginationKey, p)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetParams retrieves pagination params from request context.
func GetParams(r *http.Request) Params {
	if v := r.Context().Value(PaginationKey); v != nil {
		if p, ok := v.(Params); ok {
			return p
		}
	}
	// Fallback: parse directly
	p := Params{Page: 1, Limit: 20}
	query := r.URL.Query()
	if v, err := strconv.Atoi(query.Get("page")); err == nil && v > 0 {
		p.Page = v
	}
	if v, err := strconv.Atoi(query.Get("limit")); err == nil && v > 0 {
		p.Limit = v
	} else if v, err := strconv.Atoi(query.Get("page_size")); err == nil && v > 0 {
		p.Limit = v
	}
	p.Normalize()
	return p
}
