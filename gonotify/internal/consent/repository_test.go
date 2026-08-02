package consent

import (
	"testing"
)

func TestConsentActions(t *testing.T) {
	if ActionOptIn != "opt_in" {
		t.Errorf("ActionOptIn = %q; want %q", ActionOptIn, "opt_in")
	}
	if ActionOptOut != "opt_out" {
		t.Errorf("ActionOptOut = %q; want %q", ActionOptOut, "opt_out")
	}
}

func TestConsentLogStruct(t *testing.T) {
	cl := ConsentLog{
		ID:       "test-id",
		TenantID: "tenant-uuid",
		UserID:   "user-uuid",
		Category: "marketing",
		Channel:  "email",
		Action:   ActionOptIn,
		Source:   "user_api",
		Metadata: map[string]any{"ip": "127.0.0.1"},
	}

	if cl.Action != ActionOptIn {
		t.Errorf("expected action %q, got %q", ActionOptIn, cl.Action)
	}
	if cl.Category != "marketing" {
		t.Errorf("expected category %q, got %q", "marketing", cl.Category)
	}
}

func TestHasActiveConsentLogic(t *testing.T) {
	// Tests the conceptual logic: only opt_in actions should return true.
	testCases := []struct {
		action string
		want   bool
	}{
		{ActionOptIn, true},
		{ActionOptOut, false},
		{"", false},
		{"unknown", false},
	}

	for _, tc := range testCases {
		result := tc.action == ActionOptIn
		if result != tc.want {
			t.Errorf("action %q == ActionOptIn: got %v, want %v",
				tc.action, result, tc.want)
		}
	}
}
