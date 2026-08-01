package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestUserTokenLifecycle(t *testing.T) {
	secret := "dev_secret_key_12345678901234567890"
	userID := "usr_internal_123"
	tenantID := "00000000-0000-0000-0000-000000000001"
	extUserID := "ext_user_456"

	// 1. Generate Token
	tokenString, expiresAt, err := GenerateUserToken(secret, userID, tenantID, extUserID, 1*time.Hour)
	if err != nil {
		t.Fatalf("GenerateUserToken() unexpected error: %v", err)
	}

	if tokenString == "" {
		t.Fatal("GenerateUserToken() returned empty token string")
	}

	if expiresAt.Before(time.Now()) {
		t.Fatalf("expiresAt %v should be in the future", expiresAt)
	}

	// 2. Parse Valid Token
	claims, err := ParseUserToken(secret, tokenString)
	if err != nil {
		t.Fatalf("ParseUserToken() unexpected error: %v", err)
	}

	if claims.Subject != userID {
		t.Errorf("claims.Subject = %q; want %q", claims.Subject, userID)
	}
	if claims.TenantID != tenantID {
		t.Errorf("claims.TenantID = %q; want %q", claims.TenantID, tenantID)
	}
	if claims.ExternalUserID != extUserID {
		t.Errorf("claims.ExternalUserID = %q; want %q", claims.ExternalUserID, extUserID)
	}

	// 3. Reject Token with Wrong Secret
	_, err = ParseUserToken("wrong_secret_key", tokenString)
	if err == nil {
		t.Error("ParseUserToken() expected error for wrong secret, got nil")
	}

	// 4. Reject Expired Token
	expiredToken, _, err := GenerateUserToken(secret, userID, tenantID, extUserID, -1*time.Minute)
	if err != nil {
		t.Fatalf("GenerateUserToken(expired) unexpected error: %v", err)
	}

	_, err = ParseUserToken(secret, expiredToken)
	if err == nil {
		t.Error("ParseUserToken() expected error for expired token, got nil")
	}
}

func TestUserTokenInvalidClaims(t *testing.T) {
	secret := "dev_secret_key_12345678901234567890"

	// 1. Invalid Issuer
	badIssuerClaims := UserClaims{
		TenantID: "tenant_1",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:  "bad_issuer",
			Subject: "usr_1",
		},
	}
	tBadIssuer := jwt.NewWithClaims(jwt.SigningMethodHS256, badIssuerClaims)
	sBadIssuer, _ := tBadIssuer.SignedString([]byte(secret))
	_, err := ParseUserToken(secret, sBadIssuer)
	if err == nil {
		t.Error("ParseUserToken() expected error for bad issuer, got nil")
	}

	// 2. Empty Subject
	noSubjectClaims := UserClaims{
		TenantID: "tenant_1",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:  userTokenIssuer,
			Subject: "",
		},
	}
	tNoSub := jwt.NewWithClaims(jwt.SigningMethodHS256, noSubjectClaims)
	sNoSub, _ := tNoSub.SignedString([]byte(secret))
	_, err = ParseUserToken(secret, sNoSub)
	if err == nil {
		t.Error("ParseUserToken() expected error for empty subject, got nil")
	}

	// 3. Empty Tenant
	noTenantClaims := UserClaims{
		TenantID: "",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:  userTokenIssuer,
			Subject: "usr_1",
		},
	}
	tNoTenant := jwt.NewWithClaims(jwt.SigningMethodHS256, noTenantClaims)
	sNoTenant, _ := tNoTenant.SignedString([]byte(secret))
	_, err = ParseUserToken(secret, sNoTenant)
	if err == nil {
		t.Error("ParseUserToken() expected error for empty tenant, got nil")
	}

	// 4. SigningMethodNone (unexpected signing method)
	tNone := jwt.NewWithClaims(jwt.SigningMethodNone, badIssuerClaims)
	sNone, _ := tNone.SignedString(jwt.UnsafeAllowNoneSignatureType)
	_, err = ParseUserToken(secret, sNone)
	if err == nil {
		t.Error("ParseUserToken() expected error for none signing method, got nil")
	}
}
