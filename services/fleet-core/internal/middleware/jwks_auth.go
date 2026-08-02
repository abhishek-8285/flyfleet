package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	sharedmiddleware "github.com/fleet-platform/shared/pkg/middleware"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)


// JWKSAuthHandler handles token validation using JWKS or shared secret.
type JWKSAuthHandler struct {
	issuer   string
	audience string
	secret   string
	logger   *zap.Logger
}

// NewJWKSAuthGate creates a new JWKSAuthHandler.
func NewJWKSAuthGate(issuer, audience, secret string, logger *zap.Logger) *JWKSAuthHandler {
	return &JWKSAuthHandler{
		issuer:   issuer,
		audience: audience,
		secret:   secret,
		logger:   logger,
	}
}

// Handler returns the middleware handler.
func (j *JWKSAuthHandler) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenStr string

		// Try Authorization header first
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			if tokenStr == authHeader {
				http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
				return
			}
		}

		// Fall back to query param
		if tokenStr == "" {
			tokenStr = r.URL.Query().Get("token")
		}

		if tokenStr == "" {
			http.Error(w, `{"error":"token required"}`, http.StatusUnauthorized)
			return
		}

		claims, err := validateToken(tokenStr, j.secret, j.issuer)
		if err != nil {
			j.logger.Debug("jwt validation failed", zap.Error(err))
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}

		ctx := r.Context()

		// Resolve the effective tenant ID (prefer new claim, fall back to legacy org_id)
		effectiveTenantID := claims.TenantID
		if effectiveTenantID == uuid.Nil {
			effectiveTenantID = claims.LegacyTenantID
		}

		// Resolve role
		role := claims.Role
		if role == "" && len(claims.Roles) > 0 {
			role = claims.Roles[0]
		}

		// ── Set context values ONCE each, consistently as UUID / typed values ──
		// Internal ctxkeys (UUID type — GetTenantID/GetUserID handle both UUID and string)
		ctx = context.WithValue(ctx, ctxkeys.TenantIDKey, effectiveTenantID)
		ctx = context.WithValue(ctx, ctxkeys.UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, ctxkeys.RoleKey, role)
		ctx = context.WithValue(ctx, ctxkeys.RolesKey, claims.Roles)
		ctx = context.WithValue(ctx, ctxkeys.PermissionMaskKey, uint64(claims.PermissionMask))
		traceID, _ := ctx.Value("trace_id").(string)
		if traceID == "" {
			traceID, _ = ctx.Value("request_id").(string)
		}
		if traceID == "" {
			if u, ok := ctxkeys.GetRequestID(ctx); ok {
				traceID = u.String()
			}
		}
		j.logger.Info("JWT Auth Context Set",
			zap.String("trace_id", traceID),
			zap.Uint64("permission_mask", uint64(claims.PermissionMask)),
			zap.String("user_id", claims.UserID.String()),
			zap.String("tenant_id", effectiveTenantID.String()))



		// Shared library keys (UUID type — consumed by billing/tracking/shared domains)
		ctx = context.WithValue(ctx, sharedmiddleware.TenantIDKey, effectiveTenantID)
		ctx = context.WithValue(ctx, sharedmiddleware.UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, sharedmiddleware.RoleKey, role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type Claims struct {
	UserID            uuid.UUID  `json:"sub"`
	TenantID          uuid.UUID  `json:"tenant_id"`
	LegacyTenantID    uuid.UUID  `json:"org_id,omitempty"`
	Role              string     `json:"role"`
	Roles             []string   `json:"roles"`
	PermissionMask    int64      `json:"permissions_mask"`
	Permissions       []string   `json:"permissions"`
	jwt.RegisteredClaims
}

func validateToken(tokenStr, secret, jwksURL string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		// 1. Check if it's an RSA token with a KID
		if _, ok := t.Method.(*jwt.SigningMethodRSA); ok {
			kid, _ := t.Header["kid"].(string)
			if kid != "" && strings.HasPrefix(jwksURL, "http") {
				// Construct cache key or just fetch for now (test environment)
				resp, err := http.Get(jwksURL)
				if err == nil {
					defer resp.Body.Close()
					var jwks struct {
						Keys []struct {
							Kid string `json:"kid"`
							N   string `json:"n"`
							E   string `json:"e"`
						} `json:"keys"`
					}
					if err := json.NewDecoder(resp.Body).Decode(&jwks); err == nil {
						for _, k := range jwks.Keys {
							if k.Kid == kid {
								// Construct public key from N and E
								nBytes, _ := base64.RawURLEncoding.DecodeString(k.N)
								eBytes, _ := base64.RawURLEncoding.DecodeString(k.E)
								if len(eBytes) < 4 {
									pad := make([]byte, 4-len(eBytes))
									eBytes = append(pad, eBytes...)
								}
								e := binary.BigEndian.Uint32(eBytes)
								return &rsa.PublicKey{
									N: new(big.Int).SetBytes(nBytes),
									E: int(e),
								}, nil
							}
						}
					}
				}
			}
		}

		// 2. Fallback to HMAC with shared secret (for legacy tests)
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("token signature invalid (check JWT_SECRET synchronization): %w", jwt.ErrTokenSignatureInvalid)
	}
	return claims, nil
}

// ValidateJWTFromQuery extracts and validates a JWT from the URL query
// string. Used for WebSocket handshake where headers can't be set.
func ValidateJWTFromQuery(r *http.Request, jwtSecret string) (*Claims, error) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		return nil, jwt.ErrTokenMalformed
	}
	return validateToken(tokenStr, jwtSecret, "")
}
