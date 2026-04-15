package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"math/big"
	"net/http"
	"strings"

	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// contextKey is a private type to prevent context key collisions.
type contextKey string

const (
	CtxUserID    contextKey = "user_id"
	CtxOrgID     contextKey = "org_id"
	CtxUserRoles contextKey = "user_roles"
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

		ctx := context.WithValue(r.Context(), CtxUserID, claims.UserID)
		ctx = context.WithValue(ctx, CtxOrgID, claims.OrgID)
		
		// Set context keys for all standard locations
		ctx = context.WithValue(ctx, ctxkeys.UserIDKey, claims.UserID.String())
		ctx = context.WithValue(ctx, ctxkeys.OrgIDKey, claims.OrgID.String())
		ctx = context.WithValue(ctx, ctxkeys.TenantIDKey, claims.OrgID.String())
		
		// Handle Role (singular 'rol' or first from 'roles' slice)
		role := claims.Role
		if role == "" && len(claims.Roles) > 0 {
			role = claims.Roles[0]
		}
		ctx = context.WithValue(ctx, ctxkeys.RoleKey, role)
		ctx = context.WithValue(ctx, ctxkeys.PermissionMaskKey, float64(claims.PermissionMask))
		
		// Set IAM handler string keys (legacy support)
		ctx = context.WithValue(ctx, "user_id", claims.UserID.String())
		ctx = context.WithValue(ctx, "org_id", claims.OrgID.String())
		ctx = context.WithValue(ctx, "role", role)
		ctx = context.WithValue(ctx, "role_list", claims.Roles)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type Claims struct {
	UserID         uuid.UUID `json:"sub"`
	OrgID          uuid.UUID `json:"org_id"`
	Role           string    `json:"rol"`
	Roles          []string  `json:"roles"`
	PermissionMask uint64    `json:"pmsk"`
	Permissions    []string  `json:"prm"`
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
		return nil, jwt.ErrTokenSignatureInvalid
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
