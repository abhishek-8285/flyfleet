package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const userTokenIssuer = "gonotify"

type UserClaims struct {
	TenantID       string `json:"tenant_id"`
	ExternalUserID string `json:"external_user_id"`
	jwt.RegisteredClaims
}

func GenerateUserToken(
	secret string,
	userID string,
	tenantID string,
	externalUserID string,
	ttl time.Duration,
) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(ttl)

	claims := UserClaims{
		TenantID:       tenantID,
		ExternalUserID: externalUserID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    userTokenIssuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign user token: %w", err)
	}

	return signed, expiresAt, nil
}

func ParseUserToken(secret string, tokenString string) (UserClaims, error) {
	var claims UserClaims

	token, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}

			return []byte(secret), nil
		},
	)
	if err != nil {
		return UserClaims{}, err
	}

	if !token.Valid {
		return UserClaims{}, errors.New("invalid token")
	}

	if claims.Issuer != userTokenIssuer {
		return UserClaims{}, errors.New("invalid token issuer")
	}

	if claims.Subject == "" {
		return UserClaims{}, errors.New("invalid token subject")
	}

	if claims.TenantID == "" {
		return UserClaims{}, errors.New("invalid token tenant")
	}

	return claims, nil
}
