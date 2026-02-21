package clocksync

import (
	"fmt"
	"time"
)

// VerifySync checks if the system clock is reasonably synchronized
// This is a basic check; in production, NTP status should be checked via OS.
func VerifySync() error {
	now := time.Now()
	// Basic check: if 'now' is before project start date, something is wrong.
	projectStart := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if now.Before(projectStart) {
		return fmt.Errorf("system clock seems out of sync: %v", now)
	}
	return nil
}

// GetTimestamp returns a UTC timestamp with standard formatting
func GetTimestamp() time.Time {
	return time.Now().UTC()
}
