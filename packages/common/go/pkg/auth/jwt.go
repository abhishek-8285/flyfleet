package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type JWTManager struct {
	secretKey     []byte
	tokenDuration time.Duration
	logger        *zap.Logger
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewJWTManager(secretKey string, tokenDuration time.Duration, logger *zap.Logger) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		tokenDuration: tokenDuration,
		logger:        logger,
	}
}

func (jm *JWTManager) GenerateToken(userID, email, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jm.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "flyfleet",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jm.secretKey)
	if err != nil {
		jm.logger.Error("Failed to generate JWT token", zap.Error(err))
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	jm.logger.Debug("JWT token generated", zap.String("user_id", userID))
	return tokenString, nil
}

func (jm *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jm.secretKey, nil
	})

	if err != nil {
		jm.logger.Warn("Token validation failed", zap.Error(err))
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		jm.logger.Debug("Token validated successfully", zap.String("user_id", claims.UserID))
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

func (jm *JWTManager) RefreshToken(oldTokenString string) (string, error) {
	claims, err := jm.ValidateToken(oldTokenString)
	if err != nil {
		return "", fmt.Errorf("cannot refresh invalid token: %w", err)
	}

	// Generate new token with same claims but new expiration
	newClaims := Claims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jm.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "flyfleet",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	tokenString, err := token.SignedString(jm.secretKey)
	if err != nil {
		jm.logger.Error("Failed to refresh JWT token", zap.Error(err))
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}

	jm.logger.Debug("JWT token refreshed", zap.String("user_id", claims.UserID))
	return tokenString, nil
}