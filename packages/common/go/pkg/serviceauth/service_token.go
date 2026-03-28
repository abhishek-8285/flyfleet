// Package serviceauth provides Service-to-Service JWT token generation
// and validation for authenticated inter-service communication within
// the FlyFleet platform.
//
// Usage:
//
//	issuer := serviceauth.NewTokenIssuer("shared-secret", "web-bff")
//	token, _ := issuer.IssueToken()
//	// attach token to outgoing gRPC metadata / HTTP header
//
//	verifier := serviceauth.NewTokenVerifier("shared-secret", "web-bff", "api-gateway") // specify allowed services
//	claims, err := verifier.Verify(token)
package serviceauth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ServiceClaims carries the identity of the calling service.
type ServiceClaims struct {
	ServiceID string `json:"service_id"`
	jwt.RegisteredClaims
}

// TokenIssuer creates short-lived JWTs that identify the calling service.
type TokenIssuer struct {
	secret    []byte
	serviceID string
	ttl       time.Duration
}

// NewTokenIssuer returns a TokenIssuer using the given shared secret.
// serviceID is the logical name of the calling service (e.g. "web-bff").
func NewTokenIssuer(secret, serviceID string) *TokenIssuer {
	return &TokenIssuer{
		secret:    []byte(secret),
		serviceID: serviceID,
		ttl:       5 * time.Minute, // short-lived by design
	}
}

// IssueToken generates a signed JWT identifying this service.
func (ti *TokenIssuer) IssueToken() (string, error) {
	now := time.Now()
	claims := ServiceClaims{
		ServiceID: ti.serviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "flyfleet-s2s",
			Subject:   ti.serviceID,
			ExpiresAt: jwt.NewNumericDate(now.Add(ti.ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(ti.secret)
}

// TokenVerifier validates incoming service tokens.
type TokenVerifier struct {
	secret         []byte
	allowedServiceIDs map[string]bool // Set of service IDs that are accepted; nil or empty means no restriction (INSECURE)
}

// NewTokenVerifier creates a verifier using the shared S2S secret.
// allowedServiceIDs is a list of service IDs that this verifier will accept.
// At least one service ID must be provided for secure operation.
func NewTokenVerifier(secret string, allowedServiceIDs ...string) *TokenVerifier {
	allowed := make(map[string]bool)
	for _, id := range allowedServiceIDs {
		allowed[id] = true
	}
	return &TokenVerifier{
		secret:         []byte(secret),
		allowedServiceIDs: allowed,
	}
}

// Verify parses and validates a service JWT, returning the claims on success.
func (tv *TokenVerifier) Verify(tokenString string) (*ServiceClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ServiceClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tv.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid service token: %w", err)
	}

	claims, ok := token.Claims.(*ServiceClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid service token claims")
	}

	if claims.Issuer != "flyfleet-s2s" {
		return nil, fmt.Errorf("token issuer mismatch: got %q", claims.Issuer)
	}

	// Validate ServiceID is in allowed set
	if len(tv.allowedServiceIDs) > 0 {
		if !tv.allowedServiceIDs[claims.ServiceID] {
			return nil, fmt.Errorf("service not authorized: %s", claims.ServiceID)
		}
	} else {
		// SECURITY: No allowed services configured - this is insecure but allowed for backward compatibility.
		// In production, always specify allowedServiceIDs.
		// Consider using a build tag or environment check to panic in secure builds.
	}

	return claims, nil
}
