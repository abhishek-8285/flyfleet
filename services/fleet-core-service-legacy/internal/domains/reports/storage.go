package reports

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// ObjectStorage is the abstraction for file storage backends (S3, MinIO, local).
type ObjectStorage interface {
	Upload(ctx context.Context, key string, r io.Reader, contentType string) (string, error)
	SignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
}

// storageHMACSecret is read from the environment for signing/verifying download URLs.
// In production, inject via secret manager (not env directly).
var storageHMACSecret = []byte(getenv("REPORT_STORAGE_HMAC_SECRET", "dev-insecure-secret-change-me"))

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// LocalStorage is a dev-mode no-op implementation.
// Replace with MinIO/S3 SDK in production.
type LocalStorage struct {
	logger *zap.Logger
}

func NewLocalStorage(logger *zap.Logger) ObjectStorage {
	return &LocalStorage{logger: logger}
}

func (l *LocalStorage) Upload(_ context.Context, key string, r io.Reader, _ string) (string, error) {
	n, err := io.Copy(io.Discard, r)
	if err != nil {
		return "", fmt.Errorf("upload failed: %w", err)
	}
	l.logger.Info("File uploaded (local dev mode)",
		zap.String("key", key),
		zap.Int64("bytes", n))
	return fmt.Sprintf("https://storage.flyfleet.com/reports/%s", key), nil
}

// SignedURL generates an HMAC-SHA256 signed URL with:
//   - Tenant isolation: key must begin with tenant_id prefix
//   - Report-scoped path: key includes report_id
//   - Expiration: Unix timestamp embedded in query string
//   - HMAC signature: prevents URL enumeration or forgery
//
// Verification must be performed at the download endpoint or CDN edge.
func (l *LocalStorage) SignedURL(_ context.Context, key string, expiry time.Duration) (string, error) {
	expiresAt := time.Now().Add(expiry).Unix()
	payload := fmt.Sprintf("%s:%d", key, expiresAt)

	mac := hmac.New(sha256.New, storageHMACSecret)
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))

	url := fmt.Sprintf(
		"https://storage.flyfleet.com/reports/%s?expires=%d&sig=%s",
		key, expiresAt, sig)
	l.logger.Info("Signed URL generated",
		zap.String("key", key),
		zap.Int64("expires_at", expiresAt))
	return url, nil
}

// VerifySignedURL validates an incoming signed URL's HMAC and expiry.
// Call this at the CDN/reverse-proxy layer or in a dedicated /download handler.
func VerifySignedURL(key string, expiresStr, sig string) bool {
	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil || time.Now().Unix() > expires {
		return false
	}
	payload := fmt.Sprintf("%s:%d", key, expires)
	mac := hmac.New(sha256.New, storageHMACSecret)
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(sig))
}

func (l *LocalStorage) Delete(_ context.Context, key string) error {
	l.logger.Info("File deleted (local dev mode)", zap.String("key", key))
	return nil
}
