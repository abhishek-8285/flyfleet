package serviceauth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenVerifier_ServiceIDValidation(t *testing.T) {
	secret := "test-secret"
	allowedServices := []string{"service-a", "service-b"}
	verifier := NewTokenVerifier(secret, allowedServices...)

	issuer := NewTokenIssuer(secret, "service-a")
	token, err := issuer.IssueToken()
	require.NoError(t, err)

	// Token with allowed ServiceID should validate
	claims, err := verifier.Verify(token)
	require.NoError(t, err)
	assert.Equal(t, "service-a", claims.ServiceID)

	// Create a token with a different ServiceID (forged)
	forgedToken := createTokenWithServiceID(secret, "service-c")
	_, err = verifier.Verify(forgedToken)
	assert.Error(t, err, "should reject token with unauthorized ServiceID")
}

func TestTokenVerifier_NoAllowedServices(t *testing.T) {
	secret := "test-secret"
	// For backward compatibility, if no allowed services provided, any service is accepted
	// (but this is insecure, so in production always specify allowed services)
	verifier := NewTokenVerifier(secret)

	issuerA := NewTokenIssuer(secret, "service-a")
	tokenA, _ := issuerA.IssueToken()

	claims, err := verifier.Verify(tokenA)
	require.NoError(t, err)
	assert.Equal(t, "service-a", claims.ServiceID)

	issuerB := NewTokenIssuer(secret, "service-b")
	tokenB, _ := issuerB.IssueToken()

	claims, err = verifier.Verify(tokenB)
	require.NoError(t, err)
	assert.Equal(t, "service-b", claims.ServiceID)
}

// Helper to create a token with arbitrary ServiceID (simulating a forged token)
func createTokenWithServiceID(secret, serviceID string) string {
	now := time.Now()
	claims := ServiceClaims{
		ServiceID: serviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "flyfleet-s2s",
			Subject:   serviceID,
			ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(secret))
	return signed
}
