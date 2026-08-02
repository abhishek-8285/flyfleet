package suppression

import (
	"testing"
)

func TestIsValidChannel(t *testing.T) {
	valid := []string{"email", "sms", "push", "in_app"}
	for _, ch := range valid {
		if !IsValidChannel(ch) {
			t.Errorf("IsValidChannel(%q) = false; want true", ch)
		}
	}

	invalid := []string{"invalid", "", "carrier_pigeon", "telegram", "slack"}
	for _, ch := range invalid {
		if IsValidChannel(ch) {
			t.Errorf("IsValidChannel(%q) = true; want false", ch)
		}
	}
}

func TestValidateAddress(t *testing.T) {
	tests := []struct {
		name    string
		channel string
		address string
		wantErr bool
	}{
		{name: "valid email", channel: "email", address: "user@example.com", wantErr: false},
		{name: "valid email with uppercase", channel: "email", address: "USER@EXAMPLE.COM", wantErr: false},
		{name: "invalid email no @", channel: "email", address: "userexample.com", wantErr: true},
		{name: "invalid email empty", channel: "email", address: "", wantErr: true},
		{name: "valid sms E.164", channel: "sms", address: "+14155552671", wantErr: false},
		{name: "valid sms with formatting", channel: "sms", address: "+1 (415) 555-2671", wantErr: false},
		{name: "invalid sms missing +", channel: "sms", address: "14155552671", wantErr: true},
		{name: "invalid sms empty", channel: "sms", address: "", wantErr: true},
		{name: "invalid sms too long", channel: "sms", address: "+14155552671999999", wantErr: true},
		{name: "push non-empty", channel: "push", address: "device-token-123", wantErr: false},
		{name: "push empty", channel: "push", address: "", wantErr: true},
		{name: "in_app non-empty", channel: "in_app", address: "user-id-123", wantErr: false},
		{name: "in_app empty", channel: "in_app", address: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAddress(tt.channel, tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAddress(%q, %q) error = %v; wantErr = %v",
					tt.channel, tt.address, err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeAddress(t *testing.T) {
	tests := []struct {
		name     string
		channel  string
		address  string
		expected string
	}{
		{name: "email lowercased", channel: "email", address: "User@Example.COM", expected: "user@example.com"},
		{name: "email trimmed", channel: "email", address: "  user@example.com  ", expected: "user@example.com"},
		{name: "email trimmed and lowercased", channel: "email", address: "  USER@EXAMPLE.COM  ", expected: "user@example.com"},
		{name: "sms trimmed", channel: "sms", address: "  +14155552671  ", expected: "+14155552671"},
		{name: "push trimmed", channel: "push", address: "  device-token  ", expected: "device-token"},
		{name: "in_app trimmed", channel: "in_app", address: "  user-id  ", expected: "user-id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeAddress(tt.channel, tt.address)
			if got != tt.expected {
				t.Errorf("normalizeAddress(%q, %q) = %q; want %q",
					tt.channel, tt.address, got, tt.expected)
			}
		})
	}
}

func TestIsUniqueViolation(t *testing.T) {
	if isUniqueViolation(nil) {
		t.Error("isUniqueViolation(nil) = true; want false")
	}

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"duplicate key string", errString("ERROR: duplicate key value violates unique constraint"), true},
		{"unique_violation string", errString("ERROR: duplicate key value violates unique_violation"), true},
		{"23505 string", errString("ERROR: 23505 unique_violation"), true},
		{"generic error", errString("connection refused"), false},
		{"not found", errString("no rows in result set"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUniqueViolation(tt.err)
			if got != tt.want {
				t.Errorf("isUniqueViolation() = %v; want %v", got, tt.want)
			}
		})
	}
}

// errString implements the error interface for testing.
type errString string

func (e errString) Error() string { return string(e) }
