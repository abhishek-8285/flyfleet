package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager interface {
	GenerateToken(claims jwt.MapClaims, expiration time.Duration) (string, error)
	VerifyToken(tokenString string) (*jwt.Token, error)
	ExtractClaims(tokenString string) (jwt.MapClaims, error)
}

type JWTServiceImpl struct {
	secret []byte
}

func NewJWTManager(secret string) JWTManager {
	return &JWTServiceImpl{secret: []byte(secret)}
}

func (j *JWTServiceImpl) GenerateToken(claims jwt.MapClaims, expiration time.Duration) (string, error) {
	if claims["exp"] == nil {
		claims["exp"] = time.Now().Add(expiration).Unix()
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTServiceImpl) VerifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})
	
	if err != nil {
		return nil, err
	}
	
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	
	return token, nil
}

func (j *JWTServiceImpl) ExtractClaims(tokenString string) (jwt.MapClaims, error) {
	token, err := j.VerifyToken(tokenString)
	if err != nil {
		return nil, err
	}
	
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to extract claims")
	}
	
	return claims, nil
}

// Common claim helpers
func CreateAccessToken(userID string, orgID string, roles []string) jwt.MapClaims {
	return jwt.MapClaims{
		"user_id": userID,
		"org_id":  orgID,
		"roles":   roles,
		"type":    "access",
	}
}

func CreateRefreshToken(userID string, orgID string) jwt.MapClaims {
	return jwt.MapClaims{
		"user_id": userID,
		"org_id":  orgID,
		"type":    "refresh",
	}
}