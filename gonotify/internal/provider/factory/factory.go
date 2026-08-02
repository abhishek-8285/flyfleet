package factory

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gonotify/gonotify/internal/crypto"
	"github.com/gonotify/gonotify/internal/provider"
	"github.com/gonotify/gonotify/internal/provider/fakesms"
	"github.com/gonotify/gonotify/internal/provider/smtp"
	"github.com/gonotify/gonotify/internal/provider/twilio"
)

var ErrUnknownProviderType = fmt.Errorf("unknown provider type")

type Factory struct {
	db            *pgxpool.Pool
	encryptionKey []byte
}

func NewFactory(db *pgxpool.Pool, encryptionKey []byte) *Factory {
	return &Factory{db: db, encryptionKey: encryptionKey}
}

// Build constructs a provider.Provider from a database config row.
// It decrypts credentials (if present) and maps settings to the
// appropriate provider constructor.
func (f *Factory) Build(ctx context.Context, cfg provider.ProviderConfig) (provider.Provider, error) {
	switch cfg.Type {
	case "smtp":
		return f.buildSMTP(cfg)
	case "twilio":
		return f.buildTwilio(cfg)
	case "fakesms":
		return f.buildFakeSMS(cfg)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownProviderType, cfg.Type)
	}
}

// decryptCredentials decrypts the credentials_encrypted BLOB from the
// provider_configs row. Returns nil (with no error) when no encrypted
// credentials are stored.
func (f *Factory) decryptCredentials(cfg provider.ProviderConfig) (map[string]string, error) {
	if len(cfg.CredentialsEncrypted) == 0 {
		return nil, nil
	}

	decrypted, err := crypto.Decrypt(cfg.CredentialsEncrypted, f.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("decrypt credentials for %s: %w", cfg.Type, err)
	}

	var creds map[string]string
	if err := json.Unmarshal(decrypted, &creds); err != nil {
		return nil, fmt.Errorf("parse credentials JSON: %w", err)
	}

	return creds, nil
}

func (f *Factory) buildSMTP(cfg provider.ProviderConfig) (provider.Provider, error) {
	settings := cfg.Settings
	creds, err := f.decryptCredentials(cfg)
	if err != nil {
		return nil, err
	}

	var username, password string
	if len(creds) > 0 {
		username = creds["username"]
		password = creds["password"]
	} else {
		username = getString(settings, "username")
		password = getString(settings, "password")
	}

	return smtp.New(smtp.Config{
		Host:      getString(settings, "host"),
		Port:      getInt(settings, "port"),
		Username:  username,
		Password:  password,
		FromEmail: getString(settings, "from_email"),
		FromName:  getString(settings, "from_name"),
		UseTLS:    getBool(settings, "use_tls"),
	}), nil
}

func (f *Factory) buildTwilio(cfg provider.ProviderConfig) (provider.Provider, error) {
	settings := cfg.Settings
	creds, err := f.decryptCredentials(cfg)
	if err != nil {
		return nil, err
	}

	var accountSID, authToken string
	if len(creds) > 0 {
		accountSID = creds["account_sid"]
		authToken = creds["auth_token"]
	} else {
		accountSID = getString(settings, "account_sid")
		authToken = getString(settings, "auth_token")
	}

	return twilio.New(twilio.Config{
		AccountSID: accountSID,
		AuthToken:  authToken,
		From:       getString(settings, "from"),
	}), nil
}

func (f *Factory) buildFakeSMS(cfg provider.ProviderConfig) (provider.Provider, error) {
	return fakesms.New(fakesms.Config{
		DB:         f.db,
		ForceError: getBool(cfg.Settings, "force_error"),
	}), nil
}

// --- helpers for extracting typed values from map[string]any ---

func getString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(m map[string]any, key string) int {
	if m == nil {
		return 0
	}
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case int:
			return n
		case int64:
			return int(n)
		case float64:
			return int(n)
		}
	}
	return 0
}

func getBool(m map[string]any, key string) bool {
	if m == nil {
		return false
	}
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}
