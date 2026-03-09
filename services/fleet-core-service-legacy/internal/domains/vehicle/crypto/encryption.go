package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
)

// ErrEmptyKey is returned when the encryption key is absent in production.
var ErrEmptyKey = errors.New("VEHICLE_FIELD_ENCRYPTION_KEY not set (must be 32-byte base64 string)")

// ErrKeyLength is returned when the decoded key is not exactly 32 bytes.
var ErrKeyLength = errors.New("VEHICLE_FIELD_ENCRYPTION_KEY decoded length must be exactly 32 bytes (AES-256)")

// encryptionKey is loaded once at startup. It is unexported — never logged.
var encryptionKey []byte

// productionMode is true when APP_ENV=production. Fail-fast if key is absent.
var productionMode bool

func init() {
	productionMode = os.Getenv("APP_ENV") == "production"

	raw := os.Getenv("VEHICLE_FIELD_ENCRYPTION_KEY")
	if raw == "" {
		if productionMode {
			// Hard fail — production must never start without a key.
			panic("FATAL: VEHICLE_FIELD_ENCRYPTION_KEY is required in production (APP_ENV=production)")
		}
		// Dev/test: warn at startup but continue.
		fmt.Fprintln(os.Stderr, "[WARN] VEHICLE_FIELD_ENCRYPTION_KEY not set — chassis/engine numbers will be stored in plaintext (dev/test only)")
		return
	}

	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		panic(fmt.Sprintf("FATAL: VEHICLE_FIELD_ENCRYPTION_KEY is not valid base64: %v", err))
	}
	if len(decoded) != 32 {
		panic(fmt.Sprintf("FATAL: VEHICLE_FIELD_ENCRYPTION_KEY decoded to %d bytes, need exactly 32", len(decoded)))
	}
	encryptionKey = decoded
}

// InitForTest reinitialises the package-level encryption key from a base64 string.
// ONLY intended for use in test code to avoid process-level init() coupling.
// Passing an empty string disables encryption (dev/test mode).
func InitForTest(base64Key string) {
	if base64Key == "" {
		encryptionKey = nil
		return
	}
	decoded, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil || len(decoded) != 32 {
		encryptionKey = nil
		return
	}
	encryptionKey = decoded
}

// Used by callers that want to enforce the key in non-dev contexts (e.g. integration tests).
func IsEncryptionEnabled() bool {
	return len(encryptionKey) == 32
}

// Encrypt encrypts plaintext using AES-256-GCM.
//
// Layout of the returned base64 blob:
//
//	[12-byte random nonce][n-byte ciphertext][16-byte GCM tag]
//
// A fresh nonce is generated via crypto/rand for every call — nonces are never reused.
// SECURITY: plaintext is never logged inside this function.
func Encrypt(plaintext string) (string, error) {
	if !IsEncryptionEnabled() {
		// Dev/test fallback — only safe because productionMode panics at init().
		return plaintext, nil
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("gcm: %w", err)
	}

	// 12-byte nonce (GCM standard). Generated fresh per call.
	nonce := make([]byte, gcm.NonceSize()) // NonceSize() == 12
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("nonce generation failed: %w", err)
	}

	// gcm.Seal prepends nonce: output = nonce || ciphertext || tag
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt decrypts a base64-encoded AES-256-GCM blob produced by Encrypt.
// SECURITY: decrypted value is returned to caller — never logged here.
func Decrypt(encoded string) (string, error) {
	if !IsEncryptionEnabled() {
		return encoded, nil
	}

	sealed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		// Value may be legacy plaintext (migration window); return as-is.
		return encoded, nil
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("gcm: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(sealed) < nonceSize+gcm.Overhead() {
		// Too short to be an encrypted value — treat as legacy plaintext.
		return encoded, nil
	}

	nonce, ciphertext := sealed[:nonceSize], sealed[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		// Authentication tag mismatch — treat as legacy plaintext during migration window.
		return encoded, nil
	}

	return string(plaintext), nil
}

// EncryptPtr encrypts a *string field in-place. Nil / empty pointers are no-ops.
// SECURITY: never log the value of *s before or after this call.
func EncryptPtr(s *string) error {
	if s == nil || *s == "" {
		return nil
	}
	enc, err := Encrypt(*s)
	if err != nil {
		return err
	}
	*s = enc
	return nil
}

// DecryptPtr decrypts a *string field in-place. Nil / empty pointers are no-ops.
// SECURITY: never log the value of *s before or after this call.
func DecryptPtr(s *string) error {
	if s == nil || *s == "" {
		return nil
	}
	dec, err := Decrypt(*s)
	if err != nil {
		return err
	}
	*s = dec
	return nil
}
