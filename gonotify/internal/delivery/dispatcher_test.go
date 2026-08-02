package delivery

import (
	"encoding/json"
	"testing"

	"github.com/gonotify/gonotify/internal/queue"
)

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already formatted e164",
			input:    "+14155552671",
			expected: "+14155552671",
		},
		{
			name:     "spaces and hyphens",
			input:    "+1 (415) 555-2671",
			expected: "+14155552671",
		},
		{
			name:     "dots and parentheses",
			input:    "+1.415.555.2671",
			expected: "+14155552671",
		},
		{
			name:     "whitespace padding",
			input:    "   +1-800-555-0199   ",
			expected: "+18005550199",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizePhone(tt.input)
			if got != tt.expected {
				t.Errorf("normalizePhone(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestE164RegexValidation(t *testing.T) {
	tests := []struct {
		phone string
		valid bool
	}{
		{phone: "+14155552671", valid: true},
		{phone: "+919876543210", valid: true},
		{phone: "+442071838750", valid: true},
		{phone: "14155552671", valid: false},        // Missing +
		{phone: "+0123456", valid: false},           // Country code cannot start with 0
		{phone: "+14155552671999999", valid: false}, // Too long (>15 digits)
		{phone: "invalid_phone", valid: false},
		{phone: "", valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.phone, func(t *testing.T) {
			got := e164Regex.MatchString(tt.phone)
			if got != tt.valid {
				t.Errorf("e164Regex.MatchString(%q) = %v; want %v", tt.phone, got, tt.valid)
			}
		})
	}
}

func TestParseChannels(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected []string
	}{
		{
			name:     "empty string defaults to email and in_app",
			raw:      "",
			expected: []string{"email", "in_app"},
		},
		{
			name:     "explicit sms channel request",
			raw:      "sms",
			expected: []string{"sms"},
		},
		{
			name:     "multi-channel request with deduplication",
			raw:      "email, in_app, sms, email",
			expected: []string{"email", "in_app", "sms"},
		},
		{
			name:     "invalid channels filtered out",
			raw:      "invalid_channel, carrier_pigeon",
			expected: []string{"email", "in_app"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseChannels(tt.raw)
			if len(got) != len(tt.expected) {
				t.Fatalf("parseChannels(%q) length = %d; want %d (%v)", tt.raw, len(got), len(tt.expected), got)
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("parseChannels(%q)[%d] = %q; want %q", tt.raw, i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestStringValue(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{name: "string input", input: "+14155552671", expected: "+14155552671"},
		{name: "float64 input", input: float64(4999), expected: "4999"},
		{name: "json.Number input", input: json.Number("12345"), expected: "12345"},
		{name: "nil input", input: nil, expected: ""},
		{name: "int input", input: 99, expected: "99"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringValue(tt.input)
			if got != tt.expected {
				t.Errorf("stringValue(%v) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSMSTemplatesForEvent(t *testing.T) {
	tests := []struct {
		eventType string
		contains  string
	}{
		{eventType: "order.created", contains: "order {{.order_id}}"},
		{eventType: "payment.failed", contains: "payment for order"},
		{eventType: "unknown.event", contains: "Notification: {{.event}}"},
	}

	for _, tt := range tests {
		t.Run(tt.eventType, func(t *testing.T) {
			tmpl := smsTemplatesForEvent(tt.eventType)
			if tmpl == "" {
				t.Errorf("smsTemplatesForEvent(%q) returned empty template", tt.eventType)
			}
		})
	}
}

func TestTerminalStatusHelpers(t *testing.T) {
	terminalSuccess := []string{"sent", "delivered", "opened", "clicked"}
	for _, status := range terminalSuccess {
		if !isTerminalSuccessStatus(status) {
			t.Errorf("isTerminalSuccessStatus(%q) = false; want true", status)
		}
		if !isTerminalStatus(status) {
			t.Errorf("isTerminalStatus(%q) = false; want true", status)
		}
	}

	otherTerminal := []string{"suppressed", "canceled", "expired"}
	for _, status := range otherTerminal {
		if isTerminalSuccessStatus(status) {
			t.Errorf("isTerminalSuccessStatus(%q) = true; want false", status)
		}
		if !isTerminalStatus(status) {
			t.Errorf("isTerminalStatus(%q) = false; want true", status)
		}
	}

	nonTerminal := []string{"sending", "pending", "queued", "unknown"}
	for _, status := range nonTerminal {
		if isTerminalStatus(status) {
			t.Errorf("isTerminalStatus(%q) = true; want false", status)
		}
	}
}

func TestTemplatesForEvent(t *testing.T) {
	subj, html, text := templatesForEvent("order.created")
	if subj == "" || html == "" || text == "" {
		t.Error("templatesForEvent('order.created') returned empty template string")
	}

	subjDef, htmlDef, textDef := templatesForEvent("custom.event")
	if subjDef == "" || htmlDef == "" || textDef == "" {
		t.Error("templatesForEvent('custom.event') returned empty template string")
	}
}

func TestNotificationKeyAndInboxKey(t *testing.T) {
	nKey1 := notificationKey("evt_1", "email", "USER@example.com")
	nKey2 := notificationKey("evt_1", "email", "user@example.com")

	if nKey1 == "" {
		t.Error("notificationKey returned empty string")
	}

	if nKey1 != nKey2 {
		t.Errorf("notificationKey should be case-insensitive: %q != %q", nKey1, nKey2)
	}

	iKey := inboxKey("evt_1", "usr_123")
	if iKey == "" {
		t.Error("inboxKey returned empty string")
	}
}

func TestIsValidChannel(t *testing.T) {
	valid := []string{"email", "in_app", "sms", "push", "webhook"}
	for _, ch := range valid {
		if !isValidChannel(ch) {
			t.Errorf("isValidChannel(%q) = false; want true", ch)
		}
	}

	invalid := []string{"telegram", "slack", "carrier_pigeon", ""}
	for _, ch := range invalid {
		if isValidChannel(ch) {
			t.Errorf("isValidChannel(%q) = true; want false", ch)
		}
	}
}

func TestEventCategory(t *testing.T) {
	msgWithMeta := queue.Message{
		Metadata: map[string]string{
			"category": "security",
		},
	}
	if got := eventCategory(msgWithMeta, nil); got != "security" {
		t.Errorf("eventCategory metadata = %q; want security", got)
	}

	msgWithData := queue.Message{Metadata: map[string]string{}}
	data := map[string]any{"category": "marketing"}
	if got := eventCategory(msgWithData, data); got != "marketing" {
		t.Errorf("eventCategory data = %q; want marketing", got)
	}

	msgDefault := queue.Message{Metadata: map[string]string{}}
	if got := eventCategory(msgDefault, map[string]any{}); got != "transactional" {
		t.Errorf("eventCategory default = %q; want transactional", got)
	}
}

