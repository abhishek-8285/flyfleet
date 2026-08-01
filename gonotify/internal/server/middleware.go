package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gonotify/gonotify/internal/auth"
	"github.com/gonotify/gonotify/internal/tenant"
	"github.com/gonotify/gonotify/internal/user"
)

func APIKeyAuth(
	db *pgxpool.Pool,
	logger *slog.Logger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawKey, err := auth.ParseBearer(r.Header.Get("Authorization"))
			if err != nil {
				WriteError(
					w,
					http.StatusUnauthorized,
					"unauthorized",
					"missing or invalid authorization header",
				)
				return
			}

			keyHash := auth.HashAPIKey(rawKey)

			var apiKey auth.APIKey
			var tenantStatus string

			err = db.QueryRow(
				r.Context(),
				`
				SELECT
					ak.id::text,
					ak.tenant_id::text,
					ak.environment,
					ak.scopes,
					t.status
				FROM api_keys ak
				JOIN tenants t ON t.id = ak.tenant_id
				WHERE ak.key_hash = $1
				  AND ak.revoked_at IS NULL
				  AND (ak.expires_at IS NULL OR ak.expires_at > now())
				`,
				keyHash,
			).Scan(
				&apiKey.ID,
				&apiKey.TenantID,
				&apiKey.Environment,
				&apiKey.Scopes,
				&tenantStatus,
			)

			if errors.Is(err, pgx.ErrNoRows) {
				WriteError(
					w,
					http.StatusUnauthorized,
					"unauthorized",
					"invalid api key",
				)
				return
			}

			if err != nil {
				logger.ErrorContext(
					r.Context(),
					"api key lookup failed",
					"error", err,
				)

				WriteError(
					w,
					http.StatusInternalServerError,
					"internal_error",
					"failed to authenticate request",
				)
				return
			}

			if tenantStatus != "active" {
				WriteError(
					w,
					http.StatusForbidden,
					"tenant_inactive",
					"tenant is not active",
				)
				return
			}

			ctx := auth.WithAPIKey(r.Context(), apiKey)
			ctx = tenant.WithTenantID(ctx, apiKey.TenantID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey, ok := auth.APIKeyFromContext(r.Context())
			if !ok {
				WriteError(
					w,
					http.StatusUnauthorized,
					"unauthorized",
					"unauthenticated request",
				)
				return
			}

			if !apiKey.HasScope(scope) {
				WriteError(
					w,
					http.StatusForbidden,
					"insufficient_scope",
					fmt.Sprintf("scope '%s' is required", scope),
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func UserAuth(
	jwtSecret string,
	logger *slog.Logger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			var rawToken string
			var err error
			if authHeader != "" {
				rawToken, err = auth.ParseBearer(authHeader)
			} else if qToken := r.URL.Query().Get("token"); qToken != "" {
				rawToken = qToken
			} else {
				err = errors.New("missing authorization token")
			}
			if err != nil {
				WriteError(
					w,
					http.StatusUnauthorized,
					"unauthorized",
					"missing or invalid authorization token",
				)
				return
			}

			claims, err := auth.ParseUserToken(jwtSecret, rawToken)
			if err != nil {
				WriteError(
					w,
					http.StatusUnauthorized,
					"unauthorized",
					"invalid user token",
				)
				return
			}

			authenticatedUser := user.AuthenticatedUser{
				ID:             claims.Subject,
				TenantID:       claims.TenantID,
				ExternalUserID: claims.ExternalUserID,
			}

			ctx := user.WithAuthenticatedUser(r.Context(), authenticatedUser)
			ctx = tenant.WithTenantID(ctx, claims.TenantID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
