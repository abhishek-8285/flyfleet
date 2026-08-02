package preference

import (
	"testing"
)

func TestValidCategory(t *testing.T) {
	valid := []string{
		"transactional",
		"security",
		"marketing",
		"billing",
		"product_update",
		"social",
		"digest",
	}
	for _, c := range valid {
		if !ValidCategory(c) {
			t.Errorf("ValidCategory(%q) = false; want true", c)
		}
	}

	invalid := []string{"invalid", "unknown", "", "spam"}
	for _, c := range invalid {
		if ValidCategory(c) {
			t.Errorf("ValidCategory(%q) = true; want false", c)
		}
	}
}

func TestValidChannel(t *testing.T) {
	valid := []string{"email", "sms", "push", "in_app"}
	for _, ch := range valid {
		if !ValidChannel(ch) {
			t.Errorf("ValidChannel(%q) = false; want true", ch)
		}
	}

	invalid := []string{"telegram", "carrier_pigeon", "", "slack"}
	for _, ch := range invalid {
		if ValidChannel(ch) {
			t.Errorf("ValidChannel(%q) = true; want false", ch)
		}
	}
}

func TestDefaultEnabled(t *testing.T) {
	if DefaultEnabled("marketing", "email") {
		t.Error("DefaultEnabled('marketing', 'email') = true; want false")
	}
	if DefaultEnabled("digest", "email") {
		t.Error("DefaultEnabled('digest', 'email') = true; want false")
	}
	if !DefaultEnabled("transactional", "email") {
		t.Error("DefaultEnabled('transactional', 'email') = false; want true")
	}
	if !DefaultEnabled("security", "sms") {
		t.Error("DefaultEnabled('security', 'sms') = false; want true")
	}
}

func TestAllCategoriesAndChannels(t *testing.T) {
	cats := AllCategories()
	if len(cats) != 7 {
		t.Errorf("AllCategories() len = %d; want 7", len(cats))
	}
	chans := AllChannels()
	if len(chans) != 4 {
		t.Errorf("AllChannels() len = %d; want 4", len(chans))
	}
}
