package delivery

import (
	"errors"
	"testing"
	"time"
)

func TestRetryDelay(t *testing.T) {
	tests := []struct {
		name     string
		attempt  int
		policy   string
		expected time.Duration
	}{
		{"exponential attempt 0", 0, "exponential", 0},
		{"exponential attempt 1", 1, "exponential", 1 * time.Minute},
		{"exponential attempt 2", 2, "exponential", 5 * time.Minute},
		{"exponential attempt 3", 3, "exponential", 30 * time.Minute},
		{"exponential attempt 4", 4, "exponential", 2 * time.Hour},
		{"exponential attempt 5", 5, "exponential", 24 * time.Hour},
		{"exponential attempt 6 (capped)", 6, "exponential", 24 * time.Hour},
		{"fixed policy", 5, "fixed", 1 * time.Minute},
		{"empty policy defaults to exponential", 2, "", 5 * time.Minute},
		{"negative attempt", -1, "exponential", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RetryDelay(tt.attempt, tt.policy)
			if got != tt.expected {
				t.Errorf("RetryDelay(%d, %q) = %v, want %v",
					tt.attempt, tt.policy, got, tt.expected)
			}
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"network timeout", errors.New("dial tcp: connection refused"), true},
		{"500 server error", errors.New("twilio error: status=500"), true},
		{"provider auth failed", errors.New("SMTP 535 authentication failed"), true},
		{"recipient email required", errors.New("recipient email is required"), false},
		{"recipient phone required", errors.New("recipient phone is required"), false},
		{"invalid phone", errors.New("invalid phone number format"), false},
		{"missing consent", errors.New("missing consent for channel"), false},
		{"preference disabled", errors.New("preference disabled for email"), false},
		{"suppressed address", errors.New("address is suppressed"), false},
		{"template not found", errors.New("template not found"), false},
		{"wrapped non-retryable", errors.New("wrapped: " + "template not found"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetryableError(tt.err)
			if got != tt.expected {
				t.Errorf("isRetryableError(%v) = %v, want %v",
					tt.err, got, tt.expected)
			}
		})
	}
}
