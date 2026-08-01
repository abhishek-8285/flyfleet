package observability

import (
	"testing"
)

func TestNewLoggerLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error", "unknown"}

	for _, lvl := range levels {
		t.Run(lvl, func(t *testing.T) {
			logger := NewLogger(lvl)
			if logger == nil {
				t.Fatalf("NewLogger(%q) returned nil", lvl)
			}
		})
	}
}
