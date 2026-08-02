package delivery

import (
	"context"
	"errors"
	"fmt"

	"github.com/gonotify/gonotify/internal/provider"
)

// channelProviderSelector returns the ordered list of providers for a
// given tenant and channel, allowing the dispatcher to try them in
// priority/weight order with automatic fallback.
type channelProviderSelector interface {
	ProvidersForChannel(ctx context.Context, tenantID string, channel string) ([]provider.SelectedProvider, error)
}

var ErrNoProviderForChannel = errors.New("no provider configured for channel")

// errNoProviderSentinel wraps ErrNoProviderForChannel with channel info
// so callers can still match with errors.Is.
func noProviderErr(channel string) error {
	return fmt.Errorf("%w: %s", ErrNoProviderForChannel, channel)
}
