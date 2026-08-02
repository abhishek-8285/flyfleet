package factory

import (
	"encoding/json"
	"testing"

	"github.com/gonotify/gonotify/internal/crypto"
	"github.com/gonotify/gonotify/internal/provider"
	"github.com/gonotify/gonotify/internal/provider/smtp"
	"github.com/gonotify/gonotify/internal/provider/twilio"
)

func TestBuild_SmtpFromSettings(t *testing.T) {
	f := &Factory{encryptionKey: nil}

	cfg := provider.ProviderConfig{
		Type:    "smtp",
		Channel: "email",
		Settings: map[string]any{
			"host":       "smtp.example.com",
			"port":       587,
			"username":   "smtp_user",
			"password":   "s3cret",
			"from_email": "noreply@example.com",
			"from_name":  "Example",
			"use_tls":    true,
		},
	}

	p, err := f.Build(t.Context(), cfg)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	smtpProv, ok := p.(*smtp.Provider)
	if !ok {
		t.Fatalf("expected *smtp.Provider, got %T", p)
	}

	if smtpProv.Name() != "smtp" {
		t.Errorf("Name() = %q, want smtp", smtpProv.Name())
	}
	if string(smtpProv.Channel()) != "email" {
		t.Errorf("Channel() = %q, want email", smtpProv.Channel())
	}
}

func TestBuild_SmtpFromEncryptedCredentials(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}

	credsJSON, _ := json.Marshal(map[string]string{
		"username": "encrypted_user",
		"password": "encrypted_pass",
	})
	encrypted, err := crypto.Encrypt(credsJSON, key)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	f := &Factory{encryptionKey: key}

	cfg := provider.ProviderConfig{
		Type:    "smtp",
		Channel: "email",
		Settings: map[string]any{
			"host":       "smtp.example.com",
			"port":       587,
			"from_email": "noreply@example.com",
		},
		CredentialsEncrypted: encrypted,
	}

	p, err := f.Build(t.Context(), cfg)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	smtpProv, ok := p.(*smtp.Provider)
	if !ok {
		t.Fatalf("expected *smtp.Provider, got %T", p)
	}

	if smtpProv.Name() != "smtp" {
		t.Errorf("Name() = %q, want smtp", smtpProv.Name())
	}
}

func TestBuild_Twilio(t *testing.T) {
	f := &Factory{encryptionKey: nil}

	cfg := provider.ProviderConfig{
		Type:    "twilio",
		Channel: "sms",
		Settings: map[string]any{
			"account_sid": "AC123",
			"auth_token":  "token123",
			"from":        "+10000000000",
		},
	}

	p, err := f.Build(t.Context(), cfg)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	tw, ok := p.(*twilio.Provider)
	if !ok {
		t.Fatalf("expected *twilio.Provider, got %T", p)
	}

	if tw.Name() != "twilio" {
		t.Errorf("Name() = %q, want twilio", tw.Name())
	}
	if string(tw.Channel()) != "sms" {
		t.Errorf("Channel() = %q, want sms", tw.Channel())
	}
}

func TestBuild_FakeSMS(t *testing.T) {
	f := &Factory{encryptionKey: nil}

	cfg := provider.ProviderConfig{
		Type:    "fakesms",
		Channel: "sms",
	}

	p, err := f.Build(t.Context(), cfg)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if p.Name() != "fake_sms" {
		t.Errorf("Name() = %q, want fake_sms", p.Name())
	}
	if string(p.Channel()) != "sms" {
		t.Errorf("Channel() = %q, want sms", p.Channel())
	}
}

func TestBuild_UnknownType(t *testing.T) {
	f := &Factory{encryptionKey: nil}

	cfg := provider.ProviderConfig{
		Type:    "carrier_pigeon",
		Channel: "email",
	}

	_, err := f.Build(t.Context(), cfg)
	if err == nil {
		t.Fatal("expected error for unknown provider type")
	}
}

func TestGetString(t *testing.T) {
	m := map[string]any{"key": "value", "num": 42}

	if got := getString(m, "key"); got != "value" {
		t.Errorf("getString(key) = %q, want value", got)
	}
	if got := getString(m, "missing"); got != "" {
		t.Errorf("getString(missing) = %q, want empty", got)
	}
	if got := getString(m, "num"); got != "" {
		t.Errorf("getString(num) = %q, want empty (not a string)", got)
	}
	if got := getString(nil, "key"); got != "" {
		t.Errorf("getString(nil map) = %q, want empty", got)
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  int
	}{
		{"int", 587, 587},
		{"int64", int64(587), 587},
		{"float64", float64(587), 587},
		{"string", "587", 0},
		{"missing", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := map[string]any{"port": tt.value}
			got := getInt(m, "port")
			if got != tt.want {
				t.Errorf("getInt = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGetBool(t *testing.T) {
	m := map[string]any{"tls": true, "not_bool": "yes"}

	if !getBool(m, "tls") {
		t.Error("getBool(tls) = false, want true")
	}
	if getBool(m, "missing") {
		t.Error("getBool(missing) = true, want false")
	}
	if getBool(m, "not_bool") {
		t.Error("getBool(not_bool) = true, want false")
	}
	if getBool(nil, "tls") {
		t.Error("getBool(nil map) = true, want false")
	}
}
