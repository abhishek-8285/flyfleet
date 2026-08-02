package routing

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gonotify/gonotify/internal/provider"
	"github.com/gonotify/gonotify/internal/provider/factory"
)

// ErrNoProviderForChannel is returned by ProvidersForChannel when no
// providers (neither DB-configured nor env-based fallback) are available
// for the requested channel.
var ErrNoProviderForChannel = fmt.Errorf("no provider available for channel")

// Router resolves the ordered list of providers for a given tenant and
// channel. It first queries database-configured providers (via the
// routing repository and factory); if none are found, it falls back to
// environment-configured providers passed at construction time.
type Router struct {
	repo           *Repository
	factory        *factory.Factory
	fallbackEmail  provider.Provider
	fallbackSMS    provider.Provider
	logger         *slog.Logger
}

func NewRouter(
	repo *Repository,
	fac *factory.Factory,
	fallbackEmail, fallbackSMS provider.Provider,
	logger *slog.Logger,
) *Router {
	return &Router{
		repo:          repo,
		factory:       fac,
		fallbackEmail: fallbackEmail,
		fallbackSMS:   fallbackSMS,
		logger:        logger,
	}
}

// ProvidersForChannel returns the ordered list of providers to try for
// a channel. If database configs exist, they are built via the factory
// and returned in priority order. If no configs exist, the corresponding
// env-based fallback provider is returned. If no fallback either,
// returns nil with no error.
func (r *Router) ProvidersForChannel(
	ctx context.Context,
	tenantID string,
	channel string,
) ([]provider.SelectedProvider, error) {
	configs, err := r.repo.GetConfigs(ctx, tenantID, channel)
	if err != nil {
		return nil, fmt.Errorf("get provider configs: %w", err)
	}

	if len(configs) == 0 {
		var fb provider.Provider
		switch channel {
		case "email":
			fb = r.fallbackEmail
		case "sms":
			fb = r.fallbackSMS
		default:
			return nil, nil
		}

		if fb == nil {
			return nil, nil
		}

		return []provider.SelectedProvider{
			{Provider: fb, ConfigID: ""},
		}, nil
	}

	selected := make([]provider.SelectedProvider, 0, len(configs))
	for _, cfg := range configs {
		p, err := r.factory.Build(ctx, cfg)
		if err != nil {
			r.logger.WarnContext(ctx, "failed to build provider from config; skipping",
				"config_id", cfg.ID,
				"type", cfg.Type,
				"channel", cfg.Channel,
				"error", err)
			continue
		}
		selected = append(selected, provider.SelectedProvider{
			Provider: p,
			ConfigID: cfg.ID,
		})
	}

	return selected, nil
}
