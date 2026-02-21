package featureflags

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Flag represents a single feature flag with a fallback value.
type Flag struct {
	Name         string
	DefaultValue bool
	Description  string
}

// Common feature flags used across services
var (
	FlagGRPCTelemetry  = Flag{Name: "grpc_telemetry", DefaultValue: true, Description: "Use gRPC for telemetry ingestion"}
	FlagBatchIngestion = Flag{Name: "batch_ingestion", DefaultValue: true, Description: "Enable batch telemetry ingestion"}
	FlagStrictTenant   = Flag{Name: "strict_tenant", DefaultValue: false, Description: "Enforce strict org_id on all routes"}
	FlagAuditLogging   = Flag{Name: "audit_logging", DefaultValue: true, Description: "Enable hash-chained audit logging"}
	FlagPIIMasking     = Flag{Name: "pii_masking", DefaultValue: true, Description: "Mask PII in logs"}
	FlagCanaryDeploy   = Flag{Name: "canary_deploy", DefaultValue: false, Description: "Route 10% traffic to canary"}
)

// Store is the feature flag store with in-memory cache and fallback defaults.
type Store struct {
	mu       sync.RWMutex
	flags    map[string]bool
	logger   *zap.Logger
	provider Provider
}

// Provider is an interface for external flag sources (e.g., Redis, ConfigMap, LaunchDarkly).
type Provider interface {
	FetchFlags(ctx context.Context) (map[string]bool, error)
}

// NewStore creates a feature flag store with optional external provider.
func NewStore(logger *zap.Logger, provider Provider) *Store {
	s := &Store{
		flags:    make(map[string]bool),
		logger:   logger,
		provider: provider,
	}

	// Register defaults
	for _, f := range []Flag{
		FlagGRPCTelemetry, FlagBatchIngestion, FlagStrictTenant,
		FlagAuditLogging, FlagPIIMasking, FlagCanaryDeploy,
	} {
		s.flags[f.Name] = f.DefaultValue
	}

	return s
}

// IsEnabled checks if a feature flag is enabled, falling back to default if provider fails.
func (s *Store) IsEnabled(flag Flag) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if val, ok := s.flags[flag.Name]; ok {
		return val
	}
	return flag.DefaultValue
}

// Override sets a flag value (for testing or runtime config updates).
func (s *Store) Override(name string, value bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flags[name] = value
	s.logger.Info("Feature flag overridden", zap.String("flag", name), zap.Bool("value", value))
}

// StartPolling periodically refreshes flags from the external provider.
func (s *Store) StartPolling(ctx context.Context, interval time.Duration) {
	if s.provider == nil {
		s.logger.Info("No feature flag provider configured, using defaults")
		return
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				flags, err := s.provider.FetchFlags(ctx)
				if err != nil {
					s.logger.Warn("Failed to fetch feature flags, using cached values",
						zap.Error(err))
					continue
				}

				s.mu.Lock()
				for k, v := range flags {
					s.flags[k] = v
				}
				s.mu.Unlock()

				s.logger.Debug("Feature flags refreshed", zap.Int("count", len(flags)))
			}
		}
	}()
}
