package preference

func ValidCategory(category string) bool {
	switch category {
	case
		"transactional",
		"security",
		"marketing",
		"billing",
		"product_update",
		"social",
		"digest":
		return true
	default:
		return false
	}
}

func ValidChannel(channel string) bool {
	switch channel {
	case
		"email",
		"sms",
		"push",
		"in_app":
		return true
	default:
		return false
	}
}

func DefaultEnabled(category string, channel string) bool {
	// Marketing must be explicit opt-in.
	if category == "marketing" {
		return false
	}

	// Digest is usually optional too.
	if category == "digest" {
		return false
	}

	// Transactional, security, billing are enabled by default.
	return true
}

func AllCategories() []string {
	return []string{
		"transactional",
		"security",
		"marketing",
		"billing",
		"product_update",
		"social",
		"digest",
	}
}

func AllChannels() []string {
	return []string{
		"email",
		"sms",
		"push",
		"in_app",
	}
}
