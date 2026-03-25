package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/pkg/ctxkeys"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// GetUserID retrieves the user ID from context.
func GetUserID(ctx context.Context) uuid.UUID {
	val, ok := ctx.Value(ctxkeys.UserIDKey).(string)
	if !ok {
		return uuid.Nil
	}
	id, _ := uuid.Parse(val)
	return id
}

// GetOrgID retrieves the organization ID from context.
func GetOrgID(ctx context.Context) uuid.UUID {
	val, ok := ctx.Value(ctxkeys.OrgIDKey).(string)
	if !ok {
		return uuid.Nil
	}
	id, _ := uuid.Parse(val)
	return id
}

// NewBFFAuthGate enforces bearer token auth on API routes while allowing configured public prefixes.
func NewBFFAuthGate(jwtSecret string, publicPrefixes []string) func(http.Handler) http.Handler {
	clean := make([]string, 0, len(publicPrefixes))
	for _, p := range publicPrefixes {
		p = strings.TrimSpace(p)
		if p != "" {
			clean = append(clean, p)
		}
	}

	isPublic := func(path string) bool {
		for _, p := range clean {
			if strings.HasPrefix(path, p) {
				return true
			}
		}
		return false
	}

	validate := func(token string) (jwt.MapClaims, bool) {
		if token == "" {
			return nil, false
		}
		if jwtSecret == "" {
			// If secret is unavailable, keep gate as presence-check with empty claims for MVP.
			return jwt.MapClaims{}, true
		}

		claims := jwt.MapClaims{}
		parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !parsed.Valid {
			return nil, false
		}
		if expRaw, ok := claims["exp"]; ok {
			switch v := expRaw.(type) {
			case float64:
				if int64(v) < time.Now().Unix() {
					return nil, false
				}
			}
		}
		return claims, true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if isPublic(path) {
				next.ServeHTTP(w, r)
				return
			}

			// Enforce on API surfaces only.
			if !strings.HasPrefix(path, "/api/") {
				next.ServeHTTP(w, r)
				return
			}

			auth := strings.TrimSpace(r.Header.Get("Authorization"))
			if !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}
			token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
			claims, ok := validate(token)
			if !ok {
				http.Error(w, "invalid bearer token", http.StatusUnauthorized)
				return
			}

			// Strip incoming spoofable headers
			r.Header.Del("X-User-ID")
			r.Header.Del("X-Org-ID")
			r.Header.Del("X-Organization-ID")
			r.Header.Del("X-Role-ID")

			// Inject claims into context
			ctx := r.Context()
			if userID, ok := claims["user_id"].(string); ok {
				ctx = context.WithValue(ctx, ctxkeys.UserIDKey, userID)
				r.Header.Set("X-User-ID", userID)
			} else if sub, ok := claims["sub"].(string); ok {
				ctx = context.WithValue(ctx, ctxkeys.UserIDKey, sub)
				r.Header.Set("X-User-ID", sub)
			}

			if orgID, ok := claims["org_id"].(string); ok {
				ctx = context.WithValue(ctx, ctxkeys.OrgIDKey, orgID)
				r.Header.Set("X-Org-ID", orgID)
			}

			if role, ok := claims["rol"].(string); ok {
				ctx = context.WithValue(ctx, ctxkeys.RoleKey, role)
				r.Header.Set("X-Role-ID", role)
			} else if role, ok := claims["role"].(string); ok {
				ctx = context.WithValue(ctx, ctxkeys.RoleKey, role)
				r.Header.Set("X-Role-ID", role)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
