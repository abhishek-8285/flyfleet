package service

import "testing"

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{email: "support@test.com", valid: true},
		{email: "Support.Team+Ops@Test.COM", valid: true},
		{email: "a@b.technology", valid: true},
		{email: "missing-at-sign.example.com", valid: false},
		{email: "bad-domain@test", valid: false},
	}

	for _, tt := range tests {
		if got := isValidEmail(tt.email); got != tt.valid {
			t.Fatalf("isValidEmail(%q)=%v want %v", tt.email, got, tt.valid)
		}
	}
}
