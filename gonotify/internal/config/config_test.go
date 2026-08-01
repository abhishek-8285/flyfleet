package config

import (
	"os"
	"testing"
	"time"
)

func TestConfigDefaults(t *testing.T) {
	cfg := Load()

	if cfg.AppEnv == "" {
		t.Error("AppEnv should not be empty")
	}

	if cfg.SMSProvider != "fake" && os.Getenv("SMS_PROVIDER") == "" {
		t.Errorf("SMSProvider default = %q; want 'fake'", cfg.SMSProvider)
	}

	if cfg.HTTPAddr == "" {
		t.Error("HTTPAddr should not be empty")
	}
}

func TestConfigEnvOverride(t *testing.T) {
	os.Setenv("SMS_PROVIDER", "twilio")
	os.Setenv("SMTP_PORT", "2525")
	os.Setenv("STORAGE_USE_SSL", "true")
	os.Setenv("USER_TOKEN_TTL", "2h")
	defer func() {
		os.Unsetenv("SMS_PROVIDER")
		os.Unsetenv("SMTP_PORT")
		os.Unsetenv("STORAGE_USE_SSL")
		os.Unsetenv("USER_TOKEN_TTL")
	}()

	cfg := Load()

	if cfg.SMSProvider != "twilio" {
		t.Errorf("SMSProvider = %q; want 'twilio'", cfg.SMSProvider)
	}

	if cfg.SMTPPort != 2525 {
		t.Errorf("SMTPPort = %d; want 2525", cfg.SMTPPort)
	}

	if !cfg.StorageUseSSL {
		t.Error("StorageUseSSL = false; want true")
	}

	if cfg.UserTokenTTL != 2*time.Hour {
		t.Errorf("UserTokenTTL = %v; want 2h", cfg.UserTokenTTL)
	}
}

func TestConfigInvalidEnvFallbacks(t *testing.T) {
	os.Setenv("SMTP_PORT", "not_a_number")
	os.Setenv("STORAGE_USE_SSL", "not_a_bool")
	os.Setenv("USER_TOKEN_TTL", "not_a_duration")
	defer func() {
		os.Unsetenv("SMTP_PORT")
		os.Unsetenv("STORAGE_USE_SSL")
		os.Unsetenv("USER_TOKEN_TTL")
	}()

	cfg := Load()

	if cfg.SMTPPort != 1025 {
		t.Errorf("SMTPPort fallback = %d; want 1025", cfg.SMTPPort)
	}

	if cfg.StorageUseSSL != false {
		t.Errorf("StorageUseSSL fallback = %v; want false", cfg.StorageUseSSL)
	}

	if cfg.UserTokenTTL != time.Hour {
		t.Errorf("UserTokenTTL fallback = %v; want 1h", cfg.UserTokenTTL)
	}
}
