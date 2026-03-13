package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/pkg/ctxkeys"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/fleet-platform/shared/pkg/middleware"
)

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWKSAuthGate handles JWKS-based JWT validation
type JWKSAuthGate struct {
	jwksURL string
	logger  *zap.Logger
	keys    map[string]*rsa.PublicKey
	issuer  string
	audience string
	mu      sync.RWMutex
}

// NewJWKSAuthGate creates a new JWKSAuthGate
func NewJWKSAuthGate(jwksURL, issuer, audience string, logger *zap.Logger) *JWKSAuthGate {
	gate := &JWKSAuthGate{
		jwksURL:  jwksURL,
		logger:   logger,
		issuer:   issuer,
		audience: audience,
		keys:     make(map[string]*rsa.PublicKey),
	}

	// Initial fetch
	_ = gate.refreshKeys()

	// Start background refresher
	go gate.refreshLoop()

	return gate
}

func (g *JWKSAuthGate) refreshLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		if err := g.refreshKeys(); err != nil {
			g.logger.Error("Failed to refresh JWKS", zap.Error(err))
		}
	}
}

func (g *JWKSAuthGate) refreshKeys() error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(g.jwksURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status: %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return err
	}

	newKeys := make(map[string]*rsa.PublicKey)
	for _, key := range jwks.Keys {
		if key.Kty != "RSA" {
			continue
		}

		pubKey, err := decodeRSAPublicKey(key.N, key.E)
		if err != nil {
			g.logger.Warn("Failed to decode RSA public key", zap.String("kid", key.Kid), zap.Error(err))
			continue
		}
		newKeys[key.Kid] = pubKey
	}

	g.mu.Lock()
	g.keys = newKeys
	g.mu.Unlock()

	g.logger.Info("JWKS keys refreshed", zap.Int("count", len(newKeys)))
	return nil
}

func decodeRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, err
	}

	var e int
	if len(eBytes) < 4 {
		paddedE := make([]byte, 4)
		copy(paddedE[4-len(eBytes):], eBytes)
		e = int(big.NewInt(0).SetBytes(paddedE).Uint64())
	} else {
		e = int(big.NewInt(0).SetBytes(eBytes).Uint64())
	}

	return &rsa.PublicKey{
		N: big.NewInt(0).SetBytes(nBytes),
		E: e,
	}, nil
}

func (g *JWKSAuthGate) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, errors.New("missing kid in header")
			}

			g.mu.RLock()
			pubKey, ok := g.keys[kid]
			g.mu.RUnlock()

			if !ok {
				// Try a single on-demand refresh if key is missing
				g.logger.Info("KID not found, attempting on-demand JWKS refresh", zap.String("kid", kid))
				if err := g.refreshKeys(); err == nil {
					g.mu.RLock()
					pubKey, ok = g.keys[kid]
					g.mu.RUnlock()
				}
			}

			if !ok {
				return nil, fmt.Errorf("unknown kid after refresh: %s", kid)
			}
			
			return pubKey, nil
		}, jwt.WithIssuer(g.issuer), jwt.WithAudience(g.audience))

		if err != nil || !token.Valid {
			g.logger.Warn("Invalid token", zap.Error(err))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Inject into context
		ctx := r.Context()

		// User ID
		if sub, ok := claims["sub"].(string); ok {
			ctx = context.WithValue(ctx, ctxkeys.UserIDKey, sub) // Legacy (string)
			if uid, err := uuid.Parse(sub); err == nil {
				ctx = context.WithValue(ctx, middleware.UserIDKey, uid) // Modern (uuid.UUID)
			}
		}

		// Organization ID
		if orgID, ok := claims["org_id"].(string); ok {
			ctx = context.WithValue(ctx, ctxkeys.OrgIDKey, orgID) // Legacy (string)
			if oid, err := uuid.Parse(orgID); err == nil {
				ctx = context.WithValue(ctx, middleware.OrgIDKey, oid) // Modern (uuid.UUID)
			}
		}

		// Role
		if role, ok := claims["rol"].(string); ok {
			ctx = context.WithValue(ctx, ctxkeys.RoleKey, role)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
