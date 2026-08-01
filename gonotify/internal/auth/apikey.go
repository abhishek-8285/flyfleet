package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

func ParseBearer(header string) (string, error) {
	const prefix = "Bearer "

	if len(header) < len(prefix) {
		return "", errors.New("missing bearer token")
	}

	if !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", errors.New("invalid authorization scheme")
	}

	token := strings.TrimSpace(header[len(prefix):])
	if token == "" {
		return "", errors.New("missing bearer token")
	}

	return token, nil
}

func HashAPIKey(raw string) string {
	normalized := strings.TrimSpace(raw)
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}

func GenerateAPIKey(environment string) (string, string, string, error) {
	if environment == "" {
		environment = "dev"
	}

	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", "", "", err
	}

	token := hex.EncodeToString(b)
	raw := fmt.Sprintf("gn_%s_%s", environment, token)

	prefix := raw
	if len(raw) >= 16 {
		prefix = raw[:16]
	}

	hash := HashAPIKey(raw)

	return raw, prefix, hash, nil
}
